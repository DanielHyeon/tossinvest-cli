판정: **BLOCK**

23판은 주요 보수 조항을 반영했지만, **모드 통지 SHALL의 범위와 K3 구조 핀을 구현 전에 고쳐야 합니다.** flatten CLI의 제시된 경쟁 경로는 현재 코드에서 도달하지 않습니다.

아래 `C`는 `openspec/changes/a092-an-alert-does-not-hold-the-stop`입니다.

| # | 등급 | 주장 | 증거(파일:줄) | 권고 |
|---|---|---|---|---|
| 1 | **P0(T)** | 새 “상태를 바꾼 전이마다 통지” SHALL은 유지하기로 한 전달 실패 승격 경로와 충돌한다. 원장 기록에 성공하고 **전송만 실패**한 경우에도 승격은 `announcer=nil`이다. 23판의 예외는 **durable 기록 실패**만 덮는다. | `C/specs/engine-safety/spec.md:324`, `:77`; `internal/obs/notifier.go:223-228`, `:382-383`; `internal/app/engine/alertdelivery.go:451-452`; a124 archive `design.md:139-142` | 무통지 예외를 전달 실패에 따른 승격까지 명시하고 ADDED 요구에도 연결하거나, 해당 승격의 기록 전용 통지를 설계한다. 현재 설계와 포괄 SHALL을 함께 archive하면 안 된다. |
| 2 | **P1(T)** | K3 핀의 “Commit과 Project 사이 반환 없음”은 정상적인 **커밋 실패 반환**까지 금지한다. 문자 그대로 구현하면 현재의 안전한 오류 처리가 실패하고, 이를 제거하면 미커밋 상태를 투영할 위험이 생긴다. | `C/design.md:749`; `C/tasks.md:25`; `internal/journal/operating_mode.go:468-476` | **커밋 성공 경로**에서 반환·비동기 분리 없이 투영하도록 정의한다. 커밋 실패 시 투영 0회·통지 0회를 별도 양성 조건으로 둔다. |
| 3 | **P2(T)** | flatten을 현재의 프로세스 밖 outbox 기록자로 취급할 근거가 없다. 또한 “다음 승인은 미전달 0을 만족하지 않는다”는 설명은 전체 승인 동작과 맞지 않는다. 다음 승인은 새 행도 승인하여 0을 만들 수 있다. | `C/design.md:783-787`; `cmd/tossctl/flatten.go:233-263`; `internal/flatten/flatten.go:423-428`; `internal/execgw/replay.go:249-259`; `internal/obs/notifier.go:854-875` | 현재 도달 불가로 확정하고, 향후 replay 배선 변경 시 재검토할 조건으로 기록한다. 다음 승인·배달이 반드시 다시 잠근다는 설명은 삭제한다. |

1. **K1 — 신원 변경은 통과, 통지 적용 범위는 #1로 BLOCK.**  
   B15 `direction == 0`은 `:415`, B16 자동 완화 억제는 `:421`에서 반환한다. 투영·통지는 `:475-479` 이후이므로 “변화 없으면 통지 전 반환”은 참이다. AST의 B15·B16 좌표도 일치한다. 전이 rowid를 키에 포함하면 재강화는 이전 정착 행과 다른 키가 된다. 정본 중복 억제는 **같은 event key**에 대한 규칙이므로 모순되지 않는다. 다만 “통지 수를 늘리지 않는다”는 같은 상태의 반복 관측에 한정해야 한다. 재강화 통지 수는 의도적으로 늘어난다.  
   근거: `internal/journal/operating_mode.go:410-421`, `:475-479`; `internal/obs/notifier.go:285-293`; `C/design.md:733-739`; `openspec/specs/engine-safety/spec.md:710`.

2. **K2·K4 — 설계 규범 통과.**  
   확정~세대 읽기 사이 해제를 앞선 해제로 취급하는 허용 창, 승격 실패 뒤 무조건 차단이 a124 정본과 같다. 세대 비교와 삽입도 게이트 잠금 하나 안에서 수행된다. 제시한 설계 자체에서 추가로 게이트를 여는 경로는 찾지 못했다. 다만 세 래치 자리의 실제 적용은 미구현이므로 통과한 코드로 보고할 수 없다.  
   근거: `C/specs/engine-safety/spec.md:69`; `openspec/specs/engine-safety/spec.md:1460-1473`; `internal/execgw/retry.go:571-582`; `C/tasks.md:24-26`.

3. **K3 — 해석은 정합, 핀은 BLOCK.**  
   “같은 전이 호출 안에서 커밋 뒤 투영”이라는 명시적 해석은 현재 호출 순서와 맞는다. 커밋~투영 사이에는 이전 모드가 보일 수 있다는 한계도 적었다. 그러나 핀은 성공 경로와 실패 경로를 구별하지 않는다. #2 수정이 필요하다.  
   근거: `openspec/specs/risk-management/spec.md:133`; `C/specs/engine-safety/spec.md:310`; `internal/journal/operating_mode.go:468-479`.

4. **K5(i) — 통과.**  
   비시험 `execgw.New(` 호출은 실제로 둘이며, 모두 같은 함수에서 만든 `NewEntryGate` 결과를 `Entry`로 전달한다. tasks는 전수·역할 확인과 “호출자 추가”, “Entry 삭제” 변이를 모두 요구한다. 단순 존재 확인으로 적히지 않았다.  
   근거: `internal/app/engine/gateway.go:249`, `:296-302`; `cmd/tossctl/flatten.go:232-239`; `C/tasks.md:27`.

5. **K6·K7 — 설계 반영 확인, 교차 change 완료는 미확인.**  
   `n.mu` 아래 임차 없는 기록, 기록자별 재알림 창과 0 허용이 명시됐다. 현재 외부 `EnqueueAlert` 호출자는 `parkAlert`와 a066 `notifyRelaxation` 두 곳이다. 후자는 선행 차단 없이 critical 행을 기록하므로 실제로 이관이 필요하며, 23판도 이를 위반 형태와 미완료 작업으로 적었다. 이를 이미 닫힌 항목으로 처리하면 안 된다. exit 재무장 SHALL과 다른 기록자의 0 허용은 제시된 a094·a090 enqueue-only 계획과 양립한다.  
   근거: `C/specs/engine-safety/spec.md:41`, `:65-67`; `internal/execgw/replay.go:535-551`; `internal/app/engine/risk_relaxation_command.go:151-166`; `internal/journal/outbox.go:382-385`; `C/tasks.md:28-29`, `:37`.

6. **flatten 프로세스 밖 기록자 — 현재 P0/P1 아님; 도달 불가.**  
   추적 결과는 다음과 같다.

   - CLI는 자체 Gateway·Gate와 `Resolver`를 만든다. Gateway 옵션에는 `Replay`·`Attested`가 없다.  
     근거: `cmd/tossctl/flatten.go:232-263`.
   - 취소 해소는 `Saga.resolveStep → Resolver.Resolve`다. `Resolver.park`는 attempt와 자기 게이트를 갱신하지만 outbox에 쓰지 않는다.  
     근거: `internal/flatten/flatten.go:423-428`; `internal/execgw/indoubt.go:372-383`.
   - 청산의 안정화는 Collector·Stabiliser이며 Recovery/replayer 배선이 아니다.  
     근거: `internal/flatten/liquidate.go:577-602`.
   - `ReplayInDoubt`의 실제 호출 입구는 `Recovery.replay`이고, 설령 CLI Gateway에서 직접 호출하더라도 replay transport 부재로 `parkAlert` 전에 반환한다.  
     근거: `internal/reconcile/recovery.go:347-351`; `internal/execgw/replay.go:249-259`.

   따라서 현재 flatten 경로가 엔진의 `Acknowledge` 셈~해제 창에 해당 outbox 행을 넣는 반례는 성립하지 않는다. 향후 이 경로가 연결된다면 다른 프로세스의 Gate는 엔진을 보호하지 못하므로, 현재 선행 차단 SHALL의 보호 근거로 사용할 수 없다.  
   근거: `internal/obs/notifier.go:870-875`; `C/specs/engine-safety/spec.md:67`, `:115-117`.

7. **archive 정합성 — #1 때문에 BLOCK.**  
   “배달 실행자의 정지가…” 요구는 정본과 델타를 대조했으며, 인용 블록을 제외한 규범 본문·Scenario 차이는 없었다. 정정 표지는 승격 금지·직접 래치·종료 대기 규범을 바꾸지 않았다. 그러나 새 포괄 통지 SHALL은 a124가 유지하는 무통지 승격과 정합하지 않는다.  
   근거: `C/specs/engine-safety/spec.md:159-298`, `:324`; a124 archive `design.md:139-142`. 남은 옛 좌표 `risk-management:102-108`도 정리 대상이다(`C/specs/engine-safety/spec.md:229`).

8. **좌표·AST·반영 정확성 — 대체로 확인, 두 결함과 한 미측정 항목 잔존.**  
   AST JSON **39개 모두 현재 소스 SHA-256과 일치**했다. B15·B16·B25·B26과 `parkAlert` B2를 직접 대조했다. `EnqueueAlert :551`이 B2 바깥이라는 정정, `Announcer :639`, 생성자 두 곳, a066 `:158`, `VACUUM INTO :76`도 맞는다. K2·K4·단일 입구·캡 한정어·이력 순서·완화 뒤 재읽기는 문서와 tasks에 반영됐다. 누락은 전달 실패 승격의 무통지 범위, 과잉은 K3의 무조건 반환 금지다. flatten 도달성은 이번 코드 추적으로 해소된다.  
   근거: `C/analysis/head-ast-21/MANIFEST.txt:2-40`; `C/design.md:779-824`; `C/tasks.md:23-37`.

읽기 전용 정적 검토만 수행했습니다. 파일·git 상태 변경, 테스트 실행, 운영 원장·자격 증명 접근은 하지 않았습니다. git archive 트리이므로 커밋 식별자 자체는 독립 검증하지 않았습니다.

Recommendation: 23판의 모드 통지 예외와 K3 성공 경로 핀을 수정한 뒤 재리뷰 because 현재 문서대로는 유지할 무통지 동작이 새 SHALL을 위반하고 정상적인 커밋 실패 처리가 구조 검사에서 탈락한다.