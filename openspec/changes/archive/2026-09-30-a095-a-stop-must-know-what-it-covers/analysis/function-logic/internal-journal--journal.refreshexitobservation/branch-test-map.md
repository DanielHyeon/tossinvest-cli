# Branch Test Map: `Journal.RefreshExitObservation`

- Source: `internal/journal/exit_observation_refresh.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다. 따라서 「Test」 열은 **a095 3판이 요구하는 시험**이며 현존 증명이 아니다. `[비움 — Qn]`은 사용자 결정이 덮지 않아 Manager 에게 올린 질문이다.

| Branch | 조건 | 진입 실측 | Test (a095 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:41` `if id == "" {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B2 | `:44` `if err := request.Provenance.validate(); err != nil \|\| request.Provenance.zero() {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B3 | `:45` `if err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B4 | `:50` `if !sameExitDecisionProvenance(request.Provenance, ExitDecisionProvenance{` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B5 | `:56` `if request.ObservedAt.IsZero() {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B6 | `:60` `if _, _, ok := observationSourceOrder(source); !ok {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B7 | `:63` `if request.Snapshot.Orderable \|\| !request.Snapshot.ExecutableProposal().Zero() {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B8 | `:76` `if err := validateJudgementSnapshot(id, judgement, candidate); err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B9 | `:80` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B10 | `:85` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B11 | `:88` `if current.Completed {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B12 | `:91` `if current.Effective == nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B13 | `:95` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B14 | `:98` `if status != SnapshotStatusEvaluated \|\| proposalPending {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B15 | `:102` `if expectedLifecycle == 0 {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B16 | `:105` `if expectedLifecycle != current.LifecycleGeneration {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B17 | `:112` `if errors.Is(err, sql.ErrNoRows) {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B18 | `:114` `} else if err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B19 | `:114` `} else if err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B20 | `:117` `if lifecycleStatus != positionpolicy.StatusManaged \|\| lifecycleGeneration != expectedLifecycle {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B21 | `:121` `if err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B22 | `:124` `if decision == observationNoop {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B23 | `:127` `if current.PositionGeneration != request.Snapshot.PositionGeneration \|\|` | 예 | **a095 5.3** — issues I1에 「effective 스냅샷과 비교 — 스칼라와 일치할 때만 값 유지, 갈라지면 되돌림」으로 인용 | no | no |
| B24 | `:132` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B25 | `:148` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B26 | `:151` `if err := j.runExitWriteHook("after_refresh_state"); err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |

**미진입 분기 6개**: B9, B10, B13, B15, B24, B25
**자체 블록 없는 분기 0개**: 없음 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `switch {` 등)이며 미커버와 다르다.
