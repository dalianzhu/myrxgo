package myrxgo

import (
	"log"
	"runtime"
	"sync"
)

type Subject struct {
	observers []*Observer
	sync.Mutex
}

func NewSubject() *Subject {
	r := new(Subject)
	r.observers = make([]*Observer, 0, 5)
	return r
}

func (s *Subject) OnNext(i interface{}) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	for _, obs := range s.observers {
		go func(obs *Observer, i interface{}) {
			defer func() {
				if err := recover(); err != nil {
					stack := make([]byte, 1024*8)
					stack = stack[:runtime.Stack(stack, false)]

					f := "PANIC: %s\n%s"
					log.Printf(f, err, stack)
				}
			}()
			obs.OnNext(i)
		}(obs, i)
	}
}

func (s *Subject) OnErr(i error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	for _, obs := range s.observers {
		go func(obs *Observer, i error) {
			defer func() {
				if err := recover(); err != nil {
					stack := make([]byte, 1024*8)
					stack = stack[:runtime.Stack(stack, false)]

					f := "PANIC: %s\n%s"
					log.Printf(f, err, stack)
				}
			}()
			obs.OnErr(i)
		}(obs, i)
	}
}

func (s *Subject) Subscribe(obs *Observer) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.observers = append(s.observers, obs)
}

func (s *Subject) Unsubscribe(id string) {
	i := func() int {
		for i, item := range s.observers {
			tpID := item.ID()
			if tpID == id {
				return i
			}
		}
		return -1
	}()

	s.observers = append(s.observers[:i], s.observers[i+1:]...)
}

func (s Subject) List() []string {
	var ret []string
	for _, item := range s.observers {
		ret = append(ret, item.ID())
	}
	return ret
}
