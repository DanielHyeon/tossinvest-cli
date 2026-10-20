# Function Logic Map: `renderCalibrationReport`

- Source: `tools/a128-jev-shadow-probe/report.go`
- AST evidence: `ast.json` (`go run ./tools/logic-map --file tools/a128-jev-shadow-probe/report.go --func renderCalibrationReport`,
  2026-10-20, 편집 뒤 워킹트리 — 분기 22 · return 1)
- Risk scan: `risk-pattern-report.md` (go-float64-cast 13 건 — 전부 적중률·평균 p·중앙값 표시 계산, 돈·수량 경로 아님)
- 범위: s4(tasks 5.2) 편집은 함수 **끝에 덧붙인 루프 하나**(B19~B22)뿐. B1~B18 은 base(bbe4ceaa6)와 같은 분기이며 출력이 바이트 그대로임을
  골든 접두 시험이 고정함. High-risk 경로 아님(측정 보고서 생성기 — 주문·원장(거래)·인증 경로 무관).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `contents.Judgments` | 원장 judgment 행(옛 행 포함) | `readLedger` | 해석 실패는 readLedger 가 오류로 멈춤 — 이 함수에 안 옴 |
| `contents.Labels` | judgment id → 라벨 | `readLedger` | 없는 id = 미라벨로 셈(분모 제외) |
| `contents.StateMinutes` | state digest → minutes_since_open, nil 가능 | `readStateMinutes`(report 하위 명령만) | nil·없는 digest = 「알 수 없음」 구간 |
| `hitRules` | 질문 id → 적중 함수 4 개 | `report.go` | 없는 질문 = 표 대신 이름만(B10) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1~B18 | base 와 같음(행 집계·헤더·구간표·결선 비교) | `out` 에 씀 | 없음 | 기존 report 시험 + `TestPreS4LedgerStillReportsTheSameAndAddsTimeOfDay`(base 골든 접두) |
| B19 | `range sortedKeys(questionIDs)` — 시간대 표를 질문마다 | 없음 | 없음 | `TestPreS4LedgerStillReportsTheSameAndAddsTimeOfDay` |
| B20 | `!known` — 적중 규칙 없는 질문은 시간대 표도 건너뜀 | 없음 | `continue` | 아래 BTM B20 |
| B21 | `range sortedKeys(keysOf(lots))` — 로트마다 | 없음 | 없음 | 같은 픽스처(top10·s3 두 로트) |
| B22 | `range marketList` — 시장마다(ALL 포함) | `renderSessionTimeSection` 이 `out` 에 씀 | 없음 | 같은 픽스처(KR·US·ALL 제목 단언) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `renderSessionTimeSection` | 질문·로트·시장의 시간대 표(새 함수) | 오류 없음(순수 문자열 생성) | AST calls · `TestTimeOfDayTableJudgesAtThirtyLabels` |
| `sortedKeys`·`keysOf` | 결정적 순서 | 없음 | 기존 |
| `hitRules[...]` | 적중 규칙(J3·J4 추가) | 없음 | `TestS4HitRuleBoundaries` |

## State mutations and fallbacks

- 함수 밖 상태 변경 0(로컬 `strings.Builder` 만). 파일 읽기는 이 함수가 아니라 `reportCommand` → `readStateMinutes`(읽기 전용).
- fallback: StateMinutes 가 없으면 모든 행이 「알 수 없음」 — 조용한 생략 없음(행 수가 그 줄에 보임).

## Safety conclusion

- Safe edit boundary: 덧붙임 루프만, 앞 출력 바이트 불변(골든 접두 + 라이브 원장 base/new 대조 23,235 바이트 일치).
- High-risk impact: no
