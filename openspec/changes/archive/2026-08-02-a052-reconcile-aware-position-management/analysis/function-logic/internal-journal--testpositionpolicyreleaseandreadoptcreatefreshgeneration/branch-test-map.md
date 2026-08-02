# Branch Test Map: `TestPositionPolicyReleaseAndReadoptCreateFreshGeneration`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 133 | `if err != nil {` entered and complementary path | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration | pre-existing regression at frozen base | verified by current package suite |
| B2 | `if` line 136 | `if released.Status != positionpolicy.StatusReleased \|\| released.Version != 1 {` entered and complementary path | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration | pre-existing regression at frozen base | verified by current package suite |
| B3 | `if` line 139 | `if released.ObservedAt == "" {` entered and complementary path | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration | pre-existing regression at frozen base | verified by current package suite |
| B4 | `if` line 145 | `if err != nil {` entered and complementary path | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration | pre-existing regression at frozen base | verified by current package suite |
| B5 | `if` line 148 | `if got.AdoptionGeneration != 2 \|\| got.Version != 1 \|\| got.Status != positionpolicy.StatusManaged {` entered and complementary path | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration | pre-existing regression at frozen base | verified by current package suite |
| B6 | `if` line 151 | `if got.EffectivePolicyID != exitpolicy.RatchetPolicyID {` entered and complementary path | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration | pre-existing regression at frozen base | verified by current package suite |
| B7 | `if` line 155 | `if got.ObservedAt == "" \|\| got.ObservedAt == released.ObservedAt {` entered and complementary path | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration | pre-existing regression at frozen base | verified by current package suite |
| B8 | `if` line 162 | `if _, err := j.ApplyPositionPolicy(ctx, late); !errors.Is(err, positionpolicy.ErrVersionMismatch) {` entered and complementary path | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration | pre-existing regression at frozen base | verified by current package suite |
| B9 | `if` line 166 | `if current.AdoptionGeneration != 2 \|\| current.Version != 1 {` entered and complementary path | TestPositionPolicyReleaseAndReadoptCreateFreshGeneration | pre-existing regression at frozen base | verified by current package suite |
