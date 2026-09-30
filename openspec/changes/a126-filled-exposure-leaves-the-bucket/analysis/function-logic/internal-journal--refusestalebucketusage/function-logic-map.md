# Function Logic Map: `refuseStaleBucketUsage`

- Source: `internal/journal/risk_bucket_usage.go` (`30`–`84`)
- Qualified: `refuseStaleBucketUsage`
- AST evidence: `ast.json` (`source_sha256` 8eacbf2fea3f1234…) — **편집 전**(freeze census 와 sha · 분기 일치)
- Risk scan: `risk-pattern-report.md`
- AST branches 12 · return 11

**역할.** 공유 bucket 사용량 대조의 유일한 판정 규칙 — 원장 사용량이 snapshot 주장보다 크면 stale, latch 면 차단, 기록된 가장 작은 한도로 cap.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `buckets` · `caps` | 같은 길이 · 같은 키 | admission 계획 | B1 · B8 거절 |
| `ReadJournalBucketUsage` 결과 | 원장 사용량 | reader | B3 판독 오류 → 거절 |
| `smallestRecordedBucketLimit` 결과 | 기록 한도 | 원장 | B9 오류 · B10 없음 → continue |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:31` caps · buckets 길이 불일치 | — | `:32` | 기존 |
| B2 | `:34` bucket 순회 | — | — | 기존 |
| B3 | `:36` 사용량 판독 오류 | — | `:37` `ErrRiskBucketSnapshotMismatch` | a126 D2 거절 전파(1.1 손상 시험) |
| B4 | `:43` latch 된 사용량 | — | `:44` | a126 M7(떠난 행 플래그) |
| B5 | `:47` ledger 합 금액 아님 | — | `:48` | 기존 |
| B6 | `:51` 주장 합 금액 아님 | — | `:52` | 기존 |
| B7 | `:54` 주장 < ledger | — | `:55` stale | 기존 |
| B8 | `:63` cap 키 불일치 | — | `:64` | 기존 |
| B9 | `:67` 기록 한도 판독 오류 | — | `:68` | 기존 |
| B10 | `:70` 기록 한도 없음 | — | `continue` | 기존 |
| B11 | `:74` 합 금액 아님 | — | `:75` | 기존 |
| B12 | `:77` ledger + 예약 > 기록 한도 | — | `:78` cap 소진 | a126 M4(떠난 행 한도 유지) |

## Calls and live bindings

`riskbucket.ReadJournalBucketUsage`(원장 읽기) · `latchedUsageRefusal` · `sumMinor` · `smallestRecordedBucketLimit`(원장 읽기). 브로커 호출 없음.
오류는 거절(`RefusalError` · `ErrRiskBucketSnapshotMismatch`)로 되던진다.

## State mutations and fallbacks

없다 — 판정만.

## Safety conclusion

- **Safe edit boundary (a126 D3)**: **주석 한 줄만** — `:59–62` 의 "활성은 원장이 세는 행과 같은 모집단" 을 "한도 모집단은 떠난 행을 포함한다(a126 D3)" 로.
  분기 · 호출 · 모집단 SQL(`smallestRecordedBucketLimit`) 무변.
- **High-risk impact**: yes(진입 cap 판정) — 동작 무변.
