package bitpack

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

type BitPack struct {
	totalBits int
	bits      []uint
	data      []uint64
}

func numBits(n uint64) int {
	i := 0
	for ; n > 0; n >>= 1 {
		i++
	}
	return i
}

func New(numHint int, maxHints ...uint64) *BitPack {
	total := 0
	bits := make([]uint, len(maxHints))
	for i, n := range maxHints {
		b := numBits(n)
		bits[i] = uint(b)
		total += b
	}

	length := (numHint*total + 63) / 64
	if length == 0 {
		length = 1
	}

	return &BitPack{
		totalBits: total,
		bits:      bits,
		data:      make([]uint64, length),
	}
}

func (b *BitPack) Num() int {
	if b.totalBits == 0 {
		return int((^uint(0)) >> 1)
	}
	return len(b.data) * 64 / b.totalBits
}

func (b *BitPack) Set(n int, v ...uint64) {
	p := n * b.totalBits / 64
	r := uint(n * b.totalBits % 64)

	for i, vv := range v {
		bits := b.bits[i]

		if r+bits <= 64 {
			var mask uint64 = (1 << bits) - 1

			x := b.data[p]
			x &^= mask << r
			x |= (vv & mask) << r
			b.data[p] = x

			r += bits
			if r == 64 {
				p += 1
				r = 0
			}
		} else {
			b1 := 64 - r
			b2 := bits - b1

			var mask uint64 = (1 << b1) - 1

			x := b.data[p]
			x &^= mask << r
			x |= vv << r
			b.data[p] = x

			vv >>= b1
			mask = (1 << b2) - 1
			p += 1
			r = b2

			x = b.data[p]
			x &^= mask
			x |= vv & mask
			b.data[p] = x
		}
	}
}

func (b *BitPack) Get(n int) []uint64 {
	p := n * b.totalBits / 64
	r := uint(n * b.totalBits % 64)

	ret := make([]uint64, len(b.bits))
	for i, bits := range b.bits {

		if r+bits <= 64 {
			var mask uint64 = (1 << uint(bits)) - 1
			x := b.data[p]
			ret[i] = (x >> r) & mask

			r += bits
			if r == 64 {
				p += 1
				r = 0
			}
		} else {
			b1 := 64 - r
			b2 := bits - b1

			x1 := b.data[p] >> r

			p += 1
			r = b2

			var mask uint64 = (1 << b2) - 1
			x2 := b.data[p] & mask

			ret[i] = x2<<b1 | x1
		}
	}
	return ret
}

func (b *BitPack) Write(w io.Writer) error {
	length := []uint64{uint64(len(b.bits)), uint64(len(b.data))}
	err := binary.Write(w, binary.LittleEndian, length)
	if err != nil {
		return err
	}

	tmp := make([]uint8, len(b.bits))
	for i, bits := range b.bits {
		tmp[i] = uint8(bits)
	}

	err = binary.Write(w, binary.LittleEndian, tmp)
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, b.data)
	return err
}

func (b *BitPack) WriteFile(file string) (err error) {
	f, err := os.Create(file)
	if err != nil {
		return
	}
	defer func() {
		e := f.Close()
		if err == nil {
			err = e
		}
	}()
	err = b.Write(f)
	return
}

func Read(r io.Reader) (*BitPack, error) {
	length := make([]uint64, 2)
	err := binary.Read(r, binary.LittleEndian, &length)

	tmp := make([]uint8, int(length[0]))
	err = binary.Read(r, binary.LittleEndian, &tmp)
	if err != nil {
		return nil, err
	}

	b := BitPack{
		bits: make([]uint, int(length[0])),
		data: make([]uint64, int(length[1])),
	}

	total := 0
	for i, bits := range tmp {
		if bits > 64 {
			return nil, errors.New("invalid data")
		}
		b.bits[i] = uint(bits)
		total += int(bits)
	}
	b.totalBits = total

	err = binary.Read(r, binary.LittleEndian, &b.data)
	if err != nil {
		return nil, err
	}

	return &b, nil
}

func ReadFile(file string) (b *BitPack, err error) {
	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer func() {
		e := f.Close()
		if err == nil {
			err = e
		}
	}()

	b, err = Read(f)
	return
}
