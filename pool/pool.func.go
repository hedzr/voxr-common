/*
 * Copyright © 2019 Hedzr Yeh.
 */

package pool

import (
	"sync"
	"sync/atomic"
	"time"
)

// PoolWithFunc accept the tasks from client,it limits the total
// of goroutines to a given number by recycling goroutines.
type PoolWithFunc struct {
	// capacity of the pool.
	capacity int32

	// running is the number of the currently running goroutines.
	running int32

	// expiryDuration set the expired time (second) of every worker.
	expiryDuration time.Duration

	// workers is a slice that store the available workers.
	workers []*WorkerWithFunc

	// release is used to notice the pool to closed itself.
	release int32

	// lock for synchronous operation.
	lock sync.Mutex

	// cond for waiting to get a idle worker.
	cond *sync.Cond

	// poolFunc is the function for processing tasks.
	poolFunc func(interface{})

	// once makes sure releasing this pool will just be done for one time.
	once sync.Once

	// workerCache speeds up the obtainment of the an usable worker in function:retrieveWorker.
	workerCache sync.Pool

	// PanicHandler is used to handle panics from each worker goroutine.
	// if nil, panics will be thrown out again from worker goroutines.
	PanicHandler func(interface{})
}

// clear expired workers periodically.
func (p *PoolWithFunc) periodicallyPurge() {
	heartbeat := time.NewTicker(p.expiryDuration)
	defer heartbeat.Stop()

	for range heartbeat.C {
		currentTime := time.Now()
		p.lock.Lock()
		idleWorkers := p.workers
		if len(idleWorkers) == 0 && p.Running() == 0 && atomic.LoadInt32(&p.release) == 1 {
			p.lock.Unlock()
			return
		}
		n := -1
		for i, w := range idleWorkers {
			if currentTime.Sub(w.recycleTime) <= p.expiryDuration {
				break
			}
			n = i
			w.args <- nil
			idleWorkers[i] = nil
		}
		if n > -1 {
			if n >= len(idleWorkers)-1 {
				p.workers = idleWorkers[:0]
			} else {
				p.workers = idleWorkers[n+1:]
			}
		}
		p.lock.Unlock()
	}
}

// NewPoolWithFunc generates an instance of go-routine pool with a specific function.
func NewPoolWithFunc(size int, pf func(interface{})) (*PoolWithFunc, error) {
	return NewTimingPoolWithFunc(size, DefaultCleanIntervalTime, pf)
}

// NewTimingPoolWithFunc generates an instance of go-routine pool with a specific function and a custom timed task.
func NewTimingPoolWithFunc(size int, expiry time.Duration, pf func(interface{})) (*PoolWithFunc, error) {
	if size <= 0 {
		return nil, ErrInvalidPoolSize
	}
	if expiry <= 0 {
		return nil, ErrInvalidPoolExpiry
	}
	p := &PoolWithFunc{
		capacity:       int32(size),
		expiryDuration: expiry,
		poolFunc:       pf,
	}
	p.cond = sync.NewCond(&p.lock)
	go p.periodicallyPurge()
	return p, nil
}

// ---------------------------------------------------------------------------

// Invoke submits a task to pool.
func (p *PoolWithFunc) Invoke(args interface{}) error {
	if 1 == atomic.LoadInt32(&p.release) {
		return ErrPoolClosed
	}
	p.retrieveWorker().args <- args
	return nil
}

// Running returns the number of the currently running goroutines.
func (p *PoolWithFunc) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

// Free returns a available goroutines to work.
func (p *PoolWithFunc) Free() int {
	return int(atomic.LoadInt32(&p.capacity) - atomic.LoadInt32(&p.running))
}

// Capacity returns the capacity of this pool.
func (p *PoolWithFunc) Cap() int {
	return int(atomic.LoadInt32(&p.capacity))
}

// Tune change the capacity of this pool.
func (p *PoolWithFunc) Tune(size int) {
	if size == p.Cap() {
		return
	}
	atomic.StoreInt32(&p.capacity, int32(size))
	diff := p.Running() - size
	for i := 0; i < diff; i++ {
		p.retrieveWorker().args <- nil
	}
}

// Release Closed this pool.
func (p *PoolWithFunc) Release() error {
	p.once.Do(func() {
		atomic.StoreInt32(&p.release, 1)
		p.lock.Lock()
		idleWorkers := p.workers
		for i, w := range idleWorkers {
			w.args <- nil
			idleWorkers[i] = nil
		}
		p.workers = nil
		p.lock.Unlock()
	})
	return nil
}

// ---------------------------------------------------------------------------

// incRunning increases the number of the currently running goroutines.
func (p *PoolWithFunc) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

// decRunning decreases the number of the currently running goroutines.
func (p *PoolWithFunc) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

// retrieveWorker returns a available worker to run the tasks.
func (p *PoolWithFunc) retrieveWorker() *WorkerWithFunc {
	var w *WorkerWithFunc

	p.lock.Lock()
	idleWorkers := p.workers
	n := len(idleWorkers) - 1
	if n >= 0 {
		w = idleWorkers[n]
		idleWorkers[n] = nil
		p.workers = idleWorkers[:n]
		p.lock.Unlock()
	} else if p.Running() < p.Cap() {
		p.lock.Unlock()
		if cacheWorker := p.workerCache.Get(); cacheWorker != nil {
			w = cacheWorker.(*WorkerWithFunc)
		} else {
			w = &WorkerWithFunc{
				pool: p,
				args: make(chan interface{}, workerChanCap),
			}
		}
		w.run()
	} else {
		for {
			p.cond.Wait()
			l := len(p.workers) - 1
			if l < 0 {
				continue
			}
			w = p.workers[l]
			p.workers[l] = nil
			p.workers = p.workers[:l]
			break
		}
		p.lock.Unlock()
	}
	return w
}

// revertWorker puts a worker back into free pool, recycling the goroutines.
func (p *PoolWithFunc) revertWorker(worker *WorkerWithFunc) {
	worker.recycleTime = time.Now()
	p.lock.Lock()
	p.workers = append(p.workers, worker)
	// Notify the invoker stuck in 'retrieveWorker()' of there is an available worker in the worker queue.
	p.cond.Signal()
	p.lock.Unlock()
}
