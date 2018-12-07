package myrxgo

import (
	"log"
	"reflect"
	"sync"
	"time"
)

type Observable struct {
	C         chan interface{}
	OnClose   func()
	Name      string
	WaitClose chan int
}

func newObservable() *Observable {
	o := new(Observable)
	o.C = make(chan interface{})
	o.OnClose = func() {}
	o.WaitClose = make(chan int, 1)
	return o
}

func (o *Observable) close() {
	log.Printf("ob %v close", o.Name)
	close(o.WaitClose)
	safeGo(func(i ...interface{}) {
		time.Sleep(time.Microsecond * 10)
		close(o.C)
		o.OnClose()
	})
}

func From(arr interface{}) *Observable {
	outOb := newObservable()
	outOb.Name = UUID()[:8]
	log.Printf("ob %v run, From", outOb.Name)
	safeGo(func(i ...interface{}) {
		val := reflect.ValueOf(arr)
		if val.Kind() == reflect.Slice {
			for i := 0; i < val.Len(); i++ {
				e := val.Index(i)
				outOb.C <- e.Interface()
			}
		}
		outOb.close()
	})
	return outOb
}

func FromChan(c chan interface{}) *Observable {
	o := new(Observable)
	o.C = c
	o.OnClose = func() {}
	o.WaitClose = make(chan int, 1)
	return o
}

func (o *Observable) Merge(inputObservable *Observable,
	fc func(interface{}, interface{}) interface{}) *Observable {
	outOb := newObservable()
	outOb.Name = o.Name + "-Merge"
	log.Printf("ob %v run", outOb.Name)

	safeGo(func(i ...interface{}) {
		for item := range o.C {
			ifItem, ok := <-inputObservable.C
			if !ok {
				break
			}
			safeRun(func() {
				ret := fc(item, ifItem)
				outOb.C <- ret
			})
		}
		outOb.close()
	})
	return outOb
}

func (o *Observable) ClonePtr(ob *Observable) *Observable {
	refOb := newObservable()
	refOb.Name = o.Name + "-refClonePtr"
	log.Printf("ob %v run", refOb.Name)

	*ob = *refOb

	outOb := newObservable()
	outOb.Name = o.Name + "-ClonePtr"
	log.Printf("ob %v run", outOb.Name)

	safeGo(func(i ...interface{}) {
		for item := range o.C {
			// 对支线发数据
			safeGo(func(i ...interface{}) {
				select {
				case <-refOb.WaitClose:
					//log.Printf("%v WaitClose", refOb.Name)
				case refOb.C <- i[0]:
				case <-time.After(time.Second * 3):
					log.Printf("%v droped item %v", refOb.Name, item)
				}

			}, item)
			// 对主线发数据
			safeGo(func(i ...interface{}) {
				select {
				case <-outOb.WaitClose:
				case outOb.C <- i[0]:
				}
			}, item)
		}
		refOb.close()
		outOb.close()
	})

	return outOb
}

func (o *Observable) Subscribe(obs IOnNext) {
	log.Println("run", o.Name, "start")
	var wg sync.WaitGroup
	for item := range o.C {
		safeGo(func(i ...interface{}) {
			wg.Add(1)
			defer wg.Done()
			obs.OnNext(i[0])
		}, item)
	}
	wg.Wait()
	log.Println("run", o.Name, "exit")
}

func (o *Observable) Run(fn func(i interface{})) {
	log.Println("run", o.Name, "start")
	var wg sync.WaitGroup
	for item := range o.C {
		safeGo(func(i ...interface{}) {
			wg.Add(1)
			defer wg.Done()
			fn(i[0])
		}, item)
	}
	wg.Wait()
	log.Println("run", o.Name, "exit")
}
