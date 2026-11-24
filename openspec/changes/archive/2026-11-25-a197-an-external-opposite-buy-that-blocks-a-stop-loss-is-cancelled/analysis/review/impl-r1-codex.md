# 구현 적대 리뷰 r1 — codex (원문)

- 날짜 2026-11-25 · 보이스: codex-cli 0.165.0 `codex exec -s read-only -C ~/tossos-work/a197-impl -o …` (model 로그 표기 `gpt-6-astra`), stdin </dev/null, rc=0 · 기준 HEAD fb5a4a0bc. 프롬프트에 git -C 절대경로 · toplevel 단언 · 시험 실행 금지를 적음.

## 프롬프트

```text
TossOS(실제 돈을 다루는 Go 자동매매 엔진) change a197 의 **구현**을 적대적 Eng 관점으로 리뷰하라(구현 적대 리뷰 1회차). 파일을 수정하지 말 것(읽기 전용). git 명령은 반드시 `git -C /home/daniel/tossos-work/a197-impl …` 로만 치고, 먼저 `git -C /home/daniel/tossos-work/a197-impl rev-parse --show-toplevel` 이 /home/daniel/tossos-work/a197-impl 인지 확인하라. 다른 경로에 git init·commit·config 를 하지 말 것. 시험을 돌리지 말 것(계측은 작성자가 함).

대상: 브랜치 HEAD fb5a4a0bc 의 Go diff(origin/main 대비) — 파일 /tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/301659bc-5238-4fe8-bdaf-873fdfb91f01/scratchpad/a197-impl-review-diff.txt. 계약: openspec/changes/a197-an-external-opposite-buy-that-blocks-a-stop-loss-is-cancelled/ 의 design.md(D1~D11, §2~§5), specs/exit-policy/spec.md, specs/order-execution/spec.md, review.md 의 「Manager 결정」 절들. 새 코드: internal/app/engine/a197_external_opposite_buy.go, internal/journal/a197_external_cancel_issued.go. 편집한 기존 함수: ExitObserver.submit · alertProposalRefused(exitloop.go), endExitOrderCycle(exit_order_log.go), Context.ExitObserver(exitwiring.go), mergeEngine(config/engine.go), NewContext(engine.go). 시험: internal/app/engine/a197_*_test.go, internal/journal/a197_*_test.go, internal/config/a197_*_test.go. 변이 원장(1판): analysis/mutation/run1/.

찾아서 P0/P1/P2 로 분류하라(각 항목에 file:line 근거):
1. 사람 주문을 잘못 취소하는 경로: 매도·다른 종목·엔진 주문(귀속 판독)·식별 불완전 행·이미 낸 번호·닫힌/바뀐 포지션·토글 꺼짐.
2. 손절·청산을 늦추거나 막는 경로: 주기 끝 실행의 시간(시한·ctx), 패닉, park 실패, 알림 기록 실패(계정 래치), 같은 종목 미종결, 진입 게이트.
3. 설계와 코드의 어긋남(D1~D11, M-1~M-3, O-1) — 예: 「이미 냈다」 키, 재검증 판정식, critical 종류·key, 알림 꺼짐 처리, park 전이·사유, 조회 시한·쪽 상한·커서.
4. 기존 동작 회귀: alertProposalRefused 문장 갈래, submit 의 B15 순서, endExitOrderCycle 의 조기 반환, NewContext audit 순서, census.
5. 시험이 생산 계약을 가리는 곳(가짜 판독기·가짜 브로커), 놓친 변이.
결론은 한글로: 판정 한 줄, 발견 목록(P0/P1/P2, 근거 좌표), 각 발견에 수리 제안(새 가드·정책은 「제안」으로만 표시).
```

## 최종 메시지

판정: **수정 필요 — P1 2건·P2 2건. P0 및 확정적인 오취소 경로는 이번 정적 리뷰에서 발견하지 못했습니다.**

1. **P1 — 단계의 25초 만료가 취소를 미종결로 남겨 손절을 재시작까지 막는다.**

   근거: [a197_external_opposite_buy.go:162](/home/daniel/tossos-work/a197-impl/internal/app/engine/a197_external_opposite_buy.go:162), [dispatch.go:206](/home/daniel/tossos-work/a197-impl/internal/journal/dispatch.go:206), [gateway.go:751](/home/daniel/tossos-work/a197-impl/internal/execgw/gateway.go:751).

   `Submit.Cancel(stageCtx, …)`는 전송과 결과 원장 기록에 같은 ctx를 사용한다. 인증 갱신·예산 대기·전송 도중 25초가 끝나면, 정상 DB에서도 `MarkInDoubt`가 취소된 ctx 때문에 실패한다. 게이트웨이는 상태가 빈 결과를 반환하고, a197의 `StateInDoubt` 분기를 타지 않아 park되지 않는다.

   남은 `DISPATCH_STARTED`는 세션 스윕 대상 밖이며, 다음 손절은 [exitloop.go:1677](/home/daniel/tossos-work/a197-impl/internal/app/engine/exitloop.go:1677)의 미종결 가드에 막힌다. D11에 원장 실패 잔여는 공시돼 있지만, **새 단계 시한 자체가 정상 원장의 기록 실패를 만드는 경로**다. 빠른 5xx 시험은 이 경로를 검증하지 않는다.

   수리 **제안**: 전송 종료 후 결과 기록에 사용할 제한된 정리 ctx를 전송 시한과 분리한다. 실제 전송이 반환한 결과를 기록한 뒤 기존 park 전이를 적용해야 하며, 진행 중 attempt를 임의로 종결하면 안 된다. 단계 시한 만료 후 원장 상태·다음 손절까지 확인하는 결합 시험이 필요하다.

2. **P1 — critical 기록 실패 후 취소를 계속하고, 누락된 알림은 재시도되지 않는다.**

   근거: [a197_external_opposite_buy.go:180](/home/daniel/tossos-work/a197-impl/internal/app/engine/a197_external_opposite_buy.go:180), [같은 파일:414](/home/daniel/tossos-work/a197-impl/internal/app/engine/a197_external_opposite_buy.go:414), [같은 파일:357](/home/daniel/tossos-work/a197-impl/internal/app/engine/a197_external_opposite_buy.go:357).

   `recordEpisodeCritical`의 성공 여부를 버리므로 첫 critical 적재가 실패해도 사람 주문 취소가 진행된다. 취소 attempt가 기록되면 다음 주기에는 `ExternalCancelIssued`가 대상을 제외한다. 첫 알림뿐 아니라 park 이후 결과 알림도 실패하면 다시 기록할 경로가 없다.

   생산 알림기는 실패 시 계정 진입 래치를 건다([record_only.go:137](/home/daniel/tossos-work/a197-impl/internal/obs/record_only.go:137)). 그러나 래치는 누락된 취소 통지나 attempt 해동 안내를 복원하지 않는다. M-2의 프로세스 중단 창과 별개로, **프로세스가 살아 있어도 일시적 적재 실패가 영구 인계 누락**이 된다. 시험 기록기는 항상 성공한다([a197_external_opposite_buy_test.go:93](/home/daniel/tossos-work/a197-impl/internal/app/engine/a197_external_opposite_buy_test.go:93)).

   수리 **제안**: 첫 critical 성공을 외부 취소의 선행 조건으로 삼고, 결과 알림 실패는 취소 재발행과 독립적으로 재기록한다. 기존 계정 래치 정책은 유지한다. 실제 Notifier의 적재 실패·복구를 주입해 확인해야 한다.

3. **P2 — attempt가 없는 실패에도 “다시 취소하지 않는다”고 잘못 알린다.**

   근거: [a197_external_opposite_buy.go:197](/home/daniel/tossos-work/a197-impl/internal/app/engine/a197_external_opposite_buy.go:197), [같은 파일:219](/home/daniel/tossos-work/a197-impl/internal/app/engine/a197_external_opposite_buy.go:219), [같은 파일:397](/home/daniel/tossos-work/a197-impl/internal/app/engine/a197_external_opposite_buy.go:397).

   결정 발행 실패나 게이트웨이의 기록 전 거절에도 동일한 `uncleared` 알림을 낸다. 본문은 재취소하지 않는다고 단정하지만, CANCEL attempt가 없으므로 다음 거절 주기에 다시 대상이 된다. D7·M-1이 명시한 기록 전 실패의 의미와 반대다. 같은 알림 key가 래치되면 이후 실제 결과 설명도 생략될 수 있다.

   수리 제안: attempt 유무에 따라 문장을 구분한다. 기록 전 실패에는 다음 주기 재평가 가능성을 설명하고, “다시 취소하지 않는다”는 영속 표시가 생긴 경우에만 사용한다.

4. **P2 — 변이 원장 1판의 M22는 audit 실패 시험을 선택하지 않았다.**

   근거: [M22.txt:2](/home/daniel/tossos-work/a197-impl/openspec/changes/a197-an-external-opposite-buy-that-blocks-a-stop-loss-is-cancelled/analysis/mutation/run1/raw/M22.txt:2), [a197_startup_audit_test.go:68](/home/daniel/tossos-work/a197-impl/internal/app/engine/a197_startup_audit_test.go:68).

   실행 패턴은 `-run 'A197Startup'`이다. 실패 시험 이름 `TestA197AnUnrecordableToggleChangeRefusesStartup`은 매칭되지 않는다. 따라서 M22 생존은 해당 실패 시험의 검출 능력을 평가한 결과가 아니다.

   수리 제안: 정확한 실패 시험 이름을 선택해 무변이 대조군과 변이를 다시 계측한다. M05도 미사용 변수로 **컴파일 실패**했으므로 귀속 보호 검출 증거가 아니다. M20은 1판 기준 생존이며, 현재 HEAD의 후속 청산 시험(`:641`)으로 잡히는지는 새 계측 전까지 미확인이다.

저장소 경로와 HEAD `fb5a4a0bc00ecebfa6e3ab24ee46c13e9554771f`를 확인했습니다. 지정 diff와 현재 `origin/main → HEAD` Go diff의 SHA-256은 동일합니다. 토글 OFF, 매도·다른 종목·귀속 번호 제외, M-1 키, M-3 판정식, B15 순서, 주기 끝 조기 반환 앞 배치, audit 위치와 census 반영에서는 추가 불일치를 확인하지 못했습니다.

파일 수정·시험·계측은 수행하지 않았습니다. 작성자 변이 결과는 현재 HEAD의 재실행 결과로 취급하지 않았습니다. ai-memory CLI 조회는 접근 제한으로 실패했고, 그래프 재질의·AST 재생성 없이 기존 산출물과 현재 소스를 대조했습니다. `impeccable: N/A — 백엔드 구현 리뷰, UI 변경 없음`.