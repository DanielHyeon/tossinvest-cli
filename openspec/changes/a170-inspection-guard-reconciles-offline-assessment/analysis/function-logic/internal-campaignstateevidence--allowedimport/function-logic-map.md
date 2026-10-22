# Function Logic Map: `allowedImport`

- Source: `internal/campaignstateevidence/authority_guard_test.go`
- AST evidence: `ast.json`(revision current, 분기 11 · return 8)
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a170-inspection-guard-reconciles-offline-assessment/base-commit.txt`
- 위험 등급: 시험 파일의 정적 가드 — 생산 경로 아님(High-risk 무접촉).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `path` | 저장소 루트 기준 슬래시 경로(`TestInspectionPackageHasNoRuntimeConsumers` 가 `filepath.ToSlash(rel)`) | 호출자 walk | 허용 목록 비교가 정확 일치라 다른 표기는 불허 쪽 |
| `source` | Go 원문 바이트 | 디스크 · 시험 리터럴 | 파싱 실패 → `(false, err)` |
| 허용 목록 | lane `offline_evidence.go` 셋 + `offlineAssessment`(a170) | 이 파일 상수 | 목록 밖 evidence import → 거절 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `parser.ParseFile` 오류 | 없음 | `false, err`(@26) | (기존) 파싱 오류 — walk 가 t.Fatal |
| B2 | `range f.Imports` | 반복 | — | 모든 import 시험 |
| B3 | `strconv.Unquote` 오류 | 없음 | `false, err`(@32) | 파서가 막아 도달 어려움(방어) |
| B4 | lane 패키지 import(접두 + `/continuationlane` · `/reversallane` · `/weeklyvaluelane`) | `laneAliases` 에 별칭 추가 | — | `TestGuardRejectsRuntimeCaptureUse` |
| B5 | 이름 있는 import(`spec.Name != nil`) | 별칭 = 그 이름 | — | `TestGuardRejectsRuntimeCaptureUse`(`import lane …`) |
| B6 | 별칭이 `.`(점 import) | 없음 | `false, nil`(@40) | `TestGuardRejectsRuntimeCaptureUse` 점 import · `TestOfflineAssessmentIsTheOnlyNamedConsumer` dotCapture(M5) |
| B7 | evidence import 이고 경로가 허용 목록 밖(a170: `offlineAssessment` 추가) | 없음 | `false, nil`(@45) | `TestInspectionPackageHasNoRuntimeConsumers` forged · `…OnlyNamedConsumer` other.go(M3) |
| B8 | (a170 신규) `offlinecontinuation` import | 없음 | `false, nil`(@49) | `…OnlyNamedConsumer` 엔진 forged 2 표기(M1) |
| B9 | (a170 신규) `path == offlineAssessment` — import 루프 **뒤** | 없음 | `true, nil`(@54) | `…OnlyNamedConsumer` 실제 원문 통과 · 실제 저장소 walk(M4 · M5) |
| B10 | `ast.Inspect` 콜백 — `CaptureOfflineEvidence` selector | — | — | `TestGuardRejectsRuntimeCaptureUse` |
| B11 | selector 의 X 가 lane 별칭 식별자 | `allowed = false` | (콜백 true) | `TestGuardRejectsRuntimeCaptureUse` · `…OnlyNamedConsumer` captureOnly(M2) |

끝 return(@64 콜백 내부 `true`, @66 `allowed, nil`).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `parser.ParseFile` | import · selector 를 AST 로 읽음(문자열 매칭이 아니라 이스케이프 표기도 정규화) | 오류 → 그대로 반환 | CodeGraph + AST |
| `strconv.Unquote` | import 경로 리터럴 해독(`e` 같은 이스케이프 포함) | 오류 → 그대로 반환 | AST |
| `ast.Inspect` | lane 별칭의 `CaptureOfflineEvidence` 참조 탐색 | 없음 | AST |

## State mutations and fallbacks

- 지역 `laneAliases` · `allowed` 만 바뀐다. 전역 상태 · 파일 쓰기 0.
- a170 이 바꾼 것: B7 조건에 `&& path != offlineAssessment` · B8 신설 · B9 신설. B9 는 B6(점 import) · B8 뒤라 평가 파일에도 그 둘은 적용된다.

## Safety conclusion

- Safe edit boundary: 이름 허용은 `assessment.go` 한 파일, 그 패키지의 런타임 import 는 새로 거절 — 가드는 넓어진 곳(B9)보다 좁아진 곳(B8)이 더 크다.
- High-risk impact: no(시험 파일, 생산 바이너리에 안 들어감).
