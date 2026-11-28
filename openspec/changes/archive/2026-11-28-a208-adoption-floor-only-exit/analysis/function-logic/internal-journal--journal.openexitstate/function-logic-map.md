# Function Logic Map: `Journal.OpenExitState`

- Source: `internal/journal/exit_state.go` (117-205)
- Qualified: `Journal.OpenExitState`
- AST evidence: `ast.json` (source_sha256 `e9b375b4fcc5…`) — AST branches 20 · returns 15
- **편집 뒤 갱신(a208 F3, 2026-11-28)**: `ast.json` 을 편집 뒤 소스로 재추출(`revision: current`). 아래 분기 표는 편집 뒤 좌표이고, 편집 전 좌표는 커밋 233897096 의 같은 파일에 있음. 「Safe edit boundary」는 편집 전 계획이며 그대로 구현됨.
- Risk scan: `risk-pattern-report.md`(일치 0)
- 작성 시점: 편집 **전**(a208 F3 재작업, 2026-11-28, HEAD `468ec87ce` — base `baeddfe31` 와 이 함수 본문 동일)

**역할.** 포지션 하나의 보호 상태(exit_states 행)를 t0 로 연다. 호출자(CodeGraph 1.6.0): `OpenAdoptedExitState`(journal/adoption.go:284 — 편입 기록의 정책 ID 그대로),
`ExitObserver.openState` 비편입 갈래(exitloop.go:826), 시험 다수.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `seed.PositionID` | 존재 · exit 대상 포지션 | positions | 오류(B1 · B13~B15) |
| `seed.PolicyKind` | RATCHET · LADDER(빈 값 = RATCHET) | 호출자 | 그 밖 → 오류(B3) |
| `seed.PolicyID` | RATCHET: 빈 값 / LADDER: 빈 값(default_v1) · default_v1 · 등록 공통 정책 | 편입 기록 또는 공통 정책 | 미등록 → 오류(B10) |
| `seed.PolicyIdentity` | 비었거나 고정 정체성과 정확히 같음 | seedPolicyIdentity | 충돌 → 오류(B16) |
| 불변식 | 행이 열리는 순간 baseline = 진입 손절 | 함수 주석 | — |

## Branches and early returns

| Branch | Condition | At | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | PositionID 없음 | 119:2 | — | return (120:3) | 미측정 |
| B2 | kind 빈 값 | 123:2 | kind = RATCHET | — | 미측정 |
| B3 | kind 미지 | 126:2 | — | return (127:3) | 미측정 |
| B4 | switch kind | 131:2 | — | — | — |
| B5 | case RATCHET | 132:2 | — | — | 미측정 |
| B6 | RATCHET 에 정책 ID | 133:3 | — | return (134:4) | 미측정 |
| B7 | case LADDER | 136:2 | — | — | `TestOpenExitStatePreservesPinnedRuntimePolicyIdentity` |
| B8 | LADDER 빈 ID | 137:3 | policyID = default_v1 | — | 미측정 |
| B9 | else | 139:10 | — | — | — |
| B10 | `!ok && policyID != "default_v1" && !IsAdoptedFloorOnlyPolicyID(policyID)` — **a208 F3 조건 확장** | 139:10 | — | return 미지 정책 (141:4) | `TestA208FloorOnlyCopyNeedsAnAdoptedPosition`(사본 ID 는 통과해 B16 로) |
| B11 | OpenRatchetState 실패 | 148:2 | — | return (149:3) | 미측정 |
| B12 | BeginTx 실패 | 154:2 | — | return (155:3) | 미측정 |
| B13 | 포지션 없음 | 164:2 | — | return (165:3) | 미측정 |
| B14 | 포지션 읽기 실패 | 167:2 | — | return (168:3) | 미측정 |
| B15 | exit 대상 아님 | 170:2 | — | return (171:3) | 미측정 |
| B16 | seedPolicyIdentity 실패 | 175:2 | — | return (176:3) | `TestOpenExitStateRejectsASeedWhosePolicyMeaningDoesNotMatchItsID` · `TestA208FloorOnlyCopyNeedsAnAdoptedPosition` |
| B17 | INSERT 실패 | 179:2 | — | — | — |
| B18 | 고유 위반(이미 있음) | 188:3 | — | return ErrExitStateExists (189:4) | 미측정 |
| B19 | 이벤트 기록 실패 | 193:2 | — | return (197:3) | 미측정 |
| B20 | Commit 실패 | 199:2 | — | return (200:3) | 미측정 |

Exact AST return positions: 120:3, 127:3, 134:4, 141:4, 149:3, 155:3, 165:3, 168:3, 171:3, 176:3, 189:4, 191:3, 197:3, 200:3, 204:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `exitpolicy.CommonPolicyByID` | LADDER 정책 ID 등록 여부 | ok | AST |
| `exitpolicy.OpenRatchetState` | t0 산술 | 오류 → ErrInvalidRequest | AST |
| `seedPolicyIdentity` → `LegacyLadderPolicyIdentity(id, adopted)` | 고정 정체성 결정(편입 여부는 positions.adoption_id) | 오류 반환 | AST |
| SQL INSERT · `appendExitEventTx` | 행 · 이력 한 트랜잭션 | 오류 반환, 고유 위반은 ErrExitStateExists | AST |

Live config binding: 없음.

## State mutations and fallbacks

- exit_states 행 1 + exit 이벤트 1(한 트랜잭션). fallback 없음.

## Safety conclusion

- Safe edit boundary(a208 F3): B10 조건에 `&& !exitpolicy.IsAdoptedFloorOnlyPolicyID(policyID)` 를 더함(분기 수 무변). 사본 ID 의 정체성은 B16 의
  seedPolicyIdentity 가 `adopted`(positions.adoption_id 유무)로 정하고, 비편입이면 LegacyLadderPolicyIdentity 가 거부 → B16 오류(fail-closed).
- High-risk impact: yes(보호 상태 개시 · 정책 정체성 고정).
