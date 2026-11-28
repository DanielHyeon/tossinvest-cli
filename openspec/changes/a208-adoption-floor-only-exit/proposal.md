# a208 — 편입 보유는 부분익절 없이 보호선만 올린다(손절은 그대로, 유예 없음)

## Why

사용자는 브로커 앱에서 산 종목을 TossOS 가 편입(adoption)해 청산을 맡게 한다. 사용자 판단: 편입 보유가 잘 관리되지 않는다 — StockOS 처럼
관리해 달라. 사용자는 「편입 1번」을 승인했다(2026-11-28, Manager 전달): 편입 전용 청산 정책 · 부분익절 없음 · 보호선만 상향 · 기본 OFF ·
OFF 는 지금 동작과 정확히 같음. 손절 기준은 지금대로 **편입 시점 관측가, 그 아래 2% 이상**을 유지한다 — StockOS 의 평균단가 기준(편입 즉시
매도할 수 있는 알려진 결함)으로 바꾸지 않는다.

**사람 결정(2026-11-28, design 10)**: Q1 → 60초 보호선 상향 유예는 **넣지 않는다**(부분익절이 없는 정책에서 유예가 막는 것은 보호선 상향뿐이라
덜 보수적이고, 새 스냅숏 필드는 롤백 위험). Q2 → 모든 편입이 기존 RUNNER 편입 사본을 쓴다. Q3 → 설정 파일 토글 + 콘솔 저장 이월 + 기동
audit, 콘솔 체크박스 없음.

현재 코드의 사실(AST 근거 `analysis/ast-evidence/claims.md`):

- 편입 기록의 정책은 엔진 공통 정책을 그대로 복사한다(`internal/app/engine/adoption.go:392`, C1). 편입만의 정책을 고르는 갈래는 없다.
- 부분익절 없는 편입 사본은 공통 정책이 RUNNER 일 때만 있다(`internal/exitpolicy/common_policy.go:111`, C4). BALANCED · HYBRID_50 ·
  RATCHET(공통 정책 없음) 편입 포지션은 엔진 진입 포지션과 같은 부분익절을 받는다(C4 · C8).

StockOS 원문 재확인(`analysis/stockos-reference.md`, `fcad869e`): StockOS 도 부분익절 제거는 **RUNNER 편입에만** 한다
(`auto_exit_execution.py:6796-6807`). 그러므로 「모든 편입에 부분익절 없음」은 StockOS 보다 한 걸음 더 나간 사용자 결정이다. StockOS 의
60초 settle window 는 이식하지 않는다(Q1).

## What Changes

1. **토글** `engine.adoption.floor_only_exit`(bool, 기본 false). 참이면 **자동 편입**(`adoptOne`)이 공통 정책 대신 `COMMON_LADDER_RUNNER` 를
   편입 기록에 적는다 → 이미 있는 편입 RUNNER 사본(부분익절 0, 전량 익절 없음, 고정 정체성 `1.0.0-adopted.1`)으로 exit state 가 열린다.
   새 정책 표 · 새 digest · 새 평가 입력 · 스키마 변경 0.
2. **설정 · audit**. 설정 블록 읽기 · 병합 · 저장에 필드를 옮기고, 콘솔 저장이 그 값을 이월하며, 엔진 기동 설정 기록 끝에 한 행을 덧붙인다.
3. **편입 알림 문구**. 토글이 참일 때의 편입 알림은 「보호선 승격과 손절만 적용되고 중간 익절은 하지 않는다」를 말한다. 거짓이면 지금 문구 그대로.

토글이 거짓이면: 편입 기록 정책 ID · 알림 문구 · 설정 파일 바이트가 지금과 같다(design 5). 바뀌는 것은 엔진 기동 audit 에
`engine.adoption.floor_only_exit=false` 한 행이 처음 한 번 남는 것뿐이다. 사다리 평가기 · 관측기 · 손절 경로는 편집하지 않는다.

## Non-goals

- 편입 직후 보호선 상향 유예(StockOS settle window) — 사람 결정 Q1(b)로 기각.
- 공통 정책별 편입 사본(BALANCED · HYBRID 잠금 유지) — 사람 결정 Q2 로 기각.
- 종목별 정책 선택, 편입 해제 · 즉시 청산 버튼(사용자 항목 3, 이후).
- 편입 전용 손절 설정(사용자 항목 2, 다음 change). 결정만 「관련/다음」에 기록: 유예 기본 0분, 재난 손절(disaster stop) 항상 켬.
- 운영자 재편입(READOPT, `internal/app/engine/position_policy_command.go:359-366`) — 정책 선택은 지금 그대로, 토글을 따르지 않는다.
- 콘솔 화면에 토글 체크박스 추가(사람 결정 Q3).
- 이미 편입된 포지션의 정책 변경(편입 기록 스냅숏 규칙 — 바뀌지 않음).
- 손절 기준 변경(평균단가 기준 도입 금지).

## Impact

- High-risk: 편입 정책 선택(`adoption.go` `adoptOne`) — 익절 · 보호선 정책을 정하는 자리. 손절 갈래 · 사다리 평가기 · 관측기는 편집하지 않는다.
- 설정 · 콘솔: `internal/config/engine.go`, `internal/config/adoption_io.go`, `internal/console/settings.go`(값 이월 한 줄).
- audit: `internal/app/engine/interlock.go` `recordGateSettings` 끝 행 + `a163_source_age_audit_test.go` 의 「마지막 행」 단언 갱신.
- 스키마 · journal · 저장 스냅숏 변경 없음. 롤백 순서 제약 없음(design 7).
- 브로커 호출 · rate limit 예산 변화 없음.
