package main

const (
	crXXX = uint64(0x0a00000000000000)
	crlf0 = uint64(0x0d0a000000000000)
	crlf1 = uint64(0x000d0a0000000000)
	crlf2 = uint64(0x00000d0a00000000)
	crlf3 = uint64(0x0000000d0a000000)
	crlf4 = uint64(0x000000000d0a0000)
	crlf5 = uint64(0x00000000000d0a00)
	crlf6 = uint64(0x0000000000000d0a)
	crlf7 = uint64(0x000000000000000d)

	mask0 = uint64(0xffff000000000000)
	mask1 = uint64(0x00ffff0000000000)
	mask2 = uint64(0x0000ffff00000000)
	mask3 = uint64(0x000000ffff000000)
	mask4 = uint64(0x00000000ffff0000)
	mask5 = uint64(0x0000000000ffff00)
	mask6 = uint64(0x000000000000ffff)
	mask7 = uint64(0x00000000000000ff)

	left1 = uint64(0xff00000000000000)
	left2 = uint64(0xffff000000000000)
	left3 = uint64(0xffffff0000000000)
	left4 = uint64(0xffffffff00000000)
	left5 = uint64(0xffffffffff000000)
	left6 = uint64(0xffffffffffff0000)
	left7 = uint64(0xffffffffffffff00)
	left8 = uint64(0xffffffffffffffff)

	rght1 = uint64(0x0000ffffffffffff)
	rght2 = uint64(0x000000ffffffffff)
	rght3 = uint64(0x00000000ffffffff)
	rght4 = uint64(0x0000000000ffffff)
	rght5 = uint64(0x000000000000ffff)
	rght6 = uint64(0x00000000000000ff)

	defaultBufferSize = 1 << 16 // 32KB
)

var lefts = [9]uint64{0, left1, left2, left3, left4, left5, left6, left7, left8}
