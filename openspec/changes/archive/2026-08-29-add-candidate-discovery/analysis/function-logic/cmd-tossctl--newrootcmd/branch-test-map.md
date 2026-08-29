# Branch Test Map: `newRootCmd`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 출력 포맷 파싱 실패 → 명령 미실행 | cobra PersistentPreRunE 경로 | no (기존 동작) | yes |
| B2 | update 캐시 경로 확보 시 백오프 게이트 적용 | `TestExpiryWarningRespectsBackoffGate` | no (기존 동작) | yes |
| B3 | config status 확보 시 legacy 경고 | `TestConfigLegacyWarningOnLegacyFields` | no (기존 동작) | yes |
| B4 | table 모드에서만 온보딩 힌트 | `TestConfigLegacyWarningSilentInJSONAndSkipCommands` | no (기존 동작) | yes |
| B5 | OpenAPI 경로 해석 성공 | 온보딩 힌트 경로 | no (기존 동작) | yes |
| B6 | 자격증명 존재 → `HasOfficialCreds` | 동일 | no (기존 동작) | yes |
| B7 | 힌트 조건 충족 → stderr 한 줄 | `shouldHintOnboarding` 단위 테스트 | no (기존 동작) | yes |

이 change가 추가한 등록 한 줄의 RED/GREEN은 `TestTheDiscoveryCommandsDeclareThemselvesReadOnly`가 소유한다 — 등록 전에는 `findCommandPath`가 `tossctl candidate scan`을 찾지 못해 FAIL, 등록 후 PASS.
