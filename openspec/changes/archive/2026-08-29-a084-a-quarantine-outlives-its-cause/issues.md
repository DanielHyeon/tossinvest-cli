# a084 · Issues

## I1 — a078 I1의 설계 공백은 닫혔다

a078은 이렇게 남겼다.

> 즉 현재 코드에는 "격리를 풀고 원래 기준선을 유지하는" 경로가 없다. 이것 자체가
> 설계 공백이고, 후속 change 후보다.

a079가 사람 경로를, a084가 기계 경로를 만들었다. 저장된 진입가·최초 손절·보호선은
두 경로 모두에서 그대로 유지된다.

남은 조건: **개정이 같은 격리는 여전히 사람만 푼다.** 같은 선택기가 같은 얼어붙은
입력에 다른 답을 낼 수 없으므로 그것이 맞다.

## I2 — `ambiguous_recovery` evidence가 어느 분기에서 나왔는지 말하지 않는다

운영 원장의 세 건이 전부 접미사 없는
`exitpolicy: recovery candidate identity mismatch`다. `SelectRecoverySnapshot`에서 이
**맨** sentinel을 반환하는 경로는 두 개뿐이다 — 정체성 튜플 검사와
`compareRecoveryStage`의 rank 실패. 나머지 반환은 전부 `%w: <상세>`로 감싼다.

그 둘을 사후에 구별할 수 없다. a078은 high water가 첫 rung 문턱 바로 아래에 몰려 있다는
정황으로 특정했고, a084는 저장 snapshot을 재생해 확인했다. 둘 다 원장이 직접 말해줬어야
할 것을 추론으로 메운 것이다.

두 반환에 `%w: <어느 필드가 어긋났는가>`를 붙이면 다음 번엔 추론이 필요 없다. 격리
판정 자체는 바뀌지 않으므로 작은 change다.

## I3 — 개정 상수는 손으로 올린다

`exitpolicy.RecoverySelectorRevision`을 올릴 책임은 사람에게 있다. digest 자동 산출은
무관한 리팩터가 전 포지션을 재판정하게 만들므로 채택하지 않았다(design D3).

빠뜨린 경우의 방향은 안전하다(자동 재시도 없음 = 오늘의 동작). 다만 **선택기를 고치는
change의 tasks에 "개정 번호를 올렸는가"를 넣는 것**이 이 위험을 실질적으로 없애는
방법이고, 그것은 워크플로 문서의 몫이다.

## I4 — v27 마이그레이션 테스트가 §0.6과 충돌하고 있었다

`journalV25RowFingerprints`가 열 이름까지 해시해서, v26 이전 테이블에 nullable 열을
추가하는 것을 전부 거부했다. `schema.go`의 마이그레이션 규칙 2가 명시적으로 허용하고
`positions`·`mutation_attempts`가 이미 지난 경로인데도, v27~v29가 새 테이블만 만들어
드러나지 않았다.

fingerprint를 행만 해시하도록 좁히고 열 검사(`assertColumnsOnlyAppended`)를 분리했다.
**이 change가 없었다면 다음 additive 열 추가에서 같은 벽에 부딪혔을 것이다.**

## I5 — 독립 리뷰어 없이 배포 가능 상태에 도달했다

a083 I5와 같다. `codex`가 사용량 한도(2026-08-08 해제)로 거부됐다. 배포 전 별도 세션의
재검증이 남아 있다.

## I6 — 배포 전 `main` 대비 SchemaVersion 대조가 필요하다

`SchemaVersion` 29 → 30. 낮은 버전의 바이너리는 `ErrSchemaTooNew`로 거부하므로 롤백은
"이전 바이너리 실행"이 아니라 **DB 복원**이다. 마이그레이션 직전 자동 백업이 그 경로다.
배포 순서에서 콘솔·엔진이 같은 버전을 보게 해야 한다.

## I7 — 마이그레이션 하나가 journal 테스트 패키지를 Go 기본 timeout 밖으로 밀었다

`internal/journal`은 테스트마다 마이그레이션된 SQLite를 새로 열고 그런 테스트가 약 670개다.
그래서 **모든 새 마이그레이션 단계가 670번 곱해진다.**

```
v29 (a083 HEAD)   480.885s
v30 (a084)        795.642s      ← 둘 다 병렬 부하 아래 측정, 상대 비교용
```

v30에서 Go의 패키지당 기본 600초를 넘었고 `make test`가 timeout으로 실패했다. 매달린
것이 아니라 느린 것이다 — 넉넉한 timeout에서는 완주한다.

`make test`에 `-timeout 30m`을 붙였다. 정당한 마이그레이션이 걸리는 한계는 잘못된 실패를
보고하는 한계다.

**근본 원인은 남아 있다.** 다음 마이그레이션도 같은 곱셈을 낸다. 테스트가 스키마별
템플릿 DB를 복사해 쓰게 하면 마이그레이션 비용이 테스트 수와 무관해지지만, 그것은
별도 change다.

## I8 — `make gate`가 병행 세션의 진행 중 편집으로 막혀 있다

`make sdd-check`의 `check_agent_config_sync.py`가 drift를 보고한다.

```
 M .claude/CLAUDE.md
 M .codex/agents.md
```

두 파일 모두 **다른 세션이 지금 편집 중**이다. 한쪽에서 "에이전트 실행 순서" 절이
추가되고 advisory 관련 한 줄이 삭제된 상태이고, 두 파일의 SDD_SHARED 블록이 서로
어긋나 있다.

다른 세션의 진행 중 **안전 부트스트랩 편집을 덮어쓰지 않았다.** `--generate`는 그들의
작업을 지울 수 있다.

a084 자체의 증거는 전부 통과했다.

```
make test      EXIT=0 (go test -timeout 30m ./... 전 패키지 ok)
make vet       clean
make validate  77 passed, 0 failed
logic-map      evidence complete (11 함수)
openspec       valid --strict
```

**남은 조치**: 병행 세션이 두 파일을 커밋해 동기화한 뒤
`make gate CHANGE=a084-a-quarantine-outlives-its-cause`를 다시 돌린다.

## B1 — **격리(blocking)**: "개정당 1회" 경계가 실제로 존재하지 않는다

design D6이 기대는 경계는 코드에 없다. 격리 행에 개정을 다시 각인하는 두 writer
(`releaseReJudgedQuarantineTx`, `quarantineExitSnapshotTx`)는 **둘 다
`recordExitJudgementTx` 안**에 있다. 그런데 `judgeRatchet`/`judgeLadder`는
`o.record`에 **도달하기 전에** 반환한다.

```
exitloop.go:867   if !snapshot.Changed {   // ratchet
exitloop.go:934   if !snapshot.Changed {   // ladder
exitloop.go:1103  o.opts.Journal.RecordExitJudgementResult(...)   // 여기까지 못 온다
```

가격이 움직이지 않은 재판정은 격리 행을 건드리지 않는다. `NeedsReJudgement()`는 참으로
남고 **다음 주기에 전부 반복된다 — 영구히.**

재현: `selector_revision=NULL` 격리에 평탄한 호가로 10주기 →
`after 10 unchanged cycles: rows=1 active selector_revision=<nil> (current=2)`.

같은 성질이 `exit_state.go`의 조기 반환에도 있다 — `ErrProposalPending`(:462),
`ErrExitStateCompleted`(:425), `ErrExitLifecycleStale`(:432,:446) 전부 release 앞에서
반환한다.

결과 둘: (1) 상한이 없다. (2) `exitloop.go:556`이 5초마다 **"the generation is being
re-judged once"**라는 거짓을 기록한다 — 격리 포지션당 하루 약 17,000줄.

고칠 방향: 경계를 판정의 성패와 분리한다. `workingSet`이 통과시키기로 결정한 시점에
현재 개정을 각인하거나, 관측자 메모리에 시도한 개정을 포지션별로 들고 있는다.

## B2 — **§0.3 위반(blocking)**: 재판정 경로가 판정이 거부하기 **전에** 실주문을 취소한다

`exitloop.go:553`의 주석은 순서를 틀리게 적고 있다.

```
exitloop.go:1076   cleared, err := o.clearTheSymbol(ctx, m, snapshot.CancelPendingFirst)
exitloop.go:1103   o.opts.Journal.RecordExitJudgementResult(ctx, judgement)
```

`clearTheSymbol`은 브로커에 **실제 `Submit.Cancel`을 낸다**(:1279)。 그 다음에야
`RecordExitJudgementResult`가 돌고, 거기서 `SelectRecoverySnapshot`이 실패해 다시
격리할 수 있다(`exit_state.go:497`).

시나리오: 개정 1이 만든 `ambiguous_recovery` 격리가 통과 → 재계산이
`ActionLadderTakeProfit`(=`isFullExit`) 제안 → **working 익절 주문이 브로커에서
취소됨** → 판정 트랜잭션이 a078이 고치지 않은 사유로 `ErrRecoveryAmbiguous` → 재격리.

결과: 그 포지션은 **working 주문도 없고, 여전히 격리이며, 여전히 미판정**이다.
a084 이전에는 `record`에 도달조차 하지 않았다. **보호가 이전보다 나빠진다** — §0.3이
금지하는 방향이다.

가중: `managed`에 "이번은 재판정"이라는 표시가 없어 하위 경로가 첫 통과를 잠정적으로
다룰 수 없다.

## B3 — **롤백(blocking)**: `NeedsReJudgement`가 `!=`라서 개정이 낮아져도 재판정한다

```go
return q.SelectorRevision != exitpolicy.RecoverySelectorRevision
```

리뷰어 2명이 독립적으로 지적했다. 개정 2 바이너리가 개정 3이 각인한 행을 재판정하고,
거부 시 2로 다시 쓴다. **더 새로운 선택기가 거부한 판단을 더 오래된 선택기가 뒤집는다.**
그리고 이 배포는 프로세스가 여럿이다(`56e85c68` "the three processes"). 개정이 섞이면
`exit_snapshot_quarantines`에 주기당 한 행씩 무한 핑퐁이 생기고, 각 flip이
`quarantineAnnounced` latch를 다시 무장시켜 critical 알림이 매번 나간다.

고칠 방향: `q.SelectorRevision < exitpolicy.RecoverySelectorRevision`.
NULL은 여전히 0으로 읽히므로 backfill 없는 재판정 의도는 보존된다.

## B4 — **감사 위조(blocking)**: 선택기와 무관한 격리도 `SELECTOR_REVISED`로 닫힌다

리뷰어 3명이 독립적으로 지적했다. `NeedsReJudgement()`는 **복구 선택기에 관한 사실**인데
모든 격리 사유에 적용된다. `QuarantineExitSnapshot`은 `exitloop.go:527`
(`stored_snapshot_corrupt`)과 `:567`(`legacy_policy_identity_unknown`)에서도 불린다 —
두 경로 모두 복구 선택기를 **전혀 돌리지 않는다**.

그런데 v30 첫 주기에 그 행들이

> "the recovery selector changed since this quarantine was written; the generation was
> re-judged and a cause still holds"

라는 근거와 함께 `release_kind=SELECTOR_REVISED`로 닫힌다. **일어나지 않은 재판정을
원장에 기록한다.** `QuarantineReleaseSelectorRevised`의 doc comment가 "어느 근거가
이것을 닫았는지 기록한다"고 적은 바로 그 칸에 틀린 답을 쓴다.

재현(폐기한 테스트): `v1 reason=stored_snapshot_corrupt release_kind=SELECTOR_REVISED`.

고칠 방향: 재판정을 사유로도 게이트한다 — `ambiguous_recovery`만 대상.
같은 이유로 `ReleaseExitSnapshotQuarantine`의 validator 확장(:852)도 되돌린다.
프로덕션 호출자가 없는 dead 확장이고, 운영자 경로가 기계 전용 근거를 찍을 수 있게 만든다.

## I9 — `account_views.go`가 `selector_revision`을 읽지 않는다

`accountActiveQuarantines`(:242)는 그 열을 select하지 않으므로 모든 행이
`SelectorRevision==0`이 되고, `NeedsReJudgement()`는 그것을 "미상, 재판정 대상"으로 읽는다.
**읽지 않은 열과 각인되지 않은 열을 구별할 수 없다.** 오늘 이 map에 대해
`NeedsReJudgement()`를 부르는 호출자는 없지만, 첫 호출자가 틀린 답을 받는다.
열을 select하거나 필드를 `sql.NullInt64`/`*int64`로 바꿔 sentinel을 위조 불가하게 만든다.

## I10 — CI가 `-timeout` 없이 `go test ./...`를 돌린다

I7에서 `make test`에만 `-timeout 30m`을 넣었다. `.github/workflows/ci.yml:28`은 여전히
`go test ./...`이므로 795초짜리 journal 패키지가 Go 기본 600초에 걸려 **이 브랜치의 CI는
매번 실패한다.** `Makefile`의 `cover` 타깃(:44)도 같다.
`run: make test`로 통일한다.

## I11 — journal 테스트가 느린 원인은 마이그레이션이 아니라 fixture다

측정: 열기당 약 182ms, 열기 지점 약 534곳. 모든 테스트가 새 on-disk SQLite를 만들고
30개 마이그레이션을 `synchronous=FULL`에서 **30개 커밋으로** 재생한다.
비용은 O(열기 × 마이그레이션)이므로 **앞으로 모든 마이그레이션이 무조건 3~5초를 더한다.**
timeout을 올린 것은 같은 대화를 v31로 미룬 것이다. TestMain에서 템플릿 DB를 한 번
마이그레이션하고 파일 복사로 여는 것이 실제 해법이다.

## I12 — 보류된 익절 제안은 다음 rung을 기다린다 (D9의 잔여 성질)

D9는 재판정 통과에서 익절 계열 제안의 주문 측을 보류한다. 보류된 제안은 재제안되지
않는다 — 다음 주기는 라인이 안 움직였으면 `!snapshot.Changed`로 반환하고, 움직였으면
그때의 rung을 평가한다. 즉 보류된 rung은 **다음 rung이 교차될 때까지** 실행되지 않는다.

손절에는 적용되지 않는다. `isProtective`가 `ActionBaselineBreach`·`ActionLadderStop`을
보류 대상에서 뺀다 — 보호를 미루는 것은 §0.3이 금지하는 지연이고,
`TestAReJudgementNeverWithholdsAStop`이 그것을 고정한다.

같은 성질이 기존 `ArmSuppressedWorkingOrder` 경로에도 있다(working 주문을 못 치우면
그 주기의 제안이 사라진다). D9는 그 계약을 재사용했을 뿐 새로 만들지 않았다. 근본
해법은 `record`가 판정 트랜잭션을 주문 측보다 **먼저** 돌리는 것이고, 그것은 High-risk
기존 함수의 구조 변경이므로 별도 change다.

## B5 — **닫힘**: 재시도가 가격 읽기보다 먼저 소진됐다

개정 2의 첫 D8 구현은 `workingSet`에서 각인했다. `workingSet`은 `o.observe` **앞에서**
돈다. 따라서 호가를 못 받은 주기 — 거래정지, 매매중단, 전송 실패 — 가 그 포지션의
유일한 재판정을 아무것도 판정하지 않은 채 태웠고, 이후 매 주기 거부되며 **손절이
평가되지 않는다.** 사람이 풀거나 개정 상수를 올리기 전까지.

각인을 `judge` 진입부로 옮겼다. 거기서는 호가가 손에 있다. 각인 실패는 격리를 유지하고
`NeedsReJudgement`를 참으로 두므로 다음 주기가 다시 시도한다 — fail-closed이면서 자가
복구다.

## B6 — **닫힘**: 엔진이 쓴 arm-suppression 사유를 원장 reader가 손상으로 거부했다

개정 2가 write-side validator만 넓히고 read-side 두 곳을 그대로 뒀다.
`ExitEvents`는 hard fail하고, `AccountExitEvents`는 행을 `invalid_event_evidence`로
격하한다. D9가 남기려던 근거를 잃고, 진짜 손상을 잡아야 할 validator가 정상 행에 발화한다.
`knownArmSuppression` 하나로 write·read 양쪽을 통일했다.

## I13 — `releaseReJudgedQuarantineTx`가 사유만 보고 재판정 여부를 추론한다

개정 2가 `NeedsReJudgement` 조건을 빼고 사유 게이트만 남겼다. 단일 프로세스에서는
`workingSet`이 유일한 진입이라 정합적이지만, 프로세스가 여럿이면 A의 판정 트랜잭션이
B가 방금 연 `ambiguous_recovery` 행을 `SELECTOR_REVISED`로 닫을 수 있다 — 일어나지 않은
재판정이다. 정본 해법은 `managed.reJudge`를 판정 트랜잭션까지 전달해 추론 대신 사실로
게이트하는 것이며, 별도 change다.

## I14 — 재판정 통과의 부분 익절은 실제로 무장한다

`workingSet` 주석이 "arms and cancels nothing"이라고 적고 있었으나 `LADDER_PARTIAL`·
`RATCHET_PARTIAL`은 `CancelPendingFirst`도 `isFullExit`도 아니어서 보류 분기에 들어가지
않고 정상 제출된다. 안전하다 — 제출은 판정 커밋 **뒤**이므로 거부가 주문을 남기지
않는다 — 지만 주석이 틀렸으므로 코드가 보장하는 것으로 고쳤다.

## B7 — **닫힘**: 재판정이 아닌 판정이 격리를 풀었다 (개정 3)

세 번째 독립 리뷰가 실행 재현으로 잡았다. `releaseReJudgedQuarantineTx`가 재판정
여부를 활성 행의 **사유**로 추론했다. 그래서 지금 도는 선택기가 방금 쓴
`ambiguous_recovery` 격리를, 그 다음 성공 판정이

> "the recovery selector changed since this quarantine was written; the
> re-judgement chose one verified candidate and this generation is judged again"

라는 근거와 함께 닫았다. **일어나지 않은 재판정이다.** 그리고 그 행은 a084 이전에
운영자만 풀 수 있던 종류다. 같은 커밋이 `ReleaseExitSnapshotQuarantine`에서
`SELECTOR_REVISED`를 막아 기계가 사람 해제를 위조하지 못하게 해 놓고, 판정 경로가
그것을 뒷문으로 하고 있었다.

각인은 판별자가 될 수 없다 — `judge`가 이 트랜잭션보다 **먼저** 각인하므로 어느
쪽이든 현재 개정이 찍혀 있다. 그래서 `ExitJudgement.ReJudging`으로 **사실을 전달**한다.
추론이 아니라 호출자가 아는 것이다. `TestAJudgementThatIsNotAReJudgementReleasesNothing`.

## I15 — `awaiting`가 사이클 로그에 없었다

`Outcome.AwaitingAdjustment`의 doc comment는 "운영자가 물어볼 상태"라고 적어 두었는데
`logCycle`은 그것을 찍지 않았다. 리뷰어가 지적한 대로, 이 숫자가 보였다면 a083 I1
(missing-order 차단에 crediter가 없다)이 원장에서 바로 드러났을 것이다.
`ReconcileCycle.Awaiting`으로 세고 로그에 넣었다.

## B8 — 해제가 *어떤* 행이든 닫았다 (개정 4, 네 번째 독립 리뷰)

개정 3은 재판정이라는 **사실**을 전달해 "재판정 아닌 판정은 아무것도 안 닫는다"를
얻었다. 네 번째 리뷰가 그 사실이 아직 부족함을 실행으로 증명했다.

`releaseReJudgedQuarantineTx`는 `(position, generation)`의 활성 행을 읽을 뿐,
재시도가 실제로 소비된 `m.reJudgeVersion`과 대조하지 않았다. `judge`의 각인과 이
트랜잭션 사이에 두 가지가 끼어들 수 있다 — 운영자가 각인된 행을 `HUMAN_REPAIR`로 풀고,
병행 관측의 실패한 선택이 새 행을 연다. 그러면 커밋되는 판정이 **그 새 행**을
`SELECTOR_REVISED`로 닫는다. 새 행을 쓴 것은 지금 도는 선택기 자신이므로, 개정 3이
없앤 것과 정확히 같은 종류의 거짓 근거다 — 한 버전만큼 좁아졌을 뿐이다.

"재판정이 있었다"와 "*이* 행이 재판정되었다"는 다른 주장이고, 해제를 정당화하는 것은
두 번째뿐이다. 그래서 bool을 버리고 `ExitJudgement.ReJudgingVersion int64`로 바꿨다 —
0이 "재판정 아님"이고, 그 외는 재시도를 소비한 행의 버전이다. 두 필드가 어긋날 여지를
없애는 것이 부수 효과가 아니라 목적이다: `m.reJudge`와 `m.reJudgeVersion`은
`workingSet`의 같은 분기에서 함께 설정되므로 값 하나로 접을 수 있다.
`TestAReJudgementReleasesOnlyTheRowThatEarnedIt`.

## I16 — 각인 뒤 판정 실패는 격리를 좌초시킨다 (기존, 이 change가 만들지 않음)

`judge`가 각인한 뒤 `record`에 닿기 전에 실패하면 — `breakEven`, 정책 동일성 충돌,
`snapshotContext`, 평가기 오류, `ladderFor`/`checkLadderPolicyStillFits` — 행은 활성인 채
더는 자격이 없다. `workingSet`은 이후 모든 사이클에서 그 포지션을 거절하므로 **손절이
다시 평가되지 않는다.** 운영자 해제나 다음 선택기 개정까지.

의도된 보수 방향이고(`exit_snapshot.go`의 "Spending the retry on the attempt rather than
the outcome") 개정 4가 건드리지 않았다. 여기 적는 이유는 격리된 포지션에 **자동 손절
평가가 전혀 없다**는 사실이 별도 change로 다뤄져야 하기 때문이다. 주문에는 fail-closed,
보호에는 fail-open이다.

## I17 — `ExitArmOutcome`이 재판정 억제를 working-order 억제로 보고한다 (기존)

`exit_state.go`가 `ArmSuppressedReJudge`를 `ExitArmOutcome = "suppressed_working_order"`로
매핑한다. 타입된 `ArmSuppressedReason` 열이 진실을 담고 있으므로 관측성 문제이고 판정에
영향이 없다. 개정 4 이전부터 있었다.
