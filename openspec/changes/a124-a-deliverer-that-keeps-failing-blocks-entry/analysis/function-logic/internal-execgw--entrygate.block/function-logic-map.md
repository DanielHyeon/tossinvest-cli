# Function Logic Map: `EntryGate.Block`

- Source: `internal/execgw/retry.go`
- AST evidence: `ast.json` — **편집 전**, :526–533, 분기 1 · 반환 0 · 호출 2 · defer 1.
  source_sha256 `15d3f65344f9…`, 추출 base `4798d399` (2026-09-26, Teammate — design 5판 로트).
- Risk scan: `risk-pattern-report.md`

**a124 는 이 함수를 편집하지 않는다.** 새 `BlockUnlessClearedSince` 가 이 함수와 같은 삽입 규칙을 따른다는 주장의 근거로 만들었다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `reason` · `detail` | — | 호출자 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 그 사유의 래치가 없음 (:529) | `g.latches[reason] = detail` :530 · `g.revision++` :531 | 종단 | execgw 게이트 시험 (1.4) |
| (B1 거짓) | 이미 있음 | 무변화 — **처음 detail 이 남는다** | 종단 | 1.4 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `g.mu.Lock` :527 / `defer g.mu.Unlock` :528 | 래치 map 보호 — 밖을 부르지 않는다 | — | AST |

## State mutations and fallbacks

- 없을 때만 삽입. 재호출은 무변화(design D1 F9).

## Safety conclusion

- Safe edit boundary: 편집하지 않는다. 새 메서드는 같은 잠금 아래 세대 비교를 앞에 둔 같은 삽입이다.
- High-risk impact: yes — 진입 게이트(읽기 전용 근거).
