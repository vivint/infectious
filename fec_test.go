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

func TestBasicOperation(t *testing.T) {
	const block = 1024 * 1024
	const total, required = 40, 20

	code, err := NewFEC(required, total)
	if err != nil {
		t.Fatalf("failed to create new fec code: %s", err)
	}

	// seed the initial data
	data := make([]byte, required*block)
	for i := range data {
		data[i] = byte(i)
	}

	// encode it and store to outputs
	var outputs = make(map[int][]byte)
	store := func(s Share) {
		outputs[s.Number] = append([]byte(nil), s.Data...)
	}
	err = code.Encode(data[:], store)
	if err != nil {
		t.Fatalf("encode failed: %s", err)
	}

	// pick required of the total shares randomly
	var shares [required]Share
	for i, idx := range rand.Perm(total)[:required] {
		shares[i].Number = idx
		shares[i].Data = outputs[idx]
	}

	got := make([]byte, required*block)
	record := func(s Share) {
		copy(got[s.Number*block:], s.Data)
	}
	err = code.Rebuild(shares[:], record)
	if err != nil {
		t.Fatalf("decode failed: %s", err)
	}

	if !bytes.Equal(got, data) {
		t.Fatalf("did not match")
	}
}

func BenchmarkEncode(b *testing.B) {
	const block = 1024 * 1024
	const total, required = 40, 20

	code, err := NewFEC(required, total)
	if err != nil {
		b.Fatalf("failed to create new fec code: %s", err)
	}

	// seed the initial data
	data := make([]byte, required*block)
	for i := range data {
		data[i] = byte(i)
	}
	store := func(Share) {}

	b.SetBytes(block * required)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		code.Encode(data, store)
	}
}

func BenchmarkRebuild(b *testing.B) {
	const block = 4096
	const total, required = 40, 20

	code, err := NewFEC(required, total)
	if err != nil {
		b.Fatalf("failed to create new fec code: %s", err)
	}

	// seed the initial data
	data := make([]byte, required*block)
	for i := range data {
		data[i] = byte(i)
	}
	shares := make([]Share, total)
	store := func(s Share) {
		idx := s.Number
		shares[idx].Number = idx
		shares[idx].Data = append(shares[idx].Data, s.Data...)
	}
	err = code.Encode(data, store)
	if err != nil {
		b.Fatalf("failed to encode: %s", err)
	}

	dec_shares := shares[total-required:]

	b.SetBytes(block * required)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		code.Rebuild(dec_shares, nil)
	}
}
