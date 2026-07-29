# Review: verify-reopens-conditional-chain

## Pre-Edit Gate (High-risk)

```text
Pre-Edit Gate:
- change id / task id: verify-reopens-conditional-chain / 1.1~1.8
- 대상 심볼(패키지.함수):
    internal/verifylive.cleanupFrom          (기존 함수 내부 편집)
    internal/verifylive.RedoSet              (기존 함수 내부 편집)
    internal/verifylive.decidedAfter         (신규 leaf)
    internal/verifylive.subjectLost          (신규 leaf)
    internal/verifylive.dependsOn            (신규 leaf)
    internal/console.pageFuncs               (패키지 var — redoable → inRedo)
    internal/console templates.go 재측정 표  (템플릿 문자열)
- 기존 동작 파악 근거:
    · 호출부: CodeGraph callers(cleanupFrom) = cleanupTargets, PendingCleanup 둘뿐
    · 호출부: CodeGraph impact(RedoSet) = console/data.go readVerify·redoSet,
      console/pages.go handleStart, console/overview.go safetyPanelFrom
    · 기존 테스트: cleanup_test.go 8건, redo_test.go 6건, console/remeasure_test.go 14건
    · 실측 fixture: ~/.local/share/tossos/capability-verify.jsonl(45줄),
      capability-verify-us.jsonl(33줄) — 두 시장의 실제 판정·artifact 순서
    · runner.go:330 — awaiting-restart는 즉시 return(같은 프로세스가 cancel에 도달하지 않는다)
- upstream 상속 테스트 영향: no — internal/verifylive와 internal/console은 TossOS 신규 패키지다.
  전체 `go test ./...` 3712 → 3723 (신규 11건), 회귀 0.
- 실패 테스트 선행 작성: yes — RED 2건을 관측하고 기록했다
  (branch-test-map.md에 실패 출력 원문 포함). 콘솔 측 RED는 수정만 일시 되돌려 재현했다.
- 안전 불변식 §0 위반 여부 검토: 통과 (아래 §0 대조표)
```

### §0 대조

| 조항 | 판단 |
|---|---|
| §0.1 승인 없는 LIVE side effect 금지 | 통과 — 재측정 집합은 **제안이지 인가가 아니다**. `Plan.Authorises`와 배치 승인(새 nonce)은 무변경이며 이 change는 그 경로를 건드리지 않는다. |
| §0.2 토글 OFF = upstream | 해당 없음 — 토글 없음 |
| §0.3 손절·비상 청산 즉시성 | 해당 없음 — 여기의 조건주문은 검증 도구가 등록한 1주짜리 **측정용** 손절이며 엔진의 보호 주문이 아니다. 엔진 경로 무변경. |
| §0.4 rate limit 계상 | 통과 — 새로운 호출 **종류**가 없다. 재측정 1회가 추가하는 것은 기존 `conditional-register` 단계와 동일한 호출이다. |
| §0.5 audit | 해당 없음 |
| §0.6 원장 스키마 | 해당 없음 — 기록 스키마 무변경(`Artifact.Deliberate`·`CreatedAt`은 이미 존재) |
| §0.7 운영 토글 flip은 사람이 | 통과 — 실제 측정(task 3.1)은 사람이 콘솔에서 승인한다 |
| §0.8 scope 밖 주문·위험 코드 변경 금지 | 통과 — 변경은 verifylive 2함수 + console 표시층뿐 |
| §0.9 손절·사이징은 보수 방향만 | 통과 — 정리 가드는 취소를 **덜** 보내는 방향. 재개통이 등록하는 것은 발동가가 시장보다 한참 아래인 1주 손절로 보호를 **더하는** 방향이다. |

## 리뷰 (2026-07-29, 적대적 Eng 관점 포함)

위험 등급이 High-risk(라이브 취소·조건주문 등록)이므로 WORKFLOW "위험 등급 가중"에 따라
적대적 Eng 관점을 포함했다. 아래는 실제로 제기하고 코드·실측 기록으로 답한 반론이다.

### A1. "통과한 단계를 다시 실행한다는 것은 `data.go:331`이 명시적으로 금지한 것 아닌가"

**타당한 지적이고, 이 change의 핵심 논점이다.** 수용하되 좁혔다.

그 주석이 막으려던 것은 "이미 아는 것을 위해 실주문을 내는 것"이다. 여기서 여는 것은 그것이
아니다 — `conditional-register`의 통과가 확립한 성질은 "조건주문이 등록되어 있고 다시 읽힌다"
이고, 그 객체가 사라진 순간 **그 성질은 거짓**이다. 재실행은 아는 것을 다시 재는 것이 아니라
의존 단계가 무언가를 재기 위해 반드시 필요한 전제를 다시 세우는 것이다.

좁음의 증거는 문장이 아니라 실측이다. 현재 US 기록에 규칙을 적용하면 `conditional-register`는
집합에 **들어오지 않는다**(모든 의존 단계 pass, trigger만 deferred). KR에서만 들어온다.

### A2. "재개통이 라이브 조건주문을 하나 더 만들 수 있지 않은가" — **실재하는 위험, 완화됨**

`subjectLost`의 두 번째 조건이 "살아 있는 조건주문이 하나도 없다"이므로, 이미 하나 있으면
열리지 않는다. `TestRedoSetDoesNotReopenWhileTheConditionalIsAlive`가 고정한다.
그리고 등록 자체가 배치 승인을 다시 받는다.

### A3. "정리 가드를 닫으면 조건주문 잔여물이 영원히 남지 않는가" — **검토 후 기각**

등록이 T1, 취소 실패가 T2>T1이면 판정이 생성보다 뒤이므로 가드는 열린다
(`TestAVerdictNewerThanTheConditionalOpensCleanup`). 남는 유일한 경우는 등록 후 사람이 검증을
포기한 때인데, 그것은 이 change 이전에도 `conditional-cancel`이 terminal이 아닌 동안의 정상
동작이었고 `tossctl verify status`가 id를 출력한다.

### A4. "재측정이 등록을 통과시킨 직후, 같은 프로세스의 `conditional-cancel`이 또 지우지 않는가"

**확인함 — 지우지 않는다.** `runner.go:330`이 `awaiting-restart`에서 즉시 return한다.
2026-07-28 실측(proc-7OF)도 register·persist 두 줄만 남기고 끝났다.

실측 기록에 대한 시뮬레이션으로 끝까지 확인했다 — 실제 KR 기록 45줄에 "새 조건주문 등록 +
persist awaiting-restart"를 이어붙이면 `PendingCleanup`이 **비어 있고**, `conditional-persist`는
비-terminal이 되어 재측정 집합에서 빠지고 일반 이어하기가 받는다. 07-28의 교착이 그 지점에서
정확히 사라진다.

### A5. "`RedoableVerdict`는 이제 집합과 다른 답을 한다"

**맞다. 그래서 콘솔 표를 집합에 직접 묻도록 바꿨다**(design.md D3). 바꾸지 않았다면 체인을
여는 바로 그 행이, 조작자가 승인 전에 읽는 표에서 빠졌을 것이다 — RED로 재현해 기록했다.

`RedoableVerdict`는 판정 절반을 답하는 함수로 남겨 두었고 `TestRedoableVerdictAgreesWithTheSet`도
그대로 통과한다. 함수 이름이 이제 절반만 가리킨다는 점은 인정하되, 이름 변경은 이 change의
표면을 넓히므로 하지 않았다.

### A6. "`conditional-register`는 `NeedsHolding: true`다" — **미해결 운영 전제, 사용자에게 보고**

계좌가 KR 종목을 하나도 들고 있지 않으면 재측정은 등록을 건너뛰고 아무것도 측정하지 못한다.
2026-07-26에 정확히 그 이유로 체인 전체가 skipped였다. **내일 장중 창 전에 KR 보유가 있는지
확인해야 한다.** 코드로 막을 수 있는 것이 아니다.

자체 치유는 된다 — 그 경우 등록 판정이 `skipped`가 되고 `RedoableVerdict(skipped)`가 참이므로
집합에 계속 남는다. 창을 하나 잃을 뿐 교착으로 되돌아가지는 않는다.

### A7. "측정에 승인이 두 번 필요하다" — **사실, 운영 주의**

register(승인 1) → persist가 `awaiting-restart`로 정지 → 콘솔 재시작 → 이어하기(승인 2) →
persist·modify·cancel. 승인 창은 5분이며 이어하기는 승인이 아니다. 조작자가 두 번 모두
자리에 있어야 한다. 2026-07-27 US 체인이 같은 방식으로 성공했다.

### A8. "`idempotency-ttl-edge`가 KR·US 재측정 집합에 계속 있다"

이 change와 무관한 기존 동작이다. `--include-ttl-edge` 옵트인 단계이고 preflight가 다시
건너뛴다. 범위 밖으로 둔다.

### 결정

수용하고 진행한다. A6·A7은 코드 결함이 아니라 **운영 전제**이므로 완료 보고와 task 3.1에
남긴다. 미해결 코드 이슈 없음.

## Function Logic Map

적용함. 25개 target — 수정한 기존 함수 2건(`cleanupFrom`, `RedoSet`), 신규 leaf 3건,
이 change의 테스트 함수들, 그리고 파일 끝 덧붙임으로 diff 문맥에 걸렸을 뿐 **수정하지 않은**
함수 3건(base revision으로 고정).

`python3 tools/logic-map/check_analysis.py --change verify-reopens-conditional-chain`
→ `evidence complete or diff-proven exempt`
