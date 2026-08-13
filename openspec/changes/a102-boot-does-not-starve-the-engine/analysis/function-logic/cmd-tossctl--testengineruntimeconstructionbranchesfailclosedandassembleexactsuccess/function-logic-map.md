# Function Logic Map: `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess`

> **테스트 코드다.** a102가 이 함수를 편집한 이유는 하나뿐이고, 프로덕션 판정은 들어 있지
> 않다. 이 묶음은 `check_analysis.py`가 요구하는 증거이며, 형식은 a056·a059의 선례
> (같은 이유로 만든 테스트 함수 묶음)를 따른다.

- Source: `cmd/tossctl/engine_runtime_branch_test.go` (49-132)
- AST evidence: `ast.json` — AST 기준 branches **7** / returns 2 / calls 26
- Risk scan: `risk-pattern-report.md`
- source SHA-256: `65b8fc0f0748a3a1841c3179b0cecc5da6f2be281218b4be75d371647dfff34c`
- **a102가 바꾼 것**: §3.9b — 성공 케이스 뒤에 **조립된 `RuntimeOptions.Recover`를 실제 실행**하는 단언을 더했다. 복구 시퀀스는 seam으로 대체하고(실계좌 조회 금지), ready가 run ctx를 취소하므로 루프는 즉시 취소로 돌아온다. A2의 마지막 생존 뮤테이션(N5, 본문에서 `ready = nil`)을 죽인다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 픽스처 | 이 change로 바뀌지 않음 | 같은 파일 안의 헬퍼 | `t.Fatalf`/`t.Errorf` |

## Branches and early returns

| Branch | 위치 | Mutation/side effect | Return/이탈 |
|---|---|---|---|
| B1 | `:78` range | 테스트 코드 — 부수효과 없음 | 이 change로 바뀌지 않음 |
| B2 | `:83` if | 테스트 코드 — 부수효과 없음 | 이 change로 바뀌지 않음 |
| B3 | `:91` if | 테스트 코드 — 부수효과 없음 | 이 change로 바뀌지 않음 |
| B4 | `:97` if | 테스트 코드 — 부수효과 없음 | 이 change로 바뀌지 않음 |
| B5 | `:123` if | 테스트 코드 — 부수효과 없음 | 이 change로 바뀌지 않음 |
| B6 | `:126` if | 테스트 코드 — 부수효과 없음 | 이 change로 바뀌지 않음 |
| B7 | `:129` if | 테스트 코드 — 부수효과 없음 | 이 change로 바뀌지 않음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ast.json`의 호출 26건 | 테스트 픽스처 조립과 단언 | 이 change로 바뀌지 않음 | ast.json |

## State mutations and fallbacks

- 없음 — 테스트 코드다. tempdir과 fake seam만 쓴다.
- 새 브로커 호출·config key·audit 레코드·라우트 없음.

## Safety conclusion

- Safe edit boundary: **시그니처 반영뿐.** 단언·시나리오·기대값은 이 change가 바꾸지 않았다.
- High-risk impact: **no** — 이 함수는 프로덕션 경로가 아니다. a102의 High-risk 판정은
  `internal-enginelock--hold`·`cmd-tossctl--runenginerun`·`cmd-tossctl--engineruntime`·
  `cmd-tossctl--runconsole` 네 묶음이 진다.
