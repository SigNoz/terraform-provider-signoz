package convtypes

import (
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func Float64Pointer(v types.Float64) *float64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	f := v.ValueFloat64()
	return &f
}

func Float64FromPointer(p *float64) types.Float64 {
	if p == nil {
		return types.Float64Null()
	}
	return types.Float64Value(*p)
}

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

func NumberFloat32FromPointer(p *float32) types.Number {
	if p == nil {
		return types.NumberNull()
	}
	return types.NumberValue(big.NewFloat(float64(*p)))
}

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

func NumberFloat64FromPointer(p *float64) types.Number {
	if p == nil {
		return types.NumberNull()
	}
	return types.NumberValue(big.NewFloat(*p))
}
