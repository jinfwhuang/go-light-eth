package portalnet

import (
	"sync"
)

type Queue interface {
	Enqueue([]byte)
	Dequeue() []byte
}

// FixedFifo (First In First Out) concurrent queue
type FifoQueue struct {
	rwmutex sync.RWMutex
	arr     [][]byte
}

//func (q *FifoQueue) Len() int {
//	return q.arr.Len()
//}
//func (q *FifoQueue) Cap() int {
//	return q.cap
//}

func (q *FifoQueue) Enqueue(e []byte) {
	q.arr = append(q.arr, e)
}

func (q *FifoQueue) Dequeue() []byte {
	element := []byte{}
	if len(q.arr) > 1 {
		element = q.arr[0]
		q.arr = q.arr[1:]
	}
	return element
}

