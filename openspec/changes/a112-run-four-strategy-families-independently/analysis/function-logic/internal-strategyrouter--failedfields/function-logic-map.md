# Function Logic Map: `failedFields (새 함수)`

- Source: `internal/strategyrouter/production_family_activation.go`
- Source SHA-256: `57123f814187d9201ec6999d6375adc649b579c89beb6c4731f36cc51243b005`
- Signature: `failedFields(params=1, results=1)`
- Source range: `633:1`–`641:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 8.8.4-B).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 결정은 호출자의 `len(fields) != 0` 하나 — 이 함수가 비면 결속이 통과하므로(B3) 이름 수집이 곧 판정 입력이다.

## Branches and early returns

- Exact AST return nodes: `640:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 635:2 | 검사 순회 |
| B2 | if | 636:3 | 실패한 검사의 이름을 모음 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `append` | 637:13 |

## State mutations and fallbacks

- 상태 변경 없음.

## Safety conclusion

- 판정 입력이다 — 변이 B3 이 그것을 잡는다.

a112 8.5 응답 로트(2026-10-04): 같은 파일의 다른 함수 편집으로 줄만 밀림 — shift_same_file_bundles.py(구조 동일 확인 뒤 좌표 사상)
