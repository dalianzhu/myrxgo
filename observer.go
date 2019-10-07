package myrxgo

type IObserver interface {
	OnNext(interface{})
	OnErr(error)
	OnDone()
}

type Observer struct {
	NextHandler func(interface{})
	ErrHandler  func(error)
	DoneHandler func()
	InnerID     string
}

func NewObserverWithErrDone(onNext func(interface{}),
	onErr func(error),
	onDone func()) *Observer {
	o := new(Observer)
	o.NextHandler = onNext
	o.ErrHandler = onErr
	o.DoneHandler = onDone
	o.InnerID = UUID()[:8]
	return o
}
func NewObserverWithErr(onNext func(interface{}), onErr func(error)) *Observer {
	o := new(Observer)
	o.NextHandler = onNext
	o.ErrHandler = onErr
	o.DoneHandler = func() {}
	o.InnerID = UUID()[:8]
	return o
}

func NewObserver(onNext func(interface{})) *Observer {
	o := new(Observer)
	o.NextHandler = onNext
	o.ErrHandler = func(i error) {}
	o.DoneHandler = func() {}
	o.InnerID = UUID()[:8]
	return o
}

func NewAsyncObserver(onNext func(interface{})) *Observer {
	o := new(Observer)
	o.NextHandler = func(i interface{}) {
		safeGo(func(i ...interface{}) {
			onNext(i[0])
		}, i)
	}
	o.ErrHandler = func(i error) {}
	o.DoneHandler = func() {}
	o.InnerID = UUID()[:8]

	return o
}
func (o *Observer) OnNext(i interface{}) {
	o.NextHandler(i)
}

func (o *Observer) OnErr(i error) {
	o.ErrHandler(i)
}

func (o *Observer) OnDone() {
	o.DoneHandler()
}

func (o Observer) ID() string {
	return o.InnerID
}
