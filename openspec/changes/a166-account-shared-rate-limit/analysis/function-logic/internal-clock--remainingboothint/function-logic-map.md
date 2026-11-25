# Function Logic Map: `RemainingBootHint`

- Source: `internal/clock/bootstamp.go` L63–94
- AST evidence: `ast.json`(편집 뒤 current), 편집 전 AST `../../time-namespace-lot/pre-edit-ast/RemainingBootHint.json`
  (base `6ccc809d3`, source_sha256 `489eaa62…`, 분기 B1–B9 · return 9)
- Risk scan: `risk-pattern-report.md`
- 편집 뒤 AST(`ast.json`): 분기 id 동일(B1–B9), 문서 주석 1줄 추가로 줄 +1 — B2 L68 · B3 L71 · B4 L74 · B8 L87. 아래 표의 줄은 편집 전 기준.
- 변경: time namespace 로트(B1 수리, Manager 결정 2026-11-25 권장안 가). B2 · B3 · B4 · B8 조건에서 `Namespace !=` 항만 삭제.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `duration` | 0 ≤ d ≤ 30s | 429 응답 hint(저장값) | 범위 밖 → `ErrBootStamp`(B1) |
| `saved` | `validStamp`(UUID 형식 · 비영, ns device/inode ≠ 0, uptime ≥ 0) | hint 수신 시 `Snapshot` | 무효 → B1 |
| `now` | 같음 | 현재 프로세스 `Snapshot` | 무효 → B1 |
| `anchor` | nil 또는 유효 stamp | 다른 boot 첫 복원 시 저장 | 무효 → B1 |
| 불변식(편집 뒤) | 같은 boot 의 stamp 끼리는 같은 CLOCK_BOOTTIME — 각 stamp 를 만든 `Snapshot` 이 offset 0 · 현재=자식 ns · 앞뒤 일치를 확인했으므로 ns 신원과 무관(`time-namespace-lot/investigation.md` §5) | `Snapshot` L145 · L152 · L168 | 표본 단계에서 거절 |

## Branches and early returns

| Branch | Condition (편집 전 → 편집 뒤) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 L64 | duration 범위 · 세 stamp 유효성 (무변) | 없음 | `ErrBootStamp` | `TestBootRemainingInvalidStamps`, `TestBootCrossNamespaceStillRejectsInvalid` |
| B2 L67 | anchor≠nil ∧ anchor.boot=saved.boot ∧ (**ns 다름** ∨ anchor.up<saved.up) → **ns 항 삭제** | 없음 | `ErrBootStamp` | `TestBootRemainingRejectsAnchorBeforeSameBootProvenance`, `TestBootCrossNamespaceOrderingStillRejected`, `TestBootCrossNamespaceAnchorRestart` |
| B3 L70 | now.boot=saved.boot ∧ (**ns 다름** ∨ now.up<saved.up) → **ns 항 삭제** | 없음 | `ErrBootStamp` | `TestBootCrossNamespaceSameBootSubtractsElapsed`, `TestBootCrossNamespaceOrderingStillRejected`, `TestBootRemainingSameProvenanceStillRejectsNamespaceDriftAfterAnchor` |
| B4 L73 | anchor≠nil ∧ now.boot=anchor.boot ∧ (**ns 다름** ∨ now.up<anchor.up) → **ns 항 삭제** | 없음 | `ErrBootStamp` | `TestBootCrossNamespaceAnchorRestart`, `TestBootCrossNamespaceOrderingStillRejected` |
| B5 L76 | duration = 0 (무변) | 없음 | `{0,false}, nil` | `TestBootZeroDurationDoesNotHideInvalidClock` |
| B6 L80 | anchor≠nil → base=anchor (무변) | base 선택 | — | `TestBootRemainingConservative` |
| B7 L83 | now.boot≠base.boot (무변) | 없음 | `{duration, NeedsAnchor}` | `TestBootRemainingBoundariesAndRestarts`, `TestBootCrossNamespaceOtherBootNeedsAnchor` |
| B8 L86 | (**ns 다름** ∨ now.up<base.up) → **ns 항 삭제**. 도달 불가: B7 을 지난 now·base 는 같은 boot 이고, anchor=nil 이면 base=saved 라 B3 의 uptime 항이, anchor≠nil 이면 base=anchor 라 B4 의 uptime 항이 같은 입력을 먼저 거절 | 없음 | `ErrBootStamp` | 없음 — 변이 M6 SURVIVED(실측 `mutation-run-2.log`), 대신 막는 분기 B3 · B4. 역으로 B8 은 B4 의 uptime 항을 duration>0 에서 받쳐 1차 원장에서 M2 를 살렸다 — duration 0 판 시험으로 B4 를 직접 못 박음(Manager 결정 2: 분기 유지, 구조 시험 신설 안 함) |
| B9 L90 | elapsed ≥ duration (무변) | 없음 | `{0,false}` | `TestBootRemainingBoundariesAndRestarts`, `TestBootCrossNamespaceExpiredHintClears` |
| 끝 L93 | 남은 시간 | 없음 | `{duration-elapsed,false}` | `TestBootCrossNamespaceSameBootSubtractsElapsed` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `validStamp` ×3 (L64) | stamp 형식 검사 | bool, 실패 → B1 | AST calls L64 |

생산 호출자(무변): `pruneObservationSlots` L396, `ObservationStore.Finish` L531, `ObservationStore.PollRemaining` L598,
`validateObservationState` L230, `validObservationStamp` L165(`time-namespace-lot/census.txt`). 순수 함수라 잠금 · I/O 없음.
CodeGraph 1.6.0 `affected --filter '*_test.go'` 는 파일 단위 추이로 1,294 파일 — 선택 근거로 쓰지 않음(`time-namespace-lot/codegraph-affected-bootstamp.txt`).

## State mutations and fallbacks

- 없음(순수 함수). wall-clock fallback 없음. 다른 boot 는 기존대로 duration 전체 보류 + `NeedsAnchor`.

## Safety conclusion

- Safe edit boundary: 네 조건의 `Namespace !=` 항만. uptime 항 · boot 비교 · B1 · B5–B7 · B9 · `Snapshot` · `validStamp` · 저장 형식 무변.
- High-risk impact: yes(Manager 결정 1) — 엔진 폴링 hint 복원의 통과 여부가 바뀐다. 방향은 「거절 → 같은 시계로 계산」이고
  계산은 같은 namespace 경우와 같은 산술이다. 보수성 근거: 각 stamp 의 offset 0 이 `Snapshot` 에서 강제된다.
