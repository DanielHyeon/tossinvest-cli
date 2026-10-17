# Function Logic Map: `decodeProductionFamilyActivation`

- Source: `internal/strategyrouter/production_family_activation.go`
- Source SHA-256: `bac16c04b38d49c479a381fb326d7dd066e5525997322619d8ceb496e3ddb0d8`
- Signature: `decodeProductionFamilyActivation(params=1, results=2)`
- Source range: `509:1`–`528:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 8.8.4-B).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 판정 불변 — 같은 바이트가 같은 sentinel 로 거절.

## Branches and early returns

- Exact AST return nodes: `511:3, 518:3, 521:3, 525:3, 527:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 510:2 | 크기 0 또는 상한 초과 → `%w: manifest size …` |
| B2 | if | 517:2 | json 해석 실패(모르는 필드 포함) → `%w: manifest json: %w` |
| B3 | if | 520:2 | 문서 뒤 데이터 → `%w: trailing data …` |
| B4 | if | 524:2 | 정규 직렬화와 다름 → `%w: manifest bytes are not …canonical…` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 510:5 |
| `len` | 510:23 |
| `fmt.Errorf` | 511:44 |
| `len` | 512:4 |
| `json.NewDecoder` | 514:13 |
| `bytes.NewReader` | 514:29 |
| `decoder.DisallowUnknownFields` | 515:2 |
| `decoder.Decode` | 517:12 |
| `fmt.Errorf` | 518:44 |
| `decoder.Decode` | 520:12 |
| `fmt.Errorf` | 521:44 |
| `json.Marshal` | 523:20 |
| `bytes.Equal` | 524:20 |
| `fmt.Errorf` | 525:44 |

## State mutations and fallbacks

- 상태 변경 없음.

## Safety conclusion

- 판정 · 수락 집합 불변 — 같은 입력이 같은 sentinel 로 거절된다(errors.Is 보존, `==` 비교 0 — grep). 메시지에 필드 이름만 더해진다.

a112 8.5 응답 로트(2026-10-04): 같은 파일의 다른 함수 편집으로 줄만 밀림 — shift_same_file_bundles.py(구조 동일 확인 뒤 좌표 사상)

0.5 응답 로트: 같은 파일의 주석 · 한 글자 편집(본문 구조 불변)으로 파일 SHA 만 바뀜
