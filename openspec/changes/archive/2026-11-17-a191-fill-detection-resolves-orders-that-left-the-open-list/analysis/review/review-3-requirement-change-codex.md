# 요구 변경분 리뷰(3회차, 결정 반영분 한정) 원문 — Codex(gpt-6-astra), `codex exec -s read-only`, 2026-11-16

## 프롬프트
```
Adversarial review limited to a REQUIREMENT CHANGE in a high-risk design for a live-money Go trading engine. Read-only: no file edits, no network, no go test, no make.

Change: openspec/changes/a191-fill-detection-resolves-orders-that-left-the-open-list/. The Manager made decisions M1..M10 (review.md section "Manager 결정 (2026-11-16 …)"). These changed: specs/fill-detection/spec.md (now 4 requirements), design.md D1 (trigger excludes FailClosed single reads), D3-9 (CLOSED row must carry a readable execution.filledQuantity), D5 (per-cycle cap of one CLOSED lookup, 2s real-time deadline over the whole scan via context.WithTimeout, interval/report timing on the injected Clock, deferred candidates keep lastLookup and are not pruned), D6 (alert via Detector.Alerts <- engine Notifier, event fill.order_unresolved normal, fields intent_id/market/reason/unresolved_seconds only; TrackedOrder gains IntentID copied in JournalTracked.TrackedOrders), D10/D11/D12.

Review ONLY these changes. Cite file:line for every claim and verify against code (internal/filldetect/*.go, internal/journal/fills.go, internal/obs/notifier.go, internal/app/engine/exit_order_log.go, cmd/tossctl/engine.go, internal/execgw/indoubt.go). Questions:
1. Do spec requirements and design D1/D3/D5/D6 agree with each other and with the M decisions table? List any contradiction.
2. Is the per-cycle-one-lookup selection rule (never-asked first, then oldest lastLookup; deferred keeps lastLookup) free of starvation or livelock? Any candidate that can never be asked?
3. Does wrapping the whole ScanOrders in a 2s context actually bound wall time with the official client (internal/official) and accountbudget (internal/accountbudget)? Anything that ignores ctx?
4. Does the alert plan leak forbidden data (broker order id, account, symbol, qty, price) through obs.Notifier logging or the derived dedup Key? Check how Notifier logs fields and Body.
5. Is a 2s-bounded Notify inside collect safe (normal severity path), or can it still block the poll beyond the bound?
6. Any decision whose framing in the documents is unsafe.
Output: numbered findings (BLOCKER/MAJOR/MINOR/NIT), claim, file:line, proposed change; then a one-paragraph verdict. Concise.
```

## 응답
문서 좌표는 `openspec/changes/a191-fill-detection-resolves-orders-that-left-the-open-list/` 기준입니다.

1. **BLOCKER — `WithTimeout(2s)`는 조회 반환·wire 해제의 2초 상한을 보장하지 않는다.**  
   D5·D11은 전체 조회와 다른 주문의 추가 반영 지연을 2초로 제한한다고 명시한다(`design.md:116–117,180–183`). 그러나 공식 클라이언트의 토큰 취득은 취소 불가능한 `mu.Lock()` 뒤에서 교환까지 수행한다(`internal/official/client.go:335–336`, `internal/official/token.go:62–80,110–127`). 다른 루프가 잠금을 보유하면 CLOSED 조회의 deadline이 지나도 기다린다. accountbudget도 응답 실패 후 `WithoutCancel`로 **추가 2초** 정리를 수행하며, 이후 lease를 해제한다(`internal/accountbudget/transport_linux.go:395–422`). 예산 파일의 `Sync()` 역시 ctx를 받지 않는다(`internal/accountbudget/budget_linux.go:316–352`).  
   **변경:** “2초 취소 요청”과 “반환·wire 해제 상한”을 구분하고 M8 안전 판정을 재검토한다. 토큰 경합·관측 정리·파일 I/O까지 포함한 계약 없이는 D11의 상한을 확정하지 않는다.

2. **MAJOR — collect 안의 Notify도 2초로 묶이지 않는다.**  
   D6은 context만으로 알림을 제한한다(`design.md:136–137`). 실제 Notify는 먼저 동기 로그를 쓰며, Logger는 전달된 ctx 대신 `context.Background()`로 동기 출력한다(`internal/obs/notifier.go:149–176`, `internal/obs/log.go:100–126,154–155,194–203`). 출력이 막히면 deadline으로 중단되지 않는다. ntfy HTTP 요청은 ctx를 따르지만 이 로그 경로는 별개다(`internal/obs/ntfy.go:102–105,128–135`). 모든 원장 적용은 collect 반환 뒤이므로 함께 지연된다(`internal/filldetect/detect.go:293–317`).  
   **변경:** 알림 실행을 collect의 임계 경로에서 분리하고, 제한된 큐·실패 처리 정책을 명시한다. 조회와 알림이 동일 deadline을 공유하는지도 명확히 한다. 별도 2초씩이면 D11의 총 2초 주장과 충돌한다.

3. **MAJOR — 선택 규칙은 고정 집합에서만 공정하며, 지속 유입에서는 starvation이 가능하다.**  
   “미조회 우선, 동률은 추적 순서”이므로 매 주기 새 후보가 유입되면 이미 조회한 미해결 주문은 영원히 재조회되지 않을 수 있다(`design.md:112–120`). 미조회 후보도 자신보다 앞서는 새 후보가 계속 들어오면 최초 조회조차 못 받을 수 있다. 실제 추적 순서는 도착 순서가 아니라 주문번호 정렬이다(`internal/journal/fills.go:1652`, `internal/filldetect/ledger.go:136–146`, `internal/filldetect/detect.go:382–392`). deferred의 시간 보존은 이 우선순위 문제를 해결하지 않는다.  
   **변경:** 최초·재조회 모두 대기시간에 따라 우선권을 얻는 규칙 또는 두 집합의 처리 할당량을 Manager 결정으로 추가한다. 현 규칙을 유지하면 “N주기 내 조회”를 **고정 후보 집합·성공 주기·진행하는 Clock** 조건으로 한정한다.

4. **MAJOR — “운영자에게 한 번 알려야 한다”와 normal best-effort 전송은 다른 계약이다.**  
   R4는 운영자 통지를 보장하지만(`specs/fill-detection/spec.md:60–70`), normal Notify는 Publisher가 없거나 전송에 실패해도 nil을 반환한다(`internal/obs/notifier.go:153–155,180–192`). D6의 `reported`만으로는 전송 성공을 알 수 없다(`design.md:108,124–137`). 특히 조회 deadline을 그대로 재사용하면 조회 timeout 직후 알림도 이미 만료된 ctx를 받을 수 있다.  
   **변경:** M1의 normal 결정을 유지한다면 R4를 “이벤트 기록 및 전송 1회 시도, 전달 보장 없음”으로 정정한다. 실제 전달 보장이 필요하면 별도 전달 상태·재시도 결정을 받아야 한다.

5. **MINOR — R1·R4의 범위가 D1·D5보다 넓다.**  
   R1의 “한 주기 조회는 한 건”은 후보가 없거나 모두 간격 미달인 0건 주기를 표현하지 못한다. 후보 셋이면 세 주기에 조회된다는 시나리오도 기존 `lastLookup` 조건이 없다(`specs/fill-detection/spec.md:7–16`, `design.md:112–115`). R4는 모든 미해결 목록 이탈 주문을 대상으로 쓰였지만 D1은 FailClosed 주문을 후보에서 제외하고 D5는 비후보 상태를 삭제한다(`specs/fill-detection/spec.md:60–70`, `design.md:45–48,118–125`; 결정 `review.md:103,109`).  
   **변경:** “주기당 **최대** 1건”, 시나리오의 최초 조회·고정 집합 전제, R4의 D1 후보 한정을 명시한다.

6. **MINOR — D10의 요청 ‘상한’은 실제 wire 요청 상한이 아니다.**  
   D10은 3페이지를 최대 3요청으로 계산한다(`design.md:169`). 공식 클라이언트는 401에 최대 두 번 refresh·재전송하며 토큰 교환도 HTTP 요청이다(`internal/official/client.go:348–375`, `internal/official/token.go:132–149`). ScanOrders의 페이지 상한은 이 내부 요청을 세지 않는다(`internal/execgw/indoubt.go:743–750`).  
   **변경:** 수치를 “인증 재시도 없는 정상 경로의 목록 요청량”으로 표기하고, 인증 교환·재전송 예산을 별도로 연결한다.

7. **MINOR — M7 잔여 위험 설명이 피해 조건을 지나치게 좁힌다.**  
   D12는 실제 피해에 “절단 + 뒤 페이지의 상충 중복”이 필요하다고 한다(`design.md:197–198`). 하지만 빈 커서로 정상 종료하면 뒤 페이지의 **유일한 정상 후보**도 보지 못한다(`internal/execgw/indoubt.go:748–756`). 그러면 not-found로 추적을 유지하며, 같은 응답이 반복되면 복구가 계속 실패한다(`design.md:70,89–94`).  
   **변경:** 잘못된 종결뿐 아니라 정상 종결 누락·지속 미해결도 M7의 수용 위험에 포함한다.

**판정: 요구사항 변경 그대로는 착수 승인에 부족하다.** M2의 FailClosed 제외와 M6의 읽을 수 있는 체결 수량 요구는 결정·명세·설계가 일치한다(`review.md:103,107`, `specs/fill-detection/spec.md:5–6,24–27`, `design.md:45–48,81–85`). 지정한 네 필드와 고정 본문만 보내면 Notifier가 금지 필드를 자동 추가하는 경로는 확인되지 않았다. normal 경로는 dedup Key도 사용하지 않는다(`internal/obs/notifier.go:149–176,210–216,991–1013`); `IntentID` 복사 원천도 존재한다(`internal/journal/fills.go:1663–1664`). 우선 해결할 문제는 시간 상한, 공정성, 통지 보장이다. 현재 소스 수동 정적 검토이며 파일 변경·네트워크·시험·make 실행은 없었다. 신규 AST 생성·동적 검증은 미실행. `impeccable: N/A — backend/documentation-only`.