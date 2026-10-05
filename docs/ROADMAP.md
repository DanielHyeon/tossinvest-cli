# TossOS 구현 로드맵 (전체 작업 분할)

> 작성: 2026-07-26 · 총괄 아키텍트(Manager) 관리 문서 · SDD(OpenSpec) 기반
> 구현·테스트는 구현 에이전트(Teammate)에 위임, 리뷰 게이트는 docs/WORKFLOW.md
> 개정 v2: gstack 4보이스 리뷰(66건) 반영 — 기록: openspec/changes/add-tossos-foundation/review.md

## 제품 정의

TossOS는 `tossinvest-cli`(MIT, 커밋 `57348a7` 고정) 전체 소스·히스토리를 기반으로 한 **독립 product fork**다.
- 보존: CLI(tossctl)·MCP·공식 Open API·WTS·인증·주문·조건주문·진단 기능 전부 (CLI = 관리자 콘솔)
- 추가: 자동매매 엔진, 위험관리, 포지션/원장, 웹 API(SSE), 운영 콘솔 UI
- StockOS에서는 React UI 자산(디자인 시스템·안전 UX)과 검증된 거래 불변조건·순수 로직만 선별 이식
- 제외: KIS 연동, StockOS의 shadow/observe/canary 구현체와 runtime flag 176필드 체계, A-넘버 shim, **승격 단계(paper/capped-live) 일체 — 사용자 결정(2026-07-26): 단독 사용자 제품이므로 실전 직행**

## 운영 파라미터 (선택 — 사용자 확정 시 기록)

- 운용 시장(KR/US), 허용 자본 상한, 일일 최대 손실(절대액·%), 레인당 위험 예산
- 위 수치는 Guardian(T2.4) 설정값이 된다. 미확정 시 StockOS small_live 보수 기본값(주문 100만/노출 1,000만/일손실 10만 KRW 또는 1%)으로 시작

## 권위(Authority) 원칙

1. 주문 실행 권위: 토스 공식 Open API만. 엔진 배선은 official-only 브로커로 고정하고 **WTS 쓰기 경로가 도달 불가함을 테스트로 증명** (hybrid 라우팅 누수 차단)
2. 조회/신호: WTS는 조회 전용
3. 포지션 최종 권위: 토스 계좌. 로컬 원장은 파생 상태
4. 불일치 시 신규 진입 금지, 청산 지속
5. **무인 운영**: 핵심 매매 루프(체결 감지 포함)는 공식 API만으로 동작. **체결 감지의 권위는 공식 API 주기 폴링**(최대 신선도 SLO 명시)이고 SSE는 지연 단축용 힌트일 뿐이다. WTS 만료 시 후보 소스는 공식 API 기반(랭킹·watchlist·정적 유니버스)으로 강등되어 매매 루프가 유지된다
6. kill switch는 신규 진입 차단 전용(BLOCK-ONLY)이되, **운영 모드 체계는 별도로 존재**: NORMAL / ENTRY_BLOCKED(=kill switch) / HALT_ALL(노출 증가 중단 — EXIT_ONLY는 2d 리뷰에서 ENTRY_BLOCKED와 행동 동일로 삭제, 2c가 실제 차이 생기면 재도입). 수동 비상 청산(flatten-all)은 typed-confirmation 수동 명령으로 제공
7. **실전 직행(사용자 결정 2026-07-26)**: 승격 단계(paper/capped) 없이 실전 매매로 바로 진행한다. 안전은 단계가 아니라 **위험 한도**로 담보한다 — Guardian(일일 손실·총 노출·수량 한도)과 kill switch가 활성화되지 않으면 엔진이 기동하지 않는다(T1.9 인터록)
8. 주문 mutation은 **자동 재시도 금지**. 타임아웃·5xx 등 결과 불명(unknown outcome)은 IN_DOUBT로 표기하고 체결/거래내역 조회로 확정될 때까지 해당 레인 차단
9. 계좌당 주문 writer는 데몬 하나. CLI/MCP의 수동 주문은 데몬 경유 또는 명시적 maintenance mode에서만, reconciliation은 외부 주문을 별도 provenance로 격리

## Phase 0 — Foundation & Baseline  (change: `add-tossos-foundation`) — 진행 중

| ID | 작업 | 담당 |
|----|------|------|
| T0.1 | upstream 클론·`upstream` remote·커밋 고정 — 완료 | Teammate |
| T0.2 | build/vet/test 베이스라인(650 green) 기록 — docs/baseline.md | Teammate |
| T0.3 | OpenSpec 스캐폴딩·로드맵·워크플로 규칙 | Manager |
| T0.4 | Makefile 타겟(vet/cover/validate/gate), .gitignore 보강, gate.sh | Teammate |
| T0.5 | StockOS 인벤토리 문서화 | Manager |
| T0.6 | 발견성·스코프 정리: 루트 CLAUDE.md, AGENTS.md 스코프 헤더, openspec/project.md | Manager |
| T0.7 | upstream push URL 차단, LICENSE 보존 확인 | Teammate |

## Phase 1 — Execution Base Hardening  (change: `harden-execution-base`)

목표: 전략 손실과 주문 시스템 결함을 분리할 수 있는 실행 계층. **관측성·알림 포함** — 알림 없는 무인 엔진은 무인이 아니라 무감독이다.

| ID | 작업 | 근거 |
|----|------|------|
| T1.1 | `internal/app` 신설: `newAppContext` 승격. **위임 shim 전략** — cmd/tossctl에는 동명 얇은 래퍼를 남겨 upstream diff 최소화. 엔진 프로필은 official-only 브로커 강제 + WTS 쓰기 도달 불가 테스트 | 배선 병목 + 원칙 1 |
| T1.2 | `trading.MutationResult` → `internal/domain` 이동, 기존 위치에 type alias 유지 (upstream 호환) | 역방향 의존 제거 |
| T1.3 | 조건주문 게이트·intent 조립 `internal/trading` 이동 (shim 유지) | 엔진·ops/MCP 접근 |
| T1.4 | 주문 상태기계 코드화 — **공식 Open API 주문 status 스키마가 정본**, 응답 fixture 계약 테스트 동반. docs/trading/order-state-machine.md(WTS 관찰 기록)는 보조 증거로만. IN_DOUBT(제출 불명)·AMEND_IN_DOUBT 상태 포함 | 정본 소스 교정 |
| T1.5 | **durable intent journal(WAL)**: 주문 제출 전 의도를 영속 기록(단일 테이블 SQLite 또는 append-only JSONL). 멱등성·재시작 복구의 선행 조건. 저장 위치는 ext4 경로(XDG data dir), fuseblk 거부 가드 | P2 원장 숨은 의존 해소 |
| T1.6 | 체결 감지: 공식 API 주기 폴링(pending+체결내역)을 권위로, 신선도 SLO 정의. SSE는 힌트 — 토픽별 coalescing/single-flight, 최대 강제 갱신 간격 | SSE 권위 역전 차단 |
| T1.7 | retry matrix: endpoint×방향별 정책 — mutation 자동 재시도 금지, 조회는 bounded jitter, 필수 상태 조회 staleness 초과 시 신규 진입 차단 | 안전 방향 구분 |
| T1.8 | 재시작 reconciliation **계약 명세**: 비교 키, 허용 오차, 안정화 시간, 외부 수동 주문 분류, 충돌 규칙, 영구 불일치 시 운영자 절차 | 계약 부재 |
| T1.9 | 자동화 게이트 **설계만** (활성화는 Phase 2): 게이트는 기본 OFF, 위험 엔진(Guardian) 미주입 시 기동 거부하는 boot-time assertion. interactive auth challenge는 fail-closed 거부 + 알림 | 순서 결함 교정 |
| T1.10 | 관측성·알림: 구조화 로깅, 핵심 메트릭, push 알림 채널(예: ntfy/Telegram) — kill switch 발동·reconcile 불일치·주문 거부·세션 만료·데몬 크래시 시 통지 | Phase 6에서 이동 |
| T1.11 | `tossctl` flatten-all 명령(수동 전용, typed-confirmation): 미체결 전체 취소 + 전량 청산 | 비상 프리미티브 |
| T1.12 | 공식 API capability 검증: 자격증명 무인 갱신 N일 soak, rate limit 실측, 주문·체결 데이터 완전성 확인. FX(USD 매수 시 통화 잔고) 경로 정의 — 부족 시 fail-closed | 전제 실증 |
| T1.13 | 시간 규율: 주입 가능한 clock, 시장별 TZ, 거래일 경계(일일 한도 리셋 기준), DST 테스트 | 위험관리 선행 조건 |
| T1.14 | 실계좌 주문 경로 검증(사용자 실행, 1회성): 최소 수량·limit-only·즉시 취소 규칙으로 매도 경계(부분/전량/보유초과)·KR cancel/amend 확인. 승격 단계가 아니라 **실행 기반을 닫는 검증** — 엔진이 쓸 주문 경로의 미검증 갭 해소 | upstream 갭 |
| T1.15 | 토스 Open API 약관·자동화 허용 범위 검토 기록(사용자 협조 필요), 계정 정지 시 포지션 처리 방침 | 브로커 단일 의존 리스크 |

## Phase 2 — 실행 계약 확장 + Core Domain  (change: `extend-execution-contract` → `add-core-domain`)

**2026-07-26 재분할(3차)**: 리뷰 3라운드 108건(`openspec/changes/*/review.md`)의 결론 — 실패의 축은 "레일 vs 판단"이 아니라 **측정 의존성**이었다. 측정되지 않은 브로커 동작 위에 쓴 스펙이 라운드마다 무너졌으므로, 경계를 측정 의존성에 맞춘다.

- **2a `extend-execution-contract`** ✅ GATE PASS·archive (2026-07-26): 결정 계약(safety class·RiskIntent preimage·generation), 멱등키 기계(브로커 `clientOrderId` — P1 "멱등키 없음" 전제는 사실 오류였다), 진입 측 위험 예약, RECONCILE 상태, 한도 fail-closed·총계 계산 계약, opaque 식별자·OrderStatus 10개 정정, 엔진 Gateway 배선·봉인. **조건주문 코드 0줄.**
- **2b `verify-execution-capability`** ⏳ (사용자 측정 — 유일한 임계 경로): 멱등키 실동작·TTL 마진, 조건주문 속성(SINGLE+MARKET 손절, OCO/OTO는 LIMIT 전용), CANCEL_REJECTED 레코드 형태, 매도가능수량 의미, 실측 비용표.
- **2c `add-protection-orders`** (2b 결과 위에서 작성 — 미작성; 완료 시 ProtectionReady flip→실전 가능): 조건주문 형제 수명주기, 발동 주문 다리(`triggeredOrderId`), 미귀속 관측 격리 원장, 청산 수량 예약, PROTECTION_WEAKENING, flatten의 조건주문 취소. 기본 가설: 보호 = SINGLE+MARKET 손절 단독, 익절은 로컬 청산("한 심볼에 브로커측 매도 청구권 1개").
- **2d `add-core-domain`** ✅ GATE PASS·archive (2026-07-26, 2358 테스트): 비용 모델(KIS 수치 이식 금지, 2b 실측), Guardian 판정 체인(구조적 RR·등급배수는 P3 이관), 운영 모드(모드×클래스 표), 포지션 원장, provenance, 성과, tracer, **exit 정책(StockOS baseline ratchet·profit ladder 이식 — 기준선 단조 상승 손익 극대화, 순수 판정만; 액추에이션은 2c)**.

채택 원칙(StockOS 실행 무결성 분석 → 토스 계약으로 재구현): 불확실성은 상태로 격리, 불일치는 RECONCILE로 중단, 결정은 영속 후 실행, 브로커가 보증하지 않는 것은 타입으로 표시. MFE/MAE는 데이터 소스 부재로 P3 이관.

StockOS 순수 로직 이식 순위: costs → structural_rr → tradeplan/contract → common_admission → guardian 판정 순서 → in_flight_lifecycle → backtest/metrics. (slot_budget·capital_stage·LLM 게이트는 미채택)
**이식 상수 규칙**: 모든 정책 수치는 출처·적용 시장·검증 상태를 주석으로 기록하고, Toss 데이터로 검증 전까지 보수적 기본값(비용은 과대 추정)으로 표시한다.

| ID | 작업 | 비고 |
|----|------|------|
| T2.1 | internal/ledger: SQLite 영속 원장. **ext4 경로(XDG data dir) 강제 + fuseblk 기동 거부**, 단일 리더 락, WAL/fsync 정책, migration rollback, 백업·복구 시험 | NTFS 위험 차단 |
| T2.2 | internal/position: 포지션 상태기계. **aggregate 경계 선행 설계**: Order/Fill/Position/ProtectionSaga의 권위와 이벤트 흐름을 문서화해 T1.4 주문 상태기계·StockOS 11-상태 기계와의 중복을 명시적으로 매핑 | 이중 상태기계 방지 |
| T2.3 | internal/execution: 진입 체결→보호주문 완료까지의 **durable saga**(stop-first, 노출 SLA, 재시작 복구, 부분체결별 보호 수량 조정, oversell 방지) + fault-injection 테스트 | OCO 장애 창 |
| T2.4 | internal/risk: Guardian 판정 체인(순서·수치 보존, env 결합 제거), kill switch + 운영 모드 체계(원칙 6), 일일 손실 한도, 총 개방 위험 한도, 구조적 손절·위험 기반 수량·최소 RR. **자동화 게이트 활성화는 이 change에서** — Guardian 없이 기동 불가 인터록 포함. 고액 주문 flag는 청산 주문에 한해 상한부 허용 | T1.9와 세트 |
| T2.5 | 거래 비용 모델 (StockOS costs·cost_model 이식) | |
| T2.6 | provenance: 후보→판단→주문→체결→청산 lineage | |
| T2.7 | 성과 원시 지표: R 배수, PF, MDD, 승률 + MFE/MAE 신규 설계 | |
| T2.8 | **tracer slice(수직 슬라이스, 실전 소액)**: 하드코딩 레인 1개·종목 1개·limit 주문·최소 수량으로 신호→위험판정→preview→실체결→원장→reconcile→상태 조회 end-to-end 관통. Guardian 한도 활성 상태에서 실행 | 통합 피드백 조기화 (paper 단계 없음 — 사용자 결정) |

## Phase 3 — Strategy & Scheduling  (change: `add-strategy-engine`)

**착수 조건**: T1.14 주문 경로 검증 + T2.8 tracer slice end-to-end 검증 완료 (실행 기반 신뢰 확보 — 승격 단계 아님).

**T3.1은 이 조건에서 분리한다(2026-07-28).** 발굴은 **읽기 전용**이고 주문 경로를 건드리지
않으므로, 주문 실행 신뢰를 착수 조건으로 걸 이유가 없다. 오히려 반대다 — 발굴은 시간축
데이터를 쌓아야 쓸모가 생기고, 쌓기 시작하는 시점이 늦을수록 레인이 붙을 때 근거가 없다.
T3.1의 착수 조건은 **없다**(별도 change `add-candidate-discovery`, 격리 테스트로 주문 경로
의존 부재를 고정).

| ID | 작업 | 비고 |
|----|------|------|
| T3.1 | internal/candidate: 후보 수집 + CandidateEvidence. WTS 만료 시 공식 API 소스로 강등해도 후보가 계속 나오는 구성 명시 | 원칙 5 · **별도 change `add-candidate-discovery`로 분리, 착수 조건 없음(읽기 전용)** |
| T3.2 | internal/strategy: 독립 매수 레인 인터페이스, 레인별 ON/OFF, OFF 후 청산 지속 | |
| T3.3 | internal/scheduler: 장 시간 인지 루프 (T1.13 clock/calendar 기반) | |
| T3.4 | internal/performance: 레인별 성과(결정적 링크 없으면 표시 금지), markout 윈도우(기본 5/15/30분), 비용 후 기대값 기반 슬롯 배분 | |
| T3.5 | 실전 운영 개시: Guardian 한도·kill switch·알림 활성 확인 후 레인 ON (승격 단계 없음 — 사용자 결정) | 원칙 7 |

## Phase 4 — Service Layer  (change: `add-httpapi-daemon`)

| ID | 작업 | 비고 |
|----|------|------|
| T4.1 | cmd/tossosd: 데몬 부팅·config·graceful shutdown (internal/app 공유). 계좌당 단일 주문 writer 보장 | 원칙 9 |
| T4.2 | internal/httpapi: REST + SSE(sequence id 순서 보장·스키마 버전), 로컬 토큰 인증 | |
| T4.3 | 운영 엔드포인트: 상태, 운영 모드 전환(확인 모달), 레인 제어, reconciliation 강제 실행 | |

## Phase 5 — 운영 콘솔 → 풀 UI  (change: `add-ops-console`, `add-web-ui`)

리뷰 결정: UI를 2단계로 분할. 5a는 안전 운영에 필요한 최소 표면, 5b는 capped-live 성공 후.

**5a — 운영 콘솔 (`add-ops-console`)**: 상태·포지션·미체결·차단 사유·reconciliation 상태·운영 모드/kill switch(확인 모달)·실현손익 요약. style.css 디자인 시스템 + 안전 UX(사유 입력·차단 칩) + useDashboardStream 이식. 차트 없음.

**확인 모달의 계약**: UI에서 되돌리기 어려운 조작(운영 모드 전환, kill switch, reconciliation
강제 실행)은 **수행 내용을 문장으로 보여주는 메시지 박스 + 확인 버튼**으로 승인한다.
확인 문구 타이핑은 **금지**한다(사용자 지시 2026-07-27). 마찰이 아니라 **무엇이 일어나는지를
보여주는 것**이 이 모달의 안전 기여분이므로, 모달은 대상·현재 상태·전환 후 상태·되돌리는
방법을 명시해야 한다. 기본 포커스는 취소이고 확인은 명시적 클릭이다.
CLI(`tossctl flatten-all`, `tossctl verify run`)는 메시지 박스가 없으므로 기존 TTY
typed-confirmation을 유지한다 — 이 결정은 UI 표면에만 적용한다.

**5b — 풀 UI (`add-web-ui`, 실전 운영 안정화 후)**: 후보 랭킹·레인 성과·분석 화면·차트(경량 라이브러리 채택 우선 평가, 그린필드는 최후). API client는 TossOS 계약 신규 작성(TanStack Query 도입 여부 이 change에서 결정).

폐기 유지: canary/shadow 섹션, A-넘버 컴포넌트, 레거시 탭 shim, KIS 카드류.

## Phase 6 — Ops 잔여  (change: `add-ops-runbooks`)

| ID | 작업 |
|----|------|
| T6.1 | 장애 runbook 확장: WTS 세션 만료·WTS 계약 변경·reconcile 영구 불일치·App-Version 파손 (기본 알림·패닉 절차는 T1.10~T1.11에서 이미 확보) |
| T6.2 | upstream sync 운영화: upstream-sync 브랜치 + docs/upstream-sync-log.md 기록 |

## Phase 게이트 (공통)

- `make gate CHANGE=<id>` 통과 (tasks 완료 + review.md 존재 + test/vet/validate)
- Manager diff 리뷰 + 독립 테스트 재실행, upstream 테스트 회귀 없음
- 안전 경로(주문·위험·원장) change는 추가로: race 테스트(`go test -race`), crash/restart 시나리오, 중복·역순 이벤트 테스트 중 해당 항목
- 핵심 4패키지 커버리지 하한 유지: trading 74, orderintent 75, orderlineage 74, client 70 (%)
- `openspec archive` → specs/ 확정 반영

## upstream 동기화 정책

- 반영 경로: `upstream-sync` 브랜치에서만 선별 merge → 검증 후 main. 내역은 docs/upstream-sync-log.md에 기록
- 대상: 보안 수정, 공식 API 스펙 변경, WTS endpoint·파서 수정, 인증·세션 수정, 조건주문 수정, probe 개선
- 비대상: CLI 출력 포맷, 문서·설치기·웹사이트. 충돌 시 TossOS 실행 경로 우선
- Phase 1 리팩터는 shim 전략(T1.1~T1.3)으로 upstream 파일 원형을 최대한 보존한다

## 저장소 정책

- 비공개 독립 저장소. `upstream` = JungHoonGhae/tossinvest-cli (push URL DISABLED), `origin` = 사용자 비공개 저장소(사용자 생성)
- MIT LICENSE·원저작권 고지 유지, 시크릿·세션·DB(sidecar 포함) 커밋 금지
- NTFS 마운트: `core.filemode=false` 필수. **영속 데이터(원장·journal)는 저장소 밖 ext4 경로에 둔다**
- Go module 경로: 유지 (사용자 결정 대기 항목 — review.md의 미결정 사항 참조)

## a092 이월 · 미배정 후속 (a092 archive 뒤에도 남는 의무 — 2026-09-30)

`a092-an-alert-does-not-hold-the-stop` 이 넘긴 것. 소유 change 가 아직 없는 항목은 「미배정」이다. 각 항에 **그것에 기대는 a092 계약**과 행선을 적는다. 근거 정본은 a092 `review.md` §24.11 · §24.12 이월 표와 `tasks.md` 「미배정 후속으로 이관한 것」이다.

| 항목 | 기대는 a092 계약 | 행선 |
|---|---|---|
| `runtime.go` `Runtime.escalate` → `EscalateOperatingMode` 무기한 대기(3라운드 H2(a)) | 감독자 승격은 동기 통지자 경로(21판 Q1 문자 해석으로 범위 밖) | 미배정 |
| 엔진 종료 순서 · 종료 중 완화 통지(떨어진 ctx 에 기한 없음, 26라운드 보이스 A N1) | mode-release 는 커밋 뒤 통지 · 재읽기를 요청 ctx 에서 뗌 | 종료 순서 후속 로트 — 완화 처리기를 원장 close 앞에 기다리기 |
| `operating_modes` 최신 행 읽기의 temp B-tree 정렬(`ORDER BY rowid`, 무한 성장 — EXPLAIN 영수증 `analysis/harness/explain-operating-modes-order.txt`) | 모드 커밋 순서 = rowid(C5 · K14) | v33 마이그레이션 후보 — 스키마 핀 27 · 레인 활성화 순서와 함께 |
| mode-release 통지 확인이 PENDING 전체를 읽음 → `event_key` 점조회 | 완화 결과는 다시 읽은 값(K16 · M12) | 후속 로트(성능) |
| `Acknowledge` 의 N 행 autocommit(`n.mu` 아래) | 기록과 승인 셈~해제가 같은 `n.mu`(원칙) | 후속 로트(성능) |
| `Flush` 결과 분류 · 차단 · 전송 위 `n.mu`(생산 호출자 0, K19 핀) | 생산 배달은 a098/a124 실행자 | 생산 연결 전 수리 |
| `ClaimDisposition` 미지 값 처리 | 원장이 세 값만 반환(잠재) | 후속 로트(결과 분류 명시화) |
| exit 배선 Floor 존재 검사(`if opts.Floor == nil`) · k3 핀이 호출 모양만 셈 · 이름 기반 구조 핀이 메서드 값을 못 봄 | k3 합성(floor 조회 401 은 기록 전용) · 입구 도달 경로 핀 | 후속 로트(구조 핀 계측기 — `*types.Func` 참조 세기) |
| 제어 서버 · 클라이언트 사본(알림 제어 복제) · 기록 실패 처리 두 사본 · 승격 문구 판정 네 자리 · `claimAndDeliver` 수동 Unlock | 동작 불변 리팩터 | 후속 로트(리팩터) |
| `Notifier` 에 `Close`/`Stop` 없음 · `flatten.Saga.Notifier` nil 지뢰(M6 핀이 조립 사실만 고정) · `o.Interval()` drift | exit 기록 전용(21판) 은 이 셋에 기대지 않음 | 미배정 |
| 계좌 원문 로그 base 관행(약 20곳) · 동기 `claimAndDeliver` 게이트 설명의 원문 오류 | 불변식 8 (a) 는 a092 새 표면만 닫음 | 사람 결정 큐 「계좌 가림 설계」 |
| `check_values.py` · `coverage_gate.py` 개선(옛 a092 10.4.2~10.4.6) | 없음(문서 도구) | a092 문서 도구 후속 로트(미배정) |

## a095 이월 · 미배정 후속 (a095 archive 뒤에도 남는 의무 — 2026-09-30)

`a095-a-stop-must-know-what-it-covers` 가 넘긴 것. 근거는 a095 `review.md` §4.7 · §4.8 · §4.9 와 `design.md` D8 · `issues.md`.

| 항목 | 기대는 a095 계약 | 행선 |
|---|---|---|
| 배포 전 `main` 과 SchemaVersion 대조 · 배포 직전 원장 읽기 전용 재측정 · 배포 후 `alert_outbox` 에 편입 실패 critical 행 확인(tasks 7.1 · 7.2 · 7.4) | 공시 문안 §4.9 | 사람 항목 — 배포 승인 때 |
| 대사 goroutine 의 동기 배달 비용 — transport 사망 시 편입 실패 보유당 사이클마다 ≈34s(계산값, 상한 54s) × N(design D8, Q7 둘째 면) | critical 은 메모리 래치를 지남(6판 원칙) · 정본은 이 대기를 근거로 인용 금지(SHALL NOT) | 미배정 — 후보: PENDING 재배달을 배달 실행자에게만 |
| 수량 증가 래치가 「보고한 최대 수량」이라 감소(부분 익절 · 부분 매도) 뒤의 새 증가를 그 최대를 넘기 전까지 놓침 | 델타 「새 최대 수량마다 보고」 | 미배정 — 체결로 설명되지 않는 순증 기준 래치(델타 문언 변경) |
| 편입 커밋 뒤 보호 미개설 보유의 알림 · 오해 부르는 성공 문구(`issues.md` I7) | 범주 ③ 은 critical 요구 밖(이름 붙은 경계) | 미배정 |
| 총위험 보고(R3, `issues.md` I2) · 불타기 래칫 선행 조건(Q6, `issues.md` I1) · 운영자 재편입 하향의 승인 · audit(`issues.md` I6) | 없음(보류 · 후속) | 미배정 |

## a094 이월 · 미배정 후속 (a094 archive 뒤에도 남는 의무 — 2026-10-01)

`a094-a-stop-clears-what-blocks-it` 가 넘긴 것. 근거는 a094 `review.md` 「구현 로트」 §7~§11 · `design.md` D−2.4 · D−3.4 · D−4.2 · D−9.3 · D−9.4 · `issues.md` I3 · I4 · I5.

| 항목 | 사유 | 행선 |
|---|---|---|
| 배포 전 `main` 과 SchemaVersion 대조 · 엔진 재시작 승인(롤백 원칙 8.2) · 배포 전 운영 원장의 ACKED 행 · intent 없는 무장 발의 행을 읽기 전용으로 셈 · 배포 뒤 첫 409 사건 실물 확인 · 얼어붙은 세 포지션(475150 · 080220 · 272210) 처리(tasks 8.1~8.4) | 스키마 변경 0 · 엔진 정지 = 손절 없음 · 이 change 는 소급 보호하지 않음(I4) | 사람 항목 — 배포 승인 때 |
| 엔진 밖 주문(사람이 낸 반대 주문)의 취소 확대(tasks 3.X1 · D−2.4 선행 조건 넷). 그 전에는 새 409 의 결말이 「사람이 반대 주문을 치울 때까지 손절 보류 · 매 주기 거절 보고」 (I3) | 사용자 결정 대기 | 사람 결정 큐 |
| 판정 불가(시세 없음) 포지션의 종결 대기 알림을 a090 미관측 경보가 덮는지의 통합 시험(tasks 3.R9a, D−9.3) | a090 착지 뒤 교차 시험 | a090 착지 뒤 — a090 또는 후속 로트 |
| R1 소급 재분류(park 된 옛 409 행, tasks 4bis) | 5판 D−3.4 이연(YAGNI — 사건 두 행은 이미 park, 해동 명령으로 일원화) | 미배정 — 선택 후속 |
| 나머지 ACKED 정산 — 해소기 matcher 의 주문 번호 판별자 · 단일 일치 덮어쓰기 반례(4.N4x, D−4.2) | 기동은 기록 번호 바이트 일치만 확정 | 미배정 — 후속 change 후보 |
| UNKNOWN_BROKER_STATE 손 해소 명령 부재(D−9.4) — 종결 증거 대기 알림 본문이 「손 해소 수단은 아직 없다」 고 말함 | 명령이 저장소에 없음 | 미배정 — 후속 change 후보 |
| 거래일을 건너 살아 있는 주문을 다른 날 취소하면 그 취소가 종결 증거로 찾히지 않음(`ConfirmedCancelOf` 의 거래일 결속 — 엔진 주문 `TimeInForce DAY` 전제, review §9) | 수용된 잔여(이름 붙임) | 미배정 |
| 브로커 거절 본문의 JSON 중복 키(뒤 값이 이김, review §7 A8) | 수용된 잔여 — 이론적(브로커 실측 0) | 미배정 |


## a091 이월 · 미배정 후속 (a091 archive 뒤에도 남는 의무 — 2026-10-01)

`a091-a-stop-that-sold-nothing-is-critical` 가 넘긴 것. 근거는 a091 `design.md` D1 · D3 「새는 칸의 처분」 · D5 「후속 후보」 · D7 「이름 붙인 잔여」 · D8 와 `issues.md` 「후속 후보 — `Notifier.Acknowledge` 의 잠금 범위」.

| 항목 | 사유 | 행선 |
|---|---|---|
| 배포 전 `main` 과 SchemaVersion 대조 · 엔진 재시작(두 시장 닫힌 창) · 배포 뒤 첫 `exit.stop_sold_nothing` 실물 확인 · ntfy 구독 필터 갱신(`exit.proposal_capped` 로 거르던 필터는 보호 0주를 더 이상 못 본다 — `docs/operations.md`) | 스키마 변경 0 · 엔진 정지 = 손절 없음 | 사람 항목 — 배포 승인 때 |
| 알림 켜짐 게이트(D1, Manager Q1 — a095 정본 문자 그대로): `notifications.enabled = false` 엔진에서는 보호 0주가 옛 종류 normal + 로그 줄로 남는다 | 사용자 거부권 항목 — Manager 가 보고 | 사람 결정 큐 |
| `Notifier.Acknowledge` 의 행별 · 배치 잠금 해제(`internal/obs/notifier.go` 승인이 `n.mu` 를 쥔 채 행마다 fsync — 밀린 행 100 승인 중 exit critical 기록 대기 실측 0.30~0.49s, 앞선 조건 1.06~1.28s) | 셈-해제 배제 불변식(a092 · a124) 아래 설계 필요 — 기존 exit critical 전부가 같은 대기를 진다 | 미배정 — 후속 change 후보 |
| 보유 0 을 읽었는데 ③(no_holding)로 안 떨어지는 칸 셋 — (a) 매도가능 조회 실패 · (b) 매도가능 조회 지연으로 보유 스냅숏 낡음(StaleSnapshot) · (c) 로컬 수량 오류. 셋 다 critical 로 남김(과보고 방향). 넓히려면 `reconcileFloor` 가 보유 0 을 입력으로 넘겨야 함(riskcalc · 하한 공급자 편집). 또 ③ 는 보유 조회 한 번의 0 을 믿는다(base 와 같은 성질) | 범위 밖 — 이름 붙인 잔여 | 미배정 — 후속 change 후보 |
| 공식 클라이언트가 요청 · 본문 읽기 오류를 문자열로 감싸 취소 원인을 지움(`internal/official/client.go` `fmt.Errorf("%w: %s", ErrTransport, err)`) — HTTP 요청 중 종료 취소가 ①(floor_unknown) critical 로 기록됨(과보고, 가짜 래치는 `WithoutCancel` 로 없음) | 이름 붙인 잔여(5라운드 codex P2) | 미배정 |
| 계좌 가림 — exit 루프의 다른 `logErr` 호출자 · base 의 나머지 계좌 필드 줄(a091 은 자기가 새로 닿게 한 줄과 `Notifier.escalate` 두 줄만 가렸다) | a090 D12 · D13 와 같은 큐 | 사람 결정 큐 「계좌 가림 설계」 |

## a090 이월 · 미배정 후속 (a090 archive 뒤에도 남는 의무 — 2026-10-01)

`a090-an-unobserved-position-is-counted` 가 넘긴 것. 근거는 a090 `design.md` D1 · D2 · D9 · D11 · D12 · D13 · Q2 · Q3 와 `review.md` 「리뷰 라운드 1」 · 「Manager 판정 (2026-09-30)」.

| 항목 | 사유 | 행선 |
|---|---|---|
| Q2 실측 — 정지·0가격 종목 포함 `/prices` 읽기 전용 GET 1회(장중, 쓰기 0) → design D6 의 [미측정] 두 행 확정(tasks 0.8) | 라이브 실측 — 이 로트 실행 금지(Manager 2026-09-30). 규칙은 결과와 무관(빈도만 알려 줌) | 사람 항목 — 사전 승인 범위 안에서 사람이 장중 실행 |
| 배포 전 `main` 과 SchemaVersion 대조 · 엔진 재시작(두 시장 닫힌 창) · 배포 뒤 첫 포지션 단위 경보의 실물 확인(tasks 5.1 · 5.2 · 5.3) | 스키마 변경 0 · 엔진 정지 = 손절 없음 | 사람 항목 — 배포 승인 때 |
| settle 의 루프 체류 몫 실측(주기당 최악 2K+2P 로컬 트랜잭션 — review 「Manager 판정」 표) · 상한 캡 | 정본 「루프에 남는 몫은 이름을 갖고 편성」 — 이름·크기는 기록, 실측·캡은 frozen 범위 밖 | 미배정 — 후속 로트(측정 후 필요 시 캡) |
| 하류 임대 재검사 5자리 무음(`judge` · `judgeRatchet` · `judgeLadder` · `refreshObservation` · `record` 의 `!o.quoteUsable(quote)` → `return nil`) | Q3 — 판정 진입을 「관측됨」 으로 봄 | 미배정 — 후속 change 후보 |
| 지속 B2(작업 집합 오류)를 어느 사다리도 재지 않음 · 재시작 창(기점 소실) · workingSet B6(미관리 normal)·B10(완료 정책) | D11-2 · D2 · D1 명명 잔여 | 미배정 |
| `position_id`(계좌번호 무염 해시) 알림 탑재 · 기록 실패 시 입구 `Notifier.escalate` 로그의 `FieldAccount` 원문(a090 이 호출 경로 하나 추가) | base 관행(원장 내부 키 · a092 이전 승격 로그) — a090 신설 표면은 계좌 0(R17) | 사람 결정 큐 「계좌 가림 설계」(D12 · D13) |
| 착지 기록이 형제의 같은-파일 편집에 연쇄 거절됨 — 착지는 번들이 역사에 든 커밋(바닥) **이후**여야 하는데, 자기 Go 착지 뒤 형제가 같은 파일을 다시 고치면 그 뒤 어느 커밋도 번들 리비전과 맞지 않는다. 거절 원문(a090 2026-10-01, a094 `e5e7a67f` 가 사이에 exitloop.go 편집): "no commit at or after the evidence (6dd1e74ed666) is accepted as the landing — at the first commit walked, landing point 6dd1e74ed666 is not the revision this evidence describes: internal/app/engine/exitloop.go". a090 은 로컬 합성 커밋(A)으로 게이트를 돌렸다(review 「게이트」) | 조건 5 · 8 의 합성 효과 — 형제 로트가 같은 파일을 병행 편집하는 한 반복 | 미배정 — gate change 후보(Manager 2026-10-01: 같은-파일 로트 직렬화 또는 착지 기록 선행) |

## a112 이월 · 미배정 후속 (a112 5.2.2.2 에서 발견 — 2026-10-01)

`a112-run-four-strategy-families-independently` 태스크 5.2.2.2 가 실측으로 올린 것. 근거는 a112 `review.md` 「5.2.2.2 잔여」 절과 영수증 `openspec/changes/archive/2026-10-04-a112-run-four-strategy-families-independently/analysis/measurements/lot-5.2.2.2/schema-pin-receipt.log`.

| 항목 | 사유 | 행선 |
|---|---|---|
| **스키마 핀 27 두 자리** → `journal.SchemaVersion` 결속(또는 읽는 표가 존재하는 최소 버전 이상). ① riskbucket — `internal/riskbucket/production_snapshot_authority.go` `productionRiskJournalSchema = 27` · `PRAGMA user_version` 정확 일치 비교(생산 호출 `internal/app/engine/strategy_entry_supervisor.go` 위험 적재기). ② **strategyrouter**(2026-10-01 a127 조사로 추가) — `internal/strategyrouter/production.go:38` `productionRouteJournalV = 27` · `:609` 정확 일치, supervisor 가 route 적재기에 같은 실원장 경로를 넘긴다. `internal/journal/schema.go:6` `SchemaVersion = 35`. **핀은 SchemaVersion 이 이미 29 이던 커밋 `8022f578`(2026-08-04)에서 태어나 출하된 어느 스키마와도 맞은 적이 없다**(a127 F10 실측 — 앞 판의 「a084 부터 28 이상」은 부정확). route 적재기는 위험보다 **앞**이므로 riskbucket 만 고쳐도 1차 레그는 0 그대로다 — **생산 route · 위험 권한은 어느 범위에서도 ready 가 될 수 없다**(a112 시험 fixture 의 v27 축소 원장이 두 자리 모두를 가렸다) | **레인 활성화의 경성 선행 — route · 위험 두 적재기 모두** — 수리 전에 서명 활성화하면 route 미준비 · 전 범위 `AuthorityUnavailable`(1차 레그 0). 오늘 생산 서명 활성화 0 이라 동작 변화 없음. 시험 stub(user_version 27)이 결함을 가렸다. 수리 착지 시 a112 시험 다리 `a112MirrorLedgerIntoRiskStub` 제거 · 적재기를 실제 원장으로 단일화(트립와이어 `TestTheRiskStubBridgeIsStillNeededBecauseTheLoaderRefusesTheRealJournal` 가 실패로 알림) | **a127** 이 두 적재기를 함께 수리(Manager 2026-10-01 승인) — riskbucket 은 High-risk · a066 영역(판정 B) |
| **weekly family 활성화의 경성 선행: riskbucket horizon 매핑 결정.** weekly 레인의 horizon 은 `WEEKLY` 인데 위험 버킷 horizon 은 `SHORT`/`MEDIUM` 뿐이라 적재기(`bindProductionRiskInputs`)가 「unsupported horizon」 으로 거절한다 — weekly family 는 활성화돼도 1차 레그 0. 매핑(어느 수평선 용량을 쓰는가)은 위험 의미론 결정 | 스키마 핀과 같은 등급 — 현재 상태는 핀 시험 `TestAWeeklyLaneIsRefusedForItsHorizonUntilAMappingIsDecided`(a112 6.1, 2026-10-01 Manager 판정)가 고정 — 미래 편집이 무음으로 SHORT 에 넣는 것을 막음 | weekly 활성화 로트의 freeze 몫 — 서명 정책 형식 질문(매니페스트 v2)과 함께 |
| **lease 행의 계보 열 — 활성화 로트의 선행(a112 6.3 판정 (A)+(C)).** lease 발급 후 ~ transport 전 조정 결과 변경은 lease 행에 계보가 없어 검출 불가 — `FinalAuthorityCheck` 는 스케줄 재검증 · 가족 만료만 본다. 오늘 닫힌 지점은 **발급 시점**(1차 레그 권한의 소유자 범위 선택 + 봉인 identity 대조, admission 커밋 전 — `a112_dispatch_lineage_test.go` `TestALineageDriftIsRefusedAtIssuanceBeforeAnyReservationLeaseOrBrokerCall`). family 는 lease 의 LaneID 가 정본 표로 함의한다. **(B)** lease 행에 `lineage_identity` · `score_version` · `calibration_digest`(+ family) 열(journal schema v33 · 트리거) 추가 + transport 직전 현재 조정 결과 재대조 | 4-가족 서명 활성화를 실제로 켜는 로트의 **선행**(활성화 0 인 지금은 YAGNI — Manager 판정 2026-10-01) |
| **breakout 같은 setup 둘째 첫 레그 — 봉인 해제 로트의 면제 불가 선행(a112 6.4 판정 2026-10-01).** 손절-종결 → 재시작 → 새 매니페스트 CampaignID → 같은 setup 둘째 첫 레그: 첫 레그 캠페인이 손절로 CLOSED 되고 claim 이 풀린다 → 재시작으로 레인 prior 를 잃는다(생산에서 `BreakoutRequest.Prior` 를 채우는 곳 0) → 같은 setup 이 봉 하나만 더해져도 새 스냅숏 digest · 새 ProposalID 로 다시 PROPOSED → 서명 제안 입력 매니페스트가 새 CampaignID 를 실으면 journal 은 FLAT/CLOSED · claim 없음 · 새 캠페인 PK 로 받아들인다 → 같은 setup 의 둘째 첫 레그. 스펙(breakout-retest-strategy-lane 「breakout v1 production 권위는 first-leg 하나로 제한된다」): "Proposal replay, duplicate bar delivery, correction 또는 restart가 동일 setup/bar에서 두 번째 first-leg 권위를 만들면 안 되며 (MUST NOT)". 오늘은 벽 뒤(생산 breakout 입력 전부 거절 — 핀 `TestEveryProductionBreakoutLaneInputIsRefusedAtTheWall` · `TestNoNonTestCodeBuildsABreakoutLaneInputAroundTheWall`) | **면제 불가: 「ErrBreakoutEvidenceUnavailable 해제 로트는 B(SetupID 계보 결속) 또는 C(consumed-setup 원장 기록) 착지 전 해제 불가」.** B = breakout 계보에 SetupID 를 싣고 CampaignID 를 그것에서 유도(봉인 · 골든 — freeze 급), C = admission 과 같은 트랜잭션의 consumed-setup 원장 기록(스키마 범프) | breakout 증거 생산자(ATR · RVOL · 윗꼬리 · 거래량 확장)를 세워 벽을 여는 로트 — 그 로트 안에서 B 또는 C 를 먼저 착지 |
| **전략 dispatch crash 뒤 결과 대사 · 복구 — 활성화 로트의 면제 불가 선행(a112 6.5 판정 2026-10-01).** 근거(④ 의 위험): SUBMITTING 중 crash 면 브로커에 실주문이 존재할 수 있는데 이 빌드에는 그 결과를 알아낼 경로가 0 이다 — `internal/journal/strategy_dispatch_runtime.go:178` 「No constructor for that authority exists in this build」(ATTESTED_OUTCOME_REQUIRED), `DiscoverStrategyDispatchRecovery` · `RecoverClaimedStrategyDispatchLease` 생산 호출 0 · `internal/strategyruntime`(모델 복구) 생산 import 0 (grep 영수증 `lot-6.5-6.6/recovery-callers-grep.log`). UNKNOWN_BROKER_STATE 가족의 실체이며 활성화 상태에서는 liveness 가 아니라 안전 문제다. 같은 경로 부재로 ② admission 커밋 뒤 · lease 앞, ③ claim 뒤 · SUBMITTING 앞 crash 도 재시작 뒤 이어지지 않는다(그 종목 claim · HELD 용량이 풀리지 않음 — ② 는 lease 가 없어 복구 열거에도 안 보인다). 측정: engine `TestADispatchCrashAtEveryStepNeitherDuplicatesTheOrderNorReleasesTheCapacity`(네 crash 지점 · 재시작) | **면제 불가: 「활성화 로트는 outcome reconciliation 착지 전 해제 불가」.** 대사 경로 = 공식 브로커 정확 결과 증명의 생성자 + 복구 호출자(ATTESTED_OUTCOME_REQUIRED · REFUSE_RELEASE_REQUIRED) + admission-without-lease 열거 | 4-가족 서명 활성화를 실제로 켜는 로트 — 그 로트 안에서 먼저 착지 |
| **활성화 실패 사유 · 삼킨 사이클 오류의 운영자 표면 — 활성화 로트의 선행(a112 8.8.4 판정 2026-10-04).** 8.8.4 로트 B 로 4-가족 활성화 적재의 거절은 필드명을 오류 사슬에 싣지만(`strategyrouter/production_family_activation.go`), 엔진은 그 오류를 관문 판별(`errors.Is` 미선언 / 그 밖 되돌림)에만 쓰고 snapshot · projection 에 내지 않는다 — 운영자는 여전히 「배포 안 함」 과 「만료 · 폐기 · 결속 불일치」 를 화면에서 못 가른다(미선언과 넷 다 ON 이 둘 다 `GatedCount=0`). `refreshOnly` 갈래가 세는 `SwallowedCycleErrors` 도 같은 처지. 또 공유 읽기 함수 `readProductionRouteFile` 이 OS 원인(없음 · 심링크 · 권한 · 소유자 · 크기)을 자기 sentinel 하나로 접는다. **같은 표면(a112 8.5 보이스 1 P2-3, 2026-10-04 합류):** 활성화가 검증된 시장이 조정 앞에서 닫힌 주기(FX · 설정 · 열쇠 · 중복 · 적재 · 고장)에 ON 레인이 입력 없이 돌아 `REFUSED / ARBITRATION_SEAL_MISMATCH / "the sealed proposal does not establish this worker's lane"` 로 보고된다 — FX 장애가 봉인 경보처럼 읽힌다(latch · 원장 0, 진단 전용). 후보: 「no input this wave」 전용 detail | snapshot/projection 필드(활성화 실패 사유 · 삼킨 오류 계수/첫 원인) + 읽기 함수 원인 보존 — 화면 · OpenAPI 계약 변경이라 7.3 투영과 같은 규율 | 4-가족 서명 활성화를 실제로 켜는 로트 — 그 로트 안에서 먼저 착지 |
| ~~**SHADOW 런타임(spec four-family-strategy-runtime :88-94) — shadow 매니페스트 선행**~~ **닫힘(2026-10-05 — a112 7.3.1 구현 로트 착지로 종결; 브리프 v3.3 · 결정 63 digest 핀 · `internal/strategyshadow` · 도구 `tools/a112-family-shadow` · docs/operations.md 「KR/US SHADOW 관측」 절). 후속 = 4-가족 활성화 로트의 shadow 수집 운영(design :295 — 배포 뒤 사람 · A100). 아래 본문은 역사 기록.** (a112 7.3.1 R4, 2026-10-04; 사용자 결정 2026-10-04: a112 안에서 구축.) 오늘 SHADOW 상태 0(runtime 어휘 = {UNOBSERVED}, 핀 `TestTheRuntimeVocabularyIsExactlyUnobservedUntilAShadowLotExtendsIt`), 재시작 OFF/OFF/UNOBSERVED 핀 `TestARestartAfterAnObservedPromotionComesBackOffOffUnobservedWithNothingWritten`. a112 안에서 SHADOW 를 세울지 · 미룰지 · spec 델타를 정정할지는 사용자 결정(7.3.1 R3) | SHADOW 를 세우는 로트는 함께 결정 · 착지: ① 신뢰 앵커 — 서명(spec :89 「signed」) vs 외부 digest 핀(형제 4-가족 활성화 매니페스트가 결정 61 로 서명을 뺀 선례) ② 매니페스트 형식 · 생성기 · 골든 · docs/operations.md(8.5 가 「셋 없으면 배포 불가」 로 판정한 선례) ③ 골든 four-family-runtime-v1.json runtime 어휘 개정(SHADOW 추가 — Manager amendment) ④ 재시작 비복구(활성화 적재기 패턴 — 매 파도 재읽기) | 사용자 결정(7.3.1 R3) 뒤 배정 |
| **권한 적재 파도의 취소 응답(4 적재기 join 의 ctx select) — 활성화 로트의 면제 불가 선행 「핀 선언 전 착지」(a112 8.5 판정 2026-10-04, codex r2 P1 → (d)).** route · proposal · risk · account 적재기가 시장별 goroutine 안에서 ctx 없는 파일 읽기를 하고 `<-outcomes` 로 기다린다(ctx select 없음 — `strategy_route_authority.go` · `strategy_proposal_authority.go` `collect` · `strategy_risk_authority.go` · 계좌 적재기). 읽기에 들어간 뒤의 취소에 응답하지 않는 것은 8.8.4 이동 전부터 넷 공통 결함이고, 8.8.4 · 8.5 의 관문 계산은 핀 선언 시장의 주기에 같은 모양의 소형 읽기를 더했다(오늘 생산 핀 0 → 노출 0). 근거 `analysis/review-8.5-2026-10/design-brief-B2-P1.md` §1 · §3 · §4 | 넷을 **동시에** 한 규칙으로: 읽기 앞 `ctx.Err()` + join 의 ctx select(취소면 즉시 반환 · 미완 시장은 취소 사유). 핀: 적재기마다 stalled-reader seam — 읽기 진입 뒤 취소 → 시한 안 반환 · 막힌 시장만 취소 사유 · 다른 시장 결과 무손상 · 읽기 해제 뒤 goroutine 0. **후보(보이스 1 P2-1 제안, a112 8.5 범위 밖):** `FinalAuthorityCheck`(strategy_dispatch_cycle.go)가 만료만 재확인한다 — 핀 · 철회를 dispatch 직전에 재확인하면 제안 적재 ~ dispatch 사이 철회 경합이 판본과 무관하게 닫힌다. (d) 는 a112 8.5 의 F1(관문 스냅숏 낡음 — 판정 재계산으로 닫음)을 대신하지 못한다(직교) | 4-가족 서명 활성화의 첫 핀을 선언하는 로트보다 **먼저** 착지 |

## a070 이월 · 미배정 후속 (a070 `--skip-specs` archive 뒤 — 2026-10-01)

`a070-add-multi-market-horizon-router` 처분 ②(부분 대체)가 넘긴 것. 근거는 아카이브의 `analysis/disposition-audit.md` §5 · §7.

| 항목 | 사유 | 행선 |
|---|---|---|
| 생산 호출자 0 인 `strategyrouter` 섬 정리 — `Route()`(a112 `RouteSet` 이 대체) · durable `SchedulerState`/CAS/rollback/`MarketRecordStore` · `MigrateLegacy` · `QuotaAuthority` | a112 시험 다수가 `Route` 를 참조 대조로 부름 — quota 권한이 정해진 뒤 한 번에 | a112 7.1 뒤 정리 change 하나(미배정) |
