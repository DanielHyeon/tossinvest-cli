# Function Logic Map: `BuildAttestation`

- Source: `internal/soak/attest.go`
- Function: `internal/soak/attest.go:BuildAttestation`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `attest-covers-supervised-mutations`
- High-risk: **yes** — 이 함수의 반환값이 엔진 자동화 게이트가 읽는 파일이다.

이 change가 바꾼 것: 두 번째 증거원(감독 검증)을 받아 mutation endpoint를 합치고,
note 문장을 실제 커버 상태에 맞춘다. 기존 분기(B1~B3)와 read 경로는 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s Summary` | soak 기록 요약 | `Summarize(cycles)` | `Evaluate` 미충족이면 `ErrIncomplete`, **아무것도 쓰지 않는다** |
| `c Criteria` | `withDefaults()` 적용 후 전부 양수 | `DefaultCriteria()` | 0/음수는 기본값으로 대체 |
| `now` | 발급 시각 | 호출자 | 미래 증거 판정의 기준 |
| `supervised []attest.Proof` | nil 허용 | 검증 기록(`verifylive.SucceededEndpoints`) | nil이면 읽기만 담아 발급 — 기존 동작 |
| `s.AccountRef` | 마스킹되지 않은 계좌 | soak 기록 | 감독 증거 결속의 기준 |

불변식: **각 증거원은 자기 몫만 증명한다.** soak 기록의 비-GET은 거부되고(B2·B3),
감독 기록의 `LiveOnlyEndpoints()` 밖은 거부된다(`acceptSupervised`). 두 거부가 대칭이며
둘 중 하나만 있으면 반대편 구멍이 열린다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `!ok` — soak이 기준 미달 (line 179) | 없음 | `ErrIncomplete` + 사유 목록. **파일 미기록** | `TestBuildAttestationRefusesAnIncompleteSoak` |
| B2 | `range endpoints` (188) | 없음 | — | `TestBuildAttestationPassesTheEngineInterlock` |
| B3 | 비-GET이 soak 결과에 있다 (189) | 없음 | error — 읽기 전용 도구의 기록에 mutation은 오염 | `TestBuildAttestationNeverClaimsAnEndpointItDidNotExercise` |
| B4 | `acceptSupervised` 오류 (197) | 없음 | error. **이 change가 추가** | `TestSupervisedEvidenceCannotStandInForTheSoak`, `TestSupervisedEvidenceIsClosedToOtherMutations`, `TestSupervisedEvidenceFromAnotherAccountRefusesTheIssue` |
| B5 | `range accepted` (200) | `endpoints` append | 감독 증거가 endpoint 집합에 합쳐진다. **이 change가 추가** | `TestSupervisedEvidenceCompletesTheEnginesRequiredSet` |
| B6 | `notes` 비어 있지 않음 (211) | 없음 | 운영자 메모를 앞에 붙인다 | `TestBuildAttestationCarriesTheMeasuredRate` |

early return 2개(B1·B4). 둘 다 **아무것도 쓰지 않는 방향**이다.

`acceptSupervised`(신규 leaf)의 판정 — 거부와 건너뜀을 의도적으로 구분한다:

| 상황 | 처리 | 근거 | 테스트 |
|---|---|---|---|
| `LiveOnlyEndpoints()` 밖 | **거부** | 감독 1회가 무인 4일을 대신할 수 없고, 요구되지 않은 것을 주장해서도 안 된다 | `TestSupervisedEvidenceCannotStandInForTheSoak`, `...IsClosedToOtherMutations` |
| 계좌 불일치 | **거부** | 기대 경로에 다른 계좌 기록 = 설정 오류. 건너뛰면 "증거 없음"과 구별되지 않는다 | `TestSupervisedEvidenceFromAnotherAccountRefusesTheIssue` |
| 유효 기간 밖 | 건너뜀 | 평범한 상태(검증 후 시간이 지남). 게이트의 `MissingEndpoints`가 결과를 보고한다 | `TestSupervisedEvidenceOlderThanTheValidityIsNotEvidence` |
| 미래 시각 | 건너뜀 | 신뢰할 수 없는 시계는 아무것도 증명하지 못한다 | `TestSupervisedEvidenceFromTheFutureIsNotEvidence` |
| 중복 endpoint | 건너뜀 | 첫 항목이 이긴다 | 위 테스트들이 개수로 고정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.withDefaults` | 기준값 정규화 | 순수 | ast.json |
| `s.Evaluate` | soak 충족 판정 + 사유 | 순수, error 없음 | ast.json |
| `s.Window.SuccessfulEndpoints` | soak이 증명한 읽기 | 순수 | ast.json |
| `strings.HasPrefix` | soak 결과의 비-GET 탐지 | 순수 | ast.json |
| `acceptSupervised` | 감독 증거 판정. **이 change 신규** | 순수, error 반환 | ast.json |
| `mutationNote` | 커버 상태 문장. **이 change 신규** | 순수 | ast.json |
| `append`/`fmt`/`strings` | 조립 | — | ast.json |

라이브 바인딩 없음 — 브로커·네트워크를 호출하지 않는다. 새 브로커 호출 0건(§0.4).
파일 쓰기는 호출자(`cmd/tossctl/soak.go`)가 `attest.Save`로 한다.

## State mutations and fallbacks

- 지역 slice `endpoints`·`accepted`만 변경한다. `s`·`c`·`supervised`를 변형하지 않는다
  (`acceptSupervised`는 `p`의 복사본에 `p.Endpoint`를 정규 철자로 덮어쓴 뒤 append한다).
- fallback: 감독 증거가 없거나 전부 창 밖이면 **읽기만** 담아 발급한다 — 이 change 이전과
  같은 결과이며 인터록이 부족을 보고한다.
- 이 change는 endpoint 집합을 **넓힐 수만** 있고, 넓어지는 경로가 `LiveOnlyEndpoints()`
  두 개로 닫혀 있다.

## Safety conclusion

- Safe edit boundary: B4·B5 추가와 note 문장. B1~B3(soak 판정·비-GET 거부)은 무변경.
- High-risk impact: **yes** — 게이트 5절을 만족 가능하게 만든다. 완화:
  - **이 change 단독으로는 게이트가 열리지 않는다.** 9절 `const profileProtection =
    ProtectionUnwired`(interlock.go:175)는 설정으로 만족 불가능하며
    `TestTheGateRefusesWithoutBrokerSideProtection`이 그것을 고정한다.
  - 실릴 수 있는 것은 사람이 배치 승인한 실행이 **성공시킨** 호출뿐이고, 목록·계좌·시각
    세 조건이 각각 거부 테스트를 갖는다.
  - 최악의 오작동 방향은 "증명된 것을 못 싣는다"(게이트가 계속 거부 — 안전 방향)이다.
- 회귀 위험: `LiveOnlyEndpoints()`가 넓어지면 감독 증거의 허용 범위도 같이 넓어진다.
  2c가 그 목록을 늘릴 때 이 함수의 거부 테스트가 함께 갱신돼야 한다.
