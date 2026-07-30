# Branch Test Map: `BuildAttestation`

- Source: `internal/soak/attest.go`
- Function: `internal/soak/attest.go:BuildAttestation`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | soak 미충족이면 아무것도 쓰지 않는다 | `TestBuildAttestationRefusesAnIncompleteSoak` | no — 기존 동작 | yes |
| B2 | soak이 증명한 읽기를 싣는다 | `TestBuildAttestationPassesTheEngineInterlock` | no — 기존 동작 | yes |
| B3 | soak 기록의 비-GET은 거부한다 | `TestBuildAttestationNeverClaimsAnEndpointItDidNotExercise` | no — 기존 동작. 이 change의 대칭 거부의 짝 | yes |
| B4 | **감독 증거 판정 실패는 발급을 거부한다** | `TestSupervisedEvidenceCannotStandInForTheSoak`, `TestSupervisedEvidenceIsClosedToOtherMutations`, `TestSupervisedEvidenceFromAnotherAccountRefusesTheIssue` | **yes** | yes |
| B5 | **감독 증거가 endpoint 집합을 완성한다** | `TestSupervisedEvidenceCompletesTheEnginesRequiredSet`, `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun` | **yes** | yes |
| B6 | 운영자 메모가 note 앞에 붙는다 | `TestBuildAttestationCarriesTheMeasuredRate` | no — 기존 동작 | yes |

건너뜀 경로(오류가 아니라 미포함):

| 경로 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| 유효 기간 밖 | 오래된 증거는 증거가 아니다 | `TestSupervisedEvidenceOlderThanTheValidityIsNotEvidence` | **yes** | yes |
| 미래 시각 | 신뢰할 수 없는 시계는 증명하지 않는다 | `TestSupervisedEvidenceFromTheFutureIsNotEvidence` | **yes** | yes |
| 감독 증거 없음 | 읽기만 담아 발급되고 게이트가 거부한다 | `TestWithoutSupervisedEvidenceTheGateStillRefuses`, `TestSoakAttestDoesNotSatisfyTheEngineInterlockOnItsOwn` | no — 기존 동작 회귀 방지 | yes |

RED 실행 기록 (구현 전, `go test ./internal/soak/`):

```
internal/soak/supervised_test.go:43:68: too many arguments in call to soak.BuildAttestation
  have (soak.Summary, soak.Criteria, time.Time, string, string, []attest.Proof)
  want (soak.Summary, soak.Criteria, time.Time, string, string)
```

RED 실행 기록 (배선만 되돌린 상태, `go test ./cmd/tossctl/ -run TestSoakAttest`):

```
--- FAIL: TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun
    soak_test.go:640: the engine still lacks [POST /api/v1/orders
    POST /api/v1/orders/{id}/cancel] after a supervised check proved the mutations
--- FAIL: TestSoakAttestRefusesSupervisedEvidenceFromAnotherAccount
    soak_test.go:663: an attestation was written from another account's verification record
```

GREEN 실행 기록: `go test ./... -count=1` → 3742 passed in 57 packages (구현 전 3723).
실기록 확인: `./bin/tossctl soak attest --out /tmp/…` → endpoints 8/8, supervised 2줄,
그리고 "The gate will still refuse to start … interlock clause 9".
