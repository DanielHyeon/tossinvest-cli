# Review: verify-holds-what-it-awaits

날짜: 2026-07-30 · 보이스: CEO/Eng(적대적)/DX/QA · 시점: proposal-freeze
위험 등급: **High-risk** (라이브 취소 요청의 대상 목록)

## Pre-Edit Gate

```text
change id / task id:  verify-holds-what-it-awaits / 1.1-1.9
대상 심볼:
  verifylive.Artifact                     (record.go:177)  — 가산 필드 2개
  verifylive.cleanupFrom                  (cleanup.go:121) — 판정 규칙 통합
  verifylive.decidedAfter → heldAfter     (cleanup.go:150) — 비교 기준 확장
  verifylive.(*stepRun).markDeliberate    (runner.go:730)  — gate·사슬 함께 기록
  verifylive.(*Runner).stepConditionalRegister/Persist/Modify (steps.go:649,707,770) — 호출부 3곳
기존 동작 파악 근거:
  cleanup_test.go(340줄), record_test.go, redo_test.go, plan_test.go, static_test.go
  실기록 ~/.local/share/tossos/capability-verify{,-us}.jsonl 60+37 entry
  호출부: cleanupTargets·PendingCleanup(console 2곳)·planCleanup·runCleanup
upstream 상속 테스트 영향: no — internal/verifylive는 TossOS 전용 패키지다
실패 테스트 선행 작성: yes (tasks 1.1-1.9 전부 [T])
안전 불변식 §0 위반 여부 검토: 통과 (아래 표)
```

## §0 대조

| 조항 | 이 change |
|---|---|
| 1 승인 없는 LIVE side effect 금지 | 위반 없음 — 취소 요청은 여전히 승인 목록을 거친다. 이 change는 그 목록을 **줄이는 쪽으로만** 열린다 |
| 2 토글 OFF = upstream 동작 | **핵심 근거**. 필드 부재 = 오늘의 판정. 기존 기록 전부가 이 경로다 (task 1.8이 증명) |
| 3 손절·비상 청산 즉시성 | 무관 — verifylive는 측정 도구이고 엔진 청산 경로를 건드리지 않는다 |
| 4 rate limit 예산 | 무관 — API 호출을 추가하지 않는다. 오히려 취소 요청이 줄 수 있다 |
| 5 운영 설정 audit | 무관 |
| 6 원장 스키마 additive-nullable | 준수 — `omitempty` 2필드, `FormatVersion` 무변경 |
| 7 운영 토글 flip은 사람이 | 무관 |
| 8 scope 밖 주문·위험 코드 금지 | 준수 — `internal/verifylive` 4파일. 엔진·Guardian·journal 무접촉 |
| 9 손절 로직은 보수 방향만 | 무관. 다만 이 change의 방향 자체가 보수적이다 — 살아 있는 브로커 객체를 **덜** 취소한다 |

## 적대적 Eng 리뷰

### A1 — "취소 대상을 줄이기만 한다"는 거짓이다 (수용, proposal 수정)

**지적**: 줄어드는 게 아니라 **아무것도 안 바뀐다**. `HeldUntil`을 주문에 찍는 코드가 이
change에는 없고, 조건주문 호출부는 기본값과 같은 `conditional-cancel`을 찍는다. 목록은
동일하다. "줄인다"는 표현은 안전 방향을 과장한다.

**수용**. proposal Impact를 "이 change 단독으로는 취소 대상 목록이 바뀌지 않는다"로 고쳤다.
그리고 그 무변화 주장 자체를 task 1.8(실기록 회귀)이 증명 대상으로 삼는다. 관측 가능한 차이가
없다는 것은 이 등급의 change에서 **약점이 아니라 목표**다.

### A2 — D4의 "두 필드는 독립"이 D1과 모순 (수용, design 수정)

**지적**: D1의 붙잡음 탐색이 `Deliberate`를 읽는다면 D4가 주장한 분리는 성립하지 않는다.

**구현에서 해소됐다 — 이 지적의 전제가 틀렸다.** 정리 판정 경로가 `Deliberate`를 읽어야 할
이유가 없었다. 기본 gate는 `Kind`로 갈리고(`holdGate`), 그것이 옛 `cleanupFrom`의
`switch a.Kind`가 하던 일과 정확히 같다. 따라서 `cleanupFrom`·`holdGate`·`heldAfter` 어디도
`Deliberate`를 읽지 않으며 D4의 분리 주장은 구현에서도 성립한다.

부수 효과로 보존 범위가 넓어졌다 — `Deliberate`가 **아닌** 조건주문도 같은 gate를 받으므로
기존 판정 보존이 "대부분"이 아니라 전부다.
`TestALegacyRecordIsJudgedExactlyAsBefore`의 마지막 항목이 그것을 고정한다.

### A3 — `heldAfter`가 `decidedAfter`보다 느슨해질 수 있는가 (검토, 불가)

기준 index가 "최초 언급"에서 "가장 최근 붙잡음"으로 옮겨간다. 붙잡음 줄은 항상 최초 언급
이후이므로 기준 index는 **단조 증가**하고, 판정식이 `decided > 기준`이므로 조건은 **더
까다로워지기만** 한다. 즉 `heldAfter`는 `decidedAfter`보다 대상을 **같거나 적게** 고른다.
느슨해지는 방향은 구조적으로 없다.

붙잡음 줄이 하나도 없는 outstanding artifact(등록 직후 조회 실패로 조기 반환한 조건주문 등)는
기준이 **최초 언급**으로 떨어져 `decidedAfter`와 정확히 같다.

### A4 — `ChainID`는 이 change에 필요 없다 (부분 수용, 근거 기록)

**지적**: 정리 불변식은 `ChainID` 없이 성립한다. 정정이 붙잡음을 잇는 것도 새 artifact에
`HeldUntil`을 찍으면 된다. 범위 밖이다.

**부분 수용 — 넣되 필수가 아님을 명시한다.** 근거: ① M40이 부모→자식 링크를 산문 `note`에만
남겼고 그것이 지금 존재하는 데이터 손실이다, ② 발동 change가 어차피 필요로 하며 기록 스키마를
두 번 흔드는 것이 §0.6에 반한다. **불변식의 근거로는 쓰지 않는다** — task 1.7은 `HeldUntil`
이전으로 검증하고 `ChainID`는 함께 실리는지만 본다.

### A5 — 붙잡힌 주문이 노출 상한을 채운다 (이연, issues.md)

`MaxLiveOrders = 1`([mutate.go:75](../../../internal/verifylive/mutate.go)). 발동 change가
child 주문을 붙잡으면 그 1칸이 차서 이후 mutating 단계가 전부 `ErrExposureCap`으로 막힌다 —
M22와 같은 형태다. 이 change는 그 상태를 만들지 않지만(주문을 붙잡는 코드가 없다) **발동
change는 반드시 만든다**. 붙잡힌 child를 상한에 셀지는 그 change의 결정이며 issues.md에 이연.

### A6 — 오타난 `HeldUntil`은 영구 교착 (수용, task 1.9 확장)

카탈로그에 없는 StepID를 gate로 쓰면 `Settled`가 영원히 false여서 객체가 영구히 붙잡힌다.
fail-closed가 **조용한 교착**이 되는 유일한 경로다. task 1.9에 AST 가드 ②로 추가:
production이 쓰는 모든 `HeldUntil` 값은 카탈로그 실재 단계여야 한다.

### A7 — 시각 기반 만료 거절이 D5의 교착을 만든다 (수용, 잔여 위험으로 기록)

**지적**: 만료가 없으면 버려진 측정이 `MaxLiveConditionals`를 영구히 채우고, 콘솔에는 그
사슬을 끝낼 조작이 없다(`RedoSet`은 마지막 판정이 fail·skipped인 단계만 담으므로 gate가
`pass`면 목록에 없다). 이것은 M37을 벗어날 수 없게 만들었던 바로 그 구조다.

**수용하되 만료를 도입하지 않는다.** 경과 시간으로 살아 있는 브로커 객체를 취소 대상으로
되돌리는 규칙은 M37의 형태 그 자체이고, 발동 대기 창 한복판에서 발화한다. 대신 **차이를
명확히 한다** — M37은 조용하고 자동이었지만 이 상태는 `Outstanding`·화면·실행 요약이 매번
보고하는 **보이는 정지**다. 탈출구(CLI `--redo <gate>`)도 존재한다.

그래도 콘솔 단독 운용에서는 탈출구가 닿지 않는다. **해소하지 않고 D5·issues.md에 이연**하며,
발동 change가 운영자 조작을 함께 만들어야 한다는 것을 그 change의 선행 조건으로 적는다.

### A8 — 기존 실기록을 신 코드로 재생하면 07-28과 다른 답이 나온다 (검토, 의도)

07-28에 `grLKqiGuC…`가 실제로 취소된 것은 그때 버그가 있었기 때문이고, 현재 코드로 그 기록을
재생하면 취소하지 않는다(`verify-reopens-conditional-chain`의 수정). task 1.8의 회귀 기준은
**역사적 결과가 아니라 현재 HEAD의 판정**이어야 한다. 테스트 주석에 명시한다.

### A9 — Function Logic Map 범위

기존 함수 내부 편집이므로 면제 없음. 대상: `cleanupFrom`, `decidedAfter`(→`heldAfter`),
`markDeliberate`, 그리고 diff 문맥에 걸리는 인접 무변경 함수는 base revision AST로 충족.

## 구현 중 확인한 사실

### F1 — `conditional-persist`의 표시 호출은 오늘 no-op이다

`steps.go`의 존속 단계는 조건주문을 **되읽기만** 하고 artifact를 기록하지 않는다. 실계좌 기록이
그것을 증명한다 — `run-OJRFYBGI4UOBM4MD`의 `conditional-persist` pass 항목에 `artifacts`가
없다. 따라서 그 자리의 표시 호출은 순회할 대상이 없어 아무 일도 하지 않는다.

**이 change의 판정이 그 호출에 의존하지 않는다**는 뜻이므로 그대로 뒀다. 붙잡음은 등록 줄이
선언한 것이 계속 유효하다. `markHeld` 주석에 이 계약을 적어, 객체를 만들지 않는 단계에서
호출이 무해한 no-op임을 명시했다 — 붙잡음을 선언하는 두 번째 경로가 생기지 않게 하는 것이
여기서 중요하다.

### F2 — 정적 가드가 처음에 아무것도 검사하지 않았다

`HeldUntil:` 복합 리터럴만 찾도록 썼는데 production은 gate를 **`markHeld`의 인자**로 넘긴다.
가드는 통과했지만 검사 대상이 0건이었다. 두 철자를 모두 보도록 고치고 `checked == 0`이면
`t.Fatal`하도록 fail-closed로 만들었다.

변이로 확인했다 — 호출부 하나의 gate를 `CleanupLabel`(카탈로그 밖 상수)로 바꾸자
`the hold gate is CleanupLabel ("이전 실행이 남긴 객체 정리"), which is not a step in the
catalogue`로 실패했고, 되돌린 뒤 통과했다.

## RED → GREEN

```text
RED 1 (컴파일):
  hold_test.go:33:21: unknown field HeldUntil in struct literal of type Artifact
  hold_test.go:33:38: unknown field ChainID in struct literal of type Artifact

RED 2 (필드 추가 후 — 실제 결함 재현):
  --- FAIL: TestAHeldOrderIsNotACleanupTarget
      the prologue would cancel the child order the trigger measurement has to watch fill;
      the step that placed it has not reached a terminal verdict:
      [{Kind:order ID:child-1 ... HeldUntil:conditional-trigger ChainID:chain-A}]
  --- FAIL: TestAGateVerdictOlderThanTheHoldDoesNotRelease
  --- FAIL: TestAReDeclaredHoldOutlivesAnOlderVerdict
  --- FAIL: TestAModifiedConditionalKeepsTheChainAndTheHold
      the conditional the modify created carries no chain ... the replacement arrived unheld
  --- FAIL: TestTheCleanupDecisionDoesNotReadAClock
      heldAfter was not found in cleanup.go — this guard is asserting nothing

  RED 2에서 이미 통과한 것들(기존 규칙이 이미 옳던 자리):
  TestAHeldConditionalIsNotACleanupTarget, TestAHoldEndsWhenItsGateDecides,
  TestAFailedCancelAfterTheHoldStillReleases, TestALegacyRecordIsJudgedExactlyAsBefore(5종),
  TestAHeldArtifactSurvivesTheRecord, TestAnUnheldArtifactWritesNoHoldFields

GREEN: go test ./internal/verifylive/ -count=1 → ok (10.9s)
```

## 실계좌 기록 재생 — 판정 무변화 확인

현재 코드로 계좌의 실제 기록 두 개를 그대로 재생했다.

```text
capability-verify.jsonl    60 entries, outstanding=0, pendingCleanup=0
capability-verify-us.jsonl 34 entries, outstanding=0, pendingCleanup=0
```

이 change 전과 같다(둘 다 07-29 실행이 잔여물 0으로 끝났다). 판정 경로를 바꾸고도 실계좌
목록이 움직이지 않는다는 것이 A1이 요구한 증거이며, 합성 fixture 5종은
`TestALegacyRecordIsJudgedExactlyAsBefore`가 고정한다.

## 실행한 명령

```text
openspec validate verify-holds-what-it-awaits --strict --no-interactive   → valid
python3 tools/sdd/capture_change_base.py --change verify-holds-what-it-awaits
  → d36da6d984d48df5e739b4cd94c2265f36b64bbd
go test ./... -count=1                → 3766 passed, 0 failed (기준선 d36da6d: 3749 → +17)
go vet ./...                          → 이슈 없음
python3 tools/logic-map/check_analysis.py --change verify-holds-what-it-awaits
  → evidence complete or diff-proven exempt (13 target, base revision 2건)
```

Function Logic Map 13건 중 base revision 2건은 이 change가 **삭제·개명한** 함수다 —
`cleanup.go:decidedAfter`(→`heldAfter`), `runner.go:stepRun.markDeliberate`(→`markHeld`).
