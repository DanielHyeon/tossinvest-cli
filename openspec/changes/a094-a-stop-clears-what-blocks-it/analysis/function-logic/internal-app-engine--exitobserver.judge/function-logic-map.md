# Function Logic Map: `ExitObserver.judge`

- Source: `internal/app/engine/exitloop.go` (`872`–`911`)
- Qualified: `ExitObserver.judge`
- AST evidence: `ast.json` (`source_sha256` 9b7d9a800069b100…) — 구현 로트(2026-09-30) 편집 뒤. 편집 전 AST 는 `analysis/implementation/pre-edit/`
- Risk scan: `risk-pattern-report.md`
- 분기 8

**역할.** 포지션 하나를 판정한다. a094 는 진입에 `noteHeldProposal`(park 원인 · 종결 증거 대기 명명 critical) 한 호출을 더했다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `m` | 관리 포지션 | working set | identityErr 면 거부 알림(B2) |
| `quote` | 관측 가격 | observe | 쓸 수 없으면 반환(B1) |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 `go test ./internal/<pkg>/ -count=1 -covermode=set` 프로파일(2026-09-30, `analysis/harness/flm_tables.py`)에서 그 줄로 시작하는 블록의 count.

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:874` `if !o.quoteUsable(quote) {` | 아니오 |
| B2 | if | `:880` `if m.identityErr != nil {` | 예 |
| B3 | if | `:884` `if m.reJudge {` | 예 |
| B4 | if | `:894` `if err := o.opts.Journal.StampExitSnapshotQuarantineSelector(ctx,` | — |
| B5 | if | `:900` `if err != nil {` | 아니오 |
| B6 | switch | `:905` `switch m.state.PolicyKind {` | 예 |
| B7 | case | `:906` `case journal.ExitPolicyLadder:` | 예 |
| B8 | case | `:908` `default:` | 예 |

## Calls and live bindings

`quoteUsable` · **`noteHeldProposal`(원장 읽기 + RecordCritical 창 0, 반환값 없음 — 결과를 바꾸지 않음)** · `alertRefused` · `StampExitSnapshotQuarantineSelector` · `breakEven` · `judgeLadder`/`judgeRatchet`.

## State mutations and fallbacks

a094 몫 없음(알림 기록뿐). 기존: 재판정 도장(B4).

## Safety conclusion

- noteHeldProposal 은 B1 뒤 · B2 앞(억제 · 조기 반환 앞)에 둬서 손절 자신이 무장 발의일 때도 닿는다(D−4.4 · R10). 오류를 반환하지 않으므로 판정 순서 · 결과 무변화(3.R2a). High-risk: yes(판정 진입) — 편집은 호출 한 줄.
