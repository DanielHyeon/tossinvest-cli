# Function Logic Map: `Detector.collect` (internal/filldetect)

- Source: `internal/filldetect/detect.go` 375~471 (편집 뒤). 편집 전: `../../pre-edit-ast/filldetect--detector.collect.ast.json`(363~458, sha 544129682682, 분기 19, return 12, call 43).
- AST evidence: `ast.json`(편집 뒤 sha 889e3839105a, 분기 19, return 12, call 46). 명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/filldetect/detect.go --func Detector.collect`.
- Risk scan: `risk-pattern-report.md`(일치 0).
- 호출자(CodeGraph 1.6.0): `Detector.pollLocked` 하나 — poll 뮤텍스 아래에서만 돎.

## 재번호 (분기 번호는 위치다)

편집 전·뒤 AST 의 분기 목록을 순서대로 맞춰 보면 **종류·순서가 같고 줄만 +12** 임(B1 if@365→377 … B19 if@449→461). 원인은 함수 위
타입 선언에 더한 필드(`TrackedOrder.IntentID` 2줄, `Detector.Alerts` 2줄, `Detector.leftOpen*` 4줄, `Cycle.ClosedReads·Unresolved` 4줄 = 12줄)다.
함수 본문의 새 줄은 call@467 `d.noteLeftOpen`(+ `len`)과 call@470 `d.settleLeftOpen` 둘 — 분기 추가 0. 그래서 번호는 그대로 쓰고 줄만 옮긴다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `d.Orders` | 비nil `execgw.OrderPager` | 운영: `execgw.OfficialOrders` | OPEN 걷기 실패 → 주기 실패(B1) |
| `d.Tracked` | 추적 주문 + 선택 계좌 | 운영: `JournalTracked` | 실패·빈 계좌 → 주기 실패(B2·B3) |
| `d.Order` | 단건 raw 조회 | 운영: `OfficialOrders` | 실패 → 주기 실패(B15) |
| `cycle` | 그 주기의 `*Cycle`(포인터가 주기 신원) | `pollLocked` | — |
| 불변식 | 정규 키로만 짝짓기; 종결 목록은 상태 필드만 바꿈 | `closed_lookup.go` | 받지 않으면 단건 그대로 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@377) | OPEN 걷기 오류 | — | return@378 | `TestRateLimitedPollIsClassifiedAsOutage`·`TestTransportOutageIsClassifiedSeparately` |
| B2 (if@383) | `TrackedOrders` 오류 | — | return@384 | base 커버리지 count=0(기존 부채, a191 무관) |
| B3 (if@387) | 선택 계좌 빈 값 | — | return@388 | count=0(같음) |
| B4 (range@394) | 추적 순회 | 맵 채움 | — | 전부 |
| B5 (if@396) | 추적 키 오류 | — | return@397 | count=0(같음) |
| B6 (if@399) | 키 중복 | — | return@400 | `TestPollRejectsDuplicateTrackedCanonicalIdentity` |
| B7 (range@411) | OPEN 행 순회 | — | — | 전부 |
| B8 (if@413) | OPEN 행 파싱 실패 | — | return@416 | `TestMalformedOrderPayloadFailsTheCycle` |
| B9 (if@421) | OPEN 행 정규 키 완전 | — | — | `TestPollReadsTrackedOrdersThatLeftTheOpenList` 외(실측) |
| B11 (if@422) | 그 키가 추적됨 | lineage, `seen[key]=true` | — | 참 `TestPollSeparatesReusedOrderIDsByCanonicalScope`, 거짓 `TestPollWalksTheOpenListToTheLastPage` |
| B10 (else@426) | 정규 키 불완전 | — | — | `TestPollRejectsIncompleteOpenScopeForATrackedOrder`, `TestA191IncompleteOpenOrderNotTrackedIsStillApplied` |
| B12 (if@426) | 불완전 키인데 같은 id 추적 | — | return@427 | 참 `TestPollRejectsIncompleteOpenScopeForATrackedOrder`, 거짓 `TestA191IncompleteOpenOrderNotTrackedIsStillApplied` |
| B13 (range@435) | 추적 키 순회 | — | — | 전부 |
| B14 (if@436) | OPEN 에서 봄 → 건너뜀 | — | continue | `TestA191OrderStillInTheOpenListIsNeverLookedUp`(변이 m3) |
| B15 (if@441) | 단건 오류 | — | return@442 | count=0(기존 부채) |
| B16 (if@446) | 단건 파싱 실패 | — | return@447 | count=0(기존 부채) |
| B17 (if@452) | 응답 id 빈 값 | 추적 id | — | `TestBlankPayloadIdFallsBackToTheTrackedId` |
| B18 (if@457) | 응답 범위 불완전 | — | return@458 | count=0(기존 부채) |
| B19 (if@461) | 응답 범위 불일치 | — | return@462 | `TestPollRejectsTrackedReadFromAnotherCanonicalScope`, `TestA191ExistingTrackedReadGuardsRunBeforeTheLookup`(변이 m17) |
| call@467 `noteLeftOpen` | 가드 통과 뒤 후보 등록(분기는 새 파일) | `d.leftOpenCycle` | — | m1·m2·m21 |
| call@470 `settleLeftOpen` | 끝에서 조회·교체·정리·보고 | `snaps[i]` 상태 필드, `d.leftOpen`, `cycle.ClosedReads/Unresolved`, 알림 | return@470 | a191 시험 전부 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `execgw.ScanOrders`(@376) | OPEN 끝까지 | 오류 → 주기 실패 | AST |
| `d.Order.OrderRaw`(@440) | 목록 이탈 주문 단건 | 오류 → 주기 실패 | AST |
| `brokerstate.Derive`(@431·@466) | 우선순위 표 | 순수 | AST |
| `d.noteLeftOpen`(@467) | 후보 등록 | 실패 없음 | `../../new-code-ast/filldetect--detector.noteleftopen.ast.json` |
| `d.settleLeftOpen`(@470) | 종결 목록 확인 | **오류를 돌려주지 않음**(주기 실패 안 함), 조회 2s·알림 2s 시한 | `../../new-code-ast/filldetect--detector.settleleftopen.ast.json` |

## State mutations and fallbacks

- 지역 맵·슬라이스, `cycle.OpenOrders`, `cycle.TrackedReads++`, (새) `d.leftOpenCycle`·`d.leftOpen`·`cycle.ClosedReads`·`cycle.Unresolved`.
- fallback: 종결 목록이 확인하지 못하면 단건 스냅숏 그대로.

## 편집

- base 대비 본문 2줄: call@467 와 return@470 의 `d.settleLeftOpen` 감쌈. 기존 오류 반환 무편집. 변이 원장 `../../mutation/`.

## Safety conclusion

- 새 경로가 종결을 쓰는 경우는 design D3 하나, 그때도 상태 필드만. 그 밖은 단건 그대로. 주기 실패 경로 추가 0(변이 m10 이 잡음).
