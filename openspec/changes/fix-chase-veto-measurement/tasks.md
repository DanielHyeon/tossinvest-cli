# Tasks: fix-chase-veto-measurement

각 `[T]`는 RED → GREEN → REFACTOR → VERIFY를 거친다. 이 change는 읽기 전용이고 주문
경로에 닿지 않는다 — 검증은 fixture와 주입 clock으로 한다.

**Teammate는 1명이다.** `internal/candidate` · `internal/candidatesrc` · `cmd/tossctl` ·
`internal/console` 넷을 함께 만지고, 병행 편집은 Function Logic Map의 base-commit 대조를
깨뜨린다.

**증거 비용을 미리 줄인다.** `check_analysis.py`의 diff 경로가 `"*.go"`라 `_test.go`가
포함되고, 수정된 파일에서는 **새 함수까지** FLM 대상이 된다. 반대로 **새로 추가된 파일의
함수는 전부 면제**된다(`old_source == ""`이면 즉시 반환). 그러므로:

> **새 테스트는 전부 새 `_test.go` 파일에 넣는다.** 기존 `_test.go`를 편집하는 것은
> §7의 *정정*뿐이고, *추가*는 절대 기존 파일에 하지 않는다.

**작업 순서**: 코드를 전부 끝낸 뒤에 FLM을 만든다. `ast.json`이 함수 소스의 SHA-256을
싣기 때문에 그 뒤의 편집 한 줄이 증거를 stale로 만든다.

## 1. 신규 진입 사실을 3-상태로 만들고 실제로 채운다 (design D2)

- [x] 1.1 [T] `candidate.Row`·`Reported`의 신규 진입 사실을 3-상태로. **`unknown`이
  zero value여야 한다** — bool로 두고 첫 읽기에 `false`를 넣는 것이 이 패키지가
  `VetoState`·`Sighting`·`Expansion`·`ShadowBand`에서 네 번 거부한 형태다.
  RED: 대입하지 않은 값이 "직전에 있었다"로 읽히면 실패.
- [x] 1.2 [T] `officialRanking`·`WTSPopular`가 직전 읽기의 심볼 집합을 기억하고 보고한다.
  **소스의 첫 읽기는 `unknown`**(직전 읽기가 없다). 시장별로 분리한다 — 한 소스
  인스턴스가 두 시장을 섬기면 KR의 직전 읽기가 US의 판단 근거가 된다.
  RED: 첫 읽기가 모든 심볼을 `yes`로 보고하면 실패.
- [x] 1.3 [T] 저장소 스키마 **가산·nullable**. 기존 행은 `unknown`이다 — 기록 시점에
  아무도 이 사실을 재지 않았고 그것이 정확히 사실이다. 전진 전용, drop·rename 금지.
- [x] 1.4 [T] `internal/console/signals.go:656`의 신규 진입 표시가 3-상태를 렌더한다.
  지금까지 **뜰 수 없었던** 표시이므로, 처음으로 뜨는 경로가 생긴다.

## 2. 미상인 신규 진입 위의 최초 순위는 쓰지 않는다 (design D3)

- [x] 2.1 [T] `MeasureFirstSighting`이 최초 순위를 채택하려면 그 순위가 나온 읽기의
  신규 진입 사실이 `unknown`이 아니어야 한다. 미상이면 **사유를 명명한 미측정**.
  새 `VetoUnmeasured` 상수. 기존 사유들과 구분되는 이름이어야 한다 — 이것은 저장소
  결손도 보존 만료도 아니고 **소스가 답을 갖고 있지 않았다**는 뜻이다.
- [x] 2.2 [T] RED: 세션 시작 시나리오. 소스의 첫 읽기로 패널 전체가 승격되면 그
  후보들의 `seen_late`는 전부 미측정이고, 통과로 집계되지 않는다.
  두 번째 스캔에서 새로 나타난 심볼은 `yes`를 받아 측정 가능해진다.
- [x] 2.3 [T] 냉각 만료 후 재승격도 확인한다 — 그때 소스는 직전 읽기를 갖고 있으므로
  `no`를 보고하고, `no`는 "이미 목록에 있었다"라서 최초 관측 순위로서 정직하다.
- [x] 2.4 [T] **비율을 세지 않는다는 것을 테스트가 말하게 한다.** "한 스캔이 패널의
  N%를 승격했으면 일괄로 본다"는 대안은 그 N이 또 출처 없는 숫자다 — 이 change가
  치료하는 병을 치료 과정에서 재발시킨다. 테스트 이름·주석에 남긴다.

## 3. 절단된 읽기는 목록이 아니다 (design D4)

- [x] 3.1 [T] 두 소스가 **요청 행 수**와 **도착 행 수**를 함께 보고한다
  (`officialRanking.count` = 100, `WTSPopular.size` = 30 — 이미 알고 있고 버리는 값).
- [x] 3.2 [T] 둘이 다르면 그 읽기의 순위 백분위는 사유를 명명한 미측정.
  RED: 100개를 요청해 3개가 도착한 읽기의 1위가 백분위 66.7로 측정되면 실패.
- [x] 3.3 [T] `percentileOf(1,1) = 0`(1행 목록의 1위가 최하위로 읽힘)이 이 경로로
  걸리는지 확인한다. 별도 하한을 만들지 않는다 — **지어낸 숫자가 아니라 소스가 선언한
  크기**가 기준이라는 것이 D4의 요점이다.
- [x] 3.4 [T] 저장소 스키마 가산·nullable(1.3과 같은 규칙). 기존 행은 절단 여부 미상.

## 4. 통과 건수의 경보를 tally 항등식으로 (design D5)

- [x] 4.1 [T] 모든 code에 대해 `Passed + Raised[code] + NotMeasured[code] <= Total`을
  검사한다. `VetoTally`만으로 계산되며 **새 배선이 없다** — 두 파생 지점 모두에 이미 있다.
- [x] 4.2 [T] `Reasons[THRESHOLD_ABSENT] > 0 && Passed > 0`은 직접 모순이므로 경보.
- [x] 4.3 [T] 판단은 한 곳에 두고 두 표면(`cmd/tossctl` 스캔 출력, `/signals`)이 문구만
  각자 렌더한다. 같은 규칙을 두 번 구현하지 않는다.
- [x] 4.4 [T] RED: 미측정 veto를 통과로 세는 조작된 tally를 넣으면 경보가 뜬다.
  `THRESHOLD_ABSENT`뿐 아니라 `NO_DAY_HIGH`·`INPUT_TOO_OLD` 등 **어떤 사유든** 잡혀야
  한다. 임계 전용 경보보다 넓다는 것이 이 결정의 이유다.
- [x] 4.5 [T] 기존 `passedNote`·`signalsPassedNote`는 **그대로 둔다.** 임계가 없는
  상태가 유지되므로 그 단언은 계속 참이고, 항등식 경보는 그 옆에 추가된다.

## 5. 소비자 가드 — 파일과 심볼 단위 (design D6)

- [x] 5.1 [T] `Chase`·`Passed()`·`Verdict`·`VetoTally`를 명명할 수 있는 **파일 목록**을
  고정한다. 현재: `cmd/tossctl/candidate.go`, `cmd/tossctl/console.go`,
  `internal/console/signals.go`(+ 각각의 `_test.go`), `internal/candidatesrc` 해당 파일.
  **먼저 `rg`로 실제 집합을 구해서 쓴다** — 초안의 목록은 서술에서 쓴 것이라
  `internal/candidatesrc`가 빠져 있었다.
- [x] 5.2 [T] 그 파일들이 `execgw.`·`orderintent.`·`trading.`·`official.Place`를 **함께**
  명명하지 않음을 교집합으로 검사한다. 패키지 단위가 열린 채 실패하는 두 경로를 막는다 —
  `internal/candidatesrc`가 `candidate`와 `official`을 둘 다 import하고,
  `cmd/tossctl`은 하나의 패키지라 새 파일이 import 간선을 0개 늘린다.
- [x] 5.3 [T] **양성 대조군.** 주문 동사를 명명하는 것이 확실한 파일을 가리켜 검사가
  실제로 발화하는지 확인한다. 없으면 가드가 조용히 아무것도 금지하지 않는다
  (선례: `isolation_test.go:313`).
- [x] 5.4 [T] 왜 지금 넣는지를 테스트 주석에 적는다 — 값이 들어간 뒤에 넣으면 그 사이에
  생긴 소비자가 목록의 초기값이 되어 **결정 없이 승인된다.**

## 6. 임계 단일 출처 (design D7)

- [x] 6.1 [T] `cmd/tossctl/candidate.go:364`와 `cmd/tossctl/console.go:898`의 두
  리터럴을 하나의 생성자로 모은다.
- [x] 6.2 [T] `cmd/tossctl` 생산 파일에 다른 `VetoThresholds{…}` 복합 리터럴이 없음을
  AST로 고정. 선례는 `candidate_review_test.go:193`.
  임계가 하나(`near_high`)뿐인 지금 하는 것이 싸다.

## 7. 150행 허구 정정 (design D9)

- [x] 7.1 `add-candidate-discovery` design.md D8의 예시를 제자리에서 날짜를 붙여 정정.
  선례는 같은 문서의 `near_high` 부호 정정. 정규화 자체는 유지한다 — WTS 30행과 공식
  100행이 한 저장소에 섞이므로 정규화의 필요는 실재하고, 근거가 된 예시만 틀렸다.
- [x] 7.2 코드 주석 정정. `rg -n '150' internal/candidate internal/candidatesrc`로
  전부 찾는다 — 최소 `metrics.go:797`, `metrics.go:890`, `store.go:181`,
  `store.go:1277`, `veto.go:481`, `veto.go:74`, `store.go:29`.
  `veto.go:481`은 백분위 정규화의 **존재 이유**를 적은 자리라 특히 중요하다.
- [x] 7.3 [T] 회귀 가드: 패널 크기를 주장하는 주석과 실제 소스 선언이 어긋나면 실패하는
  테스트. `fsguard_drift_test.go`가 파일을 소스로 읽어 대조하는 방식의 선례를 따른다.
  주석은 다시 낡는다 — 이번에는 낡으면 실패해야 한다.
- [x] 7.4 §7의 기존 테스트 정정이 있으면 여기서 한다. **각각에 대해 판단한다**:
  (a) 새 상태에서 여전히 무언가를 시험하는가, (b) 아니면 이제 잘못된 것을 옳다고
  단언하는가. (b)는 삭제로 처리하지 말고 무엇을 시험하도록 고쳤는지 커밋 메시지에 적는다.

## 8. 증거와 게이트

- [x] 8.1 코드 편집을 **전부 끝낸 뒤** Function Logic Map + Branch Test Map 생성.
  대상 집합은 예측하지 말고 `python3 tools/logic-map/check_analysis.py --change
  fix-chase-veto-measurement`가 보고하는 것으로 정의한다. 편집된 `Test*` 함수가 포함된다.
  **78개** 산출. 생산이 찾은 것은 issues.md I14(음수 요청 행 수의 두 거부에 테스트가
  없었다 — `negative_request_test.go` 추가)와 I15(미커버 분기 목록).
- [x] 8.2 `docs/pm/portfolio/_registry.yaml` allowlist **와**
  `tools/pm/test_generate_master_tracker.py`의 fixture 목록 **둘 다**에 change id 등록.
  한쪽만 하면 PM tracker 테스트가 깨진다(선례 있음).
- [ ] 8.3 `make sdd-sync` → `make sdd-check` → `make gate CHANGE=fix-chase-veto-measurement`.
  sync와 gate 사이에 `.go` 편집이 들어가면 CodeGraph fingerprint가 다시 stale이 된다 —
  커밋 → sync → gate를 끊기지 않게 한 번에 돌린다.
- [ ] 8.4 별도 문맥 리뷰(구현 후). `review.md`에 추가 기록.

## 9. 구현 후 독립 리뷰 2건의 지적 반영 (2026-07-28)

리뷰 둘 다 실행 가능한 probe와 mutation으로 확인한 지적이다. 기록은 `issues.md` I3·I5–I13.

- [x] 9.1 [P0-F1] 기억된 직전 읽기에 **유효 조건**을 붙인다. 신선도 상한은
  `candidate.DefaultStalenessTTL`(= `BackoffLadder` 마지막 rung의 2배)이고 값을 새로
  고르지 않는다. 시계 역행도 거부. `OfficialRanking`·`WTSPopular`·`Panel`이
  `clock.Clock`을 받는다. issues.md I6.
- [x] 9.2 [P0-F2] 소스 자신의 비교로 **짧은 읽기는 기억을 교체하지 않는다.** 빈 읽기도
  같은 규칙. `TestAnEmptyReadingIsStillAReadingOfThisList`를 뒤집었다.
- [x] 9.3 [P0-F3] §3의 절단 사슬에 배선 테스트를 건다(`Cycle` + `Assess`).
  두 mutation이 이제 실패하는 것을 확인했다. issues.md I13.
- [x] 9.4 [P1-F4] `MeasureFirstSighting`이 `!Truncation.Known()`을 새 사유
  `REQUEST_UNRECORDED`로 거부한다. issues.md I3(뒤집음).
- [x] 9.5 [P1-F5] 소비자 가드 세 구멍(클라이언트 생성·import 게이트·alias/dot import).
  `official.New`를 금지하고 클라이언트 생성을 `cmd/tossctl/candidatepanel.go`로 옮겼다.
  issues.md I11.
- [x] 9.6 [P1-F6] 스키마 4 이전 후보는 재자격 부여되지 않는다 — 제안된 fill 대신 문장을
  고치고 양쪽을 테스트로 고정했다. issues.md I8.
- [x] 9.7 [P1-F7] 자격 부여할 수 없는 최초 순위는 저장하지 않고 `FirstRanksHeld`로 센다.
  issues.md I7.
- [x] 9.8 [P1-F8] 소스별 요청/도착 행 수와 최초 관측 census를 스캔 리포트와 `/signals`에
  올린다. issues.md I9.
- [x] 9.9 [P2-F9] `NEW_ENTRANT_UNKNOWN`의 대응 문장과 스키마 롤백 주장 정정 + 롤백 계획
  기록. issues.md I5·I7.
- [x] 9.10 [P2-F10] 패널 크기 drift 가드를 조이고 못 잡는 것을 테스트로 적었다.
  issues.md I10.
- [x] 9.11 [P2-F11] v3 fixture + 칼럼·CHECK 고정, 루프 길이 가드. issues.md I12.

## 후속 (이 change의 체크박스가 아니다)

- **`extended = 6` 투입** (사용자 결정 2026-07-28: "수리하고 후속 change").
  출처는 `internal/exitpolicy/ladder.go:138` `DefaultLadderPolicy`의 최종 목표 6.0% —
  `min_reward_risk × default_stop_pct`와 달리 **진입→목표 구간을 재는 값**이다.
  검증 상태는 `[미검증 — StockOS KOSPI 튜닝값]`을 물려받으므로 그대로 표시해야 한다.
- **`seen_late` 임계.** 이 change가 고친 측정 위에 쌓이는 실제 100행 분포에서 정한다.
  D8은 상한(< 88)만 주고 값을 정해주지 않는다. 격자는 값을 고르지 않는다(design D8).
  고를 때 밴드 `>=` 와 veto `>` 의 경계 오차를 알고 골라야 한다.
- **`archive` 순서.** `add-candidate-discovery`를 먼저 archive한 뒤에 이 change를
  archive한다. `openspec validate`도 `gate.sh`도 이 순서를 검사하지 않는다.
- **`Cycle` 옵션 블록의 fallback 유혹.** 후속 change가 상수를 만드는 순간
  `watch.go:522-533`에 두 줄을 더하는 것이 파일의 관용구가 되고, 그러면 모든 호출자의
  `THRESHOLD_ABSENT`가 도달 불가가 된다. 이 change의 §4 RED가 그때 잡는다.
