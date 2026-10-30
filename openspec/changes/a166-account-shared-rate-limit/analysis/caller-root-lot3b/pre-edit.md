# root 소유·engine/flatten 배선 사전 선언

기준 `369bf5b36`(source는 `4c92443a9`과 동일), 승인된 `analysis/caller-lifetime-plan.md`. 기존 함수 사전 AST JSON 동봉. 이번 로트는 root/engine/flatten 및 app의 official 생성 경계만, 다른 CLI/tool·priority·운영·a087 활성화는 후속이다.

| 기존 함수 | Function Logic Map / 보존 불변식 | Branch Test Map RED |
|---|---|---|
| main | 언어 설정→root→redacted stderr→Execute 오류 출력/exit1. | execute helper가 성공·오류 모두 close, 오류+close Join; monitor 출력 추가 오류 줄 없이 exit1, hardexit130 기존 경로 유지. |
| newRootCmd | rootOptions 소유, pre/post 및 모든 하위 명령에 전달. | 기존 함수/호출 테스트 호환, env snapshot 2개, OFF 경로 I/O 없음. |
| newMonitorCmd | WTS session·probe 실패 print 후 os.Exit(1). | sentinel 반환과 helper close, monitor 자체는 공식 budget 범위 밖. |
| newAppContext/app.New | 기존 cfg·credentials·prefer 분기에서만 official client 생성. | 명시 WTS/비활성은 factory 호출 0; required official factory 오류는 WTS fallback 0; nil borrowed factory OFF 기존 경로. |
| assembleEngine/NewOrderPath | config/credentials 확인 후 official.New, NewContext가 Accounts 수행·journal 소유. | required 먼저 Open→첫 token/accounts guard; nil factory OFF·기존 계좌 선택 유지, 오류 전파. |
| buildFlattenWiring | OrderPath→Accounts→journal→gateway; journal close. | 같은 root factory 전달, 첫 Accounts guard, 오류 반환, factory는 상위 owner만 Close. |

새 root factory leaf는 처음 생성이며 기존 함수 FLM N/A. 두 변수 env 표, 무명 mode/빈값/불일치 RED; 성공만 캐시·실패 후 재시도·생성/Close 동시성, closing 뒤 재생성 차단·active lease Close 오류 재시도 RED. 연구 source: stdlib mutex/errors.Join, rootOptions 기존 구조, accountbudgetruntime factory `4c92443a9` 및 Manager 계약. Go 1.25/1.26과 x/sys 기록은 기존 manager-review-v4 참조, 추가 의존성 없음. UI 없음 impeccable N/A. 실제 테스트/운영 PASS 아님.

## 분기·시험 연결 및 검토 결과

사전 AST의 main B1, monitor B1–B4, root newRootCmd B1–B7/newAppContext B1, assembleEngine B1–B2, flatten B1–B8, app.New B1–B8, engine.NewOrderPath B1–B6은 각 `.pre.json`에 고정했다. B 번호는 추출 좌표이며 함수 의미·새 분기와 test 연결은 아래 손 읽기다.

- root owner: `TestA166RootBudgetEnvMatrix`, `RootEnvSnapshot`, `RootFactoryConcurrentReuseAndClose`, `RootFailureCanRetry`, `RootWaitHonorsContext`, `RootCachedFactoryRejectsCanceledContext`; 최초 compile RED와 normal/race/tagged 원문 동봉.
- main/monitor: `TestA166ExecuteClosesOnSuccessAndFailure`, `ExecutePreservesCommandAndCloseErrors`, `MonitorFailureOutputAndNoBudgetOpen`, `MonitorMainExitOne`. monitor 실제 `main()`을 child에서 fake WTS runner로 호출해 exit1 및 출력 검증; 성공 경로·close error는 local fixture. 두 번째 신호의 hard exit130은 source 무변경이며 실행하지 않았다.
- app: `TestA166AppBudgetProviderFollowsRouting`의 WTS/disabled/official 세 갈래와 `AppNilFactoryIsError`, 기존 `TestNewOfficialClientTruthTable`. provider는 app이 기존 config/credential/prefer로 official 생성 분기임을 확인한 뒤에만 호출한다. 기존 명시 WTS와 monitor는 factory를 열지 않는다. 기존 context 없는 app.New는 직접 nil provider OFF를 보존하며 root 연결 시 Cobra command context를 공급한다.
- engine/flatten: `TestA166OrderPathUsesBorrowedFactoryBeforeBootstrap`, `OrderPathFixtureOverrideHasNoAuthority`, `EngineAndFlattenRequiredMissingFailBeforeBootstrap`, `EngineDecoratorCannotRemoveRequiredBudget`, `FlattenBootstrapUsesBorrowedBudget`. local TLS fake와 hold lease로 bootstrap wire 0 검증. flatten의 private rootOptions `officialOptions`는 테스트에서만 설정하며 CLI flag/env로 노출하지 않는다. fixture override는 금전 authority가 없다. `TestA087ProductionBoardAdmissionFlowIsOnlyLiteralNil`과 Context raw mutator 봉인 선택 시험 통과.
- 변이: env invalid→OFF, closing 무시, app provider 우회, flatten budget 제거, engine decorator 뒤 budget 재결속 삭제 각각 명시된 RED 시험에서 실패했다. fake local TLS 외 wire가 나가지 않도록 fixture를 제한했다.

검증 원문은 `a166-root-*.txt`로 동봉한다. monitor 첫 fake 시험은 Cobra writer의 실제 라우팅과 기대가 달라 실패했고, 실제 출력 두 스트림 합과 기존 문구를 검사하도록 교정했다. app provider 첫 시험은 테스트에서 잘못 쓴 `config.Default`로 컴파일 실패해 `DefaultFile`로 교정했다. 두 실패 원문도 보존한다. 테스트 결과는 변경 부분에 한정되며 live broker·전체 gate는 실행하지 않았다.

수동 gstack 범위 검토: env trust/실패, 닫힘·race·cleanup 전파, 공식 WTS 분기, optional provider null, test-only override authority, 계좌 선택/출력, raw mutator 봉인 및 hard exit 예외를 확인했다. mixed-model 전체 리뷰 및 독립 Manager 검증은 후속이다. root/app/engine/flatten 일부 연결만 완료 후보이며 다른 CLI/tool·console lifecycle·priority/replay·429·실제 참여자 범위는 여전히 미완이다. 이 부분은 required 활성화 또는 운영 승인이 아니다.

추가 설계 결정: app.New에는 인자 없는 provider 하나만 추가하고 root closure가 Cobra command context를 사용한다. 직접 required builder에서 context가 없으면 오류로 거절하며 Background 잠금 대기를 새로 추가하지 않는다. Manager 승인에 따라 적용했다. `rootOptions.officialOptions`는 flatten의 local TLS fixture용 비공개 seam이며 기본 nil, CLI flag/env로 노출하지 않는다.

최종 provider는 Manager 승인대로 `func() (*Factory,error)` 하나다. root closure가 실행 중 Cobra의 `cmd.Context()`를 캡처하고 직접 required builder에 context가 없으면 오류다. 취소된 required startup의 local TLS fake wire 0과 context 없는 직접 builder의 fail-closed를 추가했다. Manager의 caller-lifetime-plan 단일 문단만 새 signature와 동기화했다. Go 1.25 Linux tagged race 선택 시험도 통과했으나 Alpine/운영 실행이 아니다. hard exit130 및 커널 kill의 실제 child crash 시험은 미실행이다.
