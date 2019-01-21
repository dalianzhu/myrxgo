package myrxgo

import (
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
	time.Sleep(time.Second * 5)
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
	Equal(t, ret[0], "hellohahasub")
}

func TestAll(t *testing.T) {
	//go test -v myrxgo -test.run TestAll
	log.Println("TestMap")
	TestMap(t)

	log.Println("TestFromChan")
	TestFromChan(t)

	log.Println("TestClone")
	TestClone(t)

	log.Println("TestAsList")
	TestAsList(t)
	//TestSafeRun(t)

	log.Println("TestFlatMap")
	TestFlatMap(t)

	log.Println("TestSubscribe")
	TestSubscribe(t)

	log.Println("TestDistinct")
	TestDistinct(t)

	log.Println("TestFromStream")
	TestFromStream(t)
}
