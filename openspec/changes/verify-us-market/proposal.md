# Change: verify-us-market

## Why

사용자 결정(2026-07-27): US 시장 측정을 진행한다.

2b 검증 도구의 주문 단계는 KR 전용이다 — `MarketKR is the only market this tool's
mutating steps run against`(pricing.go), preflight가 비KR 심볼의 mutating 단계를 사유와
함께 건너뛴다(runner.go), 보유 종목 선택도 KR만 고른다(cmd/tossctl/verify.go). 그래서
US 정규장(22:30–05:00 KST)에는 측정할 수 있는 것이 없다.

그런데 제품의 운용 시장은 KR/US 둘 다이고(docs/ROADMAP.md, order-execution "시간 규율"의
ET·DST 세션 판정, USD 잔고 fail-closed), **2b task 2.5는 조건주문 능력을 "시장·유형별"로
측정하라고 이미 요구한다**. US를 측정하지 않으면 2c의 2.6이 US를 자동 진입 금지 목록에
올리고 엔진은 KR만 매매하게 된다.

브로커 쪽은 US를 막고 있지 않다(openapi 실문서):

- 조건주문 `symbol`: "KRX: 6자리 숫자, **US: 영문 티커**". SINGLE·MARKET에 시장 제한 문구 없음
- 주문 정정 `price`: "US: 소수점(달러). $1 미만 넷째 자리, $1 이상 둘째 자리"
- 주문 정정 `quantity`: "KR 필수 / **US 전달 불가** — 제공 시 `400 us-modify-quantity-not-supported`"
- 일일 가격제한폭(`GET /api/v1/price-limits`)은 US에서 null — 밴드가 없다

계좌 조건도 충족한다: US 보유 MWG 115주($1.27), 주문가능 USD $685.45.

## What Changes

- **시장이 run의 매개변수가 된다**: `Options.Market`(zero value = KR — §0.2 기존 동작
  보존). preflight는 하드코딩된 KR 대신 **run의 시장**과 단계 심볼의 시장을 비교한다.
  시장이 다른 mutating 단계는 종전처럼 사유와 함께 skip된다.
- **증거 기록을 시장별로 분리**: capability는 (계좌, 시장)의 속성이다. 한 파일을 공유하면
  US 판정이 KR 단계를 settled로 만들어 KR 측정을 삼킨다. KR은 `capability-verify.jsonl`
  그대로, US는 `capability-verify-us.jsonl`. StepCount·RedoSet·resume·plan digest·
  one-run-per-process는 파일 단위로 그대로 동작하며 **KR 기록은 바이트 무접촉**이다
  (M3 resume digest `sha256:fac7f233…` 보존).
- **US 정정은 가격 전용**: `AmendIntent.Quantity`를 US에서 nil로 보낸다(브로커 계약).
  승인 목록의 문구도 시장에 맞춘다. 정정이 새 ID를 발급하는지는 US에서도 측정 대상이다
  (2c 귀속 규칙 입력).
- **가격 규칙은 신설하지 않는다**: `FarBuyLimit`/`FarSellLimit`는 밴드 값이 0이면 "밴드
  없음"으로 이미 처리하고, `TickSize`는 US 그리드($1 경계, 0.0001/0.01)를 이미 구현한다.
  이 change는 그 경로를 **막지 않을 뿐**이며 새 산식을 만들지 않는다.
- **US 장시간 advisory**: 두 번째 달력을 만들지 않고 `internal/clock`(IANA
  America/New_York, DST, 정규장 09:30–16:00)을 사용한다. KR advisory는 실측 422 코드를
  계속 인용하고, US는 **아직 실측이 없음을 명시**한다(휴장 응답 미관측).
- **콘솔의 시장 선택**: `/verify?market=us` — 진행 상황·이어하기·재측정·단계 목록이 선택
  시장의 기록을 가리킨다. 승인 형식은 무변경(클릭 1회 — console-click-approval).
- **US 프로브 심볼**: 하드코딩하지 않는다. 계좌의 사용 가능한 US 보유 종목을 쓰고
  (`--holding-symbol`/`--symbol`로 덮어쓸 수 있다), 없으면 US run은 보유 필요 단계를
  종전 규칙대로 skip한다.

## Non-Goals

- 소수점(fractional) 주문 — `MinQuantity` 1주 정수 유지. US MARKET SELL 전용이며 이 도구는
  시장가를 내지 않는다(기존 판단 유지).
- 자동 환전·FX 경로 — USD 잔고 부족은 기존 fail-closed 그대로(order-execution 스펙 소유).
- 엔진의 US 매매 활성화 — 이 change는 **측정 도구**만 넓힌다. 무엇을 자동 매매해도 되는지는
  2c의 2.6 결과와 게이트 ON(§0.7)이 결정한다.
- KR 경로의 동작 변경 — 기본값·기록·문구·digest 전부 보존.
- 확장 세션(프리마켓·애프터마켓) 판정 — 미측정으로 남긴다.

## Capabilities

### Added Capabilities

- `operator-console`: "시장별 검증 화면"

## Impact

- Affected code: `internal/verifylive`(Options.Market·preflight·amendOrder·advisory·
  기록 경로 기본값), `cmd/tossctl`(verify·console 배선: 시장 플래그, 시장별 기록 경로,
  보유·프로브 심볼 선택), `internal/console`(시장 선택 화면·라우트 파라미터)
- 안전 검토(§0): 주문 1건당 노출은 종전과 같은 **1주·체결 불가 지정가·단계 내 취소**이며
  US에서는 약 $1.3(MWG 기준)이다. 시장가·소수점·자동 환전은 없다. 승인 레일(계획 목록·
  클릭 승인·계획 인가·상한)과 프로세스 경계는 무변경. KR 기록·KR 동작은 무접촉.
  새 위험: US 실주문이 처음으로 나간다 — 그것이 이 change의 목적이며 사용자 결정이다.
- PM: registry bootstrap allowlist + PM fixture 등재(archive 시 제거)
