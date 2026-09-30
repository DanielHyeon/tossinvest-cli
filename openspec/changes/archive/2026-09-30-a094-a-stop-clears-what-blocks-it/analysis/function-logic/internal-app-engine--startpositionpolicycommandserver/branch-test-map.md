# Branch Test Map: `StartPositionPolicyCommandServer`

- Source: `internal/app/engine/position_policy_transport.go`

> Test 열은 그 분기를 지나는 시험(현존 · 통과)이다. 「미진입」 분기의 인용은 그 함수를 도는 시험이며 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> RED 는 새 a094 시험에 한해 변이 원장(`analysis/implementation/mutation-ledger.md`)이 잰다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:52` `if commands == nil {` | 아니오 | `TestA094AnOperatorThawClosesTheParkAndReleasesTheProposal` | n/a | yes |
| B2 | `:56` `if dir == "" {` | 아니오 | `TestA094AnOperatorThawClosesTheParkAndReleasesTheProposal` | n/a | yes |
| B3 | `:59` `if err := positionpolicyrpc.ValidateEngineDirectory(dir); err != nil {` | 예 | `TestA094AnOperatorThawClosesTheParkAndReleasesTheProposal` | n/a | yes |
| B4 | `:64` `if err := os.Mkdir(controlDir, 0o700); err != nil {` | 예 | `TestA094AnOperatorThawClosesTheParkAndReleasesTheProposal` | n/a | yes |
| B5 | `:68` `} else {` | 예 | `TestA094AnOperatorThawClosesTheParkAndReleasesTheProposal` | n/a | yes |
| B6 | `:65` `if !errors.Is(err, os.ErrExist) {` | 아니오 | `TestA094AnOperatorThawClosesTheParkAndReleasesTheProposal` | n/a | yes |
| B7 | `:71` `if err := positionpolicyrpc.ValidateControlDirectory(controlDir); err != nil {` | 예 | `TestA094AnOperatorThawClosesTheParkAndReleasesTheProposal` | n/a | yes |
| B8 | `:72` `if createdControlDir {` | 아니오 | `TestA094AnOperatorThawClosesTheParkAndReleasesTheProposal` | n/a | yes |
| B9 | `:78` `if createdControlDir {` | 아니오 | `TestA094AnOperatorThawClosesTheParkAndReleasesTheProposal` | n/a | yes |
| B10 | `:107` `if err != nil {` | 아니오 | `TestA094AnOperatorThawClosesTheParkAndReleasesTheProposal` | n/a | yes |
| B11 | `:112` `if _, err := rand.Read(tokenBytes); err != nil {` | 아니오 | `TestA094AnOperatorThawClosesTheParkAndReleasesTheProposal` | n/a | yes |
| B12 | `:123` `if r.Method != http.MethodGet {` | 아니오 | `TestA094AnOperatorThawClosesTheParkAndReleasesTheProposal` | n/a | yes |
| B13 | `:130` `if r.Method != http.MethodGet {` | 아니오 | `TestA094AnOperatorThawClosesTheParkAndReleasesTheProposal` | n/a | yes |
| B14 | `:135` `if err != nil {` | 아니오 | `TestA094AnOperatorThawClosesTheParkAndReleasesTheProposal` | n/a | yes |
| B15 | `:148` `if quarantines, ok := commands.(exitQuarantineCommands); ok {` | 예 | `TestA094AnOperatorThawClosesTheParkAndReleasesTheProposal` | n/a | yes |
| B16 | `:153` `if relaxations, ok := commands.(riskRelaxationCommands); ok {` | 예 | `TestA094AnOperatorThawClosesTheParkAndReleasesTheProposal` | n/a | yes |
| B17 | `:157` `if thaws, ok := commands.(attemptThawCommands); ok {` | 예 | `TestA094AnOperatorThawClosesTheParkAndReleasesTheProposal` | n/a | yes |
| B18 | `:167` `if err := writePositionPolicyDescriptor(server.descriptor, descriptor); err != nil {` | 아니오 | `TestA094AnOperatorThawClosesTheParkAndReleasesTheProposal` | n/a | yes |
