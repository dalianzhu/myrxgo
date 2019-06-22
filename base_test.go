package myrxgo

import (
	"errors"
	"fmt"
	"log"
	"strings"
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
		AsList().C
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
	ret := <-ob3.AsList().C
	retlist := ret.([]interface{})
	log.Println(retlist)
	Equal(t, retlist[0], 2)
	Equal(t, retlist[1], 4)
}

func TestClone(t *testing.T) {
	arr := []string{
		"hello1",
		"hello2",
	}

	var obForked Observable
	ob := From(arr).ClonePtr(&obForked)

	var ret1 interface{}
	var ret2 interface{}
	go func() {
		ret1 = <-ob.AsList().C
		log.Println(ret1)
	}()
	go func() {
		ret2 = <-obForked.AsList().C
	}()
	time.Sleep(time.Second * 4)
	log.Println(ret1, ret2)
	ret1list := ret1.([]interface{})
	ret2list := ret2.([]interface{})
	log.Println(ret1list, ret2list)
	Equal(t, ret2list[0], ret1list[0])
	Equal(t, ret2list[1], ret1list[1])
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

	ret := <-From(arr).
		FlatMap(func(i interface{}) *Observable {
			ob := From(strings.Split(i.(string), " "))
			return ob
		}).AsList().C
	retlist := ret.([]interface{})
	log.Println(retlist)
	func() {
		for _, item := range retlist {
			if item == "idea" {
				return
			}
		}
		t.Fatal("TestFlatMap cant find 'idea'")
	}()
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
	subject := NewSubject()
	subject.Subscribe(obs1)
	subject.Subscribe(obs2)

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

func TestDistinct(t *testing.T) {
	arr := []string{
		"hello",
		"my",
		"hello",
		"world",
		"be",
	}
	ret := <-From(arr).Distinct().AsList().C
	retlist := ret.([]interface{})
	log.Println(retlist)
	Equal(t, retlist[2], "world")
}

func TestFromStream(t *testing.T) {
	arr := []string{
		"hello",
		"world",
	}
	var ret []string

	mainStream := From(arr).
		Map(func(i interface{}) interface{} {
			return i.(string) + "haha"
		})

	subStream := FromStream(mainStream).
		Map(func(i interface{}) interface{} {
			return i.(string) + "sub"
		})

	go mainStream.Run(func(i interface{}) {
		fmt.Println("main ", i)
	})

	go subStream.Run(func(i interface{}) {
		fmt.Println("sub ", i)
		ret = append(ret, i.(string))
	})

	time.Sleep(time.Second)
	Equal(t, len(ret), 2)
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
