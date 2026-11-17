# Evidence reconciliation (a193, 1단계)

기준 HEAD `791fb6ef1`. CodeGraph 1.6.0(worktree 색인)·AST 40·`rg` 를 대조했다.

| 질문 | CodeGraph | HEAD(`rg`·직접 읽기) | 결론 |
|---|---|---|---|
| `execgw.Resolver.Resolve` 의 production 호출자 | 이름 단위 15(다른 `Resolve` 섞임) | `rg '\.Resolve\(ctx'`(시험 제외) 2: recovery.go:300, flatten.go:428 | 2 — 세션 중 호출자 0 |
| 엔진 조립의 `Resolver` 인스턴스 사용자 | — | `rg 'c\.Resolver\|ectx\.Resolver'`: runtime_wiring.go:176·183 만 | `Context.Recovery` 하나 — 세션 실행자가 둘째 |
| exit 제출의 IN_DOUBT 가 outbox 에 가나 | `noteExitOrder` 호출자 다수 | AST: append@219 뿐(Notify·RecordCritical 호출 0) | 로그 줄만 — 정본 「등급화된 알림」과 차이(Q-A) |
| `ReasonUnresolvedInDoubt` 래치를 푸는 자 | — | `rg '\.Clear\('`: 그 이유를 푸는 자리 0 | 프로세스 수명 래치 — D5 공시 |
| 보조 실행자 집합을 고정한 시험 | `engineRuntime` 시험 호출자 4 | `rg 'AuxiliaryNames\(\)'`: a098_the_engine_registers_a_sender_test.go:28 하나 | 2단계에서 그 시험 함수 편집(FLM 번들) |
| a166 census 와 겹치나 | — | shapes.json 키 50 에 `engineRuntime` 없음; 새 파일은 census 선택자 미사용 | 재기준화 불필요 |
| flatten 이 IN_DOUBT PLACE 를 해소하나 | `resolveStep` 호출자 | `Saga.sell` B10(case@516) 은 step 상태만 기록; `resolveStep` 은 취소 step 만(`settleStep`) | 세션 실행자와 같은 attempt 를 다투지 않음 |

불일치: 없음(CodeGraph 의 이름 단위 과다 집계는 HEAD 로 좁혔고 결론을 바꾸지 않음).
