# Function Logic Map: `Ntfy.Publish`

> **출처 — 겹친 change.** a133(`a133-an-alert-carries-no-account`)이 편집한 함수가 a130 창
> (base `af0c0ce7` → landed-commit `727a841e`) 안에 끼어서 요구됨 — 겹친 change. 편집 커밋은
> a133 의 `35936cc8`(창 안). a130 은 이 함수를 편집하지 않음.
> AST 는 a130 쪽에서 `go run ./tools/logic-map --file internal/obs/ntfy.go --func Ntfy.Publish` 로
> **다시 추출**함. 추출한 소스 = `727a841e` 의 파일(워킹트리 · HEAD `272dba00` 와 바이트 동일,
> 파일 sha256 `0c451499c6b640fd686f3c74d6ee1014ac40abef9e428599387da46346776b8e`).
> 재추출한 `ast.json` 은 a133 번들 `openspec/changes/a133-an-alert-carries-no-account/analysis/function-logic/internal-obs--ntfy.publish/ast.json`
> 과 바이트 동일(`cmp` 일치, ast.json sha256 `9cbc9e0d6dd79810…`). 분기 · 시험 대응은 a133 번들을 출처로 삼되
> 각 행을 아래 Branch Test Map 의 시험별 coverprofile 로 다시 잼.

- Source: `internal/obs/ntfy.go` (85-142)
- Qualified: `Ntfy.Publish`
- AST evidence: `ast.json` (`revision: current` = `727a841e`, `source_sha256` 0c451499c6b640fd…) — AST branches 9 · returns 5
- Risk scan: `risk-pattern-report.md`

**역할.** 생산 코드의 `obs.Publisher` 구현 — 알림 하나를 ntfy 토픽으로 POST 함. 첫 문장
`msg = redactForTransport(msg)`(88줄, a133)가 본문 · Title · Tags · Type 을 계좌 규칙에 통과시킴.
반환이 전달 판정의 입력임: nil = 나감, 오류 = 안 나감. a130 의 편입 거래 성과 경로는 이 함수를 부르지 않음 —
창 안에 있어서 요구될 뿐임.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Topic` | 비어 있지 않음 | 설정 · 환경 변수 | 빈 값 → `ErrNtfyNotConfigured`(B1) |
| `n.BaseURL` | 비면 `DefaultNtfyBaseURL` | 설정 | B2 기본값 |
| `n.Timeout` | 0 이하 → `DefaultPublishTimeout` | 호출자 | B3 |
| `n.Token` | 선택 | 환경 변수 | 비면 Authorization 헤더 없음(B6) |
| `n.HTTPClient` | nil 가능 | 시험 주입 | nil → 기본 클라이언트(B7) |
| `msg` | 임의 — 계좌 텍스트 포함 가능 | 호출자 | 88줄에서 치환된 값으로 바뀐 뒤 쓰임 |

## Branches and early returns

| Branch | AST kind | Source line | 의미 |
|---|---|---|---|
| B1 | if at 90:2 | `if topic == "" {` | 토픽 없음 → `ErrNtfyNotConfigured` 반환(91:3) |
| B2 | if at 94:2 | `if base == "" {` | 서버 기본값 |
| B3 | if at 99:2 | `if timeout <= 0 {` | 시한 기본값 |
| B4 | if at 107:2 | `if err != nil {` | 요청 구성 실패 → 오류 반환(108:3) |
| B5 | if at 113:2 | `if title := strings.TrimSpace(msg.Title); title != "" {` | Title 헤더 |
| B6 | if at 120:2 | `if token := strings.TrimSpace(n.Token); token != "" {` | Authorization 헤더 |
| B7 | if at 125:2 | `if client == nil {` | 기본 HTTP 클라이언트 |
| B8 | if at 129:2 | `if err != nil {` | 전송 실패 → 오류 반환(130:3) |
| B9 | if at 137:2 | `if resp.StatusCode < 200 \|\| resp.StatusCode >= 300 {` | 2xx 아님 → 오류 반환(138:3) |

Exact AST return positions: 91:3, 108:3, 130:3, 138:3, 141:2.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `redactForTransport` | 전송 경계 치환(a133) | 순수 | AST calls |
| `context.WithTimeout` · `cancel` | 요청 시한 | defer 로 해제 | AST calls · defers |
| `http.NewRequestWithContext` | POST base/topic | 오류 → B4 | AST calls |
| `req.Header.Set` | Title · Tags · Priority · Content-Type · Authorization | 없음 | AST calls |
| `priorityFor` | 우선순위 | 순수 | AST calls |
| `client.Do` | 전송 | 오류 → B8 | AST calls |
| `io.ReadAll(io.LimitReader(…, 4096))` | 응답 본문(오류 문구) | 오류 무시 | AST calls |
| `fmt.Errorf` | 오류 감싸기 | — | AST calls |

Live config binding: `n.Topic` · `n.BaseURL` · `n.Token` · `n.Timeout`(구성 시점 값).

## State mutations and fallbacks

- 반환 외 부작용은 네트워크 요청 하나. `msg` 는 값 매개변수라 88줄 재대입이 호출자 값을 바꾸지 않음.

## Safety conclusion

- a130 편집 없음 — 이 번들은 창 안 남의 편집(a133)을 기술하는 증거임.
- High-risk impact: 인접 — 반환이 알림 전달 판정의 입력이나 주문 · 손절 경로는 이 함수를 기다리지 않음(a133 design D5).
