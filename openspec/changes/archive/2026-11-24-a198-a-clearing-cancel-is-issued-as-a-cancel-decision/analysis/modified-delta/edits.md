# MODIFIED 의도 편집 목록

- 원문: openspec/specs/order-execution/spec.md 의 `### Requirement: 취소의 일시적 거절 처리` 블록(75 줄)
  - 편집 1: Scenario 「다른 사유의 409」: 엔진 취소의 세 code 예외를 WHEN 에, THEN 을 「그 밖의 경우」로
  - 결과 diff: 바뀐 줄 원문 2 → 2, 그 밖의 차이 0(역치환 단언)

- 원문: openspec/specs/order-execution/spec.md 의 `### Requirement: IN_DOUBT 해소` 블록(223 줄)
  - 편집 1: D9: 기동 ACKED 문단 제목에 취소
  - 편집 2: D9: 기록 번호 있는 ACKED 취소는 그 번호로 접수 확정, 「그 외」 목록에서 취소 제거
  - 편집 3: D9: 추론 금지의 대상을 발주·정정으로(취소는 위 문장 — 추론 없는 쓰기 완료)
  - 편집 4: D10: 세션 해소 대상 문단 — 취소는 원장 종결 기록으로 닫음, 알림은 발주만
  - 편집 5: D10: 시나리오 「세션 중에는 진행 중인 전송을 종결하지 않는다」
  - 편집 6: D9: 시나리오 「재시작 뒤 ACKED로 남은 취소」 추가(기존 시나리오 앞)
  - 결과 diff: 바뀐 줄 원문 6 → 10, 그 밖의 차이 0(역치환 단언)
