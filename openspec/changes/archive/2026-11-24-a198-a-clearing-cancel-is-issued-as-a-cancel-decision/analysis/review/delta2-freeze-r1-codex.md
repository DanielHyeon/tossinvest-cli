# 보강 2 freeze 리뷰 r1 — codex (다른 계열, 1회차·마지막), 2026-11-24

- 명령: `codex exec -s read-only -C ~/tossos-work/a198-impl -o <out> "<prompt>" </dev/null`(rc 0), codex-cli 0.165.0. 프롬프트 원문: 같은 디렉터리 `delta2-freeze-r1-prompt.txt`. 기준 HEAD `f8a5bb898` + 미커밋 보강 2 문서.

## 원문(마지막 메시지)

**판정: 보강 2는 현재 설계대로 확정하면 안 됩니다. P1 3건, P2 4건. P0에 해당하는 실제 초과 매도는 입증하지 않았습니다.** 기준은 HEAD `f8a5bb89861362ddccf21857dc2ce99650c53c48`와 현재 미커밋 문서입니다.

1. **P1 — 비종결·fail-closed 증거를 제외한 뒤 “모든 증거가 일치한다”고 판단합니다.**  
   [design.md:17](…/design.md:17)의 D10은 `terminal=1`인 행만 선택한 뒤 `fail_closed`와 판정 일치를 검사합니다. 그런데 빈 범위 legacy 종결 행과 같은 소유자의 scoped 비종결·fail-closed 행은 공존할 수 있습니다. scoped 쓰기는 다른 범위인 legacy 행을 덮지 않고, CTE도 두 행을 남깁니다([fills.go:775](…/internal/journal/fills.go:775), [fills.go:888](…/internal/journal/fills.go:888)).  
   **정적 반례:** legacy `CANCELLED/terminal=1/fail_closed=0` + scoped `UNKNOWN/terminal=0/fail_closed=1`이면 후자는 검사에서 탈락하고 취소가 CONFIRMED로 닫힙니다. 기존 치움 술어에도 있는 약점이지만, D10은 이를 미종결 방벽 제거의 근거로 새로 사용합니다. 관련 행 전체의 충돌을 먼저 검사해야 합니다. 시험 21·22에는 이 혼합 반례와 변이가 없습니다.

2. **P1 — 로컬 소유자 유일성은 현재 취소 대상의 신원을 증명하지 않습니다.**  
   [증거 문서:59](…/analysis/d9-d10-evidence.md:59)는 번호와 계좌·시장·종목·방향이 맞는 과거 PLACE/AMEND 하나를 소유자로 선택합니다. 취소에는 대상 소유 intent의 식별자가 없고, 취소 intent의 거래일은 실행일입니다([gateway.go:1020](…/internal/execgw/gateway.go:1020)).  
   **번호 재사용을 전제로 한 반례:** 어제 엔진 주문 `X`는 CANCELLED, 오늘 같은 범위의 외부 주문이 `X`를 재사용하고 flatten 취소가 IN_DOUBT가 됩니다. 오늘 외부 주문에는 로컬 PLACE가 없으므로 후보는 여전히 어제의 하나입니다. D10이 어제 종결로 CONFIRMED를 쓰면 flatten은 다음 해소에서 현재 주문의 취소를 Done으로 읽습니다([flatten.go:433](…/internal/flatten/flatten.go:433)).  
   이는 “외부 주문이면 후보 0”이라는 증거 문서의 주장에 반합니다. 실제 재사용·초과 매도는 미측정입니다. 대상 소유 신원을 결속하거나, 결속을 증명할 수 없는 취소는 닫지 않아야 합니다. 전일 주문 지원 때문에 취소일 비교 하나로 해결할 수는 없습니다.

3. **P1 — “종결 뒤 최대 5초”는 성립하지 않습니다.**  
   [design.md:106](…/design.md:106)은 이를 “이전과 같은 경우”에 넣지만, 실행자는 **스윕 전체 종료 후** 5초를 쉽니다. PLACE 해소는 직렬이며 기본 예산은 건당 5분입니다([in_doubt_settle.go:148](…/internal/app/engine/in_doubt_settle.go:148), [in_doubt_settle.go:186](…/internal/app/engine/in_doubt_settle.go:186), [indoubt.go:155](…/internal/execgw/indoubt.go:155)).  
   CANCEL 처리 구간 **직후** 종결 기록이 도착하면 남은 PLACE 해소 전체를 기다립니다. CANCEL을 2단 앞에 두어도 이 창은 남습니다. base에서는 같은 조건의 종결 기록 뒤 관측 주기에 풀렸으므로 추가 회귀입니다. 별도 처리 기회 또는 명시적인 지연 계약이 필요하며, 시험은 앞선 스윕의 PLACE 해소가 시작된 뒤 종결 기록을 넣어야 합니다.

4. **P2 — ADDED의 “그 스윕은 브로커 요청 0”이 보존한 PLACE 계약과 충돌합니다.**  
   [delta:355](…/specs/order-execution/spec.md:355)는 CANCEL을 닫은 스윕 전체의 요청을 금지합니다. 같은 스윕에 IN_DOUBT PLACE가 있으면 보존된 조회 해소 의무와 동시에 만족할 수 없습니다. “취소 종결 경로가 브로커를 호출하지 않는다”로 범위를 좁히고, CANCEL+PLACE 혼합 스윕을 시험해야 합니다.

5. **P2 — §6 잔여 목록에 세션 중 `DISPATCH_STARTED` 고착이 빠졌습니다.**  
   [dispatch.go:181](…/internal/journal/dispatch.go:181)의 `MarkAcked` 또는 [dispatch.go:206](…/internal/journal/dispatch.go:206)의 `MarkInDoubt` 쓰기가 실패하면 프로세스가 살아 있어도 전송 시작 상태로 남습니다. 이후 대상의 종결 기록이 들어와도 D9·D10 모두 처리하지 않습니다.  
   [design.md:99](…/design.md:99)의 1~4 외에 별도로 남는 회귀입니다. ACKED 최종 쓰기 실패와 구분해 기록하고 시험해야 합니다. 진행 중 전송을 세션에서 임의 종결하라는 뜻은 아닙니다.

6. **P2 — D10이 gateway 차단을 풀어도 이미 보류된 flatten은 자동으로 풀리지 않습니다.**  
   [증거 문서:106](…/analysis/d9-d10-evidence.md:106)의 “flatten의 그 종목도 풀린다”는 조건이 부족합니다. `SymbolInFlight`를 받은 단계는 HELD이고, 청산 계획은 attempt가 아니라 `flatten_steps`의 HELD·PENDING·IN_DOUBT를 읽습니다([flatten.go:467](…/internal/flatten/flatten.go:467), [journal/flatten.go:324](…/internal/journal/flatten.go:324)).  
   D10 뒤에도 현재 saga의 보류가 남을 수 있습니다. resolver와 경합에서 진 경우도 IN_DOUBT 단계가 남습니다. 문서에 재실행·단계 재평가 조건을 명시하고, `CancelAll → D10 → PlanLiquidation` 순서의 시험을 추가해야 합니다.

7. **P2 — 시험 27은 “첫 관찰”과 `Resolver.Resolve`의 최종 결과를 혼동합니다.**  
   [design.md:84](…/design.md:84)는 비종결 상태의 첫 관찰을 비교하지만, `Resolve`는 첫 관찰에서 반환하는 API가 아닙니다. OPEN은 반복 관찰 후 FAILED_CONFIRMED, pending은 대기 후 park가 됩니다([amend_indoubt.go:116](…/internal/execgw/amend_indoubt.go:116)).  
   지금 문구만으로 최종 반환값을 비교하면 잘못된 동치 시험이 됩니다. 종결 5개 상태는 실제 최종 결과로 비교하고, 비종결은 첫 대기를 제어하면서 원장에 확정 전이가 없음을 관찰해야 합니다. `REPLACED/UNKNOWN`의 park와 D10의 IN_DOUBT 유지는 명시적인 차이로 남겨야 합니다.

D9 자체의 번호 덮어쓰기 결함은 발견하지 못했습니다. 기록 번호를 그대로 보존하며, a094가 철회한 PLACE 목록 matcher를 호출하지 않습니다. 다만 **접수 확정은 취소 완료 증명이 아닙니다.** `CANCEL_REJECTED` 가능성은 남고, 매수는 접수로 치움·매도는 종결 증거 대기라는 기존 계약을 그대로 따릅니다.

D10의 한 트랜잭션 설계는 기존 `BEGIN IMMEDIATE` 설정과 상태 조건 쓰기를 지키면 DB 내부 판독·쓰기 경합을 막습니다. a165 꼬리표 및 예약 해제에서 별도 결함은 확인하지 못했습니다. 다만 시험 25에는 실제 별도 원장 연결과 체결 기록 writer의 경합, 전이 기록·예약 해제 실패 시 전체 rollback 관측이 필요합니다.

읽기 전용 문서 비교에서 MODIFIED·`compare.diff`는 선언한 치환과 일치했습니다. 기존 AST 34개의 소스 해시도 일치했습니다. 지정 함수와 관련 SQL·호출부·시험 본문을 정적으로 확인했으며, 파일 수정·Git 상태 변경·빌드·시험·실계좌 호출은 하지 않았습니다. ai-memory 조회는 접근 제한으로 실패했고 그래프 재생성은 하지 않았습니다. `impeccable: N/A — 백엔드·계약 문서 리뷰, UI 변경 없음`.

(링크의 `…` 는 `/home/daniel/tossos-work/a198-impl/openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision` 또는 저장소 루트의 줄임 — 원문은 절대경로.)
