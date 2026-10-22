# Function Logic Map: `recordGateSettings`

- Source: `internal/app/engine/interlock.go` (739-813)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- 기준: 편집 뒤(브랜치 `a163-impl` GREEN 로트, 2026-10-22). 편집 전 번들 문서는 커밋 `78efe1951` · 생성 입력 `../../flm-gen/spec-base.json`
- AST 분기 6 · 반환 2 · defer 0 · go 0
- 위험 등급: High-risk(기동 인터록의 §0.5 audit — 실패하면 기동 거부)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `log *audit.Log` | 열린 audit 로그 | `openAuditLog` | nil 이면 `RecordChange` 무동작 |
| `gate` · `adoption` · `notifications` · `attestationFile` | 로드된 설정 | config | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 742 `if adoption.Rejected != "" {` | `adoptionDetail` 채움 | — | 기존 시험 |
| B2 | if 752 `if notifications.Refused != "" {` | 알림 거부 사유 | — | `TestARefusedNotificationBlockIsAudited` |
| B3 | else 755 `} else if notifications.PublicService && notifications.TopicConfigured {` | else-if 사슬 | — | `TestNotificationSettingsAreAudited` |
| B4 | if 755 `} else if notifications.PublicService && notifications.TopicConfigured {` | 공개 ntfy 경고 | — | `TestNotificationSettingsAreAudited` |
| B5 | range 807 `for _, c := range changes {` | 행 순회 | — | `TestExistingAuditEntriesKeepTheirOrder` |
| B6 | if 808 `if _, err := log.RecordChange(c.action, c.setting, c.value, c.detail); err != nil {` | — | `return err` → 기동 거부 | 기존 시험 |

반환 좌표(AST): 809 · 812

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `log.RecordChange` | 값이 바뀐 설정만 행 | 오류는 B6 반환 | AST |
| `perTradeRiskBudgetAudit` | a139 유도 값 | — | AST |

## State mutations and fallbacks

- audit 파일 append(값이 바뀐 설정만).

## a163 편집

- a163 3.6: 끝 인자 `sourceAgeReport bool` + `changes` 목록 **끝**에 `{audit.ActionExitSourceAgeReport, "engine.exit_source_age_report", strconv.FormatBool(sourceAgeReport), ""}` 한 행. 분기 6 · 반환 2 불변(design D8). — **적용됨**, 편집 전·뒤 AST 분기·반환·defer 수 대조는 `../../flm-gen/README.md`

## Safety conclusion

- Safe edit boundary: 목록 끝 행 하나, 앞 순서 무변경.
- High-risk impact: yes(audit) — 2.16 · 변이 ㉑.
