# Function Logic Map: `reasonText`

- Source: `internal/operatorview/exit_line.go` (198-248)
- Qualified: `reasonText` · revision: current(편집 뒤, worktree `~/tossos-work/a207` 브랜치 `feat/a207-blocked-exit-retry-dedupe`)
- AST evidence: `ast.json` (source_sha256 999f66b69168…) — AST branches 24 · returns 23
- 편집 전 증거: `analysis/ast-evidence/operatorview--reasonText.ast.json`(198-245, 23 갈래, sha e4af96f4bb63)
- Risk scan: `risk-pattern-report.md` (파일 단위)

**역할.** unknown 사유 코드를 사람용 한국어로 바꿈(콘솔). a207 은 `repeat_of_refused_proposal` case 하나를 더함(사람 결정 2 — 「반복 거절(요약)」).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `reason` | 코드 문자열 | 리더의 UnknownReason | 모르는 코드 → 일반 문구(코드 노출 없음) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:199` `switch strings.TrimSpace(reason) {` | — | — | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B2 | `:200` `case "engine_not_running":` | — | 문구 return | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B3 | `:202` `case "snapshot_quarantined":` | — | 문구 return | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B4 | `:204` `case "observation_older_than_limit":` | — | 문구 return | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B5 | `:206` `case "observation_in_future":` | — | 문구 return | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B6 | `:208` `case "invalid_observed_at":` | — | 문구 return | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B7 | `:210` `case "not_evaluated_yet":` | — | 문구 return | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B8 | `:212` `case "no_saved_evaluation":` | — | 문구 return | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B9 | `:214` `case "legacy_snapshot_absent", "legacy_event":` | — | 문구 return | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B10 | `:216` `case "repeat_of_refused_proposal":` | — | 문구 return | `TestA207TheRepeatReasonReadsAsASummarisedRepeatedRefusal` |
| B11 | `:219` `case "invalid_stored_snapshot", "invalid_effective_snapshot":` | — | 문구 return | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B12 | `:221` `case "legacy_policy_identity_unknown", "legacy_adoption_context_required":` | — | 문구 return | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B13 | `:223` `case "partial_snapshot_tuple", "partial_evaluated_tuple", "partial_seed_tuple", "partial_tuple":` | — | 문구 return | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B14 | `:225` `case "partial_policy_tuple", "invalid_policy_identity":` | — | 문구 return | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B15 | `:227` `case "flattened_snapshot_mismatch":` | — | 문구 return | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B16 | `:229` `case "ambiguous_exit_evidence":` | — | 문구 return | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B17 | `:231` `case "exit_evidence_unlinked":` | — | 문구 return | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B18 | `:233` `case "lineage_cycle":` | — | 문구 return | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B19 | `:235` `case "lineage_ambiguous":` | — | 문구 return | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B20 | `:237` `case "lineage_depth_exceeded":` | — | 문구 return | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B21 | `:239` `case "lineage_scope_mismatch":` | — | 문구 return | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B22 | `:241` `case "invalid_event_evidence":` | — | 문구 return | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B23 | `:243` `case "":` | — | 문구 return | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B24 | `:245` `default:` | — | 문구 return | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | 정규화 | — | AST |

## State mutations and fallbacks

- 없음(순수 함수).

## Safety conclusion

- Safe edit boundary: case 하나 추가. 다른 case · default 무변경.
- High-risk impact: no — 콘솔 표시 문구.
