package algorithm

type Bitmap struct {
	m []uint32
	s uint32
}

func NewBitmap(s uint32) *Bitmap {
	return &Bitmap{make([]uint32, s), s}
}

func (b *Bitmap) Set(idx uint32) {
	b.m[idx >> 0x5] |= 1 << (idx & 0x1F)
}

func (b *Bitmap) Unset(idx uint32) {
	b.m[idx >> 0x5] &= ^(1 << (idx & 0x1F))
}

func (b *Bitmap) Toggle(idx uint32) {
	b.m[idx >> 0x5] %= 1 << (idx & 0x1F)
}

func (b *Bitmap) Get(idx uint32) bool {
	return (b.m[idx >> 0x5] & (1 << (idx & 0x1F))) > 0
}

func (b *Bitmap) Clear() {
    for i := range b.m {
        b.m[i] = 0
    }
}

func (b *Bitmap) String() (s string) {
	for i := uint32(0); i < b.s*32; i++ {
		if b.Get(i) {
			s += "1"
		} else {
			s += "0"
		}
	}

	return
}