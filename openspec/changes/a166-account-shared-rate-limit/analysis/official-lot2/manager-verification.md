# a166 공식 클라이언트 핵심 로트 Manager 판정 — 2026-10-31

**lot2a 부분 병합 승인.** 대상 `79d45e96ca30a782991c9a53aa468279fc81ea2b`. 명시적 `NewWithAccountBudget`, private guard 봉인, 같은 token client와 a112 clone, enforcing 오류 identity 전달을 포함한다. 생산 호출자 주입은 0이며 기존 `New` 기본 OFF 경로를 유지한다. replay/probe·전체 호출자 배선·429 폴링·계좌 결속·운영 적용은 미완이다.

Manager는 `cbf53d452` 이후 lot2a 생산 diff를 손으로 읽고 후속 authority 수정 차이와 영수증 원문을 대조했다. AST 전체 증명이라고 표시하지 않는다. budget wrapper 불일치에서 기존 raw transport로 fallback하여 권한이 복원될 수 있던 경로를 발견해 수정했고, `budgetEnforced`가 있는 client의 불일치는 false로 끝난다. 이는 private 봉인 계약이며 공개 악용 경로를 확인했다는 뜻은 아니다. 기본 OFF의 기존 fallback은 유지된다.

## 검증 근거

- 독립 Terra의 b5d13bf5f 대상 official 일반 3.851초·tagged race 8.921초, execgw 취소 409 회귀 tagged race 4.090초 PASS. 원문 `/tmp/a166-lot2-independent-official-normal.log`, `...official-tagged-race.log`, `...execgw-tagged-race.log`를 직접 읽었다.
- 후속 최종 authority delta의 독립 일반 0.035초·tagged race 1.052초 PASS와 P0/P1 미발견. 원문 `/tmp/a166-lot2-authority-independent-normal.log`, `...tagged-race.log`를 읽었다.
- 최종 SHA의 독립 Go 1.25.0 official A166/A112MBUS 선택 일반 3.858초·tagged race 9.358초 PASS. `/tmp/a166-lot2-final-go125-normal.log`, `...tagged-race.log`를 읽었다. Alpine 운영 환경 검증으로 확대하지 않는다.
- 작성자 RED/GREEN·핵심 변이·사전 AST/FLM·BTM 및 gstack 수동 범위 검토는 `pre-edit.md`와 동봉 원문에 있다. 중간 시험 로그 두 개가 덮어써져 원문을 보존하지 못한 사실과 no-tests-to-run 실행을 PASS로 쓰지 않은 처분을 유지한다.

반환 `ErrNotSent`는 마지막 HTTP 요청에만 결속한다. token 전송과 주문 본체, 앞선 401 본체 전송을 합쳐 거래 전체 미전송으로 판단하지 않는다. 기존 execgw의 단조 tracker와 a165의 최초 AMBIGUOUS 보존을 유지했으며, 이후 배선에서 이 sentinel 하나로 downgrade하는 구현은 허용하지 않는다.

Terra는 독립 영수증을 보존하고 부분 완료 범위만 tasks/PM에 기록한 뒤 main에 non-force 병합·push할 수 있다. a166 전체 항목 완료 체크·archive·전체 시험/전체 gate·실계좌·운영 토글은 이번 승인에 포함되지 않는다. a087 admission nil 및 a169 측정 보류는 유지한다.
