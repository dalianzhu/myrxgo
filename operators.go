package observable

import (
	"log"
	"sync"
)

func (o *Observable) Map(fc func(interface{}) interface{}) *Observable {
	outOb := newObservable()
	outOb.Name = o.Name + "-Map"
	log.Printf("ob %v run", outOb.Name)

	safeGo(func(i ...interface{}) {
		futures := make(chan *Future, 1000)
		safeGo(func(i ...interface{}) {
			for item := range o.C {
				future := NewFuture()
				safeGo(func(i ...interface{}) {
					ret := fc(i[0])
					future.SetResult(ret)
				}, item)
				futures <- future
			}
			close(futures)
		})

		for future := range futures {
			outOb.C <- future.GetResult()
		}
		outOb.close()
	})
	return outOb
}

func (o *Observable) FlatMap(fn func(interface{}) *Observable) *Observable {
	outOb := newObservable()
	outOb.Name = o.Name + "-FlatMap"
	log.Printf("ob %v run", outOb.Name)

	safeGo(func(i ...interface{}) {
		var wg sync.WaitGroup
		for item := range o.C {
			safeGo(func(i ...interface{}) {
				wg.Add(1)
				defer wg.Done()
				applyOb := fn(i[0])
				applyOb.Run(
					func(i interface{}) {
						outOb.C <- i
					},
				)
			}, item)

		}
		wg.Wait()
		outOb.close()
	})
	return outOb
}

func (o *Observable) Filter(fc func(interface{}) bool) *Observable {
	outOb := newObservable()
	outOb.Name = o.Name + "-Filter"
	log.Printf("ob %v run", outOb.Name)
	futures := make(chan *Future, 1000)
	safeGo(func(i ...interface{}) {
		safeGo(func(i ...interface{}) {
			for item := range o.C {
				future := NewFuture()
				safeGo(func(i ...interface{}) {
					ok := fc(i[0])
					v := i[0]
					future.SetResult([]interface{}{
						ok, v,
					})
				}, item)
				futures <- future
			}
			close(futures)
		})
		for future := range futures {
			v := future.GetResult().([]interface{})
			if v[0].(bool) {
				outOb.C <- v[1]
			}
		}
		outOb.close()
	})

	return outOb
}

func (o *Observable) AsList() *Observable {
	outOb := newObservable()
	outOb.Name = o.Name + "-AsList"
	log.Printf("ob %v run", outOb.Name)

	go func() {
		ret := make([]interface{}, 0, 5)
		for item := range o.C {
			ret = append(ret, item)
		}
		outOb.C <- ret
		outOb.close()
	}()
	return outOb
}
