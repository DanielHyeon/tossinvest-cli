# a198 편집 표면 (tasks 1.4)

기준 HEAD `e1eaf720c`. 세기 방법: grep 전수(아래 명령) + a166 census JSON 키 전수 + a087 번들 목록. 시험 실행 없음.

## 1. 구조 고정·census 적중

| 고정 | 대상 | `clearTheSymbol` | `Context.ExitObserver` | `ExitObserverOptions` 필드 추가 | `RiskGuardian` 새 메서드 |
|---|---|---|---|---|---|
| a163 렌더 해시(a163_judgement_pin_test.go:41-52) | ExitObserver 판정 메서드 아홉 + stableObservationID — `record` 포함, `clearTheSymbol` **미포함** | 0 | 0 | 0 | 0 |
| a166 census(internal/accountbudgetruntime/testdata/a166-wiring/shapes.json 90 키 · sites.json 116 행) | exitloop.go·exitwiring.go·riskguardian.go·issue.go 키 0. engine.go `NewContext`·`NewOrderPath`, gateway.go `buildGateway` 는 있음 | 0 | 0 | 0 | 0 (NewContext 를 고치면 1) |
| a092 C1 options 핀(cmd/tossctl/a092_exit_options_pin_test.go) | engineRuntime 의 ExitObserverOptions 리터럴이 Alerts·Announcer·Retrier·Floor 를 넣지 않음, 리터럴 1개 | 0 | 0 | 0 (engineRuntime 이 새 필드를 넣지 않으면) | 0 |
| a092 K18 핀(a092_structure_pins_test.go:393 `TestA092TheExitIssuerDoesNotReachAnAnnouncer`) | execgw 패키지에서 **이름** `IssueReduction` 의 호출 폐포가 announcer·escalate 에 닿지 않음(이름 기반이라 RiskGuardian·Issuer 두 본문이 합쳐짐) | — | — | — | **새 이름은 핀 밖** — 같은 성질을 지키려면 시드에 추가(시험 함수 편집 1 → FLM 번들 1) |
| a087 시험 핀(a087_working_order_price_pin_test.go) | `workingOrderPrice` 값만 | 0 | — | — | — |
| seal_test.go(Context 공개 필드가 mutator 가 아님) | Context 필드 | — | Context 에 새 필드를 두면 대상(Issuer 는 mutator 아님 — 적중 예상 0, 미실행) | — | — |
| cmd/tossctl/engine_runtime_branch_test.go:21 | `authorizationOnlyGuardian` 가 ReductionIssuer 미구현 → ErrExitObserverUnavailable | — | 새 단언을 B3 뒤에 두면 기대 그대로 | — | `&execgw.RiskGuardian{}` 는 새 인터페이스도 구현 |

명령: `grep -rln 'clearTheSymbol' --include=*_test.go internal cmd tools`(5 파일: a084b·a087 pin·a094_exit_clear·a163 pin·journal a087 pin — 전부 주석·이름 언급, 본문 렌더 고정은 a163 의 `record` 뿐) · `grep -rln 'ExitObserverOptions' --include=*_test.go internal cmd`(50 파일) · census 키 python 전수.

## 2. ExitObserverOptions 를 만드는 시험 자리(36 리터럴)

| 경로 | 수 | 함수 |
|---|---|---|
| `Context.ExitObserver` 경유(운영 조립 — 공급 자동) | 12 | engine_assembly_test · a090 R17 · a091 · a092 ×2 · a094_wiring · a095 · a129 assembly · a163 ×2(443·647) · guardian_production · tracer_test(게이트에서 먼저 거절) |
| `NewExitObserver` 직접 | 10 | a087_real_gateway_e2e · a110 · a111 `a111ObserverForJournal` · exit_e2e `buildE2EStack` · exit_identity `newConcurrentObserver` · exitloop_test `newExitHarness`(:227 — a094 등 대부분의 하네스) · :1170 · :1229(Journal nil 거절) · :1239(`stubIssuer{}` 로 비용 모델 거절) · a163:627 |
| `&ExitObserver{opts: …}` 직접(NewExitObserver 우회) | 14 | a150 ×12 · a159 ×2 |

## 3. design D2 두 안 비교

| 항목 | (나) Guardian 새 메서드 `IssueCancel` + 새 인터페이스·옵션 필드 | (가-A) `execgw.Issuer` 를 새 옵션 필드로 주입 | (가-B) 관측기가 기존 필수 필드(Journal·AccountRef·clock)로 `execgw.Issuer` 를 스스로 만듦 |
|---|---|---|---|
| 고칠 기존 함수(gate ⑤ FLM) | clearTheSymbol · Context.ExitObserver · (필수 검증 시) NewExitObserver · a092 K18 시험 함수 = **3~4** | clearTheSymbol · Context.ExitObserver · (필수 검증 시) NewExitObserver = **2~3** | clearTheSymbol(호출 자리에서 구성) = **1**, 또는 NewExitObserver 에서 한 번 구성 시 **2** |
| 새 코드 | riskguardian.go 새 메서드 · exitloop.go 인터페이스·필드 | exitloop.go 필드 | 없음(또는 비공개 필드 1) |
| census·구조 고정 적중 | a166 0 · a163 0 · **a092 K18 시드 확장 필요 1** | 0 · 0 · K18 자동 포함(같은 이름) | 0 · 0 · K18 자동 포함 |
| 결정 기록이 flatten 과 같은가 | 같은 열 구성을 **손으로** 다시 씀(RecordDecision 직접 — riskguardian.go:609 와 같은 방식). TTL = g.ttl, ID = g.newID | **같은 함수**(issue.go:86 → record :137) — ID 접두만 다름(flatten 은 `flatten-`), TTL 기본 60s | (가-A) 와 같음 |
| 새 인터페이스를 채워야 하는 기존 시험 하네스 | 필드 필수 검증 시: §2 직접 10 중 9(:1229 는 Journal nil 이 먼저) + `stubIssuer` 류 1 = 대략 **9~10 함수**(각 FLM 번들). 선택 필드면 최소 `newExitHarness` 1 + 직접 리터럴 중 길 치우기에 닿는 것(미측정) | (나) 와 같음 | **0** |
| D3(배선 누락이 기동 오류) | Context.ExitObserver 단언으로 성립 | 성립 | **대상 소멸** — 이미 필수인 Journal·AccountRef 에서 유도되므로 누락할 배선이 없다. spec Scenario 「취소 결정 발행기가 조립에 없다」가 공집합 — **Manager 결정 필요** |
| 시계 | RiskGuardian 시계(게이트웨이와 같은 Context 에서) | Context 가 넣는 시계 | 관측기 시계 `o.clk`(운영 = engineRuntime 의 clk; 게이트웨이 시계와 같은지 미확인) — Issuer 주석 issue.go:36-37 「게이트웨이가 검증하는 시계와 같아야 한다」 |

권고: **(가-B)** — 편집 함수 최소(1~2), 하네스 편집 0, 결정 기록이 flatten 과 같은 함수, K18 자동 포함. 단 D3 와 spec 의 조립 Scenario 를 바꾸는 결정이고 시계 동일성 확인이 선행이라 문서는 고치지 않았다(Manager 결정 대기). (나) 를 유지하면 a092 K18 시드 확장과 하네스 9~10 개 번들이 3.3 범위에 들어온다.

## 4. 진행 중 change 와의 겹침

| change | 상태 | 겹침 |
|---|---|---|
| a087-a-protective-exit-is-a-market-order | 열림, `landed-commit.txt` **없음**, base `102d4e996`, 남은 task 9(4.4·5.4·5.5·6.1-6.6) | a087 의 **현재(revision current) 번들에 `Context.ExitObserver` 가 있다**(`analysis/function-logic/internal-app-engine--context.exitobserver/ast.json` source_sha256 `e8f131a3…` = HEAD 의 그 함수). a198 이 그 함수를 고치면 a087 의 그 번들이 낡아 a087 gate ⑤ 가 막히고, a087 창(base→워킹트리)에 a198 의 `clearTheSymbol` 편집도 새 필수 함수로 들어간다. → a087 이 착지를 먼저 기록하거나(`--record-landing`) a198 이 a087 종결 뒤 착지해야 함. (가-B) 는 Context.ExitObserver 를 건드리지 않아 앞의 충돌은 없앰(clearTheSymbol 은 남음) |
| a167-a-protective-remainder-is-re-priced | 등록 단계(0.x 전부 미착수, 코드 0) | 같은 파일 exitloop.go 의 `record`·`submit` 예정. a198 은 `record` 무편집 계획 — 함수 단위 겹침 0. 단 a163 렌더 해시가 `record` 를 고정하므로 a167 이 그것을 건드림(a198 과 무관) |
