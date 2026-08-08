# Branch Test Map: `Journal.ClaimAlertForDelivery`

측정: `go test -covermode=set -coverprofile ./internal/journal/`.
**RED 75.0% → GREEN 75.0%** — 총계가 같다. a097이 이 패키지에 새 분기를 만들지 않았기
때문이며, 그 사실 자체가 이 change가 재무장 UPDATE의 SET 목록만 바꿨다는 증거다.

`-covermode=set`은 실행 여부만 0/1로 기록하고 **횟수를 세지 않는다.**
줄번호는 GREEN 시점(구현 후 재생성한 `ast.json`) 기준이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:173` event key가 비었다 | a096 기존 (키 검증) | 진입 | 진입 (`173-175 count=1`) |
| B2 | `:176` event type이 비었다 | a096 기존 | 진입 | 진입 (`176 count=1`) |
| B3 | `:183` `BeginTx` 실패 | 없음 — 장애 주입 (`not-applicable`) | 미진입 | **미진입** (`183-185 count=0`) |
| B4 | `:194` 조회 결과 분류 `switch` | B5·B8이 대표 | 진입 | 진입 (`186-194 count=1`) |
| B5 | `:195` 기존 행을 찾았다 | a096 다수 | 진입 | 진입 (`195-197 count=1`) |
| B6 | `:197` **재무장한다** | a096 기존 + **a097 2.1·2.2·2.3** | 진입 | 진입 (`197-236 count=1`) |
| B7 | `:229` 재무장 UPDATE 실패 | 없음 — 장애 주입 (`not-applicable`) | 같은 블록이라 분리 불가 | 같은 블록(`197-236`)이라 분리 불가 |
| B8 | `:241` 조회가 `ErrNoRows`가 아닌 오류 | 없음 (`not-applicable`) | 미진입 | 미진입 (`241-242 count=0`) |
| B9 | `:249` INSERT 실패 | 없음 (`not-applicable`) | 미진입 | 미진입 (`249-251 count=0`) |
| B10 | `:253` `LastInsertId` 실패 | 없음 (`not-applicable`) | 미진입 | 미진입 (`253-255 count=0`) |
| B11 | `:256` 신규 행 커밋 실패 | 없음 (`not-applicable`) | 조건만 진입 | 조건만 진입, 오류 본문 미진입 |

## 커버리지는 이 change를 증명하지 못한다 — 단언이 증명한다

B6은 RED에서도 GREEN에서도 진입한다. a096이 재무장 자체를 이미 테스트했기 때문이다.
**진입 여부가 변경 전후로 똑같으므로 커버리지는 아무 말도 하지 않는다.**

증명한 것은 세 단언이고, 셋 다 구현 전에 실패했다(실행 로그 인용은 review.md §5):

- `title`·`body`·`payload`가 이전 에피소드의 값이었다 (`TestReArmingCarriesTheCauseThatReArmedIt`)
- `attempts = 3`이 남아 있었다 (`TestReArmingResetsTheAttemptCount`)
- `delivered_at`·`last_attempt_at`이 이전 에피소드 값을 들고 있었다
  (`TestReArmingClearsThePreviousEpisodeTimestamps`)

## 예측 하나가 틀렸다 — 기록한다

이 표의 초판은 "a097 2.4(obs의 claim 실패 테스트)가 닫힌 DB로 `B3@183`을 지나므로 GREEN에서
0→1로 바뀔 것"이라고 적었다. **틀렸다. GREEN에서도 0이다.**

이유는 Go 커버리지가 **테스트 대상 패키지 기준**으로 집계되기 때문이다.
`./internal/obs/`를 돌릴 때 실행되는 `internal/journal` 코드는 journal 패키지의 프로파일에
기록되지 않는다. 예측은 실행 사실과 집계 범위를 혼동했다.

이 칸을 "진입"으로 적고 넘어갔다면 그것이 바로 a096 3판이 지적한 **미측정 수치**가 됐을
것이다. 그래서 측정값을 그대로 두고 예측이 틀렸다고 적는다.
