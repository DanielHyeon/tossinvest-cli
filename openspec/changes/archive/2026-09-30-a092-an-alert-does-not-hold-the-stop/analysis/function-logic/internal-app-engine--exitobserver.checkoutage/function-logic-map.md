# Function Logic Map: `ExitObserver.checkOutage`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json` — **편집 뒤**, :817–859, 분기 6 · 반환 4 · 호출 21, source_sha256 `2d34b5c57f25…`, 추출 커밋 `e55102f0`. 편집 전 번들은 `analysis/pre-edit/unit5/`(없으면 단위 ④ 번들이 편집 전).
- Risk scan: `risk-pattern-report.md`
- 편집(착지 단위 ⑤ `e55102f0`): 승격 오류 안에서 `ErrModeAnnouncementFailed` 는 강화됨으로 로그하고 `cycle.Escalated = changed` 를 세움 — 분기 하나 추가(B#7).
- 재번호: difflib 정렬(편집 전 `b01e0cd0` 소스 대비): 새 분기 B6(:849).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 + 위 편집 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if since.IsZero()` (:819) | — | — | `TestA092CredentialTighteningRecordsWithoutSending`, `TestA111MissingManagedSymbolIsInvalidEvidenceAndDoesNotResetTheOutage` |
| B2 | `if o.clk.Now().Sub(since) < o.outageAfter()` (:822) | — | — | `TestA092CredentialTighteningRecordsWithoutSending`, `TestA111MissingManagedSymbolIsInvalidEvidenceAndDoesNotResetTheOutage` |
| B3 | `if o.outageRaised` (:825) | — | — | `TestASustainedOutageBlocksEntriesAndAlertsOnce` |
| B4 | `if o.opts.Escalate == nil \|\| strings.TrimSpace(o.opts.AccountRef) == ""` (:843) | — | — | (미실행) |
| B5 | `if err != nil` (:848) | — | — | `TestA092AnUnannouncedOutageTighteningStillCountsAsEscalated` |
| B6 | `if !errors.Is(err, journal.ErrModeAnnouncementFailed)` (:849) | — | — | (미실행) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Format`, `Round`, `Seconds`, `String`, `Sub`, `errors.Is`, `int`, `o.alert`, `o.clk.Now`, `o.logErr`, `o.opts.Escalate.EscalateOperatingMode`, `o.outageAfter`, `since.IsZero`, `since.UTC`, `string`, `strings.TrimSpace` | 편집 뒤 호출 | 위 편집 참조 | AST |

## State mutations and fallbacks

- 위 편집 요약이 전부 — 나머지는 편집 전과 같음(편집 전 번들 대조).

## Safety conclusion

- Safe edit boundary: 위 편집만.
- High-risk impact: 중간 — 보고 정확성(손절 무관).
