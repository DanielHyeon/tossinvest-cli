# 보호 우선순위와 재전송 출처 — Manager 세부 계약

2026-10-31. 근거는 main `ccc39f6f137236740efa924f96fa2a8f2d953bd9`의 `ExitObserver.observe/record/submit/clearTheSymbol`, `Gateway.ReplayInDoubt`, `reconcile/recovery.go`, `exitpolicy.Action.Orderable`, journal exit_events 원문이다. Terra의 `/tmp/a166-priority-origin-map.md`, `/tmp/a166-replay-origin-decision.md`와 AST 산출물은 호출 흐름과 제약을 구별한다. 이 문서는 별도 구현 로트의 계약이며 현재 배선 완료 주장이 아니다.

## 현재 호출의 분류

보호 판단을 포함한 ExitObserver.observe의 선행 Prices 호출만 목적에 따라 critical context를 전달한다. Retrier.Query 진입 전부터 지정하고 각 retry closure의 ctx까지 이어져야 한다. 일반 시세 API 전체를 승격하지 않는다. record에서 clearTheSymbol을 호출할 때는 해당 proposal의 isProtective 결과가 참인 경우에만 critical로 전달한다. 익절도 이 함수로 오므로 clearTheSymbol 또는 모든 Cancel을 무조건 승격하지 않는다. submit도 직접 호출될 수 있으므로 isProtective가 참인 proposal의 Place 사슬에 명시 context를 준다. 보호 action 집합은 기존 두 상수의 실제 값과 일치시키며 문자열 오기를 되살리지 않는다.

flatten은 비상 청산 명령의 최초 계좌 조회·사전 확인·취소·매도 전체 사슬을 critical로 전달한다. 일반 CLI 매도와 같다고 추정하지 않는다. child timeout, token exchange, 401 재전송, ACK 확인은 부모 context의 priority와 deadline을 계승한다. token과 주문은 각각 실제 wire admission을 받으며 기존 시한을 늘리지 않는다. normal/auxiliary/critical 모두 같은 strict cap을 지키고 이미 선택된 요청을 추월하지 않는다. a087 갱신기의 production nil은 그대로다.

## 복구 replay의 분류

Source=`engine`, SafetyClass=`RISK_REDUCING`, SELL이라는 조합만으로 보호 주문을 입증하지 못한다. 실제 출처를 확인하지 못한 재전송은 critical이나 normal로 임의 추정하지 않고 기존 `ReasonReplayIneligible`·`QueryFallback`으로 관찰 복구에 넘긴다. 이는 과거 보호 주문의 복구를 지연하거나 관찰로 해결하지 못할 수 있는 보수적 선택이다. 현재 replay attestation nil은 바꾸지 않는다.

이 추가 출처 판정·미입증 fallback은 budget enforcing(required) replay에만 적용한다. OFF HTTPReplay의 기존 eligibility·흐름·원장/전송 의미는 바꾸지 않으며 OFF golden으로 고정한다. required의 private budgeted replay 타입과 연결된 경계에서 적용하고 독립 플래그 누락으로 required 검사만 빠질 수 있는 구성을 만들지 않는다.

기존 terminal/unattested 및 state/kind/TTL/wire-body eligibility guard 순서는 유지한다. 추가 출처 판정은 재전송 자격을 갖춘 뒤, MarkReplayStarted와 실제 전송 전에 한다. 미입증이면 원장 상태와 durable replay count를 그대로 두고 이번 증가·전송은 0이다. 결과에 원래 count를 보존하며 임의로 0으로 덮어쓰지 않는다. DB 읽기 오류는 기존 다른 원장 오류처럼 반환하고 성공 fallback으로 숨기지 않는다. 기존 최초 409 AMBIGUOUS와 후속 NOT_SENT 처분을 보존한다.

현재 증명 가능한 정상 분류는 다음과 같다.

- engine의 persisted exposure-raising PLACE/BUY와 계좌·intent/attempt 결속이 기존 검증을 만족하면 normal이다. 단순 BUY 문자열만으로 권한을 부여하지 않는다.
- flatten-all 출처가 정확하며 PLACE/SELL/RISK_REDUCING·기존 account/intent 결속을 모두 만족하면 emergency origin으로 critical이다. Source는 원장 내부 생산자 기록이며 외부 입력의 진위 증명 수단이 아니다.
- engine의 risk-reducing PLACE는 exact intent ID의 exit event proposal action으로 분류한다. 보호 두 action이면 critical, 입증된 나머지 Orderable 익절 action이면 normal이다. 조회는 intent와 position의 실제 계좌·시장·종목 관계를 대조한다. positions에 없는 trading_day 컬럼을 가정하지 않는다.

새 journal 조회는 기존 immutable event 선례를 확장하는 leaf로 두며 schema를 추가하지 않는다. 알려진 lifecycle action만 제외한 뒤 정확히 하나의 유효 Orderable proposal이 있어야 한다. 동일 action의 중복도 미입증이다. unknown/state-only action, scope 불일치, 복수·상충, 행 누락은 미입증으로 처리한다. exact intent ID의 후보를 먼저 읽고 각각의 scope를 확인하며 SQL WHERE로 맞지 않는 scope를 조용히 걸러 모순을 숨기지 않는다. 후보 수에 제한을 두고 초과도 미입증으로 처리한다. 다른 종목에 보호 intent가 있다는 EXISTS 질의를 재사용하지 않는다. action 종류 추가 시 분류 누락을 잡는 구조 시험이 필요하다. Source와 문자열 목록만으로 전체 branch 의미를 검증했다고 보고하지 않는다.

## 구현 분리와 수용

engine exitloop의 context 및 execgw replay/journal origin leaf는 root factory 배선과 분리한다. flatten command와 engine gateway wiring은 root caller 로트가 고정된 뒤 순차 수정한다. required gateway는 concrete official client와 borrowed Budget으로 NewBudgetedHTTPReplay를 구성하고 explicit owned transport를 넘긴다. OFF의 기존 HTTPReplay와 attestation gate는 유지한다.

시험은 보호/익절 취소 구분, 토큰/ACK의 context 계승, 신규 진입 normal, flatten 선행 조회 critical, queue 경합과 취소, 정확 origin·외계좌·동일 action 중복·unknown/lifecycle·DB 오류, 기존 replay count·stored body·첫 409 보존을 다룬다. 독립 local TLS fixture와 fake journal만 사용한다. source/AST/FLM/BTM·RED/GREEN·변이·targeted normal/tag/race·gstack 수동 검토 및 독립 검증 뒤 별도 SHA로 완료를 판정한다. 전체 시험/gate·실계좌·운영·활성화는 포함하지 않는다.

search-first: 기존 isProtective/Action.Orderable·context·journal exact-intent 조회와 기존 QueryFallback을 확장한다. 새 dependency·durable priority 열·global queue 우회는 없다. grill-me/grilling: 확정된 strict cap·대기 정책을 유지하고 출처 미입증을 관찰 복구로 처분한다. impeccable N/A — backend/documentation-only.

독립 설계 검토에서 required-only 적용 누락을 지적했고 Manager가 OFF 보존과 private required 타입 결속을 추가했다. observe의 Query 진입 전 context, scope 불일치 후보를 숨기지 않는 조회도 반영했다. 수정 후 Terra 독립 재검토에서 기존 P1/P2가 해소되고 조건부 PASS임을 확인했다. Manager는 이 계약의 구현 착수를 승인한다. source/AST·fixture·변이 및 독립 구현 검증은 별도 수용 조건이며 기존 후보가 완료됐다는 뜻이 아니다.
