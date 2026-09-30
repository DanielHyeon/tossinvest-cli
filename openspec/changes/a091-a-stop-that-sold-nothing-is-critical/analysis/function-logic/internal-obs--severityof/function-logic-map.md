# Function Logic Map: `SeverityOf`

- Source: `internal/obs/event.go` (`379`–`384`)
- Qualified: `SeverityOf`
- AST evidence: `ast.json` (`source_sha256` 33221d38f60885a6…) — 편집 뒤 `3ec1efd2` 에서 `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 1 · 반환 2

**편집.** 함수 본문은 편집하지 않는다 — a091 은 등급표 `criticalEvents`(`event.go:337-361`)에 새 종류 한 줄을 더한다(tasks 2.4). 판정 방식의 근거.

**역할.** 등급은 종류에만 붙는다: `criticalEvents` 맵 조회 하나. 미등록은 normal(기본값 — 주석 `event.go:363-370`).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t` | 임의 문자열 | 호출자 | 미등록 → normal |
| `criticalEvents` | **19** 종(base b30318d6: 주문 · 브로커 · 알림 · 모드 · 루프 13 + exit 관측 5 + a095 편입 실패 1) | `event.go:337-361` | — |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 편집 뒤 `3ec1efd2` 의 `go test -count=1 -coverprofile`(covermode set, 2026-10-01)에서 그 줄로 시작하는 블록의 count (`analysis/harness/write_bundles.py`).

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:380` `if criticalEvents[t] {` | 예 |

Exact AST return positions: `381:3`, `383:2`

## Calls and live bindings

호출 없음(`ast.json` `calls` 0).

## State mutations and fallbacks

없음.

## Safety conclusion

- 같은 종류 안에서 등급을 나눌 수 없다 → 0주만 critical 로 올리려면 새 종류가 필요하다(D1 B). a095 가 같은 모양으로 `EventExitPositionAdoptionFailed` 를 신설했다(선례). High-risk: yes(등급표 = 진입 차단 스위치).
