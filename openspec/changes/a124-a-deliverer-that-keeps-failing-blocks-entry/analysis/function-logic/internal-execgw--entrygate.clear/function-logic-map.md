# Function Logic Map: `EntryGate.Clear`

- Source: `internal/execgw/retry.go`
- AST evidence: `ast.json` — **편집 전**, :537–544, 분기 1 · 반환 0 · 호출 3 · defer 1.
  source_sha256 `15d3f65344f9…`, 추출 base `4798d399` (2026-09-26, Teammate — design 5판 로트).
- Risk scan: `risk-pattern-report.md`

**a124 5판은 이 함수를 편집한다**(M1 = B′ — 해제 세대 증가). design D7 이 이 함수의 분기를 근거로 쓰기 때문에 문서보다 먼저 만들었다.
편집 뒤 `revision: current` 재추출(tasks 1.3).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `reason` | `ReasonCode` | 호출자 | — |
| `g.latches` | `g.mu` 아래에서만 | EntryGate | — |
| `ReasonAlertUndelivered` 를 지우는 자리 | **이 함수뿐** — 다른 직접 삭제는 `modegate.go:37`(`ReasonOperatingModeBlocked`)·`symbolgate.go:189`(대사 계열)뿐(grep) | HEAD | — |
| 이 사유의 호출자 | `Notifier.Acknowledge` :846(무원장) · :875(미전달 0) 둘뿐(grep) | HEAD | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 그 사유의 래치가 있음 (:540) | `delete(g.latches, reason)` :541 · `g.revision++` :542 | 종단 | obs 승인 해제 시험 (1.4 에서 이름 확정) |
| (B1 거짓) | 래치 없음 | **아무것도 바꾸지 않는다** — `revision` 도 불변 | 종단 | 1.4 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `g.mu.Lock` :538 / `defer g.mu.Unlock` :539 | 래치 map 보호 | — | AST |
| `delete` :541 | 래치 제거 | — | AST |

## State mutations and fallbacks

- `revision` 은 **실제로 지웠을 때만** 오른다(전략 진입 ABA 봉인의 세대 — `strategy_entry_gate_authority.go`). a124 는 `revision` 의 의미를 바꾸지 않는다.
- 해제 요청이 래치가 없는 게이트에 오면 흔적이 남지 않는다 — 그 요청이 「사람이 backlog 를 비웠다」는 사실이어도 그렇다. design 5판 D7 이
  해제 세대를 이 분기와 **무관하게** 올리는 이유(판별 케이스 ㉣).

## Safety conclusion

- Safe edit boundary: 함수 머리(잠금 뒤)에서 사유별 해제 세대 `clearEpochs[reason]++` 한 줄(맵은 지연 생성). B1 과 `revision` 은 그대로.
- High-risk impact: yes — 진입 게이트 해제 경로.
