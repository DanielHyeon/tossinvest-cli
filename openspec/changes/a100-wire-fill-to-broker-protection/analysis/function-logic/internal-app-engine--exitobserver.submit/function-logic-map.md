# Function Logic Map: `ExitObserver.submit`

- Source: `internal/app/engine/exitloop.go` (L1237-1312)
- AST evidence: `ast.json` — 분기 11, return 9
- Risk scan: `risk-pattern-report.md`

## 이 함수가 대상인 이유

a087 인프로세스 보호성 매도가 **실제로 발행되는 곳**이다. design D8-3은 "상주 주문이 먼저
체결되어 포지션이 이미 flat이면 보호 청산을 시도해서는 안 되고, 수량 0을 제출 경로로
흘려보내서도 안 된다"고 정했다. 그 판정이 여기 B2에 이미 있다. **High-risk이며 면제할 수
없다**(tasks 0.4).

## 이미 존재하는 것 — 수량 0은 이미 제출되지 않는다

B2(L1243)가 `isZeroQuantity(submitQuantity)`를 잡아 `release(ProposalRefused)`로 끝낸다.
주석(L1244-1248)이 이유를 적어 뒀다 — 확정된 floor가 아무것도 허가하지 않는 상태(RECONCILE 중
스냅샷 부재·노후)이며, arming을 유지하면 exit policy가 요구하는 재제안을 오히려 막는다.

⇒ **spec의 「수량 0을 제출 경로로 흘려보내지 않는다」는 이미 참이다.** a100이 추가해야 할 것은
그 앞 단계다 — **상주 주문이 먼저 체결되어 flat이 된 경우를 「보호 실패」가 아니라 「보호됨」
으로 구별해 기록하는 것.** 지금은 그 경우도 `ProposalRefused`로 뭉뚱그려진다.

## `applyFloor`가 이미 보유 수량으로 상한을 건다

B1(L1239-1242)의 `applyFloor`가 제출 수량을 계좌 상태로 제한한다. 상주 주문이 일부 체결시켜
보유가 줄었으면 그 감소가 여기에 반영된다 — **단, 원장이 그 체결을 알고 있을 때만.**
design D8의 child 귀속 문제(어댑터가 `TriggeredOrderID`를 bool로 접는다)가 해결되지 않으면
`applyFloor`는 **줄어들기 전 수량**을 통과시킨다. ⇒ 귀속은 편의가 아니라 이 함수의 정확성
전제다.

## 상태 4분기 — 그리고 어느 것도 상주 주문을 모른다

`Place` 결과는 4갈래다(B7~B10).

| 결과 | 처리 | arming |
|---|---|---|
| `StateConfirmed` (B7) | 로그 | 유지(정상 종료) |
| `InDoubt` / `UnresolvedInDoubt` (B8) | 아무것도 하지 않음 | **유지** — 주석 L1297-1299: 놓으면 다음 관측이 이미 살아 있을 수 있는 매도 위에 두 번째 매도를 낸다 |
| `ReasonSymbolInFlight` (B9) | `noteDelay` | 해제(`ProposalCancelled`) |
| 그 외 (B10) | `alertProposalRefused` | 해제(`ProposalRefused`) |

⇒ **B8의 논리가 a100에 그대로 적용된다.** 상주 주문 취소가 in-doubt면 마찬가지로 "취소됐는지
모른다"이고, 그 상태에서 재등록하면 브로커에 매도 청구권이 둘이 된다. **상주 주문의 in-doubt는
B8과 같은 방식으로 다뤄야 한다 — 놓지 말고 다음 주기가 관측한다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `quantity` | 제안 수량 | `record`가 계산 | `applyFloor`가 상한 |
| `intentID` | arming된 intent | `record`의 `recorded.ArmedProposal` | `AttachExitIntent` 실패 시 제출 중단(B4) |
| `m.position.Quantity` | 원장의 보유 | journal | 귀속 누락 시 과대 |
| `o.opts.Submit` | 실행 게이트웨이 | 조립 | — |

**불변식 — attach는 제출보다 먼저다**(주석 L1268-1271). "an attach that fails stops the
submission rather than leaving an order nothing can resolve." a100이 상주 주문 취소를
넣는다면 같은 원칙을 따른다: **해소할 수 없는 것을 만들지 않는다.**

## Branches and early returns — 측정 포함

`go test -covermode=set ./internal/app/engine` 기준. 판정은 각 분기 **본문 블록**의 실행 여부다.

| Branch | 조건 (L) | 결과 | 측정 |
|---|---|---|---|
| B1 | `applyFloor` 에러 (1240) | 에러 반환 | **미실행** |
| B2 | `isZeroQuantity` (1243) | `release(Refused)` | 실행됨 |
| B3 | `IssueReduction` 에러 (1263) | 알림 + `release(Refused)` | **미실행** |
| B4 | `AttachExitIntent` 에러 (1272) | **제출 중단** | **미실행** |
| B5 | `sellIntent` 에러 (1277) | 알림 + `release(Refused)` | **미실행** |
| B6 | `switch` (1287) | — | 블록 없음(n/a) |
| B7 | `StateConfirmed` (1288) | 로그, 성공 | 실행됨 |
| B8 | in-doubt (1296) | **arming 유지** | 실행됨 |
| B9 | `ReasonSymbolInFlight` (1301) | `noteDelay` + 취소 | **미실행** |
| B10 | default (1304) | 알림 + `release(Refused)` | 실행됨 |
| B11 | `detail == "" && err != nil` (1306) | 에러 문구 대체 | **미실행** |

**미실행 6건.** 손절 발행 함수의 **실패 경로 대부분이 한 번도 실행된 적이 없다.** a100은 이
함수에 상주 주문이라는 새 실패 원인을 더한다 ⇒ 미검증 실패 경로 위에 새 실패 원인을 얹는
것이므로, **B1·B3·B4·B5의 RED 테스트를 a100이 함께 만든다**(tasks 1절의 원칙을 이 함수에도
적용).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.applyFloor` | 확정된 계좌 상태로 제출 수량 상한 | 에러 = 즉시 반환(B1) | AST L1239 |
| `isZeroQuantity` | 0 수량 제출 차단 | 순수 함수 | AST L1243 |
| `o.opts.Issuer.IssueReduction` | 리스크 결정 발행 | 에러 = 알림 + arming 해제(B3) | AST L1251 |
| `o.opts.Journal.AttachExitIntent` | 제출 전 intent 결속 | 에러 = **제출 중단**(B4) | AST L1272, 주석 L1268-1271 |
| `o.sellIntent` | 주문 intent 구성 | 에러 = 알림 + 해제(B5) | AST L1276 |
| `o.opts.Submit.Place` | **브로커에 매도 발행** | 결과 4분기(B7~B10) | AST L1281 |
| `o.release` | arming 해제 | — | AST L1248, L1265, L1279, L1303, L1310 |

**a100이 추가할 호출은 없다.** 상주 주문 취소는 `record`에서, 상주 주문 등록·교체는 수렴
워커에서 일어난다. 이 함수는 **인프로세스 매도만** 다룬다.

## State mutations and fallbacks

- `release`는 arming을 풀어 **같은 level identity의 재제안을 허용**한다. B2·B3·B5·B10이 쓴다.
- **in-doubt는 release하지 않는다**(B8). 주문이 살아 있을 수 있으므로 arming을 유지해
  다음 관측이 두 번째 매도를 내지 못하게 한다. **a100의 상주 주문 mutation도 같은 규칙을
  따른다.**
- `capped`는 로그에만 실린다(B7). 바닥에 걸려 수량이 줄어든 사실은 알림이 아니라 로그다.
- 이 함수는 journal 상태를 **직접 쓰지 않는다** — `AttachExitIntent`와 `release`를 통해서만
  바뀐다.

## Safety conclusion

- Safe edit boundary: 「이미 flat」 구별은 **B2 앞**에서, 상주 주문 취소는 이 함수가 아니라
  `record`의 B3 블록에서(별도 산출물). 이 함수에는 **제출을 막는 새 조건을 넣지 않는다.**
- High-risk impact: **yes.** 손절 주문의 유일한 발행 지점이다.
- **in-doubt 규칙 재사용:** 상주 주문 mutation의 불확실 상태는 B8처럼 **놓지 않고 다음 주기가
  관측**한다. 놓으면 매도 청구권이 둘이 된다.
