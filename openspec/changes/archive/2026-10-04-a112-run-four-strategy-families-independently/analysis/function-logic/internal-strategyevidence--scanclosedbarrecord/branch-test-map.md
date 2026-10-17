# Branch Test Map: `scanClosedBarRecord`

- Source: `internal/strategyevidence/breakout_series.go`, SHA-256 `ece27d1ede03408e1f819f5d65f42fca9a252dd3e693b4cbadf834bee4e9abc5`; branch IDs follow `ast.json` (16 branches, generated 2026-08-17).
- New function (not in the frozen base 016da624). RED per review.md: the header↔payload binding switch was RED-first in the P1/P2 fix round (2026-08-17); the record-id binding (B5) and the issuer-mapping/effective-date/unit equalities (B9–B11) were adopted post-recheck (g2/g4) and review.md carries no RED statement for them.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 120:2 — corrupted stored row | not-applicable: out-of-band corruption; append-only trigger pinned by `TestStoreSchemaTooNewAndAppendOnlyTriggers`; no `SealBarSeries` fixture tampers a row | not-applicable | not-applicable |
| B2 | if at 125:2 — strict decode fails after `scanEnvelope` accepted | not-applicable: unreachable (defensive double decode) | not-applicable | not-applicable |
| B3 | if at 134:2 — payload session unparsable after decode | not-applicable: unreachable (defensive) | not-applicable | not-applicable |
| B4 | switch at 141:2 — binding entry | every read test, e.g. `TestSealBarSeriesRoundTripsEnvelopesBuiltByTheConstructor` (all cases pass) | yes (implementer P1/P2 fix round 2026-08-17, RED-first) | yes |
| B5 | case at 142:2 — MSFT payload under AAPL header; forged 08-14 record id on an 08-13 row; r2 moved to 13:31 under the 13:30 record id; US header / KR payload | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift`, `TestSealBarSeriesRefusesACorrectionThatMovesTheBar`, `TestSealBarSeriesRefusesAPayloadThatDisagreesWithItsHeader` | not-applicable (post-recheck g2 adoption; no RED report in review.md — pinned by the named tests) | yes |
| B6 | case at 144:2 — `AuthoritySEC` header | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /header authority is not the official api | yes (implementer P1/P2 fix round 2026-08-17, RED-first) | yes |
| B7 | case at 146:2 — schema `v2` header | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /header schema version is foreign | yes (implementer P1/P2 fix round 2026-08-17, RED-first) | yes |
| B8 | case at 148:2 — issuer `US:MSFT` header | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /header issuer identity is foreign | yes (implementer P1/P2 fix round 2026-08-17, RED-first) | yes |
| B9 | case at 150:2 — mapping `a112-bar-issuer-v2` | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /header issuer mapping version is foreign | not-applicable (post-recheck g4; no RED report in review.md — pinned by the named test) | yes |
| B10 | case at 152:2 — effective date 08-13 under session 08-14 | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /header market effective date is not the session date | not-applicable (post-recheck g4; no RED report in review.md — pinned by the named test) | yes |
| B11 | case at 154:2 — unit `major` | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /header unit is not minor | not-applicable (post-recheck g4; no RED report in review.md — pinned by the named test) | yes |
| B12 | case at 156:2 — `SourceAvailableAt = SourceEventAt` | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /header availability is not the bar close | yes (implementer P1/P2 fix round 2026-08-17, RED-first) | yes |
| B13 | case at 158:2 — payload minute 6 h earlier / 6 h later than the header minute | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /payload minute drifts earlier while every other clock agrees, `TestSealBarSeriesRefusesRatherThanHidingADriftedMinute` | yes (implementer P1/P2 fix round 2026-08-17, RED-first) | yes |
| B14 | case at 160:2 — payload observed +3 min vs header +2 min | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /payload observation instant differs from the header | yes (implementer P1/P2 fix round 2026-08-17, RED-first) | yes |
| B15 | case at 162:2 — header currency KRW, payload USD | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /payload currency differs from the header | yes (implementer P1/P2 fix round 2026-08-17, RED-first) | yes |
| B16 | case at 164:2 — header r2, payload revision 1 | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /revision identity disagrees with the payload revision | yes (implementer P1/P2 fix round 2026-08-17, RED-first) | yes |

Verification: `go test ./internal/strategyevidence -count=1` / `-race`, consumers, `go build ./...`, vet, gofmt green; all four P1 probes re-run at read by both reviewers (review.md 2026-08-17 rechecks).
