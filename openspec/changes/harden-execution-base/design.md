# Design: harden-execution-base

## Context

P0 정찰로 확인된 사실: 배선은 `cmd/tossctl/root.go:newAppContext`에 갇혀 있고, 주문 정책은 `internal/trading.Service`(Broker 인터페이스)로 이미 분리돼 있으며, 상태기계·멱등성·rate limit·관측성은 코드에 없다. SSE(`internal/push`)는 stdout 출력 소비자뿐이다. 일반 주문에는 client order id가 없다(조건주문만 ClientOrderID). upstream 테스트 650개는 회귀 금지 대상이다.

## Goals / Non-Goals

**Goals**: 엔진이 신뢰할 수 있는 주문 실행 계층 — 상태기계·journal·멱등성·체결 감지·reconciliation·안전 인터록·관측성. 사용자 결정 U2(실전 직행)를 전제로, 안전은 위험 한도와 fail-closed로 담보.
**Non-Goals**: 전략·후보·스케줄러(Phase 3), Guardian 수치 구현(Phase 2 — 여기서는 인터록 계약만), 웹 API(Phase 4), Go module rename.

## Decisions

### D1. 위임 shim으로 리팩터 (upstream 회귀 0)
`newAppContext`는 `internal/app.New(Options)`로 승격하되 `cmd/tossctl`에 동명 래퍼를 남긴다. `trading.MutationResult`는 `internal/domain`으로 옮기고 기존 위치에 `type MutationResult = domain.MutationResult` alias를 둔다. 조건주문 게이트도 동일 패턴. 검증: 전체 upstream 테스트 무수정 통과.

### D2. Intent journal은 단일 테이블 SQLite, 저장소 밖
경로 우선순위: `$TOSSOS_DATA_DIR` > `os.UserConfigDir()/tossos/journal.db`. 기동 시 파일시스템 유형 검사(fuseblk 거부). Phase 2 원장(T2.1)은 이 journal을 흡수·확장한다(스키마 호환 설계). WAL 모드 + 단일 프로세스 락.

### D3. 체결 감지 계층 분리
`internal/filldetect`: 폴링 루프(권위, SLO 기본 장중 5s)와 SSE 힌트 소비자(coalescing)를 같은 재조회 파이프에 합류. 상태 반영은 주문번호·체결 식별 기준 멱등 upsert. lineage(`orderlineage`)로 정정 후 주문번호 추적.

### D4. 알림 채널은 ntfy 우선
외부 의존 없는 HTTP POST(ntfy.sh self-host 가능) 기본, 인터페이스로 추상화해 Telegram 등 교체 가능. 알림 실패는 주문 경로를 차단하지 않는다(best-effort + 로그).

### D5. StockOS SDD 규칙 이식 범위
사용자 지시로 stockos/.claude/CLAUDE.md의 규칙 중 도구 독립적인 것만 이식: 안전 불변식(§0), 권위 경계, 위험도 분류, Pre-Edit 선언, 완료 보고 조건, 실행 순서 — docs/WORKFLOW.md 개정 2 + sdd-workflow 스펙 델타. CodeGraph/Neo4j/Function Logic Map 도구 체계는 미도입(도구 부재) — 대체 절차는 Pre-Edit 선언의 "기존 동작 파악 근거"(기존 테스트·fixture·호출부 확인). 도구 도입은 선택 과제로 보류.

## Risks / Trade-offs

- [shim이 upstream 파일에 소량 수정 유발] → 수정은 위임 1줄 수준으로 제한, 전체 테스트로 검증
- [폴링 SLO 5s가 rate limit과 충돌 가능] → T1.12 실측으로 조정, retry matrix에 예산 계상
- [journal과 Phase 2 원장의 이중 저장 기간] → journal 스키마를 원장 서브셋으로 설계해 마이그레이션 단순화
- [알림 채널 외부 서비스 의존] → self-host 옵션 문서화, 실패 시 로그 폴백

## Migration Plan

신규 패키지 추가 + shim 수정만. 롤백 = 커밋 revert. journal DB는 신규 생성물이라 롤백 부담 없음.

## Open Questions

- 폴링 SLO 기본값(5s)의 rate limit 실측 후 확정 (T1.12)
- ntfy vs Telegram 최종 선택 (T1.10 구현 시 사용자 확인)
