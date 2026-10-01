# Function Logic Map: `TestNoProductionSiteDiscardsTheSeamsAdmissionAnswer (시험)`

- Source: `internal/app/engine/strategy_dispatch_handoff_guard_test.go`
- Source SHA-256: `252e9de2ab8e6062399ac1cd6eedb5fc912e458d46857426db7d6dba24de0e83`
- Signature: `TestNoProductionSiteDiscardsTheSeamsAdmissionAnswer(params=1, results=0)`
- Source range: `930:1`–`1004:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 새 자리는 두 값을 다 받고 `!admitted` 로 거른다 — 핀은 숫자가 아니라 이름 목록이라 다음 자리도 이름을 대야 통과한다.

## Branches and early returns

- Exact AST return nodes: `946:6, 950:6, 954:6, 958:6, 973:5, 978:6, 990:5`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 936:2 | census 판정 갈래(답 미결속 · 버림 · `_ =` 침묵 · 자리 목록 대조) |
| B2 | range | 937:3 | census 판정 갈래(답 미결속 · 버림 · `_ =` 침묵 · 자리 목록 대조) |
| B3 | if | 939:4 | census 판정 갈래(답 미결속 · 버림 · `_ =` 침묵 · 자리 목록 대조) |
| B4 | if | 945:5 | census 판정 갈래(답 미결속 · 버림 · `_ =` 침묵 · 자리 목록 대조) |
| B5 | if | 949:5 | census 판정 갈래(답 미결속 · 버림 · `_ =` 침묵 · 자리 목록 대조) |
| B6 | if | 952:5 | census 판정 갈래(답 미결속 · 버림 · `_ =` 침묵 · 자리 목록 대조) |
| B7 | if | 956:5 | census 판정 갈래(답 미결속 · 버림 · `_ =` 침묵 · 자리 목록 대조) |
| B8 | type-switch | 963:5 | census 판정 갈래(답 미결속 · 버림 · `_ =` 침묵 · 자리 목록 대조) |
| B9 | case | 964:5 | census 판정 갈래(답 미결속 · 버림 · `_ =` 침묵 · 자리 목록 대조) |
| B10 | case | 966:5 | census 판정 갈래(답 미결속 · 버림 · `_ =` 침묵 · 자리 목록 대조) |
| B11 | range | 968:6 | census 판정 갈래(답 미결속 · 버림 · `_ =` 침묵 · 자리 목록 대조) |
| B12 | if | 977:5 | census 판정 갈래(답 미결속 · 버림 · `_ =` 침묵 · 자리 목록 대조) |
| B13 | switch | 982:5 | census 판정 갈래(답 미결속 · 버림 · `_ =` 침묵 · 자리 목록 대조) |
| B14 | case | 983:5 | census 판정 갈래(답 미결속 · 버림 · `_ =` 침묵 · 자리 목록 대조) |
| B15 | case | 985:5 | census 판정 갈래(답 미결속 · 버림 · `_ =` 침묵 · 자리 목록 대조) |
| B16 | case | 987:5 | census 판정 갈래(답 미결속 · 버림 · `_ =` 침묵 · 자리 목록 대조) |
| B17 | if | 1001:2 | census 판정 갈래(답 미결속 · 버림 · `_ =` 침묵 · 자리 목록 대조) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `engineProductionFiles` | 936:23 |
| `parseEngineFile` | 937:24 |
| `filepath.Base` | 942:13 |
| `len` | 945:8 |
| `isSingleCall` | 949:16 |
| `len` | 952:8 |
| `ast.Inspect` | 962:4 |
| `bind` | 965:6 |
| `make` | 967:13 |
| `len` | 967:33 |
| `append` | 969:13 |
| `bind` | 971:6 |
| `ast.Inspect` | 975:4 |
| `isSingleCall` | 977:16 |
| `append` | 980:13 |
| `t.Errorf` | 984:6 |
| `types.ExprString` | 984:95 |
| `t.Errorf` | 986:6 |
| `types.ExprString` | 986:73 |
| `identIsBlankAssigned` | 987:10 |
| `t.Errorf` | 988:6 |
| `sort.Strings` | 994:2 |
| `strings.Join` | 1001:5 |
| `strings.Join` | 1001:33 |
| `t.Fatalf` | 1002:3 |

## State mutations and fallbacks

- 시험 코드.

## Safety conclusion

- 시험 코드 — 생산 경로 없음.
