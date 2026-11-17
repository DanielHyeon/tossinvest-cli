# Function Logic Map: `Tracker.Observe` (reconcile)

- Source: `internal/reconcile/mismatch.go` (629-856)
- 편집 전: `../../pre-edit-ast/reconcile--tracker.observe.ast.json`(618~835, sha f72bcbc7d288, 분기 28).
- AST evidence: `ast.json`(편집 뒤 sha 9895f31184e5, 분기 29, return: return@660 return@731 return@855). 명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/reconcile/mismatch.go --func Tracker.Observe`.
- Risk scan: `risk-pattern-report.md`(일치 0).
- 재번호: `renumber.txt`(`../renumber.py` — 옛/새 분기를 (종류, 그 줄의 소스 문장)으로 정렬. 손 재번호 안 함).

## 재번호

```
B1:if@624 -> B1:if@635
B2:if@637 -> B2:if@650
B3:else@684 -> B3:else@703
B4:if@643 -> B4:if@656
B5:if@646 -> B5:if@659
B6:range@653 -> B6:range@666
B7:if@654 -> B7:if@667
B8:if@662 -> B8:if@675
(new/changed, insert) -> B9:if@682  | if t.orderTerminalCandidate(ctx, &block, diff, disputed) {
B9:if@670 -> B10:if@689
B10:if@689 -> B11:if@708
B11:range@702 -> B12:range@721
B12:if@703 -> B13:if@722
B13:if@711 -> B14:if@730
B14:if@715 -> B15:if@734
B15:if@718 -> B16:if@737
B16:range@719 -> B17:range@738
B17:if@720 -> B18:if@739
B18:if@727 -> B19:if@746
B19:if@738 -> B20:if@757
B20:range@768 -> B21:range@787
B21:if@771 -> B22:if@790
B22:range@776 -> B23:range@795
B23:if@778 -> B24:if@797
B24:range@783 -> B25:range@802
B25:range@802 -> B26:range@823
B26:range@805 -> B27:range@826
B27:if@806 -> B28:if@827
B28:if@820 -> B29:if@841
```

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `diff` | 한 비교 | `ReconcileDriver.RunOnce` | as-of 없음/못 읽음 → 크레딧·종결 증명 둘 다 쓰지 않음 |
| `t.Terminal` | nil 또는 `OrderTerminalJournal` | 운영: `buildGateway` 의 `*journal.Journal` | nil → 새 호출 셋이 즉시 반환(오늘과 같음) |
| `t.terminal` | 결속 기억(프로세스 로컬) | `order_terminal_release.go` | 영값 = 결속 없음 |
| 불변식 | 차단을 거는 조건·시점·순서 불변, 크레딧 규칙 먼저 | design D4·D8 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@635) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B2 (if@650) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B3 (else@703) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B4 (if@656) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B5 (if@659) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B6 (range@666) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B7 (if@667) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B8 (if@675) | 크레딧이 쓸 수 없음(`!creditUsableBy`) — 편집: 안에 B9 를 더함 | B9 거짓이면 `AwaitingAdjustment` append(기존) | continue | `TestACoincidentalAgreementStillDoesNotRelease`, `TestA192TheAdjustmentRulesHoldWithTheReaderOn` |
| B9 (if@682) | 크레딧 없는 블록에서 `t.orderTerminalCandidate` 가 참 — 이 프로세스가 삽입한 missing-order 블록의 1차 판정이 「해제」 | 블록에 `terminalRelease` 표시, `out.Cleared` append(무장이면 결속 무장만 갱신하고 거짓) | continue | `TestA192MissingOrderBlockArmsThenReleasesAfterTheTerminalProof`, `TestA192WithoutATerminalReaderTheTrackerBehavesAsToday` |
| B10 (if@689) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B11 (if@708) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B12 (range@721) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B13 (if@722) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B14 (if@730) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B15 (if@734) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B16 (if@737) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B17 (range@738) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B18 (if@739) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B19 (if@746) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B20 (if@757) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B21 (range@787) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B22 (if@790) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B23 (range@795) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B24 (if@797) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B25 (range@802) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B26 (range@823) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B27 (range@826) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B28 (if@827) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B29 (if@841) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `t.beginOrderTerminalObservation`(call@647) | 갈래 나누기 전 결속 갱신·관측 시작 블록 키 기록 | 오류 없음(기억만) | AST |
| `t.orderTerminalCandidate`(call@682) | 크레딧 없는 블록의 1차 판정 | 읽기 실패는 증명의 ReadErr → 유지 | AST |
| `t.persist`(call@786) | 영속 — 결속 생성·종결 증명 해제가 그 안 | 오류 → persistErr(기존) | AST |
| `t.updateOrderTerminalBindings`(call@808) | persist 뒤 같은 비교의 다른 신원·흔적 더함, 사라진 블록의 결속 버림 | 오류 없음 | AST |

## State mutations and fallbacks

- 새 상태: `t.terminal`(결속·무장·seen·사유). 블록 집합·streak·크레딧 갱신은 기존 그대로.
- `Terminal == nil` 이면 새 호출 셋 모두 첫 줄에서 반환 → 오늘과 같은 행동(`TestA192WithoutATerminalReaderTheTrackerBehavesAsToday`).

## Safety conclusion

- Safe edit boundary: 호출 3줄 + B8 안 분기 1(B9). 기존 분기 28개는 재번호 표대로(B9 뒤 +1).
- High-risk impact: yes(reconciliation). 차단을 거는 갈래(B3)의 블록 집합·추가 순서 무변경, 크레딧 경로 무변경, 해제는 persist 의 tx 재판정을 통과해야 커밋.
