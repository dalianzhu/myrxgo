package myrxgo

import (
	"sync"
	"time"
)

/*
Future 是一个未来的任务的抽象。和python里的那个有点类似。
在异步任务中SetResult，在GetResult的时候会等待result生成，或者超时。
使用姿势：
tasks := make([]*utils.Future, 0)
for i := 0; i < 10; i++ {
    future := utils.NewFuture()
    tasks = append(tasks, future)
    go func(result int) {
        time.Sleep(time.Second * time.Duration(rand.Int63n(10)))
        future.SetResult(result)
    }(i)
}

for _, item := range tasks {
    ret, ok := item.GetResult().(int)
    if ok {
        fmt.Println(ret)
    } else {
        fmt.Println("failed")
    }
}
*/

type Future struct {
	isfinished bool
	result     interface{}
	isTimeOut  bool
	resultchan chan interface{}
	l          sync.Mutex
	Timeout    uint
}

func (f *Future) GetResult() (interface{}, bool) {
	f.l.Lock()
	defer f.l.Unlock()
	if f.isfinished {
		return f.result, f.isTimeOut
	}

	select {
	// timeout
	case <-time.After(time.Second * time.Duration(f.Timeout)):
		f.isfinished = true
		f.result = nil
		f.isTimeOut = true
		Errorf("GetResult: timeout:%v", "")
		return nil, f.isTimeOut
	case f.result = <-f.resultchan:
		f.isfinished = true
		f.isTimeOut = false
		return f.result, f.isTimeOut
	}
}

func (f *Future) SetResult(result interface{}) {
	if f.isfinished == true {
		return
	}
	f.resultchan <- result
	close(f.resultchan)
}

func NewFuture() *Future {
	return &Future{
		isfinished: false,
		result:     nil,
		resultchan: make(chan interface{}, 1),
		isTimeOut:  false,
		Timeout:    60 * 5,
	}
}
