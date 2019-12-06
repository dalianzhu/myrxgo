package myrxgo

import (
	"context"
	"reflect"
	"sync"
	"time"
)

type Drop struct {
}

type IObservable interface {
	GetChan() chan interface{}

	SetNext(i interface{}, f func(interface{}) interface{})

	Close()
	OnClose()
	SetCloseHandler(func())
	GetCloseHandler() func()

	SetName(string)
	GetName() string

	OnStepFinish(interface{})
	SetStepFinishHandler(func(interface{}))
	GetStepFinishHandler() func(interface{})

	Map(fc func(interface{}) interface{}, configs ...interface{}) IObservable
	FlatMap(fn func(interface{}) IObservable, configs ...interface{}) IObservable
	// FlatMapSerial(fn func(interface{}) IObservable) IObservable
	Distinct() IObservable
	Filter(fc func(interface{}) bool) IObservable
	AsList() IObservable

	Merge(inputObservable IObservable,
		fc func(interface{}, interface{}) interface{}) IObservable

	Subscribe(obs IObserver) chan int

	Run(fn func(i interface{}))
	//GetBaseContext() (context.Context, context.CancelFunc)
}

type Observable struct {
	outputC chan interface{}
	//inputC     chan interface{}
	baseCtx    context.Context
	baseCancel context.CancelFunc

	closeHandler      func()
	name              string
	stepFinishHandler func(interface{})
	lock              sync.Mutex
}

func (o *Observable) GetBaseContext() (context.Context, context.CancelFunc) {
	return o.baseCtx, o.baseCancel
}

func (o *Observable) GetName() string {
	return o.name
}

func (o *Observable) Close() {
	safeGo(func(i ...interface{}) {
		Debugf("ob %v Close", o.name)
		time.Sleep(time.Microsecond * 10)
		close(o.outputC)
		o.closeHandler()
	})
}

func (o *Observable) OnClose() {
	o.closeHandler()
}

func (o *Observable) SetCloseHandler(f func()) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.closeHandler = f
}

func (o *Observable) GetCloseHandler() func() {
	o.lock.Lock()
	defer o.lock.Unlock()
	return o.closeHandler
}

func (o *Observable) OnStepFinish(i interface{}) {
	//Debugf("OnStepFinish %v", i)
	o.stepFinishHandler(i)
}

func (o *Observable) SetStepFinishHandler(f func(interface{})) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.stepFinishHandler = f
}

func (o *Observable) GetStepFinishHandler() func(interface{}) {
	o.lock.Lock()
	defer o.lock.Unlock()
	return o.stepFinishHandler
}

func (o *Observable) SetNext(i interface{}, f func(interface{}) interface{}) {
	var nextData interface{}

	//Debugf("setNext:%v", i)
	switch i.(type) {
	case error:
		nextData = i
	default:
		nextData = f(i)
	}

	switch nextData.(type) {
	case *Drop, Drop:
		Debugf("SetNext find drop")
		return
	default:
		//Debugf("setNext:%v,%v", nextData,reflect.TypeOf(nextData))
		o.outputC <- nextData
		o.OnStepFinish(nextData)
	}
}

func newObservable(ctx context.Context,
	c context.CancelFunc) IObservable {

	o := new(Observable)
	o.baseCtx = ctx
	o.baseCancel = c
	o.outputC = make(chan interface{})
	//o.inputC = make(chan interface{}, concurrent)
	o.SetCloseHandler(func() {})
	o.stepFinishHandler = func(i interface{}) {
	}

	//go func() {
	//loop:
	//	for {
	//		var ok bool
	//		var item interface{}
	//		select {
	//		case item, ok = <-o.inputC:
	//		case <-o.baseCtx.Done():
	//			break loop
	//		}
	//
	//		if !ok {
	//			break
	//		}
	//
	//		//Debugf("inputC %v %v", item, reflect.TypeOf(item))
	//		var nextData interface{}
	//		var timeout bool
	//		switch v := item.(type) {
	//		case *Future:
	//			nextData, timeout = v.GetResult()
	//			if timeout {
	//				nextData = errors.New("myrxgo future timeout")
	//			}
	//		default:
	//			nextData = v
	//		}
	//
	//		switch nextData.(type) {
	//		case Drop, *Drop:
	//		default:
	//			//Debugf("setoutput %v", nextData)
	//			o.outputC <- nextData
	//			o.OnStepFinish(nextData)
	//		}
	//	}
	//	Debugf("newObservable will break outputc")
	//	close(o.outputC)
	//}()

	return o
}

func (o *Observable) GetChan() chan interface{} {
	return o.outputC
}

func (o *Observable) SetName(name string) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.name = name
}

func From(obj interface{}) IObservable {
	ctx, cancel := context.WithCancel(context.Background())
	outOb := newObservable(ctx, cancel)
	outOb.SetName(UUID()[:8])
	Debugf("ob %v run, From", outOb.GetName())
	safeGo(func(i ...interface{}) {
		val := reflect.ValueOf(obj)
		switch val.Kind() {
		case reflect.Slice:
		loop:
			for i := 0; i < val.Len(); i++ {
				select {
				case <-ctx.Done():
					break loop
				default:
					e := val.Index(i)
					outOb.SetNext(e.Interface(), func(i interface{}) interface{} {
						return i
					})
				}
			}
		case reflect.Chan:
			cases := make([]reflect.SelectCase, 2)
			cases[0] = reflect.SelectCase{Dir: reflect.SelectRecv,
				Chan: val}
			cases[1] = reflect.SelectCase{Dir: reflect.SelectRecv,
				Chan: reflect.ValueOf(ctx.Done())}

			for {
				chosen, v, ok := reflect.Select(cases)
				if chosen == 1 {
					break
				}
				//v, ok := val.Recv()
				if !ok {
					break
				}
				outOb.SetNext(v.Interface(), func(i interface{}) interface{} {
					return i
				})
			}
		default:
			outOb.SetNext(obj, func(i interface{}) interface{} {
				return i
			})
		}
		Debugf("From will close")
		outOb.Close()
	})
	return outOb
}

func (o *Observable) Merge(inputObservable IObservable,
	fc func(interface{}, interface{}) interface{}) IObservable {
	outOb := newObservable(o.baseCtx, o.baseCancel)
	outOb.SetName(o.name + "-Merge")
	Debugf("ob %v run", outOb.GetName())

	safeGo(func(i ...interface{}) {
		for item := range o.outputC {
			ifItem, ok := <-inputObservable.GetChan()
			if !ok {
				break
			}
			safeRun(func() {
				outOb.SetNext(item, func(i interface{}) interface{} {
					return fc(i, ifItem)
				})
			})
		}
		outOb.Close()
	})
	return outOb
}

func FromStream(source IObservable) IObservable {
	inOutOb := make(chan interface{})
	outOb := From(inOutOb)
	outOb.SetName(source.GetName() + "-FromStream")
	Debugf("ob %v run, FromStream", outOb.GetName())

	sourceOnStepFinish := source.GetStepFinishHandler()
	source.SetStepFinishHandler(func(i interface{}) {
		sourceOnStepFinish(i)
		Debugf("set inOutOb: %v, source:%v", i, source)
		inOutOb <- i
	})

	sourceClose := source.GetCloseHandler()
	source.SetCloseHandler(func() {
		sourceClose()
		close(inOutOb)
	})
	return outOb
}

func (o *Observable) Subscribe(obs IObserver) chan int {
	Debugf("run %v %v", o.name, "start")
	fin := make(chan int, 1)
	findErr := false
	go func() {
		for item := range o.outputC {
			if findErr {
				//Debugf("Subscribe find err:continue:%v", "")
				continue
			}
			switch v := item.(type) {
			case error:
				obs.OnErr(v)
				findErr = true
				//Debugf("Subscribe:find err will cancel %v:%v", time.Now().UnixNano(), v)
				o.baseCancel()
			default:
				Try(func() {
					obs.OnNext(v)
				}, func(rec interface{}, stack []byte) {
					obs.OnErr(SystemPanicHandler(rec, stack))
					findErr = true
					o.baseCancel()
				})
			}
		}

		if !findErr {
			obs.OnDone()
		}

		Debugf("run %v %v", o.name, "exit")
		fin <- 1
	}()
	return fin
}

func (o *Observable) Run(fn func(i interface{})) {
	Debugf("run %v %v", o.name, "start")
	obs := NewObserver(fn)
	obs.ErrHandler = func(e error) {
		Debugf("Run error %v", e)
	}
	findErr := false
	for item := range o.outputC {
		if findErr {
			continue
		}
		switch v := item.(type) {
		case error:
			obs.OnErr(v)
			findErr = true
			o.baseCancel()
		default:
			Try(func() {
				obs.OnNext(v)
			}, func(rec interface{}, stack []byte) {
				obs.OnErr(SystemPanicHandler(rec, stack))
			})
		}
	}

	if !findErr {
		obs.OnDone()
	}

	Debugf("run %v %v", o.name, "exit")
}
