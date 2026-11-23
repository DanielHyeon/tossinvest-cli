# a194 Design

## 1. 결정

> D1·D3·D5·D6 과 §4·§5·§6 은 1단계 리뷰(review.md R1-1~R1-8) 뒤 **Manager 결정 2026-11-23** 으로 개정했다.

| # | 결정 | 근거 |
|---|---|---|
| D1 | 「운영자만 풀 수 있는 차단」 = 그 (시장, 종목)을 덮는(`Block.Covers`) 활성 블록(`Tracker.Blocks()`) 가운데 `Release == ReleaseOperatorOnly` 인 것이 하나라도 있음. `EntryAllowed` 의 사유나 `Permanent()` 로 가르지 않는다. 본문의 원인·시각은 그 블록의 `Cause`·`Since` 에서 온다 — 덮는 운영자 전용 블록이 여럿이면 가장 이른 `Since` 의 것. 편입을 막는 판정(`blocked`)은 그대로 두고, `blocked()` 가 참인데 덮는 블록을 `Blocks()` 에서 못 찾으면(두 읽기 사이의 순간 불일치) 「스스로 풀리는 차단」과 같은 길(10 분 유예 뒤 보고, 원인 「미상」)로 보낸다 — 조용한 건너뛰기를 만들지 않는다. 새 Tracker 메서드를 만들지 않는다(**Manager 결정 2026-11-23**) | `EntryAllowed` 는 `*RejectedError{Reason, Detail}` 만 돌려주고(AST: analysis/cited-ast/EntryAllowed.ast.json B2 if@1328 → return@1329) SNAPSHOT_* 는 operator_only 인데 사유가 `ReasonReconcileMismatch` 다(analysis/code-context/evidence-reconciliation.md 「F5 상세」). `Permanent()` 는 계좌 전체 수량 승격만 뜻한다(AST: analysis/cited-ast/Permanent.ast.json · Refresh.ast.json) |
| D2 | 스스로 풀리는 차단의 유예 = **10 분**(상수 하나). 포지션별 「처음 빠진 시각」을 드라이버 메모리에 두고, 그 주기에 빠지지 않으면(편입·해제·청산) 지운다. 재시작하면 시계가 다시 0 이다 | 대사 주기 약 62 초(reconcileloop.go:95 60 초 + 안정화 2 초) × 10. 수량 불일치는 조정 뒤 한두 주기에 풀리는 것이 정상이고(운영 장부 총 134 행 중 미해제는 귀속 실패 1 건뿐 — Manager 실측 2026-11-23), 10 분을 넘기면 「스스로 풀린다」는 전제가 이미 깨진 것이다. 재시작 뒤 시계 초기화는 공시한다 — 운영자 전용 차단은 유예가 없으므로 재시작에 가려지지 않는다 |
| D3 | 보유 보고. **알림 켜짐**: 기존 critical 타입 `exit.position_adoption_failed`, 사실 칸에 새 값 둘(`blocked_operator_only`, `blocked_stale`), key 는 기존 `reconcileUnmanagedKey` 형식(타입·`reconcile`·사실·포지션), 기록 입구는 기존 `d.alert → Alerts.Notify`. 메모리 래치 없이 매 관측 `Notify` 하고 중복은 outbox key + 재알림 창(Notifier 기본 1 시간)이 맡는다. **알림 꺼짐**: 기존 `alertUnmanaged` 와 같은 방식 — 일반 등급 타입 `exit.position_unmanaged` 에 같은 새 사실 값, (포지션, 사실) 메모리 래치(`d.unmanaged`)로 한 번. 이 보고는 `cycle.Unmanaged` 를 올리지 않고 새 계수 필드도 만들지 않는다. **새 보고(보유·차단)는 `judgeHoldings` 의 끝 — `adopt` 와 기존 무관리 보고 뒤 — 에서 내고, 한 주기의 보유 보고 `Notify` 는 최대 3 건(상수 하나)이다.** 넘는 보유는 다음 주기에 낸다(보유가 남아 있으면 다음 주기가 다시 본다). 순서는 결정적: 처음 빠진 시각 오름차순, 같으면 포지션 id. 차단 보고는 블록 수만큼. 이 상한은 알림을 줄이는 규칙이 아니라 한 주기의 지연 상한이다 — 재알림 창 판정은 `Notify` 안에서 일어나므로 상한은 `Notify` 호출 수에 건다(**Manager 결정 2026-11-23**) | 보호를 기대한 보유가 보호되지 않는다는 같은 사실이다. `internal/obs` 무편집. 두 타입은 등급표에서 항상 critical 이라(event.go:370·:386) 꺼진 엔진에서 critical 타입을 쓰면 outbox 행 · 진입 래치 · 모드 승격이 생긴다(R1-1). `cycle.Unmanaged` 는 「전이 상태를 빼고 센 무관리 보유」이고 `adoption_include_test.go:133` 이 그 값에 묶여 있다. 대사 루프의 critical `Notify` 는 동기 전송이다(notifier.go:354 — 기존 `alertAdoptionFailed` 와 같은 성질, 새 배선 없음). 전송이 막힌 최악의 경우 주기당 추가 지연 상한 = (3 + 블록 수) × 전송 한계(보고 하나 약 34 초 — 전송 10 초 × 3 회 + 사이 대기 2 초 × 2, r2 R2-2). 끝에서 내므로 그 주기의 편입은 밀리지 않는다 |
| D4 | 편입이 꺼지고 그 종목이 지정되지도 않은 엔진·제외 목록 종목은 이 보고의 대상이 아니다 — 그 보유는 차단이 없어도 편입되지 않으므로 「차단 때문에 보호가 빠졌다」가 거짓이다. 그 경우는 기존 무관리 보고(일반 등급)가 차단 해제 뒤에 나간다. 판정은 같은 함수의 기존 술어(`Adoption.Excludes` · `!Enabled ∧ !Included`)를 차단 분기에서 먼저 부른다 | engine-safety 「무관리 보유 보고의 등급은 사실이 정한다」의 「운영자가 고른 상태」와 같은 방향. AST: analysis/cited-ast/judgeHoldings.ast.json B12 if@163 · B13 if@171 |
| D5 | 차단 보고. **알림 켜짐**: 기존 critical 타입 `reconcile.mismatch_permanent`, 기존 `d.alert → Alerts.Notify`(새 배선 없음 — `RecordCritical` 을 끌어오지 않고 `Context.ReconcileDriver` 무편집). 「하루 한 번」은 드라이버 메모리 래치로 만든다: key = 타입·범위·종목(계좌 전체면 빈 값)·원인·KST 날짜, outbox key 도 같은 문자열. 래치는 `Notify` 가 nil 을 돌려준 **뒤에만** 선다 — 기록 실패는 보낸 것으로 세지 않고 다음 주기가 다시 시도한다. 기존 `d.alert` 는 반환값이 없고 오류를 로그로 삼키므로(reconcileloop.go:568-576) 새 함수는 `d.opts.Alerts.Notify` 의 반환값을 직접 보고, `Alerts == nil` 은 기록 성공으로 세지 않는다(r2 R2-1 — 기존 함수 편집 범위 불변). **알림 꺼짐**: critical 타입을 쓰지 않고 엔진 구조화 로그 한 줄만(그날 한 번, 같은 래치). 래치 맵은 드라이버 구조체의 새 필드이고 새 함수 안에서 nil 이면 만든다(`NewReconcileDriver` 본문 무편집). 재시작하면 그날 한 번 더 나갈 수 있다(공시). 대상은 운영자 전용 블록(D1)뿐이고 주기마다 활성 블록을 본다(**Manager 결정 2026-11-23**) | 보유가 없는 차단은 급하지 않지만(11-20 의 차단은 3 일 뒤에야 해를 냈다) 잊히면 안 된다. `Notify` 의 재알림 창은 1 시간이라(notifier.go:301·:395-399) 래치 없이 날짜 key 만으로는 같은 날 시간마다 다시 나간다(R1-2). 래치를 기록 성공 뒤에 세우는 규칙은 attestation 감시와 같다 |
| D6 | 본문은 **관측 시각 + 과거형**이다: 「<관측 시각> 대사에서 이 종목은 차단(<원인>, <Since>부터) 아래에 있어 편입하지 못했다. 그 시각 손절·익절이 걸려 있지 않았다.」 운영자 전용 차단(보유 보고·차단 보고)의 할 일 = 「이 차단은 스스로 풀리지 않는다. 풀려면 운영자 해제(`tossctl engine reconcile-resolve`)가 필요하다 — 이 명령은 엔진을 멈춰야 실행되고 멈춘 동안에는 모든 포지션의 자동 손절이 돌지 않으므로, 보유 종목의 시장이 닫힌 시간에 실행하라. 명령이 이 원인을 받지 않으면 거절하며 그 사유를 말한다.」 + 보유 보고는 「그 전까지 이 종목은 앱에서 직접 관리하라」, 차단 보고는 과거형 관측 문장 뒤에 이 할 일을 두고 「…로만 풀 수 있는 원인이었다」 같은 단정은 싣지 않는다. 해제 명령이 받는 원인 목록은 명령 쪽(a195 가 귀속 실패를 더함)이 정하므로 알림에 복제하지 않는다 — 두 판정이 되지 않게(gstack review red team CRITICAL 처분, **Manager 결정 2026-11-23**). `blocked_stale` 은 운영자 전용이라고 단정하지 않고 해제가 **필요할 수 있다**고만 말한다 — 「차단이 10 분 넘게 풀리지 않았다. 스스로 풀리지 않으면 운영자 해제(`tossctl engine reconcile-resolve`)가 필요할 수 있다 — 그 명령은 엔진을 멈춰야 실행되고 멈춘 동안에는 모든 포지션의 자동 손절이 돌지 않는다. 엔진 로그의 대사 상태를 확인하고 그 전까지 이 종목은 앱에서 직접 관리하라」(재시작 뒤 복원된 수량 불일치 블록은 조정 크레딧이 없어 일치만으로 안 풀릴 수 있음 — mismatch.go:479-483·:674-686, 구현 리뷰 P2-5). 차단 보고(D5)도 관측 시각의 과거형·조건형이고 「늦게 도착했다면 콘솔에서 현재 상태를 확인하라」를 싣는다(P2-2). 세 보고의 본문 시각은 `2026-11-23 08:09 KST` 형식(기존 `kst`), 기계가 읽는 `observed_at` 필드는 RFC3339 UTC(P2-7)(**Manager 결정 2026-11-23**, 구현 리뷰 처분 포함) | outbox 행은 전달 성공·운영자 승인으로만 정산되고 해소가 그 행을 갱신하지 않으므로 현재형은 늦게 배달되면 거짓이다(R1-3, 정본 engine-safety 「critical 무관리 보고의 본문은 … 시각의 사건」, 기존 `alertAdoptionFailed` adoption.go:481-505) |
| D7 | 제목·본문·필드에 계좌 정보를 싣지 않는다. 종목 표시는 기존 `label` 을 쓴다 | engine-safety 「시스템 밖으로 나가는 알림에는 계좌 정보가 실리지 않는다」 |

## 2. 편집 자리

- A(채택): `judgeHoldings` 의 차단 분기(if@152)에서 `continue` 하기 전에 그 보유를 모으고, 순회 뒤 새 파일(`internal/app/engine/reconcile_block_report.go`)의 새 함수들이 판정·기록한다. 차단 보고(D5)도 같은 자리에서 부른다 — `judgeHoldings` 는 `Observe` 가 성공한 주기의 마지막 단계라 블록 목록이 그 주기의 것이다. 드라이버 구조체에 메모리 상태 필드(유예 시계, 차단 보고 래치)를 더한다 — 구조체 선언은 함수 본문이 아니고, 맵은 새 함수 안에서 늦게 만든다.
- 편집하는 기존 함수 본문은 `judgeHoldings` 하나다(`NewReconcileDriver`·`Context.ReconcileDriver`·`RunOnce`·`alertUnmanaged`·`alertAdoptionFailed`·`unmanagedFact` 무편집 — 새 사실 값의 보고는 새 함수가 낸다).
- B(기각): exit 루프 `workingSet` 에서 보고 — `a090_workingset_exits_test.go:22-35`(`a090FrozenWorkingSetExits`)가 그 루프의 이탈 11 개를 고정한다.

`Observe` 가 실패한 주기에는 `judgeHoldings` 가 돌지 않는다(AST: analysis/cited-ast/RunOnce.ast.json B7 if@463 → return@467). 같은 호출 전에 반환하는 갈래가 셋 더 있다 — 안정화 실패·미수렴 B1 if@422 → return@423, 로컬 상태 읽기 실패 B2 if@428 → return@430, `Tracker.Refresh` 실패 B5 if@451 → return@455. 그 주기는 보고도 없다 — 대사 실패 자체의 기존 보고(`reconcile.mismatch` — `logCycle` 의 ERROR 로그 줄(reconcileloop.go:546)이고 알림이 아니다. 안정화 미수렴 주기는 오류가 아니라 이 줄도 없다)가 있다. 유예 시계는 그런 주기에 지우지 않는다(빠지지 않았다는 관측이 아니다).

## 3. 거절·침묵하는 정상 입력

- 수량 불일치로 막 차단된 보유: 10 분 전까지 침묵(정본의 전이 상태). 사람이 방금 산 종목의 정상 경로다.
- 편입 꺼짐 ∧ 미지정 종목·제외 목록 종목의 차단 보유: 이 보고 없음(D4).
- 이미 exit 자격을 가진(`ExitEligible` — 진입 결정 또는 편입 기록이 있는, internal/journal/position_projection.go:515-518) 보유: 이 보고 없음. 자격은 exit state 가 실제로 열렸음을 보증하지 않는다 — 편입 커밋 뒤 보호 미개설 보유(adoption.go `adoptOne`, 정본 engine-safety 「편입 기록이 커밋됐는데 보호가 아직 열리지 않은 보유」 경계)는 이 change 범위 밖이다.
- 대사 주기가 실패·미수렴인 동안: 보유 보고·차단 보고 모두 없음(§6).

## 4. 시험 전략 (RED 먼저) — Manager 결정 2026-11-23 으로 R1-7 전부 수용

새 시험 파일 `internal/app/engine/a194_reconcile_block_report_test.go`(가짜 시계). 「한 번」「창」 주장은 outbox 행 수가 아니라 **`Notify` 호출 수·전달 시도 수**로 단언한다(outbox 는 재발송 때 같은 행을 UPDATE 하므로 행 수 1 은 변이를 못 가른다 — journal/outbox.go:334-346). outbox 에 기대는 중복 방지 주장(보유 보고)은 실제 `obs.Notifier` + 실 원장 outbox 하네스(a095 하네스 형식)로 잰다. 각 단언은 **어느 타입**을 세는지 적는다(D5 가 같은 주기에 차단 보고를 따로 낸다).

1. 운영자 전용 차단(`ATTRIBUTION_FAILED` 종목 행) + 편입 켜짐 + 알림 켜짐 + 그 종목 보유 → 첫 주기에 `exit.position_adoption_failed` `Notify` 1(key·사실 `blocked_operator_only`), 편입 0, 시세 읽기 0. 실 Notifier 하네스에서 재알림 창 안의 다음 주기들은 전달 시도가 늘지 않는다.
2. 수량 불일치 차단 + 보유 → 9 분 59 초까지 0, 10 분에 1(사실 `blocked_stale`, 본문에 운영자 해제 안내 없음). 중간에 차단이 풀려 편입되면 0 이고 시계가 지워진다(다시 차단되면 새 10 분).
3. `Observe` 실패 주기가 끼어도 시계는 이어진다.
4. 편입 꺼짐 ∧ 미지정 + 운영자 전용 차단 + 보유 → 이 보고 0. include 지정 종목이면 1.
4a. SNAPSHOT_STALE(또는 SNAPSHOT_UNAVAILABLE) 종목 행 — `Release` operator_only · 사유 `ReasonReconcileMismatch` — 아래 보유 → 첫 주기에 `blocked_operator_only`(유예 없음). `Blocks().Release` 와 `EntryAllowed` 사유가 갈리는 유일한 입력이라 「`EntryAllowed` 사유로 판정」 변이를 죽인다(r2 R2-4).
5. 알림 꺼진 엔진(꺼짐·꺼짐 + topic 유지 각각) → 보유 보고는 `exit.position_unmanaged`(새 사실 값) 한 번(래치), 차단 보고는 로그 한 줄 그날 한 번. **새 outbox critical 행 0 · 진입 래치 0 · 운영 모드 전이 0**(기존 `TestA095NotificationsOffNeverRecordsACritical` 형식).
6. 제외 목록 종목 → 이 보고 0.
7. 운영자 전용 차단 활성, 보유 없음 → `reconcile.mismatch_permanent` `Notify` 그날 1, 같은 날 주기 반복에도 1, 다음 KST 날짜에 1 더. 차단이 풀리면 더는 없음. 계좌 전체 영구 차단도 1(종목마다가 아니라 한 통).
8. KST 자정 경계: KST 23:59 → 00:01(같은 UTC 날짜)에 1 더. 같은 KST 날짜 안의 UTC 자정 통과에는 더 없음.
9. 기록 실패 → 다음 주기 재시도 → 성공 뒤 래치(D5 차단 보고): `Notify` 가 오류를 돌려준 주기 뒤 다음 주기에 다시 `Notify`, 성공한 뒤로 그날 더 없음.
9a. 보유 보고(D3, 알림 켜짐)의 기록 실패 → 같은 사실 재관측 → 기록 성공: 알림 입구(`Alerts.Notify`)가 오류를 돌려주는 주기 → 다음 주기 재시도 → 성공 뒤 outbox 행 1 — 「critical 앞에 메모리 래치」 변이를 죽인다(r2 R2-4). 실 Notifier 의 원장 기록 실패와 그때의 진입 래치는 Notifier 의 기존 동작이라 이 change 의 시험 범위 밖이다(구현 리뷰 P2-3).
9b. 보유 보고(D3)의 전달 완료 뒤 재알림 창(1 시간)이 지나면 전달 시도가 한 번 더 일어난다(r2 R2-4).
10. 스스로 풀리는 차단만 활성 → 차단 보고 0.
11. `blocked()` 참 ∧ `Blocks()` 에 덮는 블록 없음(두 출처 불일치를 주입) → 10 분 유예 뒤 보유 보고 1, 원인 「미상」. 조용한 건너뛰기 0.
12. 본문은 관측 시각을 담은 과거형(D6). 늦게 배달되는 경우: PENDING 행 → 차단 해제·편입 성공 → 그 뒤 배달되는 본문이 관측 시각의 문장으로 참.
13. 제목·본문·필드에 합성 계좌·수량·가격 문자열 0.
13a. 주기 상한: 운영자 전용 차단 아래 편입 후보 보유 5 개 → 첫 주기 보유 보고 `Notify` 3 건·둘째 주기 2 건, 순서(처음 빠진 시각 오름차순, 같으면 포지션 id) 단언.
13b. 느린 전송: 보고 `Notify` 가 막히는(블로킹 가짜 알림기) 동안 같은 주기의 `adopt`(다른 편입 후보의 편입)가 그 보고보다 먼저 끝났음을 단언.
14. 기존 `TestAdoptionIsSilentUnderReconcile`·`TestActiveForeignCauseIsCountedAndBlocksBeforePriceRead`·`TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile`(`cycle.Unmanaged != 0` → Fatal)·`TestA095NotificationsOffNeverRecordsACritical` 무편집 통과.

변이(전부 CAUGHT, 변이마다 닿음 확인): 운영자 전용 판정을 항상 거짓으로 · 유예 비교 방향 뒤집기 · 시계를 지우지 않기 · D4 조건 삭제 · 차단 보고 key 에서 날짜 삭제 · critical 앞에 메모리 래치 넣기(보유 보고 — 기록 실패 뒤 재시도가 죽어야 함) · 차단 보고 래치를 `Notify` 앞에 세우기 · KST 날짜를 UTC 날짜로 · `Blocks()` 대신 `EntryAllowed` 사유로 판정 · 주기 상한 삭제 · 새 보고를 `adopt` 앞으로 옮기기.

선택 시험: `./internal/app/engine -run 'A194|Adoption|ReconcileLoop|Unmanaged|TestA095NotificationsOff|TestUSIncludedSymbolWaits'`, `-race` 는 `-run A194` 만.

## 5. 정본 delta 가 기존 문장과 어떻게 맞는가

정본 exit-policy 는 「전이 상태만 무알림」이라고 적는다. 이 change 는 그 문장을 뒤집지 않고 **전이 상태의 경계**를 정한다 — 운영자만 풀 수 있는 차단은 전이가 아니고, 스스로 풀리는 차단은 10 분까지만 전이다. ADDED 만으로는 기존 Scenario 「RECONCILE 중 편입 시도 … 무알림」(모든 RECONCILE 무결정 보유)을 좁힐 수 없으므로(R1-4) **MODIFIED 둘**을 둔다(Manager 결정 2026-11-23).

- exit-policy 「외부 취득 포지션의 자동 편입」: 알림 문단 「전이 상태(…)만 무알림이다.」 뒤에 한 문장을 더하고, Scenario 「RECONCILE 중 편입 시도」의 WHEN 을 「스스로 풀리는 RECONCILE 상태이고 유예(10분) 안」으로 좁힌다.
- engine-safety 「무관리 보유 보고의 등급은 사실이 정한다」: critical 이 되는 조건 집합 문장에 「전이 상태가 아닌 대사 차단으로 편입되지 못한 편입 후보」를 더하고 Scenario 1 개를 더한다. 그 Requirement 의 다른 문장·Scenario 는 바꾸지 않는다.

두 MODIFIED 블록은 정본 원문을 스크립트로 떠서 정확한 문자열 치환으로 만들었다 — 치환 목록·원문 1 회 출현 단언·정본 대조 diff 는 `analysis/spec-delta/`.

## 6. 잔여·후속

- 재시작이 유예 시계를 0 으로 되돌린다(D2). 반복 재시작 중에는 스스로 풀리는 차단의 보유 보고가 계속 미뤄질 수 있다 — 엔진 생존 감시(a134)의 영역.
- 재시작하면 그날의 차단 보고가 한 번 더 나갈 수 있다(D5 래치가 메모리). 같은 날 풀렸다 다시 걸린 차단은 다른 차단이라 다시 알린다(풀린 차단의 래치 key 는 그 주기에 지움 — 구현 리뷰 P2-4, Manager 결정).
- 한 주기의 포지션 인스턴스 조회 실패(judgeHoldings B5)는 그 보유를 「빠지지 않음」으로 세어 유예 시계를 지운다 — 다음 주기에 다시 빠지면 새 10 분. 원장 읽기 실패가 반복되는 상태는 그 자체가 다른 경보의 대상이고, 운영자 전용 차단은 유예가 없어 무관하다(구현 리뷰 P2-1, 잔여 — 수리하면 judgeHoldings 의 다른 분기 편집이 늘어남).
- 형제 `alertAdoptionFailed` 의 본문 시각은 UTC 표기 그대로다(범위 밖 — 이 change 의 세 보고만 KST).
- **배포 선행**: 귀속 실패 차단을 운영자 해제 명령이 받게 하는 것은 a195 다. a195 가 없는 빌드에서는 그 명령이 귀속 실패를 거절한다(cmd/tossctl/engine_reconcile.go:308-310) — a194 의 운영 반영은 a195 를 포함한 빌드여야 한다(같은 main 에 a195 가 먼저 착지한다).
- pending(비영속) 계좌 영구 승격 블록도 `ReleaseOperatorOnly` 라 운영자 전용으로 즉시 보고된다 — 그 블록은 이미 진입을 막는 실제 차단이고, 다음 깨끗한 비교에서 철회될 수 있다는 것을 가르려면 reconcile 패키지를 고쳐야 한다(gstack review adversarial #2, 기록만 — Manager 결정).
- 차단 보고는 주기당 블록 수만큼 `Notify` 한다(통상 1) — 보유 보고와 달리 상한 없음(gstack review performance, 기록만 — Manager 결정).
- `judgeHoldings` 앞에서 조기 반환하는 주기(안정화 미성공·로컬 상태 조회 실패·`Tracker.Refresh` 실패·`Observe` 실패 — RunOnce B1·B2·B5·B7)에는 보유 보고도 차단 보고도 나가지 않는다 — 그 주기에는 차단 목록·보유 판정 자체가 없다. ingest·converge 오류는 `cycle.Err` 에 남고 주기를 계속하므로(reconcileloop.go:438-449) 그 주기에는 보고가 나간다 — 구현은 `cycle.Err != nil` 을 침묵 조건으로 쓰지 않는다(r2 R2-5). 그 구간은 기존 대사 실패 보고와 루프 저하 경보의 영역이다(범위 밖 공시, Manager 결정 2026-11-23 — R1-5). 포지션 인스턴스 조회 실패(judgeHoldings B5)로 빠진 보유도 같다.
- 차단이 걸리는 순간의 원인별 상세(어느 주문·어느 전이)는 본문에 싣지 않는다. 장부의 `reconcile_states.evidence` 가 정본이다.
- 정본 reconciliation 「수량 불일치 감지 … 알림이 발송된다」(스스로 풀리는 차단의 진입 알림)는 이 change 가 메우지 않는다 — 이름 붙인 공백.
