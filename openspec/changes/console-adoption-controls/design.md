# Design: console-adoption-controls

> 개정 1(2026-07-27): proposal-freeze 적대적 리뷰 라운드 1(REVISE) 반영 — P1 3건(P1-1 raw
> Load, P1-2 include audit, P1-3 방어선 서술·확인 문구), P2 7건, P3 4건. 기록은 review.md.

## D1 — 실행 주체 불변, 그리고 방어선의 정직한 서술

콘솔에 편입 능력을 주지 않는다. 화면의 두 행위(설정 저장·종목 지정)는 전부
`engine.adoption` config 블록 편집이고, 실제 편입은 엔진 reconcile 루프의
judgeHoldings → adoptOne 경로만 수행한다. journal RO 가드·AdoptPosition 금지 가드·브로커
단일 read 인터페이스 가드는 전부 유지된다.

**방어선(리뷰 P1-3 정정)**: 방어는 "엔진이 없으면 아무 일도 없다"가 아니다 — 이 콘솔은
[엔진 시작/정지]도 갖고 있다. 실제 사슬은 ① automation gate는 콘솔에서 편집 불가(§0.7
콘솔 밖 절차) ② 엔진 기동 인터록: `AutomationStatus.Verified`(게이트 ON + 전 조건)가
아니면 ReconcileDriver가 **생성 자체를 거부**하므로 편입 판정·journal 편입 기록이 원리적
으로 불가 ③ 2c 전에는 게이트 상수 차단. 그래도 이 change는 **탈취된 세션의 blast
radius를 넓힌다**(프로세스 제어뿐이던 것이 콘솔 수명을 넘는 config 영속으로). 리뷰가
제안한 보상 통제(타이핑 확인 문구)는 **사용자 결정(2026-07-27)으로 기각** — 저장은
세션+CSRF+클릭만으로 성립한다. 수용 근거: 실효 사슬(게이트 콘솔 편집 불가·Verified
없으면 ReconcileDriver 생성 거부·2c 전 상수 차단)과 저장·기동 이중 audit이 방어의
본체이고, 잔여 위험(탈취 토큰 2개의 blind config flip)은 소유자가 인지하고 수용했다.

## D2 — include_symbols: opt-in의 표현

| 상태 | 판정 |
|---|---|
| enabled=true, 목록 무관 | exclude 제외 전부 편입 후보(현행) |
| enabled=false, include에 있음 | 그 심볼만 편입 후보 (신설) |
| include ∧ exclude 동시 | **exclude 승** — 편입 안 함, 무관리 알림 유지 |
| enabled=false, 목록 비어 있음 | 현행 그대로(무관리 알림만) — zero-value 안전 |

목록은 **시장 무관 심볼 단위**다(P3-11 — exclude의 기존 의미론 상속; 시장 한정 목록은
후속 스펙 변경 사안). validate 확장: "의미 있음" = `enabled ∨ include≠∅` — include만
있어도 합성 손절 분모가 필요하다. 위반 시 블록 전면 zeroing + Rejected(기존 규칙).
include/exclude는 normaliseSymbols 규칙 공유.

judgeHoldings의 게이트 **순서는 유지**(already-managed → RECONCILE → stale → findings)
하고 findings 분기만 확장한다 — include가 신선 조건·Stabiliser·Verified를 우회할 수 없는
구조적 이유다. **무관리 알림 사유 행렬(P2-6 — 전부 진실할 것)**:

| 상태 | why 문구 |
|---|---|
| 블록 Rejected | 설정 거부 사유 그대로(현행) |
| exclude (enabled·include 무관) | 의도적 제외(현행 — enabled 조건 제거) |
| enabled=true, 이번 주기 편입 실패 | 시도했으나 실패, 다음 주기 재시도(현행) |
| **include 심볼, 이번 주기 편입 실패** | **"지정되어 시도했으나 이번 주기 실패" — "꺼져 있다" 금지** |
| enabled=false ∧ 비include | 꺼져 있어 관찰만(현행) |

## D3 — 콘솔 표면

- `GET /settings`: adoption 블록 표시 + 편집 폼. **표시는 파일의 raw 블록**(P1-1): 주입
  seam의 Load는 config.json의 `engine.adoption` 키를 **merge·zeroing 이전의 원문**으로
  읽고, 검증 판정(거부 사유)은 별도 필드로 병기한다 — 거부된 블록의 exclude/include
  목록이 화면 왕복으로 유실되는 경로를 차단한다. automation_gate는 편집 불가 문구만.
- `POST /settings/save`: 서버측 재검증(무효면 거부·무기록). 확인 문구는 없다(D1 — 사용자
  결정). 저장 응답은 반영 시점과 **현재 엔진 실행
  여부**(엔진 마커 — 실행 중이면 "가동 중 엔진은 이전 설정으로 동작, 재시작 필요")를
  말한다(P3-13).
- `POST /settings/include`: `symbol` 하나를 include에 추가(멱등). **선행조건 미충족 시
  거부**(P2-5): 저장될 블록이 검증을 통과하지 못하면(stop pct 미설정 등) config를 건드리지
  않고 "설정 화면에서 손절폭을 먼저" 안내한다 — 엔진이 zeroing할 블록을 쓰는 것은 최악
  경로이므로 금지. 응답은 지정이 **상시 규칙**임을 말한다(P2-10 — CLOSED 후 재매수도
  재편입, 제거는 장래 편입에만 영향).
- 두 POST 모두 `session0(mutating(...))`. `consoleStateChanging`·CSRF 가드 목록 추가 +
  actVerbs에 config-쓰기 어휘 추가(P2-7 — 허용 목록 밖 신규 라우트가 침묵 통과하지 않게).
- 저장 seam: `AdoptionSettings { Load() (config.Adoption, string, error); Save(config.
  Adoption) error }` — Load의 둘째 반환은 검증 거부 사유("" = 유효). cmd/tossctl 구현:
  config.json을 raw map으로 읽어 `engine.adoption` 키만 교체, **유일 임시파일
  (os.CreateTemp) + flock 하의 read-modify-rename**(P2-4 — 동시 기록 lost-update 방지),
  파싱 불가 파일이면 저장 거부(골격 생성은 os.ErrNotExist에 한정), Save 성공 시 **audit
  로그에 자체 엔트리 append**(P2-9 — 엔진 기동 diff의 flip 붕괴 한계 보완; audit.Log는
  데이터 디렉터리 O_APPEND, 콘솔 패키지 밖이라 무기록 가드 무접촉).
- 좁힘 가드(P2-8): ① AdoptionSettings가 정확히 Load·Save 두 메서드만 선언(AST 테스트)
  ② internal/console은 config.Service/NewService를 명명하지 않음(정적 테스트) ③
  cmd/tossctl 테스트: Save 후 `engine.adoption` 밖의 모든 키가 **바이트 동일**.

## D4 — 문구·가드의 갱신

- positions의 원장 부재 공지: "이 화면의 기능이 아니다" → "편입 실행은 엔진 대사 루프의
  몫이고, 이 화면은 지정만 한다".
- 개정 가드: consoleStateChanging + CSRF 목록 + actVerbs 확장 + 전사 문장,
  TestAnUnmanagedHoldingIsLabelledExactlyOnce는 "직접 수행 금지" 유지·지정 폼 허용으로.
  계좌 동사 금지·journal 쓰기 금지·브로커 인터페이스·AdoptPosition 금지 가드는 무개정.
- audit(P1-2): `recordGateSettings`에 `engine.adoption.include_symbols` 항목 추가
  (interlock.go — 기존 함수 수정, FLM 대상). 콘솔 seam의 자체 audit과 이중 기록이 되는
  것은 의도(§0.5 — 저장 시점과 실효 시점 둘 다 남는다).

## D5 — 반영 시점과 표시

엔진은 기동 시 config를 읽는다(hot-reload·watcher 없음 — 코드 사실로 확인됨). 설정
화면·지정 응답의 고정 문구: "다음 엔진 기동부터 반영". 역방향 위험도 같은 문구가 덮는다
— **가동 중 엔진에 exclude 추가·enabled=false 저장은 즉시 멈추지 않는다**(기동 스냅샷
동작) — 저장 응답의 실행 중 안내(P3-13)가 이를 명시한다. include 제거·exclude 추가가
이미 편입된 포지션에 무효라는 문구(A5 귀결)도 설정 화면에 넣는다.

## D6 — 위험 등급

High-risk: `judgeHoldings`·`alertUnmanaged`(reconciliation)·`recordGateSettings`(audit)
기존 함수 수정 — Pre-Edit 선언 + FLM 무면제 + 적대적 Eng 리뷰 + race 테스트. 콘솔
불변식 개정 포함으로 change 전체에 적대적 관점 적용(라운드 1 완료 — review.md).

## D7 — 외부 종목 자동관리 메뉴 (요구사항 개정 2026-07-31)

기능을 복제하거나 새 route/handler를 만들지 않는다. 상단 navigation의 기존 `/settings`
링크를 `/settings#adoption`으로 좁히고 표시명을 **외부 종목 자동관리**로 바꾼다.
`settings` template의 첫 adoption section에 `id="adoption"`을 부여하므로 링크는 현재
서버측 렌더링·저장 seam·CSRF 경로를 그대로 사용한다.

화면 제목은 **외부 종목 자동관리 설정**으로 하고, 설명은 다음 경계를 분명히 한다.

- 사용자가 수동 매수한 무기록 보유가 대상이다.
- 자동 편입 ON 또는 종목별 지정이 후보를 결정하며 제외가 항상 우선한다.
- 편입 완료 후에는 현재 공통 익절·보호선·손익 극대화 정책이 적용된다.
- 저장은 편입이나 주문을 즉시 실행하지 않으며 다음 엔진 기동의 대사 루프가 수행한다.

변경 표면은 template 상수와 문자열 단언 테스트뿐이다. config, journal, reconcile,
exit-policy 계산, route 분류, CSRF, 주문·LIVE 권한은 수정하지 않는다. 기존 함수 내부
로직을 바꾸지 않으므로 이번 개정의 Function Logic Map은 not-applicable이다.
