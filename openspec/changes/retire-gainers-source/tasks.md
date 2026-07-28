# tasks

> **범위**: 답할 수 없는 원천 하나를 패널·일정에서 내리고, 그 형태를 요구사항으로 금지하고,
> **스냅샷이 이미 알고 있던 사실을 테스트가 읽게 만든다**(D8).
> `duration`을 고치지 않는다. 후퇴 사다리를 고치지 않는다(D4). 원천을 더하지 않는다.
>
> **순서**: 이 change가 **먼저** 랜딩한다. `refine-extended-shadow-bands`와 파일 표면이
> 겹치므로(`cmd/tossctl/candidate.go`) 두 change는 병행하지 않는다.
>
> **위험 등급: Normal.** 주문·손절·사이징·원장·인증·체결에 닿지 않는다(`isolation_test.go`).
> 그리고 이 change는 **호출을 하나도 만들지 않는다** — 순회당 RANKING 호출이 3에서 2로
> 줄기만 한다(D10). 429 후퇴 로직에 손대면 등급이 올라가므로 손대지 않는다.
>
> **FLM 대상 규칙**: `check_analysis.py`는 **diff hunk와 줄 범위가 겹치는 함수**만 대상으로
> 잡는다(파일 전체가 아니다). 새 파일의 함수는 전부 면제이고, 기존 파일에서 대상이 되는
> 것은 **손댄 함수**뿐이다(D7). §2가 기존 테스트 파일 하나를 반드시 고치는데, 그 파일에서
> 대상이 되는 것은 고친 함수 하나다.
>
> **새 테스트는 새 파일에 쓴다** — 이유는 FLM 범위가 아니라 diff를 좁게 유지하는 것이다(D2).
> 예외는 §2뿐이고, 그것은 **기존 테스트가 깨지기 때문**이지 선택이 아니다.
>
> **주석 함정**: `panelsize_drift_test.go`가 `internal/candidate`와 `internal/candidatesrc`의
> **비테스트** `.go` 주석을 전부 훑으면서 `숫자 + rows/행/개` 형태의 주장을 잡는다
> ([panelsize_drift_test.go:109, 174-232]). 선언하는 행 수는 `{100, 30}`뿐이므로
> 새 주석에 `"2 rows"`·`"3행"` 같은 표현을 쓰면 **무관한 이유로 RED**가 된다.
> 원천 **개수**는 rows가 아니므로 걸리지 않지만, 숫자와 단위를 붙여 쓰지 않는 편이 안전하다.

## 1. 배선 제거

- [x] 1.1 `candidatesrc.Panel`의 랭킹 타입 리터럴([candidatesrc.go:560])에서
  `RankingTopGainers`를 뺀다. 주석에 **왜** 빠졌는지와 재검토 조건(proposal "재검토 조건"
  4개)을 적는다. 리터럴 옆의 기존 주석 — 거래대금이 먼저인 이유 — 은 유지한다.
- [x] 1.2 `candidateIntervals()`([candidate.go:366])에서 `candidate.SourceOfficialGainers`
  항목을 뺀다. 1.1과 **같은 커밋**이어야 한다. 한쪽만 빼는 것이 이 change에서 가장 하기 쉬운
  실수이고(D5), 아무것도 실패시키지 않는다.
- [x] 1.3 `SourceOfficialGainers`·`RankingTopGainers`·`rankingSourceID` 항목은 **남긴다**
  (D2). `candidate.go`의 "정의상 이미 일어난 움직임의 목록" 논거도 남긴다. 지우지 않았음을
  구현 보고에 명시한다.
- [x] 1.4 `SourceOfficialGainers`가 더 이상 패널에 없다는 사실을 `candidate.go`의 상수
  주석에 한 줄로 적는다 — 상수만 읽는 사람이 그것을 살아 있는 원천으로 오해하지 않도록.

## 2. 깨지는 기존 테스트를 고친다 (이것이 첫 RED다)

- [x] 2.1 [T] `clock_wiring_test.go:63-67`의 `len(sources) != 4`를 **id 집합 단언**으로
  바꾼다(D9). **`4`를 `3`으로 바꾸지 않는다** — 길이 단언은 D5가 쓰지 말라고 한 바로 그
  형태이고, WTS 유무로 무관하게 깨지면서 정작 gainers가 되돌아오는 것은 못 잡는다.
  단언은 ① KR에서 `sources`의 id 집합이 기대 집합(거래대금·거래량·WTS 인기)과 **정확히**
  같다, ② 그 집합이 비어 있지 않다. 아래 `for range sources` 루프가 빈 슬라이스 위에서
  아무것도 단언하지 않고 통과하는 것을 막는 것이 원래 목적이고, 집합 단언이 그것을 더
  강하게 한다.
- [x] 2.2 같은 파일의 **주석과 실패 메시지에서 개수를 뺀다**. 헤더
  [clock_wiring_test.go:10-11]과 실패 메시지 [clock_wiring_test.go:101]이 *"Panel builds
  four of them"*이라고 쓰고 있다. **넷을 셋으로 바꾸지 말고 개수를 말하지 않게 고친다** —
  "Panel이 만드는 모든 원천"이면 이 파일이 실제로 지키는 성질을 그대로 말한다.
  개수를 적으면 다음 원천 변경에서 또 표류한다. 이 표류에는 가드가 없다(D9).
- [x] 2.3 [T] `go test ./internal/candidatesrc/...`를 **1.1 적용 후 2.1 적용 전**에 한 번
  돌려서 이 테스트가 실제로 RED임을 확인하고, 출력을 구현 보고에 남긴다. 초안이 이 테스트를
  놓쳤다는 사실이 이 task의 존재 이유다.

## 3. 배선을 고정하는 테스트 (새 파일, RED 먼저)

- [x] 3.1 [T] **패널에 gainers가 없다.** RED: 1.1 이전 상태에서 실패. KR·US 양쪽에서
  `candidatesrc.Panel(...)`이 만드는 id 집합에 `SourceOfficialGainers`가 없음을 단언한다.
  길이가 아니라 **id**를 본다(D5).
  **변이 확인**: `Panel`에 `RankingTopGainers`를 되돌리면 RED.
- [x] 3.2 [T] **일정에만 남은 원천이 없다.** `candidateIntervals()`의 키 집합이
  **`candidatesrc.Panel`이 만들 수 있는 id 집합의 부분집합**임을 단언한다.
  대조 대상은 `buildCandidatePanel`이 **아니다** — 그것은 `official.LoadCredentials`와
  `official.New`를 부르므로([candidatepanel.go:57-67]) 자격증명 없이는 호출할 수 없다.
  `candidatesrc.Panel`은 순수 함수이고 `cmd/tossctl`이 이미 import한다.
  최대 집합은 KR + 세 reader 전부 non-nil로 만든다(WTS는 KR 전용이므로 US는 부분집합이다).
  **방향은 한쪽뿐이다**(spec delta): 패널에 있고 일정에 없는 원천은 위반이 아니라
  `unconfiguredFloor` 15초가 적용되는 설계된 경우다([source.go:196, 248-249]).
  등호로 단언하면 그 설계를 테스트가 금지한다.
  **변이 확인**: ① 일정에만 gainers를 남기면 RED. ② 일정에서만 빼고 패널에 두면 GREEN
  (위반이 아니다). ③ 둘 다 빼면 GREEN.
- [x] 3.3 [T] 기존 `TestEveryPanelSourceHasItsOwnID`가 계속 통과하는지 확인한다(수정하지
  않는다). id 유일성과 비어 있지 않음만 보므로 원소가 줄어도 유효하다.
- [x] 3.4 [T] `panelsize_drift_test.go`가 계속 통과하는지 확인한다(수정하지 않는다).
  AST에서 `OfficialRanking(..., 100)`·`WTSPopular(..., 30)`의 **행 수**를 읽으므로 값 집합은
  `{100, 30}` 그대로다. 통과하지 않으면 그 사실이 이 change가 몰랐던 결합이므로 issues.md에
  기록하고 Manager를 부른다.

  > **정정 2026-07-29 (Manager 착오, 구현자가 잡음).** 이 항목의 초안은 근거를 *"호출이 세
  > 번에서 두 번으로 줄어도"*라고 썼다. **틀렸다.** `OfficialRanking` 호출은 소스 텍스트에
  > **한 번** 나온다 — `for` 루프 **안**이다([candidatesrc.go:593]). `declaredPanelSizes`는
  > AST의 `CallExpr`을 세므로 언제나 하나를 본다. 세 번이었던 적이 없다.
  > 줄어드는 것은 소스의 호출 개수가 아니라 **런타임 순회 횟수**다. 결론(`{100, 30}` 유지)은
  > 맞았고 근거만 틀렸다 — 그리고 근거가 틀린 채로 결론이 맞으면, 다음번에 그 근거를
  > 재사용하는 사람이 틀린다.

## 4. 스냅샷 표류 테스트 — 이 change의 핵심 산출물 (D8)

> 이 결함은 `docs/migration/openapi.latest.json`에 **46일 먼저** 적혀 있었고 참조되지
> 않았다. 배선만 고치면 다음번에 같은 일이 난다. 문자열 검색이 아니라 **두 집합의 대조**다.

- [x] 4.1 [T] **금지 집합을 스냅샷에서 읽는다.** JSON을 파싱해 랭킹 `type` enum 설명
  ([openapi.latest.json:7938])에서 `` `TOP_GAINERS`: … — `realtime` 미지원 `` 형태를 읽어
  `{TOP_GAINERS, TOP_LOSERS}`를 만든다. 같은 사실이 엔드포인트 설명(line 7934)과 오류
  예시(line 8171)에도 다른 문장으로 있다 — **어느 것을 읽었는지와 왜 그것인지 주석에 적는다.**
- [x] 4.2 [T] **배선 집합을 코드에서 읽는다.** `candidatesrc.go`를 AST로 파싱해
  ① `Panel`의 랭킹 타입 리터럴이 지목하는 상수들의 **값**([candidatesrc.go:145-147] —
  `RankingTopGainers = "TOP_GAINERS"`이므로 상수 값이 곧 API 문자열이고 매핑이 필요 없다),
  ② `Rankings(...)` 호출에 넘기는 duration 문자열 리터럴([candidatesrc.go:247])을 꺼낸다.
- [x] 4.3 [T] **단언**: 배선 집합 ∩ (그 duration의 금지 집합) = ∅.
  **금지 집합이 비면 Fatal한다** — `declaredPanelSizes`가 `len(out) == 0`에서 Fatal하는
  것과 같은 이유([panelsize_drift_test.go:76-80]). 파싱이 조용히 실패하면 이 테스트는
  어떤 배선에도 통과하는 장식이 된다.
- [x] 4.4 [T] **변이 세 가지를 확인하고 결과를 적는다**: ① `Panel`에 `RankingTopGainers`
  되돌리면 RED. ② 스냅샷의 `realtime 미지원` 문장을 지우면 **RED**(빈 금지 집합 Fatal).
  ③ duration을 `1d`로 바꾸면 **GREEN**(금지되지 않은 조합). ③이 이 테스트가 "gainers 금지"가
  아니라 **"형태 금지"**임을 보이는 변이다 — D3가 요구사항 수준에서 말한 것을 코드가 지킨다.
- [x] 4.5 파일을 파싱하는 것은 import가 아니므로 `isolation_test.go`의 금지표에 걸리지
  않는다. 선례: `panelsize_drift_test.go`(candidatesrc 소스를 텍스트로),
  `fsguard_drift_test.go`(원장 허용목록). 어느 패키지에 둘지 정하고 이유를 적는다.

## 5. 안전 근거를 고정한다

> proposal의 안전 근거는 두 불변식이 **함께** 성립해야 한다. 조건이 하나 붙어 있고,
> 초안은 그것을 빼고 썼다.

- [x] 5.1 **측정**: 저장소에 `sources`가 `official_rankings_top_gainers`를 포함하는 후보가
  0건임을 읽기 전용으로 확인한다. **한 건이라도 있으면 이 change는 그대로 진행하지 않는다** —
  그 후보의 supporter가 gainers뿐이면 `coverageAnswered`가 `present == 0`으로 false를
  반환하고 `coolAbsent`가 건너뛰어([scan.go:685-687, 704-716]) **영영 냉각되지 않는다.**
  결과를 review.md에 남긴다. 발견되면 Manager를 부른다.
- [x] 5.2 [T] **살아 있는 supporter가 하나라도 있으면 냉각은 정상 동작한다.** 후보의
  `Sources`에 패널에서 빠진 id와 살아 있는 id가 함께 있고, 살아 있는 쪽이 응답했고 그
  후보를 나열하지 않았을 때 **냉각된다**를 단언한다. 빠진 id는 `heard`에 없으므로
  건너뛰어진다.
  **변이 확인**: `coverageAnswered`가 `heard`에 없는 supporter를 건너뛰지 않게 바꾸면 RED.
- [x] 5.3 [T] **그리고 정직한 반대쪽**: supporter가 빠진 id **하나뿐**인 후보는 냉각되지
  않음을 단언한다. 이것은 바라는 동작이 아니라 **현재 동작**이고, 5.1이 그런 후보가
  없음을 측정으로 보장하는 이유다. 테스트 주석에 그 관계를 적는다 —
  "이 테스트가 RED가 되면 좋은 소식이고, 그때 5.3과 issues.md 항목을 지운다."

## 6. 기록

- [x] 6.1 **표류한 주석을 찾아 고친다.** 원천이 셋이라고 말하는 주석이 최소 세 곳 있다:
  [candidatesrc.go:556-559](*"the three types are compile-time constants"*),
  [notdue_test.go:26-27]과 [notdue_test.go:230-232](*"exactly three official sources"*).
  `notdue_test.go`의 두 곳은 **US 패널**을 말하는데, US는 이 change 이후 **둘**이 된다.
  주석을 고치되 **가능하면 개수를 말하지 않는 문장으로** 바꾼다(D9와 같은 이유).
  `notdue_test.go`를 고치면 그 파일에서 **고친 함수만** FLM 대상이 된다.
  더 있는지 `grep -rn "three official\|세 랭킹\|three rankings" --include="*.go" .`로 훑는다.
- [x] 6.2 `docs/baseline.md`(또는 발굴 원천을 세는 문서가 있으면 그것)에서 원천 수를
  갱신한다. 갱신할 곳이 없으면 없다고 보고한다 — 지어내지 않는다.
- [x] 6.3 issues.md에 D4의 남긴 관찰을 기록한다: 되찾을 수 없는 4xx가 후퇴 사다리에
  걸리지 않는다. `Missing`의 분류 비트는 `RateLimited` 하나뿐이므로([scan.go:59-62])
  이것을 고치는 것은 오류를 새로 분류하는 일이고, `off.Note`가 `Cycle` 안에 있으므로
  ([watch.go:653]) **그 change는 High-risk다.** 등급까지 적는다.
- [x] 6.4 issues.md: 원천 **개수**를 주장하는 주석에는 가드가 없다.
  `panelsize_drift_test.go`는 `숫자 + rows/행` 형태만 잡고, `"four sources"`는 숫자가
  단어라서 그 파일의 `TestWhatThisGuardCannotCatch`가 이미 못 잡는다고 적어 둔 경우다.
  이 change에서 만들지 않는다(범위 밖).
- [x] 6.5 issues.md: `1d` gainers가 장중에 갱신되는지 일 배치인지 확인하지 못했다
  (마감 후 관측이라 세 랭킹의 `ranked_at`이 모두 같았다). 재검토 조건 3번의 내용이다.

## 7. 검증

- [x] 7.1 `go test ./...`, `go vet ./...`, `gofmt -l .`, `go test -race ./internal/candidate/...
  ./internal/candidatesrc/... ./cmd/tossctl/...` 전부 통과. upstream 상속 테스트 회귀 없음.
- [x] 7.2 Function Logic Map: `check_analysis.py --change retire-gainers-source`가 대상으로
  잡는 함수 전부에 산출물을 만들거나, 대상이 0이면 `Function Logic Map: not-applicable`과
  그 근거를 review.md에 남긴다. **면제를 먼저 주장하지 않는다** — 스크립트를 돌리고 그
  출력으로 판단한다(D7). §2.1이 고치는 함수는 거의 확실히 대상이다.
- [x] 7.3 **장중 실측 1회** (D6). `./bin/tossctl candidate scan --market KR`(또는 US 장중이면
  `--market US`). 판정 기준은 **숫자가 아니라** 두 가지다:
  ① `attempted`와 `responded`가 같다 ② 그 줄에 `(degraded)`가 없다.
  기대 원천 수는 구성에 따라 다르다 — WTS 있는 KR은 3, WTS 없는 KR은 2, US는 2
  ([candidatesrc.go:566-570]). **숫자를 고정해서 읽지 않는다.**
  `(degraded)`가 남아 있으면 **이 change는 완료가 아니다** — 다른 원천도 죽어 있다는 뜻이고,
  그것이 이 change가 발견해야 할 사실이다. 읽기 전용이며 주문 side effect가 없다.
  결과 전문을 review.md에 남긴다.
- [x] 7.4 독립 리뷰: 구현과 **분리된 컨텍스트**가 diff와 테스트를 재검증하고 review.md에
  남긴다(날짜·보이스·발견·수용/거절과 근거). §4의 변이 세 가지는 **리뷰어가 직접 넣어
  확인한다** — 특히 변이 ③(duration을 `1d`로 바꾸면 GREEN)이 실제로 GREEN인지.

  > **2026-07-29 완료.** review.md "독립 리뷰 (§7.4)" 절. A~E 전부 리뷰어가 직접 변이를
  > 넣고 되돌려 확인했다(변이 ③은 gainers 복원 + `1d` 동시 적용에서 GREEN). 이 체크는
  > **리뷰를 수행하고 기록했다**는 뜻이지 랜딩 승인이 아니다 — 리뷰 판정은
  > **랜딩 불가**이고 사유는 F1·F2·F3다.
- [x] 7.5 PM 등록 확인: `docs/pm/portfolio/_registry.yaml`의 `bootstrap_change_allowlist`와
  `tools/pm/test_generate_master_tracker.py`의 fixture 튜플 **양쪽**에
  `"retire-gainers-source"`가 있는지 본다(Manager가 2026-07-28에 넣었다).
  `python3 tools/pm/generate_master_tracker.py --check`.
- [x] 7.6 `make sdd-sync` → `make sdd-check` → `make gate CHANGE=retire-gainers-source`.
  `sdd-sync`와 `gate` 사이에 **어떤 추적 파일이든** 편집이 있으면 fingerprint가 stale이
  되므로 작업을 전부 끝낸 뒤 연속으로 실행한다.

  > **정정 2026-07-29 (Manager 착오, 구현자가 잡음).** 초안은 *"`.go` 편집이 있으면"*이라고
  > 썼다. **`.md`도 무효화한다** — 구현자가 issues.md·FLM 산출물·review.md를 고친 뒤
  > `sdd-sync`→`sdd-check`→`gate` 전체를 다시 돌려야 했다. 기록 작업을 코드 뒤로 미룰
  > 이유가 여기 있다. (참고: `sdd-check`가 내는 `[context-graph] CCG adapter failed`는
  > WORKFLOW상 advisory이며 exit 0이다.)

  > **주의: gate PASS는 랜딩 승인이 아니다.** 2026-07-29 gate는 8단계 전부 통과했는데
  > 독립 리뷰의 P1 세 건은 **하나도 잡지 못했다** — F1은 산출물의 *내용*, F2는 *통과하는*
  > 가드 안의 구멍, F3은 **clean checkout에서만** 나타난다. §8이 닫히기 전에는 랜딩하지 않는다.

## 8. 독립 리뷰가 낸 P1 (2026-07-29) — 랜딩 차단

> 셋 다 Manager가 직접 재현했다. §7.6의 gate는 셋 다 통과시켰다.

- [x] 8.1 [T] **F1 — 버려지는 오류가 지키지 않는 가드에 귀속되어 있다.**
  `Panel`은 `OfficialRanking`의 오류를 버리면서([candidatesrc.go:591-594]) 주석으로
  *"TestEveryPanelSourceHasItsOwnID fails if one ever slips"*라고 쓴다. **거짓이다.**
  그 테스트는 **결과 패널**을 순회하며 id 중복과 비어 있음만 본다
  ([candidatesrc_test.go:296-311]) — 오류로 조용히 빠진 원천은 "원소가 하나 적고, id는
  여전히 유일하고, 비어 있지 않은" 패널을 만들 뿐이다. **아무것도 실패하지 않는다.**
  그리고 이 거짓 주장이 FLM 산출물로 번졌다 — `function-logic-map.md:21,32`와
  **`branch-test-map.md:7`의 "Test" 칸**. Branch Test Map의 "이 분기를 덮는 테스트" 칸이
  거짓이면 FLM 절차 전체가 형식이 된다.

  **권장 형태(구현자가 검증하고 더 나은 것이 있으면 바꾼다)**: AST가 읽은 `Panel`의 랭킹
  타입 집합과, `Panel`을 실제로 호출해 얻은 원천 id를 `rankingSourceID`로 되돌린 집합이
  **일치**함을 단언한다. 불일치 = 오류로 조용히 빠졌거나 리터럴이 시야 밖으로 옮겨졌다는 뜻이다.
  이 한 단언이 F1과 F4를 같이 닫는다.
  **변이 확인**: `rankingSourceID` 항목이 없는 실제 enum 값(예: `TOSS_SECURITIES_TRADING_AMOUNT`)을
  `Panel` 리터럴에 넣으면 RED여야 한다. 지금은 스위트 전체가 초록이다.

  > **정정 2026-07-29 ①(Manager 착오, 구현자가 잡음) — 변이 절차에 단계가 빠졌다.**
  > 위 지시를 **문자 그대로** 따라 리터럴에 **맨 문자열** `"TOSS_SECURITIES_TRADING_AMOUNT"`를
  > 넣으면 `wiredRankings`가 `*ast.BasicLit`에서 Fatal한다. RED가 나오긴 하지만 **틀린
  > 이유의 RED**이고, 그것을 보고 "결함이 없다"고 결론낼 수 있다. **먼저 상수를 선언하고**
  > 그 식별자를 리터럴에 넣어야 한다. 결론은 맞고 한 단계가 빠진 형태 — 이 change에서
  > 반복되는 바로 그 모양이다.
  >
  > **정정 2026-07-29 ② — "F1과 F4를 같이 닫는다"는 과장이다.**
  > 새 가드는 F4의 **false pass**(리터럴이 `Panel` 밖으로 나가 가드가 초록인 채 눈이 머는 것)를
  > 닫는다 — B2′ 행이 그 변이에서 스냅샷 가드가 **PASS**였음을 기록하고 있다.
  > 닫지 **못한** 것은 `wiredRankings`의 취약성이다: `Panel` 안에 랭킹과 무관한 `[]string`
  > 리터럴이 생기면(구현자 실측: `[]string{candidate.MarketKR}`) `*ast.SelectorExpr`에서
  > Fatal해 **두 테스트가 무관한 이유로 죽는다.**
  > **판정(Manager): 이대로 둔다.** 방향이 false failure이지 false pass가 아니고,
  > 읽을 수 없는 배선 앞에서 **시끄럽게 실패하는 것**이 이 가드가 지켜야 할 성질이다 —
  > 읽을 수 없는 리터럴을 건너뛰는 쪽이 가드가 눈머는 경로다. 경계는 issues.md I5에 있다.
- [x] 8.2 주석과 FLM 산출물 **세 곳**을 사실에 맞게 고친다. 8.1이 가드를 만들면 주석은
  그 가드를 가리키고, 만들지 않기로 하면 **"아무것도 지키지 않는다"고 적는다.**
  둘 중 하나여야 한다 — 지금 상태(지키지 않는 가드를 지킨다고 적음)는 안 된다.
- [x] 8.3 [T] **F2 — 핵심 가드의 완전성 검사가 "비어 있지 않음"뿐이다.**
  실파일에 대한 유일한 검사가 `len(forbidden) == 0`이고, positive control은 **하드코딩된
  사본**을 읽는다([snapshot_drift_test.go:345-350]). 그래서 스냅샷에서 `TOP_GAINERS`
  불릿의 `realtime` 절만 지우면(`TOP_LOSERS`는 남김) 금지 집합이 비지 않아 Fatal이 안 나고,
  **gainers + realtime 배선이 통과한다.** 리뷰어가 재현했다.
  **수정**: 실파일에서 읽은 `forbidden["realtime"]`가 `TOP_GAINERS`와 `TOP_LOSERS`를
  **둘 다** 담고 있음을 단언한다. 선례는 `TestTheSourcesStillDeclareTheSizesTheseCommentsClaim`이
  `{100, 30}`을 실파일에 대해 고정하는 것과 같다.
  **변이 확인**: `TOP_GAINERS` 불릿의 `realtime` 절만 지우면 RED. **반드시 되돌리고
  `git diff docs/`가 비었는지 확인한다.**
- [x] 8.4 **F3 — 커밋 절차.** `_registry.yaml`과 PM fixture에 `refine-extended-shadow-bands`가
  들어 있는데 그 change 디렉터리는 이 커밋에 들어가지 않는다. 그대로 커밋하면
  **clean checkout에서 `stale bootstrap exception`으로 `make sdd-check`가 깨진다**
  (Manager가 bands 디렉터리를 뺀 트리 사본으로 재현했다).
  체커가 **대칭**이라 단순히 등록을 지우는 것으로는 안 된다 — 디렉터리가 트리에 있는데
  등록이 없으면 `active change has no story or bootstrap exception`으로 반대편이 깨진다.
  그래서 **작업 트리는 양쪽 등록을 유지**하고, **커밋 시점에만** 다음을 한다:
  1. `_registry.yaml`과 PM fixture에서 `refine-extended-shadow-bands` 항목을 뺀다
  2. 코드 + `openspec/changes/retire-gainers-source/` + 그 두 파일을 stage하고 커밋한다
     (bands 디렉터리는 untracked이므로 커밋에 들어가지 않는다)
  3. 항목을 다시 넣는다 — 그 편집은 bands 커밋과 함께 간다
  커밋 내용이 clean checkout에서 정상임은 이미 확인했다(오류 0건).

  > **이 체크는 이 커밋이 수행하는 절차다.** task가 곧 커밋이므로, 커밋 직전에 체크하고
  > 그 상태로 게이트를 통과시킨 뒤 1~3을 실행한다. 게이트는 **양쪽 등록이 살아 있는
  > 일관된 트리**에서 돌려야 한다 — 분리된 상태에서는 `sdd-check`가 반대편으로 깨진다.
- [x] 8.5 F4~F9(비차단)를 issues.md에 옮긴다. 특히 **F7** — *"관측은 시스템이 알아보는
  것으로 되읽혀야 한다"*가 강제되지 않는다(`decodeSources`가 아무것도 검증하지 않는다).
  이 change에서 **세 번째**로 나온 "결론은 맞고 근거는 틀림" 형태다.
- [x] 8.6 [T] 재검증: `go test ./...`, `go vet ./...`, `gofmt -l .`, `-race`.
  그 다음 `make sdd-sync` → `make sdd-check` → `make gate CHANGE=retire-gainers-source`.
- [x] 8.7 §8 수정분에 대한 **2차 독립 리뷰**. 범위는 §8이 만든 diff로 한정한다.
  8.1의 새 가드에 리뷰어가 직접 변이를 넣어 확인한다.

  > **2026-07-29 완료.** review.md "2차 독립 리뷰 (§8.7)" 절. 판정 **§8 수용 — 랜딩 가능
  > (8.4만 남음)**. 변이 6종을 직접 넣고 되돌렸다. 결정적 실측 둘:
  > ① 새 가드를 실행에서 빼면 변이 A에서 패키지 **53건 전부 통과** — 결함이 실재했고
  > 나머지 53건 중 아무도 잡지 못한다는 직접 증거.
  > ② 변이 B-sharp(리터럴을 밖으로 + gainers 실제 복원)에서 스냅샷 가드는 **진짜 금지
  > 조합을 보고도 통과**했고 새 가드만 지목했다 — F4 ①이 닫혔다는 말의 실제 내용.
  > P2 5건, 랜딩 차단 없음.
- [x] 8.8 2차 리뷰 P2 중 **이 change 자신의 실패 형태인 셋**을 고친다(Manager, 2026-07-29).
  - **P2-2 `snapshot_drift_test.go:52-57`** — *"두 방향이 둘 다 조용히 통과한다"*가 틀렸다.
    리터럴을 밖으로 빼고 **다른 `[]string`을 안 남기면** 조용하지 않다 — 기존
    `len(out) == 0` Fatal이 두 가드를 함께 죽인다. 같은 파일 `:290`이 이미 그 사실을
    맞게 적고 있어 **파일이 자기모순**이었다. **이 change에서 네 번째 "결론 맞고 근거 틀림."**
    구체적 위험이 있어 고쳤다: 다음 사람이 그 Fatal을 "새 가드와 중복"이라며 지우면
    `declared`가 빌 때 **두 가드가 함께 눈머는 유일한 경로**가 열린다. 그 Fatal이
    load-bearing임을 주석에 명시했다.
  - **P2-1 `function-logic-map.md:69`** — *"KR 세 원천·US 두 원천"*을 무조건으로 적었다.
    KR이 셋인 것은 **WTS 세션이 있을 때뿐**이다. tasks §7.3이 실측 판정에서 경고한 그
    형태가 게이트 산출물 안에 남아 있었다.
  - **P2-5 `snapshot_drift_test.go:288`** — *"the snapshot guard **above** it"*이 실제로는
    130줄 아래다. 위치어 대신 이름으로 가리키게 고쳤다.
  - **P2-3·P2-4**는 issues.md로: I11(합집합이 시장별 배선을 못 본다) 신설, 그리고 파일
    머리의 자기 규칙을 정직하게 고쳤다 — I5·I6·I8이 규칙을 어긴 것이 아니라 **규칙이 과했다.**
