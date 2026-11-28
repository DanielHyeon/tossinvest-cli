**BLOCK**

현재 HEAD `490e94f9daceef7fe3a956030ec6e5054bfe6fa8` 기준 정적 리뷰입니다. 인용 AST 40개와 편집 대상 AST 2개의 소스 해시는 현재 파일과 모두 일치했습니다. 해시 일치는 시험 실행 증거가 아닙니다.

아래 `design.md`·`proposal.md`·`tasks.md`·`delta`는 지정한 a193 change 내부 파일입니다.

**[P1] 직렬 해소가 5초 스윕과 10분 알림을 함께 막는다 — 설계 결함**

- **근거:** [design.md:104](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/openspec/changes/a193-an-in-doubt-order-is-resolved-while-the-engine-runs/design.md:104), 같은 파일 112–120·178–195행, delta:92·95.
- **왜 문제인가:** `PendingAttempts → 각 attempt의 noteStale → 동기 Resolve → 전체 완료 후 Sleep` 구조다. 앞선 세 건이 각각 약 5분을 쓰면 네 번째 attempt의 알림 검사는 약 15분 뒤다. 스윕 도중 새로 생긴 attempt는 그보다 늦게 발견된다. 따라서 “5초마다 읽음”과 “10분 상한이 대기를 드러냄”은 성립하지 않는다.
  
  별도로 `PendingAttempts`가 계속 실패하면 로그 한 번만 남기고 실행자는 살아 있다. `noteStale`도 `OnStop`도 호출되지 않아 critical 침묵이 지속된다. tasks 3.7의 **즉시 오류를 반환하는 해소기**는 두 경로를 검증하지 않는다.
- **제안(최소):** 주기·알림 지연의 실제 보장과 허용 오차를 먼저 규정하고, 긴 해소·여러 attempt·반복 조회 실패를 수용 시험에 포함한다.
- **제안(범위 확장):** 10분 알림을 보장하려면 알림 스윕을 직렬 해소 대기와 분리한다. 반복 원장 조회 실패의 critical 승격은 별도 정책 결정으로 명시한다.

**[P1] 기존 Resolver는 전체 호출의 5분 상한을 강제하지 않는다 — 기전 오류**

- **근거:** design.md:131–134, delta:92, [indoubt.go:286](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/internal/execgw/indoubt.go:286), 같은 파일 336–364·414–415·743–748행.
- **왜 문제인가:** `deadline`은 다음 Sleep 여부에서만 검사한다. 목록 순회·페이지별 조회·부재 교차 확인에는 그 deadline을 전달하지 않는다. OPEN/CLOSED 각각 최대 50페이지이므로 **첫 스캔 하나만으로도** 5분을 넘을 수 있다. 개별 HTTP timeout 15초는 전체 해소 timeout이 아니다.
  
  설계의 “5분 + 마지막 스캔”과 delta의 “최대 5분을 넘어 계속되어서는 안 된다”도 서로 다르다. 정상 1페이지 요청 수를 세는 tasks 3.9로 전체 시간 상한은 증명되지 않는다.
- **제안(최소):** 현재 구현이 제공하는 상한으로 문구를 바로잡고, 다중 페이지·느린 응답을 포함한 시간 계약을 정한다.
- **제안(범위 확장):** 강제 deadline을 추가한다면 **해소 예산 만료와 엔진 종료 취소의 처분을 구별**해야 한다. 단순 `WithTimeout` 추가는 D3에 의해 park 대신 IN_DOUBT 잔류를 만들 수 있다.

**[P1] “park 원인 알림이 이미 있다”는 전체 PLACE에 적용되지 않는다 — 기전 오류**

- **근거:** design.md:178–192·201–208, delta:95, [exit_held_proposal.go:126](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/internal/app/engine/exit_held_proposal.go:126), 같은 파일 135–149·217–237행.
- **왜 문제인가:** 기존 `recordParkCause`는 **무장된 exit 발의의 intent에 연결된 attempt**만 알린다. 새 실행자의 대상은 전략 진입·수동 flatten 매도를 포함한 모든 IN_DOUBT PLACE다.
  
  그런 attempt가 10분 전에 park되면 stale 알림은 없고, 이후 `PendingAttempts`에서도 빠지며, exit 발의가 없으므로 기존 park 알림도 없다. 계정 전역 진입 래치만 걸린 채 critical이 없는 경로다. tasks 3.2는 발의가 있는 손절 사례만 검증한다.
- **제안(최소):** “기존 알림 보장”을 발의가 있는 경우로 한정하고, 나머지 PLACE의 park 통지 책임을 명시한다.
- **제안(범위 확장):** 기존 exit 알림으로 덮이지 않는 park의 통지 경로와 중복 제거 key를 추가한다. Q-A의 “IN_DOUBT 최초 발견 알림”만으로 park 결과 통지까지 충족한다고 간주해서는 안 된다.

**[P1] Critical 우선 선택은 이미 실행 중인 해소 조회를 선점하지 못한다 — 안전성 주장 오류**

- **근거:** design.md:141–146, [budget_linux.go:549](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/internal/accountbudget/budget_linux.go:549), 같은 파일 555–584행, [transport_linux.go:459](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/internal/accountbudget/transport_linux.go:459).
- **왜 문제인가:** Normal 조회가 먼저 `wire.lock`을 얻으면, 이후 도착한 Critical 손절도 그 lease가 풀릴 때까지 기다린다. lease는 응답 body 종료까지 유지된다. 우선순위 비교는 **대기 티켓 사이의 순서**만 정한다.
  
  따라서 Auxiliary goroutine 분리만으로 “다른 포지션 손절 제출을 기다리게 하지 않는다”는 결론은 나오지 않는다. 새 조회가 손절의 wire 진입을 늦추는 경로가 존재한다. tasks 3.6의 가짜 해소기 비대기 시험은 실제 a166 transport 경합을 덮지 않는다.
- **제안(최소):** “앞지르지 못한다”를 “함께 대기 중이면 Critical 우선”으로 정정하고, 이미 조회가 lease를 보유한 상태의 손절 지연을 측정·판정한다.
- **제안(범위 확장):** 조회 점유시간 제한이나 선점 정책은 별도 결정 사항이다. Auxiliary 우선순위로 바꾸는 것만으로 이미 보유한 lease 문제는 해결되지 않는다.

**[P1] D6의 프로세스 간 종목 배제 증명이 성립하지 않는다 — 설계 증거 공백**

- **근거:** design.md:224·227–230, [gateway.go:522](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/internal/execgw/gateway.go:522), 같은 파일 529·548·770–780행, [durability.go:383](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/internal/journal/durability.go:383), `cmd/tossctl/flatten.go:237–264`.
- **왜 문제인가:** `claimSymbol`은 Gateway 인스턴스 내부 잠금이고, flatten은 별도 Gateway를 만든다. `checkSymbolFree` 조회와 `Prepare` 트랜잭션은 분리돼 있으며, `Prepare`는 종목별 미종결 여부를 다시 검사하지 않는다.
  
  가능한 순서는 **엔진 제출이 빈 원장을 확인 → flatten 제출이 IN_DOUBT 기록 → 실행자가 해소 시작 → 앞서 검사한 엔진 제출이 기록·전송**이다. “이미 IN_DOUBT가 보이는 신규 호출은 거절된다”는 사실로, 앞서 검사를 통과한 다른 프로세스의 제출까지 배제할 수 없다. 같은 attempt의 전이 가드는 서로 다른 attempt를 막지 않는다.
- **제안(최소):** D6의 절대적 비중첩 주장을 철회하고 이 순서를 시험·영향 분석에 포함한다. 기존 제출 경로의 공백이며, 이번 리뷰에서 실제 중복 주문 발생을 입증한 것은 아니다.
- **제안(범위 확장):** 프로세스 간 원자적 종목 배제가 필요하다는 결론이면 별도 승인 범위로 다룬다.

**[P2] D3의 좁은 주장은 맞지만 취소 전체 계약까지 증명하지 못한다 — 경계 누락**

- **근거:** design.md:58–71·161–174, [durability.go:668](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/internal/journal/durability.go:668), `indoubt.go:376–380`, design.md:115–116, `internal/obs/record_only.go:137–149`.
- **왜 문제인가:** PLACE의 park 호출 **네 곳**과 `ResolveConfirmed`·`ResolveFailed`는 모두 같은 ctx를 전달한다. 이미 취소된 ctx로 전이에 들어가 실패하면 park 래치에 도달하지 않는다는 설명은 맞다.
  
  그러나 “커밋 전 ctx 검사”는 취소와 커밋을 원자적으로 순서화한다는 증거가 아니다. **커밋 진입 검사 이후의 취소**는 D3의 “커밋 뒤 취소” 설명과 tasks 3.4에서 빠졌다. 저장소 밖 표준 라이브러리·드라이버 원문은 이번 제한 때문에 재검증하지 않았다.
  
  또한 새 실행자는 `ctx.Err()` 검사 **전에** `noteStale`을 호출한다. 오래된 attempt에서 그 사이 취소되면 `RecordCritical` 실패가 `ReasonAlertUndelivered` 래치를 걸 수 있다. Resolver의 park 방어만으로 “취소 시 게이트 변화 0”을 보장하지 못한다.
- **제안(최소):** 취소 계약을 전이 성공 여부와 경합 시 허용 결과로 구체화하고, 오래된 attempt의 알림 기록 중 취소도 시험한다. 현재 증거만으로 Resolver 변경이 반드시 필요하다고 단정하지는 않는다.
- **제안(범위 확장):** 알림 경로에 새 취소 가드나 취소 오류 예외 정책을 넣는다면 명시적으로 승인 범위에 포함한다.

**[P2] 정본의 실측 도입 조건이 delta에서 사라지고 일부 규범이 서로 다르다 — 규범 누락**

- **근거:** `openspec/specs/order-execution/spec.md:118`, delta:6·15·92–95, proposal.md:35–37, tasks.md:19–20·45–50.
- **왜 문제인가:** 정본의 “배선·실측·공시 셋 없이 도입 금지” 문단을 교체하면서 **실측 의무가 새 규범에서 빠졌다**. tasks에는 미래 측정이 있으나 proposal은 이미 “셋을 가진다”고 쓴다. 현재는 측정 계획이다.
  
  10분 조건도 delta는 “넘게”, design과 tasks는 “이상/정확히 10분에”다. 재생은 15행에서 세션 예외를 명시했지만, 6행의 일반적 “해소 순서(SHALL)”는 그대로여서 적용 범위를 더 명확히 할 필요가 있다.
- **제안(최소):** 실측을 도입 전 조건으로 보존하고 계획과 결과를 구분한다. `> / >=`, 시간 허용 오차, 기동 재생 순서와 세션 조회 순서를 일관되게 명시한다.

**[P2] AST 종류와 park 호출 수 인용이 틀렸다 — 사실 오류**

- **근거:** design.md:25, proposal.md:32, tasks.md:40, `analysis/cited-ast/summary.txt:186`, 해당 `ast.json`.
- **왜 문제인가:** `exitIntentAttemptsQ`의 `B11(default@216)`은 AST에서 **`B11(case@216)`**이다. Go 소스가 default 절인 것은 맞지만, 문서는 `kind`를 AST 그대로 인용한다고 선언했다. tasks 3.4의 “park로 가는 다섯 자리”도 PLACE `Resolve` 기준으로는 **299·325·348·351행 네 곳**이다.
- **제안(최소):** 인용을 정정하고, 네 호출 지점과 취소 발생 시나리오 수를 구분한다.

확인했는데 문제없었던 항목:

- 기동 복구 성공 뒤 Auxiliary가 시작되고, 정상 종료 시 같은 WaitGroup으로 drain된다.
- `runAuxiliaryBody`의 Run 패닉은 보조 실행자 정지로 변환된다.
- flatten의 직접 Resolver 호출은 취소 step이며, 새 PLACE 해소와 대상 종류가 겹치지 않는다.
- 운영자 해소는 트랜잭션 안에서 `UNRESOLVED_IN_DOUBT` 출발 상태를 요구한다.
- CONFIRMED는 주문번호와 함께 기록되고 체결 추적 후보가 되며, 비수용 해제 경로로 발의를 풀지 않는다.
- park 래치는 종목 없는 계정 전역 신규 진입 차단이며, 운영자 원장 해소가 메모리 래치를 풀지 않는다는 공시는 코드와 맞는다.
- stale key의 attempt 식별자·`RecordCritical` 입구·재알림 창 0은 재시작 간 중복 억제 방식과 맞는다.
- 세션 경로의 재생 미호출과 RECORDED·DISPATCH_STARTED·ACKED 제외 설계는 재시작 규칙 오용을 피한다.
- 조회 해소 경로 자체에 주문 생성 mutator는 없다.
- 파일 쓰기·Go 시험·네트워크·브로커·운영 저널·docker·tossctl 실행 0. 외부 스킬·기억 조회는 경로 제한으로 미수행. `impeccable: N/A — backend/documentation-only`, UI 변경 없음.