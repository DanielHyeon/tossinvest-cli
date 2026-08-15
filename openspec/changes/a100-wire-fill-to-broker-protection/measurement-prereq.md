# a100 선행 실측 — 착수 전에 전제를 확인한다

이 change의 전제는 두 가지다. M-A는 아직 관측되지 않았고, M-B는 2026-08-11 정상 재기동에서
**6.7초 이하**로 관측됐다(실패 기동의 상한을 뜻하지 않는다).

1. 브로커 조건주문이 실제로 **발동해 체결된다.**
2. 엔진이 죽어 있는 동안의 **무보호 창**이 존재하고 정상 재기동 표본의 크기를 안다.

1이 거짓이면 이 change의 산출물은 보호가 아니라 보호의 외관이다. 2는 정상 재기동 표본만
측정됐고 a056의 실패 기동 창은 별개다. **따라서 구현 착수(tasks 1절) 전 남은 hard gate는
M-A 완전 PASS와 그 결과에 따른 0.11 판정표 동결이다.**

> **승인.** M-A는 실계좌 주문을 포함한다. 안전 불변식 §0-1(사람 승인 없는 LIVE 주문 side effect
> 금지)과 §0-7(live 검증은 사람이 직접 승인)에 따라 **실행 직전에 사람이 그 주문에 대해 별도로
> 승인한다.** 이전의 「소액 라이브로 진행」은 이 주문에 대한 승인이 아니다.
> 에이전트는 이 절차를 준비하고 관측을 기록하되 주문을 자동 실행하지 않는다.

## 2026-08-15 read-only preflight — HOLD

사용자 승인 뒤 Terra preflight와 별도 A-M 적대 감사를 실행했으나 주문 단계는 열리지 않았다.

- 2026-08-15는 토요일이고 다음 확인된 KR 정규장은 8/18이다. 실행 당일 market/expire를 다시 읽는다.
- official OPEN conditional은 `PAUSED` 2건·`WATCHING` 3건이었다. 기존 주문은 M-A cleanup과 섞지 않고
  각각 사람의 retain/cancel 결정 없이는 건드리지 않는다.
- 보유 046890(60), 466100(10)은 모두 journal의 OPEN/adopted 관리 상태이고, 1주 보유 108490에는
  PAUSED conditional이 있다. 따라서 그 시점에는 fresh 비관리 최소위험 1주 후보가 없었다.
- `trading.conditional=false`, official parent read는 가능했지만 WTS child `order show` session은
  invalid였고 local/deployed soak authority도 일치하지 않았다. 당일 fresh preflight가 다시 필요하다.
- 결정적 blocker는 현 verify record가 successful HTTP response의 process-local monotonic receipt와
  fsync barrier를 남기지 않는다는 점이다. 외부 wrapper나 wall time으로 대체하지 않는다.

따라서 tasks 0.2a의 M0 receipt 도구가 GREEN·A-M0 accepted되기 전에는 place하지 않는다. M0 이후에도
당일 auth/rate/market/holdings/sellable/all-page residual/attestation receipt가 fresh하지 않으면 HOLD다.

### 2026-08-15 09:14 KST 재점검 — M0 GREEN 이후에도 HOLD

사용자의 「다음 진행」 뒤 Terra read-only preflight와 별도 A-M Terra 적대 검토를 다시 수행했다.
두 검토는 독립적으로 **HOLD**에 합의했고, 주문 preview/place/modify/cancel, verify resume,
config·engine·container·soak lifecycle mutation은 모두 0건이었다.

- KR 시장은 토요일 폐장이었다. 다음 실제 세션은 세션 직전 `market hours`로 다시 확인하며 이
  기록의 날짜·expire를 재사용하지 않는다.
- official OPEN conditional all-page 결과는 pagination 종료까지 총 5건이고 raw status는
  `PAUSED` 2건·`WATCHING` 3건이었다. 기존 객체는 M-A parent와 섞거나 자동 정리하지 않는다.
- verify KR record는 ordinary outstanding 0, M0 checkpoint 0이지만 `conditional-trigger=deferred`다.
  trigger/child causal evidence는 아직 없다.
- 후보 보유는 모두 기존 엔진의 adopted/managed 기록과 겹쳤고, 1주 후보 하나는 이미 `PAUSED`
  conditional을 가졌다. 한 후보의 fresh official sellable read는 rate limit 뒤 WTS 401로 완결되지
  않았다. 따라서 사용할 종목을 자동 선택하지 않는다.
- `trading.conditional=false`이고, 장중 exact mutation 직전에만 사람이 한 키를 ON하고 cleanup 뒤
  OFF로 복원해야 한다. 일반적인 진행 승인은 symbol·trigger·expiry·confirm token이 있는 주문별
  승인이 아니다.
- PATH의 설치 CLI는 commit `676f8ae`(2026-07-30)이고 M0의 `--trigger-receipt-dir`가 없다.
  current source HEAD는 `882a0b49`이고 M0 flag를 포함한다. 설치 artifact와 reviewed source의
  identity가 일치하지 않으므로 설치 CLI로 M-A를 실행하지 않는다.
- soak는 `ready=false`, last cycle 2026-07-30, credential streak 0이었다. attestation 파일은
  존재하지만 proof 0이며 binary identity를 증명하지 않는다. official/WTS authority와 rate window를
  실제 세션 직전에 다시 맞춘다.

현재 worktree에서 **설치하지 않은** M0 후보를 `/tmp/a100-m0-artifact.wBqvWj`에 만들었다.
source manifest 36건은 current worktree와 일치하고 candidate SHA-256은
`4f8e47b93fbfc978042f5dde72b4a5af9c4ce6d0c11c76147397f50c6be5043b`이며 exact M0 flags와 local
M0/AttemptTrace tests는 PASS했다. 별도 A-M 검토도 checksum·source set·help contract를 확인했다.
그러나 이 후보는 dirty/untracked source에서 만든 `commit: unknown` 개발 빌드이고 PATH에 설치되지
않았으므로 **live-session execution authority가 아니다**. 장중 실행 전에는 reviewed source를 식별하는
설치 artifact와 atomic install/post-install read-only receipt를 새로 받아야 한다.

다음 세션의 필수 fresh input은 (1) M0 설치 artifact SHA/version/help, (2) official auth/rate와 WTS
valid 상태, (3) KR regular session과 새 expire, (4) OPEN conditional all-page raw status 및 5개 기존
객체의 사람 retain/cancel 결정, (5) holdings·official sellable·managed/unmanaged 판정, (6) binary와
일치하는 soak/attestation authority, (7) exact symbol/qty=1/SINGLE/MARKET/SELL/trigger/expiry preview와
그 주문에 대한 별도 사람 승인이다. 하나라도 없으면 preview나 config flip을 열지 않는다.

## M-A — 조건주문은 발동해서 체결되는가

### 왜 필요한가

`openspec/changes/verify-execution-capability/measurements.md:197`:

> **발동 관측과 `triggeredOrderId` 노출 지연** — `conditional-trigger`가 설계상 deferred.
> 2c의 기본 가설(SINGLE+MARKET 손절)이 실제로 **발동해 체결되는지는 양 시장 모두 미측정**이며,
> 이것이 2.5의 가장 큰 구멍이다.

등록·조회·존속·정정·취소는 실측됐다. **발동만 비어 있다.** 그리고 발동이야말로 이 change가
사려는 것 전부다.

### 순서 — probe 배포가 M-A보다 먼저다

`GET /api/v1/conditional-orders/{id}`는 **계좌에 조건주문이 있어야 증명된다**
(`probeOrderByID`와 같은 계약: 읽을 id가 없으면 Skipped다). 지금 그것을 가능하게 하는 것은
333430에 남은 2026-07-31 고아 조건주문 하나뿐이고, **아래 단계 0과 `verify run --resume`이
그것을 취소한다.**

⇒ **tasks 0.10b(probe 배포)를 먼저 하고, soak가 최소 한 사이클을 성공적으로 돌린 것을
확인한 뒤에 이 측정을 연다.** 순서를 뒤집으면 by-id 증거는 다음 조건주문이 생길 때까지
없고, 그 "다음"은 M-A가 만드는 조건주문뿐이므로 M-A 세션 중에 soak 사이클이 겹쳐야만 한다 —
운에 맡기는 순서다.

### 절차

| 단계 | 내용 | 기록할 값 |
| --- | --- | --- |
| 0 | **먼저 계좌에 남아 있는 조건주문을 조회한다.** 2026-07-31 verify 실행이 333430에 등록한 `DJwYn8P_dD9lQMEeYc-5l5_yDTzu2fo55FEJkhU8WVg`를 취소할 단계가 실행되지 않았다(`conditional-persist` = `awaiting-restart`). 살아 있으면 이 측정의 대상과 섞이지 않게 먼저 정리하거나, 대상 종목을 333430이 아닌 것으로 고른다 | 조회 시각, 남아 있는 조건주문 목록(id·종목·상태) |
| 1 | 시장이 **열려 있는** 시간에 수행한다. KR 우선(측정 선례가 KR에 있다) | 시장, 시각(KST) |
| 2 | 이미 보유 중인 유동성 높은 종목 **1주**를 대상으로 한다. 신규 매수를 만들지 않는다 | 종목, 보유 수량, 현재가 |
| 3 | 그 1주에 대해 SINGLE + MARKET 조건 매도를 등록한다. **trigger를 현재가 근처에 두어 곧 발동하게 한다** | 등록 요청 원문, trigger, 응답 |
| 4 | 등록 직후 매도가능수량을 다시 읽는다 (M13 재확인) | 등록 전/후 매도가능수량 |
| 4b | **등록한 주문을 조회해 trigger·수량이 등록 요청과 비교 가능한 형태로 돌아오는지 확인한다** | 조회 응답의 trigger·수량 필드명과 값, 요청값과의 일치 여부 |
| 5 | 발동을 관측한다 | 발동 관측 시각, 관측 경로(조회 API 필드명) |
| 6 | M0가 transport body-read 경계에서 매 HTTP attempt의 **raw status 원문**과 `triggeredOrderId`를 official response로 관측한다 | pre-trigger/triggering raw 문자열, raw-result digest, numeric HTTP status, parent local receipt sequence/monotonic 시각, child id tag, 등록→노출 지연 |
| 7 | M0가 parent receipt와 exact child cleanup checkpoint를 fsync한 뒤 official child by-id GET으로 최초 관측 fill을 기록한다 | child raw-result digest·local receipt sequence/monotonic 시각, request start, exact identity/symbol/side/quantity/filled/execution fields |
| 8 | durable child first-observed-fill receipt로 critical window를 닫는다 | parent/child ID tag equality, critical-window gap flag=false; 이후 parent terminal GET은 PASS 조건이 아님 |

### 판정

- **통과**: 발동이 관측되고 child가 체결되며 그 사실을 조회 API로 알 수 있다.
  **engine-safe local causal order에서 parent의 child-id receipt가 child 최초 fill receipt보다
  엄격히 먼저이고**, parent의 pre-trigger armed·triggering raw 문자열과 child의 terminal fill raw
  문자열이 구별된다. parent terminal 문자열은 child fill receipt 전에 관측되면 기록하지만 PASS의
  필수 조건은 아니다.
  broker server timestamp만으로 이 순서를 대체하지 않는다. → 0.11 판정표를 동결한 뒤 a100 착수.
  관측된 필드명·상태 문자열이 journal
  스키마(tasks 2.3), causal ownership(tasks 3.10), 수렴 판정(tasks 3.3)의 입력이 된다.
- **증거 불완전**: durable parent child-ID receipt부터 durable child first-observed-fill receipt까지
  parent/child read error·401·429·decode/identity gap,
  receipt write/sync 실패, process restart가 하나라도 있다.
  → **INCONCLUSIVE/HOLD.** server timestamp·외부 shell timestamp·다음 process 기록으로 메우지 않는다.
- **부분 통과**: 발동·체결은 되지만 조회로 알 수 없다(노출 지연이 크거나 필드가 없다).
  → **현재 A100 구현을 시작하지 않는다.** child durable 귀속과 terminal 재등록 방지가 성립하지
  않으므로 D2/D8/3.10을 다시 설계하고 별도 적대 리뷰로 proposal freeze를 다시 받아야 한다.
- **4b 실패**: trigger·수량이 비교 가능한 형태로 돌아오지 않는다.
  → **tasks 3.3의 수렴 정의가 성립하지 않는다.** "ACTIVE 수량 = 보유 수량이고 trigger가 유도값과
  같을 때만 수렴"을 판정할 데이터가 없기 때문이다. **현재 A100 구현을 시작하지 않고** 수렴
  정의를 다시 설계·리뷰한다. 정의를 느슨하게 바꾸지 않는다.
- **실패**: 발동하지 않거나 child가 체결되지 않는다.
  → **a100을 멈춘다.** 설계가 아니라 전제가 무너진 것이다(design D10). 상주 주문은 보호가 아니다.

### 비용과 한계

비용은 1주의 스프레드와 수수료다. 한계는 표본 1이라는 것 — 이 측정은 "발동이 가능한가"에
답하고 "항상 발동하는가"에는 답하지 않는다. US는 이 측정 범위 밖이며, KR 결과를 US로 옮겨
쓰지 않는다(a100은 US 수렴을 fixture로만 검증하고 실측은 별도로 남긴다).

### M-A 종료/cleanup

- final verdict 전에 official all-page conditional residual, 보유와 sellable을 다시 읽어 시작 전 receipt와
  비교한다. M-A가 만든 parent/child 외 기존 5개 residual은 retain/cancel 결정을 섞지 않는다.
- untriggered parent cleanup cancel, trigger 추적 중 modify가 필요하면 각각 사람이 새로 승인한다.
- triggered-but-child-unobserved, identity gap, crash checkpoint가 있으면 자동 child cancel이나 추정 cleanup을
  금지한다. owner-only verify record의 exact parent/child checkpoint를 `verify status`/`verify abort --list`로
  사람에게 보여주고 broker app과 official read를 통한 reconciliation을 별도로 승인받는다.
- config `trading.conditional` 원복과 residual=0 주장은 사람 mutation과 fresh read receipt가 모두 있을 때만
  기록한다.
- M0 entry에서 기존 verify outstanding가 있으면 trigger 실행보다 앞선 cleanup prologue도 금지하고 HOLD한다.
  create-before-response crash는 pre-create client intent를 official all-page raw read와 대조해 unique exact
  parent만 checkpoint하며, zero/multiple/mismatch는 mutation 없이 human reconciliation으로 남긴다.

### 같은 세션에서 함께 조달하는 증거 (tasks 0.10)

M-A는 사람이 승인한 조건주문 라이브 세션이다. **attestation 재발급에 필요한 증거 두 건이
정확히 같은 세션에서 나온다.** 따로 승인을 두 번 받는 것은 낭비이고, 라이브 세션을 두 번 여는
것은 위험이 두 배다.

| 필요한 증거 | 현재 | 조달 방법 |
| --- | --- | --- |
| `DELETE /api/v1/conditional-orders/{id}` | **없음** — 2026-07-31 실행이 `awaiting-restart`에서 멈춰 취소 단계에 도달하지 못했다 | `tossctl verify run --resume`으로 남은 단계를 마친다. 그 단계들이 조건주문을 취소하므로 단계 0의 정리와 같은 작업이다 |
| `POST /api/v1/conditional-orders/{id}/modify` | **없음** | D9 정정으로 교체가 자동 경로가 됐으므로 필요하다. verify 시나리오에 정정 단계가 있으면 그것을 쓰고, 없으면 M-A 절차 3~4 사이에 trigger를 한 번 올려 정정한다 |

두 증거의 유효기간은 30일이다. **M-A 세션의 날짜가 곧 재발급 마감의 기준일**이 되므로,
soak 재실행(tasks 0.12)과 M-A를 30일 안에 함께 끝낸다.

`GET /api/v1/conditional-orders`는 여기서 조달할 수 없다 — read는 supervised 증거로 인정되지
않는다(`soak/attest.go:255-277`). 그것은 soak probe 추가 + 재실행의 몫이다.

### 2026-08-15 M0 설치 후보 봉인 — 설치·실측 권위는 아님

M0 구현 acceptance 뒤의 첫 임시 후보는 `commit: unknown`이고 dirty source를 독립 재구성할 수
없어 장중 실행 권위로 거부했다. 같은 A100 안에서 Terra 구현자가 코드 편집 없이 다음 release
artifact를 다시 만들고, 별도 Terra adversary가 read-only로 재검증했다.

- receipt: `/tmp/a100-m0-artifact.M8EeNi/RELEASE-ARTIFACT-RECEIPT.md`
- candidate SHA-256:
  `899a74aca0d72411de80bb92fe89cfbdf31948dbf008635a1c4bb7125ddae882`
- embedded identity:
  `a100-m0-candidate+git.882a0b490b0b.dirty.d5a775351fba`, exact HEAD
  `882a0b490b0b6d2eb7abe5c5040c514776f49f3e`
- authoritative source archive SHA-256:
  `9e9a33a1309059acc8e5dc1d6fc8e2598c413d81f036a7cc39ea00e89c9e94f2`
- NUL-safe source manifest SHA-256:
  `d5a775351fbacc315092e24ebeedfff747df43fa4edd37db1828af2e3ff0dcf0`
- 별도 extraction 두 곳의 build는 byte-identical했고 top-level `SHA256SUMS` 50/50이 독립 PASS했다.
  candidate self-report, exact M0 help, archived-source focused tests도 receipt와 일치했다.
- 설치 전 이미지: `/home/daniel/.local/bin/tossctl`, SHA-256
  `b0e805f35b94c289da25864c129c79bc5066f950b3575357e01b743fdbda96f9`,
  mode `0755`, owner `daniel:daniel`.
- adversarial verdict: **ACCEPT — P0=0, P1=0, P2=0**.

이 결과는 이전 후보의 provenance 결함만 닫는다. PATH 바이너리는 교체하지 않았고 config/runtime,
Broker read·mutation, preview, verify resume도 실행하지 않았다. 실제 설치는 위 preimage를 다시
확인하고 같은 디렉터리·같은 filesystem에서 staged file/backup 각각 file+directory fsync, atomic
rename, identity/help postcheck, exact backup rollback을 수행하는 **별도 사람 승인 작업**이다.
설치 승인도 실계좌 주문 승인이 아니며, 설치 뒤에도 장중 fresh preflight와 exact 주문별 승인 없이는
M-A를 시작하지 않는다. 따라서 tasks 0.2와 0.11은 계속 미완료이고 T1은 열지 않는다.

### 2026-08-15 M0 PATH 설치 — 설치 gate만 ACCEPT

사람이 위에서 분리한 원자 설치 단계를 승인했다. Terra 설치 담당이 candidate와 preimage를 다시
검증한 뒤 다음 한 경로만 바꿨고, 별도 Terra adversary가 설치 후 상태를 read-only로 재검토했다.

- live target: `/home/daniel/.local/bin/tossctl`
- live SHA-256:
  `899a74aca0d72411de80bb92fe89cfbdf31948dbf008635a1c4bb7125ddae882`
- live metadata: regular non-symlink, mode `0755`, owner `daniel:daniel`
- durable same-directory backup:
  `/home/daniel/.local/bin/tossctl.backup.b0e805f35b94c289da25864c129c79bc5066f950b3575357e01b743fdbda96f9`
- backup SHA-256:
  `b0e805f35b94c289da25864c129c79bc5066f950b3575357e01b743fdbda96f9`
- install receipt: `/tmp/a100-m0-install-receipt.Z9bvJr` (`0700`, `SHA256SUMS` 7/7 PASS)
- attribution receipt: `/tmp/a100-m0-install-attribution.4rxHlA` (`0700`, manifest PASS)
- rollback: 실행하지 않음.

설치 중 target user는 0이었고 candidate/backup/stage의 SHA·mode·owner를 확인한 뒤 같은 directory에서
file·directory fsync와 atomic rename을 수행했다. postcheck는 isolated HOME의 `version --output json`과
M0 help 다섯 flag에 한정했다. `tossctl update`, Broker API, order/preview, config, verify run/resume,
Docker/engine 명령은 실행하지 않았다.

최초 adversarial review는 receipt 직후 token/marker/log/WAL mtime이 바뀐 사실을 설치자 비간섭성의
증거 공백(P1)으로 잡았다. 후속 read-only attribution은 설치 전부터 살아 있던 동일 engine PID,
두 container `RestartCount=0`, 설치 전 config mtime/hash, 전후 62~63초 `reconcile.clean` 연속 기록,
audit/order/verify mutation 0건으로 기존 엔진의 자연 쓰기임을 강하게 귀속했다. 같은 reviewer의
최종 판정은 **P0=0, P1=0, P2=1**이다. P2는 과거 개별 write syscall owner를 point-in-time audit 없이
수학적으로 복원할 수 없다는 포렌식 한계이며 깨끗한 historical filesystem audit로 바꾸어 쓰지 않는다.

이 acceptance는 설치 identity와 rollback 가능성만 닫는다. fresh Broker preflight, M0 trigger run,
conditional gate 변경과 실계좌 주문은 승인되지 않았다. tasks 0.2/0.11과 T1은 계속 HOLD다.

### 2026-08-15 10:23~10:27 KST fresh M-A read-only preflight — HOLD/NO-GO

사용자가 fresh read-only preflight만 승인했다. Terra 수집 담당은 installed M0 binary와 official/local
read surface만 읽고 `/tmp/a100-ma-preflight-final.OYpbFG`를 봉인했다. 별도 Terra adversary가
pagination 진술의 불일치를 발견했다. 수집 담당은 Broker를 다시 호출하지 않고 독립 reviewer의
2026-08-15 10:27:58 KST forced-OpenAPI terminal read를 명시적으로 귀속한 새 receipt
`/tmp/a100-ma-preflight-corrected.kSn14D`를 봉인했다.

- corrected receipt: directory `0700`, payload/manifest `0600`, self-excluding `SHA256SUMS` 3/3 PASS;
  token·account/order opaque ID 노출 0.
- installed M0 identity: SHA-256 `899a74aca0d72411de80bb92fe89cfbdf31948dbf008635a1c4bb7125ddae882`,
  commit `882a0b490b0b6d2eb7abe5c5040c514776f49f3e`, M0 필수 flag 다섯 개 일치.
- official OPEN terminal read: `count=5`, `has_next=false`, cursor 없음; raw status는
  `PAUSED` 2건·`WATCHING` 3건. 따라서 **pagination은 blocker가 아니다.** 다만 5건 각각의 사람
  retain/cancel 결정이 없으므로 M-A 객체와 섞을 수 없다.
- KR regular market은 폐장 상태였고 다음 session은 2026-08-18 09:00~15:30 KST로 관측됐다.
- official holdings는 읽혔지만 sellable은 독립 대조 기준 3/8만 완결됐다. journal의 현재 8개
  position이 모두 OPEN/adopted여서 미관리 whole-share KR 후보가 증명되지 않았다.
- soak는 `ready=false`, credential streak 0, current window cycle 0이었다. attestation은 날짜상
  유효하지만 binary/commit/build/SHA binding이 없어 installed M0 authority가 아니다.
- `trading.conditional=false`, `conditional-trigger=deferred`; WTS 401은 official 증거 대체에 쓰지 않았다.
- 최초 수집의 10:23 KST official rate-limit은 역사 snapshot으로만 보존한다. 뒤의 terminal list read가
  pagination만 정정하며 auth/rate·sellable·candidate·soak gate를 해소하지 않는다.

같은 adversary의 재검토는 receipt 품질을 **P0=0, P1=0, P2=1**로 ACCEPT했다. P2는 host process
survey가 container/runtime ownership을 증명하지 못한다는 비차단 관찰 한계다. 운영 판정은 별개로
**HOLD/NO-GO**다. preview, place/cancel/amend/modify, `verify run`/resume/abort action, config,
engine/Docker, auth/update는 모두 0건이다. 따라서 tasks 0.2와 0.11은 체크하지 않고 T1을 열지 않는다.

## M-B — 엔진 사망부터 재무장까지의 무보호 창

### 왜 필요한가

`proposal.md`가 인용하는 a056(2026-08-02)에서 엔진이 8분간 기동하지 못했고 그동안 OPEN 5건과
exit 정책 4건이 무감시로 남았다. 그 8분은 **사고 보고서에서 읽은 값**이지 측정값이 아니다.
a100이 없애려는 창의 크기를 모르면 개선 폭도 주장할 수 없다.

### 절차

| 단계 | 내용 | 기록할 값 |
| --- | --- | --- |
| 1 | **두 시장이 모두 닫힌 창(KST 05:00~09:00)에서 수행한다.** 프로젝트 기억 `tossos-engine-stop-removes-stoploss` — engine lock을 잡는 작업은 이 창에서만 | 시각(KST) |
| 2 | 정상 동작 중인 엔진의 exit observer가 무장된 상태를 확인한다 | 확인 방법, 시각 |
| 3 | 컨테이너를 정지시킨다 | 정지 시각 |
| 4 | 재기동한다 | 재기동 시각 |
| 5 | exit observer가 다시 무장되어 첫 관측을 수행할 때까지 기다린다 | 첫 관측 시각 |
| 6 | 3~5의 구간을 무보호 창으로 기록한다 | 창 크기(초) |

### 판정

판정 기준이 아니라 **기록**이다. 이 값은 a100 완료 후 "브로커 상주 보호가 덮는 시간"의
분모가 된다. 실계좌 주문을 만들지 않으므로 §0-1의 승인 대상은 아니지만, 엔진을 세우므로
시장이 닫혀 있어야 한다.

### 한계

시장이 닫힌 상태의 재무장 시간이므로, 시장이 열린 상태에서 시세 구독·대사가 함께 붙는
경우보다 **짧게 나올 수 있다.** 이 값은 하한으로 읽는다.

## 기록 위치

두 측정의 원자료와 판정은 이 파일 아래에 append한다. `verify-execution-capability`의
`measurements.md`에도 M-A 결과로 `conditional-trigger` deferred 항목을 해소하는 참조를 남긴다.

---

## 측정 기록

<!-- 실행 후 여기에 append한다. 미실행 상태에서 tasks 1절로 넘어가지 않는다. -->

- [x] **선행: tasks 0.10b — 조건주문 probe 배포.** 2026-08-11 14:00Z 완료.
      이미지 `sha256:86c6e4d2…`로 교체, 두 컨테이너 healthy. 재시작 구간을 M-B로 측정했다.
      **KST 05:00~09:00 창이 아니라 미국 정규장 중에 수행했다** — 사용자가 위험을 통보받고
      즉시 진행을 선택했다(창 6.7초, US 3종목 노출)
- [x] **0.10c — soak 1사이클 성공 확인.** 2026-08-11T23:44:54Z 시작 cycle에서 조건주문
      read 3개를 포함해 endpoints 9/9 성공했다. by-id 증거의 유일한 원천(333430 고아)을 M-A가
      없애기 전에 완료했다. soak restart는 0.10b-1/0.12로 완료됐다.
- [ ] M-A 단계 0 — 남아 있는 조건주문 조회 (읽기)
- [ ] M-A 실행 (사람 승인 필요)
- [ ] `tossctl verify run --resume` — DELETE 증거 + 333430 정리 (사람 승인 필요)
- [ ] 조건주문 정정 증거 (`.../modify`) 확보 (사람 승인 필요)
- [x] M-B 실행 — 2026-08-11, **계획된 배포 재시작**으로 측정(아래 기록)

### M-B 기록 — 2026-08-11 (계획된 배포 재시작)

tasks 0.10b(조건주문 probe 배포)의 컨테이너 재생성이 곧 M-B의 조건이므로 같은 작업에서
측정했다. **원안과 다른 점: 시장이 닫힌 창이 아니라 미국 정규장이 열린 상태에서 측정했다**
(사용자가 즉시 진행을 선택). 원안이 "닫힌 상태의 값은 하한으로 읽는다"고 적은 바로 그 한계가
여기서는 없다 — 시세 구독·대사가 함께 붙은 상태의 값이다.

| 시각(UTC) | 사건 | 기준 대비 |
| --- | --- | --- |
| 13:59:51.215 | 정지 전 마지막 엔진 생존 신호(`reconcile.clean`, 주기 ~62초) | — |
| **14:00:28.789** | `docker compose up -d` 발행 — 엔진은 이 시각 이후에만 죽을 수 있다 | 기준 0 |
| 14:00:29.511 | compose 반환(두 컨테이너 recreate + start 완료) | +0.7s |
| 14:00:32.973 | `engine.operating_mode` — 새 엔진 기동 | +4.2s |
| **14:00:35.439** | 첫 `exit.*` 관측 — exit observer 재무장 | **+6.7s** |

**무보호 창 ≤ 6.7초.** 상한이다: 엔진의 실제 사망 시각은 명령 발행 이후이므로 창은 그보다
짧다. 컨테이너 종료 유예 75초는 쓰이지 않았다(엔진이 즉시 종료했다).

당시 노출: 열린 포지션 5건(KR 272210·333430, US IONQ·MWG·TSLA) 중 **US 3건이 정규장 중**이었고
그중 IONQ(baseline 32.9935)·MWG(baseline 1.1115)가 활성 손절선을 갖고 있었다.

#### 이 값이 답하는 것과 답하지 않는 것

- **답한다**: 계획된 배포 재시작의 무보호 창은 **초 단위**다. a100이 브로커 상주 보호로 덮으려는
  시간의 분모로 쓸 수 있다.
- **답하지 않는다**: `proposal.md`가 인용한 a056(2026-08-02)의 8분은 **엔진이 기동에 실패한**
  경우다. 이번 측정은 정상 이미지로 정상 기동한 경우이며 그 실패 모드를 경계하지 못한다.
  ⇒ **a100이 사는 안전의 크기는 "6.7초"가 아니라 "기동이 실패하면 사람이 알아챌 때까지"이다.**
  6.7초는 하한이고 8분은 관측된 상한이다.
- 표본 1이다.
