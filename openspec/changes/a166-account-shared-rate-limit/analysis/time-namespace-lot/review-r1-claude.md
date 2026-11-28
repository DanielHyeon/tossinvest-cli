# 1회차 보이스 K — Claude 독립 서브에이전트 보고 원문(요약 없이 옮김)

판정: 권장안 가 채택 가능. P0/P1 없음, P2 1, P3 5.

## 코드만 읽고 낸 판단
- `Namespace !=` 네 항(bootstamp.go:67·70·73·86)을 빼고 `Snapshot` 검사를 모두 유지하면, 같은 boot_id 의 표본이 서로 다른
  시계로 비교되는 경우는 없다.
- 오프셋 기준: time ns 오프셋은 초기 ns 기준 절대값, 새 ns 는 만든 쪽의 값을 물려받고 멤버가 생긴 뒤 바뀌지 않는다.
  `Snapshot` 은 오프셋 0(L152)과 current==children(L145)을 요구 → 받아들인 표본은 모두 호스트 CLOCK_BOOTTIME. 중첩 ns ·
  suspend 도 같은 결론.
- 다중 스레드: `/proc/self` 는 스레드 그룹 리더, `timens_offsets` 는 리더의 time_for_children 오프셋. setns(time) 은 단일
  스레드에서만(EUSERS), unshare(CLONE_NEWTIME) 은 for_children 만 바꾸며 새 스레드는 부모의 current 를 물려받는다 →
  모든 스레드의 current 가 같다. x/sys `ClockGettime` 은 raw syscall. 이 커널 사실은 기억에 의존, 소스 확인 안 함.
- 다른 비교 경로: 프로세스 사이 BootStamp 비교는 `RemainingBootHint` 하나뿐. `checkOffline`(offline_linux.go:287)은 값을
  버리고, `engineready.go` 의 boot_id 는 별개 용도.
- 디스크 위조: event.state 위조는 지금도 uptime · boot_id 로 같은 효과 → 가 로 새로 생기는 공격력 없음.

## 문서 검증
1. 자리 표(§3) · 실패 순서(§4.1) file:line 모두 일치(L67/70/73/86, 호출 L165/230/396/531/598, 쓰기 L401·537·542·546·603).
   B8 죽은 분기 맞음. 영구 poison 경로(Finish L531 실패 → Begin L440–461)와 poison 해제 코드 0 확인.
2. shapes.json 에 clock · observation_linux 키 0 확인.
3. 뒤집히는 시험 3개: 정적으로도 3개(accountbudget fake 는 모두 `{1,1}`, `RemainingBootHint` 직접 호출 시험은
   bootstamp_test.go 뿐). overlay 직접 재실행에서도 같은 3개 실패. execgw `A166|EnginePoll`(무태그 · `tossos_testseams`)과
   app/engine `A166` ok.

## 지적
- [P2] execgw 통과 주장에 영수증 없음(로그에 4개 패키지만) → 재실행 원문을 남길 것(리뷰어 재실행: 두 태그 ok, 20.5s / 20.7s).
- [P3] §5 표에 「같은 boot_id 인데 커널이 다름 + 비초기 ns 참여자 + 공유 root」 누락(boot_id bind-mount 위조 · VM 스냅숏 복제;
  `procFS` 는 `/proc` 루트만 statfs, bootstamp_linux.go:35). 초기 time ns inode 는 상수 4026531834 라 호스트 ns 참여자에는
  이미 열려 있음 — 지원 전제 밖, 가드 불필요, 행만 추가.
- [P3] contract-delta §1 「진단에만 쓰고」 과장 — 생산 코드가 `Namespace` 를 읽는 곳 0 → 「형식 검사와 저장(호환)에만」.
- [P3] S5 로그는 B6(L145)과 L152 를 구분 못 함 — L152 는 추론으로 적을 것.
- [P3] §5 · §8 의 커널 전제 공백은 위 다중 스레드 근거로 채울 수 있으나 소스 인용 아님.
- [P3] FLM 4 번들은 조건부 — RED 6 을 기존 `observation_store_test.go` 에 넣으면 인접 함수가 잡힐 수 있으니 새 파일로 명시.

## 하지 않은 것
커널 소스 · 운영 컨테이너 안 봄, 전체 시험 · gate 미실행. 임시 overlay(`~/.cache/a166-tns-review2.*`) 삭제, 저장소 깨끗.
