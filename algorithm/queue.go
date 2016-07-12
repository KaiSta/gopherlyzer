package algorithm

import (
    "container/list"
)

type Queue struct {
    list *list.List
}

func NewQueue() *Queue {
    return &Queue{list:list.New()}
}

func (q *Queue) Empty() bool {
   return q.Size() == 0
}

func (q *Queue) Size() uint64 {
    return (uint64)(q.list.Len())
}

func (q *Queue) Front() interface{} {
    return q.list.Front().Value
}

func (q *Queue) Back() interface{} {
    return q.list.Back().Value
}

func (q *Queue) Push_back(e interface{}) {
    q.list.PushBack(e)
}

func (q *Queue) Pop_front() {
    q.list.Remove(q.list.Front())
}

func (q *Queue) Clear() {
    q.list.Init()
}

func (q *Queue) Iterate() chan interface{} {
    ch := make(chan interface{}, q.Size())
    
    go func() {
        for e:= q.list.Front(); e != nil; e = e.Next() {
            ch <- e.Value
        }
        close(ch)
    }()
    
    return ch
}