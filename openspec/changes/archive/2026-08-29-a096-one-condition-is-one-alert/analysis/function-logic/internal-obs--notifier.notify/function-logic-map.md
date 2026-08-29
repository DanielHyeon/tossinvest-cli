# Function Logic Map: `Notifier.Notify`

- Source: `internal/obs/notifier.go` (124–133)
- AST evidence: `ast.json` (sha256 `c9dee3479706…`, 1분기, 반환 2곳)
- Risk scan: `risk-pattern-report.md`

a096은 **이 함수를 편집하지 않는다.** 여기 있는 이유는 하나다: 이 함수가 구조화 로그를
등급 분기보다 **먼저** 남긴다는 사실이, "전송을 억제해도 관측 기록은 남는다"는 a096의
핵심 주장의 근거이기 때문이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `e.Type` | `obs.EventType` | `event.go`의 enum | 미등록 타입은 normal로 등급된다 |
| 등급 | critical / normal | `SeverityOf(e.Type)`, `criticalEvents` map | 이 함수가 판정한다 |
| `n.Log` | nil 가능 | 조립 시점 | `logEvent`가 nil을 확인하고 반환 |

계약(주석): 오류는 **critical을 durable하게 만들지 못했을 때만** 올라간다. 전송 실패는
호출자의 문제가 아니다 — 여기서 이미 처리됐다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1@128 | `severity != SeverityCritical` | `publishBestEffort` | `nil` @130 | 기존 `TestOrdinaryAlertsAreBestEffort` (진입) |
| — | (critical 경로) | `notifyCritical` | 그 반환 @132 | 기존 다수 |

## 로그가 등급보다 먼저다

```go
func (n *Notifier) Notify(ctx context.Context, e Event) error {
	severity := SeverityOf(e.Type)
	n.logEvent(e, severity)          // :126 — B1보다 앞
	if severity != SeverityCritical { // :128 B1
```

`logEvent`는 B1 **앞**에 있다. 따라서 a096이 `notifyCritical` 안에서 전송을 억제해도
구조화 로그 한 줄은 관측마다 그대로 남는다. 억제되는 것은 push뿐이고 관측 사실의
기록이 아니라는 spec 문장은 이 배치에 근거한다.

운영 로그가 이것을 확인한다: 2026-08-08 사고에서 `exit.proposal_refused` WARN 줄은
60건 남았고 outbox 행은 1건이었다. a096 이후에도 WARN 60건은 그대로 남고 push는 창당 1건이 된다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `SeverityOf` | 등급 판정 | 없음 — map 조회 | AST :125 |
| `n.logEvent` | 관측 기록 | nil 로거 안전 | AST :126 |
| `n.publishBestEffort` | normal 경로 | 오류를 삼키고 로그만 | AST :129 |
| `n.notifyCritical` | critical 경로 | 유일하게 오류가 올라오는 곳 | AST :132 |

## State mutations and fallbacks

- 이 함수 자체는 원장에 쓰지 않는다.
- a096이 `criticalEvents`를 바꾸지 않으므로 등급 분포도 바뀌지 않는다.

## Safety conclusion

- Safe edit boundary: **없음 — 편집하지 않는다.**
- High-risk impact: no.
