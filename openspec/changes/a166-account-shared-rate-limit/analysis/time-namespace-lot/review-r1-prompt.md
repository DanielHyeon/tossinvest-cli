너는 작성자(Claude)와 다른 계열의 독립 적대 리뷰어(Eng 관점)다. TossOS(실제 돈을 다루는 Go 자동매매 엔진) change a166 의
「time namespace 가 다른 참여자」 로트 1단계 설계안을 리뷰한다. 한국어로 답한다.

작업 디렉터리: /home/daniel/tossos-work/a166-time-namespace (브랜치 feat/a166-time-namespace, 커밋 aaad76d7e).
읽기 전용으로 리뷰한다. 파일을 고치지 않는다. 네트워크 · docker · 브로커 · 운영 경로 접근 금지.
시험 픽스처를 만들어야 하면 `set -euo pipefail`, 모든 git 명령은 `git -C <절대경로>`, 시작 시 `git -C <dir> rev-parse --show-toplevel`
로 toplevel 단언, 임시 디렉터리는 /home/daniel/.cache/a166-tns-review.* 아래에만. 저장소 안에 파일을 만들지 않는다.

읽을 것(이 순서):
1. 코드: internal/clock/bootstamp.go, internal/clock/bootstamp_linux.go, internal/accountbudget/observation.go,
   internal/accountbudget/observation_linux.go, internal/accountbudget/transport_linux.go(191–360행),
   internal/execgw/engine_poll.go, internal/accountbudgetruntime/offline_linux.go(278–293행), internal/accountbudgetruntime/runtime_linux.go(65–103행).
2. 설계 · 증거: openspec/changes/a166-account-shared-rate-limit/analysis/time-namespace-lot/ 의 investigation.md,
   contract-delta.md, overlay-option-ga.diff, run-current-head.log, run-overlay-option-ga.log, selected-tests-baseline-vs-overlay.log, census.txt.
3. 원 계약: openspec/changes/a166-account-shared-rate-limit/analysis/poll-clock-contract.md, design.md 70–101행.

판정할 질문:
Q1. 권장안(가) — RemainingBootHint 의 4개 조건에서 `Namespace !=` 항만 빼고 Snapshot 의 표본 안 검사(현재==자식 ns,
    offset 0, 앞뒤 ns · boot_id 동일)는 유지 — 가 같은 boot_id · offset 0 전제에서 시계 비교를 잘못되게 만드는 반례가 있는가?
    특히: 다중 스레드 Go 프로세스의 스레드별 nsproxy, /proc/self 가 가리키는 스레드, offsets 파일이 어느 ns 를 보여 주는지,
    CLOCK_BOOTTIME 의 vDSO 경로, 중첩 time ns, suspend, 컨테이너 런타임이 boot_id 를 가상화하는 경우.
Q2. investigation.md §5 「새로 받아들이는 입력」 표에 빠진 입력이 있는가? 어느 가드에도 안 걸리는 새 입력이 있으면 지목.
Q3. §3 자리 표(a/b/c)와 §4.1 실패 순서가 코드와 맞는가? 틀린 줄 · 빠진 자리를 file:line 으로.
Q4. 통과 이유가 바뀌는 기존 시험이 정말 3개뿐인가(overlay 로그 근거 포함)? FLM 번들 수(4) 계산이 맞는가?
Q5. contract-delta.md 문안이 코드와 어긋나거나 과장하는 곳.

출력 형식: 발견마다 [P0/P1/P2/P3] 제목, 근거(file:line 또는 로그 줄), 반례 또는 수정 제안. 마지막에 「권장안 가 채택 가능 / 불가」 한 줄.
새 가드 · 새 정책 · 범위 확장 제안은 해도 되지만 「범위 확장」 표지를 붙인다. 발견이 없으면 「없음」이라고 쓴다.
