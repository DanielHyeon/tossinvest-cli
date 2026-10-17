# Function Logic Map: `familyActivationRemaining`

- Source: `internal/strategyrouter/production_family_activation.go`
- Source SHA-256: `bac16c04b38d49c479a381fb326d7dd066e5525997322619d8ceb496e3ddb0d8`
- Signature: `familyActivationRemaining(params=2, results=2)`
- Source range: `384:1`–`390:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 8.8.4-B).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 이 패키지의 유일한 만료 판정 — 적재와 lease 상한이 같이 부른다(판정 불변).

## Branches and early returns

- Exact AST return nodes: `386:3, 389:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 385:2 | 만료(지금이 expires_at 이상) → `%w: expires_at … is not after …` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `now.Before` | 385:6 |
| `fmt.Errorf` | 386:13 |
| `Format` | 387:4 |
| `expires.UTC` | 387:4 |
| `Format` | 387:44 |
| `now.UTC` | 387:44 |
| `expires.Sub` | 389:9 |

## State mutations and fallbacks

- 상태 변경 없음.

## Safety conclusion

- 판정 · 수락 집합 불변 — 같은 입력이 같은 sentinel 로 거절된다(errors.Is 보존, `==` 비교 0 — grep). 메시지에 필드 이름만 더해진다.

a112 8.5 응답 로트(2026-10-04): 같은 파일의 다른 함수 편집으로 줄만 밀림 — shift_same_file_bundles.py(구조 동일 확인 뒤 좌표 사상)

0.5 응답 로트: 같은 파일의 주석 · 한 글자 편집(본문 구조 불변)으로 파일 SHA 만 바뀜
