//go:build tossos_testseams

package execgw

// a137 승인된 FX 조항: base 통화 손실은 환산 · 권한 없이, 다른 시장의 손실은 같은 파도 짝 FX 의 그 시장 권한으로만(올림), 없으면 측정 불가.

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
)

func a137USLoss(pnl string) journal.DailyLossRow {
	return journal.DailyLossRow{PositionID: "p-us-" + pnl, Market: "us", Symbol: "AAPL", ClosedAt: "2026-10-14T01:00:00Z", Outcome: &pnl}
}

func TestA137AForeignLossConvertsOnlyThroughThePairedAuthority(t *testing.T) {
	g, _, clk := a137Guardian(t)
	now := clk.Now()
	kr, err := officialfx.EvidenceForAuthorityTest("KRW", "KRW", "1", "1", now.Add(-time.Minute), now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	us, err := officialfx.EvidenceForAuthorityTest("USD", "KRW", "1400", "1.01", now.Add(-time.Minute), now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	g.dailyLossRows = a137Rows([]journal.DailyLossRow{a137Loss("-1000"), a137USLoss("-10.5")}, nil, nil)
	m := g.MeasureDailyLoss(context.Background())

	// KR 요청 + 같은 파도 US 권한: 1000 + ceil(10.5 × 1400 × 1.01 = 14847) = 15847.
	paired := qFinalManualBuyFX(QFinalEntryIssuance{FXAuthority: kr, PairedFXAuthority: us}, costs.MarketKR)
	account, err := g.dailyLossFloor(a137Account("0"), m, now, paired)
	if err != nil || account.DailyRealizedLoss.Amount != "15847" || account.DailyRealizedLoss.Currency != "KRW" {
		t.Fatalf("account=%+v err=%v, want 15847 KRW", account.DailyRealizedLoss, err)
	}
	// 짝 권한이 없으면 측정 불가 — 재조회하지 않음.
	alone := qFinalManualBuyFX(QFinalEntryIssuance{FXAuthority: kr}, costs.MarketKR)
	if _, err := g.dailyLossFloor(a137Account("0"), m, now, alone); err == nil {
		t.Fatal("a USD loss without the paired US authority must be unmeasurable")
	}
	// base 통화 손실만 있으면 권한 없이 측정.
	g.dailyLossRows = a137Rows([]journal.DailyLossRow{a137Loss("-1000")}, nil, nil)
	if account, err := g.dailyLossFloor(a137Account("0"), g.MeasureDailyLoss(context.Background()), now, nil); err != nil ||
		account.DailyRealizedLoss.Amount != "1000" {
		t.Fatalf("account=%+v err=%v, want 1000 KRW with no authority", account.DailyRealizedLoss, err)
	}
}

// 체인을 돌지 않는 execgw.Issuer.IssueEntry 는 생산 호출자가 0 이어야 함 — 측정 바닥이 걸리지 않는 문이 생산에 열리지 않게.
// Issuer 자체는 flatten 이 위험 축소(IssueReduction)에 씀(정당) — 그래서 진입 요청(EntryRequest)을 만드는 생산 자리가 0 임을 고정함.
func TestA137TheChainlessIssuerHasNoProductionCaller(t *testing.T) {
	root := filepath.Join("..", "..")
	var sites []string
	for _, dir := range []string{"cmd", "internal"} {
		err := filepath.Walk(filepath.Join(root, dir), func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return err
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				return err
			}
			ast.Inspect(file, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.CompositeLit:
					if sel, ok := node.Type.(*ast.SelectorExpr); ok && sel.Sel.Name == "EntryRequest" {
						sites = append(sites, path)
					}
					if id, ok := node.Type.(*ast.Ident); ok && id.Name == "EntryRequest" {
						sites = append(sites, path)
					}
				case *ast.SelectorExpr:
					// 메서드 값(i.IssueEntry 를 넘김)도 막음 — execgw 밖에서 Issuer.IssueEntry 를 이름으로 꺼내는 자리.
					if node.Sel.Name == "IssueEntry" && file.Name.Name != "execgw" && strings.Contains(path, "flatten") {
						sites = append(sites, path)
					}
				}
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(sites) != 0 {
		t.Fatalf("production builds an execgw.EntryRequest at %v — Issuer.IssueEntry does not apply the daily-loss floor", sites)
	}
}
