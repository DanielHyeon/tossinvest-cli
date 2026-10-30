# Branch Test Map: `TestA166ReplayAdmissionRefusalPreservesInDoubtJournal(params=1, results=0)`

기존 admission 거절 시험을 실제 engine BUY 출처로 만들어 새 required 검사를 통과시킴. 전송 0·IN_DOUBT·count 유지 확인.

| 분기 | AST 위치와 조건 원문 | 결과 범위 |
|---|---|---|
| B1 | if L26: `if err != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B2 | if L33: `if e := os.Chmod(root, 0700); e != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B3 | if L36: `if e := accountbudget.Provision(root, "scope", "instance"); e != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B4 | if L40: `if e != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B5 | if L46: `if r.URL.Path == "/oauth2/token" {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B6 | if L58: `if e != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B7 | if L61: `if _, e = auth.AuthHeaders(context.Background()); e != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B8 | if L65: `if e != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B9 | if L69: `if e != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B10 | if L73: `if e != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B11 | if L79: `if e = held.Release(); e != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B12 | if L83: `if e != nil {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
| B13 | if L86: `if posts.Load() != 0 \|\| after.State != journal.StateInDoubt \|\| after.ReplayCount != 1 {` | 선택 범위 외 기존 분기 개별 통과는 주장하지 않음 |
