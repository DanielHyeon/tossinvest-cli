판정 FAIL

정규식 변경 자체의 런타임 결함이나 새 주문 우회 경로는 발견하지 못했습니다. **blocker 0건**입니다. 다만 “주문 검증은 숫자 전용”이라는 스펙·안전성 전제가 실제 코드와 다르므로 현재 상태의 승인에는 반대합니다.

아래 경로는 사본 루트 기준입니다. `C/`는 `openspec/changes/a187-strict-quote-readers-accept-kr-alphanumeric-symbols/`입니다.

**발견사항**

| 번호 | 분류 | file:line | 문제 | 근거 | 제안 |
|---|---|---|---|---|---|
| 1 | major | `C/specs/official-read-compatibility/spec.md:13`; `internal/orderintent/intent.go:101`; `internal/orderintent/a187_order_policy_test.go:6` | **숫자 전용 시장 추론을 숫자 전용 주문 검증으로 오인함.** | `looksLikeKRSymbol`은 US→KR 자동 추론 조건일 뿐, 영숫자 KR 주문을 거절하지 않음. `NormalizePlace`는 명시적 `Market:"kr", Symbol:"0101N0", Side:"buy", OrderType:"limit", Quantity:1, Price:1000`을 그대로 반환하는 경로가 있음(`intent.go:84–145`). 새 시험은 추론 함수만 검사함. | 스펙을 실제 유지 대상인 “시장 추론·주문 정책 무변경”으로 정확히 기술. 모든 KR 주문의 숫자 전용 거절이 요구사항이라면 별도 승인 범위로 다루고 실제 주문 진입점 시험 필요. |
| 2 | minor | `C/proposal.md:48`; `tools/a128-jev-shadow-probe/collect.go:189`; `internal/official/client.go:335` | **프로브의 실제 GET 총량은 같지 않음.** 인증·재시도 영향도 빠짐. | 수집기는 리더 호출 **전** `countRead`를 증가시켜 기존 문법 거절도 셈. 변경 후 해당 종목은 실제 GET·토큰 취득 경로에 도달함. 429는 수집기 재시도(`collect.go:194`), 401은 토큰 갱신·재전송(`client.go:359`) 가능. | 리더 호출 계수와 실제 HTTP 시도를 구분. 재시도 없는 경우 영숫자 종목 N개당 GET N회 증가, 재시도·인증은 별도라고 수정. |
| 3 | minor | `internal/official/a187_kr_alphanumeric_test.go:18`, `:44`; `C/analysis/mutation/mutation.py:28` | **6개 변이 검출로 전체 문법 계약을 입증하지 못함.** | 수락 사례의 영문자는 모두 5·6번째 위치. 개행 입력과 영숫자 KR의 잘못된 응답 메아리도 없음. 아래 구체적 생존 변이 참조. | 앞자리 영문·전부 영문·개행 거절·KR 메아리 불일치 사례 추가. |
| 4 | minor | `C/review.md:32`; `C/tasks.md:16`; `tools/a087-market-sell-probe/contract_test.go:80` | **a087 주문 정책 검증의 실행 범위가 불명확하게 보고됨.** | 제공된 6개 패키지 시험에는 `tools/a087-market-sell-probe`가 없음. 로그의 `TestA087…` 두 개는 가격제한폭 필드 시험이며 주문 스키마 시험이 아님. 스키마 시험 자체도 `goodSpec()`의 `005930`을 사용함(`order_test.go:15`). | 실제 통과한 시험명을 명시하고 a087 실행 정책·스키마 검증은 별도 미실행으로 구분. |
| 5 | note | `tools/a128-jev-shadow-probe/profile.go:20`; `collect.go:485`, `:539` | 영향은 `current` 입력의 결손 해소에 그치지 않음. **`initial/compare` 양쪽도 변경됨.** | `initialState`가 `TopOfBook`을 복사함. 변경된 입력은 판단 결과를 거쳐 J1/J2 선별과 판단 원장에 반영됨. 원장에는 기존 `a128-s3` 또는 `initial-fields-v1`이 그대로 기록됨(`collect.go:562–566`). | 반영 경계를 두 입력군 모두에 적용. 입력 digest·토큰 수·선별 결과·수집 시간도 달라질 수 있음을 영향 기록에 포함. |

1번의 주문 경로는 **a187 이전부터 존재하는 경로**입니다. CLI → `NormalizePlace` → `trading.Service.Place` → 공식 주문 경로에서 영숫자 거절은 없습니다(`cmd/tossctl/order.go:214`, `internal/trading/service.go:269`, `internal/hybrid/broker.go:65`, `internal/official/orders_write.go:120`). 실행 승인·설정·confirm 가드는 유지됩니다. 따라서 이를 **a187이 새로 만든 blocker로 분류하지 않았습니다.**

**변이 검토**

다음은 시험 입력과 구현을 대조한 정적 판단입니다. 변이를 실행하지 않았습니다.

| 변이 | 판단 | 근거·보강 사례 |
|---|---|---|
| KR/US 정규식 선택 뒤바꿈 | 기존 시험에서 검출 가능 | KR `0101N0`은 US 문법에서, US `AAPL`은 KR 6자리 문법에서 거절됨. |
| `{6}` → `{5,6}` | 기존 시험에서 검출 가능 | 새 거절 사례 `0101N`, 기존 `00593`이 통과하게 됨. |
| `[A-Z0-9]` → `[A-Z]` | 기존 시험에서 검출 가능 | 숫자 KR와 영숫자 수락 사례가 실패함. |
| `^[0-9]{4}[A-Z0-9]{2}$`로 축소 | **생존 예상** | 현재 KR 수락 사례가 모두 첫 네 자리 숫자임. 계약상 유효한 `A12345`, `ABCDEF`를 거절해도 잡지 못함. |
| `(?m)^[A-Z0-9]{6}$`로 변경 | **생존 예상** | 종목 개행 사례 없음. `0101N0\n` 및 `\n0101N0`을 받아들이는 오류가 남음. |
| 현재가 메아리 조건을 `echoed != symbol && currency != "KRW"`로 변경 | **생존 예상** | 기존 메아리 불일치 시험은 US뿐(`strict_quote_reads_test.go:287`, `:304`). 새 KR 시험은 일치 응답뿐임. KR 요청에 다른 KR 종목 응답을 넣어야 검출함. |

**확인했지만 문제없던 것**

- **직접 호출 목록:** 비시험 Go 소스의 호출·참조를 검색하고 호출부를 읽었습니다. 두 호가 리더 호출은 총 **7자리**입니다. `officialbars/quote.go:113·117`, a128 `collect.go:453`, a140 `source.go:148·154`, a112 `main.go:87·91`. proposal 목록과 일치합니다. `PollQuoteL1` 운영 호출은 a140 `shadow.go:289` 하나이며 엔진 패키지 직접 호출은 없습니다.

- **a140의 추가 차단:** `officialbars.checkSymbol`은 문법 차단벽이 아니지만, a140의 유니버스 선정은 숫자 6자리만 받습니다(`shadow.go:33`, `:275`). 영숫자 코드가 이번 변경으로 a140의 `evaluateSetup`·사이징·그림자 원장까지 새로 들어가는 경로는 없습니다. 생산 evidence 파일 사용도 별도 가드가 있습니다(`main.go:102`, `guard.go:33`).

- **공용 검사기의 분봉 소비자:** proposal의 호가 목록 외에 `officialbars/producer.go:258`, a128 `collect.go:609`, a140 `source.go:142`, a128-d2 `receipts.go:188`, `official/a130_intake_candle.go:31`도 공용 검사에 연결됩니다. 다만 분봉의 기존 수락 집합은 이미 `[A-Z0-9]{6}`였으므로 이번 변경으로 입력 범위가 늘지 않습니다. `RawMinuteCandles`는 비시험 호출·메서드 참조를 찾지 못했습니다.

- **주문·체결·Guardian·대사:** 새로 수락된 호가 결과에서 해당 운영 경로로 연결되는 호출은 발견하지 못했습니다. a128은 상태·판단·라벨 원장, a112는 보고서가 소비자입니다. **인증과 공유 조회 예산에는 새 요청이 도달**하므로 이를 “영향 없음”으로 묶으면 안 됩니다.

- **응답 검증:** 통화는 종목 문자열이 아닌 명시적 시장에서 정합니다(`strict_minute_candles.go:200`). 호가는 동일한 통화·시각·첫 호가 해독기를 사용하며, 반환 `Symbol`은 요청값입니다(`strict_quote_reads.go:150`, `:156`). 현재가는 문자열 완전 일치로 메아리를 검사합니다(`:395`). 영숫자 전용 우회 분기는 없습니다. 호가 본문에 종목 필드가 없다는 제한은 공식 스냅숏 `openapi.latest.json:8443`과 일치합니다.

- **정규식:** 현재 `^[A-Z0-9]{6}$`는 정확히 ASCII 대문자·숫자 6개만 받습니다. 소문자·비 ASCII·빈 문자열·개행은 거절합니다. Go의 `$`는 여기서 마지막 `\n` 앞을 허용하지 않습니다. 로컬 Go 1.26.5 구현에서 기본 `OneLine` 플래그, `$`의 `OpEndText` 변환, `after == -1` 조건을 확인했습니다(`/usr/local/go/src/regexp/regexp.go:130`, `syntax/parse.go:960`, `syntax/prog.go:274`). 실행 시험으로 확인한 결과는 아닙니다.

- **삭제된 거절 사례:** `00593A`·`0101N0`은 새 수락 표에 실제로 옮겨졌고, 두 리더 각각 요청 계수·정확한 query·반환 종목·통화를 검사합니다(`a187_kr_alphanumeric_test.go:18–35`). 다만 `hitCount`는 OAuth 요청을 제외한 데이터 요청 계수입니다(`strict_minute_candles_test.go:34`).

- **스펙 보존:** 기존 a087 공개 명세 문구 시나리오는 delta에 보존됐습니다. `Strict engine candle contracts are preserved`의 봉 순서·커서·close-to-open·신원 검증 제한과 충돌하지 않습니다. 문제는 기존 요구사항의 a087 제한을 넘어 **전체 order intent가 숫자 전용이라는 전제를 새로 적은 부분**입니다. 정본에 옛 호가 제한이 남아 있는 것은 아직 archive 전인 상태와 구분했습니다.

- **문서·기존 증거:** design의 옛 행 번호는 base AST 기준이라고 명시돼 있어 현재 행 번호 차이를 오류로 보지 않았습니다. 제공 로그의 PASS 507개, 하위 포함 919개, FAIL 5개는 재집계와 일치합니다. 단 `review.md:23`의 래퍼 변이 “11갈래”는 로그상 **실패 하위 시험 10개＋부모 1개**입니다. 변이 대상 파일의 SHA-256은 제공 변이 기록과 일치하지만, 이는 새 시험 실행 증명이 아닙니다.

- **a184 부록:** 제공 diff에서 명백한 결함은 발견하지 못했습니다. 규칙·커밋 상수 교체가 일치하고, 추가 시험은 구판 규칙 정지·구판 승인 거부·42/45 불완전 회차의 제안 0건 및 조회 0회를 검사합니다. 보이지 않는 helper와 재개 구현의 정확성까지 확인한 것은 아닙니다.

**보지 못한 것**

- 파일 수정, Git 명령, 네트워크·공식 API 호출, 빌드·시험·변이·전체 gate 실행 **모두 하지 않았습니다**.
- 사본의 실제 HEAD와 base 출처는 독립 검증하지 못했습니다. 기존 실패 5건의 base 재현은 제공 로그 대조이며 직접 재현이 아닙니다.
- a184 나머지 소스, 공식 영숫자 응답 실측, 운영 프로브의 실제 기동 시각·사용 키·예산 모드는 확인하지 못했습니다.
- CodeGraph·AST는 제공 산출물과 관련 소스를 읽었으며 재생성하지 않았습니다. 전체 저장소·전체 분석 산출물 전수감사는 아닙니다.
- ai-memory 검색 도구는 없었고 네트워크 금지로 서비스 조회를 생략했습니다. gstack도 실행하지 않았습니다. `impeccable: N/A — UI 변경 없는 백엔드·시험·문서 리뷰`.