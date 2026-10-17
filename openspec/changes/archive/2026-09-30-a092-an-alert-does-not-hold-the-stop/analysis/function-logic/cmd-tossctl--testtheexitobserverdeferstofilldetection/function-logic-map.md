# Function Logic Map: `TestTheExitObserverDefersToFillDetection`

- Source: `cmd/tossctl/engine_test.go`
- AST evidence: `ast.json` — **편집 뒤**, :383–392, 분기 2 · source_sha256 `3d12ffb196ff…`, 추출 커밋 `c6e2e3ac`.
- Risk scan: `risk-pattern-report.md`
- 편집(착지 단위 ②의 부수 수정): SLO 배선 문자열 핀을 공백 무관 정규식으로 바꿈. a092 가 `ExitObserverOptions` 리터럴에서 `Announcer` 키를
  빼자 gofmt 정렬 폭이 바뀌어(`SLO:       ` → `SLO:      `) 문자열 일치가 깨졌음 — 판정 대상은 배선이지 정렬이 아님.
- 비례 원칙: 시험 함수 편집이라 High-risk 경로 · 게이트 판정 변경이 아님 → 변이 원장 · 다중 리뷰는 `not-applicable`. 대신 반증 1회(아래).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `engine.go` 원문 | `readSource` | 저장소 | 읽기 실패는 `t.Fatal`(헬퍼) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 정규식 `SLO:\s+detectorPressure\{detector: detector\}` 불일치 (:386) | `t.Error` | — | 자기 자신(반증: SLO 줄 삭제 사본에서 FAIL) |
| B2 | `p.detector.Health().EntryBlocked` 부재 (:389) | `t.Error` | — | 자기 자신(편집 전과 같음) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `readSource` | `engine.go` 원문 | 실패 시 Fatal | AST |
| `regexp.MustCompile(…).MatchString` | 공백 무관 배선 확인 | 상수 정규식 | AST |

## State mutations and fallbacks

- 없음(원문 읽기만).

## Safety conclusion

- Safe edit boundary: 판정 대상(SLO 에 `detectorPressure{detector: detector}` 가 배선됨)은 그대로, 공백만 너그럽게.
- High-risk impact: no — 시험 코드.
