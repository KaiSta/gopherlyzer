package concurrent

import (
    "sync/atomic"
)

type haelement struct {
    Key uint32
    Value interface{}
}

type Hasharray struct {
   Elements []haelement
   Size uint32 
}

const (
    FNV_prime = 16777619
    FNV_offset = 2166136261
)
func hash(n string) uint32 {
    hash := uint32(FNV_offset)
    for _, c := range n {
        hash = hash ^ uint32(c)
        hash = hash * FNV_prime
    }
    return hash + 1
}

func NewHasharray(size uint32) Hasharray {
    return Hasharray{Elements:make([]haelement, size), Size:size}
}

func (h *Hasharray) Insert(k string, v interface{}) {
    key := hash(k)
    var rounds uint32 = 0
    
    for idx := key;; idx++ {
        idx = idx % h.Size
        swapped :=  atomic.CompareAndSwapUint32(&h.Elements[idx].Key, 0, key)
        if swapped { //new key?
            h.Elements[idx].Value = v
            return
        } else if atomic.LoadUint32(&h.Elements[idx].Key) == key { //already stored key
            panic("tried to change existing key")
            h.Elements[idx].Value = v
            return
        }
        
        rounds += 1
        if rounds >= h.Size {
            panic("Hasharray is overfilled")
        }
    }
}

func (h *Hasharray) Get(k string) (interface{}, bool) {
    key := hash(k)
    var rounds uint32 = 0
    
    for idx := key;; idx++ {
        idx = idx % h.Size
        probedKey :=  atomic.LoadUint32(&h.Elements[idx].Key)
        
        if probedKey == key {
            return h.Elements[idx].Value, true
        } else if probedKey == 0 {
            return 0, false
        }
        
        rounds += 1
        if rounds >= h.Size {
            return 0, false
        }
    }
}
