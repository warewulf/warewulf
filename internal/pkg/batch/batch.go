
package batch

import (
	"time"
	"sync"
)

type BatchPool struct {
	active int
	mintime	int
	jobcount int
	jobs []func()
}

func New(active int, mintime int) *BatchPool {
	pool := &BatchPool{
		active: active,
		mintime: mintime,
	}
	pool.jobs = []func(){}
	return pool
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func (pool *BatchPool) Submit(f func()) {
	pool.jobcount++
	pool.jobs = append(pool.jobs, f)
}

func wrapper(wg *sync.WaitGroup, f func()) {
	defer wg.Done()
	f()
}

func (pool *BatchPool) Run() {
	var wg sync.WaitGroup
	count := pool.jobcount
	for count > 0 {
		for i:=0; i<Min(count, pool.active); i++ {
			wg.Add(1)
			go wrapper(&wg, pool.jobs[pool.jobcount-count])
			count--
		}
		if (pool.mintime > 0) {
			time.Sleep(time.Second * time.Duration(pool.mintime))
		}
		wg.Wait()
	}
}

