# Function Logic Map: `Console.handleReport`

- Source: `internal/console/pages.go` (353-366) — 편집 뒤(a173 rebase 후 main).
- Qualified: `Console.handleReport`
- AST evidence: `ast.json` (`source_sha256` f219fc446fdb5c60…) — AST branches 2 · returns 0
- Risk scan: `risk-pattern-report.md`
- 작성 시점: 완료 게이트 직전(2026-10-25, 종결 에이전트). tasks 3.1 의 비례 원칙 면제는 required 7 이라 gate ⑤ 가 받지 않아 사후에 만든 산출물임(FLM 이 구현 판단에 쓰이지 않음).

**역할.** `/report` 읽기 전용 화면. 보고서를 텍스트로 렌더.

a173 편집: 성공 갈래(B2)에서 기록 파일 경로 `rep.Record` 를 `page.Record` 로 따로 담고 본문 텍스트에서 그 한 줄(`  기록             <경로>`)을 1 회 제거 — 경로는 접힌 기술 정보로 렌더. 오류 갈래 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `market` | 요청 쿼리 | `marketOf` | 미지 값은 marketOf 기본값 |
| `c.report` 결과 | 보고서 또는 오류 | 보고서 생성 | 오류 → Error 한 줄 |

## Branches and early returns

| Branch | Source line | AST kind | Condition / 효과 | a173 |
|---|---|---|---|---|
| B1 | `357:2` | if | report 오류 → `page.Error` | 무변경 — 측정 시험 0(기존 공백, a173 무관) |
| B2 | `359:9` | else | 성공 → 텍스트 렌더 · Record 분리 | a173: Record 분리 · 본문에서 경로 줄 제거 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `marketOf` · `c.chromeFor` · `onRequest` · `c.snapshotFor` | 머리 · 스냅샷 | 읽기만 | AST calls 354 · 355 |
| `c.report` | 보고서 | 오류 → B1 | AST calls 356 |
| `rep.WriteText` | 텍스트 | 버퍼 쓰기 | AST calls 361 |
| `strings.Replace`(a173) | 경로 줄 1 회 제거 | 순수 — 줄이 없으면 원문 그대로 | AST calls 363 |
| `c.render` | 렌더 | 무변경 | AST calls 365 |

## State mutations and fallbacks

- 지역 `page` 만 채움. 파일 · 원장 쓰기 0.

## Safety conclusion

- Safe edit boundary: 성공 갈래의 표시 문자열 가공. 줄 모양이 다르면 Replace 는 무동작이라 정보 손실 없음(경로는 `Record` 로도 렌더).
- High-risk impact: no — 콘솔 읽기 전용 렌더(주문 · 손절 · 사이징 · 원장 · 인증 경로 미접촉).
