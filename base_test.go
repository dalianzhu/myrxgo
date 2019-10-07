package myrxgo

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"
	"time"
)

func Equal(t *testing.T, i1, i2 interface{}) {
	if i1 != i2 {
		log.Printf("Equal %v != %v", i1, i2)
		t.Fatalf("%v != %v", i1, i2)
	}
	log.Printf("i1 == i2 == %v", i1)
}

func IsIn(ret []string, find string) bool {
	for _, item := range ret {
		if item == find {
			return true
		}
	}
	return false
}

//go test -v myrxgo -test.run Testxxx
func TestMap(t *testing.T) {
	arr := []string{
		"hello",
		"my",
		"world",
		"be",
	}
	ret := <-From(arr).
		Filter(func(i interface{}) bool {
			return len(i.(string)) > 3
		}).
		Map(func(i interface{}) interface{} {
			return i.(string) + "haha"
		}).
		AsList().GetChan()
	retlist := ret.([]interface{})
	Equal(t, retlist[0], "hellohaha")
	Equal(t, retlist[1], "worldhaha")

	findErr := false
	obs := NewObserver(func(i interface{}) {
		log.Println("map result", i)
	})
	obs.ErrHandler = func(e error) {
		findErr = true
		log.Println(e)
	}

	<-From(arr).Map(func(i interface{}) interface{} {
		if i.(string) == "hello" {
			return errors.New("hello error")
		}
		return i
	}).Filter(func(i interface{}) bool {
		if len(i.(string)) > 3 {
			return true
		}
		return false
	}).Map(func(i interface{}) interface{} {
		log.Println(i.(string))
		return i
	}).Subscribe(obs)

	if findErr == false {
		t.Fail()
	}
}

func TestFromChan(t *testing.T) {
	c := make(chan interface{})
	ob3 := From(c).Map(
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
	ret := <-ob3.AsList().GetChan()
	retlist := ret.([]interface{})
	log.Println(retlist)
	Equal(t, retlist[0], 2)
	Equal(t, retlist[1], 4)
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
		).AsList().GetChan()
	log.Println(l)
	Equal(t, l.([]interface{})[1], "my_hi")
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
	time.Sleep(time.Second * 1)
}

func TestFlatMap(t *testing.T) {
	arr := []interface{}{
		"hello world",
		"my idea",
		"haha can",
	}

	ret := make([]string, 0)

	From(arr).
		FlatMap(func(i interface{}) IObservable {
			ob := From(strings.Split(i.(string), " "))
			return ob
		}).Run(func(i interface{}) {
		ret = append(ret, i.(string))
	})

	log.Printf("TestFlatMap ret %v", ret)
	Equal(t, len(ret), 6)

	Equal(t, true, IsIn(ret, "hello"))
	Equal(t, true, IsIn(ret, "world"))
	Equal(t, true, IsIn(ret, "my"))
	Equal(t, true, IsIn(ret, "idea"))
	Equal(t, true, IsIn(ret, "haha"))
	Equal(t, true, IsIn(ret, "can"))

}

func TestSubscribe(t *testing.T) {
	arr := []string{
		"hello",
		"my",
		"world",
		"be",
	}

	ret1 := make([]string, 0)
	ret2 := make([]string, 0)

	obs1 := NewObserver(func(i interface{}) {
		ret1 = append(ret1, i.(string))
		fmt.Println("obs1 ", i)
	})

	obs2 := NewObserver(func(i interface{}) {
		ret2 = append(ret2, i.(string))
		fmt.Println("obs2 ", i)
	})

	subject := NewSubject()
	subject.Subscribe(obs1)
	subject.Subscribe(obs2)

	go From(arr).
		Filter(func(i interface{}) bool {
			return len(i.(string)) > 3
		}).Subscribe(subject)

	time.Sleep(time.Second * 1)
	log.Printf("TestSubscribe ret1 %v", ret1)
	log.Printf("TestSubscribe ret2 %v", ret2)

	Equal(t, len(ret1), 2)
	Equal(t, len(ret2), 2)

	Equal(t, true, IsIn(ret1, "hello"))
	Equal(t, true, IsIn(ret2, "hello"))
	Equal(t, true, IsIn(ret1, "world"))
	Equal(t, true, IsIn(ret2, "world"))

}

func TestDistinct(t *testing.T) {
	arr := []string{
		"hello",
		"my",
		"hello",
		"world",
		"be",
	}
	ret := <-From(arr).Distinct().AsList().GetChan()
	retlist := ret.([]interface{})
	log.Println(retlist)
	Equal(t, retlist[2], "world")
}

func TestFromStream(t *testing.T) {
	arr := []string{
		"hello",
		"world",
	}
	var ret1 []string
	var ret2 []string

	mainStream := From(arr).
		Map(func(i interface{}) interface{} {
			return i.(string) + ":"
		})

	subStream1 := FromStream(mainStream).
		Map(func(i interface{}) interface{} {
			return i.(string) + ":sub1"
		})

	subStream2 := FromStream(mainStream).
		Map(func(i interface{}) interface{} {
			return i.(string) + ":sub2"
		})

	go mainStream.Run(func(i interface{}) {
		fmt.Println("main ", i)
	})

	var wait sync.WaitGroup

	wait.Add(1)
	go func() {
		subStream1.Run(func(i interface{}) {
			fmt.Println("sub1", i)
			ret1 = append(ret1, i.(string))
		})
		wait.Done()
	}()

	wait.Add(1)
	go func() {
		subStream2.Run(func(i interface{}) {
			fmt.Println("sub2", i)
			ret2 = append(ret2, i.(string))
		})
		wait.Done()
	}()

	wait.Wait()

	Equal(t, len(ret1), 2)
	Equal(t, len(ret2), 2)
}

func TestSubject(t *testing.T) {
	subject := NewSubject()

	var ret1 []interface{}
	var ret2 []interface{}
	var ret3 []interface{}

	subject.Subscribe(NewObserver(func(i interface{}) {
		fmt.Println("send data 1 ", i)
		ret1 = append(ret1, i)
	}))
	subject.Subscribe(NewObserver(func(i interface{}) {
		fmt.Println("send data 2 ", i)
		ret2 = append(ret2, i)
	}))

	ch := make(chan interface{})
	go func() {
		FromChan(ch).Subscribe(subject)
	}()

	loop := 0
	for {
		fmt.Println("loop", loop)
		ch <- time.Now().Unix()
		time.Sleep(time.Microsecond * 400)

		if loop == 2 {
			first := subject.List()[0]
			subject.Unsubscribe(first)
		}

		if loop == 3 {
			subject.Subscribe(NewObserver(func(i interface{}) {
				fmt.Println("send data 3 ", i)
				ret3 = append(ret3, i)
			}))
		}

		if loop == 4 {
			time.Sleep(time.Microsecond * 200)
			break
		}

		fmt.Println(subject.List())
		loop += 1
		time.Sleep(time.Microsecond * 100)
	}
	fmt.Println(ret1)
	Equal(t, len(ret1), 3)
	fmt.Println(ret2)
	Equal(t, len(ret2), 5)
	fmt.Println(ret3)
	Equal(t, len(ret3), 1)
}

func TestNewAsyncObserver(t *testing.T) {
	ch := make(chan interface{})
	start := time.Now()
	loop := 0
	go func() {
		for {
			ch <- time.Now().Unix()
			if loop == 20 {
				time.Sleep(time.Millisecond * 500)
				close(ch)
				break
			}
			loop += 1
		}
	}()

	<-FromChan(ch).Subscribe(NewAsyncObserver(
		func(i interface{}) {
			time.Sleep(time.Millisecond * 100)
			fmt.Println(i)
		}))
	end := time.Since(start)
	fmt.Println("costs time ", end)
	if end > time.Millisecond*600 {
		t.Fatal("costs time too much")
	}
}

func TestAsyncMap(t *testing.T) {
	// map中的操作是并行的
	var arr = []int{1, 2, 3, 4, 5}

	start := time.Now()
	From(arr).Map(func(i interface{}) interface{} {
		time.Sleep(time.Second)
		return i
	}).Run(func(i interface{}) {
		log.Printf("TestAsyncMap %v", i)
	})

	end := time.Since(start)

	if end > time.Second*2 {
		t.Fail()
	}
}

func TestDone(t *testing.T) {
	var source = []interface{}{
		"1 2", "3 4", errors.New("hello"), "5 6",
	}
	var ret []string
	f := From(source).Map(func(i interface{}) interface{} {
		item := i.(string)
		item += " 9"
		fmt.Println("call map", item)
		return item
	}).FlatMapPara(func(i interface{}) IObservable {
		item := i.(string)
		fmt.Println("item", i)
		return From(strings.Split(item, " "))
	}).Subscribe(NewObserverWithErrDone(func(i interface{}) {
		fmt.Println("TestDone", i)
		item := i.(string)
		ret = append(ret, item)
	}, func(e error) { fmt.Println("find onError", e) },
		func() {}))
	<-f
	Equal(t, 6, len(ret))
	Equal(t, true, IsIn(ret, "1"))
	Equal(t, true, IsIn(ret, "2"))
	Equal(t, true, IsIn(ret, "3"))
	Equal(t, false, IsIn(ret, "5"))
}
