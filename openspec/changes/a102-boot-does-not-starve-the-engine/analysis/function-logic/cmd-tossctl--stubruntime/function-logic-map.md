# Function Logic Map: `stubRuntime`

> **테스트 코드다.** a102가 이 함수를 편집한 이유는 하나뿐이고, 프로덕션 판정은 들어 있지
> 않다. 이 묶음은 `check_analysis.py`가 요구하는 증거이며, 형식은 a056·a059의 선례
> (같은 이유로 만든 테스트 함수 묶음)를 따른다.

- Source: `cmd/tossctl/engine_test.go` (59-64)
- AST evidence: `ast.json` — AST 기준 branches **0** / returns 1 / calls 3
- Risk scan: `risk-pattern-report.md`
- source SHA-256: `ee98aa2936bff2145f0b98ff32889ad3865d4422ab60fdb6df532a7d0e6931c9`
- **a102가 바꾼 것**: §3.9(D5c) — 본문을 `stubRuntimeWithReady`에 위임한다. ready 인자를 버리던 자리가 A2의 생존 뮤테이션 N1·N5가 살던 곳이다. 이 헬퍼의 계약(기존 호출자에게는 그대로)은 불변이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 픽스처 | 이 change로 바뀌지 않음 | 같은 파일 안의 헬퍼 | `t.Fatalf`/`t.Errorf` |

## Branches and early returns

| Branch | 위치 | Mutation/side effect | Return/이탈 |
|---|---|---|---|
| B1 | 분기 없음 (AST branches 0) | 테스트 코드 — 부수효과 없음 | 이 change로 바뀌지 않음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ast.json`의 호출 3건 | 테스트 픽스처 조립과 단언 | 이 change로 바뀌지 않음 | ast.json |

## State mutations and fallbacks

- 없음 — 테스트 코드다. tempdir과 fake seam만 쓴다.
- 새 브로커 호출·config key·audit 레코드·라우트 없음.

## Safety conclusion

- Safe edit boundary: **시그니처 반영뿐.** 단언·시나리오·기대값은 이 change가 바꾸지 않았다.
- High-risk impact: **no** — 이 함수는 프로덕션 경로가 아니다. a102의 High-risk 판정은
  `internal-enginelock--hold`·`cmd-tossctl--runenginerun`·`cmd-tossctl--engineruntime`·
  `cmd-tossctl--runconsole` 네 묶음이 진다.
