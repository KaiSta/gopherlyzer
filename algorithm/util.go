package algorithm

func Contains(x interface{}, slice interface{}) bool {
    switch x.(type) {
        case uint64:
            return containsuint64(x.(uint64), slice.([]uint64))
    }
    return false
}

func containsuint64(x uint64, slice []uint64) bool {
    for i := range slice {
        if x == slice[i] {
            return true
        }
    }
    return false
}