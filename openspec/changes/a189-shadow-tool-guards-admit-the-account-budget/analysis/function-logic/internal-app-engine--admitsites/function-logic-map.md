# Function Logic Map: `admitSites` (internal/app/engine)

- Source: `internal/app/engine/strategy_dispatch_handoff_guard_test.go`(편집 뒤 797~833행). 편집 전 logic-map 산출물 `../../pre-edit-ast/engine--admitsites.ast.json`(788~799행, 분기 2: B1 range 790·B2 if 792, return 2).
- AST evidence: `ast.json`(편집 뒤, 분기 11, return 2 — 829 Inspect 콜백, 832 함수 끝)
- Risk scan: `risk-pattern-report.md`(일치 0)
- 원인·경계: a087 `5c0aeb2b8` 가 `exit_lower_limit.go:58` 에 `type lowerLimitAdmission interface{ Admit(context.Context) error }` 를 넣음 — 부모 PASS, 원인 커밋 FAIL(`../../boundary/boundary.txt`, 자리 `type 블록:Admit`).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `file`(파싱된 Go 파일) | 임의 | 호출자(`TestExactlyOneProductionSiteAdmitsIntoTheSeam` — 엔진 생산 파일 전수, fixture 시험 넷) | 판정은 호출자가 함 |
| `doors` | `strategyHandoffDoors`(`Admit`·`AdmitEachOwnerScope`) | 같은 파일 | — |
| 결과 | 「감싼 최상위 선언:문 이름」 목록 | 이 함수 | 호출자가 정확한 목록과 비교 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (800) | `range file.Imports` | — | — | 전부 |
| B2 (802) | import 갈래 | — | — | 전부 |
| B3 (803) | dot import(`.`) | `dotted = true` | — | `TestAdmitCensusSeesEverySpellingItClaims`(dottedCall) — 변이 g3 |
| B4 (805) | 경로 == `strategyhandoff` 또는 `/strategyhandoff` 로 끝남 | 지역 이름 기록 | — | `TestAdmitCensusSees…`(example.com 경로) — 변이 g5 |
| B5 (807) | 별칭 있음 | 지역 이름 = 별칭 | — | `TestAdmitCensusSees…`(sh) — 변이 g1; 기본 이름은 `TestExactlyOne…`·`TestA189AdmitCensusCountsTheUnaliasedSeamName` — 변이 g7 |
| B6 (814) | `range file.Decls` | 선언마다 `qualified` 새로 만듦 | — | 전부 |
| B7 (818) | 노드 type-switch | — | — | 전부 |
| B8 (819) | `*ast.SelectorExpr` | `qualified[Sel]` 기록 | — | `TestA189AdmitCensusIsConservativeInADotImportedFile`(b.Admit 안 셈) — 변이 g4 |
| B9 (821) | X 가 경계 지역 이름이고 Sel 이 문 | 자리 추가 | — | `TestExactlyOne…`(생산 두 자리) — 변이 g6, a1·a2·a4·a5·a6 |
| B10 (824) | `*ast.Ident` | — | — | 전부 |
| B11 (825) | dot import 파일이고 문 이름이고 선택자 Sel 아님 | 자리 추가 | — | `TestAdmitCensusSees…`(dottedCall)·`TestA189AdmitCensusIsConservative…`(bare·type 블록) — 변이 a3·g3·g4 |
| return (829) | Inspect 콜백 | — | true(계속 내려감) | — |
| return (832) | 끝 | — | 자리 목록 | — |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.Trim`·`strings.HasSuffix` | import 경로 해독·끝 원소 비교 | — | AST |
| `ast.Inspect` | 선언 하나를 끝까지 훑음(함수 본문 밖 `var` 선언 포함 — 앞 판본과 같은 범위) | — | AST |
| `declName` | 자리 이름 | 무편집 | AST |

## State mutations and fallbacks

- 지역 맵 `seamNames`·`qualified`, 결과 슬라이스만. 바깥 상태 없음.

## 편집

- 앞 판본(B1 range·B2 if 「이름이 문과 같은 모든 `*ast.Ident`」)을 둘로 좁힘(design D13): 경계 패키지 지역 이름의 선택자 참조(B9)와 dot import 파일의 선택자 아닌 맨 이름(B11).
- 잃는 검출(`analysis/mutation-2/`, design R4): 경계 밖 이름으로 문을 다시 부르는 자리 — 허용 자리 안에서 문 이름 필드로 빼돌린 뒤 다른 자리에서 부름(l1), 엔진 밖 래퍼 패키지의 같은 이름(l3, `TestOnlyTheEngineImportsThisSeam` 이 대신 막음). 다른 이름 필드 빼돌리기(l2)는 앞 판본도 못 잡던 것.

## Safety conclusion

- Safe edit boundary: 시험 도우미 본문. 운영 엔진 파일 무편집.
- High-risk impact: 주문 진입 경계(strategyhandoff)의 자리 census. 진짜 둘째 자리(호출·별칭·dot·함수 값·패키지 var·둘째 문)는 변이 a1~a6 로 잡힘을 잼. 이 census 는 구조 검사이고 성질을 만드는 것은 봉투 서명임(같은 파일 주석) — 그 서명 검사(`TestOnlyTheSeamsEnvelopeCanReachTheSharedDispatch`)는 무편집.
