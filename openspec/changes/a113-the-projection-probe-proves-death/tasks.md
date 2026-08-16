# Tasks — a113-the-projection-probe-proves-death

**이 change는 a109 issues I1의 이행으로 등록만 먼저 했다(2026-08-16). 착수 선행 조건 없음.**

## 0. 착수 전 조건

- [ ] 0.1 base 재고정(`tools/sdd/capture_change_base.py --change a113-…`) — 등록 시점
  base는 착수 시점과 다르다.
- [ ] 0.2 `projectionSocketAccepts`와 그 호출자(회수 경로)의 Function Logic Map을
  **proposal 갱신 전에** 만든다. 등록 문서의 분기 주장은 전부 a109 기록(issues I1,
  review §1 A1 P1-A) 인용이며, 착수 시점 코드의 재검증이 선행이다.

## 1. 구현 (착수 시 상세화)

- [ ] 1.1 RED: 쓰기 비트가 깎인 산 socket이 삭제되지 않음을 요구하는 테스트
  (a109 `TestTheProbeRefusesASocketThatChangedUnderIt` 원형판).
- [ ] 1.2 GREEN: owner-write 절 삭제 + chmod-then-probe 교체.
- [ ] 1.3 뮤테이션: F1-N1 원형판 적용·격추 확인.
