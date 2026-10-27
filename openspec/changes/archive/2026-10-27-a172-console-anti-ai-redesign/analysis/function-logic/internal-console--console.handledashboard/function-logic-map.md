# Function Logic Map: `Console.handleDashboard`

- Source: `internal/console/pages.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- 출처: `archive/2026-10-26-a174-settings-startup-readiness-and-presets` 번들 복사, source sha 대조 일치 — a172 종결 주석(2026-10-28): a172 base `fd6afe23b` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/console/pages.go` = `d676ff005fda…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c` | 생성 완료된 `Console` | `Console.opts`와 요청별 읽기 seam | 새 기동 준비 읽기는 실패한 절을 `unknown`과 사유로만 표시하며, 기존 화면 렌더를 중단하지 않음 |
| `r` | 인증을 통과한 HTTP 요청 | `r.URL.Query()` | `notice`가 없으면 빈 문자열로 렌더함 |
| 현재 검증 run | 없음 또는 `runView` 스냅숏 | `c.currentRun()` | 없으면 `Run`과 자동 새로고침을 비움 |

- 불변식: 기존 `chrome`·CSRF·스냅숏·엔진 제어 폼·`Refresh` 계산은 바꾸지 않음.
- 불변식: 요청마다 `c.render`를 한 번 호출함.
- 불변식: 추가하는 `Readiness`는 읽기 전용 뷰 모델이며 주문·토글·엔진 제어를 호출하지 않음.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `c.currentRun()`이 nil이 아님 | run 스냅숏을 `page.Run`에 넣고, 완료·승인 대기 상태가 아닐 때만 `page.Refresh`를 켬 | 조기 반환·오류 없음 | 기존 콘솔 회귀군 `go test ./internal/console/ -count=1`; 이 change는 분기 조건을 바꾸지 않음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.chromeOnRequest` | 검증 콘솔 chrome 생성 | 오류 반환 없음 | AST, CodeGraph callees |
| `r.URL.Query().Get` | notice 읽기 | 누락 시 빈 문자열 | AST |
| `c.snapshot` | 상태 스냅숏 생성 | 기존 스냅숏 계약 유지 | AST, CodeGraph callees |
| `c.startupReadiness` | 기동 준비 읽기 전용 뷰 모델 생성 | nil seam·읽기 오류는 해당 행 `unknown`+사유; 렌더는 계속 | a174 design §2.1, RED 2.1 |
| `c.currentRun`·`run.snapshot` | 현재 검증 run 표시와 새로고침 결정 | nil이면 B1을 건너뜀 | AST |
| `c.render` | dashboard 템플릿 렌더 | 기존 렌더 오류 처리 계약 유지 | AST, CodeGraph callers/impact |

## State mutations and fallbacks

- 지역 `page`만 조립함. 공유 설정·원장·브로커 상태를 변경하지 않음.
- B1이 아니면 `page.Run`은 nil이고 `Refresh`는 false로 남음.
- 새 읽기 뷰가 필요한 seam을 받지 못하거나 읽기에 실패하면 영값을 통과로 해석하지 않고 행마다 `unknown`과 사유를 넣음.

## Safety conclusion

- Safe edit boundary: `dashboardPage`에 읽기 전용 필드 하나 추가, 리터럴에 `Readiness: c.startupReadiness(),` 한 줄 추가, `startupReadiness`는 새 파일에 구현.
- High-risk impact: no — 주문 제출·취소·정정, Guardian 판정, 엔진 시작·정지와 설정 저장 경로를 변경하지 않음.
