package concurrent

import (
    "sync/atomic"
    "unsafe"
)

type AtomicStack struct {
    head unsafe.Pointer
    count uint64
}

type stackElement struct {
    value interface{}
    next unsafe.Pointer
}

func NewAtomicStack() *AtomicStack {
    return &AtomicStack{head:nil}
}

func (s *AtomicStack) Empty() bool {
    return (*stackElement)(atomic.LoadPointer(&s.head)) == nil
}

func (s *AtomicStack) Peek() interface{} {
    el := (*stackElement)(atomic.LoadPointer(&s.head))
    if el != nil {
        return el.value
    }
	return nil
}

func (s *AtomicStack) Pop() {
    for {
        head := atomic.LoadPointer(&s.head)
        
        if (*stackElement)(head) != nil {
            newHead := (*stackElement)(head).next
            if atomic.CompareAndSwapPointer(&s.head, head, newHead) {
                atomic.AddUint64(&s.count, ^uint64(0))
                break
            }
        } else {
            return
        }
    }
}

func (s *AtomicStack) Push(v interface{}) {
    se := &stackElement{value:v}
    for {
        se.next = atomic.LoadPointer(&s.head)
        
        if atomic.CompareAndSwapPointer(&s.head, se.next, unsafe.Pointer(se)) {
            atomic.AddUint64(&s.count, 1)
            break
        }
    }
}

func (s *AtomicStack) Count() uint64 {
    return atomic.LoadUint64(&s.count)
}

func (s *AtomicStack) Iterate() chan interface{} {
    ch :=  make(chan interface{}, s.Count())
    
    go func() {
       if s.Empty() {
           close(ch)
       } else {
           curr := (*stackElement)(atomic.LoadPointer(&s.head))
           for curr != nil {
               ch <- curr.value
               curr = (*stackElement)(atomic.LoadPointer(&curr.next))
           }
           close(ch)
       }
    }()
    return ch
}
