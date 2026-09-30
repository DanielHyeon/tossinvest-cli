# Function Logic Map: `TestRiskBucketUnsafeEvidenceAndReleaseMethodsAreNotExported`

- Source: `internal/journal/risk_bucket_fill_test.go` (`324`–`349`)
- Qualified: `TestRiskBucketUnsafeEvidenceAndReleaseMethodsAreNotExported`
- AST evidence: `ast.json` (`source_sha256` 5dc1253874e74344…) — **편집 뒤**(시험 전용 census)
- Risk scan: `risk-pattern-report.md`
- AST branches 8 · return 0 · 호출 11

**역할.** 호출자가 권한을 쥐는 위험한 journal 메서드가 export 되지 않고 비시험 파일에서 불리지 않음을 문자열 census 로 지키는 시험.
a126 1.5 R7 이 census 목록에 `.releaseRiskBucketOwner(` 를 더했다 — tasks 3.1 면제 불가 의존(owner 해제 = 떠남의 생산 배선은 R3 해소 뒤)의
코드 tripwire. 비례 원칙(WORKFLOW 「비례 원칙」): 시험 전용 — 게이트가 수정 함수로 세므로 경량 번들을 둔다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `Journal` 메서드 집합 · 패키지 비시험 `.go` 원문 | 저장소 현재 트리 | reflect · `os.ReadFile` | `t.Fatal`(생산 호출 · export 발견) |

## Branches and early returns

> 표는 `analysis/harness/branch_table.py` 가 만들었다(시험 코드라 커버리지 블록 없음 — `—`).

| Branch | 종류 | 조건 (원문) | 창의 return | 진입 실측 |
|---|---|---|---|---|
| B1 | range | `:326` `for _, name := range []string{"CompleteRiskBucketFillActual", "ReleaseRiskBucketOrder"} {` | — | — |
| B2 | if | `:327` `if _, exists := typeOfJournal.MethodByName(name); exists {` | — | — |
| B3 | if | `:332` `if err != nil {` | — | — |
| B4 | range | `:335` `for _, path := range files {` | — | — |
| B5 | if | `:336` `if strings.HasSuffix(path, "_test.go") {` | — | — |
| B6 | if | `:340` `if err != nil {` | — | — |
| B7 | range | `:343` `for _, call := range []string{".completeRiskBucketFillActual(", ".releaseRiskBucketOrder(", ".releaseRiskBucketOwner("} { // a126 tasks 3…` | — | — |
| B8 | if | `:344` `if strings.Contains(string(raw), call) {` | — | — |

## Calls and live bindings

`reflect.TypeOf` · `MethodByName` · `filepath.Glob` · `os.ReadFile` · `strings.Contains` · `t.Fatal`. 생산 코드 · 브로커 호출 없음.

## State mutations and fallbacks

없음(읽기만).

## Safety conclusion

- **Safe edit boundary**: B7 의 목록 원소 하나(`.releaseRiskBucketOwner(`)와 같은 줄 주석만 더했다 — 분기 구조 불변. 비시험 파일에 그 호출이 한 줄이라도
  생기면 이 시험이 실패한다(변이 T1 CAUGHT, review 1.5.2). 문자열 census 라 간접 호출(함수 값 등)은 못 본다(codex 재리뷰 기록).
- **High-risk impact**: no — 시험. 생산 동작 변화 0.
