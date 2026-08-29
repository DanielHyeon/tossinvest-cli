# Branch Test Map: `cleanupFrom`

- Source: `internal/verifylive/cleanup.go`
- Function: `internal/verifylive/cleanup.go:cleanupFrom`

이 파일이 이 change의 **RED 원문 보관처**다. 다른 12개 target의 branch-test-map은 여기를
가리킨다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range` line 125: `for _, l := range outstandingLines(entries) {` — 취소되지 않은 객체를 그 줄의 index와 함께 훑는다 | `TestALegacyRecordIsJudgedExactlyAsBefore`(5종) | no — 순회 자체는 무변경 | yes |
| B2 | `if` line 127: `if gate == "" {` — 기다릴 대상이 없는 객체는 언제나 대상 | `TestALegacyRecordIsJudgedExactlyAsBefore/주문은 gate가 없다`, `TestALeftoverOrderIsNotSubjectToTheOrderingRule`, `TestALeftoverOrderIsCancelledOnTheNextRun` | no — 이 경로가 기존 주문 동작이며 보존이 목표다 | yes |
| B3 | `if` line 131: `if settled(gate) && heldAfter(entries, gate, l.at) {` — gate가 이 줄 **뒤에** 판정했을 때만 대상 | `TestAHeldOrderIsNotACleanupTarget`, `TestAHeldConditionalIsNotACleanupTarget`, `TestAHoldEndsWhenItsGateDecides`, `TestAGateVerdictOlderThanTheHoldDoesNotRelease`, `TestAFailedCancelAfterTheHoldStillReleases`, `TestAReDeclaredHoldOutlivesAnOlderVerdict`, `TestAVerdictOlderThanTheConditionalDoesNotCleanItUp`, `TestAVerdictNewerThanTheConditionalOpensCleanup`, `TestAConditionalWithNoCancelVerdictIsNotCleanedUp` | **yes** — 아래 원문 | yes |

## RED 원문

### RED 1 — 필드가 없어 컴파일 실패

```text
$ go test ./internal/verifylive/ -run 'Hold|Held|Chain|Clock|Legacy|Gate' -count=1
internal/verifylive/hold_test.go:33:21: unknown field HeldUntil in struct literal of type Artifact
internal/verifylive/hold_test.go:33:38: unknown field ChainID in struct literal of type Artifact
FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/verifylive [build failed]
```

### RED 2 — 필드를 더한 뒤, 실제 결함이 드러난다

```text
--- FAIL: TestAHeldOrderIsNotACleanupTarget (0.00s)
    hold_test.go:53: the prologue would cancel the child order the trigger measurement has to
    watch fill; the step that placed it has not reached a terminal verdict:
    [{Kind:order ID:child-1 Symbol:333430 ... Deliberate:true HeldUntil:conditional-trigger
      ChainID:chain-A Note:}]

--- FAIL: TestAGateVerdictOlderThanTheHoldDoesNotRelease (0.00s)
    hold_test.go:102: the gate's verdict predates the hold, so it decided nothing about this object

--- FAIL: TestAReDeclaredHoldOutlivesAnOlderVerdict (0.00s)
    hold_test.go:145: the newest line holds this object until conditional-cancel decides again,
    and the only verdict on record predates that line

--- FAIL: TestAModifiedConditionalKeepsTheChainAndTheHold (0.05s)
    hold_test.go:318: the conditional the modify created carries no chain, so the record cannot
    say it continues the one that was replaced
    hold_test.go:322: the replacement arrived unheld
    hold_test.go:338: the registered conditional carries no chain

--- FAIL: TestTheCleanupDecisionDoesNotReadAClock (0.00s)
    hold_test.go:392: heldAfter was not found in cleanup.go — this guard is asserting nothing
```

**RED 2에서 이미 통과한 것들**이 이 change의 보존 주장이다 — 기존 규칙이 이미 옳던 자리에서는
새 테스트도 처음부터 통과했다: `TestAHeldConditionalIsNotACleanupTarget`,
`TestAHoldEndsWhenItsGateDecides`, `TestAFailedCancelAfterTheHoldStillReleases`,
`TestALegacyRecordIsJudgedExactlyAsBefore`(5종), `TestAHeldArtifactSurvivesTheRecord`,
`TestAnUnheldArtifactWritesNoHoldFields`.

### GREEN

```text
$ go test ./internal/verifylive/ -count=1
ok  	github.com/JungHoonGhae/tossinvest-cli/internal/verifylive	10.927s

$ go test ./... -count=1     → 3766 passed, 0 failed  (기준선 d36da6d: 3749)
$ go vet ./...               → 이슈 없음
```

### 정적 가드 변이 검증

`TestEveryHoldGateIsACatalogueStep`이 실제로 검사하는지 확인하려고 호출부 하나의 gate를
카탈로그 밖 상수로 바꿨다.

```text
--- FAIL: TestEveryHoldGateIsACatalogueStep (0.01s)
    hold_test.go:450: steps.go: the hold gate is CleanupLabel ("이전 실행이 남긴 객체 정리"),
    which is not a step in the catalogue — nothing will ever settle it and the object it holds
    is never released
```

되돌린 뒤 통과. 이 가드는 review.md F2의 결함(검사 대상 0건인데 통과하던 상태)을 고친 뒤의
것이다.

### 게이트가 잡은 것 — FLM 행 번호 표류

첫 게이트 실행이 `GATE FAIL — Function Logic Map 산출물 미완료`로 떨어졌다. 원인은 구현이
아니라 **증거의 신선도**였다: FLM 생성 뒤에 `cleanupTargets`의 문서 주석을 고쳐 같은 파일의
함수들이 2줄씩 밀렸고, 생성기는 `check_analysis.py`의 missing 목록으로 구동되므로 "이미
완료"인 상태에서는 아무것도 다시 만들지 않았다. cleanup.go의 세 함수를 강제 재생성해 해소했다.

### 실계좌 기록 재생 — 판정 무변화

```text
capability-verify.jsonl     60 entries, outstanding=0, pendingCleanup=0
capability-verify-us.jsonl  34 entries, outstanding=0, pendingCleanup=0
```
