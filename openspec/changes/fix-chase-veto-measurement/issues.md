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

## I3. 절단 사실의 `unknown`은 refusal이다 — **2026-07-28 뒤집음 (리뷰 F4)**

**원래 결정(틀렸다)**: design D4는 "요청 행 수와 도착 행 수가 **다르면**" 거부하라고
쓴다. 요청 행 수가 기록되지 않은 읽기는 "다르다"가 아니라 "모른다"이므로 거부하지
않는다 — 그 행들은 신규 진입 사실도 `unknown`이라 I1의 규칙으로 이미 거부되므로 노출은
늘지 않는다.

**무엇이 틀렸나**: "이미 거부된다"는 *이 빌드가 쓰는 행*의 성질이지 **타입의 성질이
아니다.** (측정된 신규 진입, 미기록 요청) 쌍은 표현 가능하고 — fixture, 옛 빌드가 절반만
채운 저장소, 미래의 producer — 실제로 **측정되었다.** 1행 읽기가 그 쌍을 가지면 백분위
0(자기 목록의 최하위)이 나오고, 0은 존재할 수 있는 모든 `seen_late` 임계를 **통과한다.**
미측정이 통과로 바뀌는 것은 D10이 막으려는 바로 그 형태다.

**지금 코드가 하는 것**: `MeasureFirstSighting`이 `!Truncation.Known()`이면 새 사유
`REQUEST_UNRECORDED`로 거부한다. `READING_TRUNCATED`와 별개 사유인 이유는 이 패키지가
다섯 번 지킨 규칙이다 — 미측정은 측정된 부정이 아니다. 운영자 대응도 다르다: 절단은
엔드포인트가 짧게 주는 살아 있는 결함이고, 미기록은 스키마 4 이전 행이라 쫓을 것이 없다.
`truncation_test.go`의 해당 테스트는 이름·주석·본문이 서로 어긋나 있었고(이름과 주석은
거부를 말하고 본문은 `Measured`를 단언했다) 셋을 새 동작에 맞췄다.

## I5. 스키마 롤백 계획 (WORKFLOW §0.6) — **2026-07-28 추가 (리뷰 F9b)**

`store.go`의 두 주석이 "옛 빌드가 새 파일을 열면 선택하지 않는 칼럼을 읽을 뿐"이라고
적었는데 **거짓이다.** `migrate()`가 자기 `SchemaVersion`보다 높은 stamp를 만나면
`ErrSchemaTooNew`로 거절하고 `TestAStoreFromANewerBuildIsRefused`가 그것을 고정한다.
거절이 옳은 동작이다 — 이 행들은 veto를 먹이고, 모르는 칼럼을 "없음"으로 읽는 빌드가
정확히 이 패키지가 거부하는 것이다.

**따라서 downgrade는 스키마 조작이 아니다. 롤백 계획은 `candidates.db`를 지우거나 옆으로
치우고 옛 빌드가 새로 만들게 하는 것이다.** 비용은 관측 이력 전체와 그 안의 모든
`first_seen_at`·`first_price`·`first_rank`다. 감당 가능한 이유는 이 행들이 **파생**이기
때문이다 — 원장은 다른 파일·다른 락(D2)이고, 어떤 주문도 이 행에 의존하지 않으며, 다음
스캔부터 다시 측정한다. 공짜는 아니다: 진행 중인 모든 후보 수명이 초기화되고 각각 다시
발견될 때까지 `seen_late`·`extended`가 미측정이다. 두 주석을 제자리에서 날짜 붙여
정정했다.

## I6. 직전 읽기 집합에 유효 조건을 붙였다 — **2026-07-28 (리뷰 F1·F2)**

리뷰가 §2가 끊으려던 경로가 **여전히 열려 있음**을 실행 가능한 probe로 보였다.
`candidatesrc`의 두 소스는 읽기 실패 시 `rememberRead` **앞에서** 반환하므로, 기억된
집합이 임의로 긴 장애를 그대로 살아남는다. 429 백오프가 정확히 그 상황이고, 냉각 30분 +
staleness 10분 = 40분이면 감시목록 전체가 만료되어 회복 스캔이 패널을 일괄 재승격한다 —
그때 목록을 떠난 적 없는 심볼은 장애 이전 집합에 대해 `no`를 받고, `no`는 최초 관측을
자격 부여하는 답이다. probe: `ErrRateLimited` 200회 연속 후 회복 →
`seen_late measured=true percentile=96`.

**결정 1 (F1, 신선도)**: 직전 읽기는 `candidate.DefaultStalenessTTL`(10분)보다 오래되면
쓰지 않는다. 값은 여기서 고르지 않았다 — 그 상수 자신의 두 근거를 그대로 쓴다.
(a) `DefaultStalenessTTL`과 `MaxRankPriorAge`는 둘 다 `BackoffLadder` 마지막 rung(300s)의
**두 배**이고, 그 근거는 ladder 옆에 적혀 있다: 그만한 후퇴는 정상 운영이며 그 사이에
일어난 것을 만료시켜서는 안 된다. 소스가 백오프 안에 있는 것은 정지가 아니므로 기억은
ladder 전체를 견뎌야 한다. (b) 그 지점을 넘으면 기억이 자격 부여할 대상 자체가
사라진다 — 아무도 재승격하지 않은 후보는 `DefaultStalenessTTL`에 암묵 냉각되고 거기서부터
감시목록이 회전한다. 즉 기억은 **자기가 자격 부여할 후보 수명이 사는 동안 정확히 그만큼**
유효하다. 스캔 간격(공식 15s·WTS 5s)과 엔진 양보(2배)는 전부 ladder 마지막 rung 훨씬
안쪽이라 별도 항이 필요 없다. 시계가 뒤로 가면(음수 나이) 역시 쓰지 않는다.

**결정 2 (F2, 완전성)**: 소스 자신의 비교(요청 ≠ 도착)가 짧다고 말하는 읽기는 기억을
**교체하지 않는다.** probe: 100행 중 3행 degraded 읽기 다음의 완전 읽기가 100행 중
**97행을 신규 진입으로** 보고하고 전부 `Known()`이며 전부
`candidates.first_rank_newly_listed`에 기록되었다. 빈 읽기도 같은 비교로 짧으므로 같은
규칙에 걸린다 — `TestAnEmptyReadingIsStillAReadingOfThisList`가 반대 방향을 고정하고
있었고 뒤집었다(`TestAnEmptyReadingDoesNotBecomeThePreviousReading`).

**부수 결과**: `OfficialRanking`·`WTSPopular`·`Panel`이 `clock.Clock`을 받는다. 기억에
나이가 생겼으므로 시각 출처가 필요하고, 주입하지 않으면 그 경계를 테스트가 몰 수 없다.
nil은 `clock.System()`이다.

**부수 결과 2 (design D3 문장 하나가 틀렸다)**: D3은 "냉각 만료 후 재승격은 소스가 직전
읽기를 갖고 있으므로 `no`를 보고한다"고 쓴다. 이제 그 `no`는 **도달 불가에 가깝다.**
소스가 계속 읽고 있었다면 그 심볼이 목록에 있는 한 스캔이 계속 승격하므로 냉각되지
않는다. 냉각·만료가 일어나려면 심볼이 목록을 떠났어야 하고, 떠났다 돌아오면 답은
`yes`다. 반대로 40분간 **읽지 못했다면** 기억이 만료되어 `unknown`이다. 결론(재승격은
측정 가능하다)은 유지되지만 경로는 `yes`다. 테스트를
`TestARePromotionAfterExpiryIsQualifiedByTheReadingThatSawTheSymbolReturn`으로 고쳤다.

## I7. 자격 부여할 수 없는 최초 순위는 아예 쓰지 않는다 — **2026-07-28 (리뷰 F7·F9a)**

`first_rank`는 write-once다. 자격 미상(신규 진입 `unknown` 또는 요청 행 수 미기록) 읽기의
순위를 쓰면 그 후보 수명 전체의 `seen_late`가 영구 미측정이 되고, 나중에 채울 수도 없다
(I8). 특히 `tossctl candidate scan`은 패널을 만들고 **한 번 읽고 끝나므로** 모든 읽기가
소스의 첫 읽기다 — 즉 진단용 스캔 한 번이 그때 최초 순위를 받은 후보들에게 되돌릴 수 없는
손상을 입힌다. `watch`와 저장소를 공유하므로 실제로 일어날 수 있다.

**결정**: `recordFirsts`가 그런 읽기의 순위를 저장하지 않고 `ScanResult.FirstRanksHeld`로
센다. 저장하지 않으면 `NO_FIRST_RANK`(복구 가능)이고, 다음 tick의 자격 있는 읽기가 아직
±TTL 동일성 창(10분) 안에 있으므로 그것이 최초 관측이 된다. 공식 간격이 15초이므로
세션 시작 비용은 **한 tick**이지 한 세션이 아니다.

**절단된 읽기는 저장한다.** 두 사실 모두 취해졌고, 나오는 거부(`READING_TRUNCATED`)가
운영자가 쫓을 수 있는 진단이며, 그것이 이 후보를 처음 봤을 때의 정직한 기록이다.

**F9a와의 관계**: `veto.go`의 `NEW_ENTRANT_UNKNOWN` 주석은 "두 번째 읽기를 기다리라"고
썼고 write-once 때문에 **거짓이었다.** 이제 참이다 — 참이 된 이유가 이 결정이다. 세션
시작의 일시적 경우는 `NO_FIRST_RANK`로, 영구적 경우(자격 있는 읽기가 동일성 창 안에
도착하지 못한 채 저장된 순위)는 `NEW_ENTRANT_UNKNOWN`으로 갈린다. 사유가 이미 둘로
갈리므로 별도 카운터를 만들지 않았다.

## I8. 스키마 4 이전 후보는 재자격 부여되지 않는다 — 문장을 고쳤다 (리뷰 F6)

리뷰 F6은 `NoteFirstRank`가 제안된 읽기가 저장된 `(rank, total, at, source)`와 **정확히**
일치할 때 NULL 자격 칼럼을 채우게 하라고 제안했다. **구현하지 않았고, 대신 문장을
고쳤다.** 근거:

1. 그 경로에 **생산 호출자가 없다.** `recordFirsts`는 `FirstRank.Recorded()`인 후보에
   대해 읽기를 제안조차 하지 않고, 제안하도록 바꿔도 다음 tick의 관측은 **다른 instant**를
   가지므로 정확 일치가 성립하지 않는다. 구현하면 테스트만 있는 죽은 코드가 된다.
2. 유용해지려면 **나중 읽기의 답**을 저장된 위치 옆에 써야 하는데, 그것이 D3이 거부하는
   대입 그 자체다. 두 사실은 *그 위치가 나온 읽기*를 서술하고 그 읽기는 사라졌다.
3. 채울 원본도 없다. v3 rung의 backfill은 관측 행에서 위치를 가져오는데 그 행들의
   스키마-4 칼럼도 NULL이다(같은 rung에서 추가되고 backfill하지 않는다).

**참인 문장**: 이 후보의 `seen_late`는 **이 수명 동안** 미측정으로 남는다. 수명이 끝나면
(냉각 → 만료) `Promote`가 `first_rank`를 지우고, 다음 교차가 실행 중인 소스가 자격 부여할
수 있는 읽기에서 새 위치를 기록한다. **마이그레이션된 저장소는 스캔 속도가 아니라 후보
회전 속도로 회복한다.** `store_test.go`의 문장을 그렇게 고치고,
`migrated_firstrank_test.go`가 양쪽(자격 있는 스캔 5회 후에도 채워지지 않음 / 수명이
끝나면 측정 가능)을 고정한다.

## I9. 열린 질문을 한 세션 안에 답할 수 있게 했다 (리뷰 F8)

공식 순위 엔드포인트가 요청한 100행을 실제로 주는지는 측정된 적이 없다. 주지 않는다면
모든 읽기가 `READING_TRUNCATED`이고 후속 change가 임계를 고를 분포가 **존재하지 않는다.**
지금까지 그 상태는 후보 수명 여러 개가 지난 뒤 사유 census로만, 그것도 어느 소스인지
모른 채 보였다.

**추가한 것 둘**: (a) `ScanResult.Readings` — 소스별 요청 행 수와 도착 행 수. 스캔 한 번이
직접 답한다. 비교(1비트)가 아니라 두 숫자를 싣는다: 100 요청 99 도착과 100 요청 3 도착은
같은 `Truncation`이고 전혀 다른 대화다. (b) `candidate.TallySightingSources` — 최초 관측을
그것을 만든 소스별로 나눈 census. 두 표면(`tossctl candidate scan`, `/signals`)이 **같은
reducer**를 쓴다(§4.3의 규칙). `/signals`는 스캔 기록을 볼 수 없으므로 (a)를 렌더할 수
없고 (b)만 렌더한다.

**한계를 기록한다**: 요청 행 수는 Row를 타고 오므로 **행이 0개인 읽기는 (a)에 나타날 수
없다** — 가장 극단적인 절단이 가장 설명되지 않는다. `Reading` 자체에 수를 실으면 고쳐지고,
그것은 모든 소스가 구현하는 seam을 바꾸는 더 넓은 change다. 빈 읽기는 이미 자기 사유와
함께 `Missing`에 들어간다.

## I10. 패널 크기 drift 가드의 한계를 테스트가 말한다 (리뷰 F10)

`deniesTheSize`가 `"no panel"`·`"corrected 2026"`을 **행 전체에 대한 부분 문자열**로
면제하고 있었다. 즉 `// The KR panel returns 150 rows and the US panel 100 (no panel note)`
가 통과한다 — 평문 주장에 네 글자를 붙인 것이다.

**조인 것**: 맨 토큰 `"no panel"`/`"No panel"` 제거(실제 문장은 "no panel **has ever
returned** 150 rows"라 기존 항목이 이미 덮는다), 날짜 없는 `"corrected 2026"` 대신 완전한
날짜 정규식 요구, 부인구가 숫자와 **같은 문장** 안에 있을 것(문자 거리 상수를 만들지
않는다 — 그것이야말로 이 change가 다루는 병이다), 한국어 counter `150개 행` 추가.

**여전히 못 잡는 것을 `TestWhatThisGuardCannotCatch`가 단언한다**: 글자로 쓴 수("one
hundred and fifty rows"), 두 주석 줄에 걸쳐 쪼개진 수, `rows`·`행` 이외의 단위어("a
150-symbol list" — `roughly 150 symbols`가 이 저장소의 참인 문장이라 단위어에 넣을 수
없다), 문자열 리터럴·식별자·Markdown(주석만 읽는다 — `add-candidate-discovery` design.md에
있던 원본 허구를 이 가드는 못 찾았을 것이다), 같은 문장 안에서 부정직하게 쓰인 부인구.
텍스트 가드에는 한계가 있고, 없는 척하는 것이 한계를 적는 것보다 나쁘다.

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

## I11. 소비자 가드가 열려 있던 세 곳 — 그리고 좁힌 곳 하나 (리뷰 F5)

리뷰가 셋 다 금지된 코드를 넣어 green을 확인했다. 셋 다 고쳤고, 고친 뒤 같은 방법으로
red를 확인했다(`cmd/tossctl`에 probe 파일을 넣고 실패를 본 뒤 제거).

**(a) 클라이언트는 패키지 이름을 남기지 않는다.** `namesAnOrderVerb`는 `pkgIdent.Sel`만
본다. 그런데 공식 백엔드의 쓰기는 전부 `*official.Client`의 **메서드**이므로
(`internal/official/conditional_writes.go`) `client.CreateConditionalOrder(…)`에는
패키지 이름이 없다. `orderVerbs["official"] = {"Place","Cancel","Modify"}`는 이 트리에서
**아무것도 매치하지 않는다** — 패키지 수준 `official.Place*`는 존재하지 않는다. 보이는
것은 **생성**이므로 `official.New`를 금지하고, `cmd/tossctl/candidate.go`(veto 블록을
렌더한다)에 있던 `official.New`·`tossclient.New`를 새 파일
`cmd/tossctl/candidatepanel.go`로 옮겼다. 기존 세 접두사는 남긴다 — 비용이 없고, 누군가
패키지 수준 helper를 만드는 날 잡는다.

**(b) import 게이트가 스캔 전체를 막고 있었다.** `internal/candidate`를 import하지 않는
파일은 통째로 건너뛰었다. Go import는 파일 단위이고 `cmd/tossctl`은 한 패키지이므로,
같은 패키지의 helper에서 `candidate.Chase`를 받아 `.Passed()`를 부르는 새 파일은 import가
**0개 늘지 않는다** — design D6 case 2 그 자체다. 이제 스캔은 모든 파일에서 무조건
돌고, import는 **qualified 형태(`candidate.X`)를 해석할 수 있는지만** 결정한다.

**(c) alias를 해석하지 않았다.** candidate 쪽은 처음부터 해석했고
(`candidateImportName`), 주문 쪽은 철자를 비교했다. `import eg ".../internal/execgw"` +
`eg.Submit(…)`가 통과했다. 이제 주문 패키지들도 파일별 import 이름을 해석하고, dot
import는 selector가 아예 없으므로 **그 자체로 위반**으로 본다.

**좁힌 것 하나 — Manager 지시와 다르게 결정했다.** 지시는 unqualified 스캔 대상으로
`Passed`/`Chase`/`Verdict`/`VetoTally` 넷을 명명했다. `Verdict`를 unqualified로 스캔하면
`cmd/tossctl/verify.go:363`(`o.Verdict`, 실계좌 검증 단계의 결과),
`internal/console/data.go:318`(`s.Verdict.Terminal()`, 주문 상태)와 콘솔 테스트 셋이
걸린다 — chase 판정을 본 적 없는 파일 다섯 개다. 철자 충돌 때문에 allowlist를 다섯 줄
늘리면 목록이 "누군가 생각해 본 파일의 집합"이기를 그만두고, 읽을 수 없는 allowlist는
아무도 재검토하지 않는다. **그래서 unqualified 집합은 `Passed`·`Chase` 둘이다.** 빠져나갈
틈은 없다: import 없이 판정을 쥐려면 helper에서 받아야 하고, 쓰려면 `.Chase`나 `.Passed`를
선택해야 한다(`.SeenLate`·`.Extended`·`.NearHigh`는 `.Chase`를 통해서만 닿고 `VetoTally`의
통과 수 자체가 `.Passed`다). 타입 이름을 직접 쓰려면 import가 필요하고 그것은 qualified
스캔이 본다.

## I12. 테스트 위생 (리뷰 F11)

**(a) 새 칼럼 넷이 어디에서도 명명되지 않았다.** `candidateColumns` 단언들이
`first_rank_source`에서 멈추고 v3 fixture가 없어서 3→4 rung은 v1 테스트가 세 rung을 한
트랜잭션에서 오르며 **전이적으로만** 닿았다 — 네 ALTER에서 CHECK 문구를 지워도 아무것도
실패하지 않았다. `schema_four_test.go`가 v3 fixture와 네 칼럼, 그리고 각 CHECK를 거부해야
하는 값으로 직접 몬다. 마이그레이션된 저장소와 새로 만든 저장소를 같은 helper로 검사한다
— SQLite는 기존 테이블에 테이블 수준 제약을 추가할 수 없으므로 칼럼별로 적지 않으면 둘의
모양이 갈리고 아무도 그것을 말하지 않는다. 확인: CHECK 문구를 지우면 네 단언이 전부
실패한다.

**(b) 길이 가드 없는 루프 넷.** `previous_reading_test.go`의 행 슬라이스 루프는 행이
빠지면 공허하게 통과했다. 넷 다 `if len(x) != N { t.Fatalf }`를 앞에 뒀다.

## I13. §3의 절단 사슬에 배선 테스트가 없었다 (리뷰 F3)

`scan.go`의 두 대입(`RankRequested: r.RankRequested`, `Requested: o.Reported.RankRequested`)을
각각 `0`으로 바꿔도 **전 스위트가 green**이었다. 같은 mutation을 형제 사실(`NewlyListed`)에
하면 세 테스트가 실패하므로 §1은 고정되어 있고 §3은 아니었다 — 스캔 수준 fixture가 전부
`RankRequested = total`이라 절단된 읽기를 만드는 배선 테스트가 하나도 없었다.
`truncation_wiring_test.go`가 `Row{Rank:1, RankTotal:3, RankRequested:100}`을
`Cycle` + `Assess`로 몰아 `Chase.SeenLate.Reason() == READING_TRUNCATED`를 단언하고,
완전한 3행 읽기가 여전히 측정 가능하다는 대조군을 옆에 둔다(길이가 결정하지 않는다).
확인: 두 mutation 모두 이제 실패한다.

## I14. 음수 요청 행 수의 두 거부에 테스트가 없었다 — Function Logic Map 생산이 찾았다 (task 8.1)

Branch Test Map을 쓰면서 두 분기에 대해 "이 분기를 덮는 테스트"를 적을 수 없었다.

- `internal/candidate/candidate.go:Observation.validate` B7 — `o.Reported.RankRequested < 0`
- `internal/candidate/store.go:Store.NoteFirstRank` B5 — `first.Requested < 0`

둘 다 **이 change가 만든 분기**이고, 둘 다 아무 테스트도 넘기지 않았다. 더 나쁜 것은
`truncation_test.go:TestTruncationComparesTheTwoNumbersTheSourceDeclared`의 표에
`{"a negative request, which validate refuses at the boundary", -1, 100, TruncationUnknown()}`
라는 **부제목이 그 주장을 하고 있었다**는 점이다 — 이름이 동작을 서술하고 본문은 그것을
확인하지 않는 형태이며, 앞 change의 evidence 생산이 찾은 결함 넷 중 하나와 같은 모양이다.

**왜 조용한 결함인가**: `positive()`가 비양수를 SQL NULL로 접고 `truncationOf`가
`requested <= 0`을 `TruncationUnknown()`으로 접는다. 그래서 음수가 경계를 통과하면 그 행은
"아무도 재지 않았다"로 **위장한 채** 저장되고, 저장 뒤에는 미기록과 구분할 방법이 없다.
`first_rank_requested`는 write-once 칼럼이므로 그 후보의 남은 생명 전체가 그 값으로 답한다.

**처리**: `internal/candidate/negative_request_test.go`(신규 파일)가 두 거부를 각각 몰고,
`truncationOf(-1, 100) == TruncationUnknown()`을 먼저 단언해 거부가 무엇에 대한 것인지를
고정한다. `NoteFirstRank` 쪽은 거부가 **쓰기 이전**임(칼럼이 여전히 비어 있음)과
`Requested = 0`(부재)은 여전히 수락됨을 함께 확인한다 — 거부가 부호에 대한 것이지 자격
칼럼의 부재에 대한 것이 아니라는 것.

**확인**: 두 case를 `false && …`로 무력화하면 두 테스트가 모두 실패한다(2026-07-28 실행).
새 파일이므로 FLM 대상 집합은 78개로 그대로다(`check_analysis.py`는 새로 추가된 파일의
함수를 면제한다).

## I15. Branch Test Map이 기록한 미커버 분기 (task 8.1)

결함은 아니지만 evidence가 정직하려면 남겨야 하는 목록이다. **전부 이 change 이전부터 있던
분기**이고, 이 change가 만든 분기 중 커버되지 않은 것은 I14의 둘뿐이었다(이제 커버된다).

- `internal/candidatesrc` — 빈 심볼 skip 넷(`officialRanking.Read` B4,
  `officialRanking.rememberRead` B2, `wtsPopular.Read` B4, `wtsPopular.rememberRead` B2).
  빈 `Symbol`을 가진 fixture가 없다. 행 집합과 기억 집합이 **같은 조건으로** 걸러진다는
  것이 이 change의 전제인데, 그 대칭은 구조로만 서 있다.
- `internal/candidatesrc` — `OfficialRanking`의 `count <= 0` 반쪽, `WTSPopular`의
  `size <= 0`. 생산 호출부는 `Panel`의 리터럴 100·30뿐이다.
- `internal/candidatesrc.Panel` — 주입 clock이 네 소스에 **전달되는지**를 잡는 테스트가
  없다. 유효성 테스트들은 두 생성자를 직접 부른다.
- `internal/candidate.Collect` — 인자 검증 둘(빈 시장, zero instant), 패널·not-asked 모순,
  빈 심볼 skip, 저장 실패 전파 넷, mis-wire와 성공 읽기가 함께 있는 경우.
- `internal/candidate.Observation.validate` — 여덟 case 중 테스트가 넘기는 것은
  `Rank > 0 && RankTotal == 0` 하나(`TestARankWithoutItsListLengthIsRefused`)와 이번에
  추가한 B7뿐이다.
- `internal/candidate.Store` — 손상된 stamp·I/O 실패 경로 전반(`Promote`의 인자 검증 둘 포함).
- `cmd/tossctl.buildCandidatePanel` / `wtsPopularityReader` — **직접 테스트가 하나도 없다**
  (`rg 'buildCandidatePanel' cmd/tossctl/*_test.go` 0건). 이 change가 이 함수에 한 것은
  파일 이동과 `clock.System()` 인자 하나이고, 이동 자체는 소비자 가드 세 테스트가 잡는다.
- `cmd/tossctl.consoleSignalsMarket` B1 — `Assess` 실패 경로. 렌더 쪽만
  `internal/console`이 `Markets[0].Why`를 직접 세워 잡는다.
- `internal/candidate.assessInto` B2 — `Assess` 실패 전파.

각 항목은 해당 `analysis/function-logic/*/branch-test-map.md`에 **커버 없음**으로 표기되어
있다.
