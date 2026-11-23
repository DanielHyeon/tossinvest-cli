# 구현 적대 리뷰 r3 — codex 마지막 메시지(원문, HEAD e6383786d)

명령 프롬프트: 같은 디렉터리 `codex-prompt.txt`. 전체 로그(398 KB, 세션 출력)는 커밋하지 않음 — 아래가 마지막 메시지 원문이다(링크 경로는 저장소 절대경로를 줄임).

**판정: 조건부 P0 위험 1건. D10의 "뒤에서 매도를 다시 검사하므로 안전" 논증은 성립하지 않습니다.** HEAD `e6383786d` 확인, `internal`의 HEAD 대비 미커밋 차이 없음.

1. **[P0] legacy 종결 증거와 scoped 미종결 증거가 공존하면 살아 있는 매도를 치워진 것으로 판단할 수 있습니다.**
   - 조건: 대상의 유일한 소유 intent, 소유 시각 이후 기록된 범위 없는 legacy `CANCELLED/terminal=1`, 같은 대상의 scoped `OPEN_*/terminal=0`.
   - D10은 처음부터 **terminal 행만** 읽으므로 scoped OPEN이나 비종결 `fail_closed=1`을 충돌 증거로 보지 않습니다. (a198_cancel_target_close.go:103, exit_proposal_release.go:290)
   - 두 행은 공존 가능합니다. scoped 쓰기의 legacy mirror는 범위가 일치해야 갱신하고, UNION도 범위가 일치해야 legacy를 제외합니다. (fills.go:795, fills.go:888)
   - D10이 CANCEL을 CONFIRMED로 닫으면 미종결 가드가 풀립니다. 뒤의 `LiveOrdersForSymbol`도 **동일한 legacy 종결 증거로 매도를 목록에서 제거**하므로 `ConfirmedCancelOf`의 매도 대기 분기까지 도달하지 않습니다. (fills.go:1882)
   - 공유 판독의 기존 취약점이지만, D10이 남아 있던 차단까지 해제합니다. 실브로커 초과 체결은 미입증이며, **중복 매도 방지 논증의 반례**입니다. scoped 증거와 legacy 증거의 충돌 처리를 두 경로에서 함께 고쳐야 합니다.

2. **판정 방향·수량:** 표 자체는 맞습니다. CANCELLED 계열→CONFIRMED, FILLED/REJECTED 계열→FAILED_CONFIRMED이며 부분체결 취소도 포함합니다. (a198_cancel_target_close.go:40)
   D10은 체결 수량을 다시 쓰지 않습니다. FAILED일 때 해제하는 예약도 **취소 attempt에 연결된 것만**입니다. (a198_cancel_target_close.go:167, reservation_release.go:130)
   scoped 거래일·시장·방향 불일치, 복수 소유 intent는 닫힘을 막습니다. AMEND가 받은 후계 번호도 소유 후보입니다. 외부·사람 주문은 소유 후보가 없으면 D10 대상이 아닙니다.

3. **트랜잭션:** 읽기와 쓰기 사이 다른 연결의 writer 개입·부분 커밋 결함은 발견하지 못했습니다.
   DSN은 `BEGIN IMMEDIATE`, D10은 `WHERE state=IN_DOUBT`, 상태·전이·예약 해제를 한 트랜잭션으로 처리합니다. (journal.go:225, a198_cancel_target_close.go:150)
   Resolver/D9의 기존 `transition`도 DB 상태를 다시 읽고 from-state를 검사하므로 먼저 닫힌 상태를 덮지 않습니다. (durability.go:559)

4. **source:** 현재 엔진 조립은 `engine`, CLI flatten은 `flatten-all`을 명시합니다. D10은 후자를 닫지 않습니다. 다만 source는 인증된 소유권이 아닌 호출자 문자열이고, 일반 Gateway 기본값도 `engine`입니다. (gateway.go:187)
   빈 문자열은 제외되며 신규 intent는 빈 source를 거부합니다. NULL은 스키마상 금지이고 비정상 NULL 행이면 Scan 오류로 닫지 않습니다. 현재 확인한 운영 조립에서 flatten 오분류는 발견하지 못했습니다.

5. **D9:** 취소 식별자를 대상 번호로 오인하지 않습니다. 기록 번호를 보존하고, `ConfirmedCancelOf`는 `target_order_id`로 조회합니다. 정상 증거에서는 살아 있는 SELL은 대기, BUY는 D8에 따라 치워진 것으로 처리되어 손절 제출이 가능합니다.
   성공 알림 제거는 기존 a094 계약의 변경이지만 D9와 MODIFIED delta에 명시된 변경입니다. 쓰기 실패는 명명 알림 후 false 반환하며 기존 트랜잭션이 롤백합니다.
   **D9는 flatten-source ACKED도 확정합니다.** Q7의 격리는 D10에만 적용됩니다. flatten step을 동시에 갱신하지는 않지만 Resolver는 이미 종결된 attempt를 반환하므로, 재해소 시 이를 읽을 수 있습니다. (indoubt.go:253)

파일/Git 변경·테스트·운영 실행 0. 정적 검토이며 재현 시험은 미실행. ai-memory CLI 조회는 권한 오류로 실패했고, 기존 Graph/AST의 현 HEAD freshness도 검증하지 않았습니다. `impeccable: N/A — backend-only 검토`.
