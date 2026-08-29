# Tasks: console-click-approval

> 목적: 2026-07-28 장중 창에서 2.5·2.7·2.8 측정이 승인 마찰·무동작 기본값으로 다시
> 소모되지 않게 한다. 이 change는 승인의 **형식**만 바꾸고 레일은 건드리지 않는다.

## 1. 승인 형식 [T]

- [x] 1.1 [T] `Batch.Summary()` 분리 — 계획 목록만 렌더(확인 문자열·타이핑 지시 제외).
  `Prompt()`는 `Summary()` + 확인 문자열/만료 + 타이핑 지시로 재구성해 **TTY 출력 무변경**을
  테스트로 고정한다. RED: Summary가 확인 문자열을 담지 않음 + Prompt가 Summary로 시작함.
- [x] 1.2 [T] 콘솔 승인 = 클릭 — `handleApprove`가 nonce를 읽지 않는다. 세션·CSRF·대기 중
  배치·만료만 판정하고 승인을 전달한다. RED: nonce 없는 POST가 승인되어 실행이 진행됨,
  만료된 창의 POST는 거부되고 전송 0건, CSRF 없는 POST는 거부(기존 가드 유지).
- [x] 1.3 [T] 승인 화면이 `Summary()`를 렌더 — 화면에 확인 문자열·"입력하라"가 없고
  계획 목록은 터미널과 같은 원천이다. RED: 렌더된 HTML에 nonce 문자열·타이핑 지시 부재,
  계획 줄 존재.
- [x] 1.4 [T] `Options.ApprovalChannel` — 승인 엔트리 `approval.model` detail이 실제 채널을
  말한다(zero value = TTY 타이핑, 콘솔 배선만 클릭). RED: 콘솔 채널로 돈 run의 승인
  엔트리가 클릭으로 기록됨, 미지정 호출자는 기존 문구 유지.

## 2. 무동작 기본값 제거 [T]

- [x] 2.1 [T] 이어할 단계(`Pending`)가 비었으면 [이어하기] 비활성 + 사유 표시,
  재측정 대상이 있으면 재측정이 기본 동작. RED: 전 단계 terminal인 기록으로 렌더 시
  resume 버튼 disabled·재측정 버튼 활성, 비terminal 단계가 있으면 resume 활성.
- [x] 2.2 [T] 서버측 방어 — `mode=resume`인데 이어할 단계가 없으면 run을 시작하지 않고
  안내한다(무동작 run이 기록에 남지 않는다). RED: 그 조건의 POST가 run을 만들지 않음.

## 3. 문서 정합 [M]

- [x] 3.1 2b `verify-execution-capability/tasks.md` 1.6·1.7의 "nonce 타이핑 폼"·"승인 등가성
  3중" 서술을 이 change의 승인 형식으로 갱신(구현·계약 drift 방지).
- [x] 3.2 `measurements.md`에 2026-07-27 장중 창 소모 사실(M11) 기록 — 무동작 기본값이
  측정 0건을 만든 경위와 증거.

## 4. 완료 게이트 [M]

- [x] 4.1 Function Logic Map(수정 대상 기존 함수) + `check_analysis.py`
- [x] 4.2 `make sdd-sync && make gate CHANGE=console-click-approval`
- [x] 4.3 설치(`make install` 상당) 후 콘솔 재시작 안내 — 사용자가 새 빌드로 열어야 반영
