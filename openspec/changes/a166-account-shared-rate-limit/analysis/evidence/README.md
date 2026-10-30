# a166 evidence — design 1판·2판 AST (2026-10-29)

- 기준: `origin/main` = `5e2facaee`(a165 착지 뒤). 도구: `go build -trimpath -o /tmp/a166-logicmap ./tools/logic-map` 후
  `/tmp/a166-logicmap --file <f> --func <F> > ast/<pkg>--<F>.json`.
- 이것은 **design 이 분기를 근거로 쓰기 전에** 뽑은 AST 다(FLM-first). gate ⑤ 번들(`analysis/function-logic/…`)은 아니다 —
  번들은 base 고정(task 0.1) 뒤 task 0.5 가 `scaffold_analysis.py` 로 만든다. 그때 `source_sha256` 이 아래와 다르면 design 의 분기 인용을 다시 대조한다.
- 손으로 읽은 줄은 design 에 「손 읽기」로 따로 적었다.

| 파일 | 함수 | 줄 | 분기 수 | source_sha256(앞 16) | AST |
|---|---|---|---|---|---|
| `internal/official/client.go` | `Client.doRequest` | 191–207 | 2 | `6d145916a6797f3a` | `ast/official--Client.doRequest.json` |
| `internal/official/client.go` | `Client.send` | 320–366 | 9 | `6d145916a6797f3a` | `ast/official--Client.send.json` |
| `internal/official/token.go` | `tokenManager.token` | 60–79 | 2 | `4cadb04ff6d6691a` | `ast/official--tokenManager.token.json` |
| `internal/official/token.go` | `tokenManager.refresh` | 108–126 | 1 | `4cadb04ff6d6691a` | `ast/official--tokenManager.refresh.json` |
| `internal/official/token.go` | `tokenManager.exchange` | 130–174 | 6 | `4cadb04ff6d6691a` | `ast/official--tokenManager.exchange.json` |
| `internal/official/a112_mbus_read.go` | `a112MBUSReadAt` | 103–192 | 17 | `386c61414ce45352` | `ast/official--a112MBUSReadAt.json` |
| `internal/official/errors.go` | `ShouldFallback` | 66–76 | 1 | `5f9cc4baf967a604` | `ast/official--ShouldFallback.json` |
| `internal/execgw/retry.go` | `ClassifyQueryError` | 69–98 | 13 | `1036f9ee83942077` | `ast/execgw--ClassifyQueryError.json` |
| `internal/execgw/retry.go` | `Retrier.Query` | 335–389 | 15 | `1036f9ee83942077` | `ast/execgw--Retrier.Query.json` |
| `internal/execgw/retry.go` | `Retrier.rateLimitWait` | 430–440 | 3 | `1036f9ee83942077` | `ast/execgw--Retrier.rateLimitWait.json` |
| `internal/execgw/classify.go` | `classifyMutation` | 21–99 | 10 | `020dba811b666e94` | `ast/execgw--classifyMutation.json` |
| `internal/execgw/classify.go` | `statusOf` | 125–143 | 6 | `020dba811b666e94` | `ast/execgw--statusOf.json` |
| `internal/execgw/replay_transport.go` | `HTTPReplay.ReplayPlace` | 68–110 | 9 | `4a956d5a59a19c07` | `ast/execgw--HTTPReplay.ReplayPlace.json` |
| `internal/app/engine/exitwiring.go` | `newRetrier` | 45–56 | 0 | `6020a40b20302b0d` | `ast/engine--newRetrier.json` |
| `tools/a087-market-sell-probe/send.go` | `sendOnce` | 83–99 | 2 | `ed2cbc3e339eaffa` | `ast/a087probe--sendOnce.json` |

2판에서 더한 것(같은 HEAD `5e2facaee` 코드 트리 — 2판 커밋 전 worktree 는 문서 커밋만 앞섬):

| 파일 | 함수 | 줄 | 분기 수 | source_sha256(앞 16) | AST |
|---|---|---|---|---|---|
| `internal/official/client.go` | `New` | 99–119 | 1 | `6d145916a6797f3a` | `ast/official--New.json` |
| `internal/official/client.go` | `newOfficialHTTPClient` | 121–132 | 0 | `6d145916a6797f3a` | `ast/official--newOfficialHTTPClient.json` |
| `internal/official/client.go` | `Client.authorityOriginLocked` | 157–161 | 0 | `6d145916a6797f3a` | `ast/official--Client.authorityOriginLocked.json` |
| `internal/official/client.go` | `Client.hcTransport` | 163–169 | 1 | `6d145916a6797f3a` | `ast/official--Client.hcTransport.json` |
| `tools/a087-market-sell-probe/send.go` | `newSendClient` | 54–65 | 0 | `ed2cbc3e339eaffa` | `ast/a087probe--newSendClient.json` |
| `internal/app/engine/gateway.go` | `buildGateway` | 245–371 | 6 | `c1b7c08f60614957` | `ast/engine--buildGateway.json` |
| `internal/app/engine/engine.go` | `NewOrderPath` | 359–406 | 6 | `60e3a8d45487301e` | `ast/engine--NewOrderPath.json` |
| `internal/app/engine/precheck.go` | `officialBroker.availableActions` | 98–128 | 1 | `1ef569176914063c` | `ast/engine--officialBroker.availableActions.json` |
