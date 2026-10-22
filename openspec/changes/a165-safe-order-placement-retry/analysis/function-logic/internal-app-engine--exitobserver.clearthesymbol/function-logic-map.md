# Function Logic Map: `ExitObserver.clearTheSymbol`

- Source: `internal/app/engine/exitloop.go` (1629-1723)
- Qualified function: `ExitObserver.clearTheSymbol`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` 4d92acb6e7e17a25…)
- AST evidence: `ast.json` — AST branches 17, 반환 8, 호출 25
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official·trading·flatten 패키지 전체, journal `-run 'Classify|SendTracker|Dispatch'`, app/engine `-run 'Clear|Cancel'` — journal·engine 은 부분 실행이라 count=0 이 「미실행」을 뜻하지 않을 수 있음, cmd/tossctl 은 미측정) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

엔진 청산 전 미체결 정리. B11(1697) 취소가 오류이거나 CONFIRMED 가 아니면 `cleared=false` 로 다음 주문으로 continue — 같은 주기 안 대기 없음. B3(1639) 미정착 mutation 이 있으면 그 종목은 이번 주기 치우지 않음. 409 를 FAILED_CONFIRMED 로 분류하면 그 attempt 는 미정착이 아니므로 다음 주기에 다시 취소된다(사용자 결정 1 D).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| m | managed | ObserveOnce | — |
| withPending | bool | 호출자 | — |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 1633:3, 1637:3, 1642:3, 1707:3, 1712:4, 1716:4, 1719:4, 1722:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 1632:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 1632 이하 | 미실행 — 패키지 스위트 count=0 |
| B2 | `if at 1636:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 1636 이하 | 미실행 — 패키지 스위트 count=0 |
| B3 | `if at 1639:2` | if len(unsettled) > 0 { | 아래 「State mutations」 | 소스 줄 1639 이하 | 실행됨 — 패키지 스위트 count=1 |
| B4 | `range at 1645:2` | for _, order := range live { | 아래 「State mutations」 | 소스 줄 1645 이하 | 실행됨 — 패키지 스위트 count=1 |
| B5 | `if at 1647:3` | if !buy { | 아래 「State mutations」 | 소스 줄 1647 이하 | 실행됨 — 패키지 스위트 count=1 |
| B6 | `if at 1651:4` | if err != nil { | 아래 「State mutations」 | 소스 줄 1651 이하 | 미실행 — 패키지 스위트 count=0 |
| B7 | `if at 1656:4` | if waiting { | 아래 「State mutations」 | 소스 줄 1656 이하 | 실행됨 — 패키지 스위트 count=1 |
| B8 | `if at 1660:4` | if !withPending { | 아래 「State mutations」 | 소스 줄 1660 이하 | 실행됨 — 패키지 스위트 count=1 |
| B9 | `if at 1676:3` | if err != nil { | 아래 「State mutations」 | 소스 줄 1676 이하 | 미실행 — 패키지 스위트 count=0 |
| B10 | `if at 1682:3` | if qerr != nil \|\| perr != nil { | 아래 「State mutations」 | 소스 줄 1682 이하 | 미실행 — 패키지 스위트 count=0 |
| B11 | `if at 1697:3` | if err != nil \|\| out.State != journal.StateConfirmed { | 아래 「State mutations」 | 소스 줄 1697 이하 | 실행됨 — 패키지 스위트 count=1 |
| B12 | `if at 1701:3` | if !buy { | 아래 「State mutations」 | 소스 줄 1701 이하 | 실행됨 — 패키지 스위트 count=1 |
| B13 | `if at 1706:2` | if !res.cleared { | 아래 「State mutations」 | 소스 줄 1706 이하 | 실행됨 — 패키지 스위트 count=1 |
| B14 | `if at 1709:2` | if withPending && m.state.Pending() { | 아래 「State mutations」 | 소스 줄 1709 이하 | 실행됨 — 패키지 스위트 count=1 |
| B15 | `if at 1710:3` | if strings.TrimSpace(m.state.PendingIntentID) == "" { | 아래 「State mutations」 | 소스 줄 1710 이하 | 미실행 — 패키지 스위트 count=0 |
| B16 | `if at 1715:3` | if err != nil { | 아래 「State mutations」 | 소스 줄 1715 이하 | 미실행 — 패키지 스위트 count=0 |
| B17 | `if at 1718:3` | if !released { | 아래 「State mutations」 | 소스 줄 1718 이하 | 실행됨 — 패키지 스위트 count=1 |

## Calls and live bindings

`LiveOrdersForSymbol`, `UnsettledOnSymbol`, `ConfirmedCancelOf`, `IssueReduction`, `Submit.Cancel`.

AST 호출 이름 전체: `clearTargets`, `costs.Market`, `floatOf`, `fmt.Errorf`, `fmt.Sprintf`, `len`, `m.state.Pending`, `o.opts.Issuer.IssueReduction`, `o.opts.Journal.ConfirmedCancelOf`, `o.opts.Journal.LiveOrdersForSymbol`, `o.opts.Journal.ReleaseClearedExitProposal`, `o.opts.Submit.Cancel`, `o.opts.Submit.UnsettledOnSymbol`, `o.warnEpisode`, `onlyEngineCancelsInFlight`, `strings.EqualFold`, `strings.ToLower`, `strings.ToUpper`, `strings.TrimSpace`, `workingOrderPrice`.

## State mutations and fallbacks

- 발의 해제(release).

## Safety conclusion

- 편집 계획 없음.
