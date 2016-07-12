package cache

type FsPLRU struct {
    treeNodeCount uint32
    treeNodeQW uint32
    lruTree []uint32
    size uint32
}

func (fsp *FsPLRU) set(index uint32) {
    fsp.lruTree[index >> 0x5] |= 1 << (index & 0x1F)
}

func (fsp *FsPLRU) clear(index uint32) {
    fsp.lruTree[index >> 0x5] &= ^(1 << (index & 0x1F))
}

func (fsp *FsPLRU) isSet(index uint32) bool {
    return (fsp.lruTree[index >> 0x5] & (1 << (index & 0x1F))) > 0
}

func isPowerOfTwo(n uint32) bool {
    return ((n!=0) && !((n & (n-1))>0))
}

func NewPLRU(size uint32) *FsPLRU {
    if !isPowerOfTwo(size) {
        panic("PLRU intialized with wrong size. Must be n^2")
    }
    
    fsp := &FsPLRU{treeNodeCount: (size-1), size:size}
    fsp.treeNodeQW = (fsp.treeNodeCount + 31)/ 32
    fsp.lruTree = make([]uint32, fsp.treeNodeQW)
    return fsp
}

func (fsp *FsPLRU) Update(index uint32) {
    currNode := uint32(0)
    currStep := uint32(fsp.size/2)
    
    for currStep > 0 {
        if index < currStep {
            //left
            fsp.set(currNode)
            currNode = 2 * currNode + 1
        } else {
            fsp.clear(currNode)
            currNode = 2 * currNode + 2
            index -= currStep
        }
        currStep /= 2
    }
}

func (fsp *FsPLRU) GetIndex() uint32 {
    currNode := uint32(0)
    currVictim := uint32(0)
    currStep := uint32(fsp.size / 2)
    
    for currStep > 0 {
        if fsp.isSet(currNode) {
            //right
            currVictim += currStep
            currNode = 2 * currNode + 2
        } else {
            //left
            currNode = 2 * currNode + 1
        }
        currStep /= 2
    }
    
    return currVictim
}