// The MIT License (MIT)
//
// Copyright (C) 2016 Space Monkey, Inc.
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
	"sort"

	"github.com/spacemonkeygo/errors"
)

func (fc *FecCode) BerlekampWelch(shares []Share, output Callback) error {
	if len(shares) == 0 {
		return errors.ProgrammerError.New("must specify at least one share")
	}

	sort.Sort(byNumber(shares))

	// fast path: check to see if there are no errors by evaluating it with
	// the syndrome matrix.
	synd, err := fc.syndromeMatrix(shares)
	if err != nil {
		return err
	}
	buf := make([]byte, len(shares[0].Data))

	for i := 0; i < synd.r; i++ {
		for j := range buf {
			buf[j] = 0
		}

		for j := 0; j < synd.c; j++ {
			addmul(buf, shares[j].Data, byte(synd.get(i, j)))
		}

		for j := range buf {
			if buf[j] == 0 {
				continue
			}
			data, err := fc.berlekampWelch(shares, j)
			if err != nil {
				return err
			}
			for _, share := range shares {
				share.Data[j] = data[share.Number]
			}
		}
	}

	return fc.Decode(shares, output)
}

func (fc *FecCode) berlekampWelch(shares []Share, index int) ([]byte, error) {
	k := fc.k        // required size
	r := len(shares) // required + redundancy size
	e := (r - k) / 2 // deg of E polynomial
	q := e + k       // def of Q polynomial

	if e == 0 {
		return nil, Error.New("not enough shares")
	}

	const interp_base = gfVal(2)

	eval_point := func(num int) gfVal {
		if num == 0 {
			return 0
		}
		return interp_base.pow(num - 1)
	}

	dim := q + e

	// build the system of equations s * u = f
	s := matrixNew(dim, dim) // constraint matrix
	a := matrixNew(dim, dim) // augmented matrix
	f := make(gfVals, dim)   // constant column vector
	u := make(gfVals, dim)   // solution vector

	for i := 0; i < dim; i++ {
		x_i := eval_point(shares[i].Number)
		r_i := gfConst(shares[i].Data[index])

		f[i] = x_i.pow(e).mul(r_i)

		for j := 0; j < q; j++ {
			s.set(i, j, x_i.pow(j))
			if i == j {
				a.set(i, j, gfConst(1))
			}
		}

		for k := 0; k < e; k++ {
			j := k + q

			s.set(i, j, x_i.pow(k).mul(r_i))
			if i == j {
				a.set(i, j, gfConst(1))
			}
		}
	}

	// invert and put the result in a
	err := s.invertWith(a)
	if err != nil {
		return nil, err
	}

	// multiply the inverted matrix by the column vector
	for i := 0; i < dim; i++ {
		ri := a.indexRow(i)
		u[i] = ri.dot(f)
	}

	// reverse u for easier construction of the polynomials
	for i := 0; i < len(u)/2; i++ {
		o := len(u) - i - 1
		u[i], u[o] = u[o], u[i]
	}

	q_poly := gfPoly(u[e:])
	e_poly := append(gfPoly{gfConst(1)}, u[:e]...)

	p_poly, rem, err := q_poly.div(e_poly)
	if err != nil {
		return nil, err
	}

	if !rem.isZero() {
		return nil, Error.New("too many errors")
	}

	out := make([]byte, fc.n)
	for i := range out {
		pt := gfConst(0)
		if i != 0 {
			pt = interp_base.pow(i - 1)
		}
		out[i] = byte(p_poly.eval(pt))
	}

	return out, nil
}

func (fc *FecCode) syndromeMatrix(shares []Share) (gfMat, error) {
	// get a list of keepers
	keepers := map[int]struct{}{}
	for _, share := range shares {
		keepers[share.Number] = struct{}{}
	}

	// create a vandermonde matrix but skip columns where we're missing the
	// share.
	out := matrixNew(fc.k, len(keepers))
	for i := 0; i < fc.k; i++ {
		skipped := 0
		for j := 0; j < fc.n; j++ {
			if _, ok := keepers[j]; !ok {
				skipped++
				continue
			}

			out.set(i, j-skipped, gfConst(fc.vand_matrix[i*fc.n+j]))
		}
	}

	// standardize the output and convert into parity form
	err := out.standardize()
	if err != nil {
		return gfMat{}, err
	}

	return out.parity(), nil
}
