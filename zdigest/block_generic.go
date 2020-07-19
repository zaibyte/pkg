package zdigest

import (
	"crypto/aes"
	"encoding/binary"

	"github.com/templexxx/xorsimd"
)

func blockGeneric(d *state, p []byte) {

	for len(p) >= BlockSize {
		for i := 0; i < BlockSize; i += aes.BlockSize {
			aesenc(p[i:i+aes.BlockSize], d[i:i+aes.BlockSize], d[i:i+aes.BlockSize])
		}
		p = p[BlockSize:]
	}
}

func finalGeneric(d *state, l uint64) uint32 {

	bs := aes.BlockSize
	p := make([]byte, BlockSize)
	for i := 0; i < 4; i++ {
		binary.LittleEndian.PutUint64(p[aes.BlockSize*i:], l)
	}
	for i := range p {
		d[i] = d[i] ^ p[i]
	}

	aesenc(d[bs:], d[:], d[:])
	aesenc(d[3*bs:], d[bs:], d[2*bs:])
	aesenc(d[bs:], d[:], d[:])

	aesenc(d[:], d[:], d[:])
	aesenc(d[:], d[:], d[:])
	aesenc(d[:], d[:], d[:])

	xorsimd.Bytes8(d[:], d[:8], d[8:])
	s64 := binary.LittleEndian.Uint64(d[:8])
	return uint32(s64>>4 ^ s64)
}

func aesenc(key, dst, src []byte) {
	_ = src[15] // early bounds check
	s0 := binary.BigEndian.Uint32(src[0:4])
	s1 := binary.BigEndian.Uint32(src[4:8])
	s2 := binary.BigEndian.Uint32(src[8:12])
	s3 := binary.BigEndian.Uint32(src[12:16])

	k0 := binary.BigEndian.Uint32(key[0:4])
	k1 := binary.BigEndian.Uint32(key[4:8])
	k2 := binary.BigEndian.Uint32(key[8:12])
	k3 := binary.BigEndian.Uint32(key[12:16])

	t0 := k0 ^ te0[uint8(s0>>24)] ^ te1[uint8(s1>>16)] ^ te2[uint8(s2>>8)] ^ te3[uint8(s3)]
	t1 := k1 ^ te0[uint8(s1>>24)] ^ te1[uint8(s2>>16)] ^ te2[uint8(s3>>8)] ^ te3[uint8(s0)]
	t2 := k2 ^ te0[uint8(s2>>24)] ^ te1[uint8(s3>>16)] ^ te2[uint8(s0>>8)] ^ te3[uint8(s1)]
	t3 := k3 ^ te0[uint8(s3>>24)] ^ te1[uint8(s0>>16)] ^ te2[uint8(s1>>8)] ^ te3[uint8(s2)]

	_ = dst[15] // early bounds check
	binary.BigEndian.PutUint32(dst[0:4], t0)
	binary.BigEndian.PutUint32(dst[4:8], t1)
	binary.BigEndian.PutUint32(dst[8:12], t2)
	binary.BigEndian.PutUint32(dst[12:16], t3)
}

func sum32Generic(p []byte) uint32 {
	nn := uint64(len(p))

	h := state{}
	copy(h[:], initState[:])

	if len(p) >= BlockSize {
		n := len(p) &^ (BlockSize - 1)
		block(&h, p[:n])
		p = p[n:]
	}

	remain := len(p)

	if remain > 0 {
		buf := [BlockSize]byte{}
		copy(buf[:], p)
		block(&h, buf[:])
	}

	return final(&h, nn)
}
