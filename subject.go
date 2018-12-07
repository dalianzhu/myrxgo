package observable

import (
	"log"
	"runtime"
)

type ReplaySubject struct {
	observers []*Observer
}

func NewReplaySubject() *ReplaySubject {
	r := new(ReplaySubject)
	r.observers = make([]*Observer, 0, 5)
	return r
}

func (r *ReplaySubject) OnNext(i interface{}) {
	for _, obs := range r.observers {
		go func(obs *Observer, i interface{}) {
			defer func() {
				if err := recover(); err != nil {
					stack := make([]byte, 1024*8)
					stack = stack[:runtime.Stack(stack, false)]

					f := "PANIC: %s\n%s"
					log.Printf(f, err, stack)
					obs.OnErr(i)
				}
			}()
			obs.OnNext(i)
		}(obs, i)
	}
}

func (r *ReplaySubject) Add(obs *Observer) {
	r.observers = append(r.observers, obs)
}
