# Function Logic Map: `stageKey`

> **편집 후 지도 — 2단계 GREEN 뒤(2026-11-23).** 편집 전 원본 `../../pre-edit-ast/internal-app-engine--stagekey.ast.json`, 재번호 `renumber.txt`(옛/새 분기를 (종류, 줄 문장)으로 difflib 정렬 — 생성기 `../../renumber.py.txt`).

- Source: `internal/app/engine/attestation_watch.go` (340-353)
- AST evidence: `ast.json` — branches **4** / returns **4** / calls **9**, source SHA-256 `292dab380cb7273b0a42307786fb72e3511e08859e38c5e8dae785cdd4932bf3`(파일 단위, 편집 후).
  명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/app/engine/attestation_watch.go --func stageKey`.
- Risk scan: `risk-pattern-report.md`.
- a166 census: `attestation_watch.go`·`a135_attestation_watch_internal_test.go` 0건(1.4).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `v.Stage`·`v.DaysLeft`·`v.Deadline`·`v.Missing` | judge 결과(Missing 은 MissingEndpoints 가 정렬) | `judgeAttestation` | — |
| `now`·`kst` | 감시 시계, 고정 +9h | — | — |
| 불변식 | 키 = outbox Key 본체(eventFor 가 접두어만 붙임) — 발급 신원 없음, 계좌·경로 없음 | design D5 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (switch@343) | 단계 | — | — | — |
| B2 (case@344) | notice·imminent | 없음 | return@345 `stage:days:기한` | `TestA196SameProofReissueIsNotAnnouncedAgain`(D-7→D-6 Key 다름) · `TestA135WatchAnnouncesEachStageOnce` |
| B3 (case@346) | proof_overdue | 없음 | return@347 `stage:기한:KST날짜` | `TestA196OverdueIsCriticalOncePerDay` |
| B4 (case@348) | scope_lost | 없음 | return@350 `stage:누락,집합:KST날짜` | `TestA196RestartedWatchReusesTheRejectionKey` · `TestA196ScopeLossOncePerDayAcrossReissues` |
| (끝) | 그 밖 거절 단계 | 없음 | return@352 `stage:KST날짜` | `TestA196RejectionDayIsTheKSTDate` · `TestA135InvalidStatesAreRemindedDaily` · `TestA196ScopeLossThenExpiryAreTwoAlerts` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `now.In(kst).Format`·`v.Deadline.UTC().Format`·`fmt.Sprintf`·`strings.Join` | 문자열 조립 | 순수 | AST calls 9 |

## State mutations and fallbacks

- 없음(순수). 편집 전 `if` 하나 → `switch` 셋 갈래(renumber: changed — 구조 교체).

## Safety conclusion

- Safe edit boundary: 함수 전체(순수 키 조립).
- High-risk impact: no — 단 키가 outbox 중복 판정 입력이라 변이 ④⑤⑩⑪ 로 고정.
