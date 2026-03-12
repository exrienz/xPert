package workerpool

import "sync"

type WorkerPool struct {
	jobs     chan func()
	workers  sync.WaitGroup
	tasks    sync.WaitGroup
	stopOnce sync.Once
}

func New(workerCount int, queueSize int) *WorkerPool {
	if workerCount < 1 {
		workerCount = 1
	}
	if queueSize < workerCount {
		queueSize = workerCount
	}
	pool := &WorkerPool{
		jobs: make(chan func(), queueSize),
	}
	pool.workers.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer pool.workers.Done()
			for job := range pool.jobs {
				if job == nil {
					continue
				}
				job()
			}
		}()
	}
	return pool
}

func (p *WorkerPool) Submit(job func()) {
	p.tasks.Add(1)
	p.jobs <- func() {
		defer p.tasks.Done()
		job()
	}
}

func (p *WorkerPool) Wait() {
	p.tasks.Wait()
}

func (p *WorkerPool) Stop() {
	p.stopOnce.Do(func() {
		close(p.jobs)
		p.workers.Wait()
	})
}
