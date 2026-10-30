# Function Logic Map: `Gateway.ReplayInDoubt(params=2, results=2)`

- 원문: `internal/execgw/replay.go`. pre-AST는 priority-origin-lot, post-AST는 ast.json.
- 입력·상태·호출·오류·불변식: terminal/attestation/state/kind/key/body/TTL 순서 후 required private 타입만 출처 판정. 미입증 QueryFallback/count 유지, DB 오류 반환. 시험 TestA166RequiredReplayUnknownOriginFallsBack, TestA166ReplayAdmissionRefusalPreservesInDoubtJournal, TestReplay409DoesNotConsumeTheCap.
- CodeGraph 미초기화; source 및 Go AST 근거. High-risk.

| 분기 | AST 위치와 조건 원문 | 결과 범위 |
|---|---|---|
| B1 | if L234: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B2 | if L239: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B3 | if L246: `if rec.State.IsTerminal() {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B4 | if L252: `if g.replay == nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B5 | if L260: `if g.attested == nil \|\| !g.attested(ctx) {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B6 | if L269: `if !g.claimSymbol(symbolKey) {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B7 | if L276: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B8 | for L281: `for {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B9 | if L286: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B10 | if L291: `if rejected := g.replayGuards(rec, cfg); rejected != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B11 | if L297: `if wait := replayWait(rec, cfg, g.clk.Now()); wait > 0 {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B12 | if L298: `if waits >= cfg.MaxWaits {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B13 | if L303: `if err := g.clk.Sleep(ctx, wait); err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B14 | if L310: `if _, required := g.replay.(*budgetedHTTPReplay); required {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B15 | if L312: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B16 | if L315: `if !proven {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B17 | switch L322: `switch {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B18 | case L323: `case errors.Is(err, journal.ErrReplayCapReached):` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B19 | case L327: `case errors.Is(err, journal.ErrReplayNotInDoubt):` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B20 | case L329: `case err != nil:` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B21 | switch L339: `switch verdict.kind {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B22 | case L340: `case replayRecovered:` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B23 | if L344: `if err := attempt.ResolveConfirmed(ctx, verdict.orderID,` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B24 | case L354: `case replayInProgress:` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B25 | if L359: `if rerr != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B26 | case L367: `case replayKeyConflict:` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B27 | if L370: `if err := attempt.ResolveUnresolved(ctx, journal.ReasonReplayKeyConflict, verdict.detail); err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B28 | case L379: `default:` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
