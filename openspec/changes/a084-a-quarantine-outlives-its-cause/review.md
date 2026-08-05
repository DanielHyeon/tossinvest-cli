# a084 · Review

## 2026-08-05 · proposal-freeze

**보이스 구성**: Manager 셀프 적대적 Eng 1인.
**독립 리뷰어 부재.** `codex exec`(gpt-5.6)가 사용량 한도(2026-08-08 해제)로 거부됐다.
a083 `issues.md` I5와 같은 제약이며, 배포 전 별도 세션의 재검증이 남아 있다.

### 발견 1 (수용, 설계 유지) — 재판정이 즉시 청산을 낼 수 있다

격리 동안 high water는 얼어 있고 시장은 얼어 있지 않다. PLTR은 격리 시점 148.55에서
현재 159.62로 움직였다. 재판정이 현재가를 보고 즉시 주문을 낼 수 있다.

의도된 동작이다. 저장된 보호선 **아래**에서만 청산이 나가고, 그것은 §0.3이 지연을
금지하는 바로 그 평가다. 지금은 그 평가가 16시간째 멈춰 있다. 실제 저장값으로 재생하면
159.62에서 rung 3 · protection 151.6518 · `LADDER_HOLD_STOP_PROMOTED` — 주문 없이
보호선만 승격한다.

### 발견 2 (수용, 설계 반영) — 거부 경로가 순환할 수 있다

`quarantineExitSnapshotTx`가 활성 행을 그대로 반환하면, 각인이 다른 행은 매 관측마다
재판정되고 매번 거부되어 원장에 아무것도 남기지 않은 채 무한 재시도한다.

수용. 거부 시 옛 행을 `SELECTOR_REVISED`로 닫고 현재 개정으로 각인한 새 행을 연다.
상한은 개정 수이고 개정은 사람이 올린다(design D6). 공개 진입점
`Journal.QuarantineExitSnapshot`에도 같은 규칙을 넣었다 — 두 생성 경로가 갈리면
한쪽만 재판정된다.

### 발견 3 (수용, 별도 수정) — v27 마이그레이션 테스트가 §0.6과 충돌한다

`journalV25RowFingerprints`가 **열 이름을 해시에 포함**해서,
`TestMigrationV27AddsPairedWeeklyAuthorityWithoutChangingV26Rows`가 v26 이전 테이블에
열을 추가하는 것을 전부 거부한다. 이름은 "WithoutChangingV26**Rows**"인데 실제로는
열까지 얼린다.

이것은 마이그레이션 계약 자체와 모순이다. `schema.go`의 규칙 2는 **"New columns are
nullable or carry a DEFAULT"**로 nullable ADD COLUMN을 명시적으로 허용하고,
`positions.adoption_id`와 `mutation_attempts`의 열 11개가 이미 그 경로를 지났다.
v27~v29가 새 테이블만 만들었기 때문에 이 과도한 불변식이 한 번도 드러나지 않았을 뿐이다.

fingerprint를 **행만** 해시하도록 좁히고, 규칙 3이 실제로 금지하는 것 — 열의 삭제·
개명·순서 변경 — 을 `assertColumnsOnlyAppended`로 따로 단언했다. 이전보다 강한 검사다:
전에는 열 집합이 바뀌었다는 것만 알았고, 이제는 **무엇이 어떻게** 바뀌었는지 말한다.
v26·v27·v30 세 테스트가 모두 이 단언을 쓴다.

### 발견 4 (기록) — 개정 상수를 손으로 올린다

`RecoverySelectorRevision`은 사람이 올린다. 자동 digest 산출은 무관한 리팩터가 전
포지션을 재판정하게 만든다. 빠뜨리면 자동 재시도가 없다 — 오늘의 동작이므로 fail-safe.
불필요하게 올리면 같은 선택기가 같은 답을 내고 다시 격리한다 — 행 하나가 더 생길 뿐이다.
어느 쪽으로 틀려도 새 위험이 없다(design D3).

### 발견 5 (기록, 범위 밖) — evidence가 어느 분기에서 나왔는지 말하지 않는다

세 건의 evidence가 전부 접미사 없는 `exitpolicy: recovery candidate identity mismatch`다.
`SelectRecoverySnapshot`에서 이 맨 sentinel을 반환하는 경로는 정체성 튜플 검사와
`compareRecoveryStage`의 rank 실패 둘뿐이고, 사후에 어느 쪽인지 알 수 없다. `issues.md`.

## Pre-Edit Gate

```text
- change id / task id: a084-a-quarantine-outlives-its-cause / 6.1–6.5
- 대상 심볼(패키지.함수):
    engine.ExitObserver.workingSet                (exitloop.go)
    journal.Journal.recordExitJudgementTx         (exit_state.go)
    journal.quarantineExitSnapshotTx              (exit_snapshot.go)
    journal.Journal.QuarantineExitSnapshot        (exit_snapshot.go)
    journal.activeExitSnapshotQuarantineTx        (exit_snapshot.go)
    journal.Journal.ReleaseExitSnapshotQuarantine (exit_snapshot.go)
- 기존 동작 파악 근거:
    Function Logic Map 7건 + Branch Test Map 7건 (analysis/function-logic/, AST 분기 전수)
    운영 원장 읽기 전용 조사: 활성 격리 3건 전부 ambiguous_recovery,
      저장 snapshot을 현재 비교기로 전 가격축 재생 시 격리 0건 —
      PLTR 148.55/148.73, 032820 11340/11350.70, NNE 18.45/18.53 (첫 rung 교차 직후)
    a078 issues.md I1이 이 설계 공백을 명시적으로 후속 change 후보로 남겼다
- upstream 상속 테스트 영향: no (exit-policy·격리는 TossOS 고유 경로)
- 실패 테스트 선행 작성: yes (5.1이 RED 정본, 4.x가 원장 계층)
- 안전 불변식 §0 위반 여부 검토: 통과
    §0.3 청산 즉시성 — 회복 방향이다. 지금 격리된 포지션은 손절이 평가되지 않는다
    §0.6 additive-nullable — nullable 1열, backfill 없음, 마이그레이션 테스트로 고정
    §0.7 사람 승인 — 콘솔 해제(a079)는 그대로다. 같은 개정의 격리는 여전히 사람만 푼다
    §0.9 단방향 안전 — 손절폭·익절폭·사이징을 바꾸지 않는다. 저장 기준선을 그대로 쓴다
```

## 결정

**SHIP WITH CHANGES.** 발견 2·3을 반영했다. 발견 5는 `issues.md`. 배포 전 별도 세션의
독립 재검증과 `main` 대비 `SchemaVersion` 대조가 남아 있다.

## 2026-08-05 · 독립 리뷰 (별도 컨텍스트)

앞의 proposal-freeze 기록은 **Manager 셀프 리뷰**였고, WORKFLOW가 요구하는
"작성자와 검증자의 분리"를 충족하지 못했다. 그 지적을 받고 gstack `/review`를 실행했다.

**보이스 구성**: 별도 컨텍스트 리뷰어 4명 — 적대적 Eng, 보안·안전 불변식, 데이터
마이그레이션, 테스팅/유지보수·성능. 각자 신선한 컨텍스트에서 diff를 읽었고 작성자의
근거를 알지 못한다.
**codex(gpt-5.6)는 여전히 사용량 한도**(2026-08-08 해제)로 교차 모델은 없다.

**결과: 셀프 리뷰가 놓친 blocking 결함을 찾았다.** `issues.md`의 `B` 항목이 정본이며,
여러 건은 실행 가능한 재현으로 확인했고 나는 원본 코드를 직접 읽어 재확인했다.

**결정: 배포 보류.** blocking 항목이 닫히기 전에는 `make gate` 통과 여부와 무관하게
a084 를 배포하지 않는다. 게이트는 테스트가 green이라고 말할 뿐, 여기서 발견된 것들은
테스트가 없어서 green이었다.

### 개정 2 반영 결과

D7 자격은 사유와 방향을 본다 · D8 자격은 시도로 소진된다 · D9 재판정은 익절을 미루되 보호를 미루지 않는다 · 중복 writer 통합 · CI timeout.

blocking 항목은 전부 RED 테스트 선행 후 닫혔다. 설계는 `design.md` 개정 2,
요구사항은 spec delta 개정 2, 작업은 `tasks.md` 10장에 있다.

리뷰 과정에서 **내 수정 자체가 만든 새 위험을 하나 더 찾았다**: a084 의 초기 D9 안은
보류된 제안이 재제안되지 않는다는 성질과 합쳐져 손절을 무기한 미룰 수 있었다.
`isProtective`로 보호 계열을 보류 대상에서 제외하고
`TestAReJudgementNeverWithholdsAStop`으로 고정했다. 잔여 성질은 `issues.md` I12.
