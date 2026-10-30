# a166 replay/probe 사전 분석 초안

기준 `d13008db1205477c61731f98a836c887a587bcc8`; core는 main 고정, official lot2a 별도 branch. 이번 소유: `internal/execgw/replay_transport.go`·관련 고유 test, `tools/a087-market-sell-probe` 전송/생성 경계·관련 test, 이 분석 디렉터리. 제품 변경은 Manager의 주입·동일성 판정 전 착수하지 않는다. UI 없음: impeccable N/A — backend/tool-only. 기존 생산 caller·운영 env·설정·실계좌·live probe는 변경/실행하지 않는다.

## 현재 근거

`ast-*.json`은 현 기준 소스 SHA를 포함한 `tools/logic-map/extract_go_ast.go`의 사전 Go AST 산출물. 아래 분기 번호는 AST로 열거한 좌표이고 해석은 현재 원문을 손으로 읽은 것이다. CodeGraph는 새 worktree에서 not initialized, 호출 그래프 freshness는 미검증. `rg` 호출자 검색에서 production `internal/app/engine/gateway.go:323`과 `internal/execgw/replay_test.go:85`가 HTTPReplay를 구성한다. 실제 production Replay의 Attested는 nil이며 변경 대상이 아니다.

- `HTTPReplay.ReplayPlace`: 분기 69/72/75 입력 거절, 81 생성 실패, 85/87 헤더·오류, 90 헤더 범위, 96 Do 오류, 102 본문 오류. 요청에 journal에서 건너온 body 바이트와 기존 headers를 사용하며 status/응답 해석 계약을 보존해야 한다.
- `Gateway.ReplayInDoubt`: 사전 가드·journal 카운터 이후 ReplayPlace 호출은 AST 280 for, 309 switch 및 원문 319~324. 전송 전 admission 실패를 이유로 이전 원장의 불명확 결과를 확정하거나 replay 카운터를 임의 환급하지 않는다. 이번 소유 밖이므로 기존 함수 수정은 금지.
- `execute`: terminal·confirm·receipt 경계는 AST 186~229, 단일 send와 최종 receipt 분류는 233~246. `openOfficial`: 85/89/92 credential 실패. `sendOnce`: 87 Do 실패, 95 본문 실패. `newSendClient`는 redirect 제한과 keepalive 비활성화. `classify`는 AST 145~159의 보수 분류. token header 준비는 현재 receipt 생성 전이며 직접 POST는 그 뒤임.
- `sendResult.Sent`는 현재 실제 wire 관측이 아니라 client.Do 진입 의미다. required 입구에서 no-send를 보장하더라도 이미 존재하는 pending receipt와 prior AMBIGUOUS를 안전하게 취급하며, 관측 없이 실제 전송 여부를 새로 주장하지 않는다.

## 초안 BTM/RED

1. replay OFF: body·idempotency·authorization·account headers·timeout/status/error 바이트 기존 값 유지. Headers 오류/빈 body/nil HTTP/base 거절, 응답 read 오류, 401/409/5xx 및 journal `IN_DOUBT` 유지.
2. replay enforcing: 지정된 동일 Budget + 인증 client 결속만 허용. required nil/unsupported/mismatch에서 header token과 POST 모두 wire 0, local TLS fixture로 통제. caller context priority와 취소 보존; transport `ErrNotSent`는 현재 HTTP 시도에만 적용하고 기존 journal 판단을 덮지 않음.
3. probe OFF: 단일 POST, 기존 non-redirect·header·body·receipt classification 유지. enforcing: token과 POST 양쪽이 같은 Budget/critical context를 사용, mismatch/nil/unsupported일 때 token/POST 0, 이미 시작한 body 실패는 unknown 유지. 실제 명령/실계좌 미실행.
4. cross-process/actual TLS와 변이: detached token 또는 direct POST 예산 우회, required nil fallback, body 재직렬화·추가 POST, prior unknown을 NOT_SENT로 오인하는 변이를 적합한 최소 fixture로 포착.

## 결속 미결

기존 replay `Headers func(context.Context)`와 probe `headerSource`는 어떤 official client/budget에 결속했는지 알 수 없다. 임의 bool interface는 구현자가 위조할 수 있으므로 required 모드에서 이를 동일 instance 증명으로 받아들이면 안 된다. concrete `*official.Client`에 대해 configMu 아래 sealed 상태, budget 포인터, `c.hc.Transport` private wrapper, `tm.hc == c.hc`를 확인하는 official 소유 검증 API 또는 동등한 비위조 결속이 Manager 판정/lot2a 소유자 작업으로 필요하다. 결속 확인 없이 token 먼저 호출하는 설계는 wire 0을 보장하지 못한다. 현재 그 계약 및 lot2a merge를 기다린다.

## 기준 focused 시험

현재 기준 `d13008db1`에서 `GOFLAGS=-buildvcs=false go test -race -count=1 -run '^(TestReplaySendsExactlyTheStoredBody|TestReplayIsOffWithoutAttestation|TestReplayWithNoAnswerCountsAndFallsBack|TestReplay422KeyConflictParksAndNeverFails)$' ./internal/execgw`: PASS 13.885s (`baseline-replay.log`). `GOFLAGS=-buildvcs=false go test -race -count=1 -run '^(TestExecuteSendsOnceAndRecordsAcceptance|TestOnlySendOnceCallsDoAndItHasNoLoop|TestSendClientRefusesRedirectsAndConnectionReuse|TestUnwritableReceiptDirSendsNothing)$' ./tools/a087-market-sell-probe`: PASS 1.034s (`baseline-probe.log`). 기존 기능 baseline이며 새 enforcing 기능의 증거가 아니다.
