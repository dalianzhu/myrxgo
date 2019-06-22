package myrxgo

import (
	"sync"
)

func (o *Observable) Map(fc func(interface{}) interface{}) *Observable {
	outOb := newObservable()
	outOb.Name = o.Name + "-Map"
	Debugf("ob %v run", outOb.Name)

	safeGo(func(i ...interface{}) {
		futures := make(chan *Future, 1000)
		safeGo(func(i ...interface{}) {
			for item := range o.C {
				future := NewFuture()
				safeGo(func(i ...interface{}) {
					switch v := i[0].(type) {
					case error:
						future.SetResult(v)
					default:
						ret := fc(v)
						future.SetResult(ret)
					}

				}, item)
				futures <- future
			}
			close(futures)
		})

		for future := range futures {
			outOb.C <- future.GetResult()
			outOb.OnStepFinish(future.GetResult())
		}
		outOb.close()
	})
	return outOb
}

func (o *Observable) FlatMap(fn func(interface{}) *Observable) *Observable {
	outOb := newObservable()
	outOb.Name = o.Name + "-FlatMap"
	Debugf("ob %v run", outOb.Name)

	safeGo(func(i ...interface{}) {
		var wg sync.WaitGroup
		for item := range o.C {
			wg.Add(1)
			safeGo(func(i ...interface{}) {
				defer wg.Done()
				applyOb := fn(i[0])
				applyOb.Run(
					func(i interface{}) {
						outOb.C <- i
						outOb.OnStepFinish(i)
					},
				)
			}, item)

		}
		wg.Wait()
		outOb.close()
	})
	return outOb
}

func (o *Observable) Distinct() *Observable {
	outOb := newObservable()
	outOb.Name = o.Name + "-Distinct"
	Debugf("ob %v run", outOb.Name)

	set := make(map[interface{}]struct{})
	safeGo(func(i ...interface{}) {
		for item := range o.C {
			switch v := item.(type) {
			case error:
				outOb.C <- v
			default:
				_, ok := set[v]
				if !ok {
					outOb.C <- v
					outOb.OnStepFinish(v)
					set[v] = struct{}{}
				}
			}
		}
		outOb.close()
	})

	return outOb
}

func (o *Observable) Filter(fc func(interface{}) bool) *Observable {
	outOb := newObservable()
	outOb.Name = o.Name + "-Filter"
	Debugf("ob %v run", outOb.Name)
	futures := make(chan *Future, 1000)
	safeGo(func(i ...interface{}) {
		safeGo(func(i ...interface{}) {
			for item := range o.C {
				future := NewFuture()
				safeGo(func(i ...interface{}) {
					switch v := i[0].(type) {
					case error:
						future.SetResult([]interface{}{
							true, v,
						})
					default:
						ok := fc(v)
						future.SetResult([]interface{}{
							ok, v,
						})
					}

				}, item)
				futures <- future
			}
			close(futures)
		})
		for future := range futures {
			v := future.GetResult().([]interface{})
			if v[0].(bool) {
				outOb.C <- v[1]
				outOb.OnStepFinish(v[1])
			}
		}
		outOb.close()
	})

	return outOb
}

func (o *Observable) AsList() *Observable {
	outOb := newObservable()
	outOb.Name = o.Name + "-AsList"
	Debugf("ob %v run", outOb.Name)

	go func() {
		ret := make([]interface{}, 0, 5)
		for item := range o.C {
			switch item.(type) {
			case error:
			default:
				ret = append(ret, item)
			}
		}

		outOb.C <- ret
		outOb.OnStepFinish(ret)
		outOb.close()
	}()
	return outOb
}
