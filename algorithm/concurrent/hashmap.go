package concurrent

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Hashfunction func(interface{}) uint32

type Hashvector struct {
	elements []element
	size     uint32
	count    uint32
	hashfunc Hashfunction
	mtx      sync.RWMutex
}

type element struct {
	Key    uint32
	OuterK interface{}
	Value  interface{}
}

func NewHashVector(f Hashfunction) *Hashvector {
	return &Hashvector{elements: make([]element, 32), size: 32, hashfunc: f, count: 0}
}

func (hv *Hashvector) Insert(k interface{}, v interface{}) {
	hv.mtx.RLock()

	key := hv.hashfunc(k)
	var rounds uint32 = 0

	for idx := key; ; idx++ {
		idx = idx % atomic.LoadUint32(&hv.size)
 
		swapped := atomic.CompareAndSwapUint32(&hv.elements[idx].Key, 0, key)
		if swapped { //new key?
			hv.elements[idx].Value = v
			hv.elements[idx].OuterK = k
			atomic.AddUint32(&hv.count, 1)
			break
		} else if atomic.LoadUint32(&hv.elements[idx].Key) == key { //already stored key
			panic("tried to change existing key")
			hv.elements[idx].Value = v
			break
		}

		rounds += 1
		if rounds >= atomic.LoadUint32(&hv.size) {
			hv.mtx.RUnlock()
			hv.resize(atomic.LoadUint32(&hv.size))
			hv.Insert(k, v)
			return
		}
	}

	hv.mtx.RUnlock()
}

func (hv *Hashvector) resize(oldSize uint32) {
	hv.mtx.Lock()

	if atomic.LoadUint32(&hv.size) > oldSize {
        hv.mtx.Unlock()
		return
	}

	nsize := atomic.LoadUint32(&hv.size) * 2
	tmp := make([]element, nsize)

	for _, v := range hv.elements {
		if v.OuterK != nil {
			key := hv.hashfunc(v.OuterK)
			var rounds uint32 = 0

			for idx := key; ; idx++ {
				idx = idx % nsize
				if tmp[idx].Key == 0 {
					tmp[idx].Key = key
					tmp[idx].Value = v.Value
					tmp[idx].OuterK = v.OuterK
					break
				}

				rounds += 1
				if rounds >= nsize {
					fmt.Println("rz", rounds, nsize)
					panic("Hasharray is overfilled")
				}
			}
		} else {
			panic("nil element?")
		}
	}
    hv.elements = tmp
	atomic.StoreUint32(&hv.size, nsize)
    
    hv.mtx.Unlock()
}

func (hv *Hashvector) Get(k interface{}) (interface{}, bool) {
	hv.mtx.RLock()
	defer hv.mtx.RUnlock()

	key := hv.hashfunc(k)
	var rounds uint32 = 0

	for idx := key; ; idx++ {
		idx = idx % hv.size
		probedKey := atomic.LoadUint32(&hv.elements[idx].Key)

		if probedKey == key {
            if hv.elements[idx].OuterK != nil && hv.elements[idx].Value != nil {
                return hv.elements[idx].Value, true
            } else {
                return nil, false
            }
		} else if probedKey == 0 {
			return nil, false
		}

		rounds += 1
		if rounds >= hv.size {
			return nil, false
		}
	}
}


func (hv *Hashvector) PrintState() {
	fmt.Println("size=", hv.size, "count=", hv.count)
    
    c := 0
    for _, v := range hv.elements {
        if v.OuterK == nil {
            c++
        }
        //fmt.Println(v)
    }

	fmt.Println("free=", c)
}


// func (hv *hashvector) PrintState() {
// 	fmt.Println("size=", hv.size, "count=", hv.count)
// 	vptr := (*[]element)(hv.vals)
// 	c := 0
// 	for i := uint32(0); i < hv.size; i++ {
// 		if (*vptr)[i].OuterK == nil {
// 			c++
// 		}
// 	}
// 	fmt.Println("free=", c)
// }
// 
// func (hv *hashvector) resize() {
// 	atomic.AddUint32(&hv.cWriter, ^uint32(0))
// 
// 	if atomic.CompareAndSwapUint32(&hv.resizeV, 0, 1) {
// 		for atomic.LoadUint32(&hv.cWriter) > 0 {
// 		}
// 
// 		nsize := atomic.LoadUint32(&hv.size) * 2
// 		tmp := make([]element, nsize)
// 		vptr := (*[]element)(hv.vals)
// 
// 		for i := uint32(0); i < hv.size; i++ {
// 			if (*vptr)[i].OuterK != nil {
// 				key := hv.hashfunc((*vptr)[i].OuterK)
// 				var rounds uint32 = 0
// 
// 				for idx := key; ; idx++ {
// 					idx = idx % nsize
// 					if tmp[idx].Key == 0 {
// 						tmp[idx].Key = key
// 						tmp[idx].Value = (*vptr)[i].Value
// 						tmp[idx].OuterK = (*vptr)[i].OuterK
// 						break
// 					}
// 
// 					rounds += 1
// 					if rounds >= nsize {
// 						fmt.Println("rz", rounds, nsize)
// 						panic("Hasharray is overfilled")
// 					}
// 				}
// 			}
// 		}
// 		atomic.StorePointer(&hv.vals, unsafe.Pointer(&tmp))
// 		atomic.StoreUint32(&hv.size, nsize)
// 		atomic.StoreUint32(&hv.resizeV, 0)
// 	} else {
// 		for atomic.LoadUint32(&hv.resizeV) == 1 {
// 		}
// 	}
// }
// 
// func (hv *hashvector) Insert(k interface{}, v interface{}) {
// 	atomic.AddUint32(&hv.cWriter, 1)
// 
// 	key := hv.hashfunc(k)
// 	var rounds uint32 = 0
// 
// 	for idx := key; ; idx++ {
// 		if atomic.LoadUint32(&hv.resizeV) > 0 {
// 			atomic.AddUint32(&hv.cWriter, ^uint32(0))
// 			for atomic.LoadUint32(&hv.resizeV) == 1 {
// 			}
// 			hv.Insert(k, v)
// 			return
// 		}
// 
// 		idx = idx % atomic.LoadUint32(&hv.size)
// 		swapped := atomic.CompareAndSwapUint32(&(*(*[]element)(atomic.LoadPointer(&hv.vals)))[idx].Key, 0, key)
// 		if swapped { //new key?
// 			(*(*[]element)(atomic.LoadPointer(&hv.vals)))[idx].Value = v
// 			(*(*[]element)(atomic.LoadPointer(&hv.vals)))[idx].OuterK = k
// 			atomic.AddUint32(&hv.count, 1)
// 			break
// 		} else if atomic.LoadUint32(&(*(*[]element)(atomic.LoadPointer(&hv.vals)))[idx].Key) == key { //already stored key
// 			panic("tried to change existing key")
// 			(*(*[]element)(atomic.LoadPointer(&hv.vals)))[idx].Value = v
// 			break
// 		}
// 
// 		rounds += 1
// 		if rounds >= atomic.LoadUint32(&hv.size) {
// 			hv.resize()
// 			hv.Insert(k, v)
// 			return
// 		}
// 	}
// 
// 	atomic.AddUint32(&hv.cWriter, ^uint32(0))
// }
