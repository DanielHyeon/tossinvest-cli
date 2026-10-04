# Function Logic Map (편집 전): `run`

- Source: `tools/a112-family-shadow/main.go`
- Source SHA-256: `b205023b15dd8e0cdc5eb9263e3208adf68d14d6ceccf6f6d125f211f609b9f9`
- Signature: `run(params=0, results=1)`
- Source range: `60:1`–`109:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD 4cbcfb36).

## Inputs and invariants

- 편집 계획: 0.5 리뷰 보안#4: 쓰기 · 닫기 실패 시 0400 부분 파일을 지운다(O_EXCL 재시도 가능).

## Branches and early returns

- Exact AST return nodes: `80:3`, `84:3`, `87:3`, `92:3`, `96:3`, `99:3`, `102:3`, `108:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 79:2 | `if err != nil {` |
| B2 | if | 83:2 | `if err != nil {` |
| B3 | if | 86:2 | `if opts.out == "" {` |
| B4 | if | 91:2 | `if err != nil {` |
| B5 | if | 94:2 | `if _, err := file.Write(data); err != nil {` |
| B6 | if | 98:2 | `if err := file.Close(); err != nil {` |
| B7 | if | 101:2 | `if err := os.Chmod(opts.out, 0o400); err != nil {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `flag.StringVar` | 62:2 |
| `flag.Uint64Var` | 63:2 |
| `flag.StringVar` | 64:2 |
| `flag.StringVar` | 65:2 |
| `flag.StringVar` | 66:2 |
| `flag.StringVar` | 67:2 |
| `flag.StringVar` | 68:2 |
| `flag.StringVar` | 69:2 |
| `flag.StringVar` | 70:2 |
| `flag.StringVar` | 71:2 |
| `flag.StringVar` | 72:2 |
| `flag.BoolVar` | 73:2 |
| `flag.StringVar` | 74:2 |
| `flag.StringVar` | 75:2 |
| `flag.Parse` | 76:2 |
| `opts.document` | 78:19 |
| `strategyshadow.EncodeProductionFamilyShadow` | 82:15 |
| `fmt.Errorf` | 84:10 |
| `fmt.Errorf` | 87:10 |
| `os.OpenFile` | 90:15 |
| `fmt.Errorf` | 92:10 |
| `file.Write` | 94:15 |
| `file.Close` | 95:3 |
| `file.Close` | 98:12 |
| `os.Chmod` | 101:12 |
| `sha256.Sum256` | 104:12 |
| `fmt.Printf` | 105:2 |
| `len` | 105:48 |
| `fmt.Printf` | 106:2 |
| `strings.ToUpper` | 107:3 |
| `string` | 107:19 |
| `hex.EncodeToString` | 107:45 |

## Safety conclusion

- 도구 — 생산 경로 밖. 쓰기 실패에서만 동작이 바뀐다.
