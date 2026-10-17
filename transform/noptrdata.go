package transform

// This file moves large globals that hold no pointers to a section of their own, so a garbage
// collector that scans the data segment conservatively can leave them out. Embedded files and
// tables (a //go:embed []byte, a font) are such globals: scanned, their bytes look like pointers
// now and then, and every false pointer into a large heap object is slow to follow.

import (
	"tinygo.org/x/go-llvm"
)

// PointerFreeSection is the section MovePointerFreeGlobals puts globals in; a target's linker
// script places it apart from the data the collector scans.
const PointerFreeSection = ".noptrdata"

// pointerFreeMinimumSize is how large a global must be to move: small ones are not worth it.
const pointerFreeMinimumSize = 4096

// MovePointerFreeGlobals puts every writable, initialized global of at least
// pointerFreeMinimumSize bytes that cannot hold a pointer in PointerFreeSection: its type has no
// pointers and its initial value no addresses (an address may be kept in an integer, as a
// ptrtoint constant). Globals already placed in a section, thread-local ones and zeroed ones (in
// .bss, small on the targets that use this) stay where they are.
func MovePointerFreeGlobals(mod llvm.Module) {
	targetData := llvm.NewTargetData(mod.DataLayout())
	defer targetData.Dispose()
	for global := mod.FirstGlobal(); !global.IsNil(); global = llvm.NextGlobal(global) {
		if global.IsDeclaration() || global.IsGlobalConstant() || global.IsThreadLocal() || global.Section() != "" {
			continue
		}
		initializer := global.Initializer()
		if initializer.IsNil() || initializer.IsNull() {
			continue
		}
		valueType := global.GlobalValueType()
		if targetData.TypeAllocSize(valueType) < pointerFreeMinimumSize || typeHasPointers(valueType) || valueHasAddresses(initializer) {
			continue
		}
		global.SetSection(PointerFreeSection)
	}
}

// typeHasPointers tells whether a value of a type can hold a pointer.
func typeHasPointers(valueType llvm.Type) bool {
	switch valueType.TypeKind() {
	case llvm.PointerTypeKind:
		return true
	case llvm.StructTypeKind:
		for _, element := range valueType.StructElementTypes() {
			if typeHasPointers(element) {
				return true
			}
		}
		return false
	case llvm.ArrayTypeKind, llvm.VectorTypeKind:
		return typeHasPointers(valueType.ElementType())
	}
	return false
}

// valueHasAddresses tells whether a constant is, or holds, an expression such as the address of
// a global turned into an integer. Plain numbers (bytes, integers, floats) have no operands.
func valueHasAddresses(value llvm.Value) bool {
	if !value.IsAConstantExpr().IsNil() || !value.IsAGlobalValue().IsNil() {
		return true
	}
	for index := 0; index < value.OperandsCount(); index++ {
		if valueHasAddresses(value.Operand(index)) {
			return true
		}
	}
	return false
}
