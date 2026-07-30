## Context

현재 `engine.automation_gate.enabled`와 네 개의 LIVE 거래 정책 토글은 콘솔에서
편집되고, [엔진 시작]은 `cmd/tossctl.startEngine` seam을 호출한다. 엔진은 그
뒤에도 journal flock, 설정 한도, capability attestation, 거래 정책, Guardian,
ExecutionGateway 인터록을 자체 검증한다.

`tools/engine-autostart.sh`는 준비되어 있지만 설치·활성화되지 않으며, 사용자의
부팅 의도가 config에 기록되지 않는다. Docker 배포에서 별도 host systemd 엔진
서비스를 추가하면 호스트 바이너리와 컨테이너 바이너리가 서로 다른 journal
writer가 될 수 있고, VPN/컨테이너 배포와 수명주기도 갈라진다.

## Goals / Non-Goals

**Goals:**

- 기본 OFF인 별도 autostart 승인 상태를 config와 콘솔에 제공한다.
- ON 저장 직후와 이후 콘솔 재기동/호스트 부팅에서 같은 `startEngine` 경로를
  사용한다.
- 자동 기동도 수동 [엔진 시작]과 동일한 startup interlock·flock·로그 경로를
  사용한다.
- autostart 한 키만 surgical write하고 변경을 audit에 남긴다.
- 컨테이너가 기존 host config/data를 정확히 보도록 배포 경로를 고정한다.

**Non-Goals:**

- autostart가 게이트, Guardian, attestation, 거래 정책 또는 주문 게이트를
  대신 켜지 않는다.
- 이 변경이나 배포 과정에서 autostart를 ON으로 설정하거나 실제 엔진을
  시작하지 않는다.
- OFF 저장이 실행 중 엔진을 정지시키지 않는다.
- 엔진 crash-loop 감독 정책을 새로 만들지 않는다. 기존 엔진의 실패는 화면과
  로그에 남고, 다음 콘솔 기동 전까지 자동 재시도하지 않는다.
- VPN 자체를 설치하거나 LAN/public 인터페이스에 콘솔을 노출하지 않는다.

## Decisions

1. **독립적인 `engine.autostart` boolean**

   automation gate와 autostart를 같은 키로 쓰지 않는다. 게이트는 프로그램 주문
   능력 승인이고 autostart는 프로세스 수명주기 승인이다. 누락은 false이며
   `DefaultFile`도 false로 고정한다. 대안인 “게이트 ON이면 항상 부팅 기동”은
   기존 게이트가 ON인 머신에서 배포 순간 엔진을 시작하므로 거부한다.

2. **콘솔 seam은 load/save 한 키만 소유**

   `AutostartSettings`는 boolean load/save만 제공하고 Guardian 한도나 gate
   값을 전달받지 않는다. `internal/config`는 파일 lock 안에서
   `engine.autostart` 한 멤버만 splice한다. 이는 다른 설정의 stale overwrite를
   타입과 문법으로 막는다.

3. **ON 저장 시 기존 `StartEngine` 호출**

   설정 저장이 성공한 뒤 `StartEngine`을 호출한다. 시작 성공·이미 실행·인터록
   거부 모두 기존 엔진 note에 기록하고 설정 화면 notice에도 결과를 요약한다.
   별도 process API나 별도 interlock 구현은 만들지 않는다. OFF는 설정만 저장한다.

4. **프로세스 시작 시 한 번만 조건부 시작**

   `runConsole`은 HTTP 리스너를 열기 전에 autostart 설정을 읽고, true일 때만
   기존 `startEngine`을 한 번 호출한다. 오류는 콘솔 자체를 종료시키지 않고
   engine note와 stderr에 남긴다. config 읽기 오류는 fail-closed로 자동 기동을
   건너뛴다.

   시작 호출은 서버 생성 이후 비동기 goroutine으로 하지 않는다. 테스트 가능한
   작은 helper가 `Load`와 `StartEngine` seam을 받아 동기적으로 판단하고,
   `runConsole`이 그 결과를 초기 note로 주입한다. 따라서 경쟁적인 이중 start를
   만들지 않는다.

5. **Docker 부팅 재시작은 콘솔 컨테이너가 소유**

   `restart: unless-stopped`와 이미 enabled인 Docker service가 호스트 부팅 때
   콘솔 프로세스를 다시 띄운다. 콘솔 시작 helper가 autostart를 판단하므로 Docker
   socket, host PID namespace, privileged service가 필요 없다. 엔진은 같은
   컨테이너 안의 같은 바이너리·config·session·data mount를 사용한다.

6. **현재 app data directory를 직접 매핑**

   Compose는 `TOSSOS_DATA_DIR=/var/lib/tossos/data`를 설정한다. 그렇지 않으면
   image의 `XDG_DATA_HOME=/var/lib/tossos/data`가 `.../data/tossos`를 다시
   붙여 host의 기존 `journal.db` 대신 빈 중첩 디렉터리를 보게 된다.

## Risks / Trade-offs

- [ON 저장은 실제 주문 능력 프로세스를 즉시 띄울 수 있음] → 메뉴에 기존
  비가역 편입·프로세스 수명 보호 문장을 유지하고 autostart 의미를 추가한다.
  실제 기동은 기존 startup interlock이 최종 판정한다.
- [컨테이너 재시작 때 매번 시작 시도] → journal flock과 marker/process 검사가
  중복 writer를 거부한다. autostart OFF는 호출 자체를 하지 않는다.
- [콘솔은 살았지만 자동 엔진 시작은 실패] → 콘솔 가용성을 유지하면서 engine
  note와 로그에 엔진 자신의 거부 사유를 남긴다.
- [OFF가 즉시 정지로 오해됨] → 화면과 저장 notice에 “다음 기동만 막으며 현재
  엔진은 [엔진 정지]로 끈다”고 명시한다.
- [컨테이너 안 detached child의 수명] → 콘솔이 PID 1이고 init이 신호·좀비를
  처리한다. 컨테이너 종료 시 모든 자식이 함께 종료되며 재기동 때 다시 판단한다.

## Migration Plan

1. schema/config에 optional boolean을 추가한다. 기존 파일은 누락=OFF다.
2. config seam, 콘솔 폼·route, startup helper를 RED/GREEN 순서로 추가한다.
3. Compose data-dir override와 운영 문서를 갱신한다.
4. 전체 race/test/vet/OpenSpec/SDD gate를 통과한다.
5. 현재 autostart 값이 OFF임을 확인한 뒤 Docker 콘솔을 배포한다.
6. rollback은 컨테이너를 내리고 기존 host HTTP launcher를 다시 실행하는 것이다.
   config의 새 키는 구버전이 모르는 필드가 되므로 rollback 전 false로 제거하거나
   새 배포를 유지한다.

## Open Questions

없음. crash-loop supervisor와 VPN 설치는 별도 변경으로 다룬다.
