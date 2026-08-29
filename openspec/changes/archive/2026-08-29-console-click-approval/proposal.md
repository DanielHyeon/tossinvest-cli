# Change: console-click-approval

## Why

사용자 지시(2026-07-27, 반복): **웹 콘솔에 타이핑 확인 마찰을 두지 말 것** — 혼자 쓰는
프로그램이다. 현행 콘솔의 배치 승인은 화면에 표시된 nonce를 손으로 받아쳐야 진행되며
(2b tasks 1.6 "nonce 타이핑 폼"), 이는 이미 한 번 제거하기로 결정된 마찰과 같은 종류다.

마찰의 비용이 실측으로 드러났다. 2026-07-27 장중(09:00–15:30 KST) 창에서 검증이 두 번
시작됐고 둘 다 `0 step(s) recorded`로 끝나 **측정 0건**이었다 —
`~/.local/share/tossos/capability-verify.jsonl`의 최신 기록은 여전히 2026-07-26 22:48이다.
원인은 승인 이전 단계였다: 시작 화면의 기본 버튼 [이어하기](`mode=resume`)는 판정이
terminal인 단계를 건너뛰는데, 현재 기록은 모든 단계가 terminal이라 **구조적으로 무동작**
이다. 아래 별도 섹션의 [재측정](`mode=redo`)만이 실제로 무언가를 측정한다. 화면이
무동작을 기본 동작으로 제시했고, 그 결과 하루치 장중 창이 소모됐다.

2c(`add-protection-orders`)의 입력인 2.5·2.7·2.8은 이 창에서만 측정할 수 있다. 승인
마찰과 무동작 기본값을 남겨두면 다음 창도 같은 방식으로 소모될 수 있다.

## What Changes

- **콘솔 배치 승인 = 단일 클릭**: `/verify/approve`가 nonce 입력을 읽지 않는다. 승인의
  성립 요건은 세션 토큰 + CSRF 토큰 + 표시된 계획에 대한 명시적 POST 세 가지이며, 승인
  창 만료(현행 5분)와 "목록에 없는 요청은 전송 금지"(`Plan.Authorises`)는 그대로다.
  타이핑이 사라져도 사람이 계획을 보고 누르는 행위는 남는다 — 자동 승인·비대화 승인
  경로·플래그는 신설하지 않는다(§0.1·§0.7).
- **승인 화면의 문구·요약 분리**: `Batch.Summary()` 신설 — 계획 목록만 렌더하고 확인
  문자열 줄과 "입력하라" 지시는 담지 않는다. `Batch.Prompt()`(TTY용)는 Summary에 타이핑
  지시를 덧붙이는 형태로 유지한다. 콘솔은 Summary를 표시하므로 화면과 실제 승인 방식이
  어긋나지 않는다.
- **증거 기록의 정직성**: 승인 엔트리의 `approval.model` detail이 실제 승인 채널을
  말한다 — `Options.ApprovalChannel`(기본 TTY 타이핑, 콘솔은 클릭)을 기록한다. 현재
  문구("one typed expiring string for the whole run")는 콘솔 클릭 승인에서는 거짓이 된다.
- **무동작 기본값 제거**: 시작 화면은 아무 단계도 실행하지 않을 동작을 기본 동작으로
  제시하지 않는다 — 남은(비terminal) 단계가 없고 재측정 대상이 있으면 [재측정]이 기본
  동작이고, [이어하기]는 비활성으로 그 이유와 함께 표시된다.
- **2b tasks 문구 정정**: tasks 1.6·1.7의 "nonce 타이핑 폼"·"승인 등가성 3중" 서술을 이
  change의 승인 형식으로 갱신한다(구현과 계약 문서의 drift 방지).

## Non-Goals

- CLI TTY 경로(`tossctl verify run`)의 타이핑 확인 — 유지. 콘솔 밖에서 자동화가 승인을
  흉내낼 수 없어야 한다는 요구는 그대로다(`ErrNotATerminal`, 자동화 플래그 부재).
- `flatten-all`의 타이핑 확인(engine-safety 스펙) — 무접촉.
- 클릭조차 없는 완전 무인 승인(장 개장 시 자동 실행 등) — 만들지 않는다. 실계좌
  mutation은 사람의 행위 하나를 반드시 통과한다(§0.1).
- 승인 창 TTL 연장, 배치 목록의 축약, 게이트·주문 라우트 신설 — 전부 범위 밖.

## Capabilities

### Added Capabilities

- `operator-console`: "검증 배치 승인의 형식", "검증 시작 화면의 무동작 방지"

## Impact

- Affected code: `internal/verifylive/confirm.go`(Summary 분리), `internal/verifylive/runner.go`
  (ApprovalChannel 기록), `internal/console/pages.go`(handleApprove·renderVerify),
  `internal/console/templates.go`(승인 폼·시작 화면), `cmd/tossctl/console.go`(채널 배선)
- 안전 검토(§0): 승인 레일의 **형식**만 바꾼다 — 계획 인가(`Plan.Authorises`)·상한·즉시
  취소·`ErrOutsidePlan`·프로세스 경계·TTL·CSRF·세션은 전부 무변경. 콘솔의 계좌 접근
  능력은 늘지 않고, 비대화 경로는 신설되지 않으며, 손절·익절·사이징 경로는 무접촉.
  잔여 위험: 탈취된 콘솔 세션의 승인 비용이 "타이핑 1회"에서 "클릭 1회"로 낮아진다 —
  수용 근거는 루프백 전용 바인딩·프로세스 수명 세션·CSRF·계획 목록 표시이며, 타이핑
  마찰은 이 위협에 대한 실효 통제가 아니라는 사용자 판단이다(review.md에 기록).
- PM: registry bootstrap allowlist 등재(archive 시 제거)
