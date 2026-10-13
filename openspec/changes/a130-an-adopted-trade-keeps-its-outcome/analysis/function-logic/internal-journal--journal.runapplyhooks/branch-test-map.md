# Branch Test Map: `Journal.runApplyHooks`

편집 뒤 좌표(4판 GREEN). 분기별 RED 는 따로 재지 않았음(no) — 4판 시험 전체가 base 에서 컴파일 RED(`analysis/red/red-4.txt`), 조작 행 모양은 행동 RED(red-2 · red-4). GREEN = 패키지 시험 통과(`analysis/verify/receipts.md`). Test 칸은 그 함수의 a130 시험 · 기존 시험 묶음(무편집 갈래는 기존 시험).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:276` `if hooks.Project == nil && hooks.Campaign == nil && hooks.Exit == nil {` | 기존 runApplyHooks 시험 · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` | no | yes |
| B2 | `:283` `if hooks.Project != nil {` | 기존 runApplyHooks 시험 · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` | no | yes |
| B3 | `:284` `if err := hooks.Project(ctx, handle, fill); err != nil {` | 기존 runApplyHooks 시험 · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` | no | yes |
| B4 | `:288` `if hooks.Campaign != nil {` | 기존 runApplyHooks 시험 · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` | no | yes |
| B5 | `:289` `if err := hooks.Campaign(ctx, handle, fill); err != nil {` | 기존 runApplyHooks 시험 · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` | no | yes |
| B6 | `:293` `if err := j.applyRiskBucketOwnerBindingInTx(ctx, tx, fill); err != nil {` | 기존 runApplyHooks 시험 · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` | no | yes |
| B7 | `:297` `if hooks.Exit != nil {` | 기존 runApplyHooks 시험 · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` | no | yes |
| B8 | `:298` `if err := hooks.Exit(ctx, handle, fill); err != nil {` | 기존 runApplyHooks 시험 · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` | no | yes |
