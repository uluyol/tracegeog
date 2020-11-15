package tracer

type bitmap2 struct {
	d      []uint64
	stride int
	num    int
}

func newBitmap2(x, y int) *bitmap2 {
	return &bitmap2{
		d:      make([]uint64, (x*y+63)/64),
		stride: x,
		num:    0,
	}
}

func (b *bitmap2) Count() int { return b.num }

func (b *bitmap2) Set(x, y int) {
	bit := x + y*b.stride
	index := bit / 64
	bit %= 64
	b.d[index] = b.d[index] | (1 << bit)
	b.num++
}

func (b *bitmap2) Get(x, y int) bool {
	bit := x + y*b.stride
	index := bit / 64
	bit %= 64
	return b.d[index]&(1<<bit) != 0
}

func (b *bitmap2) Clear() {
	for i := range b.d {
		b.d[i] = 0
	}
	b.num = 0
}
