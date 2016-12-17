// Copyright (C) 2016 Space Monkey, Inc.

package infectuous

import (
	"testing"

	"sm/smtest"
	"sm/space/rand"
)

func TestBerlekampWelchSingle(t *testing.T) {
	const block = 1
	const total, required = 7, 3

	test := NewBerlekampWelchTest(t, required, total)
	shares := test.SomeShares(block)

	out, err := test.code.berlekampWelch(shares, 0)
	test.AssertNoError(err)
	test.AssertDeepEqual(out, []byte{0x01, 0x02, 0x03, 0x15, 0x69, 0xcc, 0xf2})
}

func TestBerlekampWelch(t *testing.T) {
	const block = 4096
	const total, required = 7, 3

	test := NewBerlekampWelchTest(t, required, total)
	shares := test.SomeShares(block)

	test.AssertNoError(test.code.BerlekampWelch(shares, nil))

	shares[0].Data[0]++
	shares[1].Data[0]++

	decoded_shares, callback := test.StoreShares()
	test.AssertNoError(test.code.BerlekampWelch(shares, callback))
	test.AssertDeepEqual(decoded_shares[:required], shares[:required])
}

func TestBerlekampWelchErrors(t *testing.T) {
	const block = 4096
	const total, required = 7, 3

	test := NewBerlekampWelchTest(t, required, total)
	shares := test.SomeShares(block)
	test.AssertNoError(test.code.BerlekampWelch(shares, nil))

	for i := 0; i < 500; i++ {
		shares_copy := test.CopyShares(shares)
		for i := 0; i < block; i++ {
			test.MutateShare(i, shares_copy[rand.Intn(total)])
			test.MutateShare(i, shares_copy[rand.Intn(total)])
		}

		decoded_shares, callback := test.StoreShares()
		test.AssertNoError(test.code.BerlekampWelch(shares, callback))
		test.AssertDeepEqual(decoded_shares[:required], shares[:required])
	}
}

func TestBerlekampWelchRandomShares(t *testing.T) {
	const block = 4096
	const total, required = 7, 3

	test := NewBerlekampWelchTest(t, required, total)
	shares := test.SomeShares(block)
	test.AssertNoError(test.code.BerlekampWelch(shares, nil))

	for i := 0; i < 500; i++ {
		test_shares := test.CopyShares(shares)
		test.PermuteShares(test_shares)
		test_shares = test_shares[:required+2+rand.Intn(total-required-2)]

		for i := 0; i < block; i++ {
			test.MutateShare(i, test_shares[rand.Intn(len(test_shares))])
		}

		decoded_shares, callback := test.StoreShares()
		test.AssertNoError(test.code.BerlekampWelch(test_shares, callback))
		test.AssertDeepEqual(decoded_shares[:required], shares[:required])
	}
}

func BenchmarkBerlekampWelch(b *testing.B) {
	const block = 4096
	const total, required = 40, 20

	test := NewBerlekampWelchTest(b, required, total)
	shares := test.SomeShares(block)

	b.ReportAllocs()
	b.SetBytes(required * block)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		test.AssertNoError(test.code.BerlekampWelch(shares, nil))
	}
}

func BenchmarkBerlekampWelchOneError(b *testing.B) {
	const block = 4096
	const total, required = 40, 20

	test := NewBerlekampWelchTest(b, required, total)
	shares := test.SomeShares(block)
	dec_shares := shares[total-required-2:]

	b.ReportAllocs()
	b.SetBytes(required * block)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		test.AssertNoError(test.code.BerlekampWelch(dec_shares, nil))
	}
}

func BenchmarkBerlekampWelchTwoErrors(b *testing.B) {
	const block = 4096
	const total, required = 40, 20

	test := NewBerlekampWelchTest(b, required, total)
	shares := test.SomeShares(block)
	dec_shares := shares[total-required-4:]

	b.ReportAllocs()
	b.SetBytes(required * block)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		test.AssertNoError(test.code.BerlekampWelch(dec_shares, nil))
	}
}

///////////////////////////////////////////////////////////////////////////////
// Helpers
///////////////////////////////////////////////////////////////////////////////

type BerlekampWelchTest struct {
	*smtest.SmT

	code *FecCode
}

func NewBerlekampWelchTest(t testing.TB,
	required, total int) *BerlekampWelchTest {

	smt := smtest.Wrap(t)

	code, err := NewFecCode(required, total)
	smt.AssertNoError(err)

	return &BerlekampWelchTest{
		SmT: smt,

		code: code,
	}
}

func (t *BerlekampWelchTest) StoreShares() ([]Share, Callback) {
	out := make([]Share, t.code.n)
	return out, func(idx, total int, data []byte) {
		out[idx].Number = idx
		out[idx].Data = append(out[idx].Data, data...)
	}
}

func (t *BerlekampWelchTest) SomeShares(block int) []Share {
	// seed the initial data
	data := make([]byte, t.code.k*block)
	for i := range data {
		data[i] = byte(i + 1)
	}

	shares, store := t.StoreShares()
	t.AssertNoError(t.code.Encode(data, store))
	return shares
}

func (t *BerlekampWelchTest) CopyShares(shares []Share) []Share {
	out := make([]Share, t.code.n)
	for i, share := range shares {
		out[i].Number = share.Number
		out[i].Data = append([]byte(nil), share.Data...)
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
