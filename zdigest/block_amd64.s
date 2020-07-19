#include "textflag.h"

// func blockAESNI(s *state, p []byte)
TEXT ·blockAESNI(SB), 0, $0-32
	MOVQ	s+0(FP), DI
	MOVQ	p_base+8(FP), SI
	MOVQ	p_len+16(FP), AX

	VMOVDQU	0(DI), X0
	VMOVDQU	16(DI), X1
	VMOVDQU	32(DI), X2
	VMOVDQU	48(DI), X3

loop:
	VAESENC	0(SI),  X0, X0
	VAESENC	16(SI), X1, X1
	VAESENC	32(SI), X2, X2
	VAESENC	48(SI), X3, X3

	ADDQ    $64, SI
	SUBQ	$64, AX
	CMPQ	AX, $64
	JGE	loop

	VMOVDQU	X0, 0(DI)
	VMOVDQU	X1, 16(DI)
	VMOVDQU	X2, 32(DI)
	VMOVDQU	X3, 48(DI)

	VZEROUPPER
	RET

// func finalAESNI(s *state, l uint64) uint32
TEXT ·finalAESNI(SB), 0, $0-16
	MOVQ	s+0(FP), DI
	MOVQ	l+8(FP), SI

	MOVQ	SI, X4

	VMOVDQU	0(DI), X0
	VMOVDQU	16(DI), X1
	VMOVDQU	32(DI), X2
	VMOVDQU	48(DI), X3

	VPXOR	X4, X0, X0
	VPXOR	X4, X1, X1
	VPXOR	X4, X2, X2
	VPXOR	X4, X3, X3

	VAESENC	X1, X0, X0
	VAESENC	X3, X2, X1
	VAESENC	X1, X0, X0

	VAESENC	X0, X0, X0
	VAESENC	X0, X0, X0
	VAESENC	X0, X0, X0
	VMOVDQU	X0, 0(DI)

	MOVQ    X0, R8
	VPEXTRQ $1, X0, R9
	XORQ    R8, R9
	MOVQ    R9, R10
    SHRQ    $4, R9
    XORQ    R9, R10
    MOVL    R10, r+16(FP)

	VZEROUPPER
	RET

#define n R11
#define maska X1
#define maskb X2
#define mask9 X10

// func sum32AESNIWithPadding(p *byte, n int, padding *byte) uint32
TEXT ·sum32AESNIWithPadding(SB), 0, $0
	MOVQ	p_addr+0(FP), SI
	MOVQ	p_n+8(FP), AX
	MOVQ    padding+16(FP), BX
	// Load init state.
	MOVQ	·initStateBytes(SB), DI
	VMOVDQU	0(DI), X0
	VMOVDQU	16(DI), X1
	VMOVDQU	32(DI), X2
	VMOVDQU	48(DI), X3

    MOVQ    AX, n
    MOVQ    AX, R10
    SHRQ    $6, R10       // n / 64.
    TESTQ   R10, R10
    JZ      remain           // If 0, remain.

loop_block:
	VAESENC	0(SI),  X0, X0
	VAESENC	16(SI), X1, X1
	VAESENC	32(SI), X2, X2
	VAESENC	48(SI), X3, X3

	ADDQ    $64, SI
	SUBQ    $64, AX
	SUBQ	$1, R10
	JNE	    loop_block

remain:
    CMPQ  AX, $0
    JE    final // len(p) is aligned to 64, no padding.

padding:
    VAESENC	0(BX),  X0, X0
	VAESENC	16(BX), X1, X1
	VAESENC	32(BX), X2, X2
	VAESENC	48(BX), X3, X3

final:
	MOVQ	n, X4

	VPXOR	X4, X0, X0
	VPXOR	X4, X1, X1
	VPXOR	X4, X2, X2
	VPXOR	X4, X3, X3

	VAESENC	X1, X0, X0
	VAESENC	X3, X2, X1
	VAESENC	X1, X0, X0

	VAESENC	X0, X0, X0
	VAESENC	X0, X0, X0
	VAESENC	X0, X0, X0

	MOVQ    X0, R8
	VPEXTRQ $1, X0, R9
	ADDQ    R8, R9
	MOVQ    R9, R12
    SHRQ    $4, R9
    XORQ    R9, R12
    MOVL    R12, r+24(FP)


	VZEROUPPER
	RET

// func sum32AESNINoPadding(p *byte, n int) uint32
TEXT ·sum32AESNINoPadding(SB), 0, $0
	MOVQ	p_addr+0(FP), SI
	MOVQ	p_n+8(FP), AX
	// Load init state.
	MOVQ	·initStateBytes(SB), DI
	VMOVDQU	0(DI), X0
	VMOVDQU	16(DI), X1
	VMOVDQU	32(DI), X2
	VMOVDQU	48(DI), X3

    MOVQ    AX, n
    SHRQ    $6, AX       // n / 64.

loop_block:
	VAESENC	0(SI),  X0, X0
	VAESENC	16(SI), X1, X1
	VAESENC	32(SI), X2, X2
	VAESENC	48(SI), X3, X3

	ADDQ    $64, SI
	SUBQ	$1, AX
	JNE	    loop_block

	MOVQ	n, X4

	VPXOR	X4, X0, X0
	VPXOR	X4, X1, X1
	VPXOR	X4, X2, X2
	VPXOR	X4, X3, X3

	VAESENC	X1, X0, X0
	VAESENC	X3, X2, X1
	VAESENC	X1, X0, X0

	VAESENC	X0, X0, X0
	VAESENC	X0, X0, X0
	VAESENC	X0, X0, X0

	MOVQ    X0, R8
	VPEXTRQ $1, X0, R9
	ADDQ    R8, R9
	MOVQ    R9, R12
    SHRQ    $4, R9
    XORQ    R9, R12
    MOVL    R12, r+16(FP)



	VZEROUPPER
	RET
