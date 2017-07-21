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

func TestBerlekampWelchSingle(t *testing.T) {
	const block = 1
	const total, required = 7, 3

	test := NewBerlekampWelchTest(t, required, total)
	_, shares := test.SomeShares(block)

	out, err := test.code.berlekampWelch(shares, 0)
	test.AssertNoError(err)
	test.AssertDeepEqual(out, []byte{0x01, 0x02, 0x03, 0x15, 0x69, 0xcc, 0xf2})
}

func TestBerlekampWelch(t *testing.T) {
	const block = 4096
	const total, required = 7, 3

	test := NewBerlekampWelchTest(t, required, total)
	_, shares := test.SomeShares(block)

	test.AssertNoError(test.code.decode(shares, nil))

	shares[0].Data[0]++
	shares[1].Data[0]++

	decoded_shares, callback := test.StoreShares()
	test.AssertNoError(test.code.decode(shares, callback))
	test.AssertDeepEqual(decoded_shares[:required], shares[:required])
}

func TestDecode(t *testing.T) {
	const block = 4096
	const total, required = 7, 3

	test := NewBerlekampWelchTest(t, required, total)
	origdata, shares := test.SomeShares(block)

	combined, err := test.code.Decode(nil, shares)
	test.AssertNoError(err)
	test.AssertDeepEqual(combined, origdata)

	combined, err = test.code.Decode(make([]byte, len(origdata)+1), shares)
	test.AssertNoError(err)
	test.AssertDeepEqual(combined, origdata)
}

func TestBerlekampWelchZero(t *testing.T) {
	const total, required = 40, 20

	test := NewBerlekampWelchTest(t, required, total)

	buf := make([]byte, 200)
	buf = append(buf, bytes.Repeat([]byte{0x14}, 20)...)

	shares, callback := test.StoreShares()
	test.AssertNoError(test.code.Encode(buf, callback))

	shares[0].Data[0]++

	test.AssertNoError(test.code.decode(shares, nil))
}

func TestBerlekampWelchErrors(t *testing.T) {
	const block = 4096
	const total, required = 7, 3

	test := NewBerlekampWelchTest(t, required, total)
	_, shares := test.SomeShares(block)
	test.AssertNoError(test.code.decode(shares, nil))

	for i := 0; i < 500; i++ {
		shares_copy := test.CopyShares(shares)
		for i := 0; i < block; i++ {
			test.MutateShare(i, shares_copy[rand.Intn(total)])
			test.MutateShare(i, shares_copy[rand.Intn(total)])
		}

		decoded_shares, callback := test.StoreShares()
		test.AssertNoError(test.code.decode(shares, callback))
		test.AssertDeepEqual(decoded_shares[:required], shares[:required])
	}
}

func TestBerlekampWelchRandomShares(t *testing.T) {
	const block = 4096
	const total, required = 7, 3

	test := NewBerlekampWelchTest(t, required, total)
	_, shares := test.SomeShares(block)
	test.AssertNoError(test.code.decode(shares, nil))

	for i := 0; i < 500; i++ {
		test_shares := test.CopyShares(shares)
		test.PermuteShares(test_shares)
		test_shares = test_shares[:required+2+rand.Intn(total-required-2)]

		for i := 0; i < block; i++ {
			test.MutateShare(i, test_shares[rand.Intn(len(test_shares))])
		}

		decoded_shares, callback := test.StoreShares()
		test.AssertNoError(test.code.decode(test_shares, callback))
		test.AssertDeepEqual(decoded_shares[:required], shares[:required])
	}
}

func BenchmarkBerlekampWelch(b *testing.B) {
	const block = 4096
	const total, required = 40, 20

	test := NewBerlekampWelchTest(b, required, total)
	_, shares := test.SomeShares(block)

	b.ReportAllocs()
	b.SetBytes(required * block)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		test.AssertNoError(test.code.decode(shares, nil))
	}
}

func BenchmarkBerlekampWelchOneError(b *testing.B) {
	const block = 4096
	const total, required = 40, 20

	test := NewBerlekampWelchTest(b, required, total)
	_, shares := test.SomeShares(block)
	dec_shares := shares[total-required-2:]

	b.ReportAllocs()
	b.SetBytes(required * block)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		test.AssertNoError(test.code.decode(dec_shares, nil))
	}
}

func BenchmarkBerlekampWelchTwoErrors(b *testing.B) {
	const block = 4096
	const total, required = 40, 20

	test := NewBerlekampWelchTest(b, required, total)
	_, shares := test.SomeShares(block)
	dec_shares := shares[total-required-4:]

	b.ReportAllocs()
	b.SetBytes(required * block)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		test.AssertNoError(test.code.decode(dec_shares, nil))
	}
}

///////////////////////////////////////////////////////////////////////////////
// Helpers
///////////////////////////////////////////////////////////////////////////////

type BerlekampWelchTest struct {
	*Asserter

	code *FEC
}

func NewBerlekampWelchTest(t testing.TB,
	required, total int) *BerlekampWelchTest {

	asserter := Wrap(t)

	code, err := NewFEC(required, total)
	asserter.AssertNoError(err)

	return &BerlekampWelchTest{
		Asserter: asserter,

		code: code,
	}
}

func (t *BerlekampWelchTest) StoreShares() ([]Share, func(Share)) {
	out := make([]Share, t.code.n)
	return out, func(s Share) {
		idx := s.Number
		out[idx].Number = idx
		out[idx].Data = append(out[idx].Data, s.Data...)
	}
}

func (t *BerlekampWelchTest) SomeShares(block int) ([]byte, []Share) {
	// seed the initial data
	data := make([]byte, t.code.k*block)
	for i := range data {
		data[i] = byte(i + 1)
	}

	shares, store := t.StoreShares()
	t.AssertNoError(t.code.Encode(data, store))
	return data, shares
}

func (t *BerlekampWelchTest) CopyShares(shares []Share) []Share {
	out := make([]Share, t.code.n)
	for i, share := range shares {
		out[i] = share.DeepCopy()
	}
	return out
}

func (t *BerlekampWelchTest) MutateShare(idx int, share Share) {
	orig := share.Data[idx]
	next := byte(rand.Intn(256))
	for next == orig {
		next = byte(rand.Intn(256))
	}
	share.Data[idx] = next
}

func (t *BerlekampWelchTest) PermuteShares(shares []Share) {
	for i := 0; i < len(shares); i++ {
		with := rand.Intn(len(shares)-i) + i
		shares[i], shares[with] = shares[with], shares[i]
	}
}

func (t *BerlekampWelchTest) DataDiff(a, b []byte) []byte {
	out := make([]byte, len(a))
	for i := range out {
		if a[i] != b[i] {
			out[i] = 0xff
		}
	}
	return out
}
