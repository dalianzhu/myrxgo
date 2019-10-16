package myrxgo

import (
	"sync"
)

/*
所有的error都被直接传递，不引发任何操作，setNext函数将做出操作
*/

func isError(i interface{}) bool {
	_, ok := i.(error)
	return ok
}

func (o *Observable) Map(fc func(interface{}) interface{}) IObservable {
	outOb := newObservable()
	outOb.SetName(o.name + "-Map")
	Debugf("ob %v run", outOb.GetName())

	safeGo(func(i ...interface{}) {
		for item := range o.outputC {
			outOb.SetNext(item, func(i interface{}) interface{} {
				future := NewFuture()
				safeGo(func(i ...interface{}) {
					item := i[0]
					future.SetResult(fc(item))
				}, i)
				return future
			})
		}
		outOb.Close()
	})
	return outOb
}

func (o *Observable) FlatMapPara(fn func(interface{}) IObservable) IObservable {
	outOb := newObservable()
	outOb.SetName(o.name + "-FlatMapPara")
	Debugf("ob %v run", outOb.GetName())
	safeGo(func(i ...interface{}) {
		for item := range o.outputC {
			//Debugf("flatMapPara item:%v", item)
			if isError(item) {
				outOb.SetNext(item, func(i interface{}) interface{} {
					return i
				})
				continue
			}

			applyOb := fn(item)
			<-applyOb.Subscribe(NewObserverWithErrDone(
				func(i interface{}) {
					outOb.SetNext(i, func(i interface{}) interface{} {
						return i
					})
				},
				func(e error) {
					//Debugf("FlatMapPara find error!")
					outOb.SetNext(e, func(i interface{}) interface{} {
						return i
					})
				},
				func() {}))
		}
		outOb.Close()
	})
	return outOb
}

func (o *Observable) FlatMap(fn func(interface{}) IObservable) IObservable {
	outOb := newObservable()
	outOb.SetName(o.name + "-FlatMap")
	Debugf("ob %v run", outOb.GetName())

	safeGo(func(i ...interface{}) {
		var wg sync.WaitGroup
		for item := range o.outputC {
			wg.Add(1)
			safeGo(func(i ...interface{}) {
				defer wg.Done()
				if isError(i[0]) {
					outOb.SetNext(i[0], func(i interface{}) interface{} {
						return i
					})
					return
				}

				applyOb := fn(i[0])
				<-applyOb.Subscribe(NewObserverWithErrDone(
					func(i interface{}) {
						outOb.SetNext(i, func(i interface{}) interface{} {
							return i
						})
					},
					func(e error) {
						outOb.SetNext(e, func(i interface{}) interface{} {
							return i
						})
					},
					func() {}))
			}, item)
		}
		wg.Wait()
		outOb.Close()
	})
	return outOb
}

func (o *Observable) Distinct() IObservable {
	outOb := newObservable()
	outOb.SetName(o.name + "-Distinct")
	Debugf("ob %v run", outOb.GetName())

	set := make(map[interface{}]struct{})
	safeGo(func(i ...interface{}) {
		for item := range o.outputC {
			outOb.SetNext(item, func(i interface{}) interface{} {
				_, ok := set[item]
				if !ok {
					set[item] = struct{}{}
					return item
				}
				return new(Drop)
			})
		}
		outOb.Close()
	})

	return outOb
}

func (o *Observable) Filter(fc func(interface{}) bool) IObservable {
	outOb := newObservable()
	outOb.SetName(o.name + "-Filter")
	Debugf("ob %v run", outOb.GetName())
	safeGo(func(i ...interface{}) {
		for item := range o.outputC {
			outOb.SetNext(item, func(i interface{}) interface{} {
				future := NewFuture()
				safeGo(func(i ...interface{}) {
					item := i[0]
					ok := fc(item)
					if ok {
						future.SetResult(item)
					} else {
						future.SetResult(new(Drop))
					}
				}, i)
				return future
			})

		}
		outOb.Close()
	})

	return outOb
}

func (o *Observable) AsList() IObservable {
	outOb := newObservable()
	outOb.SetName(o.name + "-AsList")
	Debugf("ob %v run", outOb.GetName())

	go func() {
		ret := make([]interface{}, 0, 5)
		for item := range o.outputC {
			switch item.(type) {
			case error:
			default:
				ret = append(ret, item)
			}
		}

		outOb.SetNext(ret, func(i interface{}) interface{} {
			return i
		})
		outOb.Close()
	}()
	return outOb
}
