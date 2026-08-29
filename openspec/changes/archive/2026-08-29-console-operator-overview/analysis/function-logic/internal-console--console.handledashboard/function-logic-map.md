# Function Logic Map: `Console.handleDashboard`

- Source: `internal/console/pages.go`
- Change: `console-operator-overview`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 change의 diff가 이 함수의 본문을 바꿨다. 아래 분석은 현재 HEAD 본문에 대한 것이다.

`/` 화면(검증 콘솔) 핸들러. 이 change가 바꾼 것은 두 문자열이다 — 404 문구의 화면 열거가 '여섯'에서 '아홉'으로, `Nav`가 `dashboard`에서 `verify-console`로. 후자는 리뷰 P2가 지적한 이름 충돌 해소다: `/`와 `/dashboard`가 같은 이름(`대시보드`)을 갖고 있었다. 판정 분기 둘은 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.URL.Path` | 정확히 `/` | HTTP 요청 | 그 외는 404 — Go mux의 `/`는 catch-all이므로 이 판정이 없으면 미등록 경로가 이 화면을 받는다 |
| `c.csrf` | 프로세스당 1회 생성 | `New` | 폼이 제출할 토큰을 페이지에 싣는다 |
| `c.snapshot()` | soak·검증·증명 기록의 현재 상태 | 로컬 파일 | 파일 부재는 화면이 보고하는 상태이지 오류가 아니다 |
| `c.opts.Relaunch` / `RestartSoak` | nil 허용 | cmd/tossctl seam | nil이면 버튼을 숨긴다 — 이 빌드가 못 하는 일의 버튼은 사과로 답하는 버튼이다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `r.URL.Path != "/"` | 없음 | 404 + 화면 아홉 열거 | `TestTheConsoleServesNothingButItsThreePages` |
| B2 | `run := c.currentRun(); run != nil` | `page.Run`, `page.Refresh` 설정 | 없음(계속) | `TestTheApprovedFlowRunsExactlyTheApprovedBatch`, `TestShutdownWaitsForARunInProgress`, `TestTheDashboardReportsAnUnstartedMachineWithoutFailing`(nil 쪽) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.refuse` | 404 | '그런 경로는 없다' + 화면 목록 | pages.go:376 |
| `c.snapshot()` | 기록 3종 읽기 | 파일 부재는 상태로 렌더 | data.go |
| `c.currentRun()` / `run.snapshot()` | 진행 중 검증 상태 | mutex 아래 복사 | run.go |
| `c.render` | 템플릿 렌더 | `html/template` — 스크립트·CDN·외부 자산 없음 | pages.go:360 |
| (금지 바인딩) | 브로커·원장 호출 없음 — 이 화면은 로컬 기록만 읽는다 | `TestTheConsoleWritesNothingButTheEvidenceItsRunnerWrites`(쓰기 금지) + import 금지 목록 | static_test.go:583 |

## State mutations and fallbacks

- 읽기만 한다. `dashboardPage` 값 하나를 만들어 렌더한다.
- `page.CSRF = c.csrf` — 이 화면이 승인 폼의 토큰을 페이지에 싣는 지점이다.
- `page.Refresh`는 검증이 돌고 승인 대기가 아닐 때만 참이다 — nonce 폼이 떠 있는 동안 재로드하면 입력이 지워진다.

## Safety conclusion

- Safe edit boundary: 문자열 둘(404 문구·Nav 라벨). 경로 판정과 run 스냅샷 분기는 무변경.
- High-risk impact: yes (인증 표면 — 이 핸들러가 템플릿에 넘기는 `c.csrf`가 승인 폼이 제출하는 그 토큰이다. 이번 diff 자체는 문구 둘이지만 함수의 소재는 그 지점이다)
