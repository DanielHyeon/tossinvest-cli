# Function Logic Map: `ExitObserver.judgeRatchet`

- Source: `internal/app/engine/exitloop.go` (lines 861–914)
- AST evidence: `ast.json` (`source_sha256: 6625c92061d5b05f566ecb0913f5c5f74a7fdde4cc4b5d8e7dfe8e75dd71de00`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk**

## What it does

ratchet 정책 평가. 개정 2에서 바뀐 것은 조기 반환 조건 한 곳뿐이다:
`if !snapshot.Changed` 가 `if !snapshot.Changed && !m.reJudge` 로 바뀌었다. 재판정에서
바뀐 것은 라인이 아니라 선택기이므로 `Changed`는 이를 알 수 없고, 여기서 돌아가면 이미
소비된 재시도를 아무것도 판정하지 않고 태운다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `m.policyIdentity` | 저널이 보존한 정책 동일성 | `exit_states` | 활성 ratchet와 불일치면 거절 — 판정 없음 |
| `m.reJudge` | bool | `workingSet` → `judge` | false면 upstream 동작과 동일: 변화 없으면 행을 남기지 않는다 |
| `quote.Price` | 관측 가격 | 시세 경로 | `snapshotContext` 오류로 거절 |
| `o.ratchet` | 활성 ratchet config | 런타임 config | 동일성 대조로 교체를 잡는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Required test |
|---|---|---|---|
| B1 | (864) `if` — if err != nil \|\| !samePolicyIdentity(m.policyIdentity, activeIdentity) { | 본문 참조 | 아래 Branch Test Map |
| B2 | (865) `if` — if err == nil { | 본문 참조 | 아래 Branch Test Map |
| B3 | (875) `if` — if err != nil { | 본문 참조 | 아래 Branch Test Map |
| B4 | (895) `if` — if err != nil { | 본문 참조 | 아래 Branch Test Map |
| B5 | (903) `if` — if !snapshot.Changed && !m.reJudge { | 본문 참조 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `exitpolicy.RatchetPolicyIdentity` | 활성 정책 동일성 | 오류면 거절 | AST `calls` L863 |
| `exitpolicy.EvaluateRatchetSnapshot` | 불변 스냅샷 평가 | 오류면 거절 — 제안 없음 | AST `calls` L894 |
| `snapshot.ChangedFromState` | 저장 상태 대비 변화 판정 | 순수 함수 | AST `calls` L901 |
| `o.record` | 판정 기록과 무장 | 오류 전파 | AST `calls` L911 |

## State mutations and fallbacks

- `o.clearRefused(m.position.ID)` — 평가가 성공했으므로 거절 래치를 푼다.
- 행 기록·무장·제출은 `record`가 한다. 이 함수는 아무것도 쓰지 않는다.

## Safety conclusion

- Safe edit boundary: 조기 반환 조건. 평가 입력 구성은 upstream 그대로다.
- High-risk impact: yes — 손절 판정 경로. `&& !m.reJudge`를 빼면 재시도가 소비된 채 격리가 영구히 남는다.
