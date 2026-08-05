# a085 · 알림이 어느 종목인지 한국어로 말한다

- **Feature**: `FEAT-TOS-004` — Operator console controls and visibility
- **Story**: `STORY-TOS-a085`
- **Spec**: `engine-safety`, `operator-console`
- **위험 등급**: **Normal** (알림 문구와 화면 표시. 주문·손절·사이징·원장 판정 경로를 건드리지 않는다.)

## Why

**보호가 멈춘 것을 알리는 알림이 영어이고, 어느 종목인지 숫자로만 말한다. 그리고 그
숫자마저, 화면에서는 아무것도 보여주지 않는다.**

### 알림 (운영 원장 `alert_outbox`, 실제 발송분)

```
title: the exit policy could not judge 032820
body:  the stored protection state or the observed price is not usable, so this
       position is not being judged at all: journal: exit snapshot generation is
       quarantined (version 1): ambiguous_recovery
```

이 알림이 말하는 것은 "032820의 손절이 지금 평가되지 않는다"이다. 모바일에서 이것을
받는 사람은 032820이 무엇인지 모른다. 콘솔은 같은 종목을 **에이치엘비**로 부르고,
포지션 화면은 한화오션·일진전기·팔란티어·스페이스X·테슬라로 부른다. 알림만 숫자다.

엔진의 알림 조립부 14곳이 전부 영어이고 전부 `position.Symbol`만 쓴다.

```
adoption.go:356        c.position.Symbol + " was adopted into exit management"
adoption.go:421        p.Symbol + " is held with no record justifying a stop, ..."
exitloop.go:1422       "the exit policy could not judge " + m.position.Symbol
exitloop.go:1441       "the exit proposal for " + m.position.Symbol + " was not submitted"
exitwiring.go:138      alert.Symbol + " was closed outside the engine while ..."
... 외 9곳
```

**종목명은 이미 엔진이 60초마다 받고 있다.** 공식 API의 `GET /api/v1/holdings` 응답
항목에 `name`이 있고, `official.apiHoldingsItem.Name`으로 파싱까지 되어 있다. 그런데
엔진이 쓰는 raw 경로가 그것을 버린다.

```go
// official.RawHolding — Name 없음
// reconcile.RawHolding — Name 없음
// reconcile.Holding    — Name 없음
```

같은 payload를 float로 읽는 `adaptHoldings`는 `Name`을 채운다. 콘솔이 종목명을 아는
것은 그 경로를 쓰기 때문이다. 즉 **추가 API 호출 없이** 이름을 알 수 있다 — §0.4
예산은 변하지 않는다.

### 화면 (PLTR, 2026-08-05 관측)

포지션 화면의 PLTR은 라인 5칸이 전부 `—`이고, 상세도 전부 `—`다.

```
익절 —   손절 —   추적 회수 —   기준 —   고점 —
진입가 — · 관측가 —      현재 보호선 — · 다음 익절 —
최초 손절 — · 워터마크 —  다음 보호선 — · 예상 수량 —
```

그런데 같은 화면이 decision id·snapshot id·observation id·정책·평가 시각은 보여준다.
그리고 **포지션 관리 화면은 같은 포지션에 대해 "유지 보호선 141.717"을 보여준다.**

값은 원장에 있다. `exit_states`에 entry 146.1, initial stop 141.717, baseline 141.717,
high water 148.55, next target 148.7298가 그대로 있다. 버리는 것은 화면이다 —
`operatorview.BuildExitLine`이 `snapshot_quarantined`에서 조기 반환하며 provenance만
남기고 숫자를 전부 `—`로 만든다.

두 화면이 같은 사실에 대해 서로 다른 말을 하고, 그중 더 위험한 쪽이 기본 화면이다.
운영자는 "이 포지션이 무엇을 붙들고 있는가"를 알 수 없다.

## What Changes

### 1. 알림에 한국어와 종목명

- `official.RawHolding`·`reconcile.RawHolding`·`reconcile.Holding`에 `Name`을 추가한다.
  비교에는 쓰이지 않는 표시 전용 필드이며, `CostBasisRaw`가 이미 만든 선례를 따른다.
- 엔진이 대사 주기마다 심볼→이름을 갱신하는 registry를 유지한다. 추가 요청 없음.
- 알림 제목·본문을 한국어로 쓰고 종목은 `한화오션(042660)` 형식으로 부른다.
  이름을 모르면 `042660`으로만 부른다 — 없는 이름을 지어내지 않는다.
- 로그의 구조화 필드(`symbol`, `position_id` 등)는 영문 키·원문 값 그대로 둔다.
  기계가 읽는 것과 사람이 읽는 것은 다른 표면이다.

### 2. 격리된 포지션이 무엇을 붙들고 있는지 보여준다

- `operatorview.BuildExitLine`이 `snapshot_quarantined`일 때, 저장된 값을 **유지 중**
  이라고 명시해 보여준다. 관리 화면이 "유지 보호선"으로 이미 하는 것과 같은 말을
  포지션 화면도 하게 한다.
- 갱신되는 값과 얼어 있는 값을 화면에서 구별할 수 있어야 한다. 얼어 있는 값을 살아
  있는 보호선처럼 보이게 해서는 안 된다.
- 다른 stale 사유(`engine_not_running`, `observation_older_than_limit` 등)의 현재
  동작은 바꾸지 않는다. 그것들은 "값이 오래됐다"이고, 격리는 "값이 고정됐고 그 값이
  무엇인지 알려져 있다"로 성격이 다르다.

## Impact

- **Specs**: `engine-safety` (ADDED 1 — 알림 가독성), `operator-console` (ADDED 1 — 격리 표시)
- **Code**: `internal/official/asset_reads.go`, `internal/reconcile/snapshot.go`,
  `internal/app/engine/` (알림 조립부 14곳 + registry),
  `internal/operatorview/exit_line.go`, `internal/console/templates_portfolio.go`
- **Schema**: 없음
- **§0.4**: 요청 수 변화 없음. 같은 `GET /api/v1/holdings` 응답의 기존 필드를 쓴다.
- **§0.8**: 종목명은 공개 정보다. 계좌번호·잔고·세션은 알림에 넣지 않는다 (현행 유지).

## Non-goals

- 알림 payload(JSON 필드)의 한국어화. 기계 판독 표면은 그대로 둔다.
- 콘솔 전체의 문구 재작성. 이미 한국어다.
- 종목명을 위한 별도 메타데이터 API 호출. holdings 응답으로 충분하고, 보유하지 않은
  종목의 이름은 알림에 필요하지 않다.
- 격리 외 stale 사유의 표시 규칙 변경.
