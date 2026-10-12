# Branch Test Map: `runSoakAttest`

- Source: `cmd/tossctl/soak.go` (476-582) — 편집 전(base)

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 479:2 | `if opts.recordRenewalStatus {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체 | 무편집 |
| B2 | if at 480:3 | `if strings.TrimSpace(opts.record) != "" \|\| strings.TrimSpace(opts.out) != "" {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체 | 무편집 |
| B3 | if at 484:3 | `if err != nil {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체 | 무편집 |
| B4 | if at 492:4 | `if result == nil {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체 | 무편집 |
| B5 | if at 495:5 | `if a, err := attest.Load(attestationPath); err != nil {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체 | 무편집 |
| B6 | else at 497:12 | `} else {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체 | 무편집 |
| B7 | if at 501:4 | `if err := saveRenewalStatus(statusPath, status); err != nil && result == nil {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체 | 무편집 |
| B8 | if at 507:2 | `if err != nil {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체 | 무편집 |
| B9 | if at 510:2 | `if opts.validity > 0 {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체 | 무편집 |
| B10 | if at 515:2 | `if base := soakSurveyedBase(root); base != "" {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체 | 무편집 |
| B11 | if at 521:2 | `if err != nil {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체 | 무편집 |
| B12 | if at 526:2 | `if err != nil {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체 | 무편집 |
| B13 | if at 532:2 | `if !opts.recordRenewalStatus {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체 | 무편집 |
| B14 | if at 535:3 | `if err != nil {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체 | 무편집 |
| B15 | if at 539:2 | `if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체 | 무편집 |
| B16 | if at 542:2 | `if err := attest.Save(path, attestation); err != nil {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체 | 무편집 |
| B17 | range at 555:2 | `for _, p := range attestation.SupervisedBy {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체 | 무편집 |
| B18 | if at 557:3 | `if market == "" {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체 | 무편집 |
| B19 | if at 565:2 | `if len(missing) > 0 {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체 | 무편집 |
| B20 | range at 568:3 | `for _, e := range missing {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체 | 무편집 |

## a131 시험

- `TestA131SoakAttestOutputCarriesNoAccount`
