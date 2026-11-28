# Function Logic Map: `LegacyLadderPolicyIdentity`

- Source: `internal/exitpolicy/identity.go` (51-79)
- Qualified: `LegacyLadderPolicyIdentity`
- AST evidence: `ast.json` (source_sha256 `25c77efa29b1…`) — AST branches 9 · returns 3
- **편집 뒤 갱신(a208 F3, 2026-11-28)**: `ast.json` 을 편집 뒤 소스로 재추출(`revision: current`). 아래 분기 표는 편집 뒤 좌표이고, 편집 전 좌표는 커밋 233897096 의 같은 파일에 있음. 「Safe edit boundary」는 편집 전 계획이며 그대로 구현됨.
- Risk scan: `risk-pattern-report.md`(일치 0)
- 작성 시점: 편집 **전**(a208 F3 재작업, 2026-11-28, HEAD `468ec87ce` — base `baeddfe31` 와 이 함수 본문 동일)

**역할.** 정체성 없이 정책 ID 만 가진 ladder 행 · 시드를 고정 version/digest 로 묶는다. 생산 호출자 셋(CodeGraph 1.6.0):
`seedPolicyIdentity`(journal/exit_state.go:215 — 새 exit state 의 정체성 결정), `legacyPolicyIdentity`(journal/exit_snapshot.go:718 — v10 열 없는 옛 행),
`managedPolicyIdentity`(engine/exitloop.go:768 — 저장 정체성이 비었을 때만).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `id` | 빈 값(=default_v1) · default_v1 · 등록 공통 정책 ID | 시드 · exit state 행 | 그 밖 → 정체성 충돌 오류(B8) |
| `adopted` | 편입 포지션 여부 | positions.adoption_id | RUNNER 만 의미가 갈림(B6) |
| 불변식 | ID 만 있는 행에 오늘 등록부를 대입하지 않음 — 고정 리터럴만 | 함수 주석 | 미지 ID 는 판정 정지 |

## Branches and early returns

| Branch | Condition | At | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if id == ""` | 53:2 | id = default_v1 | — | 미측정 |
| B2 | `switch id` | 58:2 | — | — | — |
| B3 | case default_v1 | 59:2 | digest | — | `TestLegacyPolicyIdentityIsExactAndUnknownIsFailClosed` |
| B4 | case BALANCED | 61:2 | digest | — | 미측정 |
| B5 | case RUNNER | 63:2 | digest | — | `TestApprovedPolicyTablesMatchPinnedSemanticDigests`(간접) |
| B6 | `if adopted`(RUNNER 안) | 65:3 | 편입 RUNNER version/digest | — | `TestA208OnWithRunnerKeepsTheExistingAdoptedRunnerCopy` |
| B7 | case HYBRID_50 | 68:2 | digest | — | `TestA208AdoptedHybridRowsKeepTheirMeaning` |
| B8 | case 편입 보호선 사본 — **a208 F3 새 갈래** | 70:2 | — | return adoptedFloorOnlyIdentity(adopted) (72:3) | `TestA208FloorOnlyCopyKeepsTheHybridFloors` · `TestA208FloorOnlyCopyIsAdoptionOnly` |
| B9 | default | 73:2 | — | return 정체성 충돌 (74:3) | `TestLegacyPolicyIdentityIsExactAndUnknownIsFailClosed` |

Exact AST return positions: 72:3, 74:3, 78:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `PolicyIdentity.Validate` | 형식 검사 | 오류 반환 | AST |

Live config binding: 없음.

## State mutations and fallbacks

- 없음(순수 함수). fallback 없음 — 미지 ID 는 거부.

## Safety conclusion

- Safe edit boundary(a208 F3): B7 과 B8 사이에 `case CommonLadderHybrid50AdoptedFloorOnly: return adoptedFloorOnlyIdentity(adopted)` 한 갈래.
  기존 case(HYBRID_50 편입 · 비편입 모두 1.0.0/hybrid50 digest) 무편집 → 이미 편입된 HYBRID_50 행 의미 무변. 새 ID 는 adopted 거짓이면 오류.
- High-risk impact: yes(새 exit state 정체성 결정).
