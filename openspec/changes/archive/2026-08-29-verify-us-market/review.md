# Review: verify-us-market

Function Logic Map: applied — `analysis/function-logic/` 55 target(수정된 기존 함수 전부,
`check_analysis.py` 통과). 면제 아님.

## 1. Proposal freeze (Eng)

**전제 재검증(코드로)**: 제안이 "US는 원리적으로 불가가 아니라 도구가 막고 있을 뿐"이라고
주장했고, 구현 전에 세 가지를 확인했다.

- `FarBuyLimit`/`FarSellLimit`는 `lowerLimit > 0`·`upperLimit > 0`일 때만 clamp한다 →
  밴드 없는 시장에서 이미 동작한다. `TickSize`는 US 그리드($1 경계 0.0001/0.01)를 이미 구현.
- `PriceLimits`는 US에서 0/0을 반환한다(internal/official/price_limits_reads.go 주석).
- 막고 있던 것은 `preflightStatic`의 하드코딩된 `!= MarketKR` 한 줄과 cmd의 보유 선택뿐.

따라서 이 change는 **가격 산식을 추가하지 않는다**. 리뷰에서 가장 걱정한 항목(밴드 없는
시장의 안전가)이 신규 코드가 아니라 기존 검증된 경로라는 점을 근거로 범위를 좁혔다.

**브로커 계약 인용**(openapi.latest.json 원문): 조건주문 `symbol` "KRX: 6자리 숫자, US:
영문 티커" / 정정 `quantity` "KR 필수, US 전달 불가 → 400 us-modify-quantity-not-supported"
/ 정정 `price` "US: 소수점, $1 미만 넷째 자리·이상 둘째 자리". 전부 [문서 근거]이며 US
실동작은 이 change가 만들려는 **측정 대상**이다.

## 2. Code review

- **시장 비교의 대칭성**: 기존 규칙은 "KR이 아니면 배제"였고 이는 "이 도구는 KR만 안다"와
  "이 run은 KR이다"를 한 줄에 섞은 것이었다. 후자만 남겨 `SameMarket(MarketOf(symbol),
  r.market)`으로 바꿨고, KR run이 US 심볼을 거부하는 기존 동작은 전용 테스트로 고정했다
  (`TestADefaultRunStillRefusesAUSSymbol`).
- **plan digest 보존**: `Describe`의 영문 문구는 digest의 입력이다. 밴드 절을 시장별로
  분기하되 **KR 문자열은 바이트 단위로 종전과 동일**하게 유지했고, 고정된
  `planDigest20260727` 테스트가 그대로 통과하는 것으로 확인했다.
- **정직성 결함 하나를 테스트가 잡았다**: US 계획 줄이 "clamped inside the day's price
  band"라고 말하고 있었다 — 존재하지 않는 밴드를 근거로 승인받는 문구다. 시장별 절로 교체.
- **US advisory의 근거 분리**: 초안은 "KR의 order-hours-closed 실측을 US 근거로 쓰지
  않는다"고 *설명*했는데, 그 문장 자체가 US 화면에 KR 코드를 출력한다. 코드 이름을 아예
  빼고 "다른 시장의 실측을 이 시장의 근거로 쓰지 않는다"로 고쳤다.
- **폼이 시장을 실어 나르는지**: US 화면의 [재측정]이 KR run을 시작하면 사용자가 보지 않은
  시장의 계획을 승인하게 된다. 두 start form 모두 hidden market을 싣고 전용 테스트로 고정.

**남은 관찰**: `consoleProbeSymbol`의 US 분기는 보유 종목을 프로브 심볼로 쓴다. 보유가
없으면 빈 문자열이 되고 매수측 단계는 runner의 기존 게이트가 사유와 함께 배제한다 —
빈 심볼로 주문이 나가는 경로는 없다(`farBuySpec`이 마지막 체결가를 못 읽어 실패).

## 3. Security review (CSO 관점)

**새 위험**: 이 저장소에서 처음으로 US 실주문이 나갈 수 있게 된다.

**노출 상한은 무변경**: 1주 정수(`MinQuantity`), 체결 불가 지정가(오프셋 20%), 단계 내
취소, 시장가 없음, 소수점 없음, 계획 밖 전송 금지. 실계좌 기준 MWG($1.27) 1주 ≈ $1.3.

**격리**: 시장별 기록 파일로 KR 증거가 오염되지 않는다 — M3 resume digest가 걸린 KR 파일은
바이트 무접촉이다. 한 시장의 판정이 다른 시장의 `settled`/`RedoSet`을 움직일 수 없다는 것을
콘솔·verifylive 양쪽 테스트로 고정했다.

**게이트**: 이 change는 `ProtectionReady`·automation gate·엔진 기동에 손대지 않는다. US를
자동 매매해도 되는지는 2c의 2.6과 §0.7이 정한다.

**잔여 위험(수용)**: ① 밴드 없는 시장에서 20% 떨어진 지정가를 브로커가 거부할 수 있는지
[미측정] — 거부는 전송되지 않았다는 뜻이므로 안전 방향이고, 관측되면 M-번호로 기록한다.
② US 휴장 응답 [미측정] — advisory가 그렇게 말한다. ③ USD 잔고($685.45)는 매수 프로브에
충분하나 부족 시 기존 fail-closed 경로가 처리한다.

## 4. QA

- `go build ./...`, `go vet ./...` clean. `go test ./...` 3111 passed.
- RED 관측: US 계획이 비어 있음(preflight 배제) → GREEN(계획 생성), US 정정에 수량이
  실림 → GREEN(미전송), 시장별 기록 부재 → GREEN, 콘솔 폼이 시장을 잃음 → GREEN.
- 실계좌 QA는 아직 없다 — 다음 US 정규장에서 사용자가 수행한다.

## 5. 완료 조건

- 미완료 태스크 0, `check_analysis.py` 통과, PM check 통과, gate 8/8.
- 사용자 조치: 새 빌드 설치 + 콘솔 재시작 후 `/verify?market=US`.
