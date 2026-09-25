package strategyprojection

// a115 — 「reader 자리가 비었는가」는 **한 곳에서만** 판정한다 (design D2, freeze 리뷰 P1-1).
//
// 행동 시험만으로는 판정 두 벌을 못 잡는다: 소비자가 이 파일의 판정을 옮겨 적으면 오늘은 같은 답을
// 한다(뮤테이션 원장 K13 — httpapi 가 위임 대신 사본을 둬도 스위트가 초록이었다). 갈라지는 것은 내일이다.
// 그래서 **구조**를 센다: 모듈의 생산 Go 파일 전부에서 presence 를 묻는 두 모양 — presence 인터페이스로의
// 타입 단언과 `StrategyRuntimeConfigured()` 호출 — 이 이 패키지 밖에 하나도 없어야 한다.
//
// 세는 범위는 함수 하나가 아니라 **모듈 전체**다(존재 검사는 역할 검사가 아니다 — 헬퍼 하나로 함수 범위는
// 무너진다). 시험 파일은 뺀다: a109 시험들은 wrapper 의 술어를 직접 부르는 것이 목적이다.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

func TestTheAbsenceIsJudgedInOnePlace(t *testing.T) {
	root := filepath.Join("..", "..")
	own, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	var files, asks int
	var offenders []string
	for _, top := range []string{"cmd", "internal"} {
		err := filepath.WalkDir(filepath.Join(root, top), func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			abs, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
			if err != nil {
				return err
			}
			files++
			ast.Inspect(file, func(node ast.Node) bool {
				var what string
				switch n := node.(type) {
				case *ast.TypeAssertExpr:
					if presenceType(n.Type) {
						what = "presence 타입 단언"
					}
				case *ast.CallExpr:
					if sel, ok := n.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "StrategyRuntimeConfigured" {
						what = "StrategyRuntimeConfigured() 호출"
					}
				}
				if what == "" {
					return true
				}
				asks++
				if filepath.Dir(abs) != own {
					offenders = append(offenders, path+": "+what)
				}
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	// 계측기가 눈멀지 않았는지(양성 대조군): 이 패키지의 판정 자신은 두 모양을 하나씩 갖는다.
	if files < 100 || asks < 2 {
		t.Fatalf("센 파일 %d · 이 패키지의 presence 질문 %d — 계측기가 모듈을 못 봤다", files, asks)
	}
	if len(offenders) > 0 {
		t.Errorf("부재 판정이 이 패키지 밖에서 다시 쓰였다 — 판정은 StrategyRuntimeAbsent 한 벌이어야 한다:\n  %s",
			strings.Join(offenders, "\n  "))
	}
}

// presenceType 은 타입 단언의 대상이 presence 인터페이스인가다 — 이름으로(한정자 유무 무관) 또는 같은
// 메서드를 적은 익명 인터페이스로.
func presenceType(expr ast.Expr) bool {
	switch n := expr.(type) {
	case *ast.Ident:
		return n.Name == "StrategyRuntimePresence"
	case *ast.SelectorExpr:
		return n.Sel.Name == "StrategyRuntimePresence"
	case *ast.InterfaceType:
		for _, method := range n.Methods.List {
			for _, name := range method.Names {
				if name.Name == "StrategyRuntimeConfigured" {
					return true
				}
			}
		}
	}
	return false
}
