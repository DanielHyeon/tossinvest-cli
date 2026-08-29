# Function Logic Map: `ExitObserver.announceQuarantine`

- Source: `internal/app/engine/exit_quarantine_announce.go` (lines 59–90)
- AST evidence: `ast.json` (`source_sha256: 69734107d4f6e87e0e5b9b722b6442315fde8f1fc20718c4d9f290c7b57b4c22`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal** — 알림 문구와 표시 전용 배선. 주문·손절·사이징·원장 판정 경로를 바꾸지 않는다.

## What it does

알림 조립부. a085는 제목·본문을 한국어로 바꾸고 종목을 `이름(코드)`로 부른다. 이벤트 타입·Key·Fields(기계 판독 표면)와 발송 조건·latch는 그대로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `Names` registry | nil 허용 | 대사 루프가 holdings 스냅샷에서 채운다 | nil이면 코드만 렌더 — a085 이전 동작 |
| 브로커 `name` | 빈 문자열 허용 | `GET /api/v1/holdings` 기존 필드 | 비면 코드만 렌더. 추정하지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (61) `if` — if o.quarantineAnnounced == nil | 본문 참조 | — | 아래 Branch Test Map |
| B2 | (67) `if` — if o.quarantineAnnounced[key] | 본문 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `quarantineAnnouncementKey` | (66) key := quarantineAnnouncementKey(p.ID, q.PositionGeneration, q.Version) | 호출부 계약 유지 | AST `calls` |
| `o.alert` | (71) o.alert(ctx, obs.Event{ | 호출부 계약 유지 | AST `calls` |
| `o.label` | (74) Title: o.label(p.Symbol) + " 판정 격리 — 이제 손절도 평가되지 않는다", | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 알림 제목·본문 문자열만. 이벤트 타입·Key·Fields·발송 조건·latch·원장은 바뀌지 않는다.

## Safety conclusion

- Safe edit boundary: 사람이 읽는 문자열과 표시 전용 필드. 기계 판독 표면(payload·로그 키·원장 cause)은 영문·원문 그대로다.
- High-risk impact: no — 판정도 주문도 이 경로에 없다. §0.4는 불변이다: 이름은 이미 대금을 지불한 응답에서 온다.
