# gstack 리뷰 기록 — add-tossos-foundation

- 일시: 2026-07-26 · 실행: /autoplan (자율 모드 어댑테이션: premise는 사용자 지시문에서 기확정, 최종 게이트는 보고서로 제시)
- 보이스: codex CLI(적대적 CEO+Eng+DX 통합) + 독립 Claude 서브에이전트 3(CEO/Eng/DX) — 총 발견 66건 (critical 8, high 30, medium 28)
- 대상: proposal/design/tasks/specs(product-fork, sdd-workflow), docs/ROADMAP.md, docs/WORKFLOW.md, docs/stockos-inventory.md
- 복원 지점: ~/.gstack/projects/TossOS/feat-p0-foundation-autoplan-restore-20260726-010936.md

## 합의표

CEO (Claude / codex → 합의): 전제 타당 PARTIAL/PARTIAL→**PARTIAL** · 올바른 문제 NO/NO→**NO(수직 슬라이스 필요)** · 범위 보정 PARTIAL/PARTIAL→**PARTIAL(P5 과대)** · 대안 탐색 NO/—→**NO** · 리스크 커버 PARTIAL/PARTIAL→**PARTIAL** · 궤적 건전성 NO/NO→**NO(가치 후반 집중)**

Eng (Claude / codex → 합의): 아키텍처 PARTIAL/PARTIAL · 엣지 케이스 **NO/NO** · 게이트 검증 가능성 PARTIAL/PARTIAL · 숨은 복잡성 PARTIAL/PARTIAL · task 실행 가능성 PARTIAL/— · 스펙 테스트 가능성 PARTIAL/PARTIAL

DX (Claude / codex → 합의): 발견 가능성 **NO**/— · 모순 없음 **NO**/— · 마찰 PARTIAL/PARTIAL · 게이트 자동화 PARTIAL/PARTIAL · 오류 경로 PARTIAL/PARTIAL · 규약 일관성 PARTIAL/—

교차 주제(2+ 보이스 독립 지적 = 고신뢰 신호): go-live 승격 사다리 부재 · T1.9 게이트 순서 결함 · 관측성 지연 · SSE 권위 역전 · unknown-outcome 주문 · 원장 내구성/NTFS · 리뷰 게이트 비용 · 게이트 자동화 부재 · 실계좌 보호 비기계성 · 체크박스 표류(실증) · 브랜치 모델 · P1 리팩터 vs 머지성 모순.

## 수용된 결정 (auto-decide, 6원칙 기준)

| # | 결정 | 반영 위치 | 원칙 |
|---|------|-----------|------|
| A1 | go-live 승격 사다리(replay→paper→capped→normal)를 제품 불변조건화, go-live-protocol.md 산출물 | ROADMAP 원칙7, T1.14, T2.8, T3.5 | P1 |
| A2 | 자동화 게이트 활성화를 Phase 2로 이관 + Guardian 미주입 시 기동 거부 인터록 | T1.9, T2.4 | P1 |
| A3 | 관측성·알림·flatten-all을 Phase 1로 이동 (알림 없는 무인=무감독) | T1.10, T1.11 | P1 |
| A4 | 체결 감지 권위=공식 API 폴링(SLO), SSE=힌트(coalescing) | 원칙5, T1.6 | P1 |
| A5 | durable intent journal(WAL)을 Phase 1로 — P2 원장 숨은 의존 해소 | T1.5 | P3 |
| A6 | 주문 mutation 자동 재시도 금지, IN_DOUBT/AMEND_IN_DOUBT 상태, retry matrix | 원칙8, T1.4, T1.7 | P1 |
| A7 | reconciliation 계약 명세(비교 키·오차·외부 주문 분류·운영자 절차) | T1.8 | P1 |
| A8 | kill switch 운영 모드 체계(NORMAL/ENTRY_BLOCKED/EXIT_ONLY/HALT_ALL) | 원칙6 | P1 |
| A9 | 원장·journal은 ext4(XDG) 강제 + fuseblk 기동 거부, sidecar .gitignore | T2.1, D5, .gitignore | P1 |
| A10 | clock/TZ/거래일 경계를 Phase 1로 (위험관리 선행 조건) | T1.13 | P3 |
| A11 | 계좌당 단일 주문 writer(데몬), CLI/MCP는 경유 또는 maintenance mode | 원칙9, T4.1 | P5 |
| A12 | 공식 API capability soak·FX 경로·ToS 검토 태스크 신설 | T1.12, T1.15 | P1 |
| A13 | 주문 상태기계 정본=공식 API 스키마(WTS 관찰 문서는 보조) | T1.4 | P5 |
| A14 | 엔진 배선 official-only 증명 테스트(hybrid 누수 차단) | 원칙1, T1.1 | P1 |
| A15 | Order/Fill/Position/ProtectionSaga aggregate 경계 선행 설계 + 보호주문 durable saga | T2.2, T2.3 | P5 |
| A16 | tracer slice(수직 슬라이스)와 replay/paper 하네스를 Phase 2에 신설 | T2.8, T2.9 | P6 |
| A17 | StockOS 이식 상수는 출처·검증 상태 기록, 검증 전 보수 기본값 | Phase 2 규칙 | P1 |
| A18 | Phase 5를 5a(운영 콘솔)/5b(풀 UI, capped-live 후)로 분할 | ROADMAP P5 | P3 |
| A19 | 리뷰 게이트 등급제(proposal-freeze+Requirement 수정) + review.md 필수 + make gate 자동화 + 체크-동일-커밋 규칙 | WORKFLOW, sdd-workflow spec, tools/gate.sh | P3/P6 |
| A20 | 발견성: 루트 CLAUDE.md, AGENTS.md 스코프 헤더, openspec/project.md | 신규 파일 | P5 |
| A21 | 브랜치 모델: main(제품)/upstream-sync 분리, feat/p<N>-<change-id>, 커밋 규약 | WORKFLOW, product-fork spec | P5 |
| A22 | 역할 분리를 "작성자·검증자 컨텍스트 분리"로 재정의(모델명 스펙에서 제거) | sdd-workflow spec | P5 |
| A23 | TDD SHALL을 검증 가능한 "테스트 동반+추적성" 기준으로 재정의 | sdd-workflow spec | P5 |
| A24 | 실계좌 보호를 테스트 인프라(격리 config·httptest 강제)로 기계화 | sdd-workflow spec, WORKFLOW | P1 |
| A25 | upstream push URL DISABLED(실증 완료) + sync 결정 로그 요구 | product-fork spec, git | P1 |
| A26 | 예외 경로: 결함 3분류, blocked 규칙, WIP 브랜치, change 폐기, change당 Teammate 1명 | WORKFLOW | P5 |
| A27 | 핵심 4패키지 커버리지 하한을 Phase 게이트에 명시 | ROADMAP 게이트 | P1 |
| A28 | 고액 주문 flag: 청산 주문 한정 상한부 허용 정책 | T2.4 | P1 |
| A29 | 제품 가설·중단 기준 섹션 신설(수치는 사용자 확정 대기) | ROADMAP | P6 |

## 기각·보류 (사유)

- R1 **Go module rename**: 보류 — 사용자 결정 항목(아래). shim 전략(D6)으로 rename 없이도 P1 진행 가능
- R2 **StockOS 백엔드 유지+브로커 어댑터 교체 대안 정량 분석**: 기각 — 사용자가 이미 fork 방향을 확정했고(지시문), KIS 오염 48%·17,830 LOC 단일 파일 증거로 방향 자체는 견고. 알림용으로 사유만 기록
- R3 **CI(push 트리거)**: 보류 — origin 미존재. tools/gate.sh가 로컬 동등물. origin 생성 시 재검토
- R4 **`/.claude/` ignore 해제**: 기각 — upstream 규칙 유지(머지 충돌 최소화), openspec init로 재생성 가능, 규약은 추적되는 openspec/project.md에 기록됨
- R5 **TDD red/green 증거·mutation testing**: 부분 수용 — 커버리지 하한(A27)만. 증거 수집 비용 과대
- R6 **Phase 5 전면 축소/폐지**: 부분 수용 — 사용자의 UI 이식 의도를 5b로 보존하고 순서만 조정(A18)

## 미결정 — 사용자 확인 필요

1. **U1 Go module 경로**: 유지(기본) vs 조기 rename. fork-and-forget 자세라면 rename이 유리하나 diff 오염 큼
2. **U2 go-live 승격 사다리(A1)**: 사용자 지시문은 "shadow/canary 제외"였음. 리뷰 4보이스는 구현체 제외와 별개로 승격 원칙 자체는 필수라 판단해 반영함. 뒤집으려면 원칙7 삭제
3. **U3 제품 가설·중단 기준 수치**: 자본 상한·일일 손실 한도·go/no-go 기준 (ROADMAP 섹션에 자리 마련)
4. **U4 Phase 5 분할(A18)**: 5a 운영 콘솔 우선, 풀 UI는 capped-live 후 — 원래 지시보다 UI가 늦어짐
5. **U5 origin 비공개 저장소**: 생성 시점과 이름 (사용자 GitHub 작업)
6. **U6 markout 윈도우**: StockOS 검증값 5/15/30분(기본) vs 지시문의 1/3/5분

## 재검증

- 리뷰 반영 후 `openspec validate add-tossos-foundation --strict`: 통과 (아래 tasks 3.7)
- 반영 문서: ROADMAP v2, WORKFLOW v2, sdd-workflow/product-fork spec 개정, design D5~D7, CLAUDE.md, AGENTS.md 헤더, openspec/project.md, tools/gate.sh, .gitignore
