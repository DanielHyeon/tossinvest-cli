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
6. kill switch는 신규 진입 차단 전용(BLOCK-ONLY)이되, **운영 모드 체계는 별도로 존재**: NORMAL / ENTRY_BLOCKED(=kill switch) / EXIT_ONLY / HALT_ALL(제출 전면 중단). 수동 비상 청산(flatten-all)은 typed-confirmation 수동 명령으로 제공
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

- **2a `extend-execution-contract`** (측정 무관 — 즉시 구현): 결정 계약(safety class·RiskIntent preimage·generation), 멱등키 기계(브로커 `clientOrderId` — P1 "멱등키 없음" 전제는 사실 오류였다), 진입 측 위험 예약, RECONCILE 상태, 한도 fail-closed·총계 계산 계약, opaque 식별자·OrderStatus 10개 정정, 엔진 Gateway 배선·봉인. **조건주문 코드 0줄.**
- **2b `verify-execution-capability`** (사용자 측정): 멱등키 실동작·TTL 마진, 조건주문 속성(SINGLE+MARKET 손절, OCO/OTO는 LIMIT 전용), CANCEL_REJECTED 레코드 형태, 매도가능수량 의미, 실측 비용표.
- **2c `add-protection-orders`** (2b 결과 위에서 작성 — 아직 미작성): 조건주문 형제 수명주기, 발동 주문 다리(`triggeredOrderId`), 미귀속 관측 격리 원장, 청산 수량 예약, PROTECTION_WEAKENING, flatten의 조건주문 취소. 기본 가설: 보호 = SINGLE+MARKET 손절 단독, 익절은 로컬 청산("한 심볼에 브로커측 매도 청구권 1개").
- **2d `add-core-domain`** (동결 해제 가능 — 2a 완료): 비용 모델(KIS 수치 이식 금지, 2b 실측), Guardian 판정 체인(구조적 RR·등급배수는 P3 이관), 운영 모드(모드×클래스 표), 포지션 원장, provenance, 성과, tracer, **exit 정책(StockOS baseline ratchet·profit ladder 이식 — 기준선 단조 상승 손익 극대화, 순수 판정만; 액추에이션은 2c)**.

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

| ID | 작업 | 비고 |
|----|------|------|
| T3.1 | internal/candidate: 후보 수집 + CandidateEvidence. WTS 만료 시 공식 API 소스로 강등해도 후보가 계속 나오는 구성 명시 | 원칙 5 |
| T3.2 | internal/strategy: 독립 매수 레인 인터페이스, 레인별 ON/OFF, OFF 후 청산 지속 | |
| T3.3 | internal/scheduler: 장 시간 인지 루프 (T1.13 clock/calendar 기반) | |
| T3.4 | internal/performance: 레인별 성과(결정적 링크 없으면 표시 금지), markout 윈도우(기본 5/15/30분), 비용 후 기대값 기반 슬롯 배분 | |
| T3.5 | 실전 운영 개시: Guardian 한도·kill switch·알림 활성 확인 후 레인 ON (승격 단계 없음 — 사용자 결정) | 원칙 7 |

## Phase 4 — Service Layer  (change: `add-httpapi-daemon`)

| ID | 작업 | 비고 |
|----|------|------|
| T4.1 | cmd/tossosd: 데몬 부팅·config·graceful shutdown (internal/app 공유). 계좌당 단일 주문 writer 보장 | 원칙 9 |
| T4.2 | internal/httpapi: REST + SSE(sequence id 순서 보장·스키마 버전), 로컬 토큰 인증 | |
| T4.3 | 운영 엔드포인트: 상태, 운영 모드 전환(typed-confirmation), 레인 제어, reconciliation 강제 실행 | |

## Phase 5 — 운영 콘솔 → 풀 UI  (change: `add-ops-console`, `add-web-ui`)

리뷰 결정: UI를 2단계로 분할. 5a는 안전 운영에 필요한 최소 표면, 5b는 capped-live 성공 후.

**5a — 운영 콘솔 (`add-ops-console`)**: 상태·포지션·미체결·차단 사유·reconciliation 상태·운영 모드/kill switch(typed-confirmation)·실현손익 요약. style.css 디자인 시스템 + 안전 UX(사유 입력·차단 칩) + useDashboardStream 이식. 차트 없음.

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
