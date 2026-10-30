# Branch Test Map: `ExitObserver.record(params=7, results=1)`

보호 제안일 때만 clearTheSymbol critical, 다른 제안은 normal. 기존 억제·원장 오류·release 계약 유지. 시험 TestA166ExitProposalPriority, TestA166FullTakeProfitCancelIsNormal.

| 분기 | AST 위치와 조건 원문 | 결과 범위 |
|---|---|---|
| B1 | if L1313: `if !o.quoteUsable(quote) {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B2 | if L1334: `if judgement.ObservationSource == "" {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B3 | if L1335: `if quote.FetchedAt.IsZero() {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B4 | else L1337: `} else {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B5 | if L1358: `if orderable && m.floorExpectedState != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B6 | else L1363: `} else if orderable && (snapshot.CancelPendingFirst \|\| isFullExit(proposal)) {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B7 | if L1363: `} else if orderable && (snapshot.CancelPendingFirst \|\| isFullExit(proposal)) {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B8 | if L1364: `if m.reJudge && !isProtective(proposal) {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B9 | else L1386: `} else {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B10 | if L1388: `if isProtective(proposal) {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B11 | if L1392: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B12 | if L1395: `if !cleared.cleared {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B13 | else L1402: `} else {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B14 | if L1410: `if orderable {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B15 | if L1412: `if intentID == "" {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B16 | if L1424: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B17 | if L1425: `if errors.Is(err, journal.ErrProposalPending) {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B18 | if L1431: `if errors.Is(err, journal.ErrExitSnapshotQuarantined) {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B19 | if L1444: `if recorded.ArmedProposal == nil \|\| recorded.ArmOutcome != journal.ExitArmArmed {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
