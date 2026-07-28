# Issues: fix-chase-veto-measurement

구현 중 발견한 설계 결함과, 스펙 의도가 명백해 구현하며 처리한 보완을 기록한다.
(docs/WORKFLOW.md "예외 경로" ②·③)

## I1. D3의 refusal을 관측 행에서 읽으면 `seen_late`가 영구 미측정이 된다 — **blocking에 가까운 설계 결함, safe local로 처리**

**발견**: task 2.1은 "`MeasureFirstSighting`이 최초 순위를 채택하려면 **그 순위가 나온
읽기의** 신규 진입 사실이 `unknown`이 아니어야 한다"고 쓴다. 기존 코드에서 그 사실을
찾는 유일한 경로는 `newlyListedAt(observations, first)` — 즉 `Assess`가 넘긴 관측
슬라이스에서 (source, instant, rank, total)이 일치하는 행을 찾는 것이다.

**문제**: `Assess`는 `DefaultAssessHistory`(10분) 이내의 행만 읽는다
(`internal/candidate/watch.go:346`, `:278`). 최초 관측 행은 후보 수명 대부분의 시점에서
그 창 **밖**이다. 그러므로

- 지금까지 `newlyListedAt`은 **거의 항상 false**를 돌려주고 있었고(그래서 화면의 신규
  진입 표시가 뜬 적이 없다는 결함의 두 번째 원인이다 — 소스가 채우지 않은 것이 첫 번째),
- 3-상태로 바꾸고 그 자리에서 refusal을 걸면, 10분보다 오래된 **모든 후보**의
  `seen_late`가 영구히 `NEW_ENTRANT_UNKNOWN`이 된다. `seen_late`가 구조적으로 측정
  불가가 되므로 후속 change가 임계를 넣어도 아무 일도 일어나지 않는다.

이것은 D20이 이미 한 번 고친 붕괴의 재발이다 — "seen_late collapsed to
NO_FIRST_SIGHTING on exactly the longest-running candidates, the ones the question is
about" (`store.go` first_rank 칼럼 주석). 그리고 design D3 자신의 문장("냉각 만료 후
재승격은 소스가 `no`를 보고하므로 측정 가능하다")과도 모순된다.

**처리**: 신규 진입 사실과 요청 행 수를 **`candidates` 테이블의 `first_rank` 옆 칼럼**으로
저장한다(`first_rank_newly_listed`, `first_rank_requested`, 둘 다 nullable).
D17(`first_price_at`)과 D20(`first_rank_at`)이 이미 두 번 내린 것과 같은 결정이다 —
prunable 테이블보다 오래 살아야 하는 사실은 값 옆의 칼럼이 된다.

**부수 결과**: `Store.NoteFirstRank`의 시그니처가
`(ctx, market, symbol, rank, total, at, source)`에서 `(ctx, market, symbol, FirstRank)`로
바뀐다. 위치 인자 9개를 만드는 대신 읽기 전체를 넘긴다. 사실이 위치와 **같은 문장에서**
쓰여야 한다는 것이 요점이므로(별도 setter는 "출처를 모르는 순위"가 존재하는 창을 만든다)
구조체 인자가 맞다. 호출부는 `scan.go` 1곳 + 테스트 14곳.

## I2. `Sighting`에 요청 행 수를 정수로 실을 수 없다

**발견**: `TestTheVetoCannotSeeAScoreToBeOffsetBy`는 `VetoInputs` 폐포의 모든 숫자
필드를 allowlist로 고정한다. Manager 지시는 이 테스트를 **편집하지 않고** green으로
유지하는 것이다. `Sighting.RankRequested int`를 더하면 그 테스트가 실패한다.

**처리**: `Sighting`은 절단 사실을 3-상태 `Truncation`(불리언 두 개, 숫자 없음)으로만
싣는다. 원 숫자(요청 행 수)는 `Row`·`Reported`·`FirstRank`에 남고 이 셋은 `VetoInputs`
폐포 밖이다. 결과적으로 D6("읽은 값과 파생값은 다른 칼럼")은 관측 계층에서 지켜지고,
veto 입력에는 숫자가 하나도 늘지 않는다. allowlist는 편집하지 않았다.

## I3. 절단 사실의 `unknown`은 refusal이 아니다

design D4는 "요청 행 수와 도착 행 수가 **다르면**" 거부하라고 쓴다. 요청 행 수가
기록되지 않은 읽기(스키마 4 이전의 모든 행)는 "다르다"가 아니라 "모른다"이므로 이
경로로는 거부하지 않는다. 그 행들은 신규 진입 사실도 `unknown`이라 I1의 규칙으로 이미
거부되므로 노출은 늘지 않는다. spec의 두 SHALL을 문자 그대로 읽은 결과이고,
`truncationOf`의 주석에 근거를 남겼다.

## I4. 정정한 기존 테스트 (task 7.4)

- `internal/candidate/store_test.go`
  `TestAStoreAtSchemaTwoOpensMigratesAndKeepsItsRows` — 이전에는 마이그레이션 직후
  `seen_late`가 92%로 **측정된다**고 단언했다. 스키마 4 rung은 두 사실을 backfill하지
  않으므로(할 수 있는 정직한 값이 없다 — 기존 `newly_listed` 칼럼은 모든 행에서 0이다)
  그 단언은 이제 틀린 것을 옳다고 말한다. **무엇을 시험하도록 고쳤는가**: 순위 자체는
  마이그레이션을 건너 살아남고 화면에 보이며, veto는 `NEW_ENTRANT_UNKNOWN`으로 측정을
  거부한다는 것.
- `internal/candidate/metrics_test.go`
  `TestTheSameNumberOfPlacesIsADifferentMoveInADifferentList` — 150행 vs 100행이라는
  존재하지 않는 패널 쌍으로 정규화를 설명했다. **고친 내용**: 실제로 존재하는 두 길이
  (공식 100행, WTS 30행)로 바꿔 전제까지 참이 되게 했다. 두 `NewlyListed` 단언은
  `!x` → `x.Yes()`로 좁혔다 — 이전 형태는 미측정과 측정된 yes를 구분하지 못했다.
- `internal/candidate/wiring_test.go` `pricedRow`, `internal/candidate/veto_test.go`
  `storedFirstRank` — fixture가 두 사실을 측정된 값으로 보고하도록 고쳤다. 그대로 두면
  이 패키지의 모든 스캔·veto 테스트가 refusal 경로만 시험하게 된다.
- `internal/candidate/watch_test.go`
  `TestAScanReportsTheShadowRecordForEveryCodeThatHasOne` — "첫 스캔에 측정 가능하다"는
  주석이 이제 거짓이다(세션의 첫 읽기는 거부된다). **고친 내용**: 측정 가능한 이유가
  저장된 순위 **와** 소스가 그 읽기에 대해 답했다는 사실 둘 다임을 적었다.
