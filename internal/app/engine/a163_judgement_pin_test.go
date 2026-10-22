package engine

// a163 tasks 1.2 — 「판정 경로 불변」 의 강제 수단 둘(design D7 1·2).
//
// (a) 렌더 해시: 수신자 ExitObserver 의 판정 메서드 아홉과 같은 파일의 함수 stableObservationID(수신자 없음 — design D7 은
//     「수신자 ExitObserver 의 … 열 개」 로 적었으나 HEAD 에서 이것만 자유 함수임, issues.md) 열 개를 주석 없이 go/printer 로 렌더한 sha256 을 굳힘. a163 은 이 열 개를
//     편집하지 않으므로, 본문이 바뀌면 이 change 의 계약(판정 자격 변경 0)이 깨진 것임. 기대값의 원천은
//     openspec/changes/a163-quote-freshness-uses-trade-time/analysis/render-hash/render-hash-base.txt(HEAD 60033240a — base 와 internal diff 0).
//     동명 함수(judge — alertdelivery.go · record — strategy_lane_runtime.go)가 다른 수신자로 있어 이름이 아니라 수신자로 고름.
//     파일 해시(source_sha256)를 쓰지 않는 이유: 같은 파일의 observe · ObserveOnce 를 a163 이 편집하므로 반드시 바뀜(freeze 리뷰 P1-2).
//
// (b) 금지 census: SourceTimestamp(브로커 데이터 시각)와 SourceAgeReport(보고 토글) 식별자의 **모든 출현**을 자리별로 셈.
//     존재 검사가 아니라 금지 census — 허용 자리 밖 출현이 하나라도 있으면 판정 함수나 그 헬퍼가 브로커 시각·토글을 읽는
//     길이 생긴 것임. 렌더 해시가 못 보는 「헬퍼를 거친 누출」을 이 범위가 닫음.

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// a163JudgementRenderHashes 는 render-hash-base.txt 의 ExitObserver 행 열 개임(손으로 옮김 — 변경은 a163 계약 위반이므로
// 갱신하려면 새 change 가 이유를 적어야 함).
var a163JudgementRenderHashes = map[string]string{
	"judge":                 "a3d5c5625182180a44789e394b3c9a93f0af05e70f37a8f41b004bd1052ab570",
	"judgeLadder":           "12a3a2b27be71f00697bfc0ea2474419e083dbabbe7dcf8fa58e8ccb884360b2",
	"judgeRatchet":          "de9209b6f33481e5634bf612b0224dcda3d4e629680d9b35b161b21b7bd37c47",
	"noteHeldProposal":      "2232f89dc1f62d1e63b83eacae0c9c65d44ecdd626a18d57cdd3f9aca38de015",
	"quoteUsable":           "54da8fc5f48d567d5f0dba98b99b4e87ed0318d13340c5f0d172f72b846ff49b",
	"record":                "f3458b84adf93b386bf17d6ae42e49d3a7d49e407829a7c7c10ea789efb499ea",
	"refreshObservation":    "9cf96697467eb3c3d1b15bb6bc496a0bf5d9276b19583884dc2168beeeb2eeaa",
	"requiresFullJudgement": "a0778e2b9397c7b93ba37dca99ff0ee11a8f92ff1ddca54781e6303f16ff49c6",
	"snapshotContext":       "72e0ece21373fb4556fff202e5bea7e12e73a720440545f90780901c466aab86",
	"stableObservationID":   "25fcfa904e0fcc99fd3b2e35e79707ca238338d6157e80dcf071d417665a1480",
}

// a163PinnedReceiver 는 굳힌 함수의 수신자임 — stableObservationID 만 자유 함수(수신자 없음).
func a163PinnedReceiver(name string) string {
	if name == "stableObservationID" {
		return ""
	}
	return "ExitObserver"
}

// a163ReceiverName 은 메서드 수신자의 타입 이름임(포인터 벗김). 함수면 "".
func a163ReceiverName(fd *ast.FuncDecl) string {
	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return ""
	}
	expr := fd.Recv.List[0].Type
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	if id, ok := expr.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}

func TestA163TheJudgementFunctionsRenderExactlyAsBeforeTheChange(t *testing.T) {
	got := map[string][]string{}
	for _, path := range engineProductionFiles(t) {
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0) // 주석 미적재 — 주석 편집에는 둔감, 코드 편집에는 민감
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if _, pinned := a163JudgementRenderHashes[fd.Name.Name]; !pinned {
				continue
			}
			// 동명 함수(judge · record)는 다른 수신자임 — 수신자로 고름. stableObservationID 만 자유 함수.
			if wantRecv := a163PinnedReceiver(fd.Name.Name); a163ReceiverName(fd) != wantRecv {
				continue
			}
			var buf bytes.Buffer
			if err := printer.Fprint(&buf, token.NewFileSet(), fd); err != nil {
				t.Fatalf("render %s: %v", fd.Name.Name, err)
			}
			got[fd.Name.Name] = append(got[fd.Name.Name], fmt.Sprintf("%x", sha256.Sum256(buf.Bytes())))
		}
	}
	for name, want := range a163JudgementRenderHashes {
		hashes := got[name]
		if len(hashes) != 1 {
			t.Errorf("ExitObserver.%s found %d time(s), want exactly one — a moved or duplicated judgement escapes the pin", name, len(hashes))
			continue
		}
		if hashes[0] != want {
			t.Errorf("ExitObserver.%s renders %s, want %s — a163 must not change a judgement function (design D7)", name, hashes[0], want)
		}
	}
}

// a163Site 는 식별자 출현의 자리임 — 파일 · 감싼 함수(수신자 포함) 또는 감싼 타입 선언.
type a163Site struct{ file, owner string }

func (s a163Site) String() string { return s.file + ":" + s.owner }

// a163Census 는 files 의 비시험 Go 소스에서 이름이 name 인 식별자 출현을 자리별로 셈. 선택자 · 복합 리터럴 키 · 필드 선언 ·
// method value 모두 *ast.Ident 이므로 한 번의 순회로 전부 잡힘.
func a163Census(t *testing.T, files []string, name string) map[a163Site]int {
	t.Helper()
	counts := map[a163Site]int{}
	for _, path := range files {
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, decl := range file.Decls {
			owner := "<file scope>"
			switch d := decl.(type) {
			case *ast.FuncDecl:
				owner = d.Name.Name
				if recv := a163ReceiverName(d); recv != "" {
					owner = recv + "." + owner
				}
			case *ast.GenDecl:
				var names []string
				for _, spec := range d.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						names = append(names, "type "+ts.Name.Name)
					}
				}
				if len(names) > 0 {
					owner = strings.Join(names, ",")
				}
			}
			ast.Inspect(decl, func(n ast.Node) bool {
				if id, ok := n.(*ast.Ident); ok && id.Name == name {
					counts[a163Site{file: filepath.Base(path), owner: owner}]++
				}
				return true
			})
		}
	}
	return counts
}

func a163ProductionGoFiles(t *testing.T, dir string) []string {
	t.Helper()
	all, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatalf("glob %s: %v", dir, err)
	}
	var out []string
	for _, path := range all {
		if !strings.HasSuffix(path, "_test.go") {
			out = append(out, path)
		}
	}
	if len(out) == 0 {
		t.Fatalf("no production source under %s — the census would be vacuous", dir)
	}
	return out
}

// a163CheckCensus 는 허용 자리 밖 출현 0 과 허용 자리의 정확한 수를 단언함. exact 가 -1 인 자리는 임의 수(보고 파일 전용).
func a163CheckCensus(t *testing.T, name string, counts map[a163Site]int, allowed map[a163Site]int, reportFile string) {
	t.Helper()
	var outside []string
	for site, n := range counts {
		if site.file == reportFile {
			continue
		}
		if _, ok := allowed[site]; !ok {
			outside = append(outside, fmt.Sprintf("%s ×%d", site, n))
		}
	}
	sort.Strings(outside)
	if len(outside) > 0 {
		t.Errorf("%s appears outside its allowed sites: %v — a judgement path may now read it (design D7·D8)", name, outside)
	}
	for site, want := range allowed {
		if got := counts[site]; got != want {
			t.Errorf("%s at %s = %d, want exactly %d", name, site, got, want)
		}
	}
}

func TestA163TheBrokerDataTimeIsReadOnlyWhereTheDesignAllows(t *testing.T) {
	files := append(a163ProductionGoFiles(t, "."), a163ProductionGoFiles(t, filepath.Join("..", "..", "official"))...)
	counts := a163Census(t, files, "SourceTimestamp")
	a163CheckCensus(t, "SourceTimestamp", counts, map[a163Site]int{
		{file: "exitloop.go", owner: "type observedQuote"}:   1, // 필드 선언
		{file: "exitloop.go", owner: "ExitObserver.observe"}: 2, // observedQuote{…} 리터럴의 키 + q.SourceTimestamp
		{file: "market_reads.go", owner: "adaptPrices"}:      1, // 리터럴 키
		// 편집 전부터 있던 다른 타입의 같은 이름 필드(a112 L1c 엄격 리더 — 생산 판정 소비자 0, officialbars 만 읽음).
		// design D7 2 는 「편집 전 출현 0」 이라 적었으나 HEAD 에 넷이 있음(issues.md). 정확한 수로 굳혀 새 출현은 여전히 잡힘.
		{file: "strict_quote_reads.go", owner: "type StrictLastPrice"}:      1,
		{file: "strict_quote_reads.go", owner: "type StrictTopOfBook"}:      1,
		{file: "strict_quote_reads.go", owner: "Client.StrictLastPrice"}:    1,
		{file: "strict_quote_reads.go", owner: "Client.StrictOrderbookTop"}: 1,
	}, "exit_source_age.go")
}

func TestA163TheReportToggleIsReadOnlyWhereTheDesignAllows(t *testing.T) {
	counts := a163Census(t, a163ProductionGoFiles(t, "."), "SourceAgeReport")
	a163CheckCensus(t, "SourceAgeReport", counts, map[a163Site]int{
		{file: "exitloop.go", owner: "type ExitObserverOptions"}: 1, // 필드 선언
		{file: "exitwiring.go", owner: "Context.ExitObserver"}:   1, // 생산 배선의 무조건 덮어쓰기
	}, "exit_source_age.go")
}

// 금지 census 가 실제로 세는지 — 이 시험 파일 자신의 바이트에서 알려진 출현을 찾게 함(무변이 대조군: 계측기가 눈멀면 위 두 시험이
// 언제나 초록임).
func TestA163TheCensusCountsWhatItIsPointedAt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "probe.go")
	src := "package probe\ntype observedQuote struct{ SourceTimestamp string }\n" +
		"func judge(q observedQuote) string { return q.SourceTimestamp }\n"
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	counts := a163Census(t, []string{path}, "SourceTimestamp")
	if counts[a163Site{file: "probe.go", owner: "type observedQuote"}] != 1 || counts[a163Site{file: "probe.go", owner: "judge"}] != 1 {
		t.Fatalf("census = %v, want one field declaration and one read inside judge", counts)
	}
}
