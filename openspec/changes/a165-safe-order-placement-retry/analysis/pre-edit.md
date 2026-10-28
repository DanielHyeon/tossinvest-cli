# a165 Pre-Edit Gate (2.1)

## 선언

```text
Pre-Edit Gate:
- change id / task id: a165-safe-order-placement-retry / 2.1
- 대상 심볼(패키지.함수):
  - internal/execgw.Gateway.submit (gateway.go:466): 비전략 DispatchVerified 인자를 취소 거절 바깥 래퍼로 교체함. 원 dispatch 클로저와 strategy 분기는 보존함.
  - internal/execgw.New (gateway.go:148): CancelRefusal 정책 필드만 Gateway로 이전함.
  - internal/execgw.Gateway.checkSymbolFree (gateway.go:798): B3 노출 증가 반환 뒤, 기존 unresolved 검사 뒤에 D11 원장 latch를 추가함. unsettledFor는 편집하지 않음.
  - internal/app/engine.buildGateway (gateway.go:245): EngineCancelRefusal 정책을 주입함.
  - cmd/tossctl.buildFlattenWiring (flatten.go:199): FlattenCancelRefusal 정책을 주입함.
  - 새 leaf: internal/execgw의 already-processing 판정·정책·바깥 래퍼·D11 latch, internal/journal.Journal.ConfirmedCancelAfterLatch. 기존 Journal.ConfirmedCancelOf는 편집하지 않음.
- CodeGraph definition/callers/callees/impact:
  - CodeGraph 1.6.0 current read-only DB에서 위 다섯 정의 좌표를 확인함. submit=466..762, checkSymbolFree=798..829, New=148..189, buildGateway=245..370, buildFlattenWiring=199..272.
  - 기존 hard-evidence 원문 analysis/code-context/raw/submit.txt의 callers는 Gateway.place·Gateway.Cancel·Gateway.Amend이고, callees에는 checkSymbolFree·DispatchVerified 전 단계가 있음. impact에는 execgw Gateway 계열, flatten settleStep, engine exitloop가 포함됨.
  - current CLI status는 stale lock 해제(codegraph unlock 성공) 뒤에도 대기해 CLI callers/callees/impact 재실행은 불가함. DB의 definition 좌표와 현재 AST source_sha256이 일치하므로 historical raw query와 현재 HEAD 사이 코드 불일치는 없음. 재실행 불가 상태는 구현 리뷰에서 재확인함.
- CodeGraphContext 후보와 evidence reconciliation:
  - CodeGraphContext 0.5.1 current stats는 0 nodes/edges로 보조 문맥 unavailable. GBrain은 configured brain 없음.
  - analysis/code-context/evidence-reconciliation.md의 CodeGraph↔HEAD 불일치 0 결론을 현재 AST 24개 source_sha256 일치로 재확인함. CodeGraphContext 부재는 advisory이므로 편집 차단 근거가 아님.
- 기존 동작 파악 근거: design D2~D12, delta spec, review 1~4차, analysis/code-context/, 24개 function-logic AST/FLM/BTM, gateway.go·dispatch.go·fills.go·exit_proposal_release.go·engine/gateway.go·cmd/tossctl/flatten.go·official/client.go 현재 HEAD, retry_test.go·gateway_test.go·a094_exit_clear_test.go·flatten_test.go·orders_write_test.go·client_test.go.
- Function Logic Map / Branch Test Map:
  - openspec/changes/a165-safe-order-placement-retry/analysis/function-logic/internal-execgw--gateway.submit/{function-logic-map.md,branch-test-map.md,ast.json}
  - openspec/changes/a165-safe-order-placement-retry/analysis/function-logic/internal-execgw--new/{function-logic-map.md,branch-test-map.md,ast.json}
  - openspec/changes/a165-safe-order-placement-retry/analysis/function-logic/internal-execgw--gateway.checksymbolfree/{function-logic-map.md,branch-test-map.md,ast.json}
  - openspec/changes/a165-safe-order-placement-retry/analysis/function-logic/internal-app-engine--buildgateway/{function-logic-map.md,branch-test-map.md,ast.json}
  - openspec/changes/a165-safe-order-placement-retry/analysis/function-logic/cmd-tossctl--buildflattenwiring/{function-logic-map.md,branch-test-map.md,ast.json}
  - 위 5개를 포함한 24개 번들은 HEAD source_sha256과 일치함. checkSymbolFree baseline coverprofile은 66.7%임.
- upstream 상속 테스트 영향: yes. submit은 place/cancel/amend 공용 mutation 문이고, engine clear와 flatten settle 경로가 영향권임. 새 테스트에서 place·amend·정책 nil 비회귀, engine 같은 주기 지연 0, flatten IN_DOUBT 보류를 고정함.
- 실패 테스트 선행 작성: yes. 기존 시험 함수·fake는 수정하지 않고 새 a165 전용 *_test.go 파일에 R3/R4/R5, W1~W4, D11 (a)~(k) RED를 모두 작성함. 새 fake도 새 시험 파일에만 둠.
- 설정·DB·journal 변경과 rollback: runtime flag·schema·migration·dependency·license 변경 없음(schema 36 유지). journal은 새 read-only 질의만 새 파일에 추가함. rollback은 feature branch의 task 단위 커밋 되돌림이며 운영 설정·원장 데이터 조작 없음.
- 안전 불변식 §0 위반 여부 검토: 통과 예정. 409 재시도는 cancel+명시 정책에만 한정함. place/amend·정책 nil은 오늘 동작, 엔진 취소는 재시도·대기 0, flatten 소진/중단은 AMBIGUOUS로 되돌림. D11은 raisesExposure 뒤에만 두어 cancel·청산 SELL·flatten·보호 경로를 막지 않음. live 주문·운영 토글·mutating command 실행 없음.
```

## HEAD 신선도

| 파일 | SHA-256 | 근거 |
| --- | --- | --- |
| `internal/execgw/gateway.go` | `b4b020607839345a70e7658ff488aef4282de4335eedd83b90a93520d80e53df` | `submit`·`New`·`checkSymbolFree` |
| `internal/app/engine/gateway.go` | `497ec967c114b5e7a9c1f9a80f129b381b160252dc0296ff71f95aca983d57e1` | `buildGateway` |
| `cmd/tossctl/flatten.go` | `3a2b20f4c145d7f152dad06a5ffd48e2921b05a5d2ec0ed496c6dfe7fdca3b27` | `buildFlattenWiring` |
| `internal/journal/dispatch.go` | `42165f21f0707b4d6424907acb386e3ec6475402d975e8c09e40ddd0cd058536` | `DispatchVerified` 정착 ctx 경계 |
| `internal/journal/exit_proposal_release.go` | `8faee2a9f9f1695d0d8293096a360230ab7ae700b1890127d3c8b12c3c087e3a` | `ConfirmedCancelOf` 불변 |
| `internal/journal/fills.go` | `92bc44f947eead2a045815b29358122d2fd0dd2ffd3e5bada4f2dc44eaa5c4c5` | r1 `LiveOrdersForSymbol` |
| `internal/official/client.go` | `6d145916a6797f3a32842561346c8767585a951d8883ebd81bf8af4632106639` | R5 코드 불변 |
| `internal/trading/service.go` | `631b2f02437e510698ffc06a1124bc11176d005a61956a61b4560e874d1e9f34` | broker cancel 경계 |

## 시험 기준선

- 실행 전 `pgrep -af 'go test|make .*gate'`에는 이 선언의 자체 검사 프로세스 외 실행 중인 시험/gate가 없었음.
- 실행: `go test ./internal/execgw -coverprofile=/tmp/a165-preedit-cover.out`.
- 측정: `Gateway.checkSymbolFree` 66.7%, package total 73.2%.
- 상태: test binary는 coverprofile을 완성했지만 상위 `go test` 프로세스가 종료 상태를 반환하지 않고 25분 이상 futex 대기했음. 해당 비운영 진단 프로세스만 TERM으로 종료함. PASS/FAIL은 주장하지 않으며, 후속 a165 시험은 정확한 `-run` 범위로 분리 실행함.

## 결정

- research-first 결론: 기존 `journal.DispatchFunc`·`clock.Clock`·표준 `context.WithDeadline`를 확장함. 외부 backoff 의존성은 도입하지 않음.
- `impeccable`: N/A — backend/documentation-only. 신규 UI·운영 콘솔·명령을 추가하지 않음.
- `grill-me`: N/A — Manager가 2026-10-28 proposal-freeze 1.3을 최종 승인했고, 본 단계는 승인된 계약의 구현임.
