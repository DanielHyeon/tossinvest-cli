# Branch Test Map: `engineRuntime`

- Source: `cmd/tossctl/engine.go` (618-705); **편집 뒤** 측정 — `branch_coverage.py`, 연결 워크트리 `c6e2e3ac`, `./cmd/tossctl` 시험 3개
  (`EngineRuntime|TestA092`)를 하나씩(`analysis/harness/coverage-post-engineruntime.json`). 편집 전 표(시험 10개 표본)는 `analysis/pre-edit/unit2/`.
- 재번호 없음: 편집은 분기가 아닌 리터럴 키 하나다(AST의 분기 좌표 6개가 편집 전과 같음).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 621:2 | 체결 감지기 생성 실패 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개(편집 전과 같음) |
| B2 | if at 630:2 | 대사 드라이버 생성 실패 | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` | 해당 없음(분기 불변) | 블록 630.16-632.3, PASS |
| B3 | if at 641:2 | exit 관측기 생성 실패 | 같음 | 해당 없음(분기 불변) | 블록 641.16-643.3, PASS |
| B4 | if at 646:2 | 복구 생성 실패 | 같음 | 해당 없음(분기 불변) | 블록 646.16-648.3, PASS |
| B5 | if at 650:2 | 전략 진입 감독자 생성 실패 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B6 | if at 662:2 | 배달 실행자 생성 실패 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
