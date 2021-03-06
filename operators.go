package myrxgo

import (
	"errors"
	"sync"
)

/*
所有的error都被直接传递，不引发任何操作，setNext函数将做出操作
*/

func isError(i interface{}) bool {
	_, ok := i.(error)
	return ok
}

func (o *Observable) Map(fc func(interface{}) interface{}, configs ...interface{}) IObservable {
	isSerial := false
	var con uint = 10
	var timeout uint = 60 * 5

	for _, config := range configs {
		switch v := config.(type) {
		case *serialConfig:
			isSerial = true
		case *concurrentConfig:
			con = v.value
		case *timeoutConfig:
			timeout = v.value
		}
	}
	outOb := newObservable(o.baseCtx, o.baseCancel)
	outOb.SetName(o.name + "-Map")
	Debugf("ob %v run", outOb.GetName())
	if isSerial {
		con = 0
	}

	if con >= 1 {
		con -= 1
	}

	buf := make(chan *Future, con)
	go func() {
	loop:
		for {
			var item interface{}
			var ok bool
			select {
			case item, ok = <-o.outputC:
			case <-o.baseCtx.Done():
				break loop
			}
			if !ok {
				break
			}

			future := NewFuture()
			future.Timeout = timeout
			buf <- future

			switch item.(type) {
			case error:
				future.SetResult(item)
			default:
				go Try(func() {
					future.SetResult(fc(item))
				}, func(rec interface{}, stack []byte) {
					future.SetResult(SystemPanicHandler(rec, stack))
				})
			}
		}
		close(buf)
	}()
	go func() {
		for future := range buf {
			ret, timeout := future.GetResult()
			if timeout {
				ret = errors.New("myrxgo future timeout")
			}
			//Debugf("Map future:%v", ret)
			outOb.SetNext(ret, func(i interface{}) interface{} {
				return i
			})
		}
		outOb.Close()
	}()

	return outOb
}

func (o *Observable) FlatMap(fn func(interface{}) IObservable, configs ...interface{}) IObservable {
	isSerial := false
	for _, config := range configs {
		switch config.(type) {
		case *serialConfig:
			isSerial = true
		}
	}
	outOb := newObservable(o.baseCtx, o.baseCancel)
	outOb.SetName(o.name + "-FlatMap")
	Debugf("ob %v run", outOb.GetName())

	go func() {
		var wg sync.WaitGroup
		for item := range o.outputC {
			wg.Add(1)
			tpitem := item
			tpFunc := func() {
				defer wg.Done()
				if isError(tpitem) {
					outOb.SetNext(tpitem, func(i interface{}) interface{} {
						return i
					})
					return
				}
				Try(func() {
					applyOb := fn(tpitem)
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
				}, func(rec interface{}, stack []byte) {
					e := SystemPanicHandler(rec, stack)
					outOb.SetNext(e, func(i interface{}) interface{} {
						return i
					})
				})
			}
			if isSerial {
				tpFunc()
			} else {
				go tpFunc()
			}
		}
		wg.Wait()
		outOb.Close()
	}()
	return outOb
}

func (o *Observable) Distinct() IObservable {
	outOb := newObservable(o.baseCtx, o.baseCancel)
	outOb.SetName(o.name + "-Distinct")
	Debugf("ob %v run", outOb.GetName())

	set := make(map[interface{}]struct{})
	go func(i ...interface{}) {
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
	}()

	return outOb
}

// Filter return true可以通过
func (o *Observable) Filter(fc func(interface{}) bool) IObservable {
	outOb := newObservable(o.baseCtx, o.baseCancel)
	outOb.SetName(o.name + "-Filter")
	Debugf("ob %v run", outOb.GetName())
	go func(i ...interface{}) {
		for item := range o.outputC {
			outOb.SetNext(item, func(i interface{}) interface{} {
				var ok bool
				var ret interface{}

				Try(func() {
					ok = fc(item)
					if ok {
						ret = item
					} else {
						ret = new(Drop)
					}
				}, func(rec interface{}, stack []byte) {
					ret = SystemPanicHandler(rec, stack)
				})
				return ret
			})
		}
		outOb.Close()
	}()

	return outOb
}

func (o *Observable) AsList() IObservable {
	outOb := newObservable(o.baseCtx, o.baseCancel)
	outOb.SetName(o.name + "-AsList")
	Debugf("ob %v run", outOb.GetName())

	go func() {
		ret := make([]interface{}, 0, 5)
		findErr := false
		for item := range o.outputC {
			if findErr {
				continue
			}
			switch v := item.(type) {
			case error:
				findErr = true
				outOb.SetNext(v, func(i interface{}) interface{} {
					return i
				})
			default:
				ret = append(ret, item)
			}
		}

		if findErr == false {
			outOb.SetNext(ret, func(i interface{}) interface{} {
				return i
			})
		}
		outOb.Close()
	}()
	return outOb
}
