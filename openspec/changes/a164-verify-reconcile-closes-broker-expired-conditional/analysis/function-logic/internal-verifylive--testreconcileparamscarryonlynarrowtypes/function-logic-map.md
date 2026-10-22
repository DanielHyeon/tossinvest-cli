# Function Logic Map: `TestReconcileParamsCarryOnlyNarrowTypes`

- Source: `internal/verifylive/reconcile_seal_test.go` (108-133)
- Qualified function: `TestReconcileParamsCarryOnlyNarrowTypes`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` 425387b8af8ff54a…)
- AST evidence: `ast.json` — AST branches 3, 반환 4(줄 114, 115, 129, 130), 호출 이름 9
- Risk scan: `risk-pattern-report.md`
- 이 change 에서의 역할: 봉인 2층 시험(시험 전용). a164 편집: 허용 타입에 `bool` 하나(ReconcileParams.Expired — 값 타입, 메서드 없음).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t *testing.T` | — | go test | — |

## Branches and early returns

AST 열거 순서 그대로다. 「Required test」 칸은 손으로 고르지 않았다 — 시험 함수 하나씩 `-coverprofile` 을 떠서 분기 본문
블록이 실행된 시험만 적었다(측정 도구 `analysis/measure/` — `pertest.sh` · `btm_cov.py` · `gen_flm.py`).

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | for 줄 119: `for i := 0; i < typ.NumField(); i++ {` | 없음 | `f := typ.Field(i)` | `TestReconcileParamsCarryOnlyNarrowTypes` (시험 파일 함수 — coverprofile 계측 밖, 실행 PASS 가 증거) |
| B2 | if 줄 121: `if f.Anonymous {` | 없음 | `t.Errorf("%s embeds %v — embedding promotes methods past the census", typ.Name(), f.Type)` | `TestReconcileParamsCarryOnlyNarrowTypes` (시험 파일 함수 — coverprofile 계측 밖, 실행 PASS 가 증거) |
| B3 | if 줄 124: `if !ok(f.Type) {` | 없음 | `t.Errorf("%s.%s has type %v, outside the narrow set", typ.Name(), f.Name, f.Type)` | `TestReconcileParamsCarryOnlyNarrowTypes` (시험 파일 함수 — coverprofile 계측 밖, 실행 PASS 가 증거) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `reflect.TypeOf` | 필드 타입 대조 | 순수 | AST |

AST 호출 이름 전체: `Elem`, `check`, `ft.Kind`, `ok`, `reflect.TypeOf`, `t.Errorf`, `typ.Field`, `typ.Name`, `typ.NumField`.

## State mutations and fallbacks

- 없음.

## Safety conclusion

- bool 은 메서드가 없어 쓰기 경로를 열 수 없음. Approval·Account 는 여전히 스칼라만.

## a164 GREEN 재추출

- 편집 전(구현 base `aec9bde67`) 분기 열 길이 3 에서 편집 뒤 3 로. `source_sha256` 4e72a0640290… → 425387b8af8f….
- 옛/새 분기 열(kind + 원문 줄)을 difflib 으로 정렬한 결과: equal 옛 B1..B3 = 새 B1..B3.
- 위 표의 B 번호는 **새** AST 기준이고, 시험 칸은 GREEN 코드에서 시험별 프로필을 다시 떠 측정한 값이다(손 재번호 없음).
