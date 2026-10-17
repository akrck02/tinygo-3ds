package compiler

import "tinygo.org/x/go-llvm"

// createFloatMinMax implements the min and max builtins for floating point
// values with compares and selects, with the same edge cases as the
// llvm.minimum.* and llvm.maximum.* intrinsics: a NaN argument gives NaN, and
// min(-0.0, +0.0) is -0.0 (max is +0.0). It is used where LLVM cannot select
// those intrinsics.
func (b *builder) createFloatMinMax(callName string, floatType llvm.Type, argValues []llvm.Value) llvm.Value {
	intType := b.ctx.Int32Type()
	if floatType.TypeKind() == llvm.DoubleTypeKind {
		intType = b.ctx.Int64Type()
	}
	result := argValues[0]
	for _, arg := range argValues[1:] {
		result = b.createFloatMinMaxPair(callName == "min", floatType, intType, result, arg)
	}
	return result
}

func (b *builder) createFloatMinMaxPair(isMin bool, floatType, intType llvm.Type, x, y llvm.Value) llvm.Value {
	// The smaller (or larger) of two different numbers.
	predicate := llvm.FloatOLT
	if !isMin {
		predicate = llvm.FloatOGT
	}
	picked := b.CreateSelect(b.CreateFCmp(predicate, x, y, ""), x, y, "")
	// Equal numbers can still differ in their sign when they are zeros: the
	// minimum of the two zeros has the sign bit if either has it, the maximum
	// only if both have it.
	xBits := b.CreateBitCast(x, intType, "")
	yBits := b.CreateBitCast(y, intType, "")
	var joined llvm.Value
	if isMin {
		joined = b.CreateOr(xBits, yBits, "")
	} else {
		joined = b.CreateAnd(xBits, yBits, "")
	}
	equal := b.CreateFCmp(llvm.FloatOEQ, x, y, "")
	picked = b.CreateSelect(equal, b.CreateBitCast(joined, floatType, ""), picked, "")
	// A NaN on either side is the result.
	picked = b.CreateSelect(b.CreateFCmp(llvm.FloatUNO, y, y, ""), y, picked, "")
	return b.CreateSelect(b.CreateFCmp(llvm.FloatUNO, x, x, ""), x, picked, "")
}
