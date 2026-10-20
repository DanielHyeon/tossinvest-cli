# a162 추가 통합검증 영수증

## 범위

- test-tagged `internal/weeklyvaluelane/strategyflow_testseam.go`만 확장했다. 기본 빌드의 `GoFiles`에는 이 파일이 없다. 결과는 [weekly-default-build-gofiles.txt](verification-logs/weekly-default-build-gofiles.txt)에 보존했다.
- 신규 반례와 adapter 검증은 `internal/weeklyvaluelane/strategyflow_followon_test.go` 및 기존 test-tagged `internal/offlinecontinuation/assessment_integration_test.go`에 한정했다.
- `internal/offlinecontinuation/assessment.go`는 변경하지 않았다. 현재 파일 해시는 [source-sha256sums.txt](source-sha256sums.txt)에 있다.

## 검증한 계약

- KR/US ordinal 3..7은 tagged fixture에서 native reservation과 atomic positive fill로 k-1개 prefix를 만든다. 각 prefix는 누적 state/risk를 사용한다. `PositiveLegCount=k-1`, 이전 ordinal `CONSUMED`, 현재 ordinal `ACTIVE`, `FilledMinor=20*(k-1)`, `HeldMinor=0`을 확인했다. adapter에서는 `B=1000`, `F=20*(k-1)`, `H=0`, `R=1000-F`, `U=100`, `q_candidate=q_offline=2`, planned risk 200을 정확히 단언했다.
- USD quote / KRW account fixture는 기존 tagged official-FX evidence와 native FX 생성만 사용한다. capture된 FX가 존재하며, adapter의 KRW U=14300 산술 결과는 `q_candidate=q_offline=2`, planned risk 28600이다. USD caller assumption은 수량 없이 거절된다.
- test-only FX rate 변조는 replacement seal 없이 native proposal과 adapter 모두에서 `RR_CALCULATION_INVALID`와 수량 부재를 확인한다. 이는 test input invalidation이며 권한이나 seal을 재발급하지 않는다.
- 각 ordinal 3..7의 현재 reservation map entry 제거는 seal을 재발급하지 않는 reservation-state integrity corruption이다. adapter는 정확한 native `WEEK_RESERVATION_MISSING` refusal과 수량 부재를 보인다. lookup-only missing branch를 별도로 주장하지 않는다.
- fresh control 뒤 synthetic evidence freshness만 경계 밖으로 바꾼 경우 native `EVIDENCE_STALE`가 adapter 결과에도 유지된다.
- 허용·거절 결과 모두 `InspectionOnly=true`, `LedgerCurrentness=unverified`, `ExecutionAuthority=false`다.

## 실행 결과

| 명령 | 결과 | 원문 |
| --- | --- | --- |
| `go test -count=1 ./internal/offlinecontinuation` | exit 0 | [offline-default-final.log](verification-logs/offline-default-final.log) |
| `go test -count=1 -tags tossos_testseams ./internal/offlinecontinuation` | exit 0 | [offline-tagged-final.log](verification-logs/offline-tagged-final.log) |
| tagged weekly focused tests | exit 0 | [weekly-targeted-final.log](verification-logs/weekly-targeted-final.log) |
| exact strategyflow consumer test | exit 0 | [strategyflow-consumer-final.log](verification-logs/strategyflow-consumer-final.log) |
| adapter FX-tamper and missing-reservation focused tests | exit 0 | [adapter-negative-final.log](verification-logs/adapter-negative-final.log) |
| exact adapter ordinal/risk and reservation-integrity oracle | exit 0 | [adapter-oracle-final.log](verification-logs/adapter-oracle-final.log) |

모든 최종 명령은 `-count=1`을 사용했다. 원문 해시는 [sha256sums.txt](verification-logs/sha256sums.txt)에 있다.

문서 최종 검증도 exit 0: [strict](verification-logs/final-strict.log), [PM check](verification-logs/final-pm-check.log), [gofmt·diff check](verification-logs/final-format-check.log). 제품 시험은 이 단계에서 재실행하지 않았다.

## RED 이력

- 최초 native prefix fixture는 ordinal 3..7에서 `WEEK_RESERVATION_CONFLICT`로 실패했다. raw receipt는 [weekly-prefix-red.log](verification-logs/weekly-prefix-red.log)다.
- 이후 test seam이 prior native reservation/fill 전이를 만들도록 한정 보강했다.
- 초기 구현 중 잘못된 refusal constant/import 때문에 발생한 컴파일 실패 로그는 제품 RED로 사용하지 않았다. 최종 영수증은 위 exit-0 로그만 판정 근거로 사용한다.

## 한계

- 모든 input은 tagged same-process synthetic fixture다. ordinal 3..7 실제 시장 관측·캘린더 시간순·계좌 최신성을 증명하지 않는다.
- fixture가 지원하는 follow-on 범위는 이 검증의 ordinal 2..7이며, 이번 assertion은 3..7이다. 새 runtime fixture나 authority API는 추가하지 않았다.
- FX mutation은 native refusal 검증용 test-only input mutation이다. live FX, broker, journal, dispatch, qFinal 또는 계좌 현재성을 호출·검증하지 않았다.
- 전체 suite, race, gate는 실행하지 않았다.
