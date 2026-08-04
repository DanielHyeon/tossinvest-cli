# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a075-alerts-turn-on-with-one-button/base-commit.txt`
- 위험 등급: High — 이 함수가 콘솔 프로세스에 **어떤 능력을 주입할지** 결정한다.
  여기서 seam 하나를 잘못 배선하면 화면이 하지 못해야 할 일을 하게 된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cmd.Context()` | nil 가능 | cobra | nil이면 `context.Background()` (B1) |
| `root.configDir` | 경로 또는 빈 값 | 플래그 | seam이 nil로 배선되지 않음 |
| `opts.*` | 플래그 값 | CLI | 각 검증이 B2–B8에서 error 반환 |
| journal/engine 경로 | 파일 시스템 | 프로필 | 없으면 화면이 미배선을 렌더 |
| `TOSSOS_CONTAINER` | `"1"` 또는 없음 | 환경 | 컨테이너면 self-update 배선 안 함 |

**불변식 (유지)**: seam이 없으면 **typed nil이 아니라 nil**을 넣는다. 파일의 여러
`console*Seam` 헬퍼가 전부 같은 모양이고, 이유가 주석에 적혀 있다 — 인터페이스 안의
typed nil은 화면에 배선된 것처럼 보이고 첫 클릭에서 실패한다.

**불변식 (유지)**: 이 함수는 계좌를 건드리지 않는다. 주입되는 것은 읽기 seam과
설정 편집 seam이며, 엔진 프로세스가 주문 능력을 갖는지는 §0.7로 승인된 게이트 설정과
기동 인터록이 결정한다.

**a075가 바꾸는 것**: `console.Options` 리터럴에 `Notifications:
consoleNotificationSeam(root)` **한 줄**. 조건도, 반환도, 순서도 바뀌지 않는다.

**a075가 바꾸지 않는 것**: 당시의 36개 분기 전부. 현재의 B37–B41은 이후 a072가
추가한 strategy projection endpoint 탐색이며 이 재기준화에서 그대로 보존한다.

## Branches and early returns

현재 41개 분기 중 a075가 편집한 것은 없다. 아래는 a075 당시 36개와 이후 통합된
strategy projection 5개가 서로 독립임을 보이기 위한 전수 표다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (204) | `ctx == nil` | ctx 대체 | — | 기존 |
| B2 (213) | 프로필 해석 실패 | — | error | 기존 |
| B3 (218) | 옵션 검증 실패 | — | error | 기존 |
| B4 (222) | 옵션 검증 실패 | — | error | 기존 |
| B5 (226) | 옵션 검증 실패 | — | error | 기존 |
| B6 (230) | 옵션 검증 실패 | — | error | 기존 |
| B7 (234) | 옵션 검증 실패 | — | error | 기존 |
| B8 (239) | remote 옵션 해석 실패 | — | error | 기존 |
| B9 (247) | `journalPath != ""` | 원장 리더 배선 | — | 기존 |
| B10 (249) | 원장 열기 실패 | 사유 기록 | — | 기존 |
| B11 (252) | 그 외 | 리더 주입 | — | 기존 |
| B12 (256) | 두 번째 열기 실패 | 사유 기록 | — | 기존 |
| B13 (259) | 그 외 | 리더 주입 | — | 기존 |
| B14 (270) | 엔진 journal 디렉터리 있음 | 경로 확정 | — | 기존 |
| B15 (273) | 없음 | 사유 기록 | — | 기존 |
| B16 (281) | `TOSSOS_CONTAINER == "1"` | self-update 미배선 | — | 기존 |
| B17 (284) | 그 외 · self 경로 실패 | 사유 기록 | — | 기존 |
| B18 (284) | self 경로 실패 | 사유 기록 | — | 기존 |
| B19 (286) | 그 외 | 계속 | — | 기존 |
| B20 (288) | candidate 확인 실패 | 사유 기록 | — | 기존 |
| B21 (296) | 그 외 | updater 확정 | — | 기존 |
| B22 (290) | updater 생성 실패 | 사유 기록 | — | 기존 |
| B23 (292) | 그 외 | updater 확정 | — | 기존 |
| B24 (299) | `updater != nil` | 업데이트 seam 주입 | — | 기존 |
| B25 (303) | 검사 실패 | 사유 기록 | — | 기존 |
| B26 (305) | 그 외 | 릴리스 seam 주입 | — | 기존 |
| B27 (312) | `engineDir != ""` | 엔진 마커 경로 | — | 기존 |
| B28 (315) | 마커 해석 실패 | 사유 기록 | — | 기존 |
| B29 (331) | `engineBoot != nil` | 자동 시작 seam 주입 | — | 기존 |
| B30 (338) | `engineBootNote != ""` | 기동 결과 문구 | — | 기존 |
| B31 (346) | `engineDir != ""` | 제어 평면 탐색 | — | 기존 |
| B32 (348) | descriptor 파일 있음 | dial | — | 기존 |
| B33 (360) | stat이 NotExist가 아님 | 사유 기록 | — | 기존 |
| B34 (350) | dial 실패 | 사유 기록 | — | 기존 |
| B35 (352) | 그 외 | commander 주입 | — | 기존 |
| B36 (360) | stat 오류 | 사유 기록 | — | 기존 |
| B37 (372) | strategy descriptor 파일 있음 | projection endpoint dial | — | strategy runtime integration 테스트 |
| B38 (379) | strategy descriptor stat 오류 분기 | — | — | strategy runtime integration 테스트 |
| B39 (374) | projection endpoint dial 실패 | dormant 사유 기록 | — | strategy runtime integration 테스트 |
| B40 (376) | 그 외 | strategy runtime reader 주입 | — | strategy runtime integration 테스트 |
| B41 (379) | stat이 NotExist가 아님 | dormant 사유 기록 | — | strategy runtime integration 테스트 |

**a075가 추가한 분기는 없음.** 편집은 구조체 리터럴의 필드 하나이며, 필드 대입은
조건이 아니다. B37–B41은 a072의 후속 통합분이다.
`consoleNotificationSeam`이 내부에서 nil을 판단하지만 그것은 **그 함수의** 분기이고
새 파일에 있다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `consoleSettingsSeam` 외 다수 | 각 설정 편집 seam | nil 반환 = 미배선 | AST |
| `consoleGateSwitchSeam` | 게이트 스위치 | nil 반환 = 미배선 | AST |
| **`consoleNotificationSeam`** (신규) | 알림 설정 | nil 반환 = 미배선 | 신규 |
| `console.ListenAndServe` | 서비스 | 루프백 외 거부 | AST |

`consoleNotificationSeam`은 이 함수가 부르는 다른 열 몇 개의 seam 헬퍼와 **같은
계약**이다: 프로필을 해석하지 못하면 nil, 해석하면 구현. 그 nil 판정이 구체 포인터
위에서 일어나므로 인터페이스에 typed nil이 들어가지 않는다.

## State mutations and fallbacks

- 프로세스 지역 상태만 만든다. 이 함수 자체는 파일을 쓰지 않는다 — 쓰기는 주입된
  seam이 클릭에 응답할 때 일어난다.
- 새 seam이 nil이면 알림 카드는 "미배선" 사유를 렌더하고 버튼을 렌더하지 않는다.
  프로필 없는 실행에서 편집 전과 관측 가능한 차이가 없다 (§0.2).
- 새 seam은 `Enable`이 인자를 받지 않으므로, 이 함수가 주입하는 능력은 "누를 수 있다"
  이지 "무엇을 쓸지 정할 수 있다"가 아니다.

## Safety conclusion

- Safe edit boundary: `console.Options` 리터럴의 필드 한 줄.
- High-risk impact: **no** — 주문·손절·사이징·Guardian·원장·인증·체결 경로에 닿지
  않는다. 기존 seam의 배선을 바꾸지 않는다.
- §0.2: `Notifications`가 nil인 빌드(프로필 미해석)는 편집 전과 동일하게 동작한다.
- §0.8: 이 함수는 토큰도 채널도 만지지 않는다. 토큰은 seam의 `Test`가 프로세스 환경에서
  읽고, 채널은 seam의 `Enable`이 만든다.
