package engine

// exit_source_age.go 는 a163 — 청산 관측의 체결 나이(받은 시각 − 브로커 데이터 시각)를 보고하는 부분임.
//
// # 무엇을 하고 무엇을 하지 않나
//
// 공식 /prices 의 `timestamp`(브로커 데이터 시각)는 장외 · 저유동 · 거래정지에서 몇 시간 묵을 수 있지만, 판정은 지금처럼 받은 시각
// (FetchedAt) 기준으로 돎(design D0 「A」). 이 파일은 그 사실을 **보이게만** 함: 구간이 바뀐 포지션마다 전용 로거에 한 줄,
// 브로커 시각이 없거나 해독할 수 없는 포지션은 일반 등급 알림 한 번. 어떤 판정 · 발의 · 제출도 이 파일의 결과를 읽지 않음 —
// 금지 census(a163_judgement_pin_test.go)가 SourceTimestamp · SourceAgeReport 를 읽는 자리를 셈.
//
// # 언제 도나
//
// ObserveOnce 의 판정 순회가 끝난 뒤, settleUnobserved 바로 앞에 한 번(뒤 포지션의 손절 판정 앞에 서지 않음). a090 주기 집계의
// judged · held 를 **읽기만** 함 — 그 집계를 비우는 것은 settleUnobserved 라서 이 호출이 그 앞이어야 함. 무보유 · 조기 반환 주기는
// 부르지 않음(판정이 없음).
//
// # 패닉 격리
//
// 감독자(Runtime.Run · ExitObserver.Run)에는 recover 가 없으므로 여기서의 패닉은 곧 엔진 종료 = 손절 없음임. 진입 함수 첫 줄의
// recover 경계가 거두고, 해독 · 래치 · 로그 · 알림은 전부 그 함수의 호출 트리 안에서 함(Go 의 recover 는 패닉한 프레임 위의 defer
// 에서만 받음). 래치 맵은 ObserveOnce 의 goroutine 만 쓰므로 잠금이 없음 — concurrent map write 는 recover 로 거둘 수 없는 fatal 이라
// 이 소유 전제가 격리의 일부임(-race 목록: a163_source_age_panic_test.go).
//
// # 토글
//
// 보고 전체가 기본 꺼진 운영 토글(ExitObserverOptions.SourceAgeReport ← engine.exit_source_age_report) 뒤에 있음. 판단은 recover
// 경계 바로 뒤 첫 판단 하나 — 꺼져 있으면 줄 · 알림 · 래치 0 이고 a163 이전 동작과 같음(design D8). 토글은 판정 입력이 아님.

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// 체결 나이 구간(보고 해상도일 뿐 판정 · 알림 임계가 아님 — 경계 5분 · 1시간은 사람 승인 값, 사용자 결정 2026-10-22).
const (
	sourceAgeMissing     = "missing"     // 원문이 빈 문자열(null · 키 없음 · 채우지 않는 생산자)
	sourceAgeUnparseable = "unparseable" // 비어 있지 않으나 RFC3339Nano 해독 실패(계약 형식 변화 — +0900 · 공백 구분 등)
	sourceAgeAhead       = "ahead"       // 브로커 시각이 받은 시각보다 미래(시계 차)
	sourceAgeLE15s       = "le_15s"      // 0 ≤ 나이 ≤ execgw.QueryPriceEvidenceDuration
	sourceAgeLE5m        = "le_5m"
	sourceAgeLE1h        = "le_1h"
	sourceAgeGT1h        = "gt_1h"

	// sourceAgeAlertsPerCycle 은 한 주기의 source_missing 알림 상한임(= obs.DefaultNormalRelayCapacity 64 의 1/8, 사람 승인 값).
	// 넘는 포지션은 알림함을 세우지 않으므로 다음 주기로 넘어감 — 재시작 직후 몰려도 **한 주기 안에서는** 일반 등급 버퍼의 7/8 이
	// 다른 exit 알림 몫으로 남음. 여러 주기에 걸친 누적(시각 부재 ↔ 복귀가 번갈아 재무장되는 진동)은 이 상한이 막지 않음 — issues.md I5.
	sourceAgeAlertsPerCycle = 8
	// sourceAgeRawLimit 은 해독 불가 원문을 줄 · 알림에 싣는 최대 바이트임(브로커가 보낸 긴 글자를 그대로 퍼뜨리지 않음).
	sourceAgeRawLimit = 64

	sourceAgePanicDetail = "the source-age report panicked and was recovered; this cycle's judgements and orders were already done"
)

// sourceAgeRecord 는 포지션 하나의 보고 래치임 — 마지막 구간과 「source_missing 을 이미 알렸나」.
type sourceAgeRecord struct {
	bucket  string
	alerted bool
}

// reportSourceAge 는 판정 순회 뒤 한 번 도는 보고 진입 함수임. 판정 · 제출은 이미 끝났고, 여기서 무엇이 실패해도(패닉 포함)
// 그 주기와 다음 주기의 판정은 바뀌지 않음.
func (o *ExitObserver) reportSourceAge(ctx context.Context, cycle *ExitCycle, states []managed, quotes map[string]observedQuote) {
	defer func() {
		if recover() != nil {
			o.sourceAgeRecovered()
			return
		}
		o.sourceAgePanicLogged = false // 패닉 없이 끝난 보고 — 다음 패닉 연속의 첫 줄을 다시 쓰게 함
	}()
	if !o.opts.SourceAgeReport {
		return // 꺼짐 = a163 이전 동작(줄 · 알림 · 래치 0)
	}

	tally := o.cycleUnobserved
	if tally == nil || tally.owner != cycle || len(tally.held) == 0 {
		// 이 주기의 보유 · 대상 표시가 없음 — a090 settleUnobserved 의 정리 기준과 같이 기록 전체를 비움.
		o.sourceAge = nil
		return
	}
	if o.sourceAge == nil {
		o.sourceAge = map[string]*sourceAgeRecord{}
	}
	// 정리: 이번 주기 held 밖의 id(청산 · 대상 제외 · 완료 정책)는 지움. held 안이지만 판정 안 된 포지션은 그대로 둠.
	for id := range o.sourceAge {
		if _, ok := tally.held[id]; !ok {
			delete(o.sourceAge, id)
		}
	}

	// 판정된 포지션을 id 순으로 — 주기 상한을 넘는 포지션이 결정적으로 다음 주기로 넘어가게.
	judged := make([]managed, 0, len(tally.judged))
	for _, m := range states {
		if _, ok := tally.judged[m.position.ID]; ok {
			judged = append(judged, m)
		}
	}
	sort.Slice(judged, func(i, j int) bool { return judged[i].position.ID < judged[j].position.ID })

	alerts := 0
	for _, m := range judged {
		quote, ok := quotes[strings.ToUpper(strings.TrimSpace(m.position.Symbol))]
		if !ok {
			continue // 판정 진입은 그 시세로만 일어나므로 도달하지 않는 갈래 — 그래도 보고가 판정을 추측하지 않게 건너뜀
		}
		bucket, age, parsed := classifySourceAge(quote.SourceTimestamp, quote.FetchedAt)
		rec := o.sourceAge[m.position.ID]
		if rec == nil {
			rec = &sourceAgeRecord{}
			o.sourceAge[m.position.ID] = rec
		}
		if rec.bucket != bucket {
			o.logSourceAge(m, bucket, rec.bucket, age, parsed, quote.SourceTimestamp)
			rec.bucket = bucket
		}
		if !sourceAgeAlertable(bucket) {
			rec.alerted = false // 시각이 돌아옴 — 다음 부재를 다시 알리게 재무장
			continue
		}
		if rec.alerted || alerts >= sourceAgeAlertsPerCycle {
			continue
		}
		o.alert(ctx, o.sourceMissingEvent(m, bucket, quote.SourceTimestamp))
		rec.alerted = true
		alerts++
	}
}

// classifySourceAge 는 브로커 원문과 받은 시각으로 구간을 정함. 나이 = 받은 시각 − 브로커 시각(observedQuote.FetchedAt 은
// observe 가 넣은 effectiveAt — 공식 시각이 0 이면 사이클 대체 값이라 0 시각으로 재는 일이 없음).
func classifySourceAge(raw string, receivedAt time.Time) (bucket string, age time.Duration, parsed bool) {
	if raw == "" {
		return sourceAgeMissing, 0, false
	}
	// RFC3339Nano 의 가장 긴 모양은 35 바이트 — 64 바이트를 넘는 글자는 해독하지 않고 해독 불가로 둠. 브로커가 보낸 거대 문자열을
	// 청산 goroutine 에서 매 주기 끝까지 훑지 않게(Go 의 파서는 소수부 자릿수를 제한하지 않음, Codex 리뷰 2026-10-22).
	if len(raw) > sourceAgeRawLimit {
		return sourceAgeUnparseable, 0, false
	}
	at, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return sourceAgeUnparseable, 0, false
	}
	age = receivedAt.Sub(at)
	switch {
	case age < 0:
		return sourceAgeAhead, age, true
	case age <= execgw.QueryPriceEvidenceDuration:
		return sourceAgeLE15s, age, true
	case age <= 5*time.Minute:
		return sourceAgeLE5m, age, true
	case age <= time.Hour:
		return sourceAgeLE1h, age, true
	default:
		return sourceAgeGT1h, age, true
	}
}

// sourceAgeAlertable 은 source_missing 알림 사유인 구간임(H4 — 부재 · 해독 불가). 묵음 · 미래는 로그만(H1).
func sourceAgeAlertable(bucket string) bool {
	return bucket == sourceAgeMissing || bucket == sourceAgeUnparseable
}

// sourceAgeRaw 는 해독 불가 원문의 앞 64 바이트임(그 밖의 구간에는 싣지 않음). 글자 경계에서 자르고 인쇄 불가 글자(개행 · 제어문자)는
// '?' 로 바꿈 — 브로커가 보낸 글자가 알림 본문(renderFields → ntfy)에 줄을 끼워 넣거나 깨진 UTF-8 로 전송을 망치지 않게(리뷰 2026-10-22).
func sourceAgeRaw(raw string) string {
	if len(raw) > sourceAgeRawLimit {
		cut := sourceAgeRawLimit
		for cut > 0 && !utf8.RuneStart(raw[cut]) {
			cut--
		}
		raw = raw[:cut]
	}
	return strings.Map(func(r rune) rune {
		if r == utf8.RuneError || !unicode.IsPrint(r) {
			return '?'
		}
		return r
	}, raw)
}

// logSourceAge 는 구간 전이 한 줄을 씀. 필드는 design D4 목록뿐 — 계좌 ref 없음(안전 8).
func (o *ExitObserver) logSourceAge(m managed, bucket, prev string, age time.Duration, parsed bool, raw string) {
	logger := o.opts.SourceAgeLog
	if logger == nil {
		return
	}
	args := []any{"position_id", m.position.ID, obs.FieldMarket, m.position.Market, obs.FieldSymbol, m.position.Symbol,
		"bucket", bucket, "prev_bucket", prev}
	if parsed {
		args = append(args, "age_seconds", int64(age/time.Second))
	}
	if bucket == sourceAgeUnparseable {
		args = append(args, "source_timestamp", sourceAgeRaw(raw))
	}
	logger.Event(obs.EventExitSourceAge, args...)
}

// sourceMissingEvent 는 포지션 단위 일반 등급 알림임. key 는 포지션 id — 계좌를 싣지 않음.
func (o *ExitObserver) sourceMissingEvent(m managed, bucket, raw string) obs.Event {
	fields := map[string]any{
		"position_id":   m.position.ID,
		obs.FieldMarket: m.position.Market,
		obs.FieldSymbol: m.position.Symbol,
		"bucket":        bucket,
	}
	why := "브로커가 데이터 시각을 보내지 않았다"
	if bucket == sourceAgeUnparseable {
		fields["source_timestamp"] = sourceAgeRaw(raw)
		why = "브로커 데이터 시각을 해독할 수 없다"
	}
	return obs.Event{
		Type:  obs.EventExitSourceMissing,
		Key:   string(obs.EventExitSourceMissing) + "|" + m.position.ID,
		Title: o.label(m.position.Symbol) + " 시세의 데이터 시각을 알 수 없다",
		Body: fmt.Sprintf("포지션 %s(%s) — %s. 손절 판정은 받은 시각 기준으로 그대로 돌았다. 이 시세가 언제의 체결인지는 "+
			"알 수 없으므로 장외 · 거래정지 여부를 확인하라. 같은 포지션은 시각이 돌아올 때까지 다시 알리지 않는다.",
			m.position.ID, m.position.Symbol, why),
		Fields: fields,
	}
}

// sourceAgeRecovered 는 거둔 패닉의 고정 문구 한 줄임(계좌 · 원문 없음). 이 쓰기 자체의 패닉도 거둠 — 쓰기 대상이 패닉의 원인일 수 있음.
// 같은 패닉이 매 주기 되풀이되면 연속의 첫 주기에만 씀(전이 줄 정책과 같음 — 5초마다 같은 줄은 관측이 아님). 패닉 없이 끝난 보고가 풂.
func (o *ExitObserver) sourceAgeRecovered() {
	defer func() { _ = recover() }()
	if o.sourceAgePanicLogged {
		return
	}
	o.sourceAgePanicLogged = true
	if logger := o.opts.SourceAgeLog; logger != nil {
		logger.Warn(obs.EventExitSourceAge, obs.FieldDetail, sourceAgePanicDetail)
	}
}
