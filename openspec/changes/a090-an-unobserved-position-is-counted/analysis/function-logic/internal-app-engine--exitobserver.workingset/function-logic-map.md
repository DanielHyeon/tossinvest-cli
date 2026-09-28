# Function Logic Map: `ExitObserver.workingSet`

- Source: `internal/app/engine/exitloop.go` (`493`–`612`)
- Qualified: `ExitObserver.workingSet`
- AST evidence: `ast.json` (`source_sha256` 522d5d81c4992c57…) — branches 22 · returns 3
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **a090 2판 설계 전**(1라운드 적대 보이스 F4 — "B6·B7 이 유일한 자리" 는 거짓). 이 함수의 분기를 근거로 삼는 2판 문서보다 먼저 만들었다.

**역할.** 원장의 보유 포지션과 열린 exit state 를 맞춰, 판정할 포지션 목록(`states`)을 돌려준다. 이 목록에 **들지 못한 보유 포지션은
`ObserveOnce` 의 순회에 아예 나타나지 않는다** — 그래서 B6·B7 만 세면 이 함수의 탈락을 못 본다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `positions` | 계정의 포지션 전부 | `Journal.Positions` `:494` | 오류 → 함수 오류(B1) → `ObserveOnce` B2 |
| `stateResults` | 열린 exit state | `Journal.OpenExitStateResults` `:498` | 오류 → 함수 오류(B2) |
| 보유 판정 | `p.State != PositionClosed && !isZeroQuantity(p.Quantity)` | B5 `:509` | 아니면 건너뜀(보유 아님) |

## Branches and early returns

> 조건은 소스 원문. 「진입」 은 `analysis/harness/observeonce.blocks`(commit `eac13df1`) — 그 줄에서 시작하는 블록 count 가 1 이면 예.
> 「보유 포지션의 탈락」 열이 a090 의 관심이다: 보유(B5 통과) 포지션이 `states` 에 들지 못하고 끝나는 자리.

| Branch | 종류 | 조건 (원문) | 결과 | 보유 포지션의 탈락 | 진입 |
|---|---|---|---|---|---|
| B1 | if | `:495` `if err != nil {` | **return** `:496` | 전체(보유 집합 모름) | **아니오** |
| B2 | if | `:499` `if err != nil {` | **return** `:500` | 전체 | **아니오** |
| B3 | range | `:503` `for _, result := range stateResults {` | 색인 | — | 예 |
| B4 | range | `:508` `for _, p := range positions {` | 포지션마다 | — | 예 |
| B5 | if | `:509` `if p.State == journal.PositionClosed \|\| isZeroQuantity(p.Quantity) {` | `continue` `:510` | 보유 아님 — 탈락 아님 | 예 |
| B6 | if | `:512` `if !p.ExitEligible() {` | `alertUnmanaged`(`EventExitPositionUnmanaged`, 포지션당 래치) · `continue` `:519` | **예 — 자기 경보(normal 등급)** | 예 |
| B7 | if | `:525` `if !ok {` | `openState` | — | 예 |
| B8 | if | `:527` `if err != nil {` (openState) | `cycle.Err` · `continue` `:531` | **예 — 로그뿐(`reportCycle`, critical 아님)** | **아니오** |
| B9 | if | `:528` `if cycle.Err == nil {` | 첫 오류만 | — | 아니오 |
| B10 | if | `:533` `if opened.PositionID == "" {` | `continue` `:537` | **예 — 무음**(완료된 정책) | **아니오** |
| B11 | if | `:542` `if result.Corruption != nil {` | 격리 → `refused` → `judge` 에 닿음 | 아니오(refused 로 판정 경로에 감) | 예 |
| B12 | if | `:545` `if qerr != nil {` | `cycle.Err` · `continue` `:549` | **예 — 로그뿐** | **아니오** |
| B13 | if | `:546` `if cycle.Err == nil {` | | — | 아니오 |
| B14 | if | `:556` `if q, active, qerr := …ActiveExitSnapshotQuarantine(…); qerr != nil {` | `cycle.Err` · `continue` `:560` | **예 — 로그뿐** | **아니오** |
| B15 | else | `:561` `} else if active && !q.NeedsReJudgement() {` | (else-if 사슬) | — | 예 |
| B16 | if | `:557` `if cycle.Err == nil {` | | — | 아니오 |
| B17 | if | `:561` `} else if active && !q.NeedsReJudgement() {` | `refused` → `judge` | 아니오(refused) | 예 |
| B18 | else | `:565` `} else if active {` | | — | 예 |
| B19 | if | `:565` `} else if active {` | 재판정 한 번(로그) | 아니오 | 예 |
| B20 | if | `:589` `if identityErr != nil {` | 격리 → `refused` | 아니오 | 예 |
| B21 | if | `:592` `if qerr != nil {` | `cycle.Err` · `continue` `:596` | **예 — 로그뿐** | **아니오** |
| B22 | if | `:593` `if cycle.Err == nil {` | | — | 아니오 |

마지막 **return** `:611`(`append(out, refused...)` — 격리 행은 맨 뒤, `:608-610`).

**보유 포지션이 순회 전에 탈락하는 자리는 다섯이며(B8 · B10 · B12 · B14 · B21) 전부 시험 진입 0 이다.** B8 은 진입 결정에 손절이 없으면
(`openState` `:667-671`) **매 주기** 같은 실패로 돈다. 또 보유 포지션이 전부 탈락하면 `states` 가 비어 `ObserveOnce` B3(`:431-439`)이
"Nothing is held" 로 **계정 두절 시계까지 리셋**한다 — 보유 중인데 무보유로 읽힌다.

## Calls and live bindings

`Journal.Positions` · `Journal.OpenExitStateResults` · `o.alertUnmanaged` · `o.openState` · `Journal.QuarantineExitSnapshot` ·
`o.announceQuarantine` · `Journal.ActiveExitSnapshotQuarantine` · `o.log` · `managedPolicyIdentity`. 브로커 호출 없음(원장·로컬).

## State mutations and fallbacks

원장: exit state 열기(`openState`), 격리 기록. 관측자: `unmanaged` 래치(`alertUnmanaged`), `quarantineAnnounced`.

## Safety conclusion

- **Safe edit boundary(a090 2판)**: B6 통과 뒤(`:520`, exit 대상인 보유 포지션) **한 자리에** "보유·대상 포지션으로 표시" 호출 하나를 둔다.
  다섯 탈락 자리는 편집하지 않는다 — 표시되고도 판정에 닿지 않은 포지션은 `ObserveOnce` 가 순회 뒤에 미관측으로 센다(design D1). 분기 조건·이탈 무변화.
- **High-risk impact**: yes — 손절 관측의 입력 집합.
