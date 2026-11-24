# a201 evidence reconciliation (tasks 1.3 · 1.6)

| 질의 | CodeGraph 1.6.0 | AST(`tools/logic-map`) · grep · 실행 | 결론(HEAD `86a131a51`) |
|---|---|---|---|
| RunOnce 의 생산 호출자 | `Run` 1 | grep 1 자리(:412) | 일치 |
| Observe 의 생산 호출자 | 20(이름 충돌) | grep 1 자리(:464) | CodeGraph 과다 — 이름 충돌. grep · AST 채택 |
| ConvergeQuantities 의 호출자 | RunOnce + 시험 | grep 1 자리 | 일치 |
| 수렴의 FillWatermark 호출 | 안 잡힘(인터페이스 필드) | AST call@205 | CodeGraph 누락 — AST 채택 |
| 사전 증거(7fec501f1)의 AST 20 | — | 재추출 20 바이트 동일(`../cited-ast/`) | 사전 증거 좌표 전부 유효 |
| W2 · W3 고착 | — | 실제 RunOnce 재현 6 PASS(`../repro/`) | 사전 증거의 손밟기 재현과 같은 결론 + (b) 단독 stale 입력 · 판 중단 뒤 종목 · 앞 크레딧 반박 3 건 추가 |

불일치는 전부 「CodeGraph 가 필드 · 인터페이스 호출과 동명 메서드를 못 가름」 한 종류이고 현재 HEAD 의 AST · grep 으로 해소함. 편집을 막는 미해소 불일치 0.
