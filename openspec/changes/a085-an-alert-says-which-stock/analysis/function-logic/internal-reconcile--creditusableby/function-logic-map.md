# Function Logic Map: `creditUsableBy`

- Source: `internal/reconcile/mismatch.go` (lines 379–389)
- AST evidence: `ast.json` (`source_sha256: 49fcb007cab394f144230db5b1bf330b95e2303dfef48648edc806501edd2a57`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk**

## What it does

이 credit이 비교 `observed`의 관측으로 소비될 수 있는지 판정한다.
엄격히 나중, 그 외 전부 false. 같음은 조정이 계산된 바로 그 비교라서 재조회가 아니다 —
그 동등성이 a083 결함의 전부다. 어느 쪽이든 없거나 읽을 수 없으면 순서를 세울 수 없다.
셋 다 블록을 유지하는 보수 방향이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `credit` | RFC3339 비교 스탬프 | `Converger`가 `Diff.AsOf`로 기록 | 파싱 실패면 false — 블록 유지 |
| `observed` | RFC3339 비교 스탬프 | 지금 관측의 `Diff.AsOf` | 파싱 실패면 false — 블록 유지 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Required test |
|---|---|---|---|
| B1 | (381) `if` — if !ok { | 본문 참조 | 아래 Branch Test Map |
| B2 | (385) `if` — if err != nil { | 본문 참조 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `creditStampAt` | credit 스탬프 파싱 | not-ok면 false | AST `calls` L380 |
| `time.Parse` | 관측 스탬프 파싱 | 오류면 false | AST `calls` L384 |

## State mutations and fallbacks

- 없음. 순수 함수.

## Safety conclusion

- Safe edit boundary: `After` 비교. `!Before`로 바꾸면 조정이 계산된 그 비교가 자기를 답으로 인정한다 — a083 결함의 재현이다.
- High-risk impact: yes — 진입 차단 해제 판정.
