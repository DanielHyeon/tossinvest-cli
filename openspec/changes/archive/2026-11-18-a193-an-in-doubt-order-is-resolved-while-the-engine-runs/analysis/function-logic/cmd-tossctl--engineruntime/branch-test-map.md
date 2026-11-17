# Branch Test Map: `engineRuntime`

- Source: `cmd/tossctl/engine.go` (655-753, 편집 후). 커버(cover mode set — 1 = 실행됨): 편집 전 `../../implementation/base-cover.out`, 편집 후
  `../../implementation/after-cover.out`(둘 다 `go test ./cmd/tossctl -run '<아래 시험들>' -coverprofile`, `GOFLAGS=-trimpath`). 블록 카운트는 편집 전후 같음.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 감지기 생성 실패 | 선택 시험에 이 갈래를 타는 것 없음(표 시험은 B2·B3·B4) | 해당 없음(기존 분기) | 미실행 — body 0(편집 전후 같음); 이 change 가 만든 갈래 아님 |
| B2 | 대사 드라이버 생성 실패 | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` | 해당 없음(기존 분기) | yes — body 1 |
| B3 | exit 관측자 생성 실패 | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` | 해당 없음(기존 분기) | yes — body 1 |
| B4 | 복구 생성 실패 | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` | 해당 없음(기존 분기) | yes — body 1 |
| B5 | 전략 진입 감독자 생성 실패 | 선택 시험에 없음 | 해당 없음(기존 분기) | 미실행 — body 0(편집 전후 같음) |
| B6 | 배달 실행자 생성 실패 | 선택 시험에 없음 | 해당 없음(기존 분기) | 미실행 — body 0(편집 전후 같음) |
| (성공) | 보조 집합에 `in-doubt-settle`, 감독 집합 불변 | `TestA193TheProductionRuntimeStartsTheSettlerAsAuxiliary` · `TestProductionRuntimeStartsExactlyOneAlertDeliverer` · `TestTheAlertDelivererIsNotASupervisedLoop` | yes — 배선 전 두 시험 FAIL(`../../implementation/red-cmd.txt`) | yes — 블록 709-752 count 1, 두 시험 PASS(`../../implementation/green-cmd.txt`) |
