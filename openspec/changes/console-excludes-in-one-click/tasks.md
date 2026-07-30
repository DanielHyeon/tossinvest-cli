# Tasks: console-excludes-in-one-click

위험도 **High-risk** — 쓰는 것은 config 블록뿐이지만, 그 블록의 소비자가
`ReconcileDriver`이고 제외 여부가 합성 손절의 **생성 여부**를 정한다.
1.x는 전부 RED 선행이다.

## 1. 구현

### 1.1 계약 확인 (구현 전)

- [x] 1.1 §0.4 확인 — 제외가 이미 관리 중인 포지션의 손절을 벗기지 않음을 코드로
  확인하고 근거를 FLM §0에 적는다 (`judgeHoldings`의 `ExitEligible()` 조기 반환이
  제외 판정보다 앞)
- [x] 1.2 Function Logic Map + Branch Test Map — 기존 함수 내부를 고치는 전 대상
- [x] 1.3 Pre-Edit 선언 기록 (review.md)

### 1.2 한 번 클릭 제외

- [x] 1.4 RED — `/settings/exclude`가 세션+CSRF 게이트 뒤에 있고, CSRF 없는 요청은
  거부되며 config가 바뀌지 않는다
- [x] 1.5 RED — 심볼 하나를 보내면 `exclude_symbols`에만 추가되고
  `enabled`·`default_stop_pct`·`include_symbols`는 **읽은 값 그대로**다.
  대소문자·공백은 정규화되고 중복은 1회
- [x] 1.6 RED — `remove=1`이면 그 심볼만 빠지고 나머지 목록은 보존된다
- [x] 1.7 RED — 심볼 없는 요청은 400, seam 미배선이면 501
- [x] 1.8 GREEN — `handleSettingsExclude` + 라우트 등록

### 1.3 제외는 손절폭을 발명하지 않는다 (D4)

- [x] 1.9 RED — `default_stop_pct` 미설정 상태에서 제외 지정 → 저장된 블록의
  손절폭이 여전히 미설정이고 저장이 거부되지 않는다 (편입 지정의 5% 채움과 대비)
- [x] 1.10 GREEN — 제외 경로가 `DefaultStopPct`를 건드리지 않는다

### 1.4 상호 배제 (D3)

- [x] 1.11 RED — 편입 지정된 심볼을 제외하면 같은 저장에서 include에서 빠지고
  exclude에 들어가며, 공지가 편입 지정 해제를 말한다
- [x] 1.12 RED — 제외된 심볼의 행에는 편입 체크박스가 렌더되지 않고 제외 해제
  안내가 대신 나온다
- [x] 1.13 RED — 화면을 우회한 `/settings/include` POST가 제외된 심볼을 가리키면
  공지가 "제외가 우선하여 편입되지 않는다"를 말한다 (편입 예약 성립을 보고하지 않는다)
- [x] 1.14 GREEN — 1.11~1.13 배선

### 1.5 행 표시 (D5·D7)

- [x] 1.15 RED — `positionRow.Excluded`가 seam에서 찍히고, `Label()`이 "관리 제외"를
  돌려준다. 제외+편입 동시 등재는 "관리 제외"가 이긴다. 원장 미판독은 여전히
  "관리 여부 불명"
- [x] 1.16 RED — 관리 중(`Managed`)·불명(`Unknown`) 행에는 제외 컨트롤이 없다
- [x] 1.17 GREEN — 스탬프·라벨·템플릿 컨트롤
- [x] 1.18 RED — 타이핑 확인·추가 승인 마찰이 없다 (기존 편입 컨트롤과 같은
  `confirm()` 1회) — 사용자 결정 2026-07-27

### 1.6 정적 가드 (스펙이 요구하는 동반 갱신)

- [x] 1.19 RED — 라우트 수 하한과 열거 주석, `stateChanging`,
  `consoleStateChanging` 세 곳이 새 라우트를 알고 있다. 하나라도 빠지면 실패
- [x] 1.20 GREEN — 세 가드 갱신. seam은 여전히 Load/Save 2개

## 2. 검증

- [x] 2.1 `make test` / `make vet` / `make validate` — 상속 테스트 회귀 0
- [x] 2.2 review.md — 적대적 Eng 리뷰
- [x] 2.3 `make sdd-sync` → `make sdd-check` → `make gate CHANGE=console-excludes-in-one-click`
- [x] 2.4 PM 등록
