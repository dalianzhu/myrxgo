package observable

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"
)

func TestMap(t *testing.T) {
	arr := []string{
		"hello",
		"my",
		"world",
		"be",
	}

	From(arr).
		Filter(func(i interface{}) bool {
			return len(i.(string)) > 3
		}).
		Map(func(i interface{}) interface{} {
			return i.(string) + "haha"
		}).
		Run(func(i interface{}) {
			fmt.Println(i)
		})
}

func TestFromChan(t *testing.T) {
	c := make(chan interface{})
	ob3 := FromChan(c).Map(
		func(i interface{}) interface{} {
			return i.(int) * 2
		},
	)

	i := 0
	go func() {
		for {
			i++
			if i > 3 {
				break
			}
			c <- i
			time.Sleep(time.Second)
		}
		close(c)
	}()
	go ob3.Run(func(i interface{}) {
		fmt.Println(i)
	})
	time.Sleep(time.Second * 4)
}

func TestClone(t *testing.T) {
	arr := []string{
		"hello1",
		"hello2",
	}

	var obForked Observable
	ob := From(arr).ClonePtr(&obForked)

	go ob.Run(func(i interface{}) {
		fmt.Println(i)
	})
	go obForked.Run(func(i interface{}) {
		fmt.Println(i)
	})
	time.Sleep(time.Second * 1)
}

func TestAsList(t *testing.T) {
	arr := []string{
		"hello",
		"my",
		"world",
		"be",
	}

	l := <-From(arr).
		Map(
			func(i interface{}) interface{} {
				return i.(string) + "_hi"
			},
		).AsList().C
	log.Println(l)
}

func TestSafeRun(t *testing.T) {
	arr := []interface{}{
		"hello",
		"my",
		"world",
		"be",
		123,
	}

	From(arr).
		Map(func(i interface{}) interface{} {
			return i.(string) + "_hi"
		}).
		AsList().
		Run(func(i interface{}) {
			panic("haha")
			log.Println(i)
		})
	time.Sleep(time.Second * 5)
}

func TestFlatMap(t *testing.T) {
	arr := []interface{}{
		"hello world",
		"my idea",
		"haha can",
	}

	From(arr).
		FlatMap(func(i interface{}) *Observable {
			ob := From(strings.Split(i.(string), " "))
			return ob
		}).
		Run(func(i interface{}) {
			log.Println("结果", i)
		})
}

func TestSubscribe(t *testing.T) {
	arr := []string{
		"hello",
		"my",
		"world",
		"be",
	}

	obs1 := NewObserver(func(i interface{}) {
		fmt.Println("obs1 ", i)
	})

	obs2 := NewObserver(func(i interface{}) {
		fmt.Println("obs2 ", i)
	})

	obs3 := NewObserver(func(i interface{}) {
		fmt.Println("obs3 ", i)
	})

	var ob2 Observable
	subject := NewReplaySubject()
	subject.Add(obs1)
	subject.Add(obs2)

	go From(arr).
		Filter(func(i interface{}) bool {
			return len(i.(string)) > 3
		}).
		Map(func(i interface{}) interface{} {
			return i.(string) + "haha"
		}).ClonePtr(&ob2).Subscribe(subject)

	go ob2.Map(func(i interface{}) interface{} {
		return i.(string) + " map2"
	}).Subscribe(obs3)

	time.Sleep(time.Second * 3)
}

func TestAll(t *testing.T) {
	TestMap(t)
	TestFromChan(t)
	TestClone(t)
	TestAsList(t)
	TestSafeRun(t)
	TestFlatMap(t)
	TestSubscribe(t)
}
