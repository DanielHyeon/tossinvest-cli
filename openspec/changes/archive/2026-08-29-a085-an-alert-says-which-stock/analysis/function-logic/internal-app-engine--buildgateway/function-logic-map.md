# Function Logic Map: `buildGateway`

- Source: `internal/app/engine/gateway.go` (lines 199–312)
- AST evidence: `ast.json` (`source_sha256: de44aae53a1941837cb75b4c3686801163d776701f0779e5bd18b8bf8b4f3c8b`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal** — 알림 문구와 표시 전용 배선. 주문·손절·사이징·원장 판정 경로를 바꾸지 않는다.

## What it does

조립 배선. a085는 알림 표면이 공유하는 InstrumentNames registry 하나를 만들어 넘긴다. 다른 배선은 그대로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `Names` registry | nil 허용 | 대사 루프가 holdings 스냅샷에서 채운다 | nil이면 코드만 렌더 — a085 이전 동작 |
| 브로커 `name` | 빈 문자열 허용 | `GET /api/v1/holdings` 기존 필드 | 비면 코드만 렌더. 추정하지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (204) `if` — if err := checkProjectionWired(in.journal); err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B2 | (226) `if` — if err := tracker.Restore(ctx); err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B3 | (250) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B4 | (274) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `checkProjectionWired` | (204) if err := checkProjectionWired(in.journal); err != nil { | 호출부 계약 유지 | AST `calls` |
| `execgw.NewEntryGate` | (214) entry := execgw.NewEntryGate(in.clock, nil) | 호출부 계약 유지 | AST `calls` |
| `tracker.Restore` | (226) if err := tracker.Restore(ctx); err != nil { | 호출부 계약 유지 | AST `calls` |
| `fmt.Errorf` | (227) return engineWiring{}, fmt.Errorf("engine: restoring the RECONCILE projection: %w", err) | 호출부 계약 유지 | AST `calls` |
| `entry.SetAuthorityRefresh` | (229) entry.SetAuthorityRefresh(func() error { | 호출부 계약 유지 | AST `calls` |
| `tracker.Refresh` | (230) return tracker.Refresh(context.Background()) | 호출부 계약 유지 | AST `calls` |
| `context.Background` | (230) return tracker.Refresh(context.Background()) | 호출부 계약 유지 | AST `calls` |
| `productionProtectionDigests` | (242) buildDigest, toolDigest := productionProtectionDigests() | 호출부 계약 유지 | AST `calls` |
| `productionProtectionAssemblies` | (243) assemblies := productionProtectionAssemblies(buildDigest) | 호출부 계약 유지 | AST `calls` |
| `protectionreadiness.NewProductionProvider` | (244) provider := protectionreadiness.NewProductionProvider(protectionreadiness.ProductionConfig{ | 호출부 계약 유지 | AST `calls` |
| `protection.NewPairedReadinessAdapter` | (249) readiness, err := protection.NewPairedReadinessAdapter(provider, in.accountRef, "production", provider.RuntimeContracts()) | 호출부 계약 유지 | AST `calls` |
| `provider.RuntimeContracts` | (249) readiness, err := protection.NewPairedReadinessAdapter(provider, in.accountRef, "production", provider.RuntimeContracts()) | 호출부 계약 유지 | AST `calls` |
| `execgw.New` | (253) gateway, err := execgw.New(execgw.Options{ | 호출부 계약 유지 | AST `calls` |
| `in.official.BaseURL` | (265) BaseURL: in.official.BaseURL(), | 호출부 계약 유지 | AST `calls` |
| `newNotifier` | (280) notifier := newNotifier(in.journal, entry, in.accountRef, in.logger, in.publisher, in.clock) | 호출부 계약 유지 | AST `calls` |
| `newRetrier` | (281) retrier := newRetrier(in.journal, entry, in.accountRef, notifier, nil, in.clock) | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 알림 제목·본문 문자열만. 이벤트 타입·Key·Fields·발송 조건·latch·원장은 바뀌지 않는다.

## Safety conclusion

- Safe edit boundary: 사람이 읽는 문자열과 표시 전용 필드. 기계 판독 표면(payload·로그 키·원장 cause)은 영문·원문 그대로다.
- High-risk impact: no — 판정도 주문도 이 경로에 없다. §0.4는 불변이다: 이름은 이미 대금을 지불한 응답에서 온다.
