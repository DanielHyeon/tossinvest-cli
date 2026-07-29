# Branch Test Map: `TestBuildAttestationPassesTheEngineInterlock`

- Source: `internal/soak/attest_test.go`
- Function: `internal/soak/attest_test.go:TestBuildAttestationPassesTheEngineInterlock`

RED/GREEN은 이 change를 구현하며 실제로 관측한 것이다. RED `no`는 이 change가 그 분기의
동작을 바꾸지 않아 실패 상태를 따로 만들지 않았다는 뜻이며, Test 열은 지금 그것을 덮는 것을
가리킨다. 이 change의 RED 실행 기록은 `internal-soak--buildattestation`의 branch-test-map에 있다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` line 187: `if err != nil {` | `TestBuildAttestationPassesTheEngineInterlock` | no | yes |
| B2 | `if` line 191: `if a.FormatVersion != attest.FormatVersion {` | `TestBuildAttestationPassesTheEngineInterlock` | no | yes |
| B3 | `if` line 194: `if a.SoakDays != 3 {` | `TestBuildAttestationPassesTheEngineInterlock` | no | yes |
| B4 | `if` line 197: `if !a.IssuedAt.Equal(now) {` | `TestBuildAttestationPassesTheEngineInterlock` | no | yes |
| B5 | `if` line 200: `if !a.ExpiresAt.After(now) {` | `TestBuildAttestationPassesTheEngineInterlock` | no | yes |
| B6 | `if` line 204: `if err := a.Verify(now, "123-45-678901", soak.RequiredEndpoints()); err != nil {` | `TestBuildAttestationPassesTheEngineInterlock` | no | yes |
