// Copyright (C) 2016 Space Monkey, Inc.

package infectuous

import "sort"

type FecCode struct {
	k          int
	n          int
	enc_matrix []byte
}

func NewFecCode(k, n int) (*FecCode, error) {
	if k <= 0 || n <= 0 || k > 256 || n > 256 || k > n {
		return nil, Error.New("requires 1 <= k <= n <= 256")
	}

	enc_matrix := make([]byte, n*k)
	temp_matrix := make([]byte, n*k)
	createInvertedVdm(temp_matrix, k)

	for i := k * k; i < len(temp_matrix); i++ {
		temp_matrix[i] = gf_exp[((i/k)*(i%k))%255]
	}

	for i := 0; i < k; i++ {
		enc_matrix[i*(k+1)] = 1
	}

	for row := k * k; row < n*k; row += k {
		for col := 0; col < k; col++ {
			pa := temp_matrix[row:]
			pb := temp_matrix[col:]
			acc := byte(0)
			for i := 0; i < k; i, pa, pb = i+1, pa[1:], pb[k:] {
				acc ^= gf_mul_table[pa[0]][pb[0]]
			}
			enc_matrix[row+col] = acc
		}
	}

	return &FecCode{
		k:          k,
		n:          n,
		enc_matrix: enc_matrix,
	}, nil
}

func (f *FecCode) Required() int {
	return f.k
}

func (f *FecCode) Total() int {
	return f.n
}

type Callback func(i int, n int, data []byte)

func (f *FecCode) Encode(input []byte, output Callback) error {
	size := len(input)

	k := f.k
	n := f.n
	enc_matrix := f.enc_matrix

	if size%k != 0 {
		return Error.New("input length must be a multiple of %d", k)
	}

	block_size := size / k

	for i := 0; i < k; i++ {
		output(i, n, input[i*block_size:i*block_size+block_size])
	}

	fec_buf := make([]byte, block_size)
	for i := k; i < n; i++ {
		for j := range fec_buf {
			fec_buf[j] = 0
		}

		for j := 0; j < k; j++ {
			addmul(fec_buf, input[j*block_size:j*block_size+block_size],
				enc_matrix[i*k+j])
		}

		output(i, n, fec_buf)
	}
	return nil
}

type Share struct {
	Number int
	Data   []byte
}

type byNumber []Share

func (b byNumber) Len() int               { return len(b) }
func (b byNumber) Less(i int, j int) bool { return b[i].Number < b[j].Number }
func (b byNumber) Swap(i int, j int)      { b[i], b[j] = b[j], b[i] }

func (f *FecCode) Decode(shares []Share, output Callback) error {
	k := f.k
	n := f.n
	enc_matrix := f.enc_matrix

	if len(shares) < k {
		return Error.New("not enough shares")
	}

	share_size := len(shares[0].Data)
	sort.Sort(byNumber(shares))

	m_dec := make([]byte, k*k)
	indexes := make([]int, k)
	sharesv := make([][]byte, k)

	shares_b_iter := 0
	shares_e_iter := len(shares) - 1

	for i := 0; i < k; i++ {
		var share_id int
		var share_data []byte

		if share := shares[shares_b_iter]; share.Number == i {
			share_id = share.Number
			share_data = share.Data
			shares_b_iter++
		} else {
			share := shares[shares_e_iter]
			share_id = share.Number
			share_data = share.Data
			shares_e_iter--
		}

		if share_id >= n {
			return Error.New("invalid share id: %d", share_id)
		}

		if share_id < k {
			m_dec[i*(k+1)] = 1
			output(share_id, k, share_data)
		} else {
			copy(m_dec[i*k:i*k+k], enc_matrix[share_id*k:])
		}

		sharesv[i] = share_data
		indexes[i] = share_id
	}

	if err := invertMatrix(m_dec, k); err != nil {
		return err
	}

	buf := make([]byte, share_size)
	for i := 0; i < len(indexes); i++ {
		if indexes[i] >= k {
			for j := range buf {
				buf[j] = 0
			}

			for col := 0; col < k; col++ {
				addmul(buf, sharesv[col], m_dec[i*k+col])
			}

			output(i, k, buf)
		}
	}
	return nil
}
