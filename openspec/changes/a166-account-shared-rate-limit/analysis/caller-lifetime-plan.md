# 호출자 배선과 factory 수명 — Manager 세부 계약

2026-10-31. 4판의 strict 3/s·보호 대기 정책과 공통 그룹 계약을 구현 로트로 나눈다. 운영 설정·a087 활성화는 변경하지 않는다. 근거는 현재 `cmd/tossctl/main.go`, `root.go`, `engine_assembly.go`, `flatten.go`, `internal/app/app.go`, `internal/app/engine/engine.go` 원문과 `/tmp/a166-lot3b-root-lifetime.md`다. 이 기록은 손으로 읽은 설계 근거이며 후속 편집의 AST/FLM/BTM을 대신하지 않는다.

## 실행 한 번의 소유자

app.New는 기존 config의 Enabled/prefer 판정 뒤에 공식 client 생성 여부를 정한다. 이 판정을 root에서 중복하지 않는다. app.Options는 해당 분기에서만 호출하는 lazy provider `func() (*accountbudgetruntime.Factory, error)` 하나를 받는다. 기존 New(opts)에 context 인자가 없으므로 root closure가 실제 명령 context를 capture하며 별도의 BudgetContext 옵션을 추가하지 않는다. required 초기화의 context가 없으면 오류로 처리하고 Background 잠금 대기를 추가하지 않는다. nil provider는 기존 OFF 호환이며, 설정된 provider가 nil factory/nil error를 반환하면 오류다. provider 오류를 optional client 없음이나 WTS fallback으로 바꾸지 않는다. 명시 WTS·monitor에서 공식 client를 만들지 않는 경우 factory도 열지 않는다. engine.Options는 borrowed concrete factory를 직접 받는다. 실제 env snapshot은 실행에 고정하되 그 검증 오류는 공식 client가 필요한 지점에서 드러나게 할 수 있다.

CLI의 rootOptions는 factory 하나를 지연 생성해 보유한다. 모든 공식 client builder는 이 객체를 빌려 사용한다. app/engine Options에 주입된 factory는 borrowed이며 Context.Close가 닫지 않는다. 같은 프로세스에서 새 profile/계좌 선택으로 client를 다시 만들더라도 factory를 새로 만들지 않는다. 별도 프로세스는 같은 trusted manifest의 파일 ID를 독립 검증한다.

실행 helper는 root.Execute가 성공하거나 오류로 반환한 뒤 factory Close를 호출하고 그 결과를 반환한다. main의 os.Exit 이전에 종료 처리가 끝나야 한다. Cobra PersistentPostRun은 오류 시 호출되지 않을 수 있으므로 유일한 종료 경계로 쓰지 않는다. 기존 newRootCmd의 이름과 반환 타입은 보존하고, 명시 rootOptions를 받는 생성 함수로 실행 helper와 모든 하위 명령이 같은 소유자를 사용한다. Close 오류가 없으면 기존 실행 오류 객체/출력 의미를 유지하고, Close도 실패하면 errors.Join으로 두 오류를 보존한다. 기존 redaction과 종료 코드 정책은 유지한다.

직접 종료 두 경로를 구별한다. `monitor.go:38`의 일반 실패 os.Exit(1)은 helper로 반환하는 제한된 종료 신호로 바꿔 Close를 거치게 한다. 기존 결과 출력과 exit1은 유지하고 sentinel 자체의 추가 Error 줄을 출력하지 않는다. Close 오류가 함께 있으면 이를 숨기지 않는다. main에서 처리할 종료 코드/오류 반환을 최소 형태로 구현·시험한다. `engine.go:113`의 두 번째 신호 hard exit130은 의도된 강제 종료이므로 변경하지 않으며 정상 Close 보장에서 제외한다. SIGKILL·hard exit에서 커널 flock 해제, 미완 queue/dirty의 보수 처리와 재시작 대기를 child fixture로 확인한다. 이를 정상 drain 성공으로 보고하지 않는다.

factory의 성공 인스턴스만 캐시한다. 초기화 실패는 OFF/client nil로 바꾸지 않고 해당 호출에 반환하며, 후속 명시 호출은 같은 설정으로 재시도할 수 있다. 종료 시 root는 새 client 생성을 먼저 차단한다. factory.Close가 활성 owner 때문에 실패해도 이 root의 새 생성은 열리지 않는다. 실패 오류와 factory handle은 보존하며, 이미 시작한 작업의 정리 뒤 Close 재시도는 가능하다. 프로세스 종료 시 Close 실패를 정상 종료로 숨기지 않는다.

환경은 실행에 고정한 LookupEnv로 `TOSSOS_ACCOUNT_BUDGET_MODE`와 `TOSSOS_ACCOUNT_BUDGET_MANIFEST` 두 항목을 스냅샷한다. 둘 다 부재면 기존 OFF, mode=`off`는 manifest 부재에서만 OFF, mode=`required`는 비어 있지 않은 manifest가 있어야 한다. mode가 부재인데 manifest가 있거나, mode 빈 값/알 수 없는 값, off와 manifest 동시 지정, required의 manifest 누락/빈 값은 bootstrap 전 오류다. 단일 변수 FromEnv의 부재→OFF만으로 이 판정을 대체하지 않는다. profile 경로가 달라도 같은 manifest를 사용한다. 실제 환경 값을 바꾸지 않으며 운영 launch의 참여자·required 설정 확인은 별도다. 필요한 cleanup callback은 비밀 없는 오류를 기존 stderr/redaction 경로로 드러낸다. 로컬 파일 경로·opaque group을 계좌 인증이나 권한 증거로 사용하지 않는다.

factory 초기화·Close의 동시성은 root 소유자가 직렬화한다. 생성 중 context가 취소되면 token/accounts 이전에 실패한다. 기존 context가 없는 builder는 기존 호출자 context 계약을 보존하되 Background를 새 장기 요청의 근거로 추가하지 않는다. 실행 helper를 우회하는 required 테스트는 생성한 rootOptions의 Close를 직접 소유해야 한다. 기존 OFF 명령 트리 조회 테스트는 새 예산 파일을 만들거나 읽지 않는다.

## 분리 로트

1. root 소유권과 app/engine borrowed factory 입력, engine/flatten 조립 연결. 모든 실패 경로·첫 계좌 조회 이전 guard·OFF golden·Close 결과를 집중 시험한다. 다른 CLI/tool이 아직 연결되지 않았음을 명시한다.
2. 나머지 CLI client builder, verify의 재생성, console lazy broker/metadata, MCP, soak를 같은 owner에 연결한다. console의 실패 후 명시 재시도와 verify의 호출별 새 client 계약은 유지하며 client를 전역 공유하지 않는다. 공유 대상은 factory/Budget뿐이다. console의 background 작업과 반환 후 closure 수명은 실제 소스·fixture로 확인한다. 종료 후 작업이 남는다면 Close 전에 취소·join 경계를 마련한다. 기존 WTS 경로는 예산 실패 fallback으로 쓰지 않는다.
3. 독립 tools는 각 main/run이 factory를 소유하고 모든 반환 경로에서 닫는다. method closure를 반환하는 helper도 최종 사용까지 owner가 살아 있어야 한다. 도구별 required/auxiliary 적용과 기존 계좌 선택을 기록한다.
4. 보호 판정 시세·isProtective 주문·보호 취소·flatten·ACK/token 사슬에 명시 context priority를 연결한다. 일반 SELL 또는 모든 GET을 critical로 승격하지 않는다. replay의 원장 출처가 불명인 경우 처분은 source 근거를 확인한 뒤 별도로 고정한다. a087 production admission nil은 유지한다.

## 수용 증거와 범위

required manifest 오류→첫 bootstrap wire0, 같은 owner의 복수 client 결속, 실행 성공/실패/Close 실패, 닫힘 뒤 생성 거절, 취소·동시 생성과 console 수명을 확인한다. 기존 계좌 선택·권한·journal·a165 최초 ambiguous를 보존한다. 현재 등록된 caller 전부를 연결했다는 주장은 최종 구조 검사 전 금지한다. 실제 official client 및 두 프로세스의 TLS 첫 plaintext Write 시작을 직접 기록하는 test-only fixture를 후속 추가한다. 서버 도착 시각으로 대신하지 않는다.

search-first: 기존 rootOptions·main 실행 경계·engine/app Options·stdlib errors.Join과 mutex를 확장한다. 신규 의존성·registry·hot reload는 필요 없다. grill-me/grilling: 확정된 사용자 정책을 재질문하지 않고 source에서 수명 경계를 확인했다. impeccable: N/A — backend/documentation-only. 대상 함수 편집 전 Pre-Edit·FLM/BTM·RED, 이후 영향 시험·변이·수동 gstack·독립 리뷰가 필요하다. 전체 시험/gate·실계좌·운영 배포 승인이 아니다.

## 독립 설계 검토 처분

Terra의 `/tmp/a166-caller-lifetime-independent.md`에서 일반 monitor 직접 exit와 hard-exit 예외, 두 변수 판정표 누락을 지적했다. Manager가 해당 원문을 직접 확인하고 위 계약에 반영했다. Close 실패 후 root의 새 생성 금지와 console의 호출별 client 재생성도 명시했다. 수정 후 독립 재검토는 첫 배선 로트 착수를 조건부 승인했다. Manager는 1번 로트의 개발 착수를 승인하며 console/나머지 호출자·운영·전체 완료는 포함하지 않는다. 구현 수용은 위 fixture·FLM/BTM·독립 검토를 통과한 별도 SHA로 판정한다.
