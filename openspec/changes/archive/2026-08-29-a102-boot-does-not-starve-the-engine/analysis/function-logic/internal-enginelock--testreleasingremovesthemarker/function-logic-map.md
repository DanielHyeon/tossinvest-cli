# Function Logic Map: `TestReleasingRemovesTheMarker`

> **테스트 코드다.** a102가 이 함수를 편집한 이유는 하나뿐이고, 프로덕션 판정은 들어 있지
> 않다. 이 묶음은 `check_analysis.py`가 요구하는 증거이며, 형식은 a056·a059의 선례
> (같은 이유로 만든 테스트 함수 묶음)를 따른다.

- Source: `internal/enginelock/enginelock_test.go` (169-184)
- AST evidence: `ast.json` — AST 기준 branches **3** / returns 0 / calls 12
- Risk scan: `risk-pattern-report.md`
- source SHA-256: `d460165ffc609568b4ff940dd12227409e2a20e802f9297ebdc1fa5b01945c1d`
- **a102가 바꾼 것**: `enginelock.Hold`의 D4 반환(핸들)에 맞춰 수신 변수와 두 번의 해제 호출을 바꿨다. 멱등 해제를 고정한다는 이 테스트의 주장은 그대로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 픽스처 | 이 change로 바뀌지 않음 | 같은 파일 안의 헬퍼 | `t.Fatalf`/`t.Errorf` |

## Branches and early returns

| Branch | 위치 | Mutation/side effect | Return/이탈 |
|---|---|---|---|
| B1 | `:172` if | 테스트 코드 — 부수효과 없음 | 이 change로 바뀌지 않음 |
| B2 | `:178` if | 테스트 코드 — 부수효과 없음 | 이 change로 바뀌지 않음 |
| B3 | `:181` if | 테스트 코드 — 부수효과 없음 | 이 change로 바뀌지 않음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ast.json`의 호출 12건 | 테스트 픽스처 조립과 단언 | 이 change로 바뀌지 않음 | ast.json |

## State mutations and fallbacks

- 없음 — 테스트 코드다. tempdir과 fake seam만 쓴다.
- 새 브로커 호출·config key·audit 레코드·라우트 없음.

## Safety conclusion

- Safe edit boundary: **시그니처 반영뿐.** 단언·시나리오·기대값은 이 change가 바꾸지 않았다.
- High-risk impact: **no** — 이 함수는 프로덕션 경로가 아니다. a102의 High-risk 판정은
  `internal-enginelock--hold`·`cmd-tossctl--runenginerun`·`cmd-tossctl--engineruntime`·
  `cmd-tossctl--runconsole` 네 묶음이 진다.
