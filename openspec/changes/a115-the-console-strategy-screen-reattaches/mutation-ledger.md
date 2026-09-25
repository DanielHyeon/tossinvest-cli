# a115 뮤테이션 원장

측정 2026-09-26, 비root, `analysis/harness/run.sh <id> <K*.py|none> <rev>`. 하네스 규율:

- 변이는 저장소가 아니라 **커밋된 트리의 사본**(`git archive <rev> go.mod go.sum cmd internal docs/api`)에서만
  한다 — 공유 워크트리의 병행 세션(a119·a124) 미커밋 편집이 판정에 섞이지 않는다. 사본 이름에 pid, 한 번에 한 판.
- 판정은 rc 가 아니라 출력: FAIL/panic 줄 → CAUGHT, `[build failed]` → BROKEN, 네 패키지(cmd/tossctl ·
  internal/console · internal/httpapi · internal/strategyprojection) 모두 ok → SURVIVED.
- 매 판 끝에 실제 저장소의 대상 7파일이 시작 커밋과 같은지 단언(`leak=[]` — 전 판 0).
- 시험 목록 `analysis/harness/tests.txt`.

**무변이 대조군이 먼저 빨갰다**: 1판 첫 대조군이 `TestOpenAPIStrategyRuntimeIsStrictPairedReadOnlyProjection` 에서
`../../docs/api/openapi-v1.json` 부재로 FAIL — 사본에 `docs/api` 를 더한 뒤 대조군 green(`none rc=0`). 대조군 없이
돌렸으면 모든 변이가 CAUGHT 로 찍혔을 것이다.

## 1판 — `900d7582` (구현 커밋)

| id | 변이 | 결과 | 죽인 테스트 |
| --- | --- | --- | --- |
| K1 | 부팅 nil 접힘 재도입(부팅 해석이 live 아니면 자리를 비움 — 편집 전 B35 모양) | CAUGHT(경합 의존) | `…ShowsADeadDescriptorAsUnreachable` |
| K2 | sentinel 대신 nil(dial 실패 → 부재) | CAUGHT | `…ShowsADeadDescriptorAsUnreachable` · `TestTheConsoleResolutionMatchesTheDaemons`(잔재 둘) |
| K3 | 화면 nil 판정 복귀 — 전략 페이지 | CAUGHT | `…RecoversWhenTheEngineStartsLater` · `TestAnUnconfiguredWrapperStillRendersDormant` · `TestAConfiguredButUnreachableWrapperRendersUnavailable` |
| K4 | 화면 nil 판정 복귀 — 설정 요약 | CAUGHT | `TestTheSummaryAsksThePresenceSignalToo` |
| K5 | 부재 신호 무시(판정이 nil 만 본다) | CAUGHT | a109 `TestTheAbsenceSignalIsOneJudgementForEveryConsumer` 외 5 |
| K6 | 펌프 제거 | CAUGHT | `…AttachesWithoutAnyRender` · `…ReattachesAfterTheEngineRestarts` · `…RetryResolutionIsSilent` |
| K7 | runConsole 에 부팅 1회 `strategyprojectionrpc.Dial` 재도입 | CAUGHT | `TestRunConsoleNeverDialsTheStrategyProjectionItself` |
| K8 | 펌프가 시도 대상(`failed`)일 때만 wake(a109 G2 가 기각한 게이트, freeze P1-2) | CAUGHT | `…ReattachesAfterTheEngineRestarts`(렌더 없는 재시작) |
| K9 | interval≤0 가드 제거 | CAUGHT | `TestTheConsoleStrategyPumpNeedsAPositiveInterval` |
| K10 | 재시도 경고를 errOut 으로(discard 제거) | CAUGHT | `TestTheConsoleRetryResolutionIsSilent` |
| K11 | 부팅 문구 원복(「dormant로 뜬다」) | CAUGHT | `TestTheConsoleBootWarningDoesNotPromiseDormant` |
| K12 | 펌프가 콘솔 ctx 종료를 무시(goroutine·ticker 누수) | **SURVIVED** | — |
| K13 | httpapi 가 위임 대신 판정 사본(오늘은 같은 답) | **SURVIVED** | — |
| K14 | presence 인터페이스 두 벌(alias 대신 새 정의) | CAUGHT | `TestTheAbsenceJudgementIsOneForBothPackages`(reflect 동일성) |
| K16 | 페이지가 부재 신호를 두 번 묻는다(부작용 두 배) | CAUGHT | `…StillRendersDormant` · `…RendersUnavailable`(asked==1) |
| K17 | wrapper presence 메서드 개명 + httpapi 쪽 기존 `var _` 제거(a115 결속만 남김) | BROKEN(의도된 포획) | 생산 빌드 `console_strategy_attach.go:135` 가 컴파일 거절 — 컴파일 결속의 목적 그대로 |
| K18 | 부팅이 못 붙으면 nil 반환(「언제나 non-nil」 위반) | CAUGHT | `…RecoversWhenTheEngineStartsLater` 외 3(a115Boot 의 nil 검사) |

1판: 17 중 CAUGHT 14 · BROKEN 1(의도) · **SURVIVED 2**, 그리고 K1 의 포획이 **경합에 기댔다**(펌프 첫 틱 10ms·첫 렌더
wake 가 빈 자리를 sentinel 로 승격(a109 G1)하기 전에 화면을 그려야 빨개진다).

## 2판 — `aefd7356` (생존자 시험 추가 뒤)

추가한 시험(생산 코드 무변경):

- K12 → `TestTheConsoleStrategyPumpReturnsWhenTheConsoleEnds` — 펌프를 직접 돌려 ctx 종료 뒤 1s 안에 복귀하는지.
  시도 수로는 누수가 안 보였다(wake 가 ctx 를 보므로 시도는 멈춘다).
- K13 → `TestTheAbsenceIsJudgedInOnePlace`(internal/strategyprojection) — 모듈 생산 Go 파일 전부에서 presence 타입
  단언·`StrategyRuntimeConfigured()` 호출이 이 패키지 밖에 0. 세는 범위는 모듈(함수 범위는 헬퍼 하나로 무너진다),
  양성 대조군(센 파일 ≥100 · 이 패키지 안 질문 ≥2)으로 계측기가 눈멀지 않았음을 먼저 단언.
- K1 결정적 핀 → `TestTheConsoleBootKeepsAnUnreachableEndpointUnreachable` — 간격 1h(틱 없음)·렌더 없이 부팅 직후
  자리가 sentinel 인지.

| id | 2판 결과 | 죽인 테스트(2판) |
| --- | --- | --- |
| none | green(`rc=0`, 네 패키지 ok) | — |
| K1 | CAUGHT | `…ShowsADeadDescriptorAsUnreachable` · **`…BootKeepsAnUnreachableEndpointUnreachable`**(결정적) |
| K2 | CAUGHT | 1판 + `…BootKeepsAnUnreachableEndpointUnreachable` |
| K3–K11 · K14 · K16 · K18 | CAUGHT(1판과 같은 시험) | — |
| K12 | **CAUGHT** | `TestTheConsoleStrategyPumpReturnsWhenTheConsoleEnds` |
| K13 | **CAUGHT** | `TestTheAbsenceIsJudgedInOnePlace` |
| K17 | BROKEN(의도) | 생산 빌드 컴파일 거절 |

2판: **17/17 포획**(CAUGHT 16 · 컴파일 포획 1), 누수 0.

## not-applicable (침묵한 생략 아님)

- **요청 경로 dial·single-flight·rate limit·전이 1회 로그·취소·늦은 실패·밀려난 client Close** — wrapper
  (`httpapi_strategy_attach.go`)는 a115 가 편집하지 않고 재사용한다. 그 변이는 a109 T2 원장(M8–M38)이 이미 잰다.
- **펌프의 절반 틱** — 지터 의존이라 결정적 시험이 없다(a114 원장과 같은 한계, 명기).
- **wrapper 전이 로그 「데몬」 문구** — 표면 밖(issues R1).
