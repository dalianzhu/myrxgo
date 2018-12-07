package observable

func NewHandler(fn func(i interface{})) *Observer {
	obs := new(Observer)
	obs.NextHandler = fn
	obs.ErrHandler = func(i interface{}) {}
	return obs
}
