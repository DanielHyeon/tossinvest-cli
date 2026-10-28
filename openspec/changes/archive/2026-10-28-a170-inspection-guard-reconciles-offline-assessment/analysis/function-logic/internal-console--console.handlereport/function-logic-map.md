# Function Logic Map: `Console.handleReport`

- Source: `internal/console/pages.go` (355-368)
- Qualified: `Console.handleReport`
- AST evidence: `ast.json` — AST branches 2 · returns 0
- Risk scan: `risk-pattern-report.md`
- 출처: `archive/2026-10-27-a172-console-anti-ai-redesign` 번들 복사, source sha 대조 일치 — a170 종결 주석(2026-10-28): a170 base `0d168e0a3` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/console/pages.go` = `d676ff005fda…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.
- 작성 시점: a172 종결(2026-10-28, 종결 에이전트). 사후 작성한 증거임 — FLM 이 구현 판단에 쓰이지 않음(review.md 「gate ⑤」).
- a173 번들(`archive/2026-10-25-a173-console-ux-common-cleanup`, `source_sha256` f219fc446fdb…)은 복사하지 않음: a174 `9c3dd54c7` 가 같은 파일의 `handleDashboard` 를 고쳐 파일 sha 가 `d676ff005fda…` 로 바뀜. 함수 본문은 a173 판과 같음 — 두 ast.json 의 branches·returns·calls·assignments 를 시작 줄 기준 상대 좌표로 비교해 일치(2026-10-28). 그래서 아래 분석은 a173 판을 현재 좌표로 옮긴 것이고 ast.json 은 현재 소스로 새로 뽑음. a172 의 편집은 0(a172 는 `pages.go` 를 만지지 않음).

**역할.** `/report` 읽기 전용 화면. 보고서를 텍스트로 렌더.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `market` | 요청 쿼리 | `marketOf` | 미지 값은 marketOf 기본값 |
| `c.report` 결과 | 보고서 또는 오류 | 보고서 생성 | 오류 → Error 한 줄 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `c.report` 오류 (`359:2`) | `page.Error` 채움 | 렌더 계속 | 측정상 없음(기존 공백) |
| B2 | 성공 (`361:9`) | 텍스트 렌더 · `Record` 분리 · 본문에서 기록 경로 줄 1회 제거 | 렌더 계속 | `TestOperatorBodyCarriesNoDeveloperIdentifiers` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `marketOf` · `c.chromeFor` · `onRequest` · `c.snapshotFor` | 머리 · 스냅샷 | 읽기만 | AST calls 356·357 |
| `c.report` | 보고서 | 오류 → B1 | AST calls 358 |
| `rep.WriteText` | 텍스트 | 버퍼 쓰기 | AST calls 363 |
| `strings.Replace` | 경로 줄 1회 제거 | 순수 — 줄이 없으면 원문 그대로 | AST calls 365 |
| `c.render` | 렌더 | 무변경 | AST calls 367 |

## State mutations and fallbacks

- 지역 `page` 만 채움. 파일 · 원장 쓰기 0.

## Safety conclusion

- Safe edit boundary: 성공 갈래의 표시 문자열 가공. 줄 모양이 다르면 Replace 는 무동작이라 정보 손실 없음.
- High-risk impact: no — 콘솔 읽기 전용 렌더(주문 · 손절 · 사이징 · 원장 · 인증 경로 미접촉).
