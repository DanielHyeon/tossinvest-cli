# Function Logic Map: `Service.LoadRawEngineAdoption`

- Source: `internal/config/adoption_io.go` (47-73)
- Qualified: `Service.LoadRawEngineAdoption`
- AST evidence: `ast.json` (source_sha256 `6f1df446b991…`) — AST branches 4 · returns 4
- Risk scan: `risk-pattern-report.md`(설정 규칙 일치 0)
- 작성 시점: 편집 **전**(a208 tasks 1.1, 2026-11-28, HEAD `810aec6f7` — base `baeddfe31` 와 본문 동일).

**역할.** 콘솔용 원문 읽기 — 병합의 전면 0 이전 블록과 판정 사유를 돌려준다. 호출자 `consoleAdoptionSettings.Load`(cmd/tossctl/adoptionsettings.go:59, HEAD 확인 — CodeGraph 는 엣지 누락, evidence-reconciliation).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s.path` | config.json 경로 | `Service` | — |
| 불변식 | 거부 블록도 쓰인 그대로 돌려줌(콘솔이 고쳐 저장) | 이 파일 머리 주석 | — |

## Branches and early returns

| Branch | Condition | At | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if errors.Is(err, os.ErrNotExist)` | 49:2 | — | return 0 블록 (50:3) | `TestRawLoadOfAMissingFileIsTheZeroBlock` |
| B2 | `if err != nil` | 52:2 | — | return err (53:3) | 미측정 |
| B3 | `if err := json.Unmarshal(...); err != nil` | 61:2 | — | return err (62:3) | 미측정 |
| B4 | `if raw.Enabled != nil` | 67:2 | `block.Enabled` | — | `TestRawLoadReturnsARefusedBlockAsWritten` |

Exact AST return positions: 50:3, 53:3, 62:3, 72:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `os.ReadFile` | 파일 읽기 | 오류 → B1/B2 | AST |
| `json.Unmarshal` | 파싱(엄격 아님) | 오류 → B3 | AST |
| `normaliseSymbols` · `block.validate` | 정규화 · 판정 사유 | — | AST |

Live config binding: 호출 때마다 파일을 읽음(콘솔 요청 시).

## State mutations and fallbacks

- 없음(읽기 전용).

## Safety conclusion

- Safe edit boundary(a208): `block` 조립부에 `block.FloorOnlyExit = raw.FloorOnlyExit` 한 줄. 분기 0 추가.
- High-risk impact: no.
