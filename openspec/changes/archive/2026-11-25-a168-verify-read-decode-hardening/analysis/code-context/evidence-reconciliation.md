# a168 증거 조정 (tasks 0.3)

기준 commit `e57d06cd6`. 세 출처를 현재 HEAD 로 대조했다.

| 질문 | CodeGraph 1.6.0 | HEAD grep / AST | 결론 |
| --- | --- | --- | --- |
| 네 대사 리더의 생산 호출자 | cmd 위임 1씩 | `rtk proxy grep` 동일(시험 외 cmd 위임만) · AST 호출 목록(`analysis/ast/tossctl--reconcilereads.*`) | 일치 |
| 위임의 호출 자리 | (함수 단위만 답함) | AST: `Reconcile` 6 자리(줄 249·258·262·266·271·277), `readReconcileSet` 3, 그룹 함수 1씩 | 자리 수는 AST 로 셈(design §4) |
| `unwrapAndDecode` 위치 | 228 | HEAD 228–243 | proposal 의 213 은 낡음(design §1 P7) |
| `send` 위치·2xx 판정 | 335 | AST B9 줄 377 · B4 줄 359 | proposal 의 320·362·344 는 낡음(P8) |
| 상태 코드를 리더가 볼 통로 | (질의 대상 아님) | `doRequest` 줄 220 `observeAttempt` · `trace.go` `WithAttemptObserver` · 선례 m0_reads/strict_quote/strict_minute | Q1 (a) 가능, `send` 편집 0(design §2) |
| a166 배선 census 대상 여부 | — | `shapes.json` 에 편집 대상 함수 없음(grep) | 영향 없음 |
| a121 봉인 census | — | `reconcile_structure_test.go:44` glob 이 `internal/official/reconcile*.go` 포함 · type assertion 금지 | 새 파일도 대상 — 관측자 체인 안 함(design D5) |

CodeGraphContext(보조): `codegraphcontext-context.md` 참조 — 판정 근거로 쓰지 않음. 불일치는 줄 번호 낡음뿐이고 현재 HEAD·AST 로 해소했다.
편집 차단 사유 없음.
