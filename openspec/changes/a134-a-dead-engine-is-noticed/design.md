# a134 design — 죽은 엔진은 엔진 밖에서 알아챈다

기준 commit `421bf53c`(base-commit.txt) · codegraph 1.6.0 · 증거 `analysis/code-context/`.

## D0. 「ntfy 가 박동 부재를 스스로 경보한다」는 근거가 없다

`internal/obs/heartbeat.go` 파일 주석(「ntfy does this natively」)과 `engine-safety` 「등급화된 알림」의 「heartbeat(예상 주기 초과 시 ntfy 측에서 경보)」는
같은 전제를 쓴다. 저장소 안의 ntfy 서술(`docs/operations.md` 「ntfy 구독 필터」 · `AGENTS.md` ntfy 예시 · `README.md`)은 발행과 구독 필터만 말하고
수신 측 침묵 경보는 어디에도 없다(evidence-reconciliation R5). ntfy 는 받은 메시지를 구독자에게 전달하는 서버이며 「오지 않은 메시지」를 판정하는
기능을 이 저장소는 쓰지도 문서화하지도 않는다. 그러므로 이 change 는 그 전제를 **거짓으로 취급**하고, 침묵의 판정을 엔진 밖에 산 프로세스에 둔다.

그 결과 `obs.Heartbeat`(ntfy 로 1분마다 우선순위 1 메시지를 보내는 타입)는 수신 측이 아무것도 판정하지 않으면 소음일 뿐이다. 생산 생성자도 0 이다(R1).

## D1. 엔진 쪽 생존 신호 = 이미 있는 엔진 활성 마커

엔진은 부팅 6단계에서 `enginelock.Hold` 로 저널 디렉터리의 `engine-run.json` 을 쓰고, 프로세스가 사는 동안 별도 goroutine 이 `RefreshEvery`(1분)마다
다시 쓴다. 쓰기는 임시 파일 → `Chtimes` → rename 이라 원자적이고(enginelock `write`), 정상 정지(`Release`, `runEngineRun` 의 defer)는 파일을 지운다.
내용은 pid · started_at · refreshed_at · binary · note · ready_at · proc_instance — 계좌 칸이 없다(R2 · R3).

이것이 Manager 스케치의 「엔진이 설정 디렉터리에 원자적으로 다시 쓰는 작은 heartbeat 파일」 그 자체다. 그래서 **엔진 코드를 바꾸지 않는다.**

| 안 | 판정 | 근거 |
|---|---|---|
| A. 기존 마커를 생존 신호로 읽음 | **채택** | 원자 쓰기 · 1분 주기 · 계좌 칸 없음 · 정상 정지 때 지워짐이 이미 성립. 엔진 무편집 = 손절 경로 위험 0 |
| B. 보조 실행자(`AuxiliaryExecutor`)로 새 `engine-heartbeat.json` | 기각 | 같은 사실(「이 pid 가 1분 안에 살아 있었다」)을 두 파일이 말하면 생존의 정의가 둘이 되고 둘은 갈릴 수 있다. `engineRuntime`(엔진 조립) 편집 = High-risk 인접 FLM. 그리고 `obs.Heartbeat` 의 `Publisher` 를 파일 발행자로 채우는 길은 a133 구조 핀 ①(비시험 코드에서 `Publish(ctx, Notification) error` 를 가진 타입은 `*obs.Ntfy` 하나)을 깬다 |

마커가 증명하는 것과 못 하는 것:

- 증명: 엔진 **프로세스**가 마지막 갱신 시각에 살아 있었다. OOM · SIGKILL · 호스트 정지는 파일을 지우지 못하므로 시각이 멈춘다(**stale**).
- 정상 정지 · 방어적 종료(감독 루프가 스스로 반환 → `rt.Run` 반환 → defer `Release`)는 파일을 지운다(**missing**). 방어적 종료는 런타임이 이미
  critical 알림을 내므로(engine-safety 「엔진 런타임 수명주기」 ① 방어적 종료 계약) missing 의 알림은 「돌아야 할 엔진이 없음」을 보는 두 번째 눈이다.
- 알려진 잔재(구현 리뷰 P2-2 · P2-3 — 엔진 쪽 묘비(tombstone) 없이는 못 닫음, 범위 밖):
  - autostart OFF(콘솔 버튼 기동)에서 `rt.Run` **전**의 오류 반환(`engineRuntimeFactory` · `NewPositionPolicyCommandService` 실패)과
    주 goroutine 패닉도 defer `Release` 로 마커를 지운다 → IDLE 로 읽혀 알림 없음. 그 경로에는 런타임의 critical 알림도 없다.
  - `Hold` 의 첫 쓰기가 실패하면 핸들이 비활성이라 마커가 없는 채 엔진이 돈다 → autostart ON 이면 「missing」(엔진이 도는데 끊김 —
    시끄러운 쪽), OFF 면 IDLE(감시 안 됨). 갱신 실패(`_ = h.refresh(at)`, 디스크 가득)는 「stale」. 둘 다 크게 실패하는 쪽이므로 수용하고,
    알림 문구가 「마커 쓰기가 실패하는 중」 가능성도 말한다.
  - 호스트 일시 중지(suspend) 뒤 깨어나면 엔진의 단조 시계 ticker 는 최대 1분 늦게 갱신하므로 거짓 끊김 1건 + 복구 1건이 올 수 있다(노트북 호스트에서만).
- 못 함: 프로세스는 살아 있는데 루프가 멈춘 상태(교착). 마커 goroutine 은 루프와 독립이다. 이것은 런타임의 health ladder(`superviseHealth`)의
  몫이고, 보조 실행자 안 B 도 같은 이유로 못 잡는다(루프와 독립인 goroutine). 범위 밖으로 명시한다.

## D2. 판정표

감시 명령은 마커를 `os.Stat` 으로 직접 본다(경로 철자는 `enginelock.MarkerPath` 하나). `runlock.Fresh` 는 모든 stat 실패를 「없음」으로 접어서
권한 오류와 부재를 못 가르므로 쓰지 않는다 — 「볼 수 없음」은 「없음」이 아니다.

| 마커 | `engine.autostart` | 판정 | 이유 |
|---|---|---|---|
| 있음 · 나이 ≤ 기준 | 무관 | **UP** | 프로세스가 기준 안에 갱신함 |
| 있음 · 나이 > 기준 | 무관 | **DOWN(stale)** | 죽은 프로세스만 마커를 남기고 갱신을 멈춤. 수동 운용(autostart OFF)이어도 사고 |
| 있음 · 미래 시각 | 무관 | **UP** | 시계 불일치는 죽음이 아님 — `runlock.Fresh` 선례 |
| 없음 | ON | **DOWN(missing)** | autostart ON 은 「엔진이 돌아야 한다」는 사람의 승인(engine-safety 「승인된 엔진 자동 기동」) |
| 없음 | OFF | **IDLE** | 정상 정지 · 수동 운용. 감시 대상 아님 |
| stat 실패(부재 외) | 무관 | **오류 종료** | 판정 불가를 부재로 단정하지 않음 |

**복구(끊김 → 정상) 판정은 신선한 마커만으로 하지 않는다**(구현 리뷰 P2-1): 마커 본문에 `ready_at`(재시작 복구를 마친 엔진만 씀, a102)이
있어야 한다. 마커는 부팅 6단계(Recover 전)에 쓰이고 오류 종료 때 지워지므로, 재기동 루프의 몇 초짜리 마커나 Recover 에서 멈춘 엔진을
「복구」로 읽으면 거짓 안심 알림이 된다. 신선하지만 `ready_at` 이 없거나 본문을 해석할 수 없으면 끊김을 유지하고 알림을 보내지 않는다.
끊김 판정 쪽은 `ready_at` 을 보지 않는다(신선한 마커 = 살아 있는 프로세스) — 막 기동한 엔진을 죽었다고 읽지 않기 위해서다.
IDLE(autostart OFF · 마커 없음)로의 이동도 복구 알림 한 건(「감시 대상에서 벗어남」)이다.

기준(`--stale-after`) 기본 **3분** = 갱신 두 번 놓침. 감지 지연 최악 ≈ 기준 + 타이머 1분. `RefreshEvery`(1분) 이하는 거절한다(살아 있는 엔진을
죽었다고 읽음). 엔진 쪽 stale 창(5분, 콘솔 · autostart 용)과 따로인 이유: 그 창은 「유령 마커를 믿는 최대 시간」이고, 여기 기준은 「사람을 깨우기 전
기다릴 최대 시간」이라 질문이 다르다. 5분 창은 바꾸지 않는다.

## D3. 한 번만 알리기 — 상태 파일

상태는 저널 디렉터리의 `engine-heartbeat-watch.json`(감시자 소유, 0600, 임시 파일 + rename). 칸은 `down` · `reason` · `since` 뿐.

| 이전 상태 | 이번 판정 | 행동 | 상태 쓰기 |
|---|---|---|---|
| down 아님 | DOWN | critical 「끊김」 1건 | 전송 **성공 뒤** down 기록 |
| down | DOWN | 없음(중복 억제) | 없음 |
| down | UP · IDLE | 복구 알림 1건 | 전송 **성공 뒤** down 해제 |
| down 아님 | UP · IDLE | 없음 | 없음 |

- 전송 실패 → 상태를 옮기지 않고 오류 종료 → 다음 실행(1분 뒤)이 다시 보냄. 「보냈다고 기록했는데 안 나감」이 생길 수 없는 순서.
- 상태 기록 실패(전송은 성공) → 오류 종료. 다음 실행이 같은 알림을 한 번 더 보낼 수 있다 — 중복은 허용, 누락은 금지 방향.
- 읽을 수 없는(손상된) 상태 파일 → 「down 아님」으로 읽음. 같은 방향(중복 가능 · 억제 불가).
- 상태는 엔진 · 원장과 무관하다. 엔진은 이 파일을 읽지 않는다.

## D4. 전송 — 엔진과 같은 규칙, 같은 경계

- 전송 수단은 엔진 조립이 쓰는 `resolveNotificationPublisher`(internal/app/engine/notifications.go) 를 **그대로** 부르는 내보내기 입구
  `engine.NotificationPublisher` 로 얻는다(새 파일 `a134_notification_publisher.go`). 규칙(enabled · 파일 토픽 ⊕ `TOSSCTL_NTFY_TOPIC` · 토큰 env ·
  `Rejected`)의 사본은 없다. notifications.go 무편집.
- 결과 `*obs.Ntfy` 의 `Publish` 는 a133 전송 경계(`redactForTransport`)를 지난다. 알림 본문은 애초에 마커 시각 · 기준 · 판정 사유뿐이다.
- **전달 실측 `--send-test`**(구현 리뷰 P1-1): 판정은 상태가 바뀔 때만 보내므로 「수단은 있는데 안 닿음」(호스트에 토큰 없음 · 401 · 잘못된 토픽 ·
  DNS)은 첫 끊김 때에야 드러난다. 그래서 같은 수단으로 normal 시험 알림 한 건을 보내고 실패하면 비정상 종료하는 플래그를 두고, 설치 절차에서
  **단위와 같은 환경으로** 한 번 돌리는 것을 필수 단계로 둔다(tasks 게이트 뒤 절차 2). 판정 · 상태 파일과 무관하다. 매 분 시험 송신은 하지 않는다
  (채널 소음 — 박동이 음소거되는 길).
- 수단이 없으면(알림 OFF · 토픽 없음 · 설정 거절) **오류 종료 + 이유 문구**. 감시자가 알릴 수 없는 상태는 조용한 성공이 아니다 — systemd 단위가
  failed 로 보인다. 이 검사는 판정보다 먼저 한다(알림을 못 보내는 감시자는 매번 실패해야 함).
- 등급: 끊김 = `SeverityCritical`(ntfy 우선순위 5 — 방해 금지 통과, 실계좌가 보호 없이 있을 수 있음), 복구 = `SeverityNormal`.
  이벤트 타입은 새 값 `engine.liveness_lost` · `engine.liveness_restored` · `engine.liveness_watch_test`(enginewatch 패키지 상수). `internal/obs/event.go` 는 a135 와 겹칠 수 있어
  편집하지 않고, 값이 event.go 의 어떤 값과도 겹치지 않음을 시험이 직접 읽어 확인한다(a102 유일성 시험과 같은 방법).
- outbox 를 쓰지 않는다: 감시자는 엔진이 아니고 저널을 소유하지 않는다(엔진이 flock 으로 단일 writer). 내구성은 D3 의 「성공 뒤 기록」이 대신한다.
- 호스트에는 컨테이너의 `TOSSCTL_NTFY_TOKEN`/`TOSSCTL_NTFY_TOPIC` 이 없다. 보호된 토픽이면 사람이 단위 drop-in 의 `EnvironmentFile=` 로 넣는다(tasks 3.x).
  토픽이 환경 변수에만 있는 배포는 호스트에서 「토픽 없음」 오류로 드러난다(조용히 통과하지 않음).

## D5. 생존 파일 · 상태 파일 · 알림에 계좌 칸 없음

- `enginelock.Marker` 의 JSON 키 전부가 `redact.IsAccountKey` 가 아님을 시험이 리플렉션으로 고정한다(마커에 계좌 칸이 생기면 실패 — 이 파일은 이제
  엔진 밖 프로세스가 읽는 생존 신호이므로). enginelock 무편집, 시험만 enginewatch 패키지에 둔다.
- 상태 파일 칸(`down` · `reason` · `since`)과 알림 본문에도 계좌 키 · 계좌 문자열이 없음을 시험이 확인한다.

## D6. 감시자는 호스트 systemd user 단위

- `deploy/systemd/tossos-heartbeat.service`(oneshot) + `.timer`(매 분). `tossos-attest` 와 같은 모양: `%h/.local/bin/tossctl --config-dir %h/.config/tossctl`.
  호스트의 `~/.config/tossctl` 이 컨테이너의 `/var/lib/tossos/config`(엔진 저널 디렉터리)로 bind 된다(compose `TOSSOS_CONFIG_DIR`).
- 단위에 `TimeoutStartSec=50s` — 멈춘 실행이 다음 매 분 실행을 막지 않게(ntfy 전송 상한 10초).
- 「세션을 열지 않음」의 정확한 범위(구현 리뷰 P3-2): 루트 명령의 공통 전처리(`PersistentPreRunE`)가 세션 · 자격 증명 파일을 읽고 업데이트 확인을
  할 수 있는 것은 모든 tossctl 명령(tossos-attest 포함)과 같다. 이 명령은 그것을 쓰지 않고 브로커를 부르지 않는다.
- 컨테이너 안이 아닌 이유: 컨테이너 전체가 죽거나 멈춘 경우까지 같은 감시자가 본다. 엔진 · 콘솔 어느 쪽 코드에도 의존하지 않는다.
- 설치 · 활성화는 사람이 한다(이 change 는 파일만 둔다). user 단위가 로그아웃 뒤에도 돌려면 linger 가 필요하다 — `tossos-attest.timer` 가 이미
  돌고 있다면 충족된 상태다(사람이 확인).

## D7. 외부 dead-man URL — 설계만, 구현하지 않음

기계 전체가 죽으면 호스트 타이머도 죽으므로 이 감시자는 침묵한다. 그것을 잡으려면 외부 서비스(healthchecks 류)가 「감시자의 박동이 끊김」을 판정해야 한다.

- **구현하지 않는 이유**: (1) 새 외부 HTTP 송신 경로가 생기면 a133 전송 경계 밖의 송신이 되고 구조 핀 ⑤(obs 안 net/http 는 ntfy.go 하나)의 취지와
  어긋나 별도 리뷰가 필요하다. (2) 어떤 외부 서비스 · 계정을 쓸지는 사람의 선택이고 지금 정해진 것이 없다(YAGNI). (3) Manager 가 확인한 실제 공백 —
  콘솔은 건강한데 엔진 자식이 죽음 — 은 이 change 의 호스트 감시자가 닫는다.
- **코드 없이 붙이는 길**(사람 선택, 기본 OFF): 단위 drop-in 에 `ExecStartPost=/usr/bin/curl -fsS -m 10 --retry 2 <ping-url>` 를 넣으면 감시자 실행이
  성공할 때마다 외부 서비스가 박동을 받는다. 감시자는 판정 · 전송을 마쳤을 때만 성공 종료하므로 그 박동은 「호스트 감시자가 살아서 일했다」를 뜻한다.
  URL 은 계좌와 무관한 비밀이라 저장소 · 설정 파일에 두지 않는다.

## D8. `obs.Heartbeat` — 주석만 정정, 삭제하지 않음

- 파일 · 타입 주석에서 「ntfy 가 스스로 경보」를 지우고, 생산 배선 0 · 실제 생존 경로(마커 + `tossctl engine heartbeat-check`)를 적는다. 함수 본문 ·
  함수 doc 무편집(게이트가 계산하는 수정 함수 0).
- 삭제하지 않는 이유: 생산 사용이 0 이라 남겨도 런타임 영향이 없고, 지우면 base 함수 둘(`Beat` · `Run`)의 `revision: base` 증거와 기존 시험 3 의 삭제가
  필요하다. 삭제는 후속 정리 후보로 완료 보고에 남긴다.

## D9. spec 과의 관계

새 요구 「엔진의 침묵은 엔진 밖에서 판정된다」를 `engine-safety` 에 **추가**한다. 이 요구가 「등급화된 알림」의 heartbeat 문장이 말하는 「수신 측 경보」의
이행이며 그 경보는 ntfy 가 아니라 엔진 밖 감시자가 ntfy 로 보낸다는 것을 요구 본문에 적는다. 「등급화된 알림」 자체를 MODIFIED 로 다시 쓰지 않는 이유:
그 요구는 180줄이고 critical 목록에 「자격증명 만료 임박」이 있어 병행 change(a135, attestation 만료 경고)와 같은 요구를 동시에 다시 쓰게 될 수 있다.
그 문장의 괄호 정정은 아카이브 때 Manager 가 한쪽 delta 로 접는다(완료 보고에 남김).

## D10. 위험 · 비례 원칙

- 주문 · 손절 · 익절 · 사이징 · Guardian · 원장 · 대사 · 인증 · 체결 경로: **무편집**. 엔진 프로세스 코드 무편집.
- 편집하는 기존 함수: `newEngineCmd`(cmd/tossctl/engine.go, 분기 0, `AddCommand` 인자 하나 추가). 게이트가 계산하는 수정 함수이므로 FLM 번들을
  만든다(`analysis/function-logic/cmd-tossctl--newenginecmd/`). 그 밖은 새 파일 · 새 leaf.
- 비례 원칙: 알림 · 관측 경로이고 게이트 판정을 바꾸지 않으므로 변이 원장 · 다중 적대 리뷰는 의무가 아니다. RED/GREEN 시험 + lint + 경량 독립 리뷰.
- 시험은 실 ntfy 로 보내지 않는다(`Publisher` 주입, 명령 시험은 httptest 서버). 시험은 `t.TempDir()` 저널 디렉터리만 쓴다.

## 배포

엔진 이미지 교체 불요. 호스트 `~/.local/bin/tossctl` 을 이 change 를 담은 빌드로 바꾼 뒤 사람이 단위를 설치한다(tasks 3.x). 순서 의존 없음 —
감시자는 기존 마커만 읽는다.
