package conv

import (
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// Float64Value (`types.Float64`) — direct float64 wire mapping
// ---------------------------------------------------------------------------

// Float64Pointer mirrors Int64Pointer for float64.
func Float64Pointer(v types.Float64) *float64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	f := v.ValueFloat64()
	return &f
}

// Float64FromPointer mirrors Int64FromPointer for float64.
func Float64FromPointer(p *float64) types.Float64 {
	if p == nil {
		return types.Float64Null()
	}
	return types.Float64Value(*p)
}

// ---------------------------------------------------------------------------
// NumberValue (`types.Number`) — big.Float-backed; coerces to float32/float64
// ---------------------------------------------------------------------------
//
// `types.Number` stores a `*big.Float` so it can losslessly represent
// either float32 or float64. oapi-codegen-emitted apitypes use `*float32`
// for OpenAPI `format: float` and `*float64` for `format: double` — we
// project the big.Float onto whichever the wire schema declares. The
// projection is best-effort (big.Float → float32 may lose precision)
// but matches what hand-written code does today.

// NumberFloat32Pointer extracts a `*float32` from `types.Number`.
// Null/unknown → nil; null big.Float → nil.
func NumberFloat32Pointer(v types.Number) *float32 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	bf := v.ValueBigFloat()
	if bf == nil {
		return nil
	}
	f, _ := bf.Float32()
	return &f
}

// NumberFloat32FromPointer is the inverse of NumberFloat32Pointer.
func NumberFloat32FromPointer(p *float32) types.Number {
	if p == nil {
		return types.NumberNull()
	}
	return types.NumberValue(big.NewFloat(float64(*p)))
}

// NumberFloat64Pointer extracts a `*float64` from `types.Number`.
func NumberFloat64Pointer(v types.Number) *float64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	bf := v.ValueBigFloat()
	if bf == nil {
		return nil
	}
	f, _ := bf.Float64()
	return &f
}

// NumberFloat64FromPointer is the inverse of NumberFloat64Pointer.
func NumberFloat64FromPointer(p *float64) types.Number {
	if p == nil {
		return types.NumberNull()
	}
	return types.NumberValue(big.NewFloat(*p))
}
