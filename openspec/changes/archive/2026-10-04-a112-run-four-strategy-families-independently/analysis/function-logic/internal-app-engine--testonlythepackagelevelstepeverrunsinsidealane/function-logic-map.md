# Function Logic Map: `TestOnlyThePackageLevelStepEverRunsInsideALane (시험)`

- Source: `internal/app/engine/a112_lane_runtime_test.go`
- Source SHA-256: `c0eec4047bca78202787fdabf8d0ef78eaee26237ce64fb46fd65edf901a5779`
- Signature: `TestOnlyThePackageLevelStepEverRunsInsideALane(params=1, results=0)`
- Source range: `193:1`–`235:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.5).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 자리만 세면 메서드 본문이 바뀌어도 목록은 그대로다 — 정의 본문 · 빌드 태그까지 대조한다(변이 R12 · R13 CAUGHT).

## Branches and early returns

- Exact AST return nodes: `201:6, 205:6, 213:5`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 195:2 | RunBounded 자리 세기 · 인자 수 가드 · 자리 목록 대조 |
| B2 | range | 197:3 | RunBounded 자리 세기 · 인자 수 가드 · 자리 목록 대조 |
| B3 | if | 200:5 | RunBounded 자리 세기 · 인자 수 가드 · 자리 목록 대조 |
| B4 | if | 204:5 | RunBounded 자리 세기 · 인자 수 가드 · 자리 목록 대조 |
| B5 | if | 207:5 | RunBounded 자리 세기 · 인자 수 가드 · 자리 목록 대조 |
| B6 | if | 227:2 | RunBounded 자리 세기 · 인자 수 가드 · 자리 목록 대조 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `make` | 194:11 |
| `engineProductionFiles` | 195:23 |
| `parseEngineFile` | 196:11 |
| `ast.Inspect` | 198:4 |
| `len` | 207:8 |
| `t.Fatalf` | 208:6 |
| `filepath.Base` | 209:7 |
| `len` | 209:28 |
| `append` | 211:13 |
| `filepath.Base` | 211:27 |
| `declName` | 211:51 |
| `exprSpelling` | 212:6 |
| `sort.Strings` | 217:2 |
| `strings.Join` | 227:5 |
| `strings.Join` | 227:34 |
| `t.Fatalf` | 228:3 |
| `assertLaneStepForIsTheProductionStepOutsideTestSeams` | 234:2 |

## State mutations and fallbacks

- 시험 코드.

## Safety conclusion

- 시험 코드 — 생산 경로 없음.

a112 7.3.1 SHADOW 로트 — 같은 파일의 다른 함수 편집으로 줄만 밀림(본문 불변)
