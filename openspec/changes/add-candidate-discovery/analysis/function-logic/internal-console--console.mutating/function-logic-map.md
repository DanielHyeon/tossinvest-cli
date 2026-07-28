# Function Logic Map: `Console.mutating`

- Source: `internal/console/console.go`
- Change: `add-candidate-discovery`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

**본문 무변경** — 이 change의 base 대비 함수 본문이 byte 동일하다(base·HEAD 두 판본의 선언 범위 텍스트를 직접 비교해 확인했다). 인접 hunk 교차로 evidence가 요구됐고, `ast.json`은 base revision에서 뜬 것이다.

CSRF 게이트. **본문 무변경** — base 대비 byte 동일이며, 인접한 `readOnly` 삽입의 diff hunk가 교차해 evidence가 요구됐다. AST는 base revision에서 뜬 것이고 hash도 base 파일의 것이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.Method` | `POST`만 통과 | HTTP 요청 | 그 외는 405 + `Allow: POST` |
| `r.ParseForm()` | 성공해야 한다 | 요청 본문 | 실패하면 400, 아무것도 전송되지 않는다 |
| `r.PostFormValue("csrf")` | `c.csrf`와 상수 시간 일치 | 프로세스당 1회 생성된 토큰 | 불일치면 403 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `r.Method != http.MethodPost` | `Allow` 헤더 설정 | 405 + 거부 페이지 | `TestTheApprovalRoutesRefuseAGET` |
| B2 | `r.ParseForm()` 에러 | 없음 | 400 + 거부 페이지 | 폼 파싱 실패 케이스는 이 패키지에 직접 테스트가 없다 — 무변경 분기이며 잘못된 Content-Type 요청에서만 도달한다 |
| B3 | CSRF 토큰 불일치 | 없음 | 403 + 거부 페이지 | `TestAWrongCSRFTokenSendsNothing`, `TestTheEngineButtonsNeedTheSessionAndTheCSRFToken` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `tokenEqual` | 상수 시간 비교 | `subtle.ConstantTimeCompare` — 접두 유출 없음 | console.go:657 |
| `c.refuse` | 거부 페이지 렌더 | 상태 코드와 사유를 화면에 적는다. 세 분기 모두 '아무것도 전송되지 않았다'로 끝난다 | pages.go:376 |
| (금지 바인딩) | 계좌·원장·브로커 호출 없음 | 게이트는 판정만 하고 다음 핸들러로 넘긴다 | ast.json calls |

## State mutations and fallbacks

- 무상태. 응답 헤더 하나(`Allow`)만 쓴다.
- 세 거부 모두 `next`를 부르지 않는다 — fail-closed.

## Safety conclusion

- Safe edit boundary: 무변경 — 인접 삽입(`readOnly`)만 존재한다.
- High-risk impact: yes (인증 경로 — 이 게이트가 콘솔의 모든 상태변경 라우트의 두 번째 자물쇠다)
