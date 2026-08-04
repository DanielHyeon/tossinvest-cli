# Function Logic Map: `Console.routes`

- Source: `internal/console/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a069-operator-can-lift-a-quarantine/base-commit.txt`
- 위험 등급: Normal (라우트 등록). 다만 어떤 미들웨어로 감싸느냐가 안전 성질을
  결정하므로 새 라우트의 래핑을 명시한다. Pre-Edit 선언은 `review.md`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.remote` | nil 가능 | 콘솔 옵션 | nil이면 원격 로그인 라우트 없음, security 래핑 없음 |
| `c.remote.trustedNetwork` | bool | 원격 설정 | true면 `/login`·`/logout`을 등록하지 않는다 |

**불변식**: 모든 mutation 라우트는 `c.session0(c.mutating(...))`를 통과한다.
모든 읽기 라우트는 최소한 `c.session0(...)`을 통과한다. `"/"`는 catch-all이고
`redirectRoot`가 이 콘솔이 서비스하지 않는 모든 경로의 404를 소유한다.

**a063이 추가하는 것**: 두 줄.

```
/position-management/quarantine/preview  → c.session0(c.mutating(handler, 4096))
/position-management/quarantine/apply    → c.session0(c.mutating(handler, 4096))
```

기존 `/position-management/preview`·`/apply`와 **정확히 같은 래핑**이다. 읽기용
라우트는 추가하지 않는다 — 격리 정보는 이미 `/position-management` GET이 함께
싣는다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (711) | `c.remote != nil && !c.remote.trustedNetwork` | `/login`·`/logout` 등록 | 계속 | 기존 `remote_test.go` |
| B2 (816) | `c.remote != nil` | — | `c.remote.security(mux)` 반환 | 기존 `remote_csp_test.go` |

a063은 **분기를 추가하지 않는다.** `mux.HandleFunc` 호출 두 개만 늘어난다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `http.NewServeMux` | 라우터 | — | AST |
| `c.session0` | 세션 교환·핸드오프 | 모든 라우트 필수 | AST |
| `c.mutating` | CSRF + 본문 상한 | 모든 POST 필수 | AST |
| `c.readOnly` | GET 전용 강제 | 읽기 라우트 | AST |
| `c.remote.security` | 원격일 때 보안 헤더·인증 | — | AST |

## State mutations and fallbacks

- 상태를 만들지 않는다. `http.Handler` 하나를 조립해 반환한다.
- fallback: 등록되지 않은 경로는 `"/"` catch-all의 `redirectRoot`가 404로 처리한다.
  따라서 격리 라우트가 등록되지 않아도(콘솔 구버전) 404이지 오작동이 아니다.

## Safety conclusion

- Safe edit boundary: `mux.HandleFunc` 등록 목록.
- High-risk impact: **no** — 다만 새 POST 두 개가 `mutating`을 반드시 통과해야
  한다는 점이 이 함수에서 지켜야 할 유일한 계약이고, 기존 정책 라우트와 동일한
  래핑을 그대로 사용한다. `screen_paths_test.go`의 라우트 계약 테스트가 이를 검사한다.
- §0.7: 새 라우트는 화면을 노출할 뿐 자동 실행 경로를 만들지 않는다.
