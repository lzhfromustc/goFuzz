// Generated by genwrap.go. DO NOT EDIT

package z3

import "runtime"

/*
#cgo LDFLAGS: -lz3
#include <z3.h>
#include <stdlib.h>
*/
import "C"

// Eq returns a Value that is true if l and r are equal.
func (l Int) Eq(r Int) Bool {
	ctx := l.ctx
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_eq(ctx.c, l.c, r.c)
	})
	runtime.KeepAlive(l)
	runtime.KeepAlive(r)
	return Bool(val)
}

// NE returns a Value that is true if l and r are not equal.
func (l Int) NE(r Int) Bool {
	return l.ctx.Distinct(l, r)
}

// Div returns the floor of l / r.
//
// If r is 0, the result is unconstrained.
//
// Note that this differs from Go division: Go rounds toward zero
// (truncated division), whereas this rounds toward -inf.
func (l Int) Div(r Int) Int {
	// Generated from int.go:68.
	ctx := l.ctx
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_div(ctx.c, l.c, r.c)
	})
	runtime.KeepAlive(l)
	runtime.KeepAlive(r)
	return Int(val)
}

// Mod returns modulus of l / r.
//
// The sign of the result follows the sign of r.
func (l Int) Mod(r Int) Int {
	// Generated from int.go:74.
	ctx := l.ctx
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_mod(ctx.c, l.c, r.c)
	})
	runtime.KeepAlive(l)
	runtime.KeepAlive(r)
	return Int(val)
}

// Rem returns remainder of l / r.
//
// The sign of the result follows the sign of l.
//
// Note that this differs subtly from Go's remainder operator because
// this is based floored division rather than truncated division.
func (l Int) Rem(r Int) Int {
	// Generated from int.go:83.
	ctx := l.ctx
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_rem(ctx.c, l.c, r.c)
	})
	runtime.KeepAlive(l)
	runtime.KeepAlive(r)
	return Int(val)
}

// ToReal converts l to sort Real.
func (l Int) ToReal() Real {
	// Generated from int.go:87.
	ctx := l.ctx
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_int2real(ctx.c, l.c)
	})
	runtime.KeepAlive(l)
	return Real(val)
}

// ToBV converts l to a bit-vector of width bits.
func (l Int) ToBV(bits int) BV {
	// Generated from int.go:91.
	ctx := l.ctx
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_int2bv(ctx.c, C.unsigned(bits), l.c)
	})
	runtime.KeepAlive(l)
	return BV(val)
}

// Add returns the sum l + r[0] + r[1] + ...
func (l Int) Add(r ...Int) Int {
	// Generated from intreal.go:12.
	ctx := l.ctx
	cargs := make([]C.Z3_ast, len(r)+1)
	cargs[0] = l.c
	for i, arg := range r {
		cargs[i+1] = arg.c
	}
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_add(ctx.c, C.uint(len(cargs)), &cargs[0])
	})
	runtime.KeepAlive(&cargs[0])
	return Int(val)
}

// Mul returns the product l * r[0] * r[1] * ...
func (l Int) Mul(r ...Int) Int {
	// Generated from intreal.go:16.
	ctx := l.ctx
	cargs := make([]C.Z3_ast, len(r)+1)
	cargs[0] = l.c
	for i, arg := range r {
		cargs[i+1] = arg.c
	}
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_mul(ctx.c, C.uint(len(cargs)), &cargs[0])
	})
	runtime.KeepAlive(&cargs[0])
	return Int(val)
}

// Sub returns l - r[0] - r[1] - ...
func (l Int) Sub(r ...Int) Int {
	// Generated from intreal.go:20.
	ctx := l.ctx
	cargs := make([]C.Z3_ast, len(r)+1)
	cargs[0] = l.c
	for i, arg := range r {
		cargs[i+1] = arg.c
	}
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_sub(ctx.c, C.uint(len(cargs)), &cargs[0])
	})
	runtime.KeepAlive(&cargs[0])
	return Int(val)
}

// Neg returns -l.
func (l Int) Neg() Int {
	// Generated from intreal.go:24.
	ctx := l.ctx
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_unary_minus(ctx.c, l.c)
	})
	runtime.KeepAlive(l)
	return Int(val)
}

// Exp returns l???.
func (l Int) Exp(r Int) Int {
	// Generated from intreal.go:28.
	ctx := l.ctx
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_power(ctx.c, l.c, r.c)
	})
	runtime.KeepAlive(l)
	runtime.KeepAlive(r)
	return Int(val)
}

// LT returns l < r.
func (l Int) LT(r Int) Bool {
	// Generated from intreal.go:32.
	ctx := l.ctx
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_lt(ctx.c, l.c, r.c)
	})
	runtime.KeepAlive(l)
	runtime.KeepAlive(r)
	return Bool(val)
}

// LE returns l <= r.
func (l Int) LE(r Int) Bool {
	// Generated from intreal.go:36.
	ctx := l.ctx
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_le(ctx.c, l.c, r.c)
	})
	runtime.KeepAlive(l)
	runtime.KeepAlive(r)
	return Bool(val)
}

// GT returns l > r.
func (l Int) GT(r Int) Bool {
	// Generated from intreal.go:40.
	ctx := l.ctx
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_gt(ctx.c, l.c, r.c)
	})
	runtime.KeepAlive(l)
	runtime.KeepAlive(r)
	return Bool(val)
}

// GE returns l >= r.
func (l Int) GE(r Int) Bool {
	// Generated from intreal.go:44.
	ctx := l.ctx
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_ge(ctx.c, l.c, r.c)
	})
	runtime.KeepAlive(l)
	runtime.KeepAlive(r)
	return Bool(val)
}
