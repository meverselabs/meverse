package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/meverselabs/meverse/common/queue"
)

type testdq struct {
	p1 uint64
	p2 uint64
	k1 string
	k2 string
}

func (t *testdq) Key() interface{} {
	return t.k1
}
func (t *testdq) DKey() interface{} {
	return t.k2
}
func (t *testdq) Priority() uint64 {
	return t.p1
}
func (t *testdq) DPriority() uint64 {
	return t.p2
}
func (t *testdq) Print(dq queue.IDoubleKeyQueue) {
	log.Println("p1:", t.p1, "p2:", t.p2, "k1:", t.k1, "k2:", t.k2, "size:", dq.Size())
}

func main() {
	dq := queue.NewDoubleKeyPriorityQueue(queue.LOWEST)

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	{
		im := 9
		jm := 4

		for i := 0; i < im; i++ {
			time := r1.Int63()
			for j := 0; j < jm; j++ {
				p := r1.Int63()
				dq.Push(&testdq{uint64(p), uint64(time), fmt.Sprintf("k%v", uint64(i*jm+j)), fmt.Sprintf("%v", time)})
				log.Println(time, p, dq.Size())
			}
		}
	}
	dq.Pop().(*testdq).Print(dq)
	dq.Pop().(*testdq).Print(dq)
	dq.Pop().(*testdq).Print(dq)
	dq.Pop().(*testdq).Print(dq)
	dq.Pop().(*testdq).Print(dq)
	dq.Pop().(*testdq).Print(dq)
	dq.Pop().(*testdq).Print(dq)
	log.Println("asize", dq.Size())
	dq.Peek().(*testdq).Print(dq)

	{
		im := 2
		jm := 10

		for i := 0; i < im; i++ {
			time := r1.Int63()
			for j := 0; j < jm; j++ {
				p := r1.Int63()
				dq.Push(&testdq{uint64(p), uint64(time), fmt.Sprintf("%v", uint64(i*jm+j)), fmt.Sprintf("%v", time)})
				log.Println(time, p, dq.Size())
			}
		}
	}

	dq.Iter(func(td queue.IDoubleKeyQueueItem) bool {
		td.(*testdq).Print(dq)
		return false
	})
	log.Println("size end", dq.Size())
	dq.Iter(func(td queue.IDoubleKeyQueueItem) bool {
		td.(*testdq).Print(dq)
		return false
	})
	log.Println("size end", dq.Size())
	dq.Iter(func(td queue.IDoubleKeyQueueItem) bool {
		td.(*testdq).Print(dq)
		return true
	})
	log.Println("size end", dq.Size())
}
