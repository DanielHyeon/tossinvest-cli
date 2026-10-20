package offlinecontinuation

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestAssessmentHasNoOperationalDependencies(t *testing.T) {
	file := filepath.Join("assessment.go")
	parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatal(err)
	}
	for _, spec := range parsed.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{"/journal", "/execgw", "/broker", "/engine", "/trading", "/gateway", "net/http", "os/exec"} {
			if strings.Contains(path, forbidden) {
				t.Fatalf("operational dependency %q", path)
			}
		}
	}
}

func TestResultExportsNoNativeAuthorityTypes(t *testing.T) {
	typ := reflect.TypeOf(Result{})
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		base := field.Type
		for base.Kind() == reflect.Pointer || base.Kind() == reflect.Slice || base.Kind() == reflect.Array || base.Kind() == reflect.Map {
			base = base.Elem()
		}
		if base.PkgPath() != "" {
			t.Fatalf("Result.%s exposes non-primitive type %s", field.Name, field.Type)
		}
	}
}
