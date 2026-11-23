# 증거 조정 (a199 1단계)

| 질문 | CodeGraph 1.6.0 | 현재 HEAD(AST · grep · go/types) | 결론 |
|---|---|---|---|
| 인터록 호출 사슬 | verifyGate←runInterlock←NewContext←{assembleEngine, New} | AST calls 로 자리 1·1·1·1 | 일치 |
| 청산 루프 생성자의 호출자 | engineRuntime 하나(생산) | engineRuntime AST @662·@671 | 일치 |
| `Verified` 소비자 | (CodeGraph 는 필드 읽기를 안 줌) | go/types 전수 8 자리(`../verified-census/sites.txt`) | HEAD 로 확정 |
| 감시 판정의 호출자 | prepare(생산) | prepare AST @323 | 일치 |
| `AllReasonCodes` | execgw·risk 의 같은 이름을 섞음 | grep: execgw 판 비시험 0 | HEAD 로 해소 |
| 사전 증거(scratch, 9738a0c32)와 현재 | — | AST 12개 sha·줄·분기 동일(`../cited-ast/summary.txt`) | 일치 |

편집 차단 불일치: 없음.
