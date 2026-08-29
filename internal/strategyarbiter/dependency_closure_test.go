package strategyarbiter

import (
	"go/parser"
	"go/token"
	"sort"
	"strconv"
	"testing"
)

// 중재자는 아무것도 바꾸지 않는다. 그 약속은 주석이 아니라 import 목록이
// 지킨다 — 주문 게이트웨이나 쓰기 가능한 원장을 들여오는 순간 이 테스트가 깨진다.
func TestTheArbiterCannotReachAnyMutationCapability(t *testing.T) {
	allowed := map[string]bool{
		"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow":   true,
		"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter": true,
		"time": true,
	}
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, ".", nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse package directory: %v", err)
	}
	unexpected := make([]string, 0)
	for name, pkg := range packages {
		for path, file := range pkg.Files {
			// 테스트 파일은 이 약속의 대상이 아니다. 생산 코드만 본다.
			if len(path) > 8 && path[len(path)-8:] == "_test.go" {
				continue
			}
			if name != "strategyarbiter" {
				t.Fatalf("%s declares package %q", path, name)
			}
			for _, spec := range file.Imports {
				value, err := strconv.Unquote(spec.Path.Value)
				if err != nil {
					t.Fatalf("%s: unquote %s: %v", path, spec.Path.Value, err)
				}
				if !allowed[value] {
					unexpected = append(unexpected, path+" -> "+value)
				}
			}
		}
	}
	if len(unexpected) != 0 {
		sort.Strings(unexpected)
		t.Fatalf("the arbiter imports something outside its allowed closure: %v", unexpected)
	}
}
