# a096 리뷰 기록

## §0. 독립 리뷰 상태

**1라운드: Codex (교차 모델 충족) — 2026-08-08 — 판정 FAIL.**
blocker 4건, concern 5건. 전문은 §3, 저자 검증은 §4, 2판의 조치는 §5.

**2라운드: Claude Sonnet (교차 모델) — 2026-08-09 — 판정 PASS.**
blocker 0건, concern 2건. 상세와 저자 조치는 §8에 기록한다.

a092부터 이어진 미충족 연쇄가 여기서 끊겼다 — a096이 교차 모델 리뷰를 받은 첫 change다.

아래 §1은 저자 자신의 검증이며 독립 리뷰가 아니다.

## §1. 저자 검증 — **1판 시점** (독립 아님)

아래 수치는 전부 1판 코드의 것이다. 2판의 수치는 §5 말미에 있다.

### 통과한 것

| 항목 | 결과 |
|---|---|
| `go build ./...` | 통과 |
| `go vet ./internal/journal/ ./internal/obs/ ./internal/execgw/` | 통과 |
| `go test ./internal/obs/` | 통과 (5.5s) |
| `go test ./internal/execgw/` | 통과 (40s) |
| `go test ./internal/journal/` | 통과 (598s) |
| `go test ./...` | 통과 — FAIL 패키지 없음 |
| `check_analysis.py --change a096` | evidence complete (6함수 26분기) |
| `openspec validate --strict` | valid |

### RED이 실제로 빨갰다

```text
--- FAIL: TestOneConditionIsOneSend
    sends = 3, want 1
--- FAIL: TestSuppressingTheSendKeepsTheRecord
    sends = 5, want 1
    log lines = 9, want 5
```

### 과잉 억제 반증이 RED에서도 통과했다

`TestAnUndeliveredConditionIsStillRetried`는 RED·GREEN 양쪽에서 초록이다. 이 테스트가
RED에서 빨갰다면 그것은 반증이 아니라 결함의 일부를 검증하는 테스트였다는 뜻이므로,
양쪽 초록이 요구 조건이다.

### 커버리지가 독립적으로 같은 것을 말한다

```text
RED   notifier.go:258.88,260.5 1 1
GREEN notifier.go:276.88,278.5 1 0
```

`deliver`의 `markErr != nil` 블록. 이미 전달된 행에 다시 전달 표시를 시도했을 때만
들어가며, 그 시점에 `Publish`는 이미 성공한 뒤다. 진입 → 미진입은 스위트 전체에서
불필요한 재전송이 0이 됐다는 뜻이다.

obs RED 84.8% → GREEN 84.5%, journal RED 74.9% → GREEN 74.9%.

### 측정하지 않은 값을 한 번 적었다가 되돌렸다

journal GREEN 커버리지의 첫 시도가 `go test` 기본 10분 타임아웃에 걸렸는데
(`FAIL … 600.025s`), 그 시점에 `ClaimAlertForDelivery`·`EnqueueAlert` BTM의 GREEN 열은
이미 채워져 있었다 — RED 값을 옮겨 적은 것이었다. 전부 `미측정`으로 되돌린 뒤 단독
재실행(310.3s)으로 실측을 채웠다.

타임아웃 자체는 병행 실행 경합이었다: 단독 310.3s < base 325.2s, 커버리지도 동일한 74.9%.
a096은 journal 스위트를 느리게 하지 않았다.

## §2. 스스로 발견해 고친 것

**게이트가 설계를 되돌렸다.** 1차 구현은 `EnqueueAlert`의 서명을 `(int64, string, error)`로
넓혔다. `check_analysis.py`가 증거 누락 8건을 보고했다 — `Gateway.parkAlert`와
`internal/journal/outbox_test.go`의 기존 테스트 7개. 2차 구현은 새 함수
`ClaimAlertForDelivery`를 만들고 `EnqueueAlert`는 서명 그대로 위임하게 해서, 기존
호출자를 한 줄도 건드리지 않는다. design D1에 판단 근거를 기록했다.

**`AcknowledgeAlert`의 전제를 잘못 알고 있었다.** 처음 쓴 테스트는 DELIVERED 행을
acknowledge하려 했고 `no such alert`로 실패했다. 실제로는 `WHERE state = 'PENDING'`이라
**전송된 적 없는 알림을 사람이 직접 푸는 경로**다. 이 사실이 `owed = (state ==
AlertPending)`이라는 단일 등식을 정당화한다 — DELIVERED와 ACKNOWLEDGED를 따로 열거하는
대신 PENDING 하나만 물으면 된다.

## §2bis. 1판이 리뷰어에게 요청했던 것 (기록)

1. **`owed`의 정의가 원장에 있는 것이 맞는가.** `state == AlertPending` 한 등식이고,
   호출자는 해석하지 않는다(design D1 말미). 새 alert 상태가 생기는 날 이 등식이
   조용히 틀려지는 경로가 있는지.
2. **되살아나는 조건**(design D4). DELIVERED 행은 영구히 재전송되지 않는다. 이것이
   `event_key` UNIQUE의 기존 성질과 정말 같은지, 아니면 a096이 새 한계를 만드는지.
3. **범위 밖으로 남긴 ①**(tasks 7.1). 장 운영 시간 게이트가 없어 거절되는 주문은 계속
   나간다. a096을 배포해도 그 API 부하는 그대로다 — 이 분리가 옳은지.
4. `TestNotifierIsConcurrencySafe`에 단언을 넣지 않고 남긴 판단(tasks 7.3).


## §3. 1라운드 — Codex, 판정 FAIL

### BLOCKER 1 — claim이 배타적이지 않다: 같은 key 동시 알림이 여전히 이중 전송된다

`ClaimAlertForDelivery`는 트랜잭션을 커밋하고 **연결을 놓은 뒤** 반환한다. `deliver`의
`n.mu`는 그 **뒤에** 잡힌다. 따라서:

1. Notify A가 PENDING을 읽고 `owed=true`로 나와 `Publish`에 들어가 블록
2. Notify B가 여전히 PENDING인 같은 행을 읽고 역시 `owed=true`
3. A가 발행하고 DELIVERED로 표시
4. B가 `n.mu`를 얻어 **다시 발행**하고 `MarkAlertDelivered`에서 실패

`TestNotifierIsConcurrencySafe`는 goroutine당 key가 하나라 같은 key 호출이 직렬이고,
그래서 이 경로를 덮지 못한다.

### BLOCKER 2 — a096은 a074의 진입 차단을 **덜** 일어나게 만든다. design D5는 거짓이다

전달에 성공한 행은 **transport가 나중에 죽어도** 영구히 억제된다. base에서는 다음 관측이
전송을 시도하고, 예산을 소진하고, gate를 걸고, `ENTRY_BLOCKED`로 승격했다. a096은
transport를 시험하기 **전에** 반환한다.

구체적 실패:

1. key K가 전달 실패해 `ENTRY_BLOCKED`가 걸린다
2. 운영자가 PENDING 행을 acknowledge하고 모드를 완화한다
3. 또 다른 transport 장애 중에 K가 재발한다
4. 행이 ACKNOWLEDGED라 `owed=false` — 전송도 승격도 일어나지 않는다

반대 방향도 있다: 전송이 in-flight인 동안 acknowledge해도 이미 얻은 `owed=true`는 취소되지
않아, 실패하는 in-flight 전송이 방금 푼 차단을 다시 걸 수 있다. 양쪽 다 테스트가 없다.

### BLOCKER 3 — 나중의 **다른** 발생이 영구히 삼켜진다. design D4의 "기존 한계"는 틀렸다

행 수명의 한계는 기존이었지만 **운영자에게 보이는 억제는 아니었다.** `ec29dc72`에서
UNIQUE key는 옛 행을 돌려줬지만 `notifyCritical`은 **여전히 발행했다.** a096이 처음으로
종결된 행이 모든 미래 전송을 억제하게 만든다.

그리고 `exit.proposal_refused`의 key에는 **거절 사유가 없다**(`type|position_id|action|level`).
주말 `order-hours-closed` 거절이 한 번 전달되고 나면, 같은 포지션·같은 단계의 **다른,
조치가 필요한** 거절이 조용히 억제된다.

판단: 장 운영 시간 게이트를 a096 안에서 구현할 필요는 없다. 그러나 a096은 **발생
lifecycle/재무장, 한도 있는 재알림 정책, 또는 거절 폭주를 멈추는 선행 change 중 하나
없이는 착지할 수 없다.** 영구 억제는 문서가 인정하는 범위보다 나쁘다.

### BLOCKER 4 — logic-map 증거가 현재 AST와 어긋난다

| map | 문서 헤더 | 실제 AST |
|---|---|---|
| `MarkAlertDelivered` | 154–165, `39fe…` | 186–197, `dd521…` |
| `Notify` | `d5b300…` | `6f740…` |
| `notifyCritical` | `6b77…` | `6f740…` |
| `deliver` | `6b77…` | `6f740…` |
| `EnqueueAlert` | 114–117 | 115–118 |

그리고 커버리지 해석이 틀렸다. `-covermode=set`은 블록이 **실행됐는지**를 0/1로 기록하며
**횟수를 세지 않는다.** 따라서 "블록 진입 횟수 = 불필요한 재전송 횟수"는 근거 없는 주장이다.

### CONCERN

1. `alert_outbox.state`에 CHECK 제약이 없다. PENDING이 아닌 **어떤** 문자열도 `owed=false`가
   된다 — 알 수 없는 상태는 오류이거나 owed여야지 조용한 종결이어서는 안 된다.
2. `ClaimAlertForDelivery`의 주석이 ACKNOWLEDGED를 "전송 성공 위에 얹힌 것"이라고 하지만,
   acknowledge는 PENDING에서만 허용된다.
3. 설계가 `Notifier.Flush`가 `parkAlert`의 행을 집어 간다고 전제하는데, **`Flush`에는
   non-test 호출자가 없다.** `replay.go:101`의 주석은 현재 배선이 뒷받침하지 않는다.
4. `Flush`가 `MarkAlertAttemptFailed` 오류를 버려서, 기록에 실패하고도 깨끗한 결과를 보고할 수 있다.
5. 598.9초가 병행 경합 때문이라는 인과는 경과 시간 비교만으로는 성립하지 않는다.

### Codex가 검증하고 옳다고 한 것

- 상태 기계: 없음→PENDING, PENDING→PENDING/DELIVERED/ACKNOWLEDGED, 종결 상태에서 나가는 전이 없음
- `Publish` 성공 후 표시 전 크래시는 행을 PENDING으로 남겨 재시도된다 — 조용히 버려지지 않는다
- `EnqueueAlert`가 서명을 유지하고 기존 호출자를 바꾸지 않는다
- 구조화 로그가 억제보다 먼저 실행된다
- Go diff가 손절·청산·사이징·주문 제출·Guardian 로직을 직접 바꾸지 않는다
- `ast.json` 해시는 현재 소스와 일치한다 — 어긋난 것은 Markdown 쪽이다

## §4. 저자 검증 — 네 blocker 모두 사실이다

직접 확인했다.

- **B1**: `Flush`(`notifier.go:325-353`)는 `n.mu`를 **잡지 않는다**. `deliver`와 동시에
  발행할 수 있다. Codex가 지적한 것보다 넓다.
- **B2**: `if !owed { return nil }`이 `deliver` 앞이므로 DELIVERED 행은 transport를 다시
  시험하지 않는다. D5에서 내가 쓴 반대 방향 논증("성공한 전송이 escalate를 만든 적은 없다")은
  **과거의 전송이 성공했다**는 사실을 **미래의 전송도 성공한다**로 바꿔치기한 것이다.
- **B3**: `exit.proposal_refused` key에 사유가 없음을 확인했다(`exitloop.go:1550-1551`).
  그리고 D4의 논증은 이 change가 **버그로 지목한 바로 그 혼동**의 거울상이다 —
  "행에만 적용되고 전송에는 적용되지 않는다"를 결함으로 쓴 뒤, 같은 구별을 정당화에 썼다.
- **B4**: 표의 다섯 항목 모두 재현했다. `check_analysis.py`가 못 잡은 이유는 그것이
  `ast.json`의 해시만 소스와 대조하고 **Markdown 헤더 숫자는 보지 않기** 때문이다.
  a085 §13이 남긴 순서(코드 → AST → prose → 헤더 → 측정)를 이 change에서 다시 어겼다.
- **CONCERN 3**: `Notifier.Flush`에 non-test 호출자가 **없음**을 확인했다.

`-covermode=set`이 0/1이라는 것도 맞다. 다만 GREEN의 `0`이 "그 블록이 한 번도 실행되지
않았다"를 뜻하는 것은 유효하고, 그로부터 "이미 전달된 행에 대한 재전송이 스위트에서
발생하지 않았다"는 따라 나온다. 무효인 것은 RED의 `1`을 **횟수**로 읽은 부분이다.
문구를 그렇게 좁혀 고쳐야 한다.


## §5. 2판이 무엇을 바꿨나

| 리뷰 항목 | 조치 | 어디서 |
|---|---|---|
| **B1** 동시 claim의 이중 전송 | `claimAndDeliver`가 `n.mu`를 claim부터 send까지 잡는다. `Flush`도 같은 잠금 | design D5ter, `notifier.go:220-241`, `TestConcurrentObservationsOfOneConditionSendOnce`(`-race`) |
| **B2** 영구 억제가 진입 차단을 없앤다 | 억제를 **창**으로 바꾸고 창을 넘긴 행을 **PENDING으로 재무장**. 리마인더가 최초 전달과 같은 경로를 걸으므로 예산·gate·승격이 그대로 적용된다 | design D5, `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` |
| **B3** 같은 key의 다른 원인이 삼켜진다 | 같은 창. 다른 원인의 재발은 최대 한 창 뒤 전달된다. D4의 "기존 한계"라는 주장을 철회했다 | design D4, `TestALaterOccurrenceOfTheSameKeyIsNotSwallowed` |
| **B4** 증거 불일치 + covermode 오독 | FLM 헤더를 `ast.json`에서 **기계 생성**한다(a085 §12.3). covermode 문구를 "발생 여부"로 좁혔다 | tasks 1.13, design D6 |
| **C1** 모르는 상태가 조용히 종결 | `claimOwed`의 `default`가 **owed** | design D3, `TestAnUnrecognisedAlertStateOwesDelivery` |
| **C2** ACKNOWLEDGED 주석이 틀림 | `ClaimAlertForDelivery` 주석 재작성. ACKNOWLEDGED가 PENDING에서 온다는 사실을 `claimOwed` FLM에 명시 | `outbox.go:131-166` |
| **C3** `Flush`에 non-test 호출자 없음 | a096의 문제가 아님을 확인하고 tasks §7에 기록. 다만 잠금은 지금 넣어 둔다 | `Notifier.Flush` FLM |
| **C4** `Flush`가 `MarkAlertAttemptFailed` 오류를 버림 | 기존 동작. tasks §7에 기록하고 여기서 고치지 않는다 | `Notifier.Flush` FLM |
| **C5** 598.9초의 인과가 미확립 | 단독 재실행 310.3s < base 325.2s, 커버리지 동일 74.9%로 확인했다 | §1 |

### 2판이 스스로 더한 것

- `claimOwed`를 **순수 함수로 분리**했다. 1판에서 이 판정은 SQL 사이에 끼어 있었고, 표로
  놓고 보지 않아서 두 칸(창 밖 종결, 모르는 상태)이 틀린 것을 놓쳤다. 지금은 7분기의
  독립 map을 갖는다.
- `notifyCritical` B4에 `owed &&`를 더했다. 리뷰가 지적하지 않았지만, 억제를 도입하면서
  `!sent`만으로 escalate를 판정하면 **억제될 때마다 gate가 잠긴다**. 반대 방향의 결함이다.

- **`Acknowledge`에도 같은 잠금을 넣었다.** 리뷰의 B 항목("재무장이 무엇을 깨는가")을
  스스로 밟아 보다 찾았다. 해제는 세고-판정하는 read-then-decide인데, 재무장이 그 사이에
  PENDING 행을 만들면 **미전달 critical 알림이 있는 채로 gate가 열린다.** 이전에도 이론적으로
  가능했지만(새 조건의 첫 알림), a096이 두 번째 원천을 만들어 노출을 넓혔다.
  `TestAcknowledgeCannotClearTheGateMidSend`가 상호 배제를 건다.

- **BTM 표의 `:줄` 참조를 `ast.json`과 기계 대조하는 검사를 만들었다.** 헤더를 생성으로
  바꾼 뒤에도 표 안의 숫자는 사람이 타이핑하므로 같은 방식으로 썩는다. 첫 실행에서 5건이
  나왔다. 리뷰 1라운드 B4가 헤더에서 잡은 것과 같은 부패가 표 안에 남아 있었다.

- **측정하지 않은 RED 열을 `미측정`으로 표시했다.** `Acknowledge`·`Flush`는 base 시점의
  분기 대응을 뜨지 않았는데, 처음에는 그럴듯한 값을 적어 두었다. 리뷰가 지적한 것과 같은
  종류의 거짓이라 되돌렸다.

## §6. 2라운드 리뷰어가 특히 봐야 할 것

1. **재무장이 옳은 형태인가.** 종결 행을 PENDING으로 되돌리면 `UndeliveredCount`가 올라가고,
   그것을 읽는 곳(`Acknowledge`의 해제 조건)이 영향을 받는다. 리마인더가 대기 중인 동안
   운영자의 acknowledge가 gate를 풀지 못하게 되는 시나리오가 있는지.
2. **1시간이라는 창.** 주말 60시간 지속 조건이 60건이 된다. 상한으로 충분한가,
   지수 backoff가 필요한가(design D5bis에서 검토하고 기각한 근거를 반박해 볼 것).
3. **`n.mu` 구간이 claim까지 넓어진 것.** a092가 반대 방향으로 가는 change다. 이 확장이
   관측 사이클을 붙잡는 정도, 그리고 두 change가 만났을 때의 순서.
4. **`latestStamp`가 두 시각 중 나중 것을 고르는 것.** 재무장이 이전 episode의 시각을
   남기므로 둘 다 있을 수 있다. 그 선택이 창을 늘 옳게 잡는지.
5. 1라운드의 §3 blocker들이 실제로 닫혔는지 — 특히 B2의 ACKNOWLEDGED 변형.


## §7. 2판의 수치

| 항목 | 결과 |
|---|---|
| `go build ./...` · `go vet` | 통과 |
| `go test -race ./internal/obs/` | 통과 (70.6s) |
| `go test ./internal/journal/ ./internal/execgw/` | 통과 (493s, 81s) |
| `go test ./...` | 통과 — FAIL 패키지 없음 |
| obs 커버리지 | 84.8% (RED 84.8%) |
| journal 커버리지 | **75.0%** (RED 74.9%) |
| `check_analysis.py` | evidence complete — **10함수** |
| BTM 줄 대조 | MISMATCHES 0 |
| `openspec validate --strict` | valid |
| Go diff | `internal/journal/outbox.go`, `internal/obs/notifier.go` 두 파일 |
| 스키마 | 마이그레이션 없음, `SchemaVersion` 30 |

결정적 측정 두 가지:

```text
deliver B5 (이미 전달된 행에 재표시)   RED 진입 → GREEN 미진입
ClaimAlertForDelivery B6 (재무장)      GREEN 진입
```

앞의 것은 폭주가 사라졌음을, 뒤의 것은 리마인더가 실제로 동작함을 말한다. 하나만으로는
어느 쪽도 증명되지 않는다 — 전자만 보면 영구 억제(1판)와 구별되지 않는다.

## §8. 2라운드 — Claude Sonnet, 판정 PASS

구현자와 분리된 읽기 전용 세션에서 change 산출물, 현재 Go diff, 전체 호출자와 상태 전이를
대조했다. 다음을 직접 실행했다.

- `go build ./...`
- `go vet ./internal/journal/... ./internal/obs/... ./internal/execgw/...`
- `go test -race ./internal/obs/...`의 a096 8건 + 기존 회귀 2건
- `go test ./internal/journal/...`의 a096 8건
- `go test ./internal/obs/... ./internal/execgw/...` 전체
- `openspec validate a096-one-condition-is-one-alert --strict`

리뷰어는 1라운드 blocker 네 건이 모두 닫혔다고 판정했다: claim-to-send 배타 구간,
창을 넘긴 재무장과 진입 차단 복원, 같은 key의 후속 원인 전달, AST/문서 근거 정합이다.
unknown 상태도 owed+PENDING 재무장 후 완료 표시가 성공하는 것을 확인했다.

### concern과 조치

1. `remindAfter <= 0` 주석이 unknown 상태 복구까지 금지하는 것처럼 읽힐 수 있었다.
   코드는 unknown을 PENDING으로 복구하는 안전 방향이 맞으므로, `EnqueueAlert` 주석을
   "종결 행의 시간 기반 재무장은 하지 않되 unknown 복구는 한다"로 명확히 했다.
2. claim부터 재시도 전체까지 잡는 `n.mu`가 알림 경로를 직렬화한다. 이는 a092와 합칠 때
   다시 판단해야 하는 이미 기록된 트레이드오프이며 a096의 blocker는 아니다.

최종 판정: **PASS — blocker 0**.
