# a176 design

전제: a173 착지 뒤(a174·a175 와 병행 가능 — 파일: `templates_overview.go`, `templates.go` 의 `verify` 템플릿(:762 부근), `pages.go` 는 편집하지 않는다). file:line 은 정찰(2026-10-25) 값.

## 1. 지형
| 대상 | 위치 | 비고 |
|---|---|---|
| 개요 | `templates_overview.go:30` `overview`, 요약 스트립 `:56-72`, 계좌 표 `:131-142`, 오늘 표 `:175-183`, 살아 있는 주문 `:204`, 안전·잔여물·Guardian·소진분, 최근 exit `:290` | a173 이 `causes` 배너·`unmeasured` 두 모드·기술 정보를 넣음. 스펙 704~718·1372~1390 |
| 개요 뷰 모델 | `overview.go` `accountPanelFrom`·`todayPanelFrom`·`recentPanelFrom`, 사유 열거 `:103-141` | 핸들러·패널 함수 본문 편집 0 — 요약 재배치는 템플릿 + 새 템플릿 함수로 |
| 검증 화면 | `templates.go:762` `verify`(단계 목록 `<pre>` 덤프 `:916` 근처 `--list 와 같은 내용`), `/verify/start·approve·abort` 폼, 확인 문자열 | 스펙 1326·1344~1370: 시장 선택·단계 목록·승인 채널 기록 유지 |
| 단계 데이터 | `pages.go:158` `handleVerify` 가 넘기는 단계 목록(문자열 덤프인지 구조체인지 1.1 에서 확정) | 구조체가 아니면 **텍스트를 파싱하지 않고** 기존 `verifylive` 계획 구조체를 뷰로 넘기는 새 함수 추가(핸들러 본문 편집 없이 템플릿 함수로) |
| 시험 | `overview_test.go`(한글 복사 22줄, 여섯 사실·여덟 사유 단언), `screen_paths_test.go`, a173 계약 시험, verify 관련 `pages`·`verify_*_test.go`(확인 문자열·CSRF·시장 선택) | |

## 2. 설계
### 2.1 개요 요약 세 묶음 (여섯 사실 유지)
- 상단 요약을 세 카드로: **엔진·보호**(① 실행 여부+기동 거부 사유, ⑤ 중 안전 상태) / **오늘·한도**(③ 시장별 실현손익·왕복·승패, ⑤ Guardian 한도 중 읽을 수 있는 값, 소진분 한 축) / **예외**(② 보유 종목 수·관리 외, ④ 살아 있는 주문, ⑥ 최근 exit, ⑤ 잔여물). 각 칸은 사실 1 + 상세 링크(spec 704). 재계산 0 — 기존 패널 값 그대로.
- 1280 × 900 에서 요약 세 카드가 첫 화면 안(≤ 900px). 상세 섹션은 그대로 아래에(spec 704 "제거되지 않는다").
- "구조적으로 미측정 — …" 설명은 소진분 섹션 머리에 1회, 각 축 칸은 `미측정 · structural`(사유 코드 — 열거에 없으면 **기존 여덟 사유 중** 가장 가까운 것을 쓰거나 Manager 와 상의; 새 코드를 만들지 않는다).
- 검증 기록 경로는 a173 의 `details.tech`.
### 2.2 `/verify` 단계 목록 구조화
- 단계마다 `<details class="step"><summary><span class="step-name">…</span> <span class="status-pill">mutating|read-only</span> <span class="muted">한 줄 요약</span></summary><pre>…종전 텍스트…</pre></details>`. 전체 텍스트가 DOM 에 남아 `tossctl verify run --list` 와 같은 내용(스펙 델타).
- 시작·승인·중단 폼·CSRF·확인 문자열 절차·시장 선택 링크는 마크업 바이트 불변(렌더 비교 시험).
- 접힌 상태에서도 단계 제목·mutating 배지·요약은 보여 한 화면에서 14단계를 훑을 수 있게.

## 3. 측정
a173 하네스. 기준선: `/dashboard` 4458px(a173 after), `/verify` 5188px. 목표: 개요 요약 ≤ 900px(요약 영역 하단 좌표), `/verify` 접힌 상태 ≤ 1800px, 설명 중복 0, 폼 바이트 동일.

## 4. 안전
`/verify` 는 승인 창이 있는 화면: 폼·버튼·확인 문자열·CSRF 는 건드리지 않는다. 개요는 폼 0(spec 1374). 브로커 호출·새 상수 0.
