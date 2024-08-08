package main

import (
	"encoding/binary"
	"io"
)

// dos2unix64 -- very "clever" way to convert crlf streams into just cr.
// This works by handling only 8 bytes at a time, so buffered reader and writer
// are recommended. The method here is to read 8 bytes and convert it into an
// interger, big-endian. From there, we can AND with byte masks and see if that
// equals an integer that is CRLF in the same byte position. For example:
//
// "012\r\n567" -> uint64(0x3031320d0a353637)
//
//    0x3031320d0a353637
//  & 0x000000ffff000000
// ---------------------
//    0x0000000d0a000000
//
// We know that uint64(0x00000d0a000000) means that the 4th and 5th bytes
// (offsets 3,4) must be CRLF. So we can take the left 3 bytes by ANDing
// with 0xffffff0000000000. We skip the we can take bytes 5, 6, 7, and 8 by
// ANDing with 0x0000000fffffffff, which we left-shift by 8 to consume the
// 4th byte.
//
// 0x3031320d0a353637 & 0xffffff0000000000 -> 0x3031320000000000
// 0x3031320d0a353637 & 0x0000000fffffffff -> 0x000000000a353637
// 0x000000000a353637 << 8 -> 0x0000000a35363700
//
//   0x3031320000000000
// | 0x0000000a35363700
// --------------------
//   0x3031320a35363700
//
// Since we consumed one byte, we know we should only write the first 7 bytes,
// not the full 8.
//
// The only question is, what to do if \r\n is split across an 8-byte boundary?
// We know that this can only happen if the \r appears as the last byte. In that
// case, we can make sure to ignore the last byte. To not lose that byte, we can
// write that byte into the first position and then read 7 bytes starting at
// position 1 the next time the loop goes around. That makes this case just like
// reading with \r\n at the head of the array, so it is no longer special.
func dos2unix64(w io.Writer, r io.Reader) (int, error) {
	bb := make([]byte, 8)
	offset := 0
	total := 0

	for {
		len, err := r.Read(bb[offset:])
		if err == io.EOF {
			break
		}
		if err != nil {
			return total, err
		}
		len += offset
		offset = 0
		// masked to prevent data from last cycle from confusing this cycle
		ival := binary.BigEndian.Uint64(bb) & lefts[len]
		// we need to make a copy to avoid catching duplicate "\r" values.
		// e.g. "\r\r\n" should become "\r\n", not "\n"
		orig := ival

		// last char is \r
		if (orig & mask7) == crlf7 {
			len -= 1
			offset = 1
		}

		// Working right to left, check to see if the crlf is in there.
		// If so, take everything left of the cr, OR it with the lf and
		// everything to the right. Indicate that we ate a character.
		if (orig & mask6) == crlf6 {
			ival = (ival & left6) | ((ival & rght6) << 8)
			len -= 1
		}
		if (orig & mask5) == crlf5 {
			ival = (ival & left5) | ((ival & rght5) << 8)
			len -= 1
		}
		if (orig & mask4) == crlf4 {
			ival = (ival & left4) | ((ival & rght4) << 8)
			len -= 1
		}
		if (orig & mask3) == crlf3 {
			ival = (ival & left3) | ((ival & rght3) << 8)
			len -= 1
		}
		if (orig & mask2) == crlf2 {
			ival = (ival & left2) | ((ival & rght2) << 8)
			len -= 1
		}
		if (orig & mask1) == crlf1 {
			ival = (ival & left1) | ((ival & rght1) << 8)
			len -= 1
		}
		if (orig & mask0) == crlf0 {
			ival <<= 8
			len -= 1
		}

		binary.BigEndian.PutUint64(bb, ival)
		n, err := w.Write(bb[:len])
		if err != nil {
			return total, err
		}
		total += n
		if offset == 1 {
			binary.BigEndian.PutUint64(bb, crXXX)
			bb[0] = '\r'
		}
	}

	if offset == 1 {
		n, err := w.Write(bb[:1])
		if err != nil {
			return total, err
		}
		total += n
	}

	return total, nil
}
