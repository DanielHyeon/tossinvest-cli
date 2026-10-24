# 기존 거래 기록 증거 감사 — 2026-10-21

## 최신 판정 — 브로커 조회 후 Manager 검토

2026-10-21 03:58 KST의 v3 보완 조회로 KR·US 각 1건의 공식 SELL/FILLED 상세와 실제 commission/tax 금액을 확인했다. 아래 초기 로컬 조사에서 비용을 미확인으로 둔 결론은 이 두 표본에 한해 보완됐다. 전 계좌 비용표나 정책 비용률을 확정한 것은 아니다.

Manager는 조회 스크립트를 실행 전에 검토하고, 비식별 결과·원래 영수증과의 대조 조건·보존된 값들을 확인했다. Decimal 계산으로 두 표본의 수수료·세금 bps를 독립 재계산해 모두 일치함을 확인했다. 초기 집계기의 누락값 오판은 수정했고 그 집계는 폐기했다. US의 기존 제출 수량 대조는 수행하지 않았다는 한계도 반영했다.

두 주문의 주문 수량과 체결 수량은 각각 1주로 같지만, 당시 계좌의 해당 종목 보유 전량을 매도했다는 증거는 아니다. 취소·정정 거절 레코드는 읽은 페이지에서 관측되지 않았으며, CLOSED 이력을 끝까지 읽지 않아 전체 이력에 없다고 판단하지 않는다. 2.1·2.2·2.9는 미완료 체크박스를 유지한다.

이번 보완은 기존 거래 GET 총 11회와 임시 파서의 작은 로컬 입력 검증만 수행했다. 전체 테스트, 새 주문, 서비스 재시작, 운영 설정 변경, 토큰 갱신은 수행하지 않았다. 원문은 메모리에서 읽고 개인정보 없는 표본 값만 보존했다.

## 범위와 보존

이 감사는 이미 보존된 로컬 기록만 읽었다. 외부 API, 주문, 서비스 재시작, 설정 변경, 원장 쓰기, 신규 측정은 수행하지 않았다. 계좌·종목·주문 식별자와 토큰은 출력하거나 복사하지 않았다.

- 캡처 시각: `2026-10-21T03:38:23+09:00`
- 운영 컨테이너의 데이터 마운트: `~/.config/tossctl` → `/var/lib/tossos/config`, `~/.local/share/tossos` → `/var/lib/tossos/data`.
- 현재 운영 원장: `~/.config/tossctl/journal.db`, 관찰 시 `167,333,888` bytes, mtime `2026-10-21 03:27:04 +0900`.
- 원장은 가동 중일 수 있으므로 파일 해시를 안정 스냅샷으로 주장하지 않았다. `file:...journal.db?mode=ro`와 `PRAGMA query_only=ON`으로 단일 연결을 읽었고, 조회 전후 `PRAGMA data_version=2`였다.

| 보존 입력 | SHA-256 | 이 감사에서 쓴 범위 |
| --- | --- | --- |
| `~/.config/tossctl/capability-verify.jsonl` | `7ab9420f2c52578b66e7f1f8ff62241fa9c3a74c937fa83a113ba668a860b218` | 현재 마운트 검증 기록의 비용·상태 관측 |
| `~/.local/share/tossos/capability-verify.jsonl` | `5e216f801abd44b87aa11c8886f47459e7acc5bd9a69fa65a0db6f8d7987fb6c` | 이전 KR 매도 경계 관측 |
| `~/.local/share/tossos/capability-verify-us.jsonl` | `212b08b41c39409c6211f814b1321ed57f6d3dd108f25b44e271c371108f30fc` | 이전 US 매도 경계 관측 |
| [a087 MARKET 매도 영수증](../../a087-a-protective-exit-is-a-market-order/analysis/market-sell-receipts/a087ms-20261019T145815-8bdcb862.json) | `bb86bede4143ab2d67bafd3142f1164e33da846920fd377fec4d6ca24f803666` | 접수 응답과 완료 이력 대조 범위 |

## 확인한 기록

### 매도가능 전량 접수와 포지션 전체 청산은 다르다

`internal/verifylive/steps.go:517-524`의 `sell.boundary.full_accepted` 계약은 시작 시점의 `sellableQuantity`만큼 주문이 **접수된 뒤 취소되는지**다. 이는 최종 보유량 0 또는 전량 체결의 증명이 아니다.

- 이전 KR·US 매도 경계 기록은 `partial_accepted=true`, `resting_sell_reserves=true`, `over_holding_rejected=true`, `full_accepted=unverified`다.
- a087 M50 영수증은 KR 정규장에서 일반 MARKET 매도 1주가 HTTP 200으로 접수되고, 완료 이력 화면과 시각·방향·수량으로 대조됐다는 증거다. 영수증 자체는 주문 당시 보유/매도가능수량, 잔량, 수수료·세금 또는 공식 `FILLED` 원문을 담지 않는다. 따라서 한 단위 매도 수용·완료 이력 증거로만 사용한다.
- 운영 원장 `fill_snapshots` 익명 집계에는 KR `SELL/FILLED` 39행, US `SELL/FILLED` 22행이 있고, `mutation_attempts`·`intents`와 조인된 종료 SELL 행은 67행이다. 67행의 조건은 `terminal=1`이며 `FILLED` 조건이 없으므로 체결 수로 해석하지 않는다. 스냅샷은 주문 당시 매도가능수량과 최종 계좌 보유량을 직접 저장하지 않고, 이 집계는 특정 보유량·종목·시점과 연결하지 않는다.
- `positions`의 CLOSED 82건 중 `trade_outcomes`와 직접 결속된 것은 44건이다. `trade_outcomes`는 엔진 포지션 결과이며, 원주문 당시 매도가능수량이나 계좌 전체 포지션 청산의 증명은 아니다.

따라서 과거의 특정 매도 기록이 있더라도, 이 조사 범위 안에서는 **매도가능 전량 접수**도 **포지션 전체 청산**도 확정하지 않는다.

### 비용 실측

- 현재 마운트 검증 기록의 `costs` 단계는 `orders_filled=0`, `collected=false`다. 이전 US 검증 기록도 같은 값이며, 이전 KR 기록은 `costs.collected=false`다.
- `internal/verifylive/steps.go:806-844`는 채워진 검증 주문에만 `execution.filledQuantity`, `commission`, `tax`, `filledAmount`를 합산한다. 다만 `parseDecimal`은 null·누락을 0으로 해석할 수 있으므로, `collected=true`만으로도 원문 필드의 비null성·통화·체결금액·주문별 중복/부분체결 처리가 보존되지 않으면 실측 비용 표가 되지 않는다.
- 확인한 `fill_snapshots`에는 체결 금액(`filled_amount`)은 있지만 broker commission/tax/currency 컬럼은 없다. `intents.currency`는 존재하지만 주문 의도의 통화이며 실행 수수료·세금 원문을 대신하지 않는다. `trade_outcomes.cost_total`은 44행 중 42행에 있으나, `internal/journal/trade_outcomes.go:169-260`이 `costs.Model.EstimateTradeCost`로 계산해 동결한 모델 비용이다. 브로커 실행 수수료·세금의 원문 증거가 아니다.
- a087 영수증 3개와 `verify-execution-capability`의 JSON 증거 파일을 키 수준으로 점검했지만 execution commission/tax/filledAmount 원문을 담은 보존 파일은 찾지 못했다.

따라서 실제 체결이 있었다는 원장 집계와 별개로, **실제 execution commission/tax/currency/base 비용 근거는 이 범위에서 미확인**이다.

### 거절 상태와 원주문 결속

- 보존 검증 JSONL의 `order.status.cancel_rejected.listed=false` 및 `order.status.replace_rejected.listed=false`는 해당 별도 레코드가 목록에서 관측되지 않았다는 뜻이다.
- 운영 원장의 `fill_snapshots` 및 `mutation_attempts.reason_code`에도 `CANCEL_REJECTED`/`REPLACE_REJECTED` 해당 행은 집계상 0이다.
- `REJECTED` 상태 또는 HTTP 거절 관측은 일반 주문 거절과 구분해야 한다. `CANCEL_REJECTED`·`REPLACE_REJECTED`는 브로커의 별도 주문 레코드와 원주문 링크가 필요한 계약이며, 그 형태는 [측정 문서 M31](../measurements.md#L153)에서 여전히 미측정으로 남아 있다.

따라서 HTTP 거절·일반 `REJECTED`·취소/정정 거절 별도 레코드를 혼용하지 않는다.

## 후속 증거가 충족해야 할 최소 결속

| 주장 | 같은 시점·범위에서 함께 필요한 필드 | 현재 결론 |
| --- | --- | --- |
| 매도가능 전량 접수 | 계좌·시장·종목 범위의 사전 holding/sellable/reserved, 제출 quantity, 접수 응답, 취소/종료 상태 | 미확인 |
| 포지션 전체 청산 | 위 필드 + 누적 actual filled quantity, 사후 holding/residual, 시간 순서 | 미확인 |
| 실비용 | 실행별 filled quantity/amount, commission, tax, currency/base, 부분체결·중복 주문 dedupe | 미확인 |
| 취소/정정 거절 형태 | 별도 broker record, raw status, 원주문 링크, 시간 순서 | 미확인 |

이 문서는 기존 기록의 증거 범위를 정리할 뿐, 기존 거래나 엔진 상태의 현재성·계좌 결속·실행 권한을 주장하지 않는다. `measurements.md`와 task 체크박스는 수정하지 않았다.

research-first: 기존 원장·영수증·검증 기록 재사용; grill-me: 승인 계약 변경 없음; impeccable: N/A — 문서 증거 감사, UI 영향 없음.

## 익명 원장 집계 재현 범위

원문 행을 복사하지 않은 익명 결과는 `/tmp/tossos-verify-trade-evidence-20261021-anonymous-journal-aggregate.json`에 보존했다. SHA-256은 `763e9323c988d7ac59a1a1bc0a62db79fbf041d9fcf7735a013645146683892e`다. 이 파일은 원장 경로·조회 SQL·열 이름·집계만 담으며 계좌·주문 식별자 값이나 주문 본문은 담지 않는다.

읽기 연결은 `file:~/.config/tossctl/journal.db?mode=ro`, `PRAGMA query_only=ON`이다. 다음 테이블과 열만 확인했다.

| 테이블 | 확인 열 |
| --- | --- |
| `fill_snapshots` | `order_id`, `symbol`, `market`, `state`, `terminal`, `quantity`, `filled_quantity`, `filled_amount`, `account_ref`, `trading_day`, `side` |
| `mutation_attempts` | `intent_id`, `broker_order_id`, `state`, `reason_code`, `account_ref` |
| `intents` | `id`, `account_ref`, `market`, `symbol`, `side`, `quantity`, `source` |
| `positions` | `id`, `account_ref`, `market`, `symbol`, `state`, `quantity`, `closed_at` |
| `position_adjustments` | `position_id`, `prev_quantity`, `new_quantity`, `broker_as_of`, `evidence` |
| `position_adoptions` | `symbol`, `market`, `quantity`, `observed_at`, `preimage_digest` |
| `trade_outcomes` | `position_id`, `realized_pnl_after_costs`, `initial_quantity`, `closed_at`, `cost_total` |

실행한 집계는 다음 다섯 개다.

1. `fill_snapshots`에서 `side=SELL AND state=FILLED`를 market별로 센다.
2. `fill_snapshots → mutation_attempts → intents` 조인에서 SELL·terminal 행을 센다.
3. `positions(state=CLOSED) → trade_outcomes` 직접 결속 행을 센다.
4. `trade_outcomes.cost_total` 비공란 행을 센다.
5. `fill_snapshots.state`와 `mutation_attempts.reason_code`에서 CANCEL/REPLACE_REJECTED 행을 센다.

앞서 쓴 61·67·44는 각각 **스냅샷 행**, **3테이블 조인 행**, **포지션-성과 행**이다. 서로 다른 단위이며, 이 감사는 distinct broker order·계좌·종목 수로 해석하지 않는다. 조인은 한 주문의 재관측이나 이력 구조에 따라 중복될 수 있다.

특히 67행 집계의 조인은 주문 ID와 intent ID 기준이며 계좌 일치 조건을 추가하지 않았다. 이는 탐색 집계이며 특정 계좌의 거래 결속 증거로 사용하지 않는다. 조회 전후 data_version 일치는 해당 연결에서 관측한 변화 여부일 뿐 브로커의 일관된 시점 증명은 아니다.

이 범위에서는 주문 전 `sellable` 스냅샷, 주문별 사후 계좌 보유량, 또는 broker execution fee/tax/currency 필드가 없다. 이는 원장 전체나 다른 저장소에 그런 필드가 없다는 주장이 아니라, 위에 명시한 현재 원장 테이블·열만으로 전량 및 비용 결속을 만들 수 없다는 제한이다. `position_adoptions`·`position_adjustments`도 이 조사에서 주문 ID와 동일 시점의 sellable 값을 결속하지 못했다.

## Manager 최종 검토

익명 집계의 실제 SQL·열 목록·결과와 현재 `stepCosts`, `sell.boundary.full_accepted`, `EstimateTradeCost` 호출 소스를 대조했다. terminal 행과 체결 행의 혼동, 주문 의도 통화와 실제 비용 통화의 혼동을 정정했다. 확인된 기존 매도 기록은 보완 근거로 남기되 2.1·2.2·2.9를 완료로 승격하지 않는다. 신규 테스트·전체 테스트는 실행하지 않았다.

Manager가 동일한 5개 집계를 별도 `mode=ro`·`query_only=ON` 연결의 읽기 트랜잭션에서 재조회해 Terra 결과와 모두 일치함을 확인했다. 이 재조회도 계좌·주문 식별자 없이 일치 여부만 출력했다.

다음 보완은 새 주문보다 **기존 거래의 브로커 원문 영수증 확보**가 우선이다. 실제 commission/tax와 통화·체결금액, 전량 여부의 거래 전후 잔고, 취소·정정 거절 별도 레코드와 원주문 링크가 필요하다. 위 결론 뒤 사용자 승인으로 아래의 제한된 브로커 원문 읽기 보완을 수행했다. 해당 원문에도 필요한 필드가 없으면 미확인 상태를 유지하고, 실측 주문은 별도 계획·승인 없이는 실행하지 않는다.

## 초기 로컬 조사와 v2 브로커 목록 조회 — 2026-10-21

사용자가 승인한 읽기 전용 범위에서 기존 명시 프로필의 **유효한 캐시 토큰**만 메모리에 읽어 공식 HTTPS 호스트에 조회했다. 공식 클라이언트는 401에서 refresh/retry할 수 있으므로 사용하지 않았다. 직접 요청은 TLS 검증 유지, 리디렉션·프록시·재시도·캐시 쓰기 금지, 요청당 10초 timeout으로 제한했다. 토큰·계좌·주문 ID·원문 body는 저장하거나 출력하지 않았다.

- 익명 결과: `/tmp/tossos-verify-trade-evidence-20261021/broker-orders-readonly-summary.json`
- SHA-256: `75bb21c422157489db2d6b002560f9a19882035c9d8422417949c4e299b84de5`
- 최초 로컬 집계기 오류 전 accounts 1회 + CLOSED 1회가 실행됐고 원문은 저장되지 않았다. 이후 로컬 합성 입력으로 집계기만 고친 뒤 accounts 1회, CLOSED 2페이지, OPEN 1페이지, detail 2회가 실행됐다. 총 HTTP GET은 8회다.
- HTTP 401/429/리디렉션/재시도는 없었고, 토큰 캐시 SHA는 전후 동일했다.
- `/accounts` 첫 행의 ASCII 계좌 참조·양수 sequence 검사는 통과했고, attestation 결속 비교는 메모리에서만 bool로 처리했다. 이는 실행 중 엔진이 같은 토큰·선택 sequence로 기동됐다는 증명은 아니다.

### 목록 범위와 원문 실행 필드

`internal/official/orders_reads.go:13-141`의 API 구조에 따라 `GET /api/v1/orders`는 `status`가 필수다. CLOSED는 cursor 페이지이고 OPEN은 pending 집합이다. 이번 CLOSED 범위는 현재 원장의 terminal SELL 날짜 범위 `2026-08-03`~`2026-10-19`, `limit=100`이었다.

| 조회 | 결과 | 해석 제한 |
| --- | --- | --- |
| CLOSED | 2페이지 후 `hasNext=true` | 2페이지 이후 이력은 읽지 않았다. 전체 기간·전체 주문 집계가 아니다. |
| OPEN | 1페이지, `hasNext=false` | 조회 시점 OPEN 집합이다. v2 집계기는 CLOSED와 OPEN을 한 set에 dedupe했으므로 서로 결합하지 않았다는 주장은 철회한다. |
| M50 | 영수증의 broker order ID가 CLOSED 목록과 정확히 일치하고 detail 1건을 읽음 | 일반 MARKET 매도 1주의 접수·체결 필드 확인이다. 전량 매도 증명은 아니다. |
| 대표 detail | 요청 상한 때문에 KR SELL 체결 1건만 추가 읽음 | US는 목록의 원문 필드 presence 집계만 있으며 detail은 읽지 않았다. |

v2 목록 집계는 local numeric classifier가 비문자형을 numeric으로 오인할 수 있었음이 확인돼, numeric count·비용 합계·상태별 execution 수치 집계를 모두 **폐기**했다. CLOSED와 OPEN을 결합한 v2 결과는 broker execution 비용 존재나 비용률의 근거로 사용하지 않는다. attestation account normalized equality도 v2에서는 검사하지 않았다.

따라서 이 절의 목록 결과는 페이지 경계·M50 list ID 일치·거절 상태의 미관측 범위만 남긴 초기 조회 기록이다. 비용 원문 증거는 아래 v3 detail 표본만 사용한다.

### 전량·거절 결론의 갱신

- M50 detail과 목록은 `status=FILLED`, 양수 filled quantity, filledAmount·commission·tax의 수치 필드와 timestamp/settlementDate 필드 존재를 보인다. 비식별 요약에는 개별 quantity·filledAmount 값이 보존되지 않았으므로 비용률은 계산하지 않는다. 그러나 같은 주문 직전 sellable·reserved, 사후 holdings/residual은 이 주문 읽기에서 제공되지 않는다. **매도가능 전량 접수 및 포지션 전체 청산은 계속 미확정**이다.
- 현재 읽은 CLOSED/OPEN 페이지에는 `CANCEL_REJECTED`·`REPLACE_REJECTED`가 0행이었다. 따라서 별도 broker record와 원주문 링크의 실제 형태는 여전히 미측정이다. 상한 소진으로 special-rejection detail은 실행하지 않았다.
- 결과 JSON은 field presence/null/type, 상태·방향·통화별 집계, 페이지 경계, ID dedupe 개수만 보존한다. 원문 주문 body·계좌·주문 ID는 보존하지 않았다.

### v3 detail 표본 보완 — 2026-10-21

v2 목록 집계의 숫자 판정은 `None`·빈 값이 아닌 비문자형을 수치처럼 처리할 수 있었으므로, 그 목록 기반 numeric count·비용 합계는 비용 근거에서 제외한다. v3는 nonempty decimal **문자열**만 허용하고 finite·nonnegative를 확인하며 missing/null/type/invalid을 분리했다. `NaN`·`Infinity`·bool·정수·빈 문자열·음수는 로컬 합성 검증에서 모두 거절했다.

- 스크립트: `/tmp/tossos-verify-broker-orders-readonly-20261021-v3.py`
- 스크립트 SHA-256: `6587d8c8a7b78c88e604227010f3f14a103dc0377f68cc0095857d0dc46b8215`
- 익명 결과: `/tmp/tossos-verify-trade-evidence-20261021/broker-orders-readonly-v3-summary.json`
- 결과 SHA-256: `27f2771e2bc4711e73655b410aec7fa39c3def882ba9b628b656263aa1899902`
- 이번 요청: accounts 1회, M50 detail 1회, 현재 원장의 같은 account/US/SELL/FILLED/symbol 후보 detail 1회. 이전 8회와 합쳐 **GET 11회**다. fallback 목록 조회·재시도·refresh·cache write·redirect는 없었다.

accounts 첫 행의 positive `accountSeq`와 ASCII-normalized account reference는 현재 attestation `account_ref`와 일치했다. 이 결과는 현재 캐시 토큰으로 받은 첫 account 행과 attestation의 결속만 확인한다. 캐시 토큰 provenance나 실행 중 엔진의 startup guard 통과를 증명하지 않는다.

두 detail은 모두 요청 ID echo, SELL, FILLED, 기대 symbol, 기대 currency를 만족했다. M50만 이전 제출 quantity와도 일치한다. US detail은 현재 원장의 account/market/symbol-bound 후보로 선택했으나 기존 제출 quantity와의 대조값이 없으므로 quantity check는 not_checked다; broker detail의 quantity와 filled quantity가 각각 1인 것만 관측했다. M50의 기대 symbol/quantity는 이전 보존 영수증의 `request.wire_body`에서 읽은 제출값이고, 영수증 response는 order ID echo만 보존한다. 따라서 이는 이전 제출 요청의 기대값과 새 detail의 대조이며, 이전 요청 전체를 새로 검증한 결과는 아니다.

| 표본 | 통화 | quantity / filled quantity | filled amount | commission | tax | commission bps | tax bps | timestamp / settlement field |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| M50 exact detail | KRW | 1 / 1 | 13050 | 2 | 26 | 1.532567049808429118773946360 | 19.92337164750957854406130268 | 둘 다 present |
| current-journal-bound US SELL detail | USD | 1 / 1 | 33.09 | 0.03 | 0.01 | 9.066183136899365367180417044 | 3.022061045633121789060139015 | 둘 다 present; submitted quantity check는 not_checked |

이 bps는 각각의 **관측 표본 계산값**이다. 비용 정책, 전체 계좌 비용률, 부분 체결·중복 주문을 제거한 집계, 모델 보정값으로 사용하지 않는다. 비식별·재현용 보충 요약은 [broker-receipt-samples-2026-10-21.json](broker-receipt-samples-2026-10-21.json)에 두며, 원본 v3 tmp summary의 SHA와 원본 스크립트 SHA를 포함하되 token cache hash와 raw body는 포함하지 않는다. 두 detail 모두 사전 sellable/reserved와 사후 holding/residual을 제공하지 않으므로 매도가능 전량 접수·포지션 전체 청산은 계속 미확정이다. CANCEL_REJECTED/REPLACE_REJECTED detail은 이번 상한 내에서 발견·조회하지 못했으므로 원주문 링크 형태도 미측정이다.
