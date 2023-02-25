package batch

import (
	"sync"
)

type BatchPool struct {
	active int
	jobs   []func()
}

func New(active int) *BatchPool {
	pool := &BatchPool{
		active: active,
	}
	pool.jobs = []func(){}
	return pool
}

func (pool *BatchPool) Submit(f func()) {
	pool.jobs = append(pool.jobs, f)
}

func wrapper(wg *sync.WaitGroup, limiter chan struct{}, f func()) {
	f()
	<-limiter
	wg.Done()
}

func (pool *BatchPool) Run() {
	var wg sync.WaitGroup
	limiter := make(chan struct{}, pool.active)
	for i := range pool.jobs {
		limiter <- struct{}{}
		wg.Add(1)
		go wrapper(&wg, limiter, pool.jobs[i])
	}

	wg.Wait()
	close(limiter)
}
