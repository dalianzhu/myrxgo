package observable

type IOnNext interface {
	OnNext(interface{})
}

type Observer struct {
	NextHandler func(interface{})
	ErrHandler  func(interface{})
}

func NewObserver(fn func(interface{})) *Observer {
	o := new(Observer)
	o.NextHandler = fn
	o.ErrHandler = func(i interface{}) {}
	return o
}

func (o *Observer) OnNext(i interface{}) {
	o.NextHandler(i)
}

func (o *Observer) OnErr(i interface{}) {
	o.ErrHandler(i)
}
