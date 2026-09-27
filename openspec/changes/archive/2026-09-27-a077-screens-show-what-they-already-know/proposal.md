# a077 · 화면은 이미 알고 있는 것을 보여준다

- **Feature**: `FEAT-TOS-004` — Operator console controls and visibility
- **Story**: `STORY-TOS-a077`
- **Spec**: `operator-console`

## Why

2026-08-03 운영자가 세 화면에서 같은 종류의 결함을 신고했다. 세 건 모두 **콘솔이
이미 손에 들고 있는 사실을 화면에 쓰지 않는 것**이다.

1. `/positions`와 `/dashboard`의 `라인` 열이 전 종목 `—`다. 익절·손절·추적 회수·
   기준·고점이 하나도 없다.
2. `/position-management`의 계좌·종목 열에 종목명이 없다.
3. (`/orders`도 같은 증상이지만 원인이 다르다 — 후속 change에서 다룬다.)

## 결함 1 — `라인` 열: 엔진은 "안 바뀌면 저장하지 않고", 콘솔은 그 저장 시각을
"최근 관측 시각"으로 읽는다

`exit_states.last_observed_at`을 쓰는 코드 경로는 하나뿐이다 —
`internal/journal/exit_state.go`의 `record`. 그리고 관측 루프는 값이 바뀌지 않으면
`record`를 부르지 않는다.

```go
// internal/app/engine/exitloop.go — judgeRatchet, judgeLadder 둘 다
snapshot = snapshot.ChangedFromState(m.state.HighWater, m.state.Baseline, …)
if !snapshot.Changed {
        // exit_events is append-only and the loop runs every five seconds. A
        // judgement that moved nothing is not a judgement worth a row; …
        return nil
}
return o.record(ctx, m, snapshot, …)
```

그래서 그 컬럼은 **"마지막으로 관측한 시각"이 아니라 "마지막으로 값이 바뀐 시각"**이다.
콘솔은 그것을 신선도 하트비트로 읽는다.

```go
// internal/console/portfolio_pages.go
snapshot := row.Exit.Snapshot.WithFreshness(asOf, holdingsTTL)   // holdingsTTL = 30초
```

그 30초는 **브로커 캐시 TTL**이다. 보호선의 수명과는 아무 관계가 없다. 넘으면
`BuildExitLine`이 `StaleReason`에서 곧장 반환하며 익절·손절·추적 회수·기준·고점을
전부 `—`로 닫는다.

**결과: 5초마다 정상 관측 중인 엔진이라도, 판정이 아무것도 바꾸지 않으면 30초 뒤
전 종목의 라인 열이 빈다.** 2026-08-03 11:27~11:34 운영 원장이 정확히 그 상태였다.

| 종목 | 기록된 고점 | 현재가 | 마지막 기록 |
|---|---|---|---|
| 439960 | 18,060 | 17,880 | 05:56 |
| IONQ | 37.41 | 36.59 | 08:05 |
| TSLA | 315.19 | 313.27 | 08:15 |
| 042660 | 83,500 (seed) | 82,000 | **없음 — `SEED`** |

전부 현재가가 기록된 고점 아래다. 올릴 고점도 올릴 rung도 없으니 판정이 아무것도
바꾸지 않고, 그래서 아무것도 저장되지 않는다. 20초 간격 3회 폴링에서 네 행의
`last_observed_at`은 전혀 움직이지 않았다.

`holdingsTTL`을 60초로 넓히는 것은 **고치는 게 아니다.** 평평한 포지션은 몇 시간씩
바뀌지 않는다.

## 결함 2 — 종목명: 그 열에 이름 필드가 없다

`/position-management`의 계좌·종목 셀은 이렇게 생겼다.

```html
<code>{{.State.AccountRef}}</code><span class="submetric">{{.State.Market}} · {{.State.Symbol}}</span>
```

이름을 쓸 자리가 아예 없다. `positionpolicy.State`에 `Name` 필드가 없고 journal의
`positions` 테이블에도 이름 컬럼이 없기 때문이다.

**US가 "나오는" 것처럼 보인 것은 착시다.** `IONQ`·`TSLA`는 티커가 사람이 읽을 수 있을
뿐이고, 아이온큐도 테슬라도 출력되지 않는다. KR만 6자리 숫자라 그 부재가 드러났다.

이름은 콘솔이 이미 갖고 있다. `/positions`와 `/dashboard`가 한화오션·코스모로보틱스·
아이온큐를 정확히 그리고 있고, 출처는 holdings 응답의 `name`이다. 그 응답은
`holdingsCache`에 있고 `peek`는 **브로커 호출 0회**로 읽는다
(change `console-operator-overview` D4).

## What Changes

### 라인 열 — 신선도 판정 근거를 바꾼다

콘솔은 **snapshot 자신의 나이**가 아니라 **그 값을 유지하는 주체가 살아 있는지**로
판정한다. 보호선·다음 익절·최초 손절·고점은 시간에 따라 변하는 측정값이 아니라
마지막 판정이 확정한 **상태**다. 상태는 나이로 썩지 않는다. 못 믿게 만드는 것은
셋뿐이다 — 그것을 갱신할 프로세스가 없거나, 그 포지션이 판정 대상에서 빠졌거나,
generation이 어긋났거나.

| 콘솔이 아는 것 | 판정 | 화면 |
|---|---|---|
| 엔진 마커 = 실행 중 | 라인은 살아 있다 | 다섯 값 표시 |
| 엔진 마커 = 정지 | 갱신 주체 없음 | `—` + `엔진 정지` |
| 엔진 마커 미배선 | 알 수 없음 | **기존 관측 나이 판정 그대로** |
| 활성 quarantine | 판정 대상 아님 | `—` + `판정 격리` |

세 번째 줄이 중요하다. 기존 spec은 "미배선을 정지로 표시해서는 안 된다"고 못박는다.
그래서 미배선 콘솔은 지금과 **완전히 같은 동작**을 유지한다 — 회귀 0.

### quarantine을 화면에 올린다 (이 change의 안전 전제)

지금 운영 원장에는 활성 quarantine이 하나 있다.

```
pos-3b14217c40e2a96c3f16c35e  466100  ambiguous_recovery
  evidence: exitpolicy: recovery candidate identity mismatch
  quarantined_at 2026-08-03T09:03:40Z   released_at NULL
```

엔진 로그의 표현으로 "this position is not being judged at all"이다.
**콘솔은 quarantine을 한 번도 읽은 적이 없다.** 지금 그 행이 `—`인 것은 우연히
나이로 걸린 덕이다. 신선도 근거만 바꾸고 quarantine을 안 읽으면, 판정되지 않는
포지션에 실행 가능한 보호선을 그리게 된다. 그래서 quarantine 읽기는 스코프 확장이
아니라 이 change가 안전하려면 반드시 함께 있어야 하는 절반이다.

### `/position-management`가 이름을 보여준다

핸들러가 `holdingsCache.peek`로 이름을 붙인다. 브로커 호출 0회, 모든 행이 보유
포지션이므로 커버리지 100%다. 캐시가 비어 있으면 이름 없이 지금과 같이 그린다 —
없는 이름을 지어내지 않는다.

## Non-Goals

- **엔진을 고치지 않는다.** `last_observed_at`을 정직한 하트비트로 만드는 것은
  옳은 수정이지만 exit 경로(High-risk)이고 별도 change다. a061은 그것 없이도
  화면을 정직하게 만들고, 그 change가 오면 판정 조건 하나가 **추가**될 뿐 여기서
  만든 것은 버려지지 않는다.
- **quarantine을 해제하지 않는다.** 466100의 해제는 운영자 판단이고 §0.7이다.
  a061은 그 사실을 화면에 **표시**만 한다.
- `/orders`의 종목명은 다루지 않는다. 원인이 다르고(공식 주문 스키마에 이름이 없다)
  새 API capability와 rate budget 항목이 필요하다.
- `cmd/tossctl/httpapi_reader.go`도 같은 결함을 갖고 있다(`issues.md` I1).
  a061은 콘솔 화면만 고친다.
- 보호선 값을 계산하거나 원장 raw 값으로 대체하지 않는다. 기존 fail-closed 계약은
  그대로다.

## Impact

- `internal/journal/account_views.go` — `PositionExit`에 활성 quarantine 부착(additive),
  `ReadOnly`에 조회 전용 read 1건.
- `internal/operatorview/exit_line.go` — 새 사유 문자열 3건과 그에 맞는 상태 문구.
  값 계산은 그대로 없다.
- `internal/console/portfolio_pages.go` — 신선도 판정 교체.
- `internal/console/portfolio.go` — join이 quarantine을 행에 옮긴다.
- `internal/console/position_policy.go`, `templates_position_policy.go` — 이름.
- `openspec/specs/operator-console` — ADDED 2건. 기존 요구사항은 MODIFY하지 않는다
  (a055 `issues.md` I1의 미아카이브 MODIFY 부채를 늘리지 않는다 — a059·a060과 같은 이유).
