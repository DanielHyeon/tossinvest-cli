# Change: console-adoption-controls

## Why

사용자 요청(2026-07-27): ① **자동 편입 켜기**를 웹 콘솔의 별도 설정 화면에서 할 수 있어야
하고, ② **종목별(항목별) 편입 지정**을 웹 화면에서 할 수 있어야 한다. 현행 계약은 둘 다
막는다 — adoption 설정은 config.json 수동 편집뿐이고(adopt-external-positions design A3),
콘솔의 상태변경 행위는 검증 제어·프로세스 기동/정지로 한정되며(operator-console 안전
불변식), 편입 단위 제어는 exclude 목록(opt-out)뿐이라 "체크한 종목만 편입"이 표현되지
않는다.

## What Changes

- **요구사항 개정(2026-07-31 — 메뉴 발견 가능성)**: 이미 구현된 `/settings`의 편입
  설정을 새 엔진 기능으로 복제하지 않고, 상단 메뉴에서 **외부 종목 자동관리**라는
  이름과 `#adoption` 직접 링크로 노출한다. 화면 첫 제목·설명은 수동 매수 종목이 기존
  익절·보호선·손익 극대화 정책에 편입된다는 결과와, 실제 실행 주체가 엔진 대사
  루프라는 점을 명시한다. 기존 enabled·보호폭·제외·지정 저장 seam과 검증·audit을
  그대로 사용한다.
- **종목별 편입(config)**: `adoption.include_symbols` 신설 — enabled=false여도 이 목록의
  심볼은 편입 후보가 된다. 판정: `(enabled ∨ include(심볼)) ∧ ¬exclude(심볼)` — exclude가
  항상 우선. include 경유 편입도 신선 조건·staleness·합성 손절 규칙 전부 동일하다.
  zero-value 안전: 빈 목록 = 현행 동작 그대로(§0.2). include가 비어 있지 않으면
  `default_stop_pct` 검증이 요구되고, 거부 시 블록 전체 zeroing(기존 규칙 유지).
- **콘솔 설정 화면**: `/settings`(GET) — adoption 블록 상태(거부 사유 포함) 표시,
  `/settings/save`(POST, 세션+CSRF) — enabled·default_stop_pct·exclude·include 편집.
  콘솔은 **config만 쓴다**: 주입된 저장 seam이 config.json의 `engine.adoption` 블록만
  외과적으로 교체(다른 키 보존, 원자적 tmp+rename). journal·브로커·계좌는 계속 무접촉.
- **포지션 화면 행별 지정**: 관리 외(미편입) 보유 행에 [관리 편입 지정] 버튼
  (`/settings/include` POST) — include 목록에 그 심볼을 추가한다. 버튼은 편입을 수행하지
  않는다 — **편입 실행 주체는 변함없이 엔진 대사 루프**이며, 화면은 "다음 대사 주기에
  편입 후보"가 됨을 안내한다.
- **반영 시점의 정직성**: 엔진은 기동 시 config를 읽으므로 저장은 "다음 엔진 기동부터
  반영"으로 안내한다(가동 중 hot-reload는 범위 밖). 지금은 automation gate OFF(2c 전)라
  이 change는 **가동 준비 표면**이다 — 저장돼도 실계좌 행동은 게이트 ON(§0.7 사람 승인)
  이후에만 생긴다.
- **audit(§0.5)**: 이중 기록 — ① 콘솔 저장 seam(cmd/tossctl)이 저장 시점에 audit 엔트리를
  직접 append하고(기동 없는 flip·flip 시퀀스 붕괴 보완), ② 엔진 기동 시
  `recordGateSettings`가 직전 대비 변경분을 남긴다. `include_symbols`를
  `recordGateSettings` 항목에 추가한다(종목별 지정은 특정 실보유를 매도 가능 관리 대상으로
  만드는 설정이므로 무audit일 수 없다). 콘솔 저장 자체는 사람의 직접 행위(루프백
  세션+CSRF)다(§0.7 — TossOS가 자동 flip하는 경로는 여전히 없다). 타이핑 확인 문구는
  두지 않는다(사용자 결정 2026-07-27 — 리뷰의 보완 통제 제안을 기각).
- **위협 모델의 정직한 확장**: 이 change는 탈취된 콘솔 세션의 blast radius를 넓힌다 —
  기존에는 프로세스 제어뿐이었으나 이제 콘솔 수명을 넘는 config 영속이 가능하다. 수용
  근거: 실효 사슬이 기동 인터록(Verified 아니면 ReconcileDriver 생성 거부 — 편입 판정
  자체가 불가)과 콘솔 밖 §0.7 게이트 절차에 막혀 있고, 저장·기동 이중 audit이 남는다(확인 문구
  마찰은 사용자 결정으로 기각 — 잔여 위험은 review.md에 기록).

## Non-Goals

- 별도 편입 엔진, 즉시 journal 편입, 수동 매수 주문 감지용 신규 API, 정책 계산 복제.
  이번 개정은 기존 자동 편입 기능의 메뉴 발견 가능성과 설명만 바꾼다.
- **편입 해제 없음 유지**(adopt-external-positions design A5): include 제거·exclude 추가는
  이미 편입된 포지션에 아무 효과가 없다(화면에 명시). 보호 제거 UI는 만들지 않는다.
- automation_gate(운영 게이트)·Guardian 한도·kill switch의 콘솔 편집 — 금지 유지.
- 콘솔의 직접 편입(journal 쓰기)·주문 라우트 — 금지 유지(가드 테스트 존치).
- 엔진 config hot-reload.

## Capabilities

### Modified Capabilities

- `exit-policy`: "외부 취득 포지션의 자동 편입" — include_symbols 종목별 편입 조건 추가
- `operator-console`: "콘솔 안전 불변식" — 상태변경 행위에 편입 설정 저장·종목 편입 지정
  추가(주문·게이트·자격증명 라우트 부재 불변), "read-only 불변식" — 유일한 쓰기 예외로
  주입된 편입 설정 seam 명시(journal RO·브로커 조회 전용 불변), 설정 화면 요구사항 추가

## Impact

- Affected code: `internal/config`(Adoption.IncludeSymbols·validate·raw 블록 Load·외과적
  Save), `internal/app/engine/adoption.go`(judgeHoldings·alertUnmanaged — **High-risk:
  reconciliation 경로, Pre-Edit 선언·FLM 필수**), `internal/app/engine/interlock.go`
  (recordGateSettings에 include_symbols 항목 — FLM 대상), `internal/console`(설정
  화면·라우트·가드 테스트), `cmd/tossctl`(설정 seam·audit append 배선)
- 안전 검토(§0): 편입은 보호 추가 방향만(include는 편입 대상 확대 = 보호 확대, exclude
  우선 유지), 편입 tx 무발의(A2) 불변, 손절 즉시성 무접촉, zero-value 안전, 실효는 게이트
  ON 이후. 콘솔 신규 쓰기는 config 한정이며 계좌 접근 능력은 늘지 않는다
- PM: registry bootstrap allowlist 등재(활성 story 체계 부재 — archive 시 제거)
