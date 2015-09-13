package bitpack

import (
	"os"
	"testing"
)

func TestNumBits(t *testing.T) {
	if numBits(1) != 1 {
		t.FailNow()
	}

	if numBits(2) != 2 {
		t.FailNow()
	}
	if numBits(3) != 2 {
		t.FailNow()
	}

	if numBits(10) != 4 {
		t.FailNow()
	}

	if numBits(1025) != 11 {
		t.FailNow()
	}

	if numBits((1<<63)+111111) != 64 {
		t.FailNow()
	}
}

func TestNum(t *testing.T) {
	b := New(30, 1)
	if b.Num() < 30 {
		t.Error(b.Num(), "less than", 30)
	}

	b = New(999, 255)
	if b.Num() < 999 {
		t.Error(b.Num(), "less than", 999)
	}
}

func TestSetGet(t *testing.T) {
	n := 65536
	b := New(n, 127)

	for i := 0; i < n; i++ {
		b.Set(i, uint64(i%128))
	}
	for i := 0; i < n; i++ {
		b.Set(i, uint64(i*37%128))
	}

	for i := 0; i < n; i++ {
		if b.Get(i)[0] != uint64(i*37%128) {
			t.Error(i, b.Get(i), uint64(i*37%128))
		}
	}
}

func TestWriteRead(t *testing.T) {
	n := 255
	b := New(n, 127)

	for i := 0; i < n; i++ {
		b.Set(i, uint64(i*37%128))
	}

	err := b.WriteFile("hoge.tmp")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer os.Remove("hoge.tmp")

	b, err = ReadFile("hoge.tmp")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	for i := 0; i < n; i++ {
		if b.Get(i)[0] != uint64(i*37%128) {
			t.Error(i, b.Get(i), uint64(i*37%128))
		}
	}
}

func BenchmarkSet(b *testing.B) {
	bp := New(50000000, 4*1024*1024*1024*1024, 10*1024*1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bp.Set(i*37%50000000, 0, 0)
	}
}

func BenchmarkGet(b *testing.B) {
	bp := New(50000000, 4*1024*1024*1024*1024, 10*1024*1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bp.Get(i * 37 % 50000000)
	}
}
