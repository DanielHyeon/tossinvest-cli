package execgw

import (
	"math"
	"testing"
)

func TestCanonicalProtectionQuantityRejectsFractionalNonFiniteAndUnsafeFloatValues(t *testing.T) {
	for _, value := range []float64{0, -1, 1.5, math.Inf(1), math.Inf(-1), math.NaN(), float64(uint64(1) << 53)} {
		if quantity, ok := canonicalProtectionQuantity(value); ok || quantity != 0 {
			t.Fatalf("value %v canonicalized as %d", value, quantity)
		}
	}
	for _, value := range []float64{1, 50, float64((uint64(1) << 53) - 1)} {
		quantity, ok := canonicalProtectionQuantity(value)
		if !ok || quantity != uint64(value) {
			t.Fatalf("value %v refused quantity=%d ok=%v", value, quantity, ok)
		}
	}
}
