// Copyright (C) 2016 Space Monkey, Inc.

// +build !amd64

package infectuous

func addmul(z []byte, x []byte, y byte) {
	if y == 0 {
		return
	}

	// hint to the compiler that we don't need bounds checks on x
	x = x[:len(z)]

	// TODO(jeff): loop unrolling for SPEEDS
	gf_mul_y := gf_mul_table[y][:]
	for i := range z {
		z[i] ^= gf_mul_y[x[i]]
	}
}
