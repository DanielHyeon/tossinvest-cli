# Branch Test Map: `seedQualifyingRecord`

- Source: `cmd/tossctl/soak_test.go`
- Function: `cmd/tossctl/soak_test.go:seedQualifyingRecord` (base revision `8fff3aa44569`)

이 change는 이 함수를 수정하지 않았다. RED 열이 전부 `no`인 것은 그 사실의 기록이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` base line 552: `if err != nil {` | `seedQualifyingRecord` | no | yes |
| B2 | `for` base line 558: `for day := 0; day < 3; day++ {` | `seedQualifyingRecord` | no | yes |
| B3 | `for` base line 559: `for half := 0; half < 2; half++ {` | `seedQualifyingRecord` | no | yes |
| B4 | `range` base line 576: `for _, e := range soak.RequiredEndpoints() {` | `seedQualifyingRecord` | no | yes |
| B5 | `if` base line 581: `if err := rec.Append(c); err != nil {` | `seedQualifyingRecord` | no | yes |
