# proposal-freeze r2 — codex (다른 계열, 마지막 회차), 2026-11-23

- 명령: `codex exec -s read-only -C ~/tossos-work/a198-impl -o <out> "<prompt>" </dev/null` (rc 0), codex-cli 0.165.0. 프롬프트는 이 파일을 만들기 직전 스크래치에서 옮겼고 a198 문안임을 확인(첫 줄 「TossOS change a198 proposal-freeze 리뷰 2회차」).

## 프롬프트

```text
TossOS change a198 proposal-freeze 리뷰 **2회차(마지막)**. 적대적 Eng 관점.
대상: openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/{proposal.md,design.md,tasks.md,specs/exit-policy/spec.md} 와 analysis/{claim-check.md,edit-surface.md,rate-limit.md}, docs/pm/portfolio/stories/STORY-TOS-a198.yaml. 정본 openspec/specs/exit-policy/spec.md. 1회차 원문 analysis/review/proposal-freeze-r1-codex.md.
현재 소스로 직접 확인할 것(internal/app/engine/{exitloop.go,exit_order_log.go,exit_held_proposal.go,exitwiring.go,engine.go,gateway.go,precheck.go}, cmd/tossctl/{engine.go,engine_assembly.go}, internal/execgw/{issue.go,gateway.go,guardian.go}, internal/journal/*, internal/reconcile/acked_boot.go, internal/flatten/flatten.go, internal/official/client.go).

1회차 발견과 Manager 처분(2026-11-23):
- P1-1 D4 warnEpisode 운영 무기록 → D4 를 a129 운영 로그 경로(noteExitOrder→ExitOrderLog)의 한 줄(종류 취소·상태·사유 코드만, 계좌·수량·가격·Detail 금지)로 변경. 그 경로의 기존 함수 편집은 exitOrderLine.args 1 개(취소 줄만 side/종류). Log 전체 연결은 기각.
- P1-2 CANCEL ACKED 고착 → 범위 밖·이름 붙인 잔여로 proposal·design §6·Story limits 에 공시, 시험 6 에 ACKED 고착 변형(지연 경보가 실제로 나가는지 단언; 안 나가면 범위 재판단).
- P1-3 전일 주문 거래일 결속 → 범위 밖, design §6 에 사실 기재, 고치지 않음.
- P1-4 사전 GET → 정본 「브로커 조회를 더하지 않는다」는 대상 찾기에 대한 것, 사전 단건 조회는 주문 경로 기존 계약(retry matrix 행 37)으로 봄. proposal Impact·design §2-4 에 「취소 1건 = 조회 1 + 취소 1, 동기」와 최악 지연(대상 수 × (2s + 15s))을 공시. spec delta 무변경.
- P2-5 D3 조립 실패 → D2 를 (가-B)(관측기 안에서 execgw.Issuer 구성, flatten 과 같은 IssueReduction)로 바꾸고 D3 삭제, 조립 Scenario 삭제. 운영 조립이 지도자인 시험(§4-7)은 유지.
- P2-6 결정 종류 단언 불가 → §4 단언을 「attempt kind=CANCEL·CONFIRMED·대상 주문 번호, 결정 행 멱등 키 없음, preimage 방향 = 주문 방향」으로, 변이를 「멱등 키가 붙는 발행으로 되돌리기」로 수정.
- 그 밖: 매수 취소 접수 뒤 재취소는 고치지 않고 design §2-5·§6 공시. a087 순서는 기다리지 않음(design §6).

요청: (1) 각 처분이 틀렸으면 파일:줄 근거와 함께 반박하라(특히 (가-B)의 시계·TTL 동일성, D4 의 「args 1개 편집」 충분성·계좌 비노출, ACKED 고착에서 지연 경보가 실제로 나가는지, 최악 지연 계산). (2) 새 발견은 P0/P1/P2 로, 사실과 추측을 구분하라. (3) 처분 결과 문서들 사이의 모순(proposal·design·spec delta·tasks·Story)을 찾아라.
파일을 수정하지 말 것. 읽기 전용 명령만(git·빌드·시험 실행 금지).
```

## 원문(마지막 메시지)

**판정: 현 문서 그대로 freeze 불가.** 핵심 수정 방향은 타당하지만, D4의 주문별 기록 보장과 최악 지연 계산은 반박됩니다. ACKED 지연 경보는 조건부로 도달 가능하며, 이번 읽기 전용 리뷰에서 실제 발송을 검증한 것은 아닙니다. P0 근거는 찾지 못했습니다.

1. **[P1 · 새 발견] D4의 기존 refusal 래치는 서로 다른 취소 실패를 합칩니다.**

   [design.md:10](openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/design.md:10)은 실패마다 대상 주문 번호를 기록하되, 같은 결과가 **연속 주기**에 이어질 때만 묶는다고 설명합니다.

   실제 [exit_order_log.go:203](internal/app/engine/exit_order_log.go:203)의 signature는 `event·action·level·reason·state`뿐이고, 저장 키는 `positionID`입니다. 대상 주문 번호·방향·attempt가 없습니다. :209에서 같은 signature면 **같은 주기 안에서도** 반환합니다.

   따라서 같은 포지션에서 주문 A와 B의 취소가 같은 사유로 실패하고 helper가 같은 action/level을 사용하면, A만 기록되고 B의 첫 실패는 버려집니다. 반대로 A와 B의 사유가 다르면 signature가 서로 덮어써져 다음 주기에도 두 줄이 반복됩니다. 이는 [spec delta:25](openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/specs/exit-policy/spec.md:25)의 무기록 금지와 D4의 중복 억제 설명을 동시에 만족하지 못합니다.

   **처분:** `noteExitOrder` 무편집을 유지하려면 새 helper가 주문별 래치 신원을 어떻게 제공하는지 설계에 명시해야 합니다. 현재 문서만으로 “기존 함수는 args 하나면 충분”이라는 결론은 성립하지 않습니다. 복수 주문의 동일 사유·상이 사유를 각각 두 주기 관측하는 시험이 필요합니다. 현재 §4-4의 단일 취소 시험으로는 잡히지 않습니다.

   수량·가격도 구분해야 합니다. [args:145](internal/app/engine/exit_order_log.go:145)는 해당 키를 항상 출력합니다. helper에서 값을 비우면 **값 비노출**은 가능하지만 **필드 부재**는 아닙니다. D4의 허용 필드만 출력하려면 취소 전용 args 분기가 필요합니다. 이 부분 자체는 args 한 함수 안에서 해결 가능합니다.

2. **[P1 · 새 발견] 사전 조회의 2초는 반환 상한이 아니므로 `N × 17초`를 최악 지연으로 사용할 수 없습니다.**

   [design.md:22](openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/design.md:22), [proposal.md:48](openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/proposal.md:48)의 계산에는 다음 누락이 있습니다.

   - [precheck.go:99](internal/app/engine/precheck.go:99)는 2초 context를 만들지만, [client.go:336](internal/official/client.go:336)은 HTTP 전송 전에 `tm.token(ctx)`를 호출합니다.
   - [token.go:63](internal/official/token.go:63)은 context로 중단할 수 없는 `mu.Lock()`을 사용합니다.
   - 다른 루프의 `refresh`도 같은 mutex를 잡고([token.go:111](internal/official/token.go:111)), 토큰 HTTP 요청과 응답 처리·캐시 저장까지 잠금을 유지합니다([token.go:149](internal/official/token.go:149)).

   **확인된 사실:** 다른 루프가 토큰 갱신으로 mutex를 보유하면 사전 조회의 context가 만료되어도 잠금 대기는 끝나지 않습니다. “사전 조회 시한 2초”를 전체 호출 소요시간 상한으로 사용할 수 없습니다. 실제 운영 지연 크기는 미측정입니다.

   또한 [client.go:359](internal/official/client.go:359)는 401에서 최대 두 번 추가 전송하므로, [rate-limit.md:9](openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/analysis/rate-limit.md:9)의 “주기당 POST 1 + GET 1 상한”은 물리 HTTP 요청 상한이 아닙니다. design은 401을 제외한다고 적었지만 rate-limit 문서는 제외하지 않습니다.

   **처분:** `N × 17초`를 “토큰 획득·경합·인증 재전송·원장 처리 등을 제외한 HTTP 시간 예산”으로 낮춰 표현하고, **전체 최악 지연은 미확정**이라고 명시해야 합니다. 이 발견이 토큰 경로 수리까지 자동으로 승인하는 것은 아닙니다.

3. **[P2 · 처분의 조건 누락] “ACKED는 연속 실패 계수에서 빠진다”는 항상 참이 아닙니다.**

   [proposal.md:33](openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/proposal.md:33), [design.md:55](openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/design.md:55), [Story:24](docs/pm/portfolio/stories/STORY-TOS-a198.yaml:24)는 이를 무조건적으로 기술합니다.

   실제 [onlyEngineCancelsInFlight:61](internal/app/engine/exit_held_proposal.go:61)은 다음 조건을 모두 요구합니다.

   - 치울 대상 목록이 비어 있지 않음.
   - 미종결 attempt가 모두 해당 대상의 진행 중 CANCEL임.
   - 모든 대상에 대응하는 CANCEL이 있음.

   대상 주문의 종결 체결 기록이 들어와 live 목록에서 빠지면, CANCEL이 여전히 ACKED여도 이 제외 조건은 깨집니다. 다른 미취소 대상이 남아 있어도 같습니다. **손절 차단은 계속되지만 연속 실패는 다시 셀 수 있습니다.**

   지연 경보 쪽은 처분을 뒤집을 근거가 없습니다. `clearTheSymbol`의 ACKED 반환 뒤 [record:1402](internal/app/engine/exitloop.go:1402)는 countable과 무관하게 `noteDelay`를 호출합니다. 주문 가능한 보호 발의가 계속되는 경우 `ChangedFromState`도 재판정을 유지합니다([snapshot.go:245](internal/exitpolicy/snapshot.go:245)).

   다만 [noteDelay:1925](internal/app/engine/exitloop.go:1925)는 첫 호출에서 시계만 시작하고, **30초 이상 경과한 후 다시 도달한 주기**에 알림을 요청합니다. 독립적인 30초 타이머가 아닙니다. 운영에서는 [exitwiring.go:333](internal/app/engine/exitwiring.go:333)의 기록 전용 경로를 거칩니다.

   **처분:** 공시를 위 조건으로 좁히고, §4-6은 전체 관측 주기를 반복해 실제 알림 기록을 단언해야 합니다. `noteDelay` 직접 호출이나 가짜 시계 이동만으로 “운영 경보 발송”을 증명하면 안 됩니다.

4. **[P2 · 문서 정합성] 잔여 위험 공시는 정본 위반의 해소와 구별해야 합니다.**

   전일 주문 결속을 이번 범위에서 고치지 않는 결정은 존중합니다. 다만 [design.md:56](openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/design.md:56)의 잔여는 [정본:579](openspec/specs/exit-policy/spec.md:579)의 확정 취소 재전송 금지 및 [delta:13](openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/specs/exit-policy/spec.md:13)의 “한 번 취소”를 충족한 상태가 아닙니다. **알려진 미충족 잔여**로 남겨야 합니다. 새 발견으로 중복 계산하지 않습니다.

   그 밖의 처분 반영 불일치는 다음과 같습니다.

   | 위치 | 남아 있는 충돌 |
   |---|---|
   | [rate-limit.md:11](openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/analysis/rate-limit.md:11) | 사전 GET 해석이 아직 “Manager 판단 대기”. design에서는 결정 완료 |
   | [edit-surface.md:37](openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/analysis/edit-surface.md:37) | 운영 시계 동일성 “미확인”. claim-check 후반에서는 확인 완료 |
   | [edit-surface.md:47](openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/analysis/edit-surface.md:47) | a087 착지 선행 필요. design §6에서는 기다리지 않기로 결정 |
   | [review.md:21](openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/review.md:21) | 여섯 처분 모두 결정 대기, :29는 “반영한 수정 없음” |
   | [tasks.md:38](openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/tasks.md:38) / [Story:25](docs/pm/portfolio/stories/STORY-TOS-a198.yaml:25) | task에는 승인 후 운영 반영, Story에는 운영 반영 제외 |

   역사적 기록은 보존하되, 현재 처분을 가리키는 정정·후속 기록이 필요합니다.

5. **[P2 · 증거 출처 오류] 1회차 파일의 프롬프트 원문이 다른 리뷰입니다.**

   [proposal-freeze-r1-codex.md:10](openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/analysis/review/proposal-freeze-r1-codex.md:10)은 a196 `attestation_watch.go` 구현 리뷰이며, Git diff 실행도 지시합니다. 뒤의 결과 본문은 a198 설계 리뷰입니다.

   본문 발견은 현재 소스로 재확인할 수 있지만, 이 파일을 a198의 정확한 실행 프롬프트 증거로 사용할 수 없습니다. 원본을 덮어쓰지 말고 별도 정정 기록으로 출처를 연결해야 합니다.

**반박하지 않는 처분도 명확합니다.**

- **(가-B) 시계:** 운영은 [engine.go:217](cmd/tossctl/engine.go:217)의 같은 `clk`를 assembly와 runtime에 전달하고, gateway와 observer가 각각 받습니다. 동일성 확인됨.
- **TTL:** flatten과 Issuer 기본값 모두 현재 60초입니다. 서로 다른 상수라는 [claim-check.md:81](openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/analysis/claim-check.md:81)의 설명도 정확합니다.
- **D3 삭제:** 새 주입 필드·인터페이스가 없어지므로 해당 신규 조립 실패 조건은 사라집니다.
- **D4 계좌 비노출 방향:** 고정 사유 코드와 명시 필드만 복사하면 가능합니다. `positionID`는 계좌 원문이 아니라 해시입니다([position_projection.go:441](internal/journal/position_projection.go:441)). 신규 helper는 아직 없으므로 구현의 비노출까지 검증한 것은 아닙니다.
- **P2-6 수정:** attempt 종류·대상 번호·키 부재·방향 단언은 관측 가능한 계약입니다. 다만 CANCEL→AMEND 발행 종류 변이는 여전히 저장 결과로 구별되지 않습니다.

검증 범위: 현재 작업 파일의 관련 함수·호출 경로를 정적으로 읽고, 기존 cited AST **41개 모두 소스 해시 일치**를 확인했습니다. 이는 시험 실행이나 전체 `internal/journal/*` 감사가 아닙니다. 파일 변경·Git·빌드·시험·AST/그래프 재생성은 실행하지 않았습니다. ai-memory 조회는 네트워크 권한 오류로 실패했습니다. `impeccable: N/A — backend/documentation-only`.