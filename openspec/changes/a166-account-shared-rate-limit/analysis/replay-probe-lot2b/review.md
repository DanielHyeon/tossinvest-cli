# a166 replay/probe lot 2b 구현·검증

기준: 제품 첫 로트 `d13008db1`에서 시작, 공식 lot2a main `2c5617321`과 concrete binding API main `f44143788`을 ff-only로 반영했다. 상세 사전 읽기/사전 AST와 계획은 `pre-edit.md`; `ast-*.json`은 `d13008db1`에서 생성하고 official 두 차례 ff 이후 대상 원문의 SHA가 여전히 같은 것을 대조했다. `post-*.json`은 현재 소스의 AST다. CodeGraph 새 worktree는 미초기화이며 graph hard evidence/freshness PASS를 주장하지 않는다. 기존 실행 함수의 분기 좌표는 AST로 열거하고 의미·호출 관계는 원문을 손으로 읽었다.

## 범위와 불변식

`HTTPReplay`의 기존 public OFF 경로·stored body·headers·status 해석은 수정하지 않았다. `NewBudgetedHTTPReplay`는 concrete `*official.Client`의 sealed exact-budget/token-client 결속을 확인하고 client 자신의 BaseURL과 정확히 일치한 URL만 허용한다. private wrapper가 토큰 획득 전후 결속을 재확인하며 원래 요청 context를 같은 budget을 쓰는 직접 POST에 전달한다. `http.Client`를 복사하고 enforcing에서만 redirect를 막는다. nil/unsupported/mismatch는 헤더 발급 전 거절한다. 오류는 기존 ReplayPlace의 `%w` 경계를 통해 보존되며 journal의 `ReplayInDoubt` 분기·기존 count/불명 결과를 변경하지 않았다.

probe의 required 경계는 현재 실행파일 내부 `deps`와 `newBudgetedSendClient`다. concrete official 결속·전송 template 확인 뒤 token과 직접 POST에 동일 Budget, Critical context를 사용한다. required에서는 기존 임의 opener를 부르지 않는다. main의 기존 구성은 OFF 그대로이며 CLI/env 옵션이나 운영 설정은 추가하지 않았다. lot3 runtime factory가 주입하기 전 이 경로는 실제 main에서 켜지지 않는다. 기존 `sendOnce`의 Sent는 client.Do 진입 의미이며 wire 입증이 아니다. admission 실패의 기존 receipt는 보수적으로 unknown 처리한다. local TLS fixture의 테스트 endpoint는 거래 권한을 얻지 않는다.

## 함수 흐름/branch test map

- `HTTPReplay.ReplayPlace`의 사전 AST 69/72/75 입력 거절, 81 request 생성 오류, 85/87/90 헤더 분기, 96 Do 오류, 102 본문 오류는 원문 그대로. OFF body/경계: `TestReplaySendsExactlyTheStoredBody`, `TestReplayIsOffWithoutAttestation`; lost response·409·422·다른 echoed key: 선택된 기존 Replay 테스트. 새 private wrapper의 nil/foreign/URL 거절·token 및 POST 공유·client 불변·redirect·held-budget은 `TestA166Replay*`로 검증했다.
- 유일한 기존 제품 함수 변경 `execute`: 사전 AST `ast-execute.json`, 변경 후 `post-execute.json`. terminal 193, confirm/body 198~205, 경로/receipt 존재 209~215, 기존 opener 오류 234, 헤더/필수값 240/246, request 및 receipt 오류 250/255, 최종 쓰기/accepted/refused/unknown 264~273은 기존 분기다. 필수 구성/결속 222~225, template 오류 226~228, Critical context 230~231, header 후 결속 243~244가 새 분기다. `TestA166Probe*`는 필수 nil·foreign·unsupported template의 token/POST/receipt 0, 동일 budget token→POST, held budget POST 0, 닫힌 budget의 unknown receipt를 검증한다. OFF 성공/재시도 금지/confirm·terminal·receipt 오류 및 정적 단일 POST·redirect·권한 API 금지는 선택된 기존 테스트로 검증한다. `filepath.Abs` 오류 및 header 직후 외부 동시 변조의 독립 강제 fixture는 미실행이며 sealed/private official 조건에 의존한다.
- 관련 기존 테스트 함수 `TestNoProductionOrderSurfaceIsNamed`, `TestImportsStayInsideTheAllowlist`는 `pre-*.json`/`post-*.json` 보존. 새 concrete `official.Client` 타입과 `accountbudget` import만 허용하고 생산 주문·계좌 API 금지는 유지했다. 변경 직후 기존 static RED `probe-static-red.log`, 최소 변경 후 GREEN `probe-static-green.log`.
- `Gateway.ReplayInDoubt`는 수정하지 않았다. 실제 journal fixture `TestA166ReplayAdmissionRefusalPreservesInDoubtJournal`에서 기존 IN_DOUBT·replay count 1을 유지하고 현재 replay POST 0을 확인한다. 이를 최초 주문의 wire 미전송 증거로 전용하지 않는다.

## 실행·변이

`final-{replay,probe}-go{125,126}-{normal,race}.log`는 각각 Go 1.25.0/1.26.5에서 대상 패키지에 한정한 `go test [-race] -count=1 -run ...` 원문이다. replay는 `TestA166Replay.*`와 위의 기존 OFF·409·422·unknown 테스트, probe는 `TestA166Probe.*`와 기존 실행·단일전송·redirect·static allowlist·receipt 테스트를 선택했다. 전체 패키지/저장소 시험은 아님. baseline `baseline-*.log`는 변경 전 d13008db1의 소수 시험이다. `*-overlay-candidate.log`/`*-overlay2.log`는 승인 전 official 후보 917a91023을 Go overlay한 잠정 시험이며 최종 main 검증으로 계산하지 않는다.

`mutation-*.log`는 공유 소스를 건드리지 않는 Go overlay 임시 변이 원문이다. replay·probe의 guard 우회, required 차단 제거, token-client 결속 해제, 다른 budget 허용, 저장 body 변형, redirect 허용, prior IN_DOUBT를 부당 확정하는 변이는 각각 관련 local fixture에서 RED. 모든 변이의 네트워크 상대는 `httptest` local server이며 live broker는 사용하지 않았다.

## gstack 및 남은 범위

로컬 gstack review checklist를 수동 적용했다. critical 관점에서 concrete 결속·동일 Budget·원장상 기존 불명 상태·single POST·body와 header 바이트 및 callback/URL 경계를 검토했고, 발견된 기존 static allowlist 실패는 최소 수정했다. 변경 없음/미해결 항목을 추가 자동 승인한 것은 아니다. telemetry·설치·업그레이드·다른 specialist 호출은 하지 않았다. 독립 Manager 검증은 별도다.

전체 저장소/전체 gate, 실제 계좌·브로커·운영·실행 probe·배포, a087 activation, a169 측정, 타 OS 실행, 전체 caller wiring 및 권한/동일 namespace·clock 운영 검증은 미실행. 기존 a166은 진행 중이며 push/merge/archive하지 않았다. 신규 `accountbudget` test import는 probe의 static allowlist에 명시적으로 추가했으나 기존 생산 주문 API 금지는 변경하지 않았다.

## Manager 후속: enforcing nil transport 거절

`6099998ed`의 budgeted constructor가 nil client transport를 `http.DefaultTransport`로 대체하던 경로는 명시 constructor-owned template 계약과 어긋났다. 사전 AST `pre2-NewBudgetedHTTPReplay.json`, 변경 후 AST `post2-NewBudgetedHTTPReplay.json`. 새 `TestA166ReplayRequiredNilTransportRejectedBeforeHeaders`는 로컬 TLS fixture의 token/POST 0을 확인하며 `nil-transport-red.log`에 이전 실패를 보존했다. 강제 모드에서 명시 `*http.Transport`가 없으면 `ErrUnavailable`을 반환하도록 수정했고 OFF `HTTPReplay`는 그대로다. `TestA166ReplayOffNilTransportStillWorks`는 OFF에서 nil/default가 기존대로 동작함을 local HTTP fixture로 확인한다. `nil-transport-green.log` GREEN, fallback 복원 Go overlay 변이 `nil-transport-mutation.log` RED. 명시적으로 전역 DefaultTransport를 수정하거나 임의 TLSNextProto를 허용하지 않았다.

`nil-go{125,126}-{normal,race}.log`는 해당 replay 패키지의 기존 OFF·journal·새 forcing 선택 시험이다. probe 코드는 이 후속 수정에서 변경하지 않아 기존 `final-probe-go{125,126}-{normal,race}.log`를 재사용한다. gstack checklist 수동 재검토: global fallback 삭제와 헤더 이전 거절 및 OFF 경계 확인. probe static guard의 concrete client selector 제한 범위는 독립 reviewer가 별도 지적했으며 이 작은 후속 수정으로 해결했다고 주장하지 않는다.
