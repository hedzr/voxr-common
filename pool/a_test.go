/*
 * Copyright © 2019 Hedzr Yeh.
 */

package pool_test

import (
	"fmt"
	"sync"
)

// A Pool

type workerRoutine struct {
	Func func()
}

type PoolLite struct {
	wg       sync.WaitGroup
	channels chan workerRoutine
}

func (p *PoolLite) Start(size int) {
	p.channels = make(chan workerRoutine, size*2)

	for i := 0; i < size; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for ch := range p.channels {
				// reflect.ValueOf(ch.Func).Call(ch.Args)
				ch.Func()
			}
		}()
	}
}

func (p *PoolLite) Stop() {
	close(p.channels)
	p.wg.Wait()
}

func (p *PoolLite) Add(worker func()) {
	wk := workerRoutine{
		Func: func() {
			fmt.Println(j + j)
		},
	}
	p.channels <- wk
}

func PoolStart(size int) {
	pool := &PoolLite{}
	pool.Start(size)

	for i := 0; i < 100; i++ {
		j := i
		pool.Add(func() {
			fmt.Println(j + j)
		})
	}

	pool.Stop()
}
