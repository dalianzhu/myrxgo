package myrxgo

import (
	"go/types"
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

	Map(fc func(interface{}) interface{}) IObservable
	FlatMap(fn func(interface{}) IObservable) IObservable
	Distinct() IObservable
	Filter(fc func(interface{}) bool) IObservable
	AsList() IObservable

	Merge(inputObservable IObservable,
		fc func(interface{}, interface{}) interface{}) IObservable

	Subscribe(obs IObserver) chan int

	Run(fn func(i interface{}))
}

type Observable struct {
	c                 chan interface{}
	closeHandler      func()
	name              string
	stepFinishHandler func(interface{})
	lock              sync.Mutex
	buffer            chan interface{}
}

func (o *Observable) GetName() string {
	return o.name
}

func (o *Observable) Close() {
	safeGo(func(i ...interface{}) {
		Debugf("ob %v Close", o.name)
		time.Sleep(time.Microsecond * 10)
		close(o.buffer)
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
	case *Future:
		o.buffer <- nextData
	default:
		o.c <- nextData
		o.OnStepFinish(nextData)
	}
}

func newObservable() IObservable {
	o := new(Observable)
	o.c = make(chan interface{})
	o.SetCloseHandler(func() {})
	o.stepFinishHandler = func(i interface{}) {
	}
	o.buffer = make(chan interface{}, 10)

	go func() {
		for item := range o.buffer {
			Debugf("buffer %v %v", item, reflect.TypeOf(item))
			var nextData interface{}

			switch v := item.(type) {
			case *Future:
				nextData = v.GetResult()
			default:
				nextData = v
			}

			switch nextData.(type) {
			case Drop, *Drop:
			case types.Nil:
			default:
				o.c <- nextData
				o.OnStepFinish(nextData)
			}
		}
		close(o.c)
	}()

	return o
}

func (o *Observable) GetChan() chan interface{} {
	return o.c
}

func (o *Observable) SetName(name string) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.name = name
}

func From(obj interface{}) IObservable {
	outOb := newObservable()
	outOb.SetName(UUID()[:8])
	Debugf("ob %v run, From", outOb.GetName())
	safeGo(func(i ...interface{}) {
		val := reflect.ValueOf(obj)
		switch val.Kind() {
		case reflect.Slice:
			for i := 0; i < val.Len(); i++ {
				e := val.Index(i)
				outOb.SetNext(e.Interface(), func(i interface{}) interface{} {
					return i
				})
			}
		case reflect.Chan:
			for {
				v, ok := val.Recv()
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
		outOb.Close()
	})
	return outOb
}

func FromChan(c chan interface{}) IObservable {
	outOb := newObservable()
	outOb.SetName(UUID()[:8])
	Debugf("ob %v run, FromChan", outOb.GetName())
	safeGo(func(i ...interface{}) {
		for item := range c {
			outOb.SetNext(item, func(i interface{}) interface{} {
				return i
			})
		}
		outOb.Close()
	})
	return outOb
}

func (o *Observable) Merge(inputObservable IObservable,
	fc func(interface{}, interface{}) interface{}) IObservable {
	outOb := newObservable()
	outOb.SetName(o.name + "-Merge")
	Debugf("ob %v run", outOb.GetName())

	safeGo(func(i ...interface{}) {
		for item := range o.c {
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
	outOb := FromChan(inOutOb)
	outOb.SetName(source.GetName() + "-FromStream")

	sourceOnStepFinish := source.GetStepFinishHandler()
	source.SetStepFinishHandler(func(i interface{}) {
		sourceOnStepFinish(i)
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
	go func() {
		for item := range o.c {
			safeRun(func() {
				switch v := item.(type) {
				case error:
					obs.OnErr(v)
				default:
					obs.OnNext(v)
				}
			})
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
	for item := range o.c {
		safeRun(func() {
			switch v := item.(type) {
			case error:
				obs.OnErr(v)
			default:
				obs.OnNext(v)
			}
		})
	}
	Debugf("run %v %v", o.name, "exit")
}
