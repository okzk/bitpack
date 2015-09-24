package bitpack

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

type BitPack struct {
	bits int
	data []uint64
}

func numBits(n uint64) int {
	i := 0
	for ; n > 0; n >>= 1 {
		i++
	}
	return i
}

func New(numHint int, maxHints uint64) *BitPack {
	bits := numBits(maxHints)

	length := (numHint*bits + 63) / 64
	if length == 0 {
		length = 1
	}

	return &BitPack{
		bits: bits,
		data: make([]uint64, length),
	}
}

func (b *BitPack) Num() int {
	if b.bits == 0 {
		return int((^uint(0)) >> 1)
	}
	return len(b.data) * 64 / int(b.bits)
}

func (b *BitPack) Set(n int, v uint64) {
	p := n * b.bits / 64
	r := uint(n * b.bits % 64)

	bits := uint(b.bits)
	if r+bits <= 64 {
		var mask uint64 = (1 << bits) - 1
		x := b.data[p]
		x &^= mask << r
		x |= (v & mask) << r
		b.data[p] = x
	} else {
		b1 := 64 - r
		b2 := bits - b1

		var mask1 uint64 = (1 << b1) - 1
		x := b.data[p]
		x &^= mask1 << r
		x |= v << r
		b.data[p] = x

		v >>= b1
		var mask2 uint64 = (1 << b2) - 1
		x = b.data[p+1]
		x &^= mask2
		x |= v & mask2
		b.data[p+1] = x
	}
}

func (b *BitPack) Get(n int) uint64 {
	p := n * b.bits / 64
	r := uint(n * b.bits % 64)

	bits := uint(b.bits)
	if r+bits <= 64 {
		var mask uint64 = (1 << bits) - 1
		return (b.data[p] >> r) & mask
	} else {
		b1 := 64 - r
		b2 := bits - b1

		x1 := b.data[p] >> r

		var mask uint64 = (1 << b2) - 1
		x2 := b.data[p+1] & mask

		return x2<<b1 | x1
	}
}

func (b *BitPack) Write(w io.Writer) error {
	err := binary.Write(w, binary.LittleEndian, uint8(b.bits))
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, uint64(len(b.data)))
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
	var bits uint8
	err := binary.Read(r, binary.LittleEndian, &bits)
	if err != nil {
		return nil, err
	}
	if bits > 64 {
		return nil, errors.New("invalid data")
	}
	var length uint64
	err = binary.Read(r, binary.LittleEndian, &length)
	if err != nil {
		return nil, err
	}
	data := make([]uint64, int(length))
	err = binary.Read(r, binary.LittleEndian, &data)
	if err != nil {
		return nil, err
	}

	return &BitPack{
		bits: int(bits),
		data: data,
	}, nil
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
