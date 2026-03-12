package queue

type MemoryQueue struct {
	jobs chan string
}

func NewMemoryQueue(size int) *MemoryQueue {
	if size < 1 {
		size = 16
	}
	return &MemoryQueue{jobs: make(chan string, size)}
}

func (q *MemoryQueue) Put(jobID string) {
	q.jobs <- jobID
}

func (q *MemoryQueue) Jobs() <-chan string {
	return q.jobs
}
