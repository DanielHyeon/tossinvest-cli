# Design: harden-execution-base

## Context

P0 정찰 + proposal-freeze 리뷰(3보이스 49건, review.md)로 확인된 실코드 제약: 공식 API status는 실질 OPEN/CLOSED 2값, per-fill ID 없음(누적 filledQuantity만), cancel/modify는 새 주문번호 발급, `Orders`는 pagination을 버림(첫 페이지만), official 클라이언트에 `GetOrderAvailableActions` 부재, `newAppContext`는 root_test.go가 거의 커버하지 않음, hybrid 브로커는 자격증명 부재 시 WTS로 조용히 폴백. 이 제약들이 본 설계의 입력이다.

## Goals / Non-Goals

**Goals**: 엔진이 신뢰할 수 있는 주문 실행 계층 — mutation 중심 상태 모델·journal·멱등성·체결 감지·reconciliation·안전 인터록·관측성. 실전 직행(U2) 전제.
**Non-Goals**: 전략·스케줄러(P3), Guardian 수치 구현(P2 — 인터페이스 초안만), 웹 API(P4), **계좌 단일 writer lease와 CLI/MCP 중재(P4 데몬과 함께 — 잔존 리스크는 proposal에 기록)**, module rename. capability soak·실계좌 검증·약관 검토는 별도 change `verify-execution-capability`.

## Decisions

### D1. 신규 orchestrator가 감싸고, upstream은 수정하지 않는다
엔진 주문 파이프는 신규 패키지 `internal/execgw`(ExecutionGateway)가 `*trading.Service`를 **감싼다** — `internal/trading/service.go`는 수정하지 않는다. Gateway가 journal 기록→GuardianDecision 검증→제출→확정을 orchestrate한다. CLI/MCP 경로는 upstream 그대로(§0.2 보존) — journal 의무는 엔진 프로필 경로에만 적용된다.

### D2. Journal: 저장소 밖 XDG data 경로, 엄격 내구성
경로: `$TOSSOS_DATA_DIR` > `$XDG_DATA_HOME/tossos` > `~/.local/share/tossos`. 로컬 저널링 FS allowlist(ext4/xfs/btrfs) 검사, 실패 시 기동 거부. `BEGIN IMMEDIATE` + `synchronous=FULL`(intent 쓰기), 스키마 버전 필드, additive migration, 손상 시 기동 거부. lineage는 journal 트랜잭션 내 기록(기존 orderlineage JSON은 CLI용으로 보존). Phase 2 원장 import 계약: 안정 PK(intent id), 불변 intent 필드, MutationAttempt 이력 보존.

### D3. 브로커 어댑터: official 직접 + 사전 확인 대체
엔진 브로커는 `official.Client`를 직접 감싸는 어댑터(신규, `internal/execgw` 내) — hybrid 미사용, config `OpenAPI.Enabled/Prefer` 무시, 자격증명 없으면 기동 거부. cancel/amend 사전 확인은 `OrderByID` 상태 파생으로 구현(공식에 available-actions 부재). 상태 파생 함수는 `(status, canceledAt, filledQuantity, quantity, lineage)` 우선순위 표 + UNKNOWN_BROKER_STATE fail-closed. 파생 fixture는 upstream 테스트 fixture에서 출발하고, 실계좌 확정은 verify-execution-capability에서 보강(미지 값은 fail-closed라 enum 미완성이 안전을 깨지 않음).

### D4. 체결 감지: 누적 스냅샷 + 폴링 권위
`internal/filldetect`: 폴링(미체결 pagination 완주 + OrderByID + 잔고) 권위, SSE 힌트 coalescing. per-fill ID 부재 → lineage 노드 단위 누적 filledQuantity 스냅샷, 양의 delta만 반영, 감소·역순은 fail-closed. 선행 과제: `official.Orders` pagination 완주 헬퍼(cursor loop 방어) — upstream 파일 수정 없이 어댑터 측에서 반복 호출.

### D5. 알림: ntfy 구체 구현 + critical outbox
`internal/obs`: 구조화 로그(주문 전이·reconcile·오류, 셀 수 있는 이벤트 규약) — 메트릭 서버는 P4로. 알림은 ntfy 구체 구현(추상화 없음). critical 이벤트는 SQLite outbox(journal DB 내 테이블) 경유 재전송, 지속 실패 시 신규 진입 차단. heartbeat: 엔진이 주기 publish, 수신 측(ntfy 클라이언트 설정)이 간격 초과 경보.

### D6. 시간 규율
`internal/clock`: 주입 가능 Clock, 시장 TZ(KST/ET), 거래일 경계, DST 테이블 테스트. journal 타임스탬프·staleness·안정화 간격·SLO 측정이 모두 이 패키지를 쓴다.

### D7. StockOS SDD 규칙 이식 범위
(기존 D5 유지) 도구 독립 규칙만 — WORKFLOW 개정 2 + sdd-workflow 델타. CodeGraph류 도구는 미도입, 대체는 Pre-Edit 선언. Pre-Edit 전문 작성은 upstream 파일 수정 task(1.1, 1.2, 1.4, 3.2)에만 적용하고 신규 패키지 task는 §0 검토 + race/crash 테스트로 갈음.

## Risks / Trade-offs

- [shim이 upstream 파일 소량 수정] → characterization 테스트를 먼저 작성(태스크 1.2a), DoD = 기존 테스트 + characterization 통과
- [MCP 표면이 Gateway 우회 가능] → Phase 1 잔존 리스크로 문서화, P4 데몬 단일 writer에서 해소. 그 전까지 MCP 주문은 사용자 수동 영역
- [폴링 SLO vs rate limit] → retry matrix 표에 폴링 예산 계상, verify-execution-capability 실측으로 수치 확정
- [fingerprint 매칭의 이론적 모호성] → 심볼당 in-flight 1건 제한으로 유일성 보장

## Migration Plan

신규 패키지 + shim 최소 수정. 롤백 = revert. journal DB는 신규 생성물.

## Open Questions

- retry matrix·SLO·안정화 수치의 최종값 (verify-execution-capability의 실측 입력, 태스크 2.6에서 보수 기본값으로 시작)
- ntfy self-host 여부 (태스크 4.3 구현 시 사용자 확인)
