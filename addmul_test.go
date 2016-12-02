// Copyright (C) 2016 Space Monkey, Inc.

package infectuous

import (
	"bytes"
	"testing"

	"sm/smtest"
	"sm/space/rand"
)

func addmulSlow(z []byte, x []byte, y byte) {
	gf_mul_y := gf_mul_table[y][:]
	for i := range z {
		z[i] ^= gf_mul_y[x[i]]
	}
}

func TestAddmul(t *testing.T) {
	for i := 0; i < 10000; i++ {
		size := rand.Intn(1024)
		y := byte(rand.Intn(256))
		x := smtest.RandomBytes(size)
		z := smtest.RandomBytes(size)
		z1 := append([]byte(nil), z...)
		z2 := append([]byte(nil), z...)

		addmulSlow(z1, x, y)
		addmul(z2, x, y)

		if !bytes.Equal(z1, z2) {
			t.Logf("size: %d", size)
			t.Logf("x: %x", x)
			t.Logf("z: %x", z)
			t.Logf("z1: %x", z1)
			t.Logf("z2: %x", z2)
			t.Fatal("mismatch")
		}
	}
}
