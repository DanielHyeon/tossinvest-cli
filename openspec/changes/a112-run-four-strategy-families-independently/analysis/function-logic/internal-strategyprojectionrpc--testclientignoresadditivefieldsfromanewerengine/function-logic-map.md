# Function Logic Map: `TestClientIgnoresAdditiveFieldsFromANewerEngine (시험)`

- Source: `internal/strategyprojectionrpc/a112_additive_fields_test.go`
- Source SHA-256: `3d365481c525d3099026c29999bdaf048b5f7da7bedcdacf0cb8bc8873b5a0c6`
- Signature: `TestClientIgnoresAdditiveFieldsFromANewerEngine(params=1, results=0)`
- Source range: `25:1`–`84:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 주입은 세 층 모두 이 client 가 모르는 이름이다 — read 가 실패하면 시험 실패.

## Branches and early returns

- Exact AST return nodes: `67:4`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 28:2 | arrangement 오류 가드 |
| B2 | if | 36:2 | arrangement 오류 가드 |
| B3 | if | 41:2 | arrangement 오류 가드 |
| B4 | if | 46:2 | arrangement 오류 가드 |
| B5 | if | 51:2 | arrangement 오류 가드 |
| B6 | if | 55:2 | arrangement 오류 가드 |
| B7 | if | 59:2 | arrangement 오류 가드 |
| B8 | if | 65:3 | 토큰 불일치 → 401 |
| B9 | if | 76:2 | 구 reader 가 additive 필드에 죽음 → 실패 |
| B10 | if | 80:2 | 기존 필드 의미 · 형식 유지 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `strategyprojection.DormantSnapshot` | 26:14 |
| `time.Date` | 26:49 |
| `json.Marshal` | 27:14 |
| `t.Fatal` | 29:3 |
| `json.Unmarshal` | 36:12 |
| `t.Fatal` | 37:3 |
| `json.RawMessage` | 39:31 |
| `json.Unmarshal` | 41:12 |
| `t.Fatal` | 42:3 |
| `json.RawMessage` | 44:33 |
| `json.Marshal` | 45:23 |
| `t.Fatal` | 47:3 |
| `json.Unmarshal` | 51:12 |
| `len` | 51:69 |
| `t.Fatalf` | 52:3 |
| `len` | 52:44 |
| `json.RawMessage` | 54:32 |
| `json.Marshal` | 55:30 |
| `t.Fatal` | 56:3 |
| `json.Marshal` | 58:15 |
| `t.Fatal` | 60:3 |
| `strings.Repeat` | 63:11 |
| `httptest.NewServer` | 64:12 |
| `http.HandlerFunc` | 64:31 |
| `r.Header.Get` | 65:6 |
| `w.WriteHeader` | 66:4 |
| `Set` | 69:3 |
| `w.Header` | 69:3 |
| `w.Write` | 70:10 |
| `t.Cleanup` | 72:2 |
| `server.Client` | 74:61 |
| `client.Read` | 75:14 |
| `context.Background` | 75:26 |
| `t.Fatalf` | 77:3 |
| `len` | 80:62 |
| `t.Fatalf` | 82:3 |

## State mutations and fallbacks

- 시험 코드.

## Safety conclusion

- 시험 코드 — 생산 경로 없음.
