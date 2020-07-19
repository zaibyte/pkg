package zdigest

import "github.com/templexxx/cpu"

func init() {
	if cpu.X86.HasAES && cpu.X86.HasAVX { // After 2008, all Intel CPUs have.
		block = blockAESNI
		final = finalAESNI
		sum32 = sum32AESNI
	}
}

func sum32AESNI(p []byte) uint32 {

	n := len(p)
	aligned := n &^ blockSizeMask
	remain := n - aligned
	if remain > 0 {
		buf := [BlockSize]byte{}
		copy(buf[:], p[aligned:])
		return sum32AESNIWithPadding(&p[0], n, &buf[0])
	}

	return sum32AESNINoPadding(&p[0], n)
}

//go:noescape
func blockAESNI(s *state, p []byte)

//go:noescape
func finalAESNI(s *state, l uint64) uint32

//go:noescape
func sum32AESNIWithPadding(p *byte, n int, padding *byte) uint32

//go:noescape
func sum32AESNINoPadding(p *byte, n int) uint32
