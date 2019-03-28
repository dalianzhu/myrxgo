package myrxgo

type IObserver interface {
	OnNext(interface{})
	OnErr(error)
}

type Observer struct {
	NextHandler func(interface{})
	ErrHandler  func(error)
	InnerID     string
}

func NewObserver(fn func(interface{})) *Observer {
	o := new(Observer)
	o.NextHandler = fn
	o.ErrHandler = func(i error) {}
	o.InnerID = UUID()[:8]
	return o
}

func NewAsyncObserver(fn func(interface{})) *Observer {
	o := new(Observer)
	o.NextHandler = func(i interface{}) {
		safeGo(func(i ...interface{}) {
			fn(i[0])
		}, i)
	}
	o.ErrHandler = func(i error) {}
	o.InnerID = UUID()[:8]
	return o
}
func (o *Observer) OnNext(i interface{}) {
	o.NextHandler(i)
}

func (o *Observer) OnErr(i error) {
	o.ErrHandler(i)
}

func (o Observer) ID() string {
	return o.InnerID
}
