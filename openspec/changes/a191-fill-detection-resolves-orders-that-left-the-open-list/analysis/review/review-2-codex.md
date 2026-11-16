# 적대 리뷰 2회차(최종) 원문 — Codex(gpt-6-astra), codex-cli 0.165.0, `codex exec -s read-only`, 2026-11-16

## 프롬프트
```
Second and FINAL adversarial review round of a HIGH-RISK design in a live-money Go trading engine. Read-only: do not modify files, no network, no `go test ./...`, no make targets.

Change: openspec/changes/a191-fill-detection-resolves-orders-that-left-the-open-list/ (proposal.md, design.md, specs/fill-detection/spec.md, tasks.md, review.md). Round 1 findings and the author's dispositions are in review.md ("리뷰 1회차" table) and the raw round-1 text is analysis/review/review-1-codex.md. The author corrected design D3-2 (duplicate rows), D3-7 (only status fields come from the CLOSED row; quantities/price/amount/filledAt stay from the single read), D3-8/D3-9 (facts about ScanOrders and absent execution), D10 (budget numbers), D11 (indirect delay), Q3 safety-net paragraph, N4 row, Q4, Q5(c)(d), proposal recovery wording. New guards were deliberately NOT added; they are listed as Manager decisions M6-M10.

Tasks, citing file:line for every claim:
1. For each round-1 finding marked "수정" (3, 4, 7, 8, 12 and the factual parts of 1, 2, 5, 9, 10), say whether the correction is now right or still wrong.
2. Is D3-7 (replace RawStatus/Canceled/CanceledAt/Derived only) internally consistent with Snapshot fields in internal/filldetect/detect.go and with how JournalLedger.Apply (internal/filldetect/ledger.go) and RecordFill (internal/journal/fills.go) consume them — e.g. does any downstream code read a field the author forgot (Terminal flag, Reason/Detail, state string)? Does a terminal state with a single-read execution.filledAt cause any ordering/refusal problem (classifyFillRefusal)?
3. Any NEW defect introduced by the corrections, or any remaining claim in design.md that the code contradicts.
4. Are the Manager-decision options M6-M10 framed correctly (no option that is unsafe presented as safe)?
Do not repeat round-1 findings that the author already deferred to a Manager decision unless the framing is wrong.

Output: numbered findings with severity (BLOCKER/MAJOR/MINOR/NIT), claim, file:line evidence, proposed change; then a one-paragraph verdict. Concise, no praise.
```

## 응답
아래 문서 경로는 `openspec/changes/a191-fill-detection-resolves-orders-that-left-the-open-list/` 기준입니다.

1. **MAJOR — 수정한 D3 계약이 spec에 반영되지 않았다.**  
   D3-2는 raw 바이트 동일 중복을 허용하고, D3-7은 상태 필드만 교체한다(`design.md:68–77`). 그러나 spec은 여전히 “행이 정확히 하나”, “그 행이 스냅숏을 대신해야 한다”고 명령한다(`specs/fill-detection/spec.md:7–8`). 사유 표도 ambiguous를 서로 다른 범위에만 한정한다(`design.md:97`).  
   **변경:** spec에 raw 바이트 중복 제거 규칙과 교체할 네 필드를 명시하고, ambiguous에 같은 범위·다른 payload도 포함한다. 1회차 **#3·#4는 설계 본문에서 수정됐지만 계약 전체로는 미완료**다.

2. **MAJOR — 기존 래치·legacy 예외를 인정한 수정과 나머지 계약이 충돌한다.**  
   N4는 CLOSED 미해결 시 기존 UNKNOWN refusal·즉시 래치가 발생한다고 정정했다(`design.md:210`). 하지만 spec은 비종결 단건이라는 조건만으로 “체결 기록·새 진입 래치 없음”을 보장한다(`specs/fill-detection/spec.md:29–31`). 단건 자체가 체결 증가를 보고하면 기록되고, UNKNOWN이면 래치가 생긴다(`internal/journal/fills.go:385–409,464–475`; `internal/filldetect/detect.go:324–327`). 또한 Q5(a)/(d)는 legacy 예외를 인정하지만 Q5(b)와 성공 시나리오는 여전히 열린 주문에서 빠진다고 단정한다(`design.md:236–247`; `specs/fill-detection/spec.md:12`). 실제 로컬 열린 주문은 legacy가 남을 수 있는 `TrackedFillOrders` 결과를 소비한다(`internal/journal/fills.go:1604–1614`; `internal/reconcile/compare.go:201–212`).  
   **변경:** “CLOSED 조회 자체가 추가하는 기록·래치 없음”으로 한정하고, 추적 종료 보장에 legacy 전제를 붙인다. **#9 사실 수정은 맞고, #10 수정은 부분 완료**다.

3. **MAJOR — M8의 ‘Apply 뒤로 이동’만으로 보호 경로 지연을 해결한다는 선택지 구성이 불완전하다.**  
   M8은 짧은 시한·총량 제한과 Apply 이후 이동을 대안으로 제시한다(`review.md:90`). 후자는 이미 수집한 스냅숏 적용 지연은 줄이지만, 진행 중 CLOSED 요청이 공유 wire를 점유하는 문제는 남긴다. Critical은 다음 소유자 선택에서 우선할 뿐 현재 소유자를 선점하지 않는다(`internal/accountbudget/budget_linux.go:549–584`); 응답 body 종료까지 소유권이 유지된다(`internal/accountbudget/transport_linux.go:177–188`).  
   **변경:** (나)의 미해결 위험을 명시하고, 어느 배치든 허용 지연·총량 한도를 결정하도록 한다. (가)의 2초도 페이지별인지 전체 scan인지 명시한다. **#5의 D11 사실 수정은 맞지만 M8 선택지 설명은 부족하다.**

4. **MAJOR — M9는 대사 차단을 보장된 대체 안전장치로 설명한다.**  
   추천 근거는 “미해결 주문은 대사가 심볼·계좌로 막는다”다(`review.md:91`). 그러나 filldetect와 reconcile은 별도 루프이며(`cmd/tossctl/engine.go:725–742`), filldetect는 대사 완료를 기다리지 않고 freshness를 찍는다(`internal/filldetect/detect.go:346–348`). missing-order의 최초 블록도 계좌 전체가 아닌 심볼 범위다(`internal/reconcile/mismatch.go:1418–1427`).  
   **변경:** 도장 선택을 “기존 동작 유지에 따른 잔여 위험 수용”으로 표현하고, 대사 관측 전 공백과 차단 범위를 명시한다. 정책 자체를 다시 지적하는 것이 아니라 **안전하다는 추천 근거가 성립하지 않는 문제**다.

5. **MAJOR — 새 Manager 결정이 착수 게이트에 연결되지 않았고 안전 통과 선언도 남았다.**  
   M8은 2단계 착수 전 필수다(`design.md:171`; `review.md:90`). 그러나 tasks의 착수 조건과 승인 항목은 M1–M4만 요구한다(`tasks.md:3,15`). Pre-Edit 초안은 여전히 “exit 루프 호출 0”으로 손절 즉시성 통과를 선언한다(`review.md:49–50`). 이는 간접 지연을 인정한 D11과 충돌한다(`design.md:166–171`).  
   **변경:** M6–M10의 결정·처분을 착수 체크리스트에 연결하고, M8 결정 및 검증 전 안전 판정을 미확정으로 바꾼다.

6. **MINOR — Q5의 ‘위험 버킷은 종결 해제 경로만’은 여전히 틀렸다.**  
   Q5는 이제 양의 delta를 인정하면서도 마지막에는 위험 버킷이 종결 해제만 한다고 적는다(`design.md:252–256`). 실제로 `RecordFill`은 체결 적용 후 종결 해제를 호출하며(`internal/journal/fills.go:526–533`), 등록된 BUY의 양의 delta는 위험 버킷 체결 전이를 계산·저장한다(`internal/journal/risk_bucket_fill.go:208–243`).  
   **변경:** 해제만이라는 설명을 delta=0 사례로 한정하고, BUY·delta>0이면 체결 전이도 발생한다고 적는다.

요청한 수정별 판정과 D3-7 확인 결과:

| 대상 | 판정·근거 |
|---|---|
| #1 사실 | **맞음.** execution 부재/null의 0 처리 설명은 코드와 일치한다(`design.md:81–84`; `internal/filldetect/payload.go:81–99,178–184`). |
| #2 사실 | **맞음.** 빈 커서를 성공 종료로 처리한다(`design.md:78–80`; `internal/execgw/indoubt.go:754–756`). |
| #3 | **D3-2는 맞음.** 계약 반영은 finding 1처럼 미완료다. |
| #4 / D3-7 | **필드 전달은 맞음.** Snapshot에 독립 Terminal/State/Reason/Detail은 없고, Apply가 모두 `Derived`에서 읽는다(`internal/filldetect/detect.go:194–226`; `internal/filldetect/ledger.go:54–70`). 네 필드 교체에서 빠진 downstream 상태 필드는 발견하지 못했다. |
| 체결 시각·거부 | **새 거부 문제 없음.** `HasBrokerTime`일 때만 단건 시각을 전달한다(`internal/filldetect/ledger.go:72–78`). 과거 시각이어도 누적 수량이 같으면 허용한다. 수량 감소 또는 과거 시각+다른 수량은 기존대로 거부한다(`internal/journal/fills.go:682–704`). 종결 자체는 역순 거부 조건이 아니다. |
| #5 사실 | **맞음.** collect가 Apply보다 먼저이고 기본 요청 시한은 15초다(`internal/filldetect/detect.go:293–317`; `internal/official/client.go:22,137`). M8 설명 문제는 finding 3이다. |
| #7 | **0.1 req/s 계산은 맞음.** 조회당 최대 3페이지이므로 3/30이다(`design.md:155`; `internal/execgw/indoubt.go:743–752`). |
| #8 | **맞음.** 외부 포지션·외부 주문은 그 자체로 `BlocksEntry`를 만들지 않는다(`design.md:218–224`; `internal/reconcile/compare.go:306–313,410–417,468–472`). |
| #9 | **N4 사실 수정은 맞음.** 후속 주문이 있는 PENDING은 UNKNOWN이고 기존 Apply 경로가 차단한다(`internal/brokerstate/derive.go:392–394`; `internal/filldetect/detect.go:324–327`). |
| #10 | **proposal·Q5(c)/(d)는 수정 방향이 맞음.** 수량 차이 없으면 수렴은 즉시 반환하고, 크레딧 없는 블록은 남는다(`internal/reconcile/converge.go:142–145`; `internal/reconcile/mismatch.go:662–668`). 남은 무조건적 표현은 finding 2다. |
| #12 | **맞음.** B45는 `case@475`다(`design.md:210`; `internal/brokerstate/derive.go:475`). |

M6의 존재성 확인 추천, M7의 CLOSED 전용 검사, M10의 명시적 잔여 위험 선택에는 별도 framing 결함을 발견하지 못했다(`review.md:88–89,92`; 위험 명시 `design.md:218–224`). M8·M9는 위 수정이 필요하다.

**판정:** 이 문서 묶음 그대로는 구현 착수 불가다. D3-7의 상태·체결 필드 분리는 downstream과 맞지만, spec의 이전 계약, 미완성 착수 게이트, M8/M9의 안전 설명을 먼저 정합화해야 한다. HEAD `a392308e8`와 현재 미커밋 문서를 정적으로 검토했다. 파일 수정·네트워크·시험·make 실행은 없었다. ai-memory MCP 부재로 추가 조회하지 않았고 저장소의 1회차 원문을 사용했다. `impeccable: N/A — backend/documentation-only`.