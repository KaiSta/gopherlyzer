package algorithm

type Stack struct {
	data    []interface{}
	currTop int
}

func NewStack() *Stack {
	return &Stack{currTop: -1, data: make([]interface{}, 0)}
}

func (s *Stack) Peek() interface{} {
	if s.currTop > -1 {
		return s.data[s.currTop]
	}
	return nil
}

func (s *Stack) Pop() {
    if s.currTop > -1 {
       s.currTop-- 
    }
}

func (s *Stack) Push(v interface{}) {
    if len(s.data) > s.currTop+1 {
        s.currTop++
        s.data[s.currTop] = v
        return
    }
    s.data = append(s.data, v)
    s.currTop++
}

func (s *Stack) Count() int {
    if s.currTop == -1 {
        return 0
    }
    return s.currTop+1
}

func (s *Stack) Empty() bool {
    return s.currTop == -1
}

func (s *Stack) Iterate() chan interface{} {
    ch :=  make(chan interface{}, s.Count())
    
    go func() {
       if s.Empty() {
           close(ch)
       } else {
           for i := 0; i < s.Count(); i++ {
               ch <- s.data[i]
           }
           close(ch)
       }
    }()
    return ch
}