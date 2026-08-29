# Change: console-excludes-in-one-click

## Why

사용자 요청(2026-07-30): 종목 제외를 **한 번의 클릭**으로 할 수 있어야 한다. 사용자가
심볼을 일일이 읽고 설정 파일 목록을 손으로 고치는 것은 너무 어렵다.

이 요청은 직전 change(`verify-observes-the-trigger`)가 사용자에게 넘긴 3.1 —
"`engine.adoption.exclude_symbols`에 측정 종목을 고정한다" — 에서 나왔다. 그 작업이
사용자 몫인 것은 §0.7(운영 설정 flip은 사람이 한다) 때문이지 **손으로 해야 하기**
때문이 아니다. 사람이 누르는 버튼은 사람이 한 것이다.

현행 화면은 비대칭이다.

- **편입 지정**: 보유 화면(`/positions`)의 그 심볼 행에서 체크 한 번.
  핸들러는 `Load → IncludeSymbols 한 필드만 수정 → Save`의 외과적 경로다.
- **제외 지정**: 경로가 없다. `/settings`를 열고, 쉼표로 이어붙은 `exclude_symbols`
  텍스트를 **읽고**, 거기에 심볼을 타이핑해 넣고, 블록 **전체**(enabled ·
  default_stop_pct · include · exclude)를 다시 쓰는 폼을 저장해야 한다.

두 번째 경로의 문제는 불편만이 아니다. 그것은 사람의 손을 거치는 read-modify-write이고,
목록이 길수록 한 심볼을 조용히 떨어뜨리기 쉽다. 떨어진 심볼은 **제외가 풀린 심볼**이고,
제외가 풀리면 엔진이 다음 대사 주기에 그 보유를 편입한다. `LoadRawEngineAdoption`이
raw 블록을 따로 읽도록 만들어진 이유가 정확히 그것 — "설정 화면의 왕복이 장기 보유를
지키던 exclude 목록을 지우는 것"(console-adoption-controls 리뷰 1회차 P1-1) — 인데,
코드에서 막은 그 실패를 지금은 사람이 키보드로 재현할 수 있다.

## What Changes

- **보유 화면의 행별 제외 컨트롤**: `/settings/exclude`(POST, 세션+CSRF) 신설.
  `symbol` 하나와 선택적 `remove=1`을 받아 `Load → ExcludeSymbols만 수정 → Save`.
  기존 편입 지정 경로와 같은 seam, 같은 검증, 같은 감사 기록을 쓴다.
- **행 상태 표시**: `positionRow.Excluded`를 핸들러가 seam에서 찍는다(표시 전용,
  `Designated`와 같은 방식).
- **모순 상태를 UI에서 원천 차단**: 엔진은 제외를 편입보다 **우선**한다
  (adoption.go — exclude가 항상 우선). 따라서
  - 제외를 걸 때 그 심볼에 편입 지정이 있으면 **같은 저장에서 지정을 함께 내린다**.
    보수 방향이므로 자동으로 한다. 무엇이 함께 내려갔는지는 공지에 적는다.
  - 제외된 심볼의 행에는 편입 체크박스를 **렌더하지 않는다**. 대신 제외 컨트롤과
    "편입하려면 제외를 먼저 해제한다"를 보여준다.
  - 화면이 감췄다는 것은 강제가 아니므로, `handleSettingsInclude`는 직접 POST로
    도달한 제외 심볼에 대해 **지정은 저장하되 공지에서 편입되지 않는다고 말한다**.
- **제외는 손절폭을 발명하지 않는다**: 편입 지정은 블록을 "의미 있게" 만들어
  `default_stop_pct` 검증을 부르므로 기본값 5%를 명시적으로 채운다. 제외 목록은
  `Adoption.validate()`의 그 조건에 들어가지 않는다 — 제외 경로는
  `DefaultStopPct`를 **건드리지 않는다**.

## Non-Goals

- 편입 체크박스의 철자·동작 변경. "미체크 = 관리 외(미편입), 체크 = 관리 편입"은
  사용자 UX 결정(2026-07-27)이며 이 change는 그것을 3-상태 라디오로 재설계하지 않는다.
- `/settings` 전체 폼의 제거. 이 change는 빠른 경로를 **추가**하고, 여러 심볼을 한 번에
  손보는 일반 경로는 그대로 둔다.
- **엔진 관리 중(`Managed`)인 행의 제외**. 제외는 편입(진입)만 막는다 —
  `judgeHoldings`는 `ExitEligible()`에서 제외 판정보다 **먼저** 반환하므로 이미 관리
  중인 포지션에는 아무 효과가 없다. 효과 없는 버튼을 그리는 대신 그 행에는 컨트롤을
  두지 않는다(전체 폼은 여전히 열려 있다).
- 타이핑 확인·추가 승인 마찰. 사용자 결정(2026-07-27)이며 이 change도 지키지 않는다 —
  기존 편입 컨트롤과 동일한 `confirm()` 한 번뿐이다.

## Impact

- `operator-console`: "콘솔 안전 불변식"의 상태변경 행위 열거에 **종목 제외 지정**을
  더한다. 이 열거는 정적 검사(`consoleStateChanging`)가 코드에서 강제하고 있으므로
  스펙 문장과 검사 목록이 함께 움직여야 한다.
- 코드: `internal/console`(settings.go · portfolio.go · portfolio_pages.go ·
  templates_portfolio.go · console.go 라우트 등록), 정적 가드 3종.
- 계좌·journal·브로커: 무접촉. 쓰는 것은 `engine.adoption` 블록뿐이다.
- 엔진: 코드 변경 없음. 소비자는 기존 `Adoption.Excludes`이며 의미도 그대로다.
