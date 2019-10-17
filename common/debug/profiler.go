package debug

import (
	"log"
	"sync"
	"time"
)

var gProfiler = &Profiler{
	PointTimeMap:    map[string]int64{},
	PointCountMap:   map[string]int{},
	AverageValueMap: map[string]int64{},
	AverageCountMap: map[string]int{},
}

func Start(name string) *Timer {
	return gProfiler.Start(name)
}

func Average(name string, num int64) {
	gProfiler.Average(name, num)
}

func Result() {
	gProfiler.Result()
}

type Profiler struct {
	sync.Mutex
	PointTimeMap    map[string]int64
	PointCountMap   map[string]int
	AverageValueMap map[string]int64
	AverageCountMap map[string]int
}

func (p *Profiler) Start(name string) *Timer {
	return &Timer{
		p:     p,
		Name:  name,
		Begin: time.Now().UnixNano(),
	}
}

func (p *Profiler) Average(name string, num int64) {
	p.Lock()
	defer p.Unlock()

	p.AverageValueMap[name] += num
	p.AverageCountMap[name]++
}

func (p *Profiler) Result() {
	p.Lock()
	defer p.Unlock()

	if len(p.PointTimeMap) > 0 {
		log.Println("[Point]")
		for name, t := range p.PointTimeMap {
			cnt := int64(p.PointCountMap[name])
			if cnt > 0 {
				log.Println(name, time.Duration(t), cnt, time.Duration(t/cnt))
			} else {
				log.Println(name, time.Duration(t), cnt)
			}
		}
		p.PointTimeMap = map[string]int64{}
		p.PointCountMap = map[string]int{}
	}

	if len(p.AverageValueMap) > 0 {
		log.Println("[Average]")
		for name, t := range p.AverageValueMap {
			cnt := int64(p.AverageCountMap[name])
			if cnt > 0 {
				log.Println(name, int64(t), cnt, int64(t/cnt))
			} else {
				log.Println(name, int64(t), cnt)
			}
		}
		p.AverageValueMap = map[string]int64{}
		p.AverageCountMap = map[string]int{}
	}
}

type Timer struct {
	p     *Profiler
	Name  string
	Begin int64
}

func (t *Timer) Stop() {
	now := time.Now().UnixNano()

	t.p.Lock()
	defer t.p.Unlock()

	t.p.PointTimeMap[t.Name] += now - t.Begin
	t.p.PointCountMap[t.Name]++
}
