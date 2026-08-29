# Review: console-click-approval

Function Logic Map: applied — `analysis/function-logic/` 34 target(수정된 기존 함수 전부,
`check_analysis.py` 통과). 면제 아님.

## 1. Proposal freeze (Eng)

**범위 판정**: 승인의 *형식*만 바꾼다. 승인 이후의 레일(계획 인가 `Plan.Authorises`, 수량
상한, 즉시 취소, `ErrOutsidePlan`, 프로세스 경계, TTL, CSRF, 세션)은 코드·테스트 모두
무변경임을 diff로 확인했다.

**사용자 지시의 위치**: "웹 콘솔에 타이핑 확인을 두지 말라"는 2026-07-27 반복 지시이며
[[no-typed-confirmation-friction]] 기억과 일치한다. 지시는 UX 결정이고, 안전 불변식(§0.1
사람 승인 없는 LIVE mutation 금지)의 변경이 아니다 — 그래서 클릭은 유지하고 자동 승인은
만들지 않았다(design A2).

**대안 기각 기록**: 원클릭(시작=승인)은 계획을 보기 전에 승인하게 되어 배치 모델의 근거를
없앤다. 체크박스는 마찰의 형태만 바꾼 것이라 지시의 취지에 어긋난다. 둘 다 기각(design A6).

## 2. Code review

- `Batch.Summary`/`Prompt` 분리: 목록의 원천은 여전히 `Plan.WriteLines` 하나다.
  `Prompt`가 `Summary`로 시작함을 테스트로 고정했으므로 두 렌더링이 갈라질 수 없다.
- `Batch.Expired`: 창 규칙을 한 번만 쓰기 위한 추출. `Verify`가 이 함수를 호출하므로
  콘솔과 터미널의 만료 판정이 같은 코드다. static guard가 `pages.go`의 `ExpiresAt` 직접
  참조를 금지한다.
- `handleApprove`: nonce 폼 값 읽기 제거. 도달 조건은 세션·CSRF·대기 배치·창이며, 이
  넷을 모두 통과한 POST만 `deliver(nil)`에 이른다. 두 번째 제출은 `deliver`가 false를
  반환해 "창이 닫혔다"로 끝난다(기존 동작).
- `handleStart`: 무동작 방지 분기는 `mode=redo`를 건드리지 않고, `Pending`(비terminal
  단계)이 남아 있으면 그대로 통과시킨다 — awaiting-restart 이어하기가 막히지 않는 것을
  전용 테스트로 고정했다(`TestResumeStaysOfferedWhileAStepIsPending`).
- `ApprovalChannel`: zero value가 기존 문구를 유지하므로 다른 호출자·테스트는 무변경
  (§0.2). 콘솔 배선 한 곳만 클릭 채널을 선언한다.

**남은 관찰**: `writeBanner`의 문구 분기는 채널 상수 비교다. 채널이 세 번째 값을 갖게
되면 배너는 기본(타이핑) 문구로 떨어진다 — 안전 방향이지만 그때 문구를 함께 늘려야 한다.

## 3. Security review (CSO 관점)

**위협 모델 변화**: 콘솔 세션을 탈취한 주체의 승인 비용이 "화면에 보이는 문자열을 타이핑"
에서 "버튼 클릭"으로 낮아진다.

**판정: 수용**. 근거 —
- 타이핑은 이 위협에 대한 통제가 아니었다. nonce는 같은 화면에 평문으로 표시되므로,
  화면에 접근할 수 있는 주체에게는 비용이 0이다. 통제 역할을 하는 것은 루프백 전용
  바인딩·프로세스 수명 세션 토큰·CSRF이며 전부 유지된다.
- CSRF가 유지되므로 외부 페이지가 유도하는 POST는 여전히 불가능하다(전용 테스트 존치).
- 승인은 계획을 표시하는 화면에서만 가능하고, 승인되는 것은 그 계획뿐이다.
- 비대화 승인 경로·환경변수·플래그는 신설되지 않았다. `AUTO_APPROVE` 계열 문자열 금지
  static guard가 그대로 있다.
- 잔여 위험(수용): 이 머신에서 콘솔 화면에 도달할 수 있는 주체는 클릭 한 번으로 최소
  수량 검증 주문을 승인시킬 수 있다. 상한(1주·즉시 취소·목록 밖 금지)은 무변경이다.

**§0.7**: 게이트 ON·운영 토글은 이 change의 범위 밖이며 콘솔에 라우트가 없다(가드 유지).

## 4. QA

- `go test ./...` 3096 passed, `go vet ./...` clean, `make validate` 통과.
- RED 관측: `Summary`/`Expired` 부재로 인한 빌드 실패 → 구현 후 GREEN. 콘솔 클릭 승인·
  무동작 방지 6건은 실패(승인 거부/무동작 run 생성) 관측 후 GREEN.
- 실계좌 QA는 하지 않았다(장 종료 후). 다음 장중 창에서 재측정 실행 시 확인한다.

## 5. 완료 조건

- 미완료 태스크 0, `check_analysis.py` 통과, PM check 통과.
- 사용자 조치: 새 빌드 설치 후 콘솔 재시작 필요(구 프로세스는 구 화면을 서빙한다).
