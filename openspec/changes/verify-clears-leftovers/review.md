# Review: verify-clears-leftovers

Function Logic Map: applied — `analysis/function-logic/` 6 target, `check_analysis.py` 통과.

## 1. Proposal freeze (Eng)

**전제를 실제 기록으로 재검증했다.** 제안이 "도구가 자기 잔여물에 갇혔다"고 주장했으므로,
구현 전에 사용자의 실제 US 기록(19줄)을 그대로 읽어 대상 선정 로직을 돌렸다:

```
outstanding: kind=order id=OsBakhtsu…afmp1j2 symbol=MWG deliberate=false
redo set:    [idempotency-ttl-edge order-cancel order-amend sell-boundary]
cleanup target: order OsBakhtsu…afmp1j2 (MWG)
cleanup line:   step=cleanup kind=cancel-order symbol=MWG maxQty=0
```

즉 이 change는 가상의 상태가 아니라 **지금 계좌에 살아 있는 객체 1건**을 대상으로 한다.
(이 확인은 임시 테스트로 했고 커밋 전에 삭제했다 — 실계좌 기록을 읽는 테스트를 저장소에
남기지 않는다.)

## 2. Code review

- **승인 모델이 그대로인가**: 정리도 `PlannedMutation`이고, `r.gate` → `r.authorise` →
  `Plan.Authorises`를 통과한다. 목록에 없으면 `ErrOutsidePlan`으로 전송되지 않는다.
  "취소니까 그냥 보낸다"는 경로를 만들지 않았다 — 그것이야말로 배치 승인이 막으려는 것이다.
- **순서가 승인 뒤인가**: prologue는 `approveBatch` **다음** 줄에서 시작한다. 앞이면 승인
  없는 라이브 요청이 된다. `TestARefusedBatchSendsNoCleanup`이 거절 시 `/cancel`이 한 건도
  나가지 않음을 브로커 요청 로그로 고정한다.
- **존속 측정을 죽이지 않는가**: 가장 위험한 실수는 prologue가 conditional-persist의 관측
  대상을 먼저 취소하는 것이었다. 규칙을 "생성 단계가 settled면 대상"으로 잡았다면 실제로
  그렇게 됐을 것이다(conditional-register는 pass다). 실제 규칙은 **취소하는 단계
  (`conditional-cancel`)가 settled일 때만** 조건주문이 대상이 된다이고,
  `TestTheConditionalLeftForPersistenceIsNotCleanedUp`이 이를 고정한다.
- **주문과 조건주문을 다르게 다루는 근거**: catalogue의 어떤 단계도 *이전 실행의* 주문을
  취소하지 않는다(각 단계는 자기가 낸 것만 취소한다). 반면 `conditional-cancel`은 기록에서
  조건주문을 찾아 취소한다 — 실제로 3차 실행이 그렇게 취소했다. 그래서 주문은 항상 대상,
  조건주문은 조건부 대상이다.
- **실패가 정직한가**: 정리 실패는 `fail`로 기록되고 실행은 계속된다. 잔여물은 계속
  `Outstanding`에 남는다(취소가 실제로 일어난 것만 상쇄하므로 구조적으로 그렇다).
  `TestAFailedCleanupIsRecordedAndDoesNotStopTheRun`이 세 가지를 함께 검사한다.
- **정리가 측정으로 오인되지 않는가**: `KindCleanup`. `isStepEntry`가 유일한 판정 지점이고
  `StepCount`가 그것을 쓴다. `RedoSet`·`BuildProgress`·리포트는 `Steps()`를 순회하므로
  카탈로그에 없는 것은 구조적으로 들어올 수 없다. `TestCleanupIsNotAMeasuredStep`이
  `StepCount`와 실제 측정 줄 수가 일치하는지 센다.
- **라벨**: `stepLabels`에 넣지 않았다 — 그 맵은 "카탈로그와 정확히 일치"가 테스트로 걸려
  있다. `StepLabel`에 분기를 두었고, 이유를 주석으로 남겼다.
- **콘솔**: `ShowStart`는 `v.Done`일 때만 참이다. 진행 중 실행에서 시작 제어가 보이면
  승인 대기 화면에서 두 번째 실행을 시작할 수 있게 된다 — 전용 테스트로 막았다.
  `Spent`와 "이어할 단계가 없다"는 종전대로 버튼을 비활성화한다(전용 테스트로 고정).

**남은 관찰**: 정리는 `preflightStatic`의 시장 일치 검사를 거치지 않는다. 의도적이다 —
그 검사는 *주문을 내는* 쪽의 규칙이고, 다른 시장 심볼이라는 이유로 취소를 거부하면 지금
고치려는 교착을 그대로 재현한다. 취소는 식별자만으로 이뤄지고 시장별 가격 규칙을 쓰지 않는다.

## 3. Security review (CSO 관점)

**새 위험**: 실행이 라이브 취소를 한 건 더 보낼 수 있다.

- 방향이 **노출 감소뿐**이다. 정리는 취소만 하고 아무것도 만들지 않는다.
- 대상이 **기록으로 한정**된다. "이 도구가 만들었다"가 기록의 사실이 아니면 대상이 아니므로,
  계좌의 다른 주문은 어떤 경로로도 대상이 될 수 없다.
- **승인 없이는 전송되지 않는다.** 상한(`MaxLiveOrders`)·1주 규칙·체결 불가 지정가·계획
  인가는 전부 무변경이다. 상한이 비는 것은 잔여물이 실제로 사라졌기 때문이지 상한을
  늘렸기 때문이 아니다.
- **실패 방향이 안전하다**: 대상 선정이 좁으면 종전처럼 사람이 처리한다(현행). 넓어도
  최악은 "이 도구가 만든 미체결 주문을 취소한다"이며, 그것은 모든 실행이 끝에 하려는 일이다.

**잔여 위험(수용)**: ① 정리가 상한까지 실패하면 주문이 계좌에 남는다 — 종전과 같고
화면·기록이 잔여물로 보고한다. ② `ErrConfirmationExpired` 문구를 바꿨다. `errors.Is`로
비교하는 곳은 무영향이고, 문자열을 비교하는 유일한 곳(`refusalNotice`)은 같은 sentinel의
`Error()`를 쓰므로 함께 움직인다.

**게이트**: 이 change는 `ProtectionReady`·automation gate·엔진 기동에 손대지 않는다.

## 4. QA

- `go build ./...`, `go vet ./...` clean. `go test ./...` **3126 passed**.
- RED 관측(순서대로): 잔여 주문이 있어도 계획에 정리 줄이 없음 → GREEN, 다음 실행이
  잔여물을 남긴 채 끝남 → GREEN, 막혀 있던 `--redo order-cancel`이 상한으로 거절됨 →
  GREEN(pass, 잔여물 0), 정리 실패가 기록되지 않음 → GREEN, 끝난 실행 뒤 콘솔에 시작
  제어가 없음 → GREEN(있음), spent 프로세스에서 시작 제어가 사라짐 → GREEN(보이되 비활성).
- 실계좌 확인은 다음 US 실행에서 이뤄진다: `[재측정]`의 승인 목록 첫 줄에 `OsBakht…`의
  취소가 보여야 하고, 승인 후 그 주문이 사라져야 한다.

## 5. 완료 조건

- 미완료 태스크 0, FLM 통과, PM check 통과, gate 8/8.
- 사용자 조치: 새 빌드 설치 후 콘솔 재시작 → `/verify?market=US` → `[재측정]`.
