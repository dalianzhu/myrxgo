package myrxgo

import (
	"context"
	"log"
	"reflect"
	"sync"
	"time"
)

type Observable struct {
	C            chan interface{}
	OnClose      func()
	Name         string
	WaitClose    context.Context
	Cancel       context.CancelFunc
	OnStepFinish func(interface{})
	Lock         sync.Mutex
}

func newObservable() *Observable {
	o := new(Observable)
	o.C = make(chan interface{})
	o.OnClose = func() {}
	ctx, cancel := context.WithCancel(context.Background())
	o.WaitClose = ctx
	o.Cancel = cancel
	o.OnStepFinish = func(i interface{}) {
	}

	return o
}

func (o *Observable) close() {
	safeGo(func(i ...interface{}) {
		log.Printf("ob %v close", o.Name)
		o.Cancel()
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
				outOb.OnStepFinish(e.Interface())
			}
		}
		outOb.close()
	})
	return outOb
}

func FromChan(c chan interface{}) *Observable {
	outOb := newObservable()
	outOb.Name = UUID()[:8]
	log.Printf("ob %v run, FromChan", outOb.Name)
	safeGo(func(i ...interface{}) {
		for item := range c {
			outOb.C <- item
			outOb.OnStepFinish(item)
		}
		outOb.close()
	})
	return outOb
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
				outOb.OnStepFinish(ret)
			})
		}
		outOb.close()
	})
	return outOb
}

func FromStream(source *Observable) *Observable {
	source.Lock.Lock()
	defer source.Lock.Unlock()

	inOutOb := make(chan interface{})
	outOb := FromChan(inOutOb)
	outOb.Name = source.Name + "-FromStream"

	sourceOnStepFinish := source.OnStepFinish
	source.OnStepFinish = func(i interface{}) {
		sourceOnStepFinish(i)
		inOutOb <- i
	}

	sourceClose := source.OnClose
	source.OnClose = func() {
		sourceClose()
		close(inOutOb)
	}
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
			//log.Printf("ClonePtr %v enqueue %v", refOb.Name, item)
			select {
			case <-refOb.WaitClose.Done():
				log.Printf("%v WaitClose %v %v", refOb.Name, item,
					refOb.WaitClose.Err())
			case refOb.C <- item:
				refOb.OnStepFinish(item)
				//log.Printf("%v send %v", refOb.Name, x)
			}

			// 对主线发数据
			select {
			case <-outOb.WaitClose.Done():
			case outOb.C <- item:
				outOb.OnStepFinish(item)
			}
		}

		refOb.close()
		outOb.close()
	})

	return outOb
}

func (o *Observable) Subscribe(obs IObserver) chan int {
	log.Println("run", o.Name, "start")
	fin := make(chan int, 1)
	go func() {
		for item := range o.C {
			safeRun(func() {
				switch v := item.(type) {
				case error:
					obs.OnErr(v)
				default:
					obs.OnNext(v)
				}
			})
		}
		log.Println("run", o.Name, "exit")
		fin <- 1
	}()
	return fin
}

func (o *Observable) Run(fn func(i interface{})) {
	log.Println("run", o.Name, "start")
	for item := range o.C {
		fn(item)
	}
	log.Println("run", o.Name, "exit")
}
