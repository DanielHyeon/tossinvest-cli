package risk_test

import (
	"reflect"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
)

func TestAccountBaseFXHasNoCallerWritableFields(t *testing.T) {
	typeOf := reflect.TypeOf(risk.AccountBaseFX{})
	for i := 0; i < typeOf.NumField(); i++ {
		if typeOf.Field(i).PkgPath == "" {
			t.Fatalf("AccountBaseFX field %q is exported; raw FX must not be caller writable", typeOf.Field(i).Name)
		}
	}
}
