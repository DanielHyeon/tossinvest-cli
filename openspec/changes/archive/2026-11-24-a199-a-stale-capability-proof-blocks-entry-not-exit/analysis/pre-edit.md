# 2.1 Pre-Edit 선언 (High-risk) — a199

base `603046a23`(a198 착지 merge 뒤 재고정). 편집 표면은 `analysis/edit-surface.md` 확정본(승인 14 함수 + 새 파일 2, `critical` · `stageKey` 는 들어가면 보고).

## 바꾸는 것
1. `attest` 새 파일: `Attestation.NamesAccount` — 비공개 `sameAccount`(Verify B6 와 같은 비교)를 부름. `Verify` · `Load` 무편집.
2. `engine` 새 파일: 공용 낡음 판정 함수 — `Verify` 성공 → 수락(가드 없음), 실패가 `ErrExpired`/`ErrEndpointNotAttested` 이고 형식 1..현재 · endpoint 항목 전부 비지 않음 · (만료면) `NamesAccount` 참 → 청산 전용, 그 밖 → 거절 오류. `AutomationStatus` 칸(`ExitOnly` 와 낡음 사유 문자열) · 청산 쪽 판정 메서드 · 감사 Detail · 기동 출력 문장 · 래치 detail.
3. `verifyGate` B6 → 공용 판정. 낡으면 반환하지 않고 B7~B11 끝까지, 다 통과하면 `status.ExitOnly`. 실패하면 그 조항 오류만.
4. `runInterlock`: 청산 전용 갈래 — Verified=false 유지 · EntryPermitted=false · 진입 래치(`capability_attestation_stale`) · 감사 `automation_gate.exit_only` · 로그. 래치를 세울 게이트가 없으면(`gateFacts.entry == nil`) 거절(게이트웨이 조항 오류) — 래치가 유일한 집행이라 fail-closed.
5. `logGateDecision`: 청산 전용이면 Warn + reason, 「보호는 프로세스 수명」 문장을 청산 전용에도.
6. `NewContext`: `gateFacts` 에 `entry: wiring.entry` 한 줄, B16 조건을 청산 쪽 판정으로. a166 census 키 하나 갱신.
7. 청산 쪽 소비자: `runEngineRun` B7(+ 출력 한 줄), `Context.ExitObserver` B2, `Context.ReconcileDriver` B2.
8. 감시: `judgeAttestation` 이 공용 판정을 부름, `newAttestationWatch` 가 기동 표지를 읽음, `prepare` 가 청산 전용이면 healthy 여도 회복 안내(키 = 날짜 · 에피소드, 기한 변경 초기화에서 보존), `eventFor` 문장.
9. `execgw`: 사유 상수 · `AllReasonCodes` · `latchOrder`(중앙 무결성 뒤 · 모드 앞) · golden. `audit` action 상수.
10. `soak.go` 문구(`runSoakAttest` · `newSoakAttestCmd`), `interlock.go` 머리 주석.

## 바꾸지 않는 것
- `Verified` 의 뜻 · 쓰는 자리, 전략 worker 조건(`Verified` 그대로 — 시험 고정), `Gateway.submit` · `checkEntry` · `checkProtection` · EntryGate 메서드, `attest.Verify` · `Load`, 게이트 OFF 경로, 루프 집합, 운영 모드 · 트리거 · outbox 타입, 설정 키, 유효한 attestation 의 판정 경로(Manager Q-B), a196 단계 · 등급 · 키 규칙.

## 되돌리는 법
- 이 change 의 Go 커밋을 revert(새 파일 삭제 · 편집 함수 원복 · census 키 원복 · golden 원복). 운영 반영 전이면 운영 영향 0. 반영 뒤라면 이전 이미지로 롤백 — 그때 낡은 attestation 은 다시 기동 거절이 된다(지금 동작).

## 고정 장치 점검(편집 뒤 선택 시험에 넣음)
- `./internal/accountbudgetruntime`(a166 census — NewContext 키 외 무변), a071 금지 문자열, 감독 루프 집합 시험, `TestA092*` 구조 핀, a098 사유 열거, golden.
