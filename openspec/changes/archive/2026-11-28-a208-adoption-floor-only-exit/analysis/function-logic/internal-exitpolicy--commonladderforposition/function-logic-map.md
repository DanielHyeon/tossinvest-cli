# Function Logic Map: `CommonLadderForPosition`

- Source: `internal/exitpolicy/common_policy.go` (106-131)
- Qualified: `CommonLadderForPosition`
- AST evidence: `ast.json` (source_sha256 `34b6b0fa45f1…`) — AST branches 6 · returns 5
- **편집 뒤 갱신(a208 F3, 2026-11-28)**: `ast.json` 을 편집 뒤 소스로 재추출(`revision: current`). 아래 분기 표는 편집 뒤 좌표이고, 편집 전 좌표는 커밋 233897096 의 같은 파일에 있음. 「Safe edit boundary」는 편집 전 계획이며 그대로 구현됨.
- Risk scan: `risk-pattern-report.md`(go-panic 3 건 — 전부 이 함수 밖: buildCommonPolicies 74·78, RegisteredCommonPolicyDescriptors 167)
- 작성 시점: 편집 **전**(a208 F3 재작업, 2026-11-28, HEAD `468ec87ce` — base `baeddfe31` 와 이 함수 본문 동일)

**역할.** exit state 에 스냅숏된 정책 ID 를 판정용 사다리 표로 해소한다. 생산 호출자 `ExitObserver.ladderFor`(exitloop.go:1324) 하나(CodeGraph 1.6.0).
시험 호출자: common_policy_test · policy_identity_contract_test · journal a078.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `id` | 등록 공통 정책 ID(BALANCED · RUNNER · HYBRID_50) | `exit_states.policy_id` | 미등록 → 오류(판정 거부, B1) |
| `adopted` | 포지션이 편입 기록을 가짐 | `Position.Adopted()` | — |
| 불변식 | 같은 (id, adopted) 는 언제나 같은 표 · 같은 정체성 | 고정 digest | 표가 바뀌면 정체성 충돌 오류(B4 · B5) |

## Branches and early returns

| Branch | Condition | At | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if IsAdoptedFloorOnlyPolicyID(id)` — **a208 F3 새 갈래** | 108:2 | — | return adoptedFloorOnlyLadder(adopted) (109:3) | `TestA208FloorOnlyCopyKeepsTheHybridFloors` · `TestA208FloorOnlyCopyIsAdoptionOnly` |
| B2 | `if !ok` — 미등록 ID | 112:2 | — | return 오류 (113:3) | 미측정 |
| B3 | `if adopted && policy.ID == CommonLadderRunner` | 115:2 | 사본의 부분익절 0 · 정체성 교체 | — | `TestExternalRunnerHasNoAutomaticPartials` |
| B4 | `range policy.Ladder.Rungs` | 116:3 | `PartialRatio = "0"` | — | `TestExternalRunnerHasNoAutomaticPartials` |
| B5 | `if err != nil` — Identity 실패 | 122:3 | — | return 오류 (123:4) | 미측정(고정 표라 도달 불가) |
| B6 | `if identity.Digest != adoptedRunnerPolicyDigest` | 125:3 | — | return 정체성 충돌 (126:4) | 미측정(고정 표라 도달 불가) |

Exact AST return positions: 109:3, 113:3, 123:4, 126:4, 130:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `CommonPolicyByID` | 등록 표의 복사본 | 없으면 ok=false | AST |
| `LadderPolicy.Identity` | 편입 RUNNER 사본 정체성 재계산 | 오류 반환 | AST |

Live config binding: 없음 — 설정이 아니라 exit state 의 정책 ID 를 읽음.

## State mutations and fallbacks

- 복사본(cloneCommonPolicy)만 고침 — 전역 표 무변. fallback 없음(미등록은 거부).

## Safety conclusion

- Safe edit boundary(a208 F3): B1 **앞**에 새 갈래 하나 — `id` 가 편입 보호선 사본 ID(`COMMON_LADDER_HYBRID_50_ADOPTED_FLOOR_ONLY`)일 때만 새 파일 함수로 해소하고 반환.
  B1~B5 는 무편집 → 기존 ID(편입 HYBRID_50 행 포함)의 해소 결과 무변. 새 갈래는 `adopted` 거짓이면 오류(fail-closed).
- High-risk impact: yes(익절 · 보호선 표 해소). 새 ID 는 공통 정책 등록부에 없으므로 `CommonPolicyByID` 를 쓰는 설정 · 콘솔 · override 검증은 새 ID 를 받지 않음.
