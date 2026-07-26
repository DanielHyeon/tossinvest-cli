package costs

// sellgate_test.go pins task 1.2: the cost model is **not** a liquidation gate,
// and the break-even price it feeds cannot run away.
//
// Both halves are safety invariants rather than behaviour, so both are pinned
// structurally — a behavioural test only covers the call sites that exist today.
//
//  1. **§0.3 / exit-policy: 비용은 청산 게이트가 아니다.** 청산·부분익절 발의는
//     예상 비용을 이유로 차단되지 않는다(SHALL NOT). StockOS's counterpart
//     (`SELL_COST_BUFFER_EXCEEDED`, guardian.py:467-468) refuses a *sale* when
//     its estimated cost exceeds a buffer — that is a gate that keeps a position
//     open because closing it looks expensive, which is the one direction a risk
//     control must never fail in. It is on risk-management's 제외 list and no
//     identifier of it exists in this tree.
//  2. **exit-policy: 상한 없는 과대 추정은 본전 기준선을 무한히 끌어올려 승자
//     포지션을 즉시 청산시킨다.** The cap is what bounds that, so the bound is
//     stated as a number here rather than left implicit in MaxRate's comment.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// forbiddenGateIdentifiers are the StockOS sell-cost gate's names. Any of them
// appearing as an identifier or a string constant means someone ported the gate
// that §0.3 excludes.
var forbiddenGateIdentifiers = []string{
	"sell_cost_buffer",
	"sellcostbuffer",
	"sell_cost_buffer_exceeded",
}

// TestNoSellCostGateExistsInTheTree scans the Go sources for the excluded gate.
//
// Comments are deliberately not scanned: the exclusion has to be *explained*
// somewhere (docs/guardian-chain.md, and the chain's own comments), and a
// substring scan that could not tell an explanation from an implementation would
// force the explanation out of the code. Parsing without comments makes the test
// precise — it fails on a use, never on a mention.
func TestNoSellCostGateExistsInTheTree(t *testing.T) {
	roots := []string{filepath.Join("..", "..", "internal"), filepath.Join("..", "..", "cmd")}
	scanned := 0
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			// No parser.ParseComments: comments are not code.
			file, err := parser.ParseFile(token.NewFileSet(), path, src, 0)
			if err != nil {
				return err
			}
			scanned++
			ast.Inspect(file, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.Ident:
					checkForbidden(t, path, node.Name)
				case *ast.BasicLit:
					if node.Kind == token.STRING {
						if unquoted, err := strconv.Unquote(node.Value); err == nil {
							checkForbidden(t, path, unquoted)
						}
					}
				}
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatalf("walking %s: %v", root, err)
		}
	}
	if scanned < 100 {
		t.Fatalf("only %d Go files scanned; the walk did not reach the tree and would pass vacuously", scanned)
	}
}

func checkForbidden(t *testing.T, path, text string) {
	t.Helper()
	lowered := strings.ToLower(text)
	for _, forbidden := range forbiddenGateIdentifiers {
		if strings.Contains(lowered, forbidden) {
			t.Errorf("%s uses %q: §0.3 excludes the sell-cost gate — a liquidation must never be refused "+
				"because its estimated cost looks high", path, text)
		}
	}
}

// TestCostModelAnswersAmountsNeverVerdicts pins the shape that makes the
// exclusion durable.
//
// A cost model that returns a yes/no answer is a cost model something can wire
// into a refusal path. Every exported function here returns money, a rate or an
// error — nothing that reads as a verdict. The check is over this package's own
// AST because "nobody has wired it yet" is not the same guarantee as "there is
// nothing to wire".
func TestCostModelAnswersAmountsNeverVerdicts(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("parsing the package: %v", err)
	}
	if len(pkgs) == 0 {
		t.Fatal("no package parsed; the check would pass vacuously")
	}

	verdictish := []string{"decision", "verdict", "gate", "allowed", "blocked", "refus"}
	exported := 0
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || !fn.Name.IsExported() || fn.Type.Results == nil {
					continue
				}
				exported++
				for _, result := range fn.Type.Results.List {
					name := typeName(result.Type)
					if name == "bool" {
						t.Errorf("%s returns a bool: a cost model must not answer a yes/no question (§0.3)", fn.Name.Name)
					}
					for _, word := range verdictish {
						if strings.Contains(strings.ToLower(name), word) {
							t.Errorf("%s returns %q, which reads as a verdict rather than an amount (§0.3)", fn.Name.Name, name)
						}
					}
				}
			}
		}
	}
	if exported == 0 {
		t.Fatal("no exported functions found; the check would pass vacuously")
	}
}

func typeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return typeName(t.X)
	case *ast.SelectorExpr:
		return t.Sel.Name
	case *ast.ArrayType:
		return typeName(t.Elt)
	default:
		return ""
	}
}

// TestBreakEvenCannotRunAwayAboveTheCap states the bound the cap buys.
//
// With every sell-side component pinned at MaxRate — the worst configuration the
// gate will accept — the break-even price is at most 1/(1 − 3×MaxRate) of the
// entry plus the buy leg. That is a bounded, if hostile, baseline. Without the
// cap the same expression has no bound at all: as the sell-side rate approaches
// 1 the break-even price goes to infinity, and an exit policy whose 실질 본전 is
// unreachable closes every winner the instant it opens.
func TestBreakEvenCannotRunAwayAboveTheCap(t *testing.T) {
	atCap := strconv.FormatFloat(MaxRate, 'f', -1, 64)
	worst := modelWith(t, map[string]string{
		KeyUSBuyCommissionRate:     atCap,
		KeyUSSellCommissionRate:    atCap,
		KeyUSSellRegulatoryFeeRate: atCap,
		KeyUSFXConversionFeeRate:   atCap,
	})

	got, err := worst.BreakEvenSellPrice("100", "1", MarketUS)
	if err != nil {
		t.Fatalf("BreakEvenSellPrice: %v", err)
	}
	// Buy leg at the cap adds commission + FX = 2×MaxRate; the sell side grosses
	// up by 1/(1 − 3×MaxRate).
	bound := 100 * (1 + 2*MaxRate) / (1 - 3*MaxRate)
	if v := amount(t, got); v > bound {
		t.Fatalf("break-even %v exceeds the bound %v the cap is supposed to buy", v, bound)
	}

	// And the placeholders in use are nowhere near it: the KR round trip's
	// break-even sits within a percent of the entry.
	def, err := DefaultModel().BreakEvenSellPrice("10000", "10", MarketKR)
	if err != nil {
		t.Fatalf("BreakEvenSellPrice(defaults): %v", err)
	}
	if v := amount(t, def); v > 10000*1.01 {
		t.Fatalf("default break-even %v is more than 1%% above entry; the placeholders are no longer conservative-but-plausible", v)
	}
}

// TestRateAboveCapIsRejectedForEveryKey is the 1.2 half of the cap: the
// configuration that would produce a runaway baseline never loads.
func TestRateAboveCapIsRejectedForEveryKey(t *testing.T) {
	for _, key := range OverrideKeys() {
		for _, tooBig := range []string{"0.051", "0.2", "5", "20"} {
			if _, err := NewModel(map[string]string{key: tooBig}); err == nil {
				t.Fatalf("%s accepted %s, above the %v cap", key, tooBig, MaxRate)
			}
		}
	}
}
