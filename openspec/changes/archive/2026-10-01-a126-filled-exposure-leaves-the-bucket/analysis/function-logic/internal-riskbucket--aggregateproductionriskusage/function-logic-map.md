# Function Logic Map: `aggregateProductionRiskUsage`

- Source: `internal/riskbucket/production_snapshot_authority.go` (`494`–`534`)
- Qualified: `aggregateProductionRiskUsage`
- AST evidence: `ast.json` (`source_sha256` c5e0c64b2de65248…) — **편집 뒤**(구현 로트, 격리 워크트리). 편집 전 판은 `467322df`
- Risk scan: `risk-pattern-report.md`
- AST branches 6 · return 5 · 호출 31

**역할.** 예약 행을 검증하고 합을 내고 latch 를 모으며 RowDigest 를 만든다. a126 뒤에는 떠남 규칙(D1 · D2)의 **유일한** 적용 자리다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `rows` | `readProductionRiskUsage` 결과(사실 열 포함) | 원장 | 행 계약 위반 B2 · 사본 불일치 B3 · 영수증 행 손상/불일치 B4 → `ErrJournalUsageInvalid` |

## Branches and early returns

> 분기 표는 `analysis/harness/branch_table.py` 가 ast · 소스 · 편집 뒤 커버리지(`analysis/impl/coverage-post-edit.out`, `-coverpkg=./internal/riskbucket,./internal/journal`)로 만들었다. 「창의 return」은 위치다.

| Branch | 종류 | 조건 (원문) | 창의 return | 진입 실측 |
|---|---|---|---|---|
| B1 | range | `:498` `for _, row := range rows {` | — | 예 |
| B2 | if | `:501` `if !filledOK \|\| !heldOK \|\| rowFilled.Sign() < 0 \|\| rowHeld.Sign() < 0 \|\| rowFilled.BitLen() > 256 \|\| rowHeld.BitLen() > 256 \|\|` | :505 | — |
| B3 | if | `:510` `if (row.Receipted != 0 \|\| row.DecisionReceipted != 0) && row.OwnerKeyMatches == 0 {` | :511 | 예 |
| B4 | if | `:513` `if row.Receipted != 0 && (row.OwnerReleasedAt == "" \|\| row.OwnerReleasedAt != row.ReceiptReleasedAt \|\| row.State == "HELD" \|\| rowHe…` | :514 | 예 |
| B5 | if | `:522` `if !departed {` | — | 예 |
| B6 | if | `:526` `if filled.BitLen() > 256 \|\| held.BitLen() > 256 {` | :527, :532 | 예 |

- B3 — 영수증이 (r.* 또는 d.* 로) 있는데 예약 행의 owner 키가 결정 사본과 다르면 거절(D1 조건 4).
- B4 — r.* 영수증이 있는데 owner released_at 이 없거나 다르거나(D1 조건 2), 행이 HELD 이거나 held≠0 이면(D2) 거절. **scope latch 판정보다 앞** — 되돌려진 행도 손상 검사를 받음(codex #4).
- B5 — `departed = 영수증 ∧ scope latch 없음`(D1 조건 1 · 3, D4 되돌림). 떠난 행은 합에서만 빠지고 latch 플래그는 그 앞에서 이미 집계됨(Q3).

## Calls and live bindings

`big.Int.SetString` · `canonicalIdentity` · `fmt.Errorf`(`%w ErrJournalUsageInvalid`) · `productionRiskDigest`. 원장 · 브로커 호출 없음(순수). 결과 `JournalBucketUsage`.

## State mutations and fallbacks

없다 — 순수. RowDigest parts 형식은 편집 전과 같음(D2 — 떠남은 snapshot digest 의 filled 로 반영).

## Safety conclusion

- **Safe edit boundary**: 감소는 영수증에만 걸리고 모르는 것(불일치 · 손상)은 거절, scope latch 는 되돌림. 한도 모집단은 이 함수 밖(D3, 무변).
- **High-risk impact**: yes — 감소는 진입을 여는 방향이다. 변이 M1 · M1b · M2 · M3 · M6 · M6b-1~3 · M6c · M7 전부 CAUGHT(`analysis/impl/mutation-1.log`).
