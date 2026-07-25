# Tasks: harden-execution-base

> [M]=Manager, [T]=Teammate. TDD 필수, 완료 체크는 산출물 커밋과 동일 커밋.
> Pre-Edit 선언 전문 작성: **upstream 파일을 수정하는 task(1.1, 1.2, 1.4, 3.4)에만** 적용. 신규 패키지 task는 §0 검토 + race/crash 테스트로 갈음.
> 아키텍처 고정(design.md): 신규 `internal/execgw`가 `*trading.Service`를 감싼다 — `internal/trading/service.go` 수정 금지. 엔진 journal 의무는 엔진 프로필 경로에만 적용(CLI/MCP는 upstream 유지).

## 1. Shim 리팩터 (upstream 회귀 0)

- [x] 1.1 [T][High-risk] `trading.MutationResult`(internal/trading/result.go) → `internal/domain` 이동 + 기존 위치 type alias. 전체 upstream 테스트 무수정 통과 [Pre-Edit]
- [x] 1.2 [T][High-risk] `newAppContext`(cmd/tossctl/root.go:480) characterization 테스트 선행 작성: configDir 경로 override 4종, official 클라이언트 구성 truth table(creds×Enabled×Prefer), lineage 배선, 세션 부재 허용 — 이후 `internal/app.New(Options)`로 승격 + cmd 쪽 위임 래퍼. DoD = 기존 테스트 + characterization 통과 [Pre-Edit]
- [ ] 1.3 [T] `internal/app` 엔진 프로필: official 직접 구성(자격증명 없으면 기동 거부, config Prefer/Enabled 무시), hybrid·WTS mutator 정적 import 부재 테스트, 전 mutation matrix WTS-spy 0회 테스트
- [ ] 1.4 [T][High-risk] 조건주문 게이트(cmd/tossctl/conditional_gate.go)·intent 조립 → `internal/trading` 이동 + `Service`에 Conditional Preview/Execute 메서드, cmd 쪽 shim [Pre-Edit]

## 2. 시간·Journal·상태 모델

- [ ] 2.0 [T] `internal/clock`: 주입 가능 Clock, 시장별 TZ(KST/ET), 거래일 경계, DST 전환 테스트
- [ ] 2.1 [T][High-risk] `internal/journal` 스키마: intent·MutationAttempt·lineage edge 테이블, 스키마 버전, XDG data 경로 해석·FS allowlist 가드(fuseblk 픽스처 테스트 — 본 저장소가 fuseblk라 실검증 가능)
- [ ] 2.2 [T][High-risk] journal 내구성: BEGIN IMMEDIATE + synchronous=FULL, 커밋 성공 후에만 제출 진행, 손상 감지 기동 거부, disk-full·crash-during-write 테스트
- [ ] 2.3 [T] 브로커 상태 파생 함수: (status, canceledAt, filledQuantity, quantity, lineage) 우선순위 표 구현, upstream fixture 기반 표 테스트, 미지 status → UNKNOWN_BROKER_STATE fail-closed
- [ ] 2.4 [T] MutationAttempt 수명주기: RECORDED→DISPATCH_STARTED→ACKED/IN_DOUBT 전이, 재시작 시 RECORDED→NOT_DISPATCHED 안전 종결, transport fault-injection(전송 전/중/후) 테스트
- [ ] 2.5 [T][High-risk] `internal/execgw` Gateway 뼈대: journal 선기록→GuardianDecision 검증(one-shot nonce, 초안 인터페이스)→`trading.Service` 위임→결과 확정. raw mutator 미노출 [High-risk]
- [ ] 2.6 [T][High-risk] retry matrix 표 작성(스펙 산출물, 보수 기본값) + 구현: mutation 무재시도, 조회 예산·jitter, 429 Retry-After, 401/403 즉시 차단, staleness 진입 차단
- [ ] 2.7 [T][High-risk] IN_DOUBT 해소 엔진: fingerprint(OPEN+CLOSED pagination 완주) 안정화 N회 + 잔고/보유 delta 교차 확인, UNRESOLVED_IN_DOUBT 운영자 해소, 심볼당 in-flight 1건 제한, 2페이지 주문 발견 테스트
- [ ] 2.8 [T][High-risk] AMEND/CANCEL IN_DOUBT 해소: OrderByID(원주문) + 심볼 범위 후계 주문 스캔, lineage 트랜잭션 기록
- [ ] 2.9 [T] `official.Orders` pagination 완주 어댑터(cursor loop·max-pages 방어) — upstream 파일 수정 없이 execgw 측 구현
- [ ] 2.10 [T] fail-closed 분기: interactive auth·USD 잔고 부족·미지원 유형 거부 + reason-code enum 정의 (fixture 테스트)

## 3. 체결 감지 + Reconciliation

- [ ] 3.1 [T][High-risk] `internal/filldetect` 폴링 권위 루프: 미체결(pagination)+OrderByID+잔고, SLO 측정점(broker-visible→local-commit) fake clock 테스트, outage 분류, 위반 시 진입 차단
- [ ] 3.2 [T][High-risk] 누적 스냅샷 멱등 반영: 양의 delta만, 감소·역순 fail-closed, 중복 수신 무변화 테스트
- [ ] 3.3 [T][High-risk] SSE 힌트 소비자: push.Listener(internal/push/listen.go) 연결, 토픽 coalescing·최소 간격, 폴링 파이프 합류 [Pre-Edit 불요 — 신규 소비자]
- [ ] 3.4 [T][High-risk] `internal/reconcile` 스냅샷 계약: 고정 순서(미체결→보유→잔고)·as-of·부분 실패 폐기, decimal 문자열 비교+epsilon, 안정화 간격, external provenance 분류
- [ ] 3.5 [T][High-risk] 재시작 복구 시퀀스: journal 해소→계좌 조회→상태 재구성→완료 전 주문 거부, crash/restart 테스트
- [ ] 3.6 [T] 불일치 처리: 진입 차단·청산 유지, 재대사 간격·카운터 리셋, 영구 불일치 + 차단 범위 상태표

## 4. 엔진 안전 + 관측성

- [ ] 4.1 [T][High-risk] cancel/amend 사전 확인 대체: OrderByID 파생 기반, WTS 세션 nil/만료 상태 cancel/amend 성공 테스트
- [ ] 4.2 [T] 기동 인터록: 게이트 기본 OFF, ON 시 Guardian(초안 인터페이스, Phase 2 확정 예정) + attestation(존재·미만료·계좌 일치) 검증 실패 시 기동 거부, config 스키마 추가(별도 커밋) + audit 로그
- [ ] 4.3 [T] `internal/obs`: 구조화 로그(전이·reconcile·오류, 셀 수 있는 이벤트 규약), ntfy 알림(등급화: critical은 journal DB outbox 경유 재전송·지속 실패 시 진입 차단, heartbeat publish)
- [ ] 4.4 [T][High-risk] flatten-all 1/2 — cancel-all saga: 진입 차단→미체결 각각 취소·확정(IN_DOUBT 규칙), journal 기록, crash 재개 테스트
- [ ] 4.5 [T][High-risk] flatten-all 2/2 — reduce-only 청산: 계좌 재조회 안정화→매도가능수량 기준 공격적 limit 청산→반복 reconcile, `--dry-run`(mutation 0), 확인 문자열(마스킹 계좌·포지션 수·예상 수량·nonce·TTY 전용), oversell 방지 테스트
- [ ] 4.6 [T] 테스트 격리 가드: 격리 config 디렉터리 헬퍼(t.Setenv), 실 Toss hostname POST hard-fail transport 가드 테스트 — 완료 게이트 포함

## 5. 완료 게이트 [M]

- [ ] 5.1 diff 리뷰: upstream 테스트 650 green + characterization 통과, shim 최소성
- [ ] 5.2 High-risk task Pre-Edit 선언·race(-race)/crash 테스트 확인
- [ ] 5.3 `make gate CHANGE=harden-execution-base`
- 5.4 (사용자 확인 후) archive — 단, verify-execution-capability 완료 전 자동화 게이트 ON 불가
