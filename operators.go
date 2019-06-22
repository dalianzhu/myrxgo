package myrxgo

import (
	"sync"
)

func (o *Observable) Map(fc func(interface{}) interface{}) IObservable {
	outOb := newObservable()
	outOb.SetName(o.name + "-Map")
	Debugf("ob %v run", outOb.GetName())

	safeGo(func(i ...interface{}) {
		for item := range o.c {
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

func (o *Observable) FlatMap(fn func(interface{}) IObservable) IObservable {
	outOb := newObservable()
	outOb.SetName(o.name + "-FlatMap")
	Debugf("ob %v run", outOb.GetName())

	safeGo(func(i ...interface{}) {
		var wg sync.WaitGroup
		for item := range o.c {
			wg.Add(1)
			safeGo(func(i ...interface{}) {
				defer wg.Done()
				applyOb := fn(i[0])
				applyOb.Run(
					func(i interface{}) {
						outOb.SetNext(i, func(i interface{}) interface{} {
							return i
						})
					},
				)
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
		for item := range o.c {
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
		for item := range o.c {
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
		for item := range o.c {
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
