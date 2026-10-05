package verifylive

// reconcile_g1_test.go — a121 tasks 2.2·2.2.1: design G1(부재 증명) 의 거절 사례와 수락 양성 대조.
//
// 각 시험은 newRCHarness 의 수락 픽스처에서 하나만 바꾼다. 거절은 코드(ReconcileRefusalCode)로 단언하고, 기록이
// 한 바이트도 바뀌지 않았음을 같이 잰다.

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// --- 수락(양성 대조) ---------------------------------------------------------------

// TestReconcileAppendsOneReconciledAbsentLineWhenEveryConditionHolds 는 spec 「every condition of authoritative absence
// holds」 다. 이 시험이 GREEN 이 아니면 아래 거절 시험들의 "거절" 은 아무것도 증명하지 못한다.
func TestReconcileAppendsOneReconciledAbsentLineWhenEveryConditionHolds(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	res, line, _ := rcAccept(t, h)
	if res.Artifact.ID != rcTargetID || res.Artifact.Kind != KindConditional {
		t.Fatalf("result names %s %s, want the a063 target", res.Artifact.Kind, res.Artifact.ID)
	}
	if line["kind"] != KindReconcile {
		t.Fatalf("appended line kind %v, want %q", line["kind"], KindReconcile)
	}
	// 재개 정리 계획에서 빠지고 투영에서도 빠진다 — 새 취소가 계획되지 않는다.
	entries, err := LoadEntries(h.path)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range PendingCleanup(entries) {
		if a.ID == rcTargetID {
			t.Fatal("a reconciled artifact is still offered for cleanup (a new DELETE would be planned)")
		}
	}
	for _, a := range Outstanding(entries) {
		if a.ID == rcTargetID {
			t.Fatal("a reconciled artifact is still outstanding")
		}
	}
}

// TestReconcileAcceptsExactlyThePageCap 는 페이지 상한(maxFixturePages = 10, limit 100 — design G1 「치르는 값」)의
// 경계 양성 대조다: 10 번째 페이지에서 끝나는 그룹은 거절이 아니다.
func TestReconcileAcceptsExactlyThePageCap(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	var pages []rcPage
	for i := 0; i < maxFixturePages; i++ {
		last := i == maxFixturePages-1
		next := ""
		if !last {
			next = fmt.Sprintf("closed-c%d", i+1)
		}
		pages = append(pages, rcCondPage(next, !last))
	}
	h.reader.set(-1, ReconcileGroupConditionalClosed, pages...)
	rcAccept(t, h)
}

// TestReconcileApprovalWaitDoesNotConsumeTheFreshnessWindow 는 승인 순서(freeze 재검 P2)다: 사람 승인은 목록 읽기
// 전에 받으므로, 승인 대기가 Q3 창을 쓰면 안 된다.
func TestReconcileApprovalWaitDoesNotConsumeTheFreshnessWindow(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	h.onApprove = func() { h.clock.advance(100 * rcFreshness) }
	rcAccept(t, h)
	if len(h.approvals) != 1 {
		t.Fatalf("want exactly one approval, got %d", len(h.approvals))
	}
}

// TestReconcileQueriesTheRecordLineSymbolBytes 는 양성 대조 (i)다: 조회 심볼은 호출자 입력이 아니라 artifact 줄의
// 심볼 바이트 그대로다. 공백·대소문자 정규화가 끼면 다른 종목을 묻게 된다.
func TestReconcileQueriesTheRecordLineSymbolBytes(t *testing.T) {
	const odd = "brk.b " // 정규화(TrimSpace·ToUpper)가 바꾸는 바이트
	entries := rcA063Entries()
	entries[1].Artifacts[0].Symbol = odd
	h := newRCHarness(t, entries)
	rcAccept(t, h)
	calls := h.reader.allCalls()
	if len(calls) == 0 {
		t.Fatal("no reads at all")
	}
	for _, c := range calls {
		if c.Symbol != odd {
			t.Fatalf("%s %s queried symbol %q, want the record line's bytes %q", c.Kind, c.Group, c.Symbol, odd)
		}
	}
}

// TestReconcileReadsTheGroupsInTheFixedOrder 는 G1-6 의 고정 읽기 순서(조건주문 OPEN → 일반 OPEN → CLOSED)를 두 번,
// 그 앞뒤의 종목 조회, 그리고 limit 100 을 호출 열로 고정한다.
func TestReconcileReadsTheGroupsInTheFixedOrder(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	rcAccept(t, h)
	var got []string
	for _, c := range h.reader.allCalls() {
		if c.Kind == "instrument" {
			got = append(got, "instrument")
			continue
		}
		got = append(got, c.Group)
		if c.Limit != 100 {
			t.Fatalf("%s read with limit %d, want 100", c.Group, c.Limit)
		}
	}
	want := []string{"instrument",
		ReconcileGroupConditionalOpen, ReconcileGroupPlainOpen, ReconcileGroupConditionalClosed,
		ReconcileGroupConditionalOpen, ReconcileGroupPlainOpen, ReconcileGroupConditionalClosed,
		"instrument"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("read order\n got %v\nwant %v", got, want)
	}
}

// --- G1-1 · G1-2 · G1-3 (부재 판정) -----------------------------------------------

// TestReconcileRefusesALiveSuccessorInTheOpenConditionalGroup — 옛 id 는 없지만 정정이 발급한 새 id 가 OPEN 에 있다.
func TestReconcileRefusesALiveSuccessorInTheOpenConditionalGroup(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	h.reader.set(-1, ReconcileGroupConditionalOpen, rcCondPage("", false, rcCondRow("CO-SUCCESSOR", "WATCHING")))
	rcRefuse(t, h, RefuseOpenConditional)
}

// TestReconcileRefusesTheTargetFoundInClosed — 대상 id 가 CLOSED 에 COMPLETED + triggeredOrderId 로 있다(부재가 아니다).
func TestReconcileRefusesTheTargetFoundInClosed(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	row := rcCondRow(rcTargetID, "COMPLETED")
	row.TriggeredOrderID = "ORD-CHILD-1"
	h.reader.set(-1, ReconcileGroupConditionalClosed, rcCondPage("", false, row))
	rcRefuse(t, h, RefuseClosedTargetPresent)
}

// TestReconcileRefusesTheTargetExpiredInClosedWithThePermanentMessage 는 Q6 이다 — 거절 메시지가 영구 거절을 적는다.
func TestReconcileRefusesTheTargetExpiredInClosedWithThePermanentMessage(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	h.reader.set(-1, ReconcileGroupConditionalClosed, rcCondPage("", false, rcCondRow(rcTargetID, "EXPIRED")))
	r := rcRefuse(t, h, RefuseClosedTargetExpired)
	const want = "만료된 artifact 는 이 경로로 영구히 대사되지 않는다"
	if !strings.Contains(r.Error(), want) {
		t.Fatalf("Q6 refusal must say %q; got %q", want, r.Error())
	}
}

// TestReconcileRefusesAnotherIdentifierThatFiredInClosed — 정정으로 id 가 바뀐 후속이 발동하면 **다른 id** 로 CLOSED 에
// 남는다(리뷰 P0). triggeredOrderId 비공란 또는 COMPLETED 면 어느 id 든 거절.
func TestReconcileRefusesAnotherIdentifierThatFiredInClosed(t *testing.T) {
	triggered := rcCondRow("CO-OTHER-1", "EXPIRED")
	triggered.TriggeredOrderID = "ORD-CHILD-2"
	for name, row := range map[string]official.ReconcileConditionalRow{
		"triggered-order-id": triggered,
		"completed":          rcCondRow("CO-OTHER-2", "COMPLETED"),
	} {
		t.Run(name, func(t *testing.T) {
			h := newRCHarness(t, rcA063Entries())
			h.reader.set(-1, ReconcileGroupConditionalClosed, rcCondPage("", false, row))
			rcRefuse(t, h, RefuseClosedFiredTrace)
		})
	}
}

// TestReconcileRefusesAnyOpenPlainOrderOnTheSymbol — 발동한 child 가 체결 전이면 일반 주문으로 호가에 있다(G1-3).
func TestReconcileRefusesAnyOpenPlainOrderOnTheSymbol(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	h.reader.set(-1, ReconcileGroupPlainOpen, rcPlainPage("", false,
		official.ReconcileOrderRow{ID: "ORD-CHILD-3", Symbol: rcSymbol, Status: "PENDING"}))
	rcRefuse(t, h, RefuseOpenPlainOrder)
}

// --- G1-5 · F4 · F6 (행 검증) ------------------------------------------------------

func TestReconcileRefusesARowForAnotherSymbol(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	row := rcCondRow("CO-OTHER-3", "EXPIRED")
	row.Symbol = "000660"
	h.reader.set(-1, ReconcileGroupConditionalClosed, rcCondPage("", false, row))
	rcRefuse(t, h, RefuseRowSymbol)
}

func TestReconcileRefusesARowForAnotherMarket(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	row := rcCondRow("CO-OTHER-4", "EXPIRED")
	row.Market = MarketUS
	h.reader.set(-1, ReconcileGroupConditionalClosed, rcCondPage("", false, row))
	rcRefuse(t, h, RefuseRowMarket)
}

// TestReconcileRefusesARowMissingRequiredFields — id·symbol·market 결측(F4 필수 필드).
func TestReconcileRefusesARowMissingRequiredFields(t *testing.T) {
	for name, mutate := range map[string]func(*official.ReconcileConditionalRow){
		"id":     func(r *official.ReconcileConditionalRow) { r.ID = "" },
		"symbol": func(r *official.ReconcileConditionalRow) { r.Symbol = "" },
		"market": func(r *official.ReconcileConditionalRow) { r.Market = "" },
	} {
		t.Run(name, func(t *testing.T) {
			h := newRCHarness(t, rcA063Entries())
			row := rcCondRow("CO-OTHER-5", "EXPIRED")
			mutate(&row)
			h.reader.set(-1, ReconcileGroupConditionalClosed, rcCondPage("", false, row))
			rcRefuse(t, h, RefuseRowIncomplete)
		})
	}
}

// TestReconcileRefusesOCORowsInEitherGroup 는 codex F6 이다 — second 다리가 있는 행이 하나라도 보이면 거절.
// OPEN 에서도 G1-1 보다 먼저 OCO 로 거절돼야 이 가드가 실제로 서 있음이 보인다.
func TestReconcileRefusesOCORowsInEitherGroup(t *testing.T) {
	for _, group := range []string{ReconcileGroupConditionalOpen, ReconcileGroupConditionalClosed} {
		t.Run(group, func(t *testing.T) {
			h := newRCHarness(t, rcA063Entries())
			status := "EXPIRED"
			if group == ReconcileGroupConditionalOpen {
				status = "WATCHING"
			}
			row := rcCondRow("CO-OCO-1", status)
			row.HasSecond = true
			h.reader.set(-1, group, rcCondPage("", false, row))
			rcRefuse(t, h, RefuseOCO)
		})
	}
}

// TestReconcileRefusesClosedStatusOutsideTheAllowlist 는 codex F4 다 — 결측·미지·그룹 모순(CLOSED 에 OPEN 계열)은
// 그 자체로 거절. 세 값은 어떤 전사 결과의 허용 목록에도 들 수 없다.
func TestReconcileRefusesClosedStatusOutsideTheAllowlist(t *testing.T) {
	for name, status := range map[string]string{"missing": "", "unknown": "FOO", "open-family": "WATCHING"} {
		t.Run(name, func(t *testing.T) {
			h := newRCHarness(t, rcA063Entries())
			h.reader.set(-1, ReconcileGroupConditionalClosed, rcCondPage("", false, rcCondRow("CO-OTHER-6", status)))
			rcRefuse(t, h, RefuseClosedStatus)
		})
	}
}

// TestReconcileClosedStatusAllowlistIsATranscribedGolden 는 F4 의 "allowlist 가 골든/영수증 전사" 를 고정할 자리다.
//
// RED 로트 전사 결과: 출처 부재. verify-execution-capability 의 measurements.md·issues.md·tasks.md 어디에도 조건주문
// CLOSED status 의 실측이 없다(M12·M17·M18·M19·M36·M39 는 WATCHING 과 OPEN/CLOSED 필터만 적는다). 조건주문 종결 어휘를
// 적은 곳은 문서 enum(docs/migration/openapi.latest.json:940-949 — 문서 enum, 측정 아님)과 그 문서를 옮긴 코드 산문
// (internal/official/conditional_reads.go:158-164)뿐이다. 지어내지 않는다 — 출처가 생기면 이 skip 을 그 전사 대조로 바꾼다.
func TestReconcileClosedStatusAllowlistIsATranscribedGolden(t *testing.T) {
	t.Skip("a121 RED: 출처 부재 — verify-execution-capability 영수증/골든에 조건주문 CLOSED status 측정 없음(analysis/red-lot/q3-freshness-and-closed-allowlist.md §2). 전사할 원본이 생기면 이 skip 을 대조 시험으로 교체")
}

// TestReconcileClosedStatusAllowlistStaysEmptyWithoutASource 는 위 skip 의 짝이다: 원본이 없는 동안 허용 목록은 비어
// 있어야 한다(설계 산문·문서 enum 으로 채우면 이 시험이 막는다). 결과로 CLOSED 행이 하나라도 있는 심볼은 거절된다.
func TestReconcileClosedStatusAllowlistStaysEmptyWithoutASource(t *testing.T) {
	if len(reconcileClosedStatusAllowlist) != 0 {
		t.Fatalf("CLOSED status allowlist has %d entries but no measured/golden source was transcribed: %v",
			len(reconcileClosedStatusAllowlist), reconcileClosedStatusAllowlist)
	}
}

// --- G1-4 (Q1 보존 측정) ------------------------------------------------------------

// TestReconcileRefusesWhileTheRetentionMeasurementIsAbsent 는 spec 「the Q1 retention measurement is absent」 다.
// 생산 빌드의 상태(nil)가 바로 이것이다.
func TestReconcileRefusesWhileTheRetentionMeasurementIsAbsent(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	rcSetPolicies(t, nil, rcFreshness)
	rcRefuse(t, h, RefuseRetentionUnmeasured)
}

// TestReconcileRefusesAnyEvictionModelButDuration 는 codex F5·R2-4 의 세 사례다 — 축출 모형 결측·행 수 기반·판별 불능.
func TestReconcileRefusesAnyEvictionModelButDuration(t *testing.T) {
	for name, model := range map[string]evictionModel{
		"missing": evictionUnknown, "count-based": evictionCount, "indeterminate": evictionIndeterminate,
	} {
		t.Run(name, func(t *testing.T) {
			h := newRCHarness(t, rcA063Entries())
			rcSetPolicies(t, &retentionMeasurement{Bound: rcRetentionBound, Eviction: model}, rcFreshness)
			rcRefuse(t, h, RefuseRetentionEviction)
		})
	}
}

// TestReconcileRefusesAnArtifactOlderThanTheRetentionBound — 나이의 기준점은 outstanding 줄의 CreatedAt.
func TestReconcileRefusesAnArtifactOlderThanTheRetentionBound(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	rcSetPolicies(t, &retentionMeasurement{Bound: time.Hour, Eviction: evictionDuration}, rcFreshness) // 나이 2h
	rcRefuse(t, h, RefuseArtifactTooOld)
}

// TestReconcileRefusesAZeroCreatedAtOnTheOutstandingLine — 영 시각에서는 나이를 잴 수 없다.
func TestReconcileRefusesAZeroCreatedAtOnTheOutstandingLine(t *testing.T) {
	entries := rcA063Entries()
	entries[1].Artifacts[0].CreatedAt = time.Time{}
	h := newRCHarness(t, entries)
	rcRefuse(t, h, RefuseCreatedAtZero)
}

// TestReconcileRechecksTheRetentionAgeImmediatelyBeforeAppend 는 codex F7 이다 — 마지막 종목 조회가 느려 그 사이
// artifact 가 보존 한도를 넘으면, 추가 직전 재검사가 거절한다(신선도 창은 넉넉히 둬서 그 가드와 섞이지 않게 한다).
func TestReconcileRechecksTheRetentionAgeImmediatelyBeforeAppend(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	age := rcNow.Sub(rcCreated)
	rcSetPolicies(t, &retentionMeasurement{Bound: age + time.Minute, Eviction: evictionDuration}, time.Hour)
	h.reader.instrument = func(n int, symbol string) (string, error) {
		if n == 1 {
			h.clock.advance(5 * time.Minute)
		}
		return symbol, nil
	}
	rcRefuse(t, h, RefuseArtifactTooOld)
}

// --- G1-6 (두 번 읽기 · 신선도 · 페이지) ---------------------------------------------

// TestReconcileRefusesWhileTheFreshnessBoundIsUnfixed 는 Q3 상수 부재다(골격의 생산 상태).
func TestReconcileRefusesWhileTheFreshnessBoundIsUnfixed(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	rcSetPolicies(t, &retentionMeasurement{Bound: rcRetentionBound, Eviction: evictionDuration}, 0)
	rcRefuse(t, h, RefuseFreshnessUnfixed)
}

// TestReconcileRefusesWhenThePairExceedsTheFreshnessBound — 둘째 읽기가 한도를 넘겨 끝났다.
func TestReconcileRefusesWhenThePairExceedsTheFreshnessBound(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	h.reader.hook = func(c rcCall) {
		if c.Kind == "list" && c.Pass == 1 && c.Group == ReconcileGroupConditionalClosed {
			h.clock.advance(rcFreshness + time.Second)
		}
	}
	rcRefuse(t, h, RefuseFreshnessExceeded)
}

// TestReconcileFreshnessWindowCoversAppendAdmission 는 codex F7 이다 — Q3 창의 끝은 둘째 읽기 완료가 아니라 추가 승인
// 직전이다. 읽기 뒤 마지막 종목 조회가 느리면 거절.
func TestReconcileFreshnessWindowCoversAppendAdmission(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	h.reader.instrument = func(n int, symbol string) (string, error) {
		if n == 1 {
			h.clock.advance(rcFreshness + time.Second)
		}
		return symbol, nil
	}
	rcRefuse(t, h, RefuseFreshnessExceeded)
}

// TestReconcileRefusesAStatusTransitionBetweenTheTwoReads — (그룹, id, status, triggeredOrderId) multiset 이 다르다.
func TestReconcileRefusesAStatusTransitionBetweenTheTwoReads(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	first := rcCondRow("CO-OTHER-7", "EXPIRED")
	second := rcCondRow("CO-OTHER-7", "COMPLETED")
	second.TriggeredOrderID = "ORD-CHILD-7"
	h.reader.set(0, ReconcileGroupConditionalClosed, rcCondPage("", false, first))
	h.reader.set(1, ReconcileGroupConditionalClosed, rcCondPage("", false, second))
	rcRefuse(t, h, RefuseReadsDiffer)
}

// TestReconcileRefusesARowAppearingInOnlyOneRead — 둘째 읽기에만 행이 있다.
func TestReconcileRefusesARowAppearingInOnlyOneRead(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	h.reader.set(1, ReconcileGroupConditionalClosed, rcCondPage("", false, rcCondRow("CO-OTHER-8", "EXPIRED")))
	rcRefuse(t, h, RefuseReadsDiffer)
}

// TestReconcileRefusesADuplicateRowWithinOneRead 는 freeze P1-4 다 — 페이지 경계 이동이 만든 (그룹, id) 중복은
// 집합 비교가 접어 숨기므로 한 읽기 안에서 그 자체로 거절한다(S7 포괄이 아닌 자기 코드).
func TestReconcileRefusesADuplicateRowWithinOneRead(t *testing.T) {
	for _, pass := range []int{0, 1} {
		t.Run(fmt.Sprintf("read-%d", pass+1), func(t *testing.T) {
			h := newRCHarness(t, rcA063Entries())
			row := rcCondRow("CO-OTHER-9", "EXPIRED")
			h.reader.set(pass, ReconcileGroupConditionalClosed,
				rcCondPage("closed-c1", true, row), rcCondPage("", false, row))
			rcRefuse(t, h, RefuseDuplicateRow)
		})
	}
}

// TestReconcileRefusesPaginationFaultsInEitherRead 는 m0RecoverPending B4·B6·B13·B14 와 같은 모양을 세 그룹 × 두 읽기에서
// 잰다: 반복 커서 · hasNext 인데 빈 커서 · 상한 도달 · 읽기 오류.
func TestReconcileRefusesPaginationFaultsInEitherRead(t *testing.T) {
	groups := []string{ReconcileGroupConditionalOpen, ReconcileGroupPlainOpen, ReconcileGroupConditionalClosed}
	page := func(group, next string, hasNext bool) rcPage {
		if group == ReconcileGroupPlainOpen {
			return rcPlainPage(next, hasNext)
		}
		return rcCondPage(next, hasNext)
	}
	faults := map[string]struct {
		code  ReconcileRefusalCode
		pages func(group string) []rcPage
	}{
		"repeated-cursor": {RefuseRepeatedCursor, func(g string) []rcPage {
			return []rcPage{page(g, "same", true), page(g, "same", true), page(g, "", false)}
		}},
		"empty-cursor-with-has-next": {RefuseEmptyCursor, func(g string) []rcPage {
			return []rcPage{page(g, "", true)}
		}},
		"page-cap": {RefusePageCap, func(g string) []rcPage {
			var out []rcPage
			for i := 0; i < maxFixturePages+1; i++ {
				out = append(out, page(g, fmt.Sprintf("%s-c%d", g, i+1), true))
			}
			return out
		}},
		"read-error": {RefuseReadError, func(g string) []rcPage {
			return []rcPage{{err: errors.New("official: HTTP 503 server error")}}
		}},
	}
	for name, f := range faults {
		for _, g := range groups {
			for _, pass := range []int{0, 1} {
				t.Run(fmt.Sprintf("%s/%s/read-%d", name, g, pass+1), func(t *testing.T) {
					h := newRCHarness(t, rcA063Entries())
					h.reader.set(pass, g, f.pages(g)...)
					rcRefuse(t, h, f.code)
				})
			}
		}
	}
}

// --- 양성 대조 (P1-3) ----------------------------------------------------------------

// TestReconcileRefusesWhenTheInstrumentControlFails — 읽기 전후 종목 조회가 실패하거나 다른 심볼을 되돌린다.
func TestReconcileRefusesWhenTheInstrumentControlFails(t *testing.T) {
	for name, fn := range map[string]func(n int, symbol string) (string, error){
		"before-fails": func(n int, s string) (string, error) {
			if n == 0 {
				return "", errors.New("official: HTTP 500")
			}
			return s, nil
		},
		"after-fails": func(n int, s string) (string, error) {
			if n == 1 {
				return "", errors.New("official: HTTP 500")
			}
			return s, nil
		},
		"echoes-another-symbol": func(n int, s string) (string, error) { return "000660", nil },
	} {
		t.Run(name, func(t *testing.T) {
			h := newRCHarness(t, rcA063Entries())
			h.reader.instrument = fn
			rcRefuse(t, h, RefuseInstrumentControl)
		})
	}
}
