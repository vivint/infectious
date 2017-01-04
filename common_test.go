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
	crand "crypto/rand"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func RandomBytes(size int) []byte {
	buf := make([]byte, size)
	_, err := crand.Read(buf)
	if err != nil {
		panic(fmt.Sprintf("rand.Read failed: %s", err))
	}
	return buf
}

// we want all of our test runs to be with a different seed
func init() { rand.Seed(int64(time.Now().UnixNano())) }

type Asserter struct {
	tb testing.TB
}

func Wrap(tb testing.TB) *Asserter {
	return &Asserter{
		tb: tb,
	}
}

func (a *Asserter) AssertNoError(err error) {
	if err != nil {
		a.tb.Fatalf("expected no error; got %v", err)
	}
}

func (a *Asserter) AssertDeepEqual(x, y interface{}) {
	if !reflect.DeepEqual(x, y) {
		a.tb.Fatalf("expected\n%#v\n%#v\nto be equal", x, y)
	}
}
