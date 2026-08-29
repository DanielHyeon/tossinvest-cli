# Branch Test Map: `Notifier.Flush`

Source: `internal/obs/notifier.go` (727-826). AST 기준 branches **19** / returns 4.

**프로덕션 호출자는 여전히 0이다** — 실측(2026-08-12):
`grep -rn '\.Flush(' internal/ cmd/ --include='*.go' | grep -v _test.go` → 전부
`bufio.Writer`·`csv.Writer`·`http.Flusher`다. `Notifier.Flush`는 **한 건도 없다.**
테스트 호출자는 **여섯**이다(1판 셋 + a099가 더한 셋).

## 커버리지는 주장이 아니라 **측정값**이다

아래 표의 「본문 실행」 칸은 눈으로 고른 것이 아니라
`go test ./internal/obs/ -count=1 -coverprofile`(exit 0 · **81건 통과**)의
블록 카운트를 `notifier.go:727-826` 범위로 잘라 읽은 것이다.
`count=0`이면 미실행, `count>=1`이면 실행이다.

| Branch | 위치 | 조건 평가 | **본문 실행** | 근거 블록 |
|---|---|---|---|---|
| B1 | `:728` `n.Journal == nil` | yes | **no** | `728.22-730.3` count=0 |
| B2 | `:738` `PendingAlerts` 오류 | yes | **no** | `738.16-740.3` count=0 |
| B3 | `:741` `range pending` | yes | **yes** | `741.2-741.32` count=1 |
| B4 | `:742` `n.Publisher == nil` | yes | **no** | `742.25-743.9` count=0 |
| B5 | `:753` `ClaimAlertByID` 오류 | yes | **no** | `753.18-759.20` count=0 |
| B6 | `:759` `n.Log != nil` (B5 안) | **no — 도달 자체가 없다** | **no** | `759.20-761.5` count=0 |
| B7 | `:764` `Disposition != Acquired` | yes | **yes** | `764.49-767.55` count=1 |
| B8 | `:767` `== ClaimHeldElsewhere` | yes | **yes** | `767.55-769.5` count=1 |
| B9 | `:779` `Publish` 오류 | yes | **yes** | `779.57-782.19` count=1 |
| B10 | `:782` `MarkAlertAttemptFailed` 오류 | yes | **no** | `782.19-783.21` count=0 |
| B11 | `:786` (B10의 `else`) | yes | — | `786.10-786.54` count=1 |
| B12 | `:783` `n.Log != nil` (B10 안) | **no — 도달 자체가 없다** | **no** | `783.21-785.6` count=0 |
| B13 | `:786` `failed.Outcome != Applied` | yes | **no** | `786.54-788.5` count=0 |
| B14 | `:803` `ReleaseAlertClaim` 오류 | yes | **no** | `803.19-804.21` count=0 |
| B15 | `:807` (B14의 `else`) | yes | — | `807.10-807.56` count=1 |
| B16 | `:804` `n.Log != nil` (B14 안) | **no — 도달 자체가 없다** | **no** | `804.21-806.6` count=0 |
| B17 | `:807` `released.Outcome != Applied` | yes | **no** | `807.56-809.5` count=0 |
| B18 | `:813` `MarkAlertDelivered` 오류 | yes | **no** | `813.18-815.4` count=0 |
| B19 | `:816` `settled.Outcome != Applied` | yes | **no** | `816.47-820.12` count=0 |

**본문이 실행되는 분기는 넷이다 — B3 · B7 · B8 · B9.** 나머지 열다섯 중 셋
(B6 · B12 · B16)은 **조건조차 평가되지 않는다**: 셋 다 한 번도 안 도는 오류 팔 안의
`n.Log != nil` 가드이기 때문이다.

### ⚠ 이 측정이 찾은 것 하나 — a098의 것이 아니므로 이름을 붙여 넘긴다

a099가 이 함수에 더한 **정산·반납 실패 경로**(B10 · B12 · B13 · B14 · B16 · B17 · B18 · B19)는
**`Flush`를 통해서는 한 번도 실행되지 않는다.** a099의 테스트는 같은 규범을
`deliver` 쪽에서 잰다. 즉 *"두 발송 경로가 같은 규범을 진다"*는 주장 중
**`Flush` 쪽 절반이 이 함수에서는 미측정**이다.

**a098은 이 함수를 안 건드리므로 이것을 지지 않는다.** 그러나 **조용히 넘기지도 않는다** —
tasks의 「안 하는 것」 표에 이름을 붙인다. 고치는 자리는 `internal/obs`이고
그것은 a092의 표면이다(design D3·D4).

## a098이 이 표에서 지는 RED는 **없다** — 그리고 왜 없는지가 중요하다

> **⛔ 1판은 여기에 *"tasks §3의 R1~R5가 이 표에서 나온다"*고 적었다. 그 문장은 죽었다.**
>
> 두 가지가 동시에 무효로 만들었다. ① 사용자 결정(2026-08-10, 안 1)이 **`Flush`를
> 안 부르는 설계 D**를 고정했다 — 이 함수의 분기는 a098의 실행 경로가 아니다.
> ② tasks §3의 R 등록부가 5판·6판·7판을 거치며 **다른 성질로 다시 세워졌다** —
> 오늘의 R1은 *"기록된 critical 알림은 발송된다"*이고 R5는 *"운영자가 밀린 것을 읽고
> 승인으로 게이트를 푼다"*이다. **번호는 같고 가리키는 것이 다르다.**
>
> 번호를 그대로 두고 뜻만 바꾼 문서를 근거로 쓰면 **읽는 사람이 반증할 수 없다.**

**대신 a098이 지는 것은 이것이다.** a098의 배달 실행자(tasks 4.0)는 이 함수를
안 부르고 **같은 규범을 자기 경로에 다시 구현한다.** 그러므로 위 표에서 a098로
건너오는 것은 분기가 아니라 **규범**이고, 그 규범을 지는 R은 이미 §3에 있다.

| `Flush`의 규범 | a098 루프의 대응 | 지는 R |
|---|---|---|
| B7 `:764` — `Acquired`가 아니면 건너뛴다 | 같음 (4.0의 `HeldElsewhere → 건너뛴다`) | **R19** (동시 발송이 한 번만) |
| B8 `:767` — held를 **기록한다** | 같음 — 조용히 안 넘긴다 | **task 0.2 3번** |
| B9 `:779` → `:800` — publish 실패 뒤 **반납한다** | 같음 (4.0의 `포기할 때만 Release`) | **R11**의 대조군 |
| B19 `:816` — 정산 실패 뒤 **해제하지 않는다** | 같음 (4.0의 ⛔) | **R11** |
| `:734` — 잠금을 **전송 위에서** 쥔다 | **반대로 한다** — 어떤 뮤텍스도 안 쥔다 | **R4b** |

마지막 줄이 이 번들이 a098에 있는 이유 전부다(design D1.1).

## 산출물 근거

- 분기·이탈 열거: `ast.json` (branches **19**, returns 4) — `go run ./tools/logic-map`
- 커버리지: `go test ./internal/obs/ -count=1 -coverprofile` exit 0 · 81건 통과
- `Flush` 테스트 호출 지점 전수: `grep -rn '\.Flush(' internal/obs/ --include='*_test.go'` → **6건**
  (`a097_exclusion_is_an_event_test.go:109` · `a099_claim_excludes_the_second_sender_test.go:210` ·
  `a099_round4_test.go:288`·`:311` · `obs_test.go:440`·`:590`)
- 프로덕션 호출자 0건: 위 grep의 non-test 매치는 전부 `bufio`·`csv`·`http.Flusher`
