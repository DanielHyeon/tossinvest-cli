# a208 — 편입 보유는 부분익절 없이 보호선만 올린다(손절은 그대로, 유예 없음)

## Why

사용자는 브로커 앱에서 산 종목을 TossOS 가 편입(adoption)해 청산을 맡게 한다. 사용자 판단: 편입 보유가 잘 관리되지 않는다 — StockOS 처럼
관리해 달라. 사용자는 「편입 1번」을 승인했다(2026-11-28, Manager 전달): 편입 전용 청산 정책 · 부분익절 없음 · 보호선만 상향 · 기본 OFF ·
OFF 는 지금 동작과 정확히 같음. 손절 기준은 지금대로 **편입 시점 관측가, 그 아래 2% 이상**을 유지한다 — StockOS 의 평균단가 기준(편입 즉시
매도할 수 있는 알려진 결함)으로 바꾸지 않는다.

**사람 결정(2026-11-28, design 10)**: Q1 → 60초 보호선 상향 유예는 **넣지 않는다**(부분익절이 없는 정책에서 유예가 막는 것은 보호선 상향뿐이라
덜 보수적이고, 새 스냅숏 필드는 롤백 위험). Q2 → ~~모든 편입이 기존 RUNNER 편입 사본~~ **F3 결정(2026-11-28, 사용자 「F3 2」)으로 대체**:
편입은 설정된 공통 정책의 보호선 표(단계 트리거 · 보호선 · 러너 트레일)를 그대로 쓰고 부분익절 비율만 0 인 사본을 쓴다 — RUNNER 표로 바꾸면
운영 공통 정책(HYBRID_50)보다 보호선이 느슨해져 보수 방향이 아니었다(리뷰 F3). Q3 → 설정 파일 토글 + 콘솔 저장 이월 + 기동
audit, 콘솔 체크박스 없음.

현재 코드의 사실(AST 근거 `analysis/ast-evidence/claims.md`):

- 편입 기록의 정책은 엔진 공통 정책을 그대로 복사한다(`internal/app/engine/adoption.go:392`, C1). 편입만의 정책을 고르는 갈래는 없다.
- 부분익절 없는 편입 사본은 공통 정책이 RUNNER 일 때만 있다(`internal/exitpolicy/common_policy.go:111`, C4). BALANCED · HYBRID_50 ·
  RATCHET(공통 정책 없음) 편입 포지션은 엔진 진입 포지션과 같은 부분익절을 받는다(C4 · C8).

StockOS 원문 재확인(`analysis/stockos-reference.md`, `fcad869e`): StockOS 도 부분익절 제거는 **RUNNER 편입에만** 한다
(`auto_exit_execution.py:6796-6807`). 그러므로 「모든 편입에 부분익절 없음」은 StockOS 보다 한 걸음 더 나간 사용자 결정이다. StockOS 의
60초 settle window 는 이식하지 않는다(Q1).

## What Changes

1. **토글** `engine.adoption.floor_only_exit`(bool, 기본 false). 참이면 **자동 편입**(`adoptOne`)이 공통 정책 대신 그 공통 정책의
   **편입 보호선 사본** ID 를 편입 기록에 적는다(F3 결정).
   - 공통 정책 HYBRID_50(운영) → 새 사본 `COMMON_LADDER_HYBRID_50_ADOPTED_FLOOR_ONLY`(고정 정체성 `1.0.0-adopted.1` · 고정 digest). 표는 등록된
     HYBRID_50 에서 유도해 트리거 · 보호선 · 러너 트레일이 같고 부분익절만 0. 공통 정책 등록부 밖이라 설정 · 콘솔 · override 에서 고를 수 없다.
   - 공통 정책 RUNNER → `COMMON_LADDER_RUNNER` 그대로(기존 편입 RUNNER 사본이 이미 「RUNNER 표 + 부분익절 0」 — 새 정체성 불필요).
   - 공통 정책 BALANCED · 없음(RATCHET) → **토글 무시 = OFF 와 같은 동작**(fail-closed) + 기동 audit 행 값 `ignored` · 사유.
   스키마 · 평가 입력 변경 0. 원장은 사본 ID 하나만 더 받는다(편입 기록 · 편입 포지션 exit state).
2. **설정 · audit**. 설정 블록 읽기 · 병합 · 저장에 필드를 옮기고, 콘솔 저장이 그 값을 이월하며, 엔진 기동 설정 기록 끝에 한 행을 덧붙인다
   (값 `true` · `false` · `ignored`).
3. **편입 알림 문구**. 사본이 적용된 편입의 알림은 「보호선 승격과 손절만 적용되고 중간 익절은 하지 않는다」를 말한다. 적용되지 않으면 지금 문구 그대로.

토글이 거짓이면: 편입 기록 정책 ID · 알림 문구 · 설정 파일 바이트가 지금과 같다(design 5). 바뀌는 것은 엔진 기동 audit 에
`engine.adoption.floor_only_exit=false` 한 행이 처음 한 번 남는 것뿐이다. 사다리 평가기 · 관측기 · 손절 경로는 편집하지 않는다.
이미 편입된 포지션(편입 HYBRID_50 행 포함)의 판정은 바뀌지 않는다 — 기존 정책 ID 의 해소 갈래는 무편집.

## Non-goals

- 편입 직후 보호선 상향 유예(StockOS settle window) — 사람 결정 Q1(b)로 기각.
- BALANCED 의 편입 사본 — 최종 rung 이 전량 익절이라 「부분익절만 0」이 보호선만의 정책이 되지 않음(F3 결정 범위 밖, 토글 무시).
- 종목별 정책 선택, 편입 해제 · 즉시 청산 버튼(사용자 항목 3, 이후).
- 편입 전용 손절 설정(사용자 항목 2, 다음 change). 결정만 「관련/다음」에 기록: 유예 기본 0분, 재난 손절(disaster stop) 항상 켬.
- 운영자 재편입(READOPT, `internal/app/engine/position_policy_command.go:359-366`) — 정책 선택은 지금 그대로, 토글을 따르지 않는다.
- 콘솔 화면에 토글 체크박스 추가(사람 결정 Q3).
- 이미 편입된 포지션의 정책 변경(편입 기록 스냅숏 규칙 — 바뀌지 않음).
- 손절 기준 변경(평균단가 기준 도입 금지).

## Impact

- High-risk: 편입 정책 선택(`adoption.go` `adoptOne` — 새 파일 함수 경유), 판정 표 해소(`exitpolicy.CommonLadderForPosition` · `LegacyLadderPolicyIdentity`
  — 사본 ID 갈래 추가, 기존 갈래 무편집), 원장 정책 ID 문(`journal.AdoptionRequest.record` · `Journal.OpenExitState` — 사본 ID 하나 추가).
  손절 갈래 · 사다리 평가기 · 관측기는 편집하지 않는다.
- 설정 · 콘솔: `internal/config/engine.go`(필드 · `mergeEngine` 무시 사유 한 줄), `internal/config/adoption_io.go`, `internal/console/settings.go`(값 이월 한 줄).
- audit: `internal/app/engine/interlock.go` `recordGateSettings` 끝 행 + `a163_source_age_audit_test.go` 의 「마지막 행」 단언 갱신.
- 스키마 · journal migration · 저장 스냅숏 변경 없음. **롤백 순서 제약 있음**(design 7 개정 2): 사본 ID 를 모르는 옛 이미지는 그 포지션의 판정(손절 포함)을 거부한다.
- 브로커 호출 · rate limit 예산 변화 없음.
