package myrxgo

import (
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
		safeGo(func(i ...interface{}) {
			i[0].(*Observer).OnNext(i[1])
		}, obs, i)
	}
}

func (s *Subject) OnErr(err error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	for _, obs := range s.observers {
		safeGo(func(i ...interface{}) {
			i[0].(*Observer).OnErr(i[1].(error))
		}, obs, err)
	}
}

func (s *Subject) Subscribe(obs *Observer) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.observers = append(s.observers, obs)
}

func (s *Subject) Unsubscribe(id string) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

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
