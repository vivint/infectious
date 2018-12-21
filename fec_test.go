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
	"fmt"
	"math/rand"
	"strconv"
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
		outputs[s.Number] = s.DeepCopy().Data
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

func TestEncodeSingle(t *testing.T) {
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
	for i := 0; i < total; i++ {
		outputs[i] = make([]byte, block)
		err = code.EncodeSingle(data[:], outputs[i], i)
		if err != nil {
			t.Fatalf("encode failed: %s", err)
		}
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

func BenchmarkEncodeSingle(b *testing.B) {
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

	output := make([]byte, block)

	b.SetBytes(block * required)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < total; j++ {
			code.EncodeSingle(data, output, j)
		}
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

func BenchmarkMultiple(b *testing.B) {
	b.ReportAllocs()
	data := make([]byte, 8<<20)
	output := make([]byte, 8<<20)

	confs := []struct{ required, total int }{
		{2, 4},
		{20, 50},
		{30, 60},
		{50, 80},
	}

	dataSizes := []int{
		64,
		128,
		256,
		512,
		1 << 10,
		256 << 10,
		1 << 20,
		5 << 20,
		8 << 20,
	}

	for _, conf := range confs {
		confname := fmt.Sprintf("r%dt%d/", conf.required, conf.total)
		for _, tryDataSize := range dataSizes {
			dataSize := (tryDataSize / conf.required) * conf.required
			var testname string
			if dataSize < 1024 {
				testname = strconv.Itoa(dataSize) + "B"
			} else {
				testname = strconv.Itoa(dataSize/1024) + "KB"
			}
			fec, _ := NewFEC(conf.required, conf.total)

			b.Run("Encode/"+confname+testname, func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(dataSize))
				for i := 0; i < b.N; i++ {
					err := fec.Encode(data[:dataSize], func(share Share) {})
					if err != nil {
						b.Fatal(err)
					}
				}
			})

			b.Run("Single/"+confname+testname, func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(dataSize))
				output := make([]byte, dataSize/conf.required)
				for i := 0; i < b.N; i++ {
					for j := 0; j < conf.total; j++ {
						err := fec.EncodeSingle(data[:dataSize], output, j)
						if err != nil {
							b.Fatal(err)
						}
					}
				}
			})

			shares := []Share{}
			err := fec.Encode(data[:dataSize], func(share Share) {
				shares = append(shares, share.DeepCopy())
			})
			if err != nil {
				b.Fatal(err)
			}

			b.Run("Decode/"+confname+testname, func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(dataSize))
				for i := 0; i < b.N; i++ {
					rand.Shuffle(len(shares), func(i, k int) {
						shares[i], shares[k] = shares[k], shares[i]
					})

					offset := i % (conf.total / 4)
					n := conf.required + 1 + offset
					if n > conf.total {
						n = conf.total
					}

					_, err = fec.Decode(output[:dataSize], shares[:n])
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
