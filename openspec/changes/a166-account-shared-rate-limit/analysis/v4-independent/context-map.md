# a166 v4 context map — read-only 2026-10-31

기준 main e32606d4bd44de3142f3c732ab86ecee913e3ab0. a087 checkout 1a153af2c과 tracked Go 차이는 a166 transport-prototype lifecycle_test.go뿐(`git diff --name-only 1a153af2c e32606d4b -- '*.go'`). 원본 경로는 저장소 상대경로, live/API/secret 호출 없음.

## 수동 원본 대조

- `internal/official/client.go:99-118,191-205,320-354`: New는 기본 transport 만들고 option 적용 뒤 tokenManager에 같은 hc/base를 주입; doRequest는 c.hc.Do 및 body read, send는 최초 및 401 재전송별 doRequest. `WithBaseURL`·`WithHTTPClient`는 :56-78, seal 이후 거부.
- `internal/official/token.go:60-78,108-125,130-150`: token/refresh→exchange의 hc.Do는 doRequest 외부. mutex와 ctx propagation 존재.
- `internal/official/a112_mbus_read.go:116-124,147-153`: sealed same-client 검증 뒤 원 transport를 clone, 독립 http.Client.Do. `tools/a112-mb-us-source/official_reader.go:16`의 별도 token cache.
- `internal/app/engine/gateway.go:320-325`→`internal/execgw/replay_transport.go:48-59,79-95`: replay는 별도 http.Client 및 official base URL. 안전 계약 결정 시 취급 명시 필요.
- `tools/a087-market-sell-probe/send.go:38-40,53-85`: official AuthHeaders를 쓰되 별도 client로 orderPath 전송.
- `internal/monitor/probes.go:57-58,161-183`: 별도 WTS host·쿠키, 공식 API와 계좌 공유 여부 미확인. `internal/client/*.go`도 별도 WTS 계열을 조사 대상으로 남김. `tools/a128-jev-shadow-probe/judge.go:177-187`는 별도 judge POST, 브로커와 동일하다고 가정 금지.
- constructor callers 전체 rg: `cmd/tossctl/openapi.go:44`, `instrument_names.go:125`, `mcp.go:68`, `verify_reconcile.go:183,204`, `candidatepanel.go:67`, `engine_reconcile.go:254`, `verify.go:889,906`, `soak.go:727`; `internal/app/app.go:132`, `internal/app/engine/engine.go:393` (NewContext :441), `cmd/tossctl/flatten.go:200` (NewOrderPath). tools `a128-jev-shadow-probe/main.go:127`, `a112-l1c-quote-probe/main.go:60`, `a130-outcome-backfill/main.go:89`, `a112-mb-us-source/official_reader.go:16`, `a087-market-sell-probe/main.go:95`, `a140-breakout-signal-shadow/main.go:133`, `a128-d2-receipts/main.go:99`. console의 client 소유·실제 실행 경로 별도 완전 추적 미완료.
- `internal/app/app.go:121-132` 및 `internal/app/paths/paths.go:48`은 configDir별 자격·토큰 경로를 고를 수 있다. 다른 config/process/host에서도 동일 계좌를 공유할 수 있다는 문제와 host 간 공통 filesystem·안정 identity 존재는 원본만으로 보장되지 않는다.

## AST 및 CodeGraph

main의 a166 기존 AST 23개 파일 SHA 대조 `ast-freshness.tsv`: 21 fresh, 2 stale (NewOrderPath, newRetrier; 파일 전체 hash 기준). 재추출 `engine--NewOrderPath.json`, `engine--newRetrier.json`; 각 AST의 branch kind/call text/return kind 목록은 이전과 동일하나 NewOrderPath 좌표 차이. 동등성 주장은 본문 전체 의미 동등성 증명이 아님.

main checkout `codegraph status`는 `Index is up to date`, commit SHA는 출력하지 않음. `codegraph-callers-*`는 해당 index의 검색 결과(보조 교차 확인)이며 완전성·HEAD 바인딩은 미검증. a087 및 a169 worktree는 CodeGraph 미초기화. 원본 소스와 AST 우선. index 재구축·sync 미실행.

## 미검증

다른 프로세스·다른 config·다른 host의 동일 계좌 안정 식별 및 상호 배제, WTS와 공식 API budget 동질성, console 경로 완전 열거, direct request 별 구동 조건, 메모리 색인 commit binding, a166 v4 특정 계약 승인·동작, 전체 게이트·테스트·운영 상태. 실제 통신·계좌·환경 비밀 확인 안 함.
