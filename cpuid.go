// The MIT License (MIT)
//
// Copyright (C) 2016 Space Monkey, Inc.
// Copyright (c) 2015 Klaus Post
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

package infectuous

var hasAVX2, hasSSE3 bool

func init() {
	mfi, _, _, _ := cpuidex(0, 0)
	if mfi < 0x1 {
		return
	}
	_, _, c, _ := cpuidex(1, 0)
	if (c & 1) != 0 {
		hasSSE3 = true
	}
	if mfi >= 7 && c&(1<<26) != 0 && c&(1<<27) != 0 && c&(1<<28) != 0 {
		eax, _ := xgetbv(0)
		if (eax & 0x6) == 0x6 {
			_, ebx, _, _ := cpuidex(7, 0)
			if (ebx & 0x00000020) != 0 {
				hasAVX2 = true
			}
		}
	}
}
