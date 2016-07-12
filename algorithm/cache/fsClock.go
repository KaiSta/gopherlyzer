package cache

type FsClock struct {
    s uint32
    c []uint32
    p uint32
}

func NewFsClock(size uint32) *FsClock {
    return &FsClock{size, make([]uint32, size), 0}
}

func (fc * FsClock) Update(idx uint32) {
    fc.c[idx]++
}

func (fc *FsClock) GetIndex() uint32 {    
    for {
        if fc.c[fc.p] == 0 {
            tmp := fc.p
            fc.p = (fc.p+1) % uint32(len(fc.c))
            return tmp
        }
        fc.c[fc.p]--
        fc.p = (fc.p+1) % uint32(len(fc.c))
    }  
}