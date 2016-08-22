package algorithm

type HashSet struct {
	data map[interface{}]struct{}
	size uint64
}

func NewHashSet() *HashSet {
	return &HashSet{data: make(map[interface{}]struct{})}
}

func (s *HashSet) Insert(el interface{}) {
	_, ok := s.data[el]
	if !ok {
		s.data[el] = struct{}{}
	}
}

func (s *HashSet) Clear() {
	s.data = make(map[interface{}]struct{})
}

func (s *HashSet) Size() uint64 {
	return uint64(len(s.data))
}

func (s *HashSet) Empty() bool {
	return s.size == 0
}

func (s *HashSet) Iterate() chan interface{} {
	ch := make(chan interface{}, len(s.data))
	go func() {
		for k := range s.data {
			ch <- k
		}
		close(ch)
	}()
	return ch
}
