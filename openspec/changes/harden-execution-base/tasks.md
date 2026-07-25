# Tasks: harden-execution-base

> 실행 주체 — [M]=Manager, [T]=Teammate. High-risk task는 Pre-Edit 선언 필수 (WORKFLOW.md).
> 순서는 의존성 순. 각 task는 TDD(실패 테스트 → 최소 구현 → 통과 → 커밋), 완료 체크는 산출물 커밋과 동일 커밋.

## 1. 리팩터 기반 (shim, upstream 회귀 0)

- [ ] 1.1 [T] `internal/domain`에 MutationResult 이동 + `internal/trading`에 type alias — 전체 테스트 무수정 통과 확인 [High-risk]
- [ ] 1.2 [T] `internal/app` 신설: `newAppContext` 로직 승격(`app.New(Options)`), `cmd/tossctl`에 위임 래퍼 유지 [High-risk]
- [ ] 1.3 [T] `internal/app` 엔진 프로필: official-only 브로커 주입, WTS 쓰기 도달 불가 테스트, WTS 전용 intent 거부
- [ ] 1.4 [T] 조건주문 게이트·intent 조립 `internal/trading` 이동 + `trading.Service`에 Conditional Preview/Execute 메서드, cmd 쪽 shim [High-risk]

## 2. 주문 상태기계 + Journal

- [ ] 2.1 [T] `internal/journal`: SQLite intent journal (스키마, WAL, 단일 락, fuseblk 거부 가드, ext4 경로 해석) — 계약 테스트
- [ ] 2.2 [T] 주문 상태기계 타입·전이 검증 (공식 API status fixture 기반 매핑 테스트, IN_DOUBT/AMEND_IN_DOUBT 포함)
- [ ] 2.3 [T] 제출 파이프: journal 기록 → 제출 → 결과 확정 (기록 실패 시 미제출, 결과 불명 시 IN_DOUBT + 재제출 차단) [High-risk]
- [ ] 2.4 [T] IN_DOUBT 해소 루틴: 미체결+체결+거래내역 조회로 확정, 해소 전 심볼·레인 진입 차단
- [ ] 2.5 [T] retry matrix: 조회 bounded-jitter/mutation 재시도 금지/429 유형별 정책, staleness 임계 진입 차단 [High-risk]
- [ ] 2.6 [T] fail-closed 분기: interactive auth·USD 잔고 부족·미지원 유형 거부 + 사유 기록 (fixture 테스트)

## 3. 체결 감지 + Reconciliation

- [ ] 3.1 [T] `internal/filldetect`: 공식 API 폴링 루프(SLO 설정형), 멱등 상태 반영
- [ ] 3.2 [T] SSE 힌트 소비자: push.Listener 연결, 토픽별 coalescing·최소 간격, 폴링 파이프 합류
- [ ] 3.3 [T] `internal/reconcile`: 대사 계약 구현(비교 키·오차·안정화·external 분류) [High-risk]
- [ ] 3.4 [T] 재시작 복구 시퀀스: journal 해소 → 계좌 조회 → 상태 재구성 → 완료 전 주문 거부 — crash/restart 테스트 [High-risk]
- [ ] 3.5 [T] 불일치 처리: 진입 차단·청산 유지·3회 초과 영구 불일치 표기 + 알림

## 4. 안전 인터록 + 관측성

- [ ] 4.1 [T] 자동화 게이트 설계 구현: 기본 OFF, Guardian 인터페이스 계약 정의, 미주입+ON 기동 거부 boot assertion [High-risk]
- [ ] 4.2 [T] `internal/obs`: 구조화 로깅(주문 전이·reconcile·오류), 핵심 메트릭 카운터
- [ ] 4.3 [T] push 알림 채널(ntfy 기본, 인터페이스 추상화): 지정 이벤트 통지, 실패 시 로그 폴백
- [ ] 4.4 [T] `tossctl` flatten-all 명령: typed-confirmation, 미체결 취소→전량 청산, 진행 출력 [High-risk]

## 5. 실증·검증 (사용자 협조 필요)

- [ ] 5.1 [T] 공식 API capability soak 도구: 자격증명 무인 갱신 N일 기록, rate limit 실측 — 실행은 사용자 환경에서
- [ ] 5.2 [M] 실계좌 주문 경로 1회성 검증 절차 문서(최소 수량·limit-only·즉시 취소, 매도 경계·KR cancel/amend) — 실행·승인은 사용자
- [ ] 5.3 [M] 토스 Open API 약관·자동화 허용 범위 검토 기록 (사용자 협조)

## 6. 완료 게이트 [M]

- [ ] 6.1 diff 리뷰: upstream 테스트 650개 무수정 green, shim 수정 최소성 확인
- [ ] 6.2 High-risk task의 Pre-Edit 선언·race/crash 테스트 존재 확인
- [ ] 6.3 분리 커밋·체크박스 동시성 확인 후 `make gate CHANGE=harden-execution-base`
- 6.4 (사용자 확인 후) archive
