# a101 — soak는 배포에서 살아남는다

## Why

**2026-08-11 14:00Z에 측정했다.** a100의 조건주문 probe를 배포하려고 문서화된 절차
(`docker compose build` → `up -d`)를 그대로 실행했다. 컨테이너가 재생성됐고, 살아난 것은
engine·console·httpapi 셋뿐이었다. **capability soak는 죽은 채로 남았다.**

이것은 사소한 일이 아니다.

- 엔진의 automation gate는 capability attestation이 유효할 때만 기동한다
  (`engine-safety` spec, `internal/app/engine/interlock.go:518`).
- 그 attestation을 갱신하는 **유일한 경로**가 `tossctl soak run`이 쌓는 기록이다
  (`internal/soak/attest.go`). 기록이 48시간보다 오래되면 `soak attest`는 거부한다.
- 현재 attestation은 **2026-08-29에 만료된다.**

⇒ **문서화된 배포 절차가 무인 거래를 계속 가능하게 하는 시계를 조용히 멈춘다.**
아무 경고도 없고, 콘솔에도 "서베이가 죽었다"고 뜨지 않는다. 다음 배포가 8월 29일 전에
있고 아무도 soak을 다시 켜지 않으면, 엔진은 **a100과 무관하게** 기동하지 못한다.

### 왜 지금까지 문제가 되지 않았나

과거에는 호스트에 `soak-autostart.sh`가 있었다. a060이 2026-08-03에 그것을 은퇴시켰고
(archive `issues.md` I3), 은퇴 근거는 **"콘솔이 컨테이너로 옮겨가며 이 스크립트의 역할은
끝났다"**였다.

**그 전제에 구멍이 있었다.** 콘솔은 스크립트의 *수동 재시작* 역할은 물려받았지만
(`RestartSoak` seam, 대시보드 버튼) **"떠 있지 않으면 띄운다"는 감시자 역할은 물려받지
않았다.** a060은 죽은 산출물을 치운 것이 옳았고, 이 change는 그것을 되돌리는 것이 아니라
**아무도 물려받지 않은 절반을 마저 물려받는다.**

a060이 제거한 위험은 이 change에 없다. 그 스크립트는 플래그 없이 자식을 띄워
**기본 프로필에 두 번째 서베이**를 만들 수 있었다. 여기서는 콘솔이 이미 가진
`restartSoak` seam을 그대로 쓰므로 프로필 플래그(`soakArgs`)가 붙고, 소유 판정
(`soakFindProcesses(recordPath)`)이 이 기록에 붙은 서베이만 대상으로 한다 —
`operator-console` spec이 이미 요구하는 바로 그 판정이며, 그 조문은 **"기본 경로를 명시한
콘솔과 생략한 autostart는 같은 인스턴스다"**라고 autostart를 이미 상정하고 있다.

## What Changes

`engine.autostart`를 **그대로 본뜬다.** 그것의 audit 주석이 원리를 이미 적어 뒀다 —
"the separate process-lifecycle approval. **It does not grant order capability**".
프로세스 수명주기 승인은 능력 부여와 별개이고, soak은 능력을 아예 가질 수 없는
조회 전용 도구이므로 같은 형태가 더 강한 이유로 맞는다.

- `soak.autostart` config 키를 추가한다. **없으면 false**이고, 그것이 지금까지의 동작이다.
- 콘솔 기동 시 그 키를 읽어, ON이면 **이미 있는 `restartSoak` seam**으로 서베이를 세운다.
  실패는 콘솔 기동을 막지 않는다 — 조회 전용 도구가 없다고 운영자 화면을 못 뜨게 할 이유가 없다.
- **승인을 기록하는 곳은 새 화면이 아니라 이미 있는 soak 재시작 버튼이다.** 운영자가 그
  버튼을 누르는 행위가 곧 "이 프로필에서 서베이를 돌린다"는 의사표시이므로, 성공한 재시작이
  그 의사를 영속시키고 audit에 남긴다(`soak.autostart`).

## Non-goals

- **서베이가 무엇을 재는지는 바꾸지 않는다.** endpoint 목록도, 주기도, 판정도 그대로다.
- **주문 능력과 무관하다.** soak은 GET만 낸다(`internal/soak` 패키지 doc과 import-graph 테스트).
- **토글을 이 change가 켜지 않는다.** 안전 불변식 §0-7에 따라 운영 토글 flip은 사람이 한다.
  이 change는 키와 배선을 만들고, 켜는 것은 운영자의 버튼 한 번이다.
- **끄는 화면을 만들지 않는다.** 콘솔에는 애초에 soak 정지 버튼이 없다. 시작만 가능한 표면에
  중지 승인 UI를 새로 파는 것은 YAGNI이며, 끄려면 config를 고친다.
- a100의 조건주문 probe와 무관하다. 그 배포가 이 결함을 **드러냈을** 뿐이다.
