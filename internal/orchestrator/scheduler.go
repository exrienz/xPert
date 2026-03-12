package orchestrator

import (
	"sync"

	"xpert/internal/queue"
)

type Scheduler struct {
	queue      queue.Queue
	jobManager *JobManager
	stop       chan struct{}
	wg         sync.WaitGroup
}

func NewScheduler(queue queue.Queue, jobManager *JobManager) *Scheduler {
	return &Scheduler{
		queue:      queue,
		jobManager: jobManager,
		stop:       make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.stop:
				return
			case jobID := <-s.queue.Jobs():
				s.jobManager.ProcessJob(jobID)
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	close(s.stop)
	s.wg.Wait()
}
