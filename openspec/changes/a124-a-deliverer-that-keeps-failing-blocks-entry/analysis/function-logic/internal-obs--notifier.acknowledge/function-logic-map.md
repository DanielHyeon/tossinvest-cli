# Function Logic Map: `Notifier.Acknowledge`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json` — **편집 전**, :840–878, 분기 10 · 반환 6 · 호출 12.
  source_sha256 `0bc75668ff17…`, 추출 base `4798d399` (2026-09-26, Teammate — 재freeze 로트).
- Risk scan: `risk-pattern-report.md`

**a124 4판은 이 함수에 한 줄을 넣는다** — `n.mu.Lock` (:851) 직후 승인 세대 `ackGen` 증가(design D7). 분기·해제 조건(B4~B10)은 그대로다.
이 번들은 design 2판 D7 이 「승인은 `n.mu` 아래에서 세고-푼다」를 근거로 쓰기 때문에 문서보다 먼저 만들었고, 편집 뒤 `revision: current`
로 재추출한다(tasks 1.3).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `operator` | 비어 있지 않음 | 운영자 명령 | B1 → 오류 |
| `ids` | 비면 PENDING 전부 | 운영자 | — |
| `n.Journal` | nil 허용 | 배선 | B2 → 게이트만 해제 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 운영자 공백 (:841) | 없음 | 오류 (:842) | obs 승인 시험 |
| B2 | `n.Journal == nil` (:844) | B3 참이면 `Gate.Clear` :846 | `nil` (:848) | 1.4 |
| B3 | `n.Gate != nil` (:845) | 해제 | — | 1.4 |
| — | (B2 거짓) `n.mu.Lock` :851 · `defer Unlock` :852 | **이후 전부 `n.mu` 아래** | — | — |
| B4 | `len(ids) == 0` (:854) | `PendingAlerts(ctx, 0)` :855 | — | `TestPersistentDeliveryFailureBlocksEntries` 계열 |
| B5 | 나열 오류 (:856) | 없음 | 오류 (:857) | 1.4 |
| B6 | 나열 행 순회 (:859) | ids 적재 | — | — |
| B7 | ids 순회 (:863) | `AcknowledgeAlert` :864 | — | — |
| B8 | 승인 오류이고 `ErrAlertNotFound` 아님 (:864) | 앞선 승인은 커밋됨 | 오류 (:866) | 1.4 |
| B9 | `UndeliveredCount` 오류 (:871) | 없음 | 오류 (:872) | 1.4 |
| B10 | `remaining == 0 && n.Gate != nil` (:874) | **`Gate.Clear(ReasonAlertUndelivered)`** :875 | `nil` (:877) | `TestAcknowledgeWhileStillPendingKeepsTheBlock` · `TestAcknowledgeCannotClearTheGateMidSend` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.mu.Lock` :851 | 동기 발송자와의 배제 — 세고-푸는 구간 | — | AST |
| `n.Journal.AcknowledgeAlert` :864 | 행 → ACKNOWLEDGED, 임차 해제 (outbox.go:481-504) | NotFound 는 무시 | AST |
| `n.Journal.UndeliveredCount` :870 | 해제 조건 | 오류 → B9 | AST |
| `n.Gate.Clear` :846 · :875 | 해제 | — | AST |

## State mutations and fallbacks

- 해제 조건은 「승인 + 미전달 0」이고 그 판단 전체(B4~B10)가 `n.mu` 아래다. 배달 실행자는 오늘 `n.mu` 를 잡지 않으므로
  이 배제 밖에 있다 — 실행자가 이 구간 사이에 `Block` 하면 빈 backlog 위의 래치가 된다(review F2).
- `ids` 가 비고 PENDING 도 없으면 B10 이 바로 해제한다 — 늦은 래치를 운영자가 다시 풀 수 있는 이유(보수 방향의 잔여).

## Safety conclusion

- Safe edit boundary: `n.mu.Lock`/`defer Unlock` 뒤에 원자 증가 한 줄. B1~B10 의 조건·순서·반환은 불변. 실행자는 같은 `n.mu` 를 메모리 전용 울타리
  (`LatchUnlessAcknowledgedSince`, 새 leaf)로만 빌린다.
- High-risk impact: yes — 진입 게이트 해제 경로.
