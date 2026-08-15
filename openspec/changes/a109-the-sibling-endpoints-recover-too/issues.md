# a109 구현 중 발견 — 분류와 처리

Teammate가 구현 중 발견한 계약 결함을 기록한다. 분류는 ① blocking(안전·동작 모순 —
구현 중단 후 보고) ② safe local(스펙 의도가 명백한 사소한 보완 — 구현하며 사후 기록)
③ editorial(즉시 수정) 셋이다.

**blocking 0건.** (있으면 여기 맨 위에 적고 구현을 멈춘다.)

---

## T1 — ③ editorial

### E1. tasks §1a의 "자기 표면 9개"는 슬러그 수를 잘못 셌다 (2026-08-15)

tasks.md §1a는 T1의 FLM 표면을 "9개: prepare/stat/validate×2계열, Start×3, Close×3"으로
적었다. 실제 scaffold 된 T1 슬러그는 **17개**다:

- `internal/positionpolicyrpc` 8 — PreparePrivateSocket · PrepareRuntimeSocket ·
  statPrivateSocket · ValidatePrivateSocket · ValidateRuntimeSocket ·
  ValidatePrivateControlDirectory · ValidateRuntimeControlDirectory · validatePrivateDirectory
- `internal/app/engine` 9 — Start×3 · Close×3 · descriptor 발행×3

§1a 문구를 17로 고치고 17개를 전부 완성했다. (9는 "prepare 2 + stat 1 + validate 소켓 2 +
Start 3 + Close 3 = 11"과도 맞지 않아 어느 셈으로도 성립하지 않는다.)

### E2. `validatePrivateDirectory` scaffold의 Source 경로가 실제 정의 파일과 다르다

`analysis/function-logic/internal-positionpolicyrpc--validateprivatedirectory/`의
`function-logic-map.md` scaffold는 Source를 `internal/positionpolicyrpc/private_fs_unix.go`로
적었으나, 같은 디렉터리의 `ast.json`은 `internal/positionpolicyrpc/client.go`를 가리킨다
(실제 정의도 client.go:140). `check_analysis.py`가 ast.json의 경로를 정본으로 요구하므로
맵을 client.go로 정정했다. `risk-pattern-report.md`는 처음부터 client.go로 옳게 적혀 있었다.

---

## T1 — ② safe local

### S1. `PreparePrivateSocket`/`PrepareRuntimeSocket`을 **삭제하지 않고 남긴다**

design D2는 두 함수의 동작(probe 없는 unlink)을 거부로 바꾼다고만 하고, 함수의 존폐를
말하지 않는다. a109 GREEN 이후 두 함수는 기동 경로에서 호출되지 않는다.

**남기기로 했다.** 근거: ① `PrepareRuntimeSocket`에는 기존 테스트
(`TestPrepareRuntimeSocketNeverDeletesNonSocket`)가 있어 삭제는 기존 테스트 삭제를 동반한다.
② 둘 다 공개 API이고 `_other.go` 플랫폼 스텁까지 짝이 있다. ③ 삭제하면 FLM이
`revision: base`로 바뀌어 게이트 형태가 달라진다.

대신 **본문은 한 줄도 고치지 않는다** — 고치면 공유 helper(`statPrivateSocket`)를 통해
완화가 클라이언트 검증으로 번진다(freeze P1-3). 남은 위험: 다음 사람이 이 함수를 다시
기동 경로에 배선할 수 있다. 그 위험은 §1.2의 "산 주인 거부" 핀이 잡는다 — 배선을 되돌리면
그 테스트가 죽는다.

### S2. 회수는 잔재 descriptor의 **모양**을 검사한다(내용은 관용)

design D2는 "descriptor는 rename이 덮는다 / 0바이트·잘린 잔재는 관용"만 말하고, 최종
descriptor 이름 자리의 **모양**을 어떻게 다룰지는 말하지 않는다(P1-7②가 "descriptor 자리의
이물은 회수하지 않는다"고만 적었다).

`ValidatePrivateFile`(0600 정규 파일·우리 uid·nlink 1)을 쓰기로 했다. 근거: ① a108의
회수도 `openVerifiedDescriptor`로 같은 형식을 요구한다(이식 원형과 일관). ② 우리 발행
경로는 언제나 chmod 0600 뒤 rename 하므로 이 요구가 자기 잔재를 거부할 수 없다.
③ "이물은 회수하지 않는다"의 구현이 곧 거부다(지우지 않고 기동을 실패시키면 D3 강등이
받는다). 내용은 **파싱하지 않는다** — 0바이트 관용은 그대로다.

뮤테이션 M14가 이 절을 지웠을 때 1차에서는 아무 테스트도 죽지 않았고,
`TestReclaimRefusesADescriptorOfTheWrongShape`를 추가해 죽였다.

### S3. `SweepPrivateStagingLeftovers`는 오류를 돌려주지 않는다

design D2a는 "낯선 엔트리는 오늘처럼 무시한다"까지만 말한다. 위생이 **실패를 보고할지**는
정하지 않았다. 오류 없는 형태로 만들었다 — 이 endpoint에 새 실패 경로를 만들지 않는다는
D2a의 이유(이물 하나가 격리 해제 표면을 매 부팅 지운다)가 "디렉터리를 못 읽었다"에도 같은
힘으로 적용되기 때문이다. 못 치운 잔재는 다음 부팅이 다시 시도한다.

### S4. 회수 후 control 디렉터리는 **언제나 이번 기동이 만든 것**이다

a108 모형을 이식하면 회수가 디렉터리째 지우고 Start가 다시 만든다. 그 결과 두 socket
transport의 `createdControlDir` 플래그가 의미를 잃어 제거했다(실패 정리가 조건 없이
디렉터리를 지운다). command endpoint는 회수를 넣지 않았으므로 그 플래그를 그대로 뒀다.

이것은 동작 변화다: 편집 전에는 "남이 만든 디렉터리는 실패해도 남긴다"였고, 편집 후에는
"회수를 통과한 디렉터리만 존재하므로 지워도 우리 것"이다. 회수가 낯선 엔트리를 거부하는
것이 그 전제이고, 그 거부를 §1.3 핀과 뮤테이션 M1b가 지킨다.

---

## T1 §1-fix — A1 적대 리뷰 판결 구현 중 남긴 것 (2026-08-15)

### I1. a108 원형(`projectionSocketAccepts`)도 같은 owner-write 추정을 가지고 있다

§1-fix F1은 `privateSocketAccepts`의 owner-write 사망 추정을 chmod-then-probe로 바꿨다.
그 절은 a108 `internal/strategyprojectionrpc/transport_unix.go`의 `projectionSocketAccepts`에서
이식한 것이고, **원형에는 그대로 남아 있다.**

**후속 change 등록 후보.** a108 확정 코드는 이 change의 표면이 아니다(pre-edit-gate T1의
선언된 무변경). 병의 모양은 동일하다 — 쓰기 비트가 깎인 **수락 중인** projection socket을
죽었다고 읽고 지운다. 다만 노출은 형제보다 좁다: 그 endpoint의 최종 이름은 이미 D1 의례를
지나 0600으로만 나타나므로, 그 상태에 이르려면 외부에서 chmod가 일어나야 한다.

수정은 기계적이다(chmod 0600 → dial, owner-write 절 삭제). 같은 뮤테이션(F1-N1의 원형판)이
그것을 지킬 수 있다.

### I2. `SweepPrivateStagingLeftovers`에는 F5의 probe가 없다 — 범위를 주석에 적었다

§1-fix F5는 "수락 중인 socket은 이름과 무관하게 unlink하지 않는다"를 **회수**
(`ReclaimStalePrivateEndpoint`)에 통일했다. 위생(`SweepPrivateStagingLeftovers`)에는
그 probe를 넣지 않았다.

**근거:** 이 함수를 쓰는 유일한 endpoint(policy command)는 loopback TCP라 socket을 발행하지
않는다 — 자기 접두의 socket을 만들 수 있는 경로가 없다. 그리고 `private_staging.go`는
플랫폼 무관 파일이라 probe(unix 전용)를 부르려면 `_other.go` 스텁 짝을 새로 들여야 하는데,
그것은 Manager 판결의 범위 밖이다("임의 확장 금지").

**남은 위험:** 다음 사람이 socket을 발행하는 endpoint에 이 위생을 배선하면 그 방어가 없다.
함수 doc에 그 조건을 명시했다("socket을 발행하는 endpoint에서 쓰려면 probe를 먼저 들여야
한다"). 코드가 아니라 문서로 막은 것이므로 **약한 방어**다.

### I3. 원장 §B에 남은 M26(Close의 listener 직접 닫기)은 잠정 생존이다

M22("경합이라 못 죽인다")가 A1에게 반증된 이상, 같은 사유를 든 M26도 그 지위를 잃는다.
지금 원장은 M26을 "경합(잠정)"으로 적고 있다 — "죽지 않는다"가 아니라 "내가 쓴 방법으로는
안 죽었다"는 뜻이다.

**후속 후보.** `Shutdown`이 대개 listener를 닫아 주므로 "누가 닫았는가"를 파일 관측으로
가르는 방법을 아직 찾지 못했다. M22를 죽인 A1의 수법(정리 대상 자리에 관측용 파일을 두고
그 생사를 본다)이 여기에도 통할 가능성이 있다.

### I4. P2-F(산 주인 + 이물 공존의 오귀속)는 잔존 수용이다

A1 P2-F: 산 주인이 살아 있는 디렉터리에 이물이 함께 있으면, 회수는 열거 루프에서 먼저
만나는 **이물**을 이유로 거부하므로 운영자에게는 "이물을 치워라"만 보인다 — 실제로는
주인이 살아 있는 것이 더 근본적인 원인인데도.

**수용한다.** §1-fix F6으로 거부 오류가 위반 엔트리를 지목하므로, 안내는 이제 **실행
가능하다**: 운영자가 그 이물을 치우고 재시작하면 다음 거부가 "the endpoint owner is still
alive"로 바뀐다. 두 원인을 한 번에 보여 주려면 열거를 끝까지 돌고 오류를 모아야 하는데,
그것은 "낯선 엔트리를 만나면 아무것도 건드리지 않고 즉시 멈춘다"는 D2의 성질을 약화한다.
두 번의 부팅으로 두 원인이 순서대로 보이는 쪽을 택했다.
