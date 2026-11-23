# 보강분 freeze 리뷰 r1 — codex (다른 계열, 1회차·마지막), 2026-11-24

- 명령: `codex exec -s read-only -C ~/tossos-work/a198-impl -o <out> "<prompt>" </dev/null`(rc 0), codex-cli 0.165.0. 프롬프트 원문: 같은 디렉터리 `delta-freeze-r1-prompt.txt`(실행에 넘긴 파일을 그대로 복사).

## 원문(마지막 메시지)

**판정: 보강안 그대로 동결하기에는 차단 사유가 있습니다. P1 1건, P2 5건. P0는 입증하지 못했습니다.** 기준 HEAD `72be1707e7396dc66760f3918947ddafac848051`. D7·D8·D4a는 예정 설계로 평가했으며, 현재 구현 결함과 구분했습니다.

1. **P1 — `already-modified`가 후계 매도 주문을 남긴 채 청산 보류를 해제합니다.**
   
   **사실:** [D7]( /home/daniel/tossos-work/a198-impl/openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/design.md:13)은 이를 확정 거절로 포함하지만, 현재 [resolveCancel:145](/home/daniel/tossos-work/a198-impl/internal/execgw/amend_indoubt.go:145)는 `REPLACED`를 의도적으로 park합니다. D7 뒤 flatten은 [classifyCancel:458](/home/daniel/tossos-work/a198-impl/internal/flatten/flatten.go:458)에서 `Failed`로 종결하고, [UnsettledCancelSymbols:324](/home/daniel/tossos-work/a198-impl/internal/journal/flatten.go:324)는 `Failed`를 매도 보류 대상에서 제외합니다. 단순한 실패 보고 변경이 아니라 **청산 허용 변경**입니다.
   
   엔진도 옛 주문의 종결 스냅숏이 들어오면 [미체결 목록:1882](/home/daniel/tossos-work/a198-impl/internal/journal/fills.go:1882)에서 제외합니다. 후계 주문이 원장에 없다면 그 주문을 치우지 못합니다.
   
   **조건부 위험:** 조회·취소 사이 정정되어 후계 SELL이 살아 있으면 추가 SELL 경로가 열립니다. 실제 초과 매도는 브로커 가용수량·거절에 달려 있어 확정하지 않습니다. flatten의 가용수량 제한도 남아 있습니다. **후계 노출 해소 계약 없이 `already-modified`를 종결 목록에 넣지 않는 처분이 필요합니다.** 증거 문서의 “Manager 결정 필요”와 D7의 “결정됨”도 불일치합니다.

2. **P2 — 마지막 재전송의 거절을 전체 CANCEL attempt의 ‘취소 미성립’으로 일반화합니다.**
   
   [withCancelRefusal:90](/home/daniel/tossos-work/a198-impl/internal/execgw/cancel_refusal.go:90)는 같은 attempt에서 재전송합니다. `already-processing → already-canceled`라면 마지막 요청이 취소할 대상이 없다는 사실과 앞선 전송이 취소를 성립시켰는지는 별개입니다. [API 계약:6588](/home/daniel/tossos-work/a198-impl/docs/migration/openapi.latest.json:6588)은 취소 주체를 증명하지 않습니다.
   
   현재 관측 해소는 취소 상태를 확인하면 `CONFIRMED`로 기록하지만, D7은 `FAILED_CONFIRMED → FlattenStepFailed`로 고정하고 재개 때도 건너뜁니다([flatten:317](/home/daniel/tossos-work/a198-impl/internal/flatten/flatten.go:317)). 기동 복구도 종결 attempt를 재관측하지 않습니다([recovery:271](/home/daniel/tossos-work/a198-impl/internal/reconcile/recovery.go:271)).
   
   **처분:** 단일 요청 거절과 재전송 묶음의 결과를 구분해야 합니다. 위 응답 순서는 가능한 반례이며 실물 관측 사실은 아닙니다. “code가 서로소라 재시도 창에 들어가지 않는다”만으로 해결되지 않습니다.

3. **P2 — 새 정본 delta와 기존 정본이 같은 입력에 반대 결과를 요구합니다.**
   
   [새 delta:5](/home/daniel/tossos-work/a198-impl/openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/specs/order-execution/spec.md:5)는 네 code의 취소를 확정 거절로 요구합니다. 기존 [Scenario:592](/home/daniel/tossos-work/a198-impl/openspec/specs/order-execution/spec.md:592)는 `already-processing`이 아닌 취소 409를 결과 불명으로 요구합니다. ADDED만으로는 두 요구를 동시에 만족할 수 없습니다.
   
   또한 delta는 정책 유무·status 제한이 없는데, [증거 §2:30](/home/daniel/tossos-work/a198-impl/openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/analysis/d7-d8-evidence.md:30)는 “모호인 409만”, 정책 nil은 제외하는 안입니다. **기존 Scenario를 명시적으로 수정하고 적용 범위를 하나로 확정해야 합니다.**
   
   판정 순서는 모순 code 보존 → D7 대상 판정 → 해당하지 않을 때 a165 판정이어야 합니다. 첫 응답과 재전송 응답 모두 다뤄야 하며, 재전송 중단·미전송 결과의 기존 보수 처분도 보존해야 합니다.

4. **P2 — D8 체결 경합의 회복 설명이 현재 프로젝션과 맞지 않습니다.**
   
   [증거 §6:60](/home/daniel/tossos-work/a198-impl/openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/analysis/d7-d8-evidence.md:60)은 뒤늦은 매수 체결이 보유를 늘리고 다음 청산 판정이 보호한다고 설명합니다. 하지만 손절이 진행되어 `CLOSING`이면 [전이표 E31–E33](/home/daniel/tossos-work/a198-impl/internal/position/table.go:78)은 `ENTRY_WHILE_CLOSING` 거절입니다. [Apply:249](/home/daniel/tossos-work/a198-impl/internal/position/position.go:249)는 기존 수량으로 반환하고, [프로젝션:101](/home/daniel/tossos-work/a198-impl/internal/journal/position_projection.go:101)은 정상 수량 갱신 대신 RECONCILE로 갑니다.
   
   **처분:** 매수 체결이 손절 접수 전·후인 경우를 나눠야 합니다. 후자는 대사·보호 수량 회복까지 시험해야 하며, “다음 주기 프로젝션으로 보호”를 회복 근거로 삼을 수 없습니다. D8 자체의 신규 초과 매도는 입증하지 않았습니다.

5. **P2 — §4-10~16과 변이가 핵심 판정 경계를 충분히 고정하지 못합니다.**
   
   [시험·변이 목록:55](/home/daniel/tossos-work/a198-impl/openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/design.md:55)에 다음 수용 조건이 빠져 있습니다.
   
   - 네 `already-*` 각각, 특히 **후계 주문 없는 `already-modified`**와 flatten 매도 보류.
   - `already-processing → already-canceled`, 중단·미전송, 재시작 후 상태. “첫 응답만 D7 적용” 변이.
   - status 독립성·nil 정책·읽을 수 없는 code 및 **AMEND에도 적용**하는 변이.
   - D8의 다른 계좌·시장·종목·거래일 취소 증거, 조회 오류, 늦은 매수 체결. [ConfirmedCancelOf:157](/home/daniel/tossos-work/a198-impl/internal/journal/exit_proposal_release.go:157)는 범위를 결속하지만 취소 intent 거래일과 주문 거래일이 같아야 합니다. 전일 매수는 D8을 통과하지 못하므로 “재취소 금지” delta의 무조건적 문언과도 맞지 않습니다.
   - D4a의 **가격 해석 실패**, 첫 로그 쓰기 패닉 후 다음 주기 재기록. §16은 수량 실패만 명시하여 “표시를 쓰기 앞에” 변이의 관측점이 불완전합니다.
   
   마지막 exit-policy Scenario도 다른 미치움 주문·미종결 attempt가 없다는 전제가 필요합니다. 현재 “다음 주기 손절 재제출”은 본문의 다른 주문 판정 보존 의무보다 강합니다.

6. **P2 — 보강 증거가 인용한 `clearTheSymbol` AST는 현재 소스와 불일치합니다.**
   
   [증거 서두:3](/home/daniel/tossos-work/a198-impl/openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/analysis/d7-d8-evidence.md:3)는 같은 작업 트리 AST라고 하지만, [해당 AST:3](/home/daniel/tossos-work/a198-impl/openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/analysis/cited-ast/ExitObserver_clearTheSymbol.ast.json:3)의 `source_sha256`는 현재 `exitloop.go`와 다릅니다. 확인한 나머지 다섯 AST는 일치했습니다. **이 AST의 B9·B10을 현재 분기 번호의 증거로 사용할 수 없습니다.** 이번 판단은 현재 소스를 직접 대조했습니다.

D7의 확정 거절 자체가 원래 PLACE 주문을 직접 은퇴시키지는 않습니다. 위험은 종목 차단과 flatten 보류의 해제에서 생깁니다. D8의 `CONFIRMED`도 취소 완료 관측이 아니라 접수 확정입니다. 취소에는 [round-trip 확인이 없습니다](/home/daniel/tossos-work/a198-impl/internal/execgw/roundtrip.go:71).

파일·Git 상태 변경, 빌드·시험·실거래 실행은 하지 않았습니다. 기존 탐침 결과는 문서 증거로만 읽었으며 재현하지 않았습니다. ai-memory CLI 조회는 연결 권한 오류로 실패했고, 그래프 갱신·새 AST 생성도 하지 않았습니다. `impeccable: N/A — backend/documentation-only`입니다.

