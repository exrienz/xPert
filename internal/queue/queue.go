package queue

type Queue interface {
	Put(jobID string)
	Jobs() <-chan string
}
