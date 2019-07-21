package debug

import (
	"log"
	"time"
)

var gProfiler = &Profiler{
	PointTimeMap:  map[string]int64{},
	PointCountMap: map[string]int{},
}

func Start(name string) *Timer {
	return gProfiler.Start(name)
}

func Result() {
	gProfiler.Result()
}

type Profiler struct {
	PointTimeMap  map[string]int64
	PointCountMap map[string]int
}

func (p *Profiler) Start(name string) *Timer {
	return &Timer{
		p:     p,
		Name:  name,
		Begin: time.Now().UnixNano(),
	}
}

func (p *Profiler) Result() {
	for name, t := range p.PointTimeMap {
		log.Println(name, time.Duration(t))
	}
	p.PointTimeMap = map[string]int64{}
	p.PointCountMap = map[string]int{}
}

type Timer struct {
	p     *Profiler
	Name  string
	Begin int64
}

func (t *Timer) Stop() {
	t.p.PointTimeMap[t.Name] += time.Now().UnixNano() - t.Begin
	t.p.PointCountMap[t.Name]++
}
