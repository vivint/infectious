// The MIT License (MIT)
//
// Copyright (C) 2016-2017 Vivint, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package infectious

import (
	"bytes"
	"math/rand"
	"testing"
)

func addmulSlow(z []byte, x []byte, y byte) {
	gf_mul_y := gf_mul_table[y][:]
	for i := range z {
		z[i] ^= gf_mul_y[x[i]]
	}
}

func TestAddmul(t *testing.T) {
	for i := 0; i < 10000; i++ {
		align := rand.Intn(256)
		size := rand.Intn(1024) + align
		y := byte(rand.Intn(256))
		x := RandomBytes(size)
		z := RandomBytes(size)
		z1 := append([]byte(nil), z...)
		z2 := append([]byte(nil), z...)

		addmulSlow(z1[align:], x[align:], y)
		addmul(z2[align:], x[align:], y)

		if !bytes.Equal(z1, z2) {
			t.Logf("align: %d", align)
			t.Logf("size: %d", size)
			t.Logf("x: %x", x)
			t.Logf("z: %x", z)
			t.Logf("z1: %x", z1)
			t.Logf("z2: %x", z2)
			t.Fatal("mismatch")
		}
	}
}
