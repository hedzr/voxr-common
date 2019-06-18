/*
 * Copyright © 2019 Hedzr Yeh.
 */

package pool

import (
	log "github.com/sirupsen/logrus"
	"time"
)

// WorkerWithFunc is the actual executor who runs the tasks,
// it starts a goroutine that accepts tasks and
// performs function calls.
type WorkerWithFunc struct {
	// pool who owns this worker.
	pool *PoolWithFunc

	// args is a job should be done.
	args chan interface{}

	// recycleTime will be update when putting a worker back into queue.
	recycleTime time.Time
}

// run starts a goroutine to repeat the process
// that performs the function calls.
func (w *WorkerWithFunc) run() {
	w.pool.incRunning()
	go func() {
		defer func() {
			if p := recover(); p != nil {
				w.pool.decRunning()
				if w.pool.PanicHandler != nil {
					w.pool.PanicHandler(p)
				} else {
					log.Printf("worker exits from a panic: %v", p)
				}
			}
		}()

		for args := range w.args {
			if args == nil {
				w.pool.decRunning()
				w.pool.workerCache.Put(w)
				return
			}
			w.pool.poolFunc(args)
			w.pool.revertWorker(w)
		}
	}()
}
