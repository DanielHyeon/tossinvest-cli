# a099 tasks — claim이 두 번째 발송자를 배제한다

base: `285c7619110d2f8c53a1d9ddfbadd16ad0e9e53e` (`base-commit.txt`)

**High-risk**: 원장 스키마(불변식 5). Pre-Edit 선언은 design D5.

## 1. 증거 (완료)

AST 산출물 **열셋**을 **proposal·design보다 먼저** 만들었다 — 열하나에 3라운드가 둘을 더했다.
2라운드 A-T2·B-T2가 「여덟/아홉/열하나」 혼재를 잡았다 — **디스크가 정본이고 지금은 열셋**이다.
`.claude/CLAUDE.md`의 규칙 — 분기를 근거로 쓰는 문서는 proposal이라도 AST가 먼저다.
**규칙은 편집 여부가 아니라 「근거로 쓰는가」에 걸린다** — 그래서 안 고치는 셋에도 번들이 있다.

| 번들 | 함수 | 분기 / 이탈 | 이 change에서 |
|---|---|---|---|
| `internal-journal--journal.claimalertfordelivery` | `ClaimAlertForDelivery` `outbox.go:169-261` | 11 / 10 | **편집** — 임차 CAS |
| `internal-journal--claimowed` | `claimOwed` `outbox.go:269-315` | 8 / 7 | **인용만** — B2 `:276`가 반증의 근거 |
| `internal-journal--journal.markalertdelivered` | `MarkAlertDelivered` `outbox.go:337-348` | 1 / 2 | **편집** — 토큰 술어 + 성공 시 **해제** |
| `internal-journal--journal.markalertattemptfailed` | `MarkAlertAttemptFailed` `outbox.go:352-363` | 1 / 2 | **편집** — 토큰 술어 + 임차 **유지** ⚠ 3판까지 이 칸이 *"임차 해제"*였다 |
| `internal-journal--journal.undeliveredcount` | `UndeliveredCount` `outbox.go:408-415` | 1 / 2 | **인용만** — 술어 불변이 D1의 안전 논거 |
| `internal-journal--journal.pendingalerts` | `PendingAlerts` `outbox.go:392-405` | 2 / 2 | **인용만** — 같음 |
| `internal-obs--notifier.claimanddeliver` | `claimAndDeliver` `notifier.go:238-277` | 4 / 3 | **편집** — doc + **취득 결과 셋을 가르는 분기 하나** ⚠ 3판까지 이 칸이 *"doc만 · 제어 흐름 불변"*이었다 |
| `internal-obs--notifier.flush` | `Flush` `notifier.go:427-462` | 6 / 4 | **편집** — 행마다 claim (D7) |
| `internal-obs--notifier.deliver` | `deliver` `notifier.go:341-407` | 12 / 3 | **편집** — 포기 시 해제. **이 번들이 설계를 하나 뒤집었다** |
| `internal-journal--journal.enqueuealert` | `EnqueueAlert` `outbox.go:115-122` | **0** / 1 | **편집** — no-claim 위임 (D13) |
| `internal-journal--journal.acknowledgealert` | `AcknowledgeAlert` `outbox.go:370-384` | 2 / 3 | **인용만** — 술어가 D1·D10의 근거 |
| `internal-obs--notifier.acknowledge` | `Notifier.Acknowledge` `notifier.go:476-514` | 10 / 6 | **인용만** — C6이 `:481-482`·`:510-511`을 *"a099가 안 건드린다"*의 근거로 쓴다 |
| `internal-journal--open` | `Open` `journal.go:108-185` | 14 / 8 | **편집** — `Options.AlertLease` 기본값 (C4) |

> **`deliver` 번들은 나중에 생겼고, 그것이 규칙이 작동한 증거다.**
> `deliver`는 처음에 편집 대상이 아니었다. 그 AST를 뽑고 나서야
> B8 `:384` → B9 `:387` → B10 `:388` → 루프 복귀가 보였고,
> **초안의 「실패 기록이 임차를 푼다」가 이중 발송을 만든다**는 것이 나왔다.
> 손으로 읽었으면 `:384`에서 멈췄을 자리다.

- [x] 1.1 **열셋** 번들의 `ast.json`·`risk-pattern-report.md` 생성 (열하나 + 3라운드의 둘)
- [x] 1.2 `base-commit.txt` 고정 (`capture_change_base.py`)
- [x] 1.3 D0.3 반증의 근거를 AST 열거로 확정 — B6 `:197`이 유일한 UPDATE 관문,
  `claimOwed` B2 `:276` → `:278` `return true, false`

## 2. Function Logic Map / Branch Test Map 완성

- [x] 2.1 **열셋** 번들의 두 Markdown map을 AST 열거로 채웠다
- [x] 2.2 `check_analysis.py --change a099-a-claim-excludes-the-second-sender` 통과
- [x] 2.3 `Notifier.Flush`의 a099 전용 번들 생성 (a098 것과 별개)
- [x] 2.4 **번들 둘을 더 만든다** (1라운드 A-T2·B-T2):
      `Journal.AcknowledgeAlert` — design D1이 그 술어를 근거로 쓴다 ·
      `Journal.EnqueueAlert` — A-P5가 이 함수의 편집을 요구한다.
      **총 열하나가 된다**
- [x] 2.5 **branch-test-map의 Test 칸을 전부 실측한다.** 1라운드 B-T1이
      `internal/obs/notifier_test.go`가 **존재하지 않는다**는 것을 잡았다 —
      네 곳에서 인용했었다. 남은 「기존」 칸도 파일·함수명을 실제로 확인한다

      **잰 방법 — 인용을 「읽은」 것이 아니라 「푼」 것이다.** 열셋 번들의
      `branch-test-map.md`에서 `*_test.go` 인용과 `Test*` 이름을 전부 뽑고,
      줄 번호마다 **그 줄을 감싸는 `func`를 실제로 찾아** 문서가 적은 이름과 맞췄다.

      | 검사 | 결과 |
      |---|---|
      | 인용한 `*_test.go` 파일이 실재하는가 | **전부 실재** (13/13 번들) |
      | 인용한 `Test*` 함수가 어딘가에 있는가 | **전부 있음** |
      | **`파일:줄`이 문서가 적은 함수 안인가** | **12 맞고 1 틀림** |

      **틀린 하나 — `open` 번들 B7.** `durability_test.go:902 openTestJournalWithBusy`라고
      적혀 있었는데 **902는 빈 줄**이고 그 함수는 `:903`이다. 한 줄 밀렸다.
      부르는 자리 둘(`:616`·`:617`)도 같이 적었다 — 둘 다 `100ms`를 **주므로**
      「기본값 쪽 미검증」이라는 판정은 그대로 옳다.

      > **오탐 하나도 적어 둔다.** 검사가 `acknowledgealert` 번들에서
      > `internal/obs/notifier_test.go`를 잡았는데, 그 줄은 인용이 아니라
      > **1라운드 B-T1이 잡은 오류를 기록한 문장**이었다. 없는 파일의 이름은
      > 근거로 쓸 때만 결함이고, **없다고 적을 때는 기록**이다.

      **덤으로 고친 것 셋** (같은 실측에서 나왔다):

      - `flush` B1 — *"`Flush`를 부르는 두 파일"*이 실제로는 **네 함수**다.
        이름을 다 적었고, **넷 다 journal을 배선하므로 B1의 참 쪽은 안 덮인다**는
        사실이 드러났다. 칸을 `yes (기존)`에서 **`아니오 — 거짓 쪽만 덮인다`**로 고쳤다.
      - `flush` B1·B4의 *"§2.4가 실측한다"* — §2.4는 닫혔는데 미룸이 남아 있었다.
        **미룬 곳이 아니라 잰 값으로** 바꿨다.
      - R1·R2·R3·R4의 *"planned RED — 미관측"* 넉 줄 — §3.1이 넷 다 관측했다.
        번들에도 반영했다. **관측이 문서보다 앞서 있으면 문서가 거짓말을 한다.**
- [x] 2.6 **열한 번들을 D-CORE와 §3.0 등록부에 맞춘다** (2라운드 A-T1 = B-P5의 처방).
      번들이 인용하던 `owed=false`·옛 R 번호·옛 API를 **정본 하나로 모은다.**
      `enqueuealert` 번들이 `R10`, tasks가 같은 것을 `R11`이라고 부르던
      **번호 분기**가 이 라운드에 실제로 있었다
- [x] 2.7 **번들 둘을 더 만든다** (3라운드):
      `obs.Notifier.Acknowledge` — design C6이 `:481-482`·`:510-511`의 분기를
      *"오늘 이렇게 동작하고 a099가 하나도 안 건드린다"*의 근거로 쓴다.
      **인용만 하는 함수에도 규칙이 걸린다** — 편집 여부가 아니라 「근거로 쓰는가」다 ·
      `journal.Open` — `Options.AlertLease` 기본값 채우기가 이 함수 **내부**를 바꾼다(§4.1b).
      **총 열셋이 된다**
- [x] 2.8 **번들 열셋의 map을 4판 D-CORE에 맞춘다.** 3라운드가 잡은 사본은 여덟 자리다 —
      `claimant` 없는 시그니처 · *"열 둘"* · *"제어 흐름 불변"* · *"`:381` 직전 해제"* ·
      *"34초"* · `MarkAlertAttemptFailed`의 *"해제"*.
      **정본 절을 만드는 것만으로는 부족하다는 것이 세 라운드로 증명됐다** —
      §5.6이 편집 후 다시 뽑을 때도 같은 대조를 한다

## 3. RED — 실패하는 것을 먼저 관측한다

> **⛔ 2라운드가 이 절을 통째로 다시 쓰게 했다.**
> 2판의 R 목록은 **작성 불가 셋 · born-GREEN 여섯**을 섞어 놓고 3.1이
> *"R1~R9를 RED로 작성한다"*라고만 적었다. R10~R13은 **어느 task도 안 가리켰다.**
> 그리고 「오늘 실패한다」가 거짓인 R을 RED라고 불렀다.
>
> **RED는 시점이 있는 주장이다.** 임차 개념 자체가 없는 오늘, 임차에 관한 단언은
> 대부분 **자동으로 참**이다. 그러면 그것은 RED가 아니라 born-GREEN이고, a092 19라운드
> A-P6이 같은 실패를 이미 잡았다. 그래서 아래 표는 **각 R이 언제 빨간불이 되는지**를
> 적는다. 그 시점을 안 적으면 관측할 수 없는 RED다.

### 3.0 R 등록부 — 이것이 정본이다

**종류가 셋이다.**

| 종류 | 뜻 |
|---|---|
| **RED 오늘** | 현재 HEAD에서 실패한다. 실패 출력을 3.1에 붙인다 |
| **RED — 시점** | 오늘은 통과한다(개념이 없다). 적힌 시점에 실패해야 하고, **그 실패를 관측해서 붙인다** |
| **회귀 핀** | 어느 시점에도 빨간불이 아니다. **안 고른 설계를 배제한다.** 테스트 주석에 그렇게 적는다 |

> **⛔⛔ 4판이 이 표를 다시 손봤다. 3라운드가 「RED라고 부른 것의 절반이 안 빨갛다」를
> 잡았기 때문이다.** born-GREEN을 RED라고 부르는 것은 a092 19라운드 A-P6이 이미 잡은
> 실패이고, 3판은 그것을 **줄여 놓고 없애지 못했다.**
>
> **번호는 안 옮긴다.** 3판이 2판에서 번호를 옮기다 번들과 tasks가 갈라졌다.
> R5는 사용자 결정 5-1로 **삭제**되고 그 번호는 **빈 채로 남긴다** — 재사용하지 않는다.

| # | 무엇 | 종류 / RED 시점 | 근거 |
|---|---|---|---|
| **R1** | 발송자 둘이 같은 PENDING 행을 claim하면 **하나만** 발송 권한을 얻는다 | **RED 오늘** — `claimOwed` B2 `:276` → `:278`이 **둘 다에게** `true, false`를 준다 | AST |
| **R2** | claim에 진 쪽은 **publish를 안 한다** | **RED 오늘** — `claimAndDeliver` 이탈 `:276`이 owed면 무조건 발송한다 | AST |
| **R3** | 그 배제가 **서로 다른 뮤텍스**를 쓰는 두 발송자 사이에서도 성립한다 | **RED 오늘** — 오늘 배제는 `n.mu` 하나뿐(`notifier.go:242` defer) | AST + `outbox.go:166-168` |
| **R4** | `Flush`도 claim을 거친다 — 못 잡은 행은 **publish 안 한다** | **RED 오늘** — `Flush`는 claim을 아예 안 부른다 | `notifier.flush` 분기 6 / 이탈 4 |
| ~~R5~~ | ~~시도조차 안 한 미전달 행이 신규 진입을 막는다~~ | **삭제 — 사용자 결정 5-1.** 진입 게이트의 트리거를 안 바꾼다(C6). 번호는 **재사용하지 않는다** | — |
| **R6** | 실패한 시도는 임차를 **안 푼다** — 같은 발송자가 예산 안에서 재시도하는 동안 두 번째 발송자가 못 들어온다 | **회귀 핀** ← 3판은 RED라고 적었다. **어느 시점에도 안 빨갛다**: §4.5 전에는 `MarkAlertAttemptFailed`가 임차를 아예 모른다. 배제하는 설계 = **「실패 기록이 임차를 푼다」**(초안) | `deliver:384` → `:387-391` · 3라운드 B |
| **R7** | 발송자가 **예산을 다 쓰고 포기하면** 임차가 풀린다 | **RED — §4.5b 전** (`ReleaseAlertClaim`이 없어 임차가 만료까지 남는다) | design D3 |
| **R8** | publish 성공 + **정산 실패**면 임차를 **안 푼다** — 만료가 풀 때까지 억제 표시다 | **회귀 핀** ← 3판은 RED. 배제하는 설계 = **「`deliver:381` 이탈에서 해제한다」**(내가 셋으로 늘렸던 것, 1라운드 B-P3이 되돌렸다) | design D3 |
| **R9** | claim한 발송자가 죽으면 **만료 뒤** 다른 발송자가 집는다 | **RED — §4.3 뒤 §4.3d 전.** 취득 CAS만 있고 만료 술어가 없으면 그 행은 **영원히** 못 집힌다 | design C2 |
| **R10** | **시계가 뒤로 가도 임차가 스큐만큼 길어지지 않는다** | **RED — §4.3d 뒤 §4.3e 전.** 만료 술어만 있으면 역행에서 `claim_expires_at`이 미래로 남아 안 열린다 | design C2 · D4 |
| **R11** | **정확히 skew만큼 역행한 경계에서 훔치지 않는다** | **회귀 핀** — 배제하는 설계 = **역행 경계를 `>=`로 쓴 구현**. ⛔ **5라운드 B-T5: 4판의 근거 문장이 거짓이었다** — *"`>=`면 **방금 발급한** 임차가 자기 규칙에 걸린다"*고 적었는데, C2는 `claimed_at`과 `now+skew`를 비교하고 fresh claim은 `claimed_at = now`이므로 **`>=`여도 안 걸린다.** 참인 핀은 **경계값 그 자체**다 | 1라운드 A-P6 · 5라운드 B-T5 |
| **R12** | **임차 기간이 다른 경합자가 남의 유효한 임차를 못 훔친다** | **회귀 핀** — 배제하는 설계 = **「만료를 저장하지 않고 판정자의 lease로 유도한다」**(3판 C1). **3판 설계에서는 쓸 수조차 없었다** — lease가 원장 것 하나뿐이었다. 사용자 결정 6-1이 이 핀을 **쓸 수 있게** 만들었다 | 1라운드 B-P4 · 3라운드 A-P5 = B-R12 |
| **R13** | 임차를 잃은 발송자의 결과 기록이 **새 보유자의 임차를 안 지운다** | **RED — §4.5 전** (토큰 술어가 없으면 `id + state`만으로 남의 행을 친다) | design D3 |
| **R14** | **같은 이름의 발송자가 재취득해도** 옛 발송자가 새 임차를 못 친다 (ABA) | **RED — 같은 시점** | 1라운드 A-P4 = B-P5 |
| **R15** | 임차를 잃은 발송자가 **남은 전송을 즉시 멈춘다** | **RED — §4.5 전** ← 4라운드 B 정정. §4.5 자체가 `SettleLeaseLost → 즉시 중단`이라는 호출자 계약을 만든다. §4.5c는 **로그** 단계라 그때는 이미 GREEN이다 | design C3의 `SettleLeaseLost` |
| **R16** | 기록 전용 경로(`replay.go:551`)로 만든 행을 **배달 실행자가 즉시 집을 수 있다** | **RED — §4.3 뒤 §4.4b 전** ⛔ **5라운드 B-P7 = B-T6: 4판의 정정이 틀렸다.** 4판은 *"시그니처가 바뀌면 컴파일이 강제하므로 빨간 시점이 없다"*며 회귀 핀으로 내렸는데, **claimant 문자열을 넘겨 컴파일만 맞추는 나쁜 수정이 가능하다** — no-claim 위임이 강제되지 않는다. **이 change의 Branch Test Map(`enqueuealert/branch-test-map.md:17-29`)이 처음부터 단계 RED라고 적고 있었고 등록부만 달랐다** | design D13 · 5라운드 B-P7 |
| **R17** | **재무장이 이전 episode의 임차를 지운다** — 운영자가 승인한 행에 남은 임차가 새 episode로 안 넘어간다 | **RED — §4.6 전.** ⚠ 조건: **`AlertLease > remindAfter`인 설정에서 관측한다.** 기본 remindAfter(1시간)가 lease보다 길면 그 임차는 재무장 시점에 **이미 만료**라 born-GREEN이다 (4라운드 B) | design C5 · `acknowledgealert` 번들 |
| **R18** | claim 경합·만료 탈취·토큰 불일치가 **전송 실패와 다른 이벤트로 남는다** | **RED — §4.5c 전.** 3판 API로는 작성 불가였고 `ClaimResult.Stole*`가 그 절반을 열었다. **나머지 절반(토큰 불일치 시 새 보유자·임차 나이)은 4라운드 B가 여전히 불가라고 판정했고, C3의 `SettleResult`가 그것을 연다** | design D12 · C3 |
| **R19** | v30 DB를 v31로 올릴 때 **부분 실패가 `ALTER TABLE` 넷과 `user_version`을 함께 되돌린다** | **회귀 핀** ← 3판은 RED. **migration 트랜잭션의 일반 성질**은 이미 단언돼 있다 — `TestFailedMigrationLeavesTheJournalRestorable`(`internal/journal/migration_v5_test.go:302`). ⛔ **5라운드 B-T7: 4판이 `:341`이라고 적었다** — 그 줄은 함수 선언이 아니라 본문이다. **4라운드에 이 인용을 고쳤는데 고친 값이 또 틀렸다.** ⚠ **그 테스트는 v4→v5 케이스이지 v31이 아니다**(4라운드 B). a099는 **v31 케이스를 그 테스트에 더하고**, 그 케이스는 §4.1 전에 존재할 수 없으므로 **빨간 시점이 없다** | design D11 |
| **R20** | **`DefaultAlertLease > DefaultAlertDeliveryBound()`** — 설정 상수가 바뀌면 이 테스트가 깨진다 | **RED — §4.1b 전** (두 값이 없어 **컴파일이 안 된다**). ⚠ 4라운드 B: 오늘 `defaultBusyTimeout`은 비공개이고 `Ntfy`의 10초는 **리터럴**(`ntfy.go:95`)이라 유도를 **하드코딩 없이 쓸 수 없다.** 그래서 **C4가 `obs.DefaultAlertDeliveryBound()`를 요구한다**. ⛔ **5라운드 B-P8: 부등식만으로는 약하다** — **거짓 `DefaultAlertDeliveryBound()`(예: 0을 반환)나 과도하게 큰 lease도 통과한다.** 그 함수가 읽는 **입력값 넷**(`Attempts`·`RetryDelay`·`Ntfy.Timeout`·`BusyTimeout`)을 각각 바꿔 **결과가 따라 움직이는 것**까지 단언한다 | design C4 · C7 · 5라운드 B-P8 |
| **R21** | 발송 중인 행이 `UndeliveredCount`·`PendingAlerts`에 **계속 잡힌다** | **회귀 핀** — 배제하는 설계 = **D1의 안 A(`SENDING` 상태)** | `outbox.go:411` |
| **R22** | **운영자 승인이 임차의 유무에 영향받지 않는다** | **회귀 핀** — 배제하는 설계 = **승인 술어에 임차를 넣는 것**(C5) | `acknowledgealert` 번들 |
| **R23** | **정상 발송 중의 경합이 진입 게이트를 잠그지도 풀지도 않는다** | **회귀 핀** — 배제하는 설계 = **2판 D10(경합에 래치)** 과 **3판 C6(경합에 유도)** **둘 다**. ⛔ **5라운드 B-P8: 최종 `Blocks()`만 보면 중간 `Block→Clear`가 통과한다** — `EntryGate.revision`(`retry.go:498-515`·`:568-576`)이 **안 움직인 것**이나 호출 spy로 **호출 이력 자체**를 고정한다 | 2라운드 A-P1 · 3라운드 A-P1 · 5라운드 B-P8 |
| **R24** | 발송자가 하나이고 아무도 안 죽었을 때 **오늘과 같은 발송 결정** | **회귀 핀** — 배제하는 설계 = **「임차 취득 실패를 발송 안 함의 근거로 확대해 단독 발송자의 결정까지 바꾸는 구현」**. ⛔ **5라운드 B-P8: 한 경우만 보면 약하다** — 오늘의 **결정표 전체**(PENDING · settled-window · unknown-state, `outbox.go:275-314`)를 **행마다** 고정한다 | 1라운드 A-P7 · B-P1 · 5라운드 B-P8 |
| **R26** | **운영자가 승인한 행은 발송자의 정산을 거부한다** — 임차를 든 발송자가 publish 중에 운영자 승인이 들어오면 그 발송자의 `MarkAlertDelivered`가 **실패한다** | **회귀 핀** — 오늘은 `state = ?`/`AlertPending` 술어(`outbox.go:342-343`)가 이것을 **공짜로** 준다. 배제하는 설계 = **§4.5의 토큰 술어를 더하면서 상태 술어를 느슨하게 만드는 구현**. 상태가 `ACKNOWLEDGED`인데 토큰이 맞아 통과하면 **운영자가 안 본 것을 봤다고 원장이 적는다** | a098 spec Scenario `spec.md:186`(「발송 중인 행도 운영자가 승인할 수 있다」)의 **둘째 주장**. 6라운드 P5 — ①은 R22, ③은 R17이 지는데 **②만 주인이 없었다** |
| **R25** | 운영자의 밀린 목록이 **임차 상태를 보여준다** — 누가 언제부터 들고 있는지 | **RED — §4.10 전** (`alertSelect`에 열이 없다). ⛔ **5라운드 B-P8: 보유자·시각만 보면 D12를 다 안 잰다** — **토큰 원문이 안 나오는 것**까지 같이 단언해야 불변식 8이 고정된다 | design D12 · 5라운드 B-P8 |

**3판 → 4판에서 바뀐 것** — 침묵한 변경을 금지하므로 전부 적는다.

| R | 3판 | 4판 | 왜 |
|---|---|---|---|
| R5 | RED 오늘 | **삭제** | 사용자 결정 5-1 |
| R6 · R8 | RED — 시점 | **회귀 핀** | 어느 시점에도 안 빨갛다 (3라운드 B) |
| R9 · R10 | RED — 시점(§4.3) | **RED — 시점(§4.3d·§4.3e)** | **시점이 실재하도록 §4.3을 셋으로 쪼갰다.** 3판은 한 덩어리라 그 사이가 없었다 |
| R11 · R12 | RED — 시점 | **회귀 핀** | 관측하려면 **일부러 틀린 구현**을 써야 한다. 그것은 RED가 아니라 배제다 |
| R19 | RED — 시점 | **회귀 핀** | `migration_v5_test.go:302`가 일반 성질을 이미 단언한다 (4판은 `:341`이라고 적었다 — 5라운드 B-T7) |
| R18 | RED — 시점 (**작성 불가**) | **RED — 시점 (작성 가능)** | C3의 `Stole*` |
| R23 | 회귀 핀 (*"사람 개입 없이 풀린다"*) | **회귀 핀** (*"잠그지도 풀지도 않는다"*) | 결정 5-1이 배제 대상을 **둘로** 만들었다 |
| — | — | **R25 신설** | §4.10(임차 투영)을 **어느 R도 안 보고 있었다** |

**종류별 수**

| 종류 | 수 | 어느 것 |
|---|---|---|
| RED 오늘 | **4** | R1 · R2 · R3 · R4 |
| RED — 시점 | **11** | R7 · R9 · R10 · R13 · R14 · R15 · **R16** · R17 · R18 · R20 · R25 |
| 회귀 핀 | **10** | R6 · R8 · R11 · R12 · R19 · R21 · R22 · R23 · R24 · **R26** |
| **합** | **25** | 빈 번호 하나(R5) — 재사용하지 않는다 |

> **⛔ 5판 — 5라운드 B가 R16을 되돌렸다.** 4판이 **회귀 핀으로 강등**한 것을
> **RED — 시점(§4.3 뒤 §4.4b 전)으로 되돌린다.** 수는 그대로 4 / 11 / 9인데
> **11의 구성원이 바뀌었다** — 4판 표에는 R16이 이미 11에 들어 있었고 등록부만
> 회귀 핀이라고 적고 있었다. **표와 등록부가 서로 다른 말을 하고 있었다**(B-T6).
> 같은 모순이 이 change의 `enqueuealert/branch-test-map.md:17-29`에도 있었고,
> **그 map이 처음부터 옳았다.**

- [x] 3.1 **RED 오늘** 넷(R1~R4)을 작성하고 **실패를 실제로 관측한다.** 출력을 여기 붙인다

      **작성한 파일 둘 — 기존 테스트 파일은 안 건드렸다** (새 코드는 새 파일에):

      | 파일 | R |
      |---|---|
      | `internal/journal/a099_claim_excludes_the_second_sender_test.go` | R1 |
      | `internal/obs/a099_claim_excludes_the_second_sender_test.go` | R2 · R3 · R4 |

      **관측 (2026-08-11, `go test -race`) — 넷 다 빨갛고, 넷 다 「2, want 1」이다.**

      ```text
      a099_…_test.go:83:  senders granted the right to send = 2, want 1 — two senders each
                          believing the send is theirs is the double delivery this change
                          exists to stop
      --- FAIL: TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend (2.58s)
      FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	2.593s

      a099_…_test.go:147: publishes while one sender held the row = 2, want 1 — the second
                          sender lost nothing, because today there is nothing to lose
      --- FAIL: TestTheSenderThatLosesTheClaimDoesNotPublish (2.46s)
      a099_…_test.go:184: sends = 2, want 1 — n.mu excludes senders inside one Notifier and
                          there are two Notifiers here
      --- FAIL: TestExclusionHoldsBetweenSendersWithDifferentMutexes (2.52s)
      a099_…_test.go:221: publishes = 2, want 1 — Flush sent a row another sender was
                          already sending
      --- FAIL: TestFlushDoesNotPublishARowAnotherSenderHolds (2.76s)
      FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/obs	7.766s
      ```

      > **네 실패가 전부 같은 수다 — `2, want 1`.** 우연이 아니라 **하나의 원인**이다:
      > 오늘의 배제는 `n.mu` 하나뿐이고, 네 시나리오는 전부 **그 뮤텍스 밖에서** 두 번째
      > 발송자를 세운다. `-race` 경고는 없다 — **경합이 아니라 설계가 둘을 다 통과시킨다.**
      >
      > **겹침은 스케줄러에 안 맡겼다.** `a099Publisher`가 첫 publish를 붙잡아 두 번째
      > 발송자가 **반드시** 그 안에서 도착하게 한다. 가끔만 겹치는 테스트는
      > **가끔만 빨간 테스트**이고, 좋은 날에 통과하는 RED는 관측이 아니다.
      >
      > ⚠ **R1의 호출은 §4.3에서 다시 타입이 바뀐다**(design C1이 boolean을 claim 결과로
      > 바꾼다). 단언하는 성질은 안 바뀐다 — 한 행에 닿은 발송자 둘 중 **정확히 하나**만
      > 발송 권한을 들고 나간다.

      **회귀 확인 (2026-08-11) — 두 패키지 전량이 실제로 끝났고, 빨간 것은 그 넷뿐이다.**

      게이트와 같은 조건(`go test -timeout 30m -count=1`, `-race` 없이)으로 재고,
      `-json` 전량을 직접 받아서 `run` 이벤트와 `pass|fail|skip` 이벤트를 대조했다.

      | 패키지 | 시작 | 종료 | 미완 | pass | fail | skip | 소요 |
      |---|---:|---:|---|---:|---:|---:|---:|
      | `internal/journal` | 1390 | 1390 | 없음 | 1389 | 1 | 0 | 331.13s |
      | `internal/obs` | 67 | 67 | 없음 | 64 | 3 | 0 | 9.22s |

      실패 넷은 R1·R2·R3·R4 그대로다. `test timed out` 0건, `DATA RACE` 0건.

      > ⚠ **요약 줄만 읽었으면 못 볼 뻔했다.** 앞선 `-race` 실행은 요약이
      > `660 passed, 4 failed`였고 그 넷이 정확히 의도한 RED라서 **깨끗한 통과처럼
      > 보였다.** 진짜 신호는 패키지 꼬리표에 있었다 — `FAIL … internal/journal
      > 1200.028s`, 그 1200초가 곧 `-timeout 20m`이다. `panic: test timed out after
      > 20m0s`, 603 시작 / 597 종료. **넉넉하다고 믿은 예산이 부족했던 것**이고,
      > 교착은 아니었다(panic 헤더의 실행 중 테스트가 8초·1초, goroutine은
      > SQLite 토크나이저 안에서 `[runnable]`, 오래 멈춘 것들은 전부
      > `testing.(*T).Parallel` 정상 대기).
      >
      > ⚠ **세는 도구도 한 번 거짓말했다.** rtk tee 로그는 **1 MiB에서 잘린다**
      > (`--- truncated at 1048576 bytes ---`). 그 잘린 로그로 세면 끝나지 않은
      > 테스트가 **있는 것처럼** 보이고, timeout panic은 스트림 **끝**에 찍히므로
      > **없는 것처럼** 보인다. 위 표는 `rtk proxy`로 필터를 우회해 받은
      > 1.43 MiB 전량(파싱 불가 줄 0)에서 센 것이다.
      >
      > **`-timeout`은 트리 전체가 아니라 테스트 바이너리 하나마다 걸린다.**
      > 그래서 `make test`의 `-timeout 30m`(`Makefile:35`)이 `./...`에 걸리는 방식은
      > "트리 전체 30분"이 아니라 "패키지마다 30분"이다. 가장 큰 `internal/journal`이
      > `-race` 없이 331초이므로 예산은 약 5.4배 남는다. **`-race`를 붙이면 그 여유가
      > 사라진다** — 위 20분 초과가 그 증거다.
- [ ] 3.2 **RED — 시점** 열하나(R7·R9·R10·R13·R14·R15·R16·R17·R18·R20·R25)는 §4를
      진행하며 **그 시점에 실패를 관측하고** 출력을 붙인다. 시점을 못 관측하면
      **회귀 핀으로 강등해서 그렇게 적는다** — *"오늘도 통과한다"*를 RED라고 부르지 않는다.
      **3판이 이 강등을 다섯 번 안 해서 3라운드가 BLOCK했다**
- [ ] 3.3 **회귀 핀** 열(R6·R8·R11·R12·R19·R21·R22·R23·R24·**R26**)은 테스트 주석에
      *"이것은 RED가 아니다"*와 위 표의 **「배제하는 설계」 문구를 그대로** 적는다
- [x] 3.4 **임차 기간을 유도식으로 정한다** (design C4). 상수 하나를 고르는 것이 아니다.
      `bound`는 **재시도 예산 전체 + SQLite 쓰기 대기**를 덮어야 한다 — 오늘 기본값
      유도는 `3×(10+5) + 2×2 + 5 = 54초`(design D4).
      `margin`의 하한은 1.5이고 **여기에 숫자를 적은 뒤** R20이 그것을 단언한다

      **먼저 유도가 인용한 네 값을 현재 HEAD에서 대조했다** — 베낀 것이 아니라 읽었다:

      | 항목 | 값 | 선언 자리 | 확인 |
      |---|---|---|---|
      | 시도 횟수 | 3 | `DefaultCriticalAttempts` `internal/obs/notifier.go:45` | ✓ |
      | 한 시도의 상한 | 10초 | `internal/obs/ntfy.go:95`의 **리터럴** (`Timeout` 0일 때) | ✓ — 상수가 아니다 |
      | 시도 사이 대기 | 2초 | `DefaultRetryDelay` `internal/obs/notifier.go:48` | ✓ |
      | 쓰기 대기 | 5초 | `defaultBusyTimeout` `internal/journal/journal.go:32` | ✓ |

      산수도 폈다 — `3×(10+5)=45`, `2×2=4`, `+5` → **`bound = 54초`**.

      **정한 값:**

      ```text
      margin = 1.5                        ← 하한을 그대로 쓴다
      lease  = ceil(54 × 1.5) = 81초      → journal.DefaultAlertLease = 81 * time.Second
      ```

      > **하한을 쓰는 것이 이 경우엔 기본값이 아니라 옳은 선택이다.** `bound` 54초는
      > 이미 **최악 경로**다 — 세 시도가 전부 10초를 꽉 쓰고, 세 번의 실패 기록과
      > 해제가 전부 5초 busy timeout을 꽉 쓰고, 재시도 대기 둘이 전부 붙는 경로다.
      > 이미 비관적인 값에 여유를 더 얹어도 사는 것은 적고, **치르는 값은 크다**:
      > 임차가 길수록 **발송자가 죽었을 때 그 행이 잠겨 있는 시간**이 그만큼 길어진다.
      > 그 행은 critical 알림이고 — 지금 272210이 만들고 있는 바로 그 종류다 —
      > 늦게 도착하는 알림은 중복 알림보다 나쁘다. 27초 여유면 스케줄링·GC·
      > 시계 해상도를 덮는다.
      >
      > ⚠ **`lease`를 81초로 적는 것과 유도를 코드에 두는 것은 다르다.** `journal`은
      > `obs`를 import할 수 없으므로(역방향이다) `DefaultAlertLease`는 리터럴일
      > 수밖에 없다. 그것이 조용히 짧아지지 않게 하는 것은 **R20뿐이다** —
      > `DefaultAlertLease > obs.DefaultAlertDeliveryBound()`. 그래서 §4.1b가
      > `DefaultAlertDeliveryBound()`를 **공개**로 만들어야 하고, R20이 그 함수를
      > 읽어야 한다. 54나 81을 테스트에 다시 적으면 그 테스트는 유도를 안 지킨다.
- [x] 3.4b `:skew` 값을 정한다. **하한은 2초**다 — RFC3339 절삭 1초 + 여유 1초(C2).
      만료 경계는 등호 포함, 역행 경계는 엄격 부등호

      **정한 값: `skew = 2초`** — 하한 그대로.

      > **이 술어가 어느 방향으로 여는지 먼저 확인했다.** `claimed_at > :now + :skew`는
      > C2의 OR 사슬에 있고, 참이면 그 행을 **claim 가능하게 연다** — 발급 시각이
      > 미래면 시계가 역행한 것이고 저장된 만료를 믿을 수 없기 때문이다.
      > 그래서 **skew를 키우면 덜 열린다**(미래로 보이는 행이 줄어든다). 즉 skew는
      > 「탈취를 얼마나 참을 것인가」가 아니라 **「시계를 얼마나 못 믿을 것인가」**다.
      >
      > 오늘 두 발송자는 **같은 프로세스**다(사용자 결정 9-2 — 보조 실행자를 engine
      > 안에 둔다). 같은 프로세스면 시계가 하나이므로 실제 skew는 0이고, 하한 2초를
      > 지배하는 것은 **RFC3339 초 절삭뿐**이다. 프로세스가 갈라지면 이 값을 다시
      > 유도해야 하고, 그 유도는 이 change가 아니라 그 change의 것이다.
      >
      > 경계 부등호는 C2 그대로다 — 만료는 `<=`(여는 방향), 역행은 **엄격 `>`**
      > (방금 발급한 임차가 자기 규칙에 걸리면 안 된다).
- [ ] 3.5 **합격 기준을 먼저 선언하고 그다음 잰다** (1라운드 A-P8):

      | 무엇 | 값 |
      |---|---|
      | 대상 | `claimAndDeliver` 진입 → `deliver`의 첫 `Publish` 호출까지의 체류 |
      | 기준선 | a099 **이전** HEAD에서 같은 지점 |
      | 허용 회귀 | **p99 +100ms · 최댓값 +1초** (중앙값 참고치 +10ms) ← 재기 **전에** 적었다 |
      | 조건 | 다른 쓰기가 같은 연결을 쓰는 상태 포함 (`SetMaxOpenConns(1)`) |

      > **이 숫자가 어디서 왔는지**: a099가 이 구간에 더하는 것은 **쓰기 트랜잭션
      > 하나**(claim UPDATE)다. 경합이 없으면 SQLite 쓰기는 밀리초 이하이고,
      > 경합하면 `SetMaxOpenConns(1)` 때문에 **앞선 쓰기 뒤에 줄을 선다** — 그
      > 줄서기가 p99를 만든다. 그래서 중앙값이 아니라 **p99를 게이트로 삼는다.**
      > 최댓값 +1초는 `defaultBusyTimeout` 5초의 1/5이다 — 여기 닿으면 줄이 아니라
      > **막힌 것**이고, 그때는 이 설계가 기각이다.
      >
      > ⚠ **아직 재지 않았다.** 위 숫자는 선언이고 측정이 아니다. §5에서 재고
      > 그 출력을 여기 붙인다. **재기 전에는 불변식 4를 만족한다고 적지 않는다.**

      **넘으면 이 설계는 기각이고 임차 취득을 정지 임계 경로 밖으로 옮긴다.**
      「네트워크보다 작다」는 안전 기준이 아니다. **재기 전에는 불변식 4를
      만족한다고 적지 않는다**
- [x] 3.6 R1·R3는 **동시성 테스트**다. `-race`로 돌린다(High-risk는 race/crash 테스트 필수)

      **관측 (2026-08-11, `go test -race`) — 넷 다 `-race`로 돌렸고 넷 다 빨갛다.**
      `DATA RACE` 경고는 **0건**이다.

      > **경고가 없다는 것이 이 change의 논거다.** 두 발송자가 같은 행을 다 보내는데
      > race detector는 할 말이 없다 — **경합이 아니라 설계가 둘을 다 통과시킨다.**
      > `-race`가 잡는 것은 동기화 없는 메모리 접근이고, 여기서 둘은 각자
      > **자기 뮤텍스를 올바르게 잡고** 원장에게 물은 뒤 각자 답을 받는다.
      > 배제가 없는 것이지 동기화가 깨진 것이 아니다.
      >
      > ⚠ `-race`는 이 패키지에서 **비용이 크다** — `internal/journal`이
      > `-race` 없이 331초인데 `-race`로는 20분을 넘겼다(§3.1). `-race` 실행에는
      > `-timeout`을 따로 넉넉히 준다.
- [ ] 3.7 **진입 게이트의 다섯 자리가 안 바뀌었음을 R23이 고정한다**(C6의 표).
      3판은 이 자리에 *"R5가 프로덕션 동작 변화를 고정한다"*를 적었다 —
      **사용자 결정 5-1로 그 동작 변화 자체가 없어졌다.**
      대신 R23이 *"경합은 게이트를 잠그지도 풀지도 않는다"*를 고정하고,
      §5.3이 다섯 자리의 diff가 비어 있음을 확인한다

## 4. GREEN — 최소 구현

**모든 값은 design D-CORE가 정본이다.** 아래 task는 그것을 가리키고 다시 적지 않는다.

- [ ] 4.1 `schemaV31` — `alert_outbox`에 **열 넷**(design C1): `claim_token`·`claimed_by`·
      `claimed_at`·`claim_expires_at`. `SchemaVersion` 30 → 31.
      design D6의 additive 규칙 4행 대조를 커밋 메시지에 남긴다
- [ ] 4.1b `Options.AlertLease`를 더하고 `Open`이 비면 `DefaultAlertLease`를 채운다
      (design C4). **`obs.DefaultAlertDeliveryBound()`도 같이 만든다** — R20이
      하드코딩 없이 유도를 단언하려면 그 값이 공개여야 한다 (4라운드 B).
      **`Open`이 값을 저장했는지도 단언한다** — 상수만 맞고 배선이 빠지면 lease 0으로
      claim해 임차가 즉시 만료된다 (4라운드 A-P5). `journal.go:139-142`의 `BusyTimeout` 자리 **바로 옆**이고 같은 모양이다.
      **a099는 값을 주입하지 않는다** — 주입은 a098이다(§7.4)
- [ ] 4.2 마이그레이션 테스트 — v30 DB를 만들고 v31로 올린 뒤 **기존 행이 claim 가능**임을
      확인한다(`claim_token=''`·`claim_expires_at IS NULL`)
- [ ] 4.3 `ClaimAlertForDelivery(ctx, a, remindAfter, claimant)`가 임차를 **CAS로**
      취득하고 **`ClaimResult`를 돌려준다** (design C3).
      이 단계의 취득 조건은 **`claim_token = ''`뿐이다** — 만료·역행은 4.3d·4.3e다.
      취득 UPDATE는 **네 열을 함께 쓴다**(C1).
      **⛔ 취득 실패는 `owed=false`가 아니다** — `ClaimHeldElsewhere`다
- [ ] 4.3d **만료 술어를 더한다** — `claim_expires_at <= :now` (design C2).
      `:now`는 원장이 만들고 만료는 **행에 저장된 값**을 읽는다.
      **판정자의 `AlertLease`는 이 술어에 안 들어간다** (사용자 결정 6-1).
      **R9가 이 직전에 빨갛다**
- [ ] 4.3e **역행 술어를 더한다** — `claimed_at > :now + :skew` (design C2).
      경계는 **엄격 부등호**이고 skew는 §3.4b가 정한다. **R10이 이 직전에 빨갛다**
- [ ] 4.3b **`claimAndDeliver`가 `ClaimHeldElsewhere`를 받으면 발송을 건너뛴다**(C6).
      `owed && !sent`(`notifier.go:217`)는 그것을 못 본다 — 그 조건도 같이 본다.
      **⛔ 진입 게이트를 아무 방향으로도 안 건드린다** — `Gate.Block`도 `Gate.Clear`도
      부르지 않는다 (2라운드 A-P1 · 3라운드 A-P1 · 사용자 결정 5-1).
      남기는 것은 **구조화 로그 한 줄**뿐이다(§4.5c)
- [ ] 4.4 `ClaimAlertByID(ctx, id, claimant)` 신설 (design D7·C3).
      **`lease`를 인자로 받지 않는다** — 임차 기간은 원장의 것이다 (C4)
- [ ] 4.4b `recordAlert`를 비공개로 빼고 `EnqueueAlert`가 그것만 부른다 (design D13).
      **플래그 인자로 가르지 않는다.** 안 그러면 `replay.go:551`이 쓴 행이
      **아무도 안 푸는 임차 뒤에 갇힌다**
- [ ] 4.5 정산 셋을 **`SettleOutcome`으로 가른다** (design C3 · D3의 ⚠⚠):
      `MarkAlertDelivered` 성공 시 임차를 푼다 ·
      `MarkAlertAttemptFailed` **임차를 유지한다** ·
      `ReleaseAlertClaim` **신설** — 포기할 때만 푼다.
      셋 다 **토큰**을 CAS 술어에 넣는다(이름이 아니다 — ABA).
      **0행일 때 같은 트랜잭션 안에서 한 번 더 읽어** 네 갈래를 가른다.
      **호출자의 행동은 C3의 계약 표가 정본이다** — 특히
      `SettleAlreadySettled`는 **오류가 아니고**(운영자 승인이 정상 경로다)
      `SettleNotFound`는 **오류다**.
      기존 호출 넷을 고친다 — `notifier.go:356`·`:384`·`:452`·`:455`
- [ ] 4.5b `ReleaseAlertClaim`을 **두 자리**에서 부른다
      (내가 셋으로 늘렸고 1라운드 B-P3이 하나를 되돌렸다):

      | 자리 | 상황 | 해제 |
      |---|---|---|
      | `deliver` 이탈 `:406` 직전 | 예산 소진 | **푼다** |
      | `Flush`의 행마다 (D7) | 그 행의 일이 끝났다 | **푼다** |
      | `deliver` 이탈 `:381` 직전 | publish 성공 + **정산 실패** | **⛔ 안 푼다** — 풀면 다음 관측이 다시 보내 **a096 폭풍이 돌아온다.** 임차가 억제 표시로 남고 만료가 푼다 |
      | `deliver` 이탈 `:358` | 성공 | `MarkAlertDelivered`가 정산과 함께 푼다 |

      **루프 안(B8~B10)에서 부르면 안 된다**: `:387-391`이 아직 재시도한다
- [ ] 4.5c 「전송 실패」와 「임차 상실」과 「만료 탈취」를 **서로 다른 로그 이벤트**로
      낸다 (design D12). 오늘 `deliver:385`와 `Flush:452`가 전부 같은 자리로 흘려보낸다.
      **이벤트를 내는 것은 obs다** — `Journal`에는 sink가 없고, 그래서 원장은
      `ClaimResult.Stole*`·`SettleOutcome`으로 **사실만 돌려준다**(C3).
      **로그에 넣는 것**: 이벤트 이름 · 행 id · `claimed_by` · 임차 나이.
      **넣지 않는 것**: **토큰 원문** · 계좌 · 알림 본문 · payload (불변식 8) —
      토큰은 배제의 근거이고, 새면 그것을 읽은 프로세스가 남의 임차를 정산할 수 있다
- [ ] 4.6 재무장 UPDATE(`outbox.go:229-236`)가 **임차 열 넷도 초기화**한다 —
      값은 **C5의 「네 열 초기화」가 정본**이다.
      새 episode는 새 발송이고, 이전 episode의 임차를 물려받으면 안 된다.
      `:198-201`의 주석이 그 원칙을 이미 적는다 — **임차 열은 그 원칙의 예외가 아니다**
- [ ] 4.7 `Flush`가 행마다 `ClaimAlertByID`를 부르고 `ClaimHeldElsewhere`인 행을 건너뛴다.
      **`n.mu`의 구간은 안 건드린다** — 그것이 a092
- [ ] 4.8 `outbox.go:165-168`의 문장을 지운다 —
      *"Exclusion against a concurrent claimer is the caller's"*.
      **그 문장이 이 change가 반증한 것이다.** 지우고 원장이 무엇을 보증하는지,
      그리고 **무엇을 보증하지 않는지**(C7) 적는다
- [ ] 4.9 `claimAndDeliver`의 doc `notifier.go:230-234`도 같은 이유로 고친다.
      **뮤텍스는 그대로 두되** 무엇이 배제를 지는지는 이제 다르다
- [ ] 4.10 `Alert`·`alertSelect`·`scanAlerts`에 **임차 상태를 투영한다** (design D12).
      운영자의 밀린 목록이 *"누가 언제부터 언제까지 들고 있는지"*를 보여야 한다 —
      `claimed_by`·`claimed_at`·`claim_expires_at`.
      **토큰 원문은 투영하지 않는다.** **R25가 이 직전에 빨갛다**
- [ ] 4.11 **RED 열다섯(R1~R4·R7·R9·R10·R13~R18·R20·R25)이 GREEN이고,
      회귀 핀 아홉(R6·R8·R11·R12·R19·R21~R24)이 계속 GREEN이다**

## 5. VERIFY

- [ ] 5.1 `go test ./...` — upstream 회귀 없음
- [ ] 5.2 `go test -race ./internal/journal/... ./internal/obs/...`
- [ ] 5.3 **읽기 경로의 술어가 안 바뀌었음을 diff로 확인** (design D1의 표 · C5) —
      `UndeliveredCount`·`PendingAlerts`·`AcknowledgeAlert`의 `WHERE`가 그대로
- [ ] 5.3b **진입 게이트 다섯 자리의 diff가 비어 있음을 확인한다** (design C6의 표 1~5행):
      `notifier.go:261-262`·`:378-379`·`:403-404`·`:481-482`·`:510-511`.
      **사용자 결정 5-1은 이 다섯을 하나도 안 건드린다.** 하나라도 바뀌면
      승인된 규범을 건드린 것이고 `MODIFIED` 없이는 못 나간다.

      > **⚠ C6의 표에는 이제 여섯째 행이 있다 — 그것은 a099가 안 진다.**
      > 사용자 결정 8-1이 배포 단위에 **「배달 실행자가 죽으면 `Block`」**을 더했고,
      > 그 자리는 **a098의 `internal/execgw` 편집**이다(a098 D8.2 · task 4.3.3).
      > **이 task가 재는 것은 다섯뿐이고**, 여섯째가 a099의 diff에 나타나면
      > **범위 위반**이다 — 5.4가 `internal/execgw`의 빈 diff로 그것을 잡는다
- [ ] 5.4 `internal/app/engine`·`internal/execgw`·`cmd/`의 diff가 **비어 있음**을 확인.
      a099는 journal과 obs만 건드린다. **살아 있는 설정 주입(C4)과 `tossctl`은 a098이다**
- [ ] 5.5 **만료 판정이 `claim_expires_at`만 읽는지 확인한다** —
      `rg 'AlertLease' -- internal/journal`의 결과가 **`Open`의 기본값 채우기와
      claim UPDATE 한 자리뿐**이어야 한다. 판정 술어에 lease가 나타나면
      3판의 결함이 살아 있는 것이다 (사용자 결정 6-1 · 3라운드 A-P5).
      **역행 술어만 `claimed_at`을 읽는다**
- [ ] 5.6 편집 후 AST를 다시 뽑고 Map을 최신화한다 (WORKFLOW.md §Function Logic Map)
- [ ] 5.7 **3.5의 지연 값을 GREEN 이후에 다시 잰다.** 3.5는 예측이고 이것이 실측이다

## 6. Gate

- [ ] 6.1 `make sdd-sync`
- [ ] 6.2 `make sdd-check`
- [ ] 6.3 `openspec validate a099-a-claim-excludes-the-second-sender --strict --no-interactive`
- [ ] 6.4 `make gate CHANGE=a099-a-claim-excludes-the-second-sender`
- [ ] 6.5 독립 리뷰 (적대적 Eng 관점 필수 — High-risk 원장 스키마).
      **4라운드다.** 1·2·3라운드가 각각 BLOCK했다
- [ ] 6.6 **⛔ 배포 제약 — a099를 혼자 배포하지 않는다** (사용자 결정 3 · design D9).
      `Flush`의 프로덕션 호출자가 0이므로 claim 뒤 죽은 행을 **다시 집을 주체가 없다.**
      a098과 **같은 배포 단위**로 나간다. `SchemaVersion` 31이므로 콘솔과 엔진이
      같은 빌드여야 한다 — 낮은 버전 바이너리는 이 DB를 거부한다(설계이고 옳다)
- [x] 6.7 **짝 gate를 기계가 강제하게 만든다** (사용자 결정 7-1 · design D9의 ⛔).
      3라운드가 *"묶은 것은 문장뿐"*이라고 두 보이스에서 각각 잡았다.

      | 무엇 | 어디 |
      |---|---|
      | 짝 선언 | `openspec/changes/<id>/deploy-pair.txt` — 한 줄에 change id 하나, `#`은 주석 |
      | 검사 | `tools/gate.sh`의 **새 3단계** — 짝 디렉터리 존재 · 짝의 미완료 task 0건 · **상호 선언** |
      | 교착 예외 | 짝의 미완료 중 **`make gate`를 포함한 줄은 안 센다.** 안 그러면 두 gate가 서로를 기다린다 |
      | 선언이 없으면 | 그 단계는 **통과한다** — 기존 change 전부에 영향이 없다 |

      **선언 파일 둘을 이 change에서 만든다** — a099와 a098 양쪽. 한쪽만 두면
      상호 선언 검사가 잡는다.

      **실측 (2026-08-11, 임시 change 둘로 여섯 경우를 돌렸다):**

      | 경우 | 결과 |
      |---|---|
      | 정상 짝 (짝의 `make gate`만 미완료) | **3단계 통과** → 4단계에서 멈춤 |
      | 짝에 `make gate` 아닌 미완료 1건 | `GATE FAIL … 아직 완료가 아닙니다` |
      | 상호 선언 없음 (한쪽만 적음) | `GATE FAIL … 짝으로 선언하지 않았습니다` |
      | 짝 디렉터리가 없음 | `GATE FAIL … tasks.md 가 없습니다` |
      | 주석만 있는 빈 선언 | `GATE FAIL … 빈 선언은 선언이 아닙니다` |
      | 선언 파일 없음 | `OK: 짝 선언 없음` — **기존 change에 영향 0** |

      `bash -n tools/gate.sh` 통과. 임시 change 둘은 삭제했다

## 7. 후속 change에 넘기는 것

- [ ] 7.1 **a092의 D0.3·§8.7·D7·R17-3을 재정렬한다.** a099가 그 전제를 만들었으므로
      *"배제가 SQL 술어로 옮겨간다"*는 이제 참이다 — 그러나 **a099 뒤에** 참이다.
      `a092/analysis/function-logic/internal-journal--journal.markalertdelivered/`의
      **17판 잔재**도 같이 정리한다 (design D0의 ⚠⚠)
- [ ] 7.2 **a098의 공존 이야기가 a099를 전제로 한다고 적는다.** 배달 루프와
      `claimAndDeliver`를 가르는 것은 임차다
- [ ] 7.3 a092 19라운드의 A-P3 = B-P1(*"어느 착지 순서도 안전하지 않다"*)을
      **해소로 기록한다** — 순서를 고른 것이 아니라 단계를 다시 나눠서 없앴다
- [ ] 7.4 **a098이 지는 것 넷** (사용자 결정 3·4로 확정):

      | 무엇 | 왜 a098인가 |
      |---|---|
      | 만료된 임차를 다시 집는 주체 | 배달 루프가 주기마다 PENDING을 훑는다 |
      | **기동 직후 `Block` 복원** (C6) — **`Clear`는 복원하지 않는다** | 엔진의 **`Recover` 단계**에서 원장을 읽는다. 루프의 첫 tick은 **늦다** — 그 사이에 진입이 열린다 (3라운드) |
      | **살아 있는 설정에서 임차 기간 주입** (C4) | engine 배선이 a098에 있다. **C7의 보장이 여기에 달려 있다** |
      | **`tossctl` 운영자 표면** (목록·승인) | `Acknowledge`의 프로덕션 호출자가 0이다 |
- [ ] 7.5 **a092가 져야 하는 것 넷** (6판 신설 — 사용자 결정 9-2·11-1·12-2, 2026-08-11).
      **셋은 「미배정」이 아니라 「반대」다** — a098의 델타와 a092의 델타가 **서로 반대를
      적고 있는 상태**이고, 둘 다 archive되면 정본에 **모순된 SHALL 둘**이 남는다.

      | a092의 자리 | 무엇을 적고 있나 | a098과의 관계 | 종류 |
      |---|---|---|---|
      | `specs/engine-safety/spec.md:24-25` | *"「엔진 런타임 수명주기」는 루프를 **하나 더한다** — 알림 배달 루프"* | 결정 9-2는 **안 더한다** (보조 실행자다) | **반대** |
      | 같은 파일 MODIFIED 「등급화된 알림」 | *"그 실행자는 엔진의 **루프 감독 아래에** 있어야 하며(SHALL)"* | 결정 9-2가 **감독 밖**에 둔다 | **반대** |
      | 같은 파일 MODIFIED 「등급화된 알림」 | *"그 정지는 진입 게이트 래치와 **운영 모드 승격**으로 이어져야 하고(SHALL)"* | 결정 11-1이 **승격을 뺐다** | **반대** |
      | 같은 파일 `:126-130` Scenario 「배달 실행자가 죽는다」 | 래치 사유가 `ALERT_UNDELIVERED` + 모드 승격 | 결정 8-1은 **자기 사유 코드**, 11-1은 **모드 없음** | **반대 + 미배정** |

      > **결정 12-2가 고른 것은 「a092가 진다」이지 「무시한다」가 아니다.**
      > a092는 a098보다 **뒤에 착지**하므로 그 시점에 자기 델타를 정합하게
      > 고칠 수 있다. **고치기 전에는 a092의 gate를 통과시키지 않는다** —
      > 통과시키면 정본에 모순이 들어간다.
      >
      > **a098은 이 넷을 자기 change에서 고치지 않는다.** 남의 change의 델타를
      > 고치면 그 change의 base-commit·리뷰 기록과 어긋나고, 그것이
      > `stacked-changes-break-the-gate`가 적는 형태다.

## 안 하는 것 — 이름 붙여 둔다

| 무엇 | 어디로 |
|---|---|
| `n.mu` 구간 축소 | a092 |
| 배달 루프 신설 · 운영자 CLI · 기동 복원 · 설정 주입 | a098 (§7.4) |
| `Flush`·인라인 재시도·`wait`·죽은 상수 제거 | contract (3번째 배포 단위) |
| 알림 등급 이관·예산 상수 | a092 |
| `SetModeProjector` 배선 | 미배정 후속 (a092가 발견) |
| **정확히 한 번 발송** | **안 한다.** 유계 실행 밖에서는 at-least-once다 (design C7) |
| 전송 수단 idempotency 키 | ntfy에 없다. 넣으려면 전송 계약 자체가 바뀐다 (design C7·D8) |
