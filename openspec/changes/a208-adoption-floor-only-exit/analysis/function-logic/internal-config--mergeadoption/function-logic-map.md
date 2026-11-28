# Function Logic Map: `mergeAdoption`

- Source: `internal/config/engine.go` (299-317)
- Qualified: `mergeAdoption`
- AST evidence: `ast.json` (source_sha256 `3275a148ad43…`) — AST branches 3 · returns 2
- **편집 뒤 갱신(a208 3.3, 2026-11-28)**: `ast.json` 을 편집 뒤 소스로 재추출(`revision: current`), 좌표는 옛→새 AST 노드 순서 대응으로 옮김(분기 · 반환 개수 무변, 새로 생긴 분기 없음). 아래 「Safe edit boundary」는 편집 전 계획이며 그대로 구현됨.
- Risk scan: `risk-pattern-report.md`(설정 규칙 일치 0)
- 작성 시점: 편집 **전**(a208 tasks 1.1, 2026-11-28, HEAD `810aec6f7` — base `baeddfe31` 와 본문 동일).

**역할.** 설정 파일의 `engine.adoption` 블록을 기본값 위에 병합하거나, 거부하면 전면 0 으로 둔다. 호출자 `mergeEngine`(engine.go) 하나.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `raw *rawAdoption` | nil(키 없음) 또는 파싱 결과 | `json.Unmarshal`(service.go:250, 엄격 아님) | nil → 기본값 유지 |
| 불변식 | 거부 블록은 부분 유지 없이 전면 0 | 이 함수 주석 · spec 「범위 밖 손절폭 설정」 | — |

## Branches and early returns

| Branch | Condition | At | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if raw == nil` | 300:2 | — | return (301:3) | `TestAdoptionDefaultsOff` |
| B2 | `if raw.Enabled != nil` | 304:2 | `next.Enabled = *raw.Enabled` | — | `TestAdoptionBlockParsed` |
| B3 | `if why := next.validate(); why != ""` | 312:2 | `cfg.Adoption = Adoption{Rejected: why}` | return (314:3) | `TestAdoptionRefusesAStopFractionOutsideTheBand` |

Exact AST return positions: 301:3, 314:3

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `normaliseSymbols` | 목록 정규화 | — | AST |
| `Adoption.validate` | 손절폭 범위 검사 | 거부 사유 문자열 | AST |

Live config binding: 설정 파일 → 기동 때 1 회.

## State mutations and fallbacks

- `cfg.Adoption` 대입 하나(수용) 또는 전면 0(거부).

## Safety conclusion

- Safe edit boundary(a208): `next` 조립부에 `next.FloorOnlyExit = raw.FloorOnlyExit` 한 줄(B3 앞 — 거부되면 B3 가 전면 0 으로 덮어 거짓). 분기 0 추가.
- High-risk impact: no(설정 병합). 토글 거짓 · 키 없음은 a208 이전과 같다.
