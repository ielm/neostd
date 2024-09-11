// This file contains the SipHash implementation originally from https://github.com/aead/siphash/tree/master
package hash

import (
	"encoding/binary"
	"hash"
	"strconv"
)

const (
	// KeySize is the size of the SipHash secret key in bytes.
	KeySize = 16
	// BlockSize is the block size of SipHash in bytes.
	BlockSize = 8
)

const (
	c0 = 0x736f6d6570736575
	c1 = 0x646f72616e646f6d
	c2 = 0x6c7967656e657261
	c3 = 0x7465646279746573
)

type KeySizeError uint

func (k KeySizeError) Error() string {
	return "siphash: invalid key size " + strconv.Itoa(int(k))
}

// Sum64 returns the 64 bit authenticator for msg using a 128 bit secret key.
func Sum64(msg []byte, key *[KeySize]byte) uint64 {
	k0 := binary.LittleEndian.Uint64(key[0:])
	k1 := binary.LittleEndian.Uint64(key[8:])

	var hVal [4]uint64
	hVal[0] = k0 ^ c0
	hVal[1] = k1 ^ c1
	hVal[2] = k0 ^ c2
	hVal[3] = k1 ^ c3

	n := len(msg)
	ctr := byte(n)

	if n >= BlockSize {
		n &= (^(BlockSize - 1))
		core(&hVal, msg[:n])
		msg = msg[n:]
	}

	var block [BlockSize]byte
	copy(block[:], msg)
	block[7] = ctr

	return finalize64(&hVal, &block)
}

// New64 returns a hash.Hash64 computing the SipHash-64 checksum.
// This function returns a non-nil error if len(key) != 16.
func New64(key []byte) (hash.Hash64, error) {
	if k := len(key); k != KeySize {
		return nil, KeySizeError(k)
	}
	h := new(digest64)
	h.key[0] = binary.LittleEndian.Uint64(key)
	h.key[1] = binary.LittleEndian.Uint64(key[8:])
	h.Reset()
	return h, nil
}

type digest64 struct {
	hVal  [4]uint64
	key   [2]uint64
	block [BlockSize]byte
	off   int
	ctr   byte
}

func (d *digest64) BlockSize() int { return BlockSize }

func (d *digest64) Size() int { return 8 }

func (d *digest64) Reset() {
	d.hVal[0] = d.key[0] ^ c0
	d.hVal[1] = d.key[1] ^ c1
	d.hVal[2] = d.key[0] ^ c2
	d.hVal[3] = d.key[1] ^ c3

	d.off = 0
	d.ctr = 0
}

func (d *digest64) Write(p []byte) (n int, err error) {
	n = len(p)
	d.ctr += byte(n)

	if d.off > 0 {
		dif := BlockSize - d.off
		if n < dif {
			d.off += copy(d.block[d.off:], p)
			return
		}
		copy(d.block[d.off:], p[:dif])
		core(&(d.hVal), d.block[:])
		p = p[dif:]
		d.off = 0
	}
	if nn := len(p) &^ (BlockSize - 1); nn >= BlockSize {
		core(&(d.hVal), p[:nn])
		p = p[nn:]
	}
	if len(p) > 0 {
		d.off = copy(d.block[:], p)
	}
	return n, nil
}

func (d *digest64) Sum64() uint64 {
	hVal := d.hVal
	block := d.block
	for i := d.off; i < BlockSize-1; i++ {
		block[i] = 0
	}
	block[7] = d.ctr
	return finalize64(&hVal, &block)
}

func (d *digest64) Sum(sum []byte) []byte {
	var out [8]byte
	binary.LittleEndian.PutUint64(out[:], d.Sum64())
	return append(sum, out[:]...)
}

func core(hVal *[4]uint64, msg []byte) {
	v0, v1, v2, v3 := hVal[0], hVal[1], hVal[2], hVal[3]

	for len(msg) > 0 {
		m := binary.LittleEndian.Uint64(msg)
		msg = msg[BlockSize:]

		v3 ^= m

		v0 += v1
		v1 = v1<<13 | v1>>(64-13)
		v1 ^= v0
		v0 = v0<<32 | v0>>(64-32)
		v2 += v3
		v3 = v3<<16 | v3>>(64-16)
		v3 ^= v2
		v0 += v3
		v3 = v3<<21 | v3>>(64-21)
		v3 ^= v0
		v2 += v1
		v1 = v1<<17 | v1>>(64-17)
		v1 ^= v2
		v2 = v2<<32 | v2>>(64-32)

		v0 += v1
		v1 = v1<<13 | v1>>(64-13)
		v1 ^= v0
		v0 = v0<<32 | v0>>(64-32)
		v2 += v3
		v3 = v3<<16 | v3>>(64-16)
		v3 ^= v2
		v0 += v3
		v3 = v3<<21 | v3>>(64-21)
		v3 ^= v0
		v2 += v1
		v1 = v1<<17 | v1>>(64-17)
		v1 ^= v2
		v2 = v2<<32 | v2>>(64-32)

		v0 ^= m
	}

	hVal[0], hVal[1], hVal[2], hVal[3] = v0, v1, v2, v3
}

func finalize64(hVal *[4]uint64, block *[BlockSize]byte) uint64 {
	v0, v1, v2, v3 := hVal[0], hVal[1], hVal[2], hVal[3]

	m := binary.LittleEndian.Uint64(block[:])
	v3 ^= m

	v0 += v1
	v1 = v1<<13 | v1>>(64-13)
	v1 ^= v0
	v0 = v0<<32 | v0>>(64-32)
	v2 += v3
	v3 = v3<<16 | v3>>(64-16)
	v3 ^= v2
	v0 += v3
	v3 = v3<<21 | v3>>(64-21)
	v3 ^= v0
	v2 += v1
	v1 = v1<<17 | v1>>(64-17)
	v1 ^= v2
	v2 = v2<<32 | v2>>(64-32)

	v0 += v1
	v1 = v1<<13 | v1>>(64-13)
	v1 ^= v0
	v0 = v0<<32 | v0>>(64-32)
	v2 += v3
	v3 = v3<<16 | v3>>(64-16)
	v3 ^= v2
	v0 += v3
	v3 = v3<<21 | v3>>(64-21)
	v3 ^= v0
	v2 += v1
	v1 = v1<<17 | v1>>(64-17)
	v1 ^= v2
	v2 = v2<<32 | v2>>(64-32)

	v0 ^= m

	v2 ^= 0xff
	for i := 0; i < 4; i++ {
		v0 += v1
		v1 = v1<<13 | v1>>(64-13)
		v1 ^= v0
		v0 = v0<<32 | v0>>(64-32)
		v2 += v3
		v3 = v3<<16 | v3>>(64-16)
		v3 ^= v2
		v0 += v3
		v3 = v3<<21 | v3>>(64-21)
		v3 ^= v0
		v2 += v1
		v1 = v1<<17 | v1>>(64-17)
		v1 ^= v2
		v2 = v2<<32 | v2>>(64-32)
	}
	return v0 ^ v1 ^ v2 ^ v3
}
