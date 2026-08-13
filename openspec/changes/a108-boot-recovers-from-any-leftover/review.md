# a108 review — 적대 리뷰 라운드와 판정

파이프라인: Manager(Fable) openspec → T1·T2(Opus) 병렬 구현 → A1·A2(Opus) 적대 리뷰 →
Manager 판정(design D1-2~D5-2) → Fix 라운드(tasks §6) → gstack → Manager 독립 검증.

## §1. A1 (T1: `internal/strategyprojectionrpc`) — FIX-FIRST

뮤테이션 독립 재현 2건(M4·M5c)이 원장과 정확히 일치. RED 진위는 커밋 순서(테스트 커밋이
GREEN보다 81초 선행, 구현 파일 0)와 detached 체크아웃 재실행으로 확증 — **A2의 T2 RED와
달리 역사적 증거가 실재한다.**

| # | 발견 | 판정 |
| --- | --- | --- |
| F1 P1 | pre-chmod socket(umask 077→0700) 잔재가 "stale socket is unsafe"로 영구 거부 — 버그가 보안 핀(M7)으로 봉인됨. 실측 3/3 | **수용 → 6.1/6.2/6.5.** 생산자 stage+rename + 검증-사망 한정 perm 완화(D1-2) |
| F2 P1 | 0바이트·잘린 descriptor 영구 거부. D2 이후 내용 파싱은 판정에 안 쓰이는 거부-전용 게이트. 실측 3/3 | **수용 → 6.1/6.2.** descriptor stage+rename + 내용 파싱 실패=사망 입증 시 회수(D1-2) |
| F3 P1 | spec delta의 일반화된 SHALL("엔진이 소유한 런타임 endpoint")이 거짓 — 형제 셋도 같은 병(실측: policy runtime·alert control pre-chmod 영구 거부, alert는 산-주인 socket 탈취) | **수용 → spec 좁힘 + a109 등록(D5-2).** 거짓 SHALL을 정본에 넣지 않는다 |
| F4 P1 | tasks 2.5 핀은 관용이 자명한 모양만 만들어 실패 불가 — 정지 조건 무력화 | **수용.** 2.5는 미완으로 남기고 사고급 모양 핀은 a109 tasks가 소유 |
| F5 P2 | 늦은 unlink가 후계자 socket을 경로 기준으로 지움(300라운드 중 3회 재현). 오늘의 방어는 flock이 아니라 "goroutine이 프로세스와 함께 죽는다"뿐 | **수용 → 6.3.** SetUnlinkOnClose(false)+명시 제거+in-process 재시도 금지(D2-2) |
| F6 P3 | 절 단위 미핀(symlink·uid·SameFile), rmdir ENOENT 비대칭, 새-주인 경합의 flock 미인용 | **수용 → 6.6** |
| F7 | probe TOCTOU·자기 오판·좀비 매달림(DialTimeout 200ms 1회)·backlog 포화(안전 방향) — 결함 못 찾음 | 기록 |
| F8 | processAlive 삭제의 손실 — 못 찾음. 비root(uid 10001)에서 EPERM 분기는 거짓 양성 증폭기였다(사고 PID 16이 그 자리) | 기록 — D2 논증 강화 |

## §2. A2 (T2: `cmd/tossctl`) — FIX-FIRST

뮤테이션 독립 재현 2건(M1·M8)이 원장과 정확히 일치. 회귀 -race 전체 통과.

| # | 발견 | 판정 |
| --- | --- | --- |
| F1 P1 | 강등의 durable critical 행이 **다음 부팅의 entry gate를 잠근다** — `UndeliveredCount` Type 무필터 + `restoreAlertEntryLatch` + publisher 미설정이면 영원히 PENDING. obs 교리("화면 오탈자가 실계좌 매매를 멈출 수 없어야") 위반. execgw 선례는 등급표 등재+의도적 Block이라 선례 아님(오독) | **수용 → 6.7.** outbox 철회(D3-2, 결정 반전 기록). 방향이 보수적(진입 차단)이라 안전 불변식 위반은 아니나 이 change의 논지·교리 위반 |
| F2 P1 | "미해소 유지" 미구현 — transport 살아 있으면 ~2s 뒤 소멸. 확인 테스트는 전달 루프 없는 harness에서 PENDING을 봄(미측정) | **수용 → D3-2.** 문서를 코드가 지킬 수 있는 계약으로 재작성, 약화를 선언 |
| F3 P2 | S3(socket 잔존 사망)에서 Dial이 연결하지 않아 강등 미발동 + 집계 스냅샷 전체 실패(probe 실측) — 전원 단절의 기본 모양 | **수용 → 6.4(T1)/6.8(T2).** Dial connect probe + reader dormant 흡수(D4-2) |
| F4 P2 | 비-NotExist stat fatal이 콘솔과 갈라져 ENOTDIR 볼륨 오배치에서 crash loop 재생산 | **수용 → 6.8.** 콘솔 패리티(D4-2, 원 D4 결정 철회) |
| F5 P3 | T2 RED는 git 히스토리에 없음(한 커밋) — 재구성 증거. 실질은 A2의 M1 재현이 독립 확증 | **수용 → 6.9 기록.** T1의 RED-선행 커밋과 대비하여 관행 교훈으로 남김 |
| F6 P3 | ready 시점 테스트는 seam 존재만 측정(실제 after-Recover 순서는 a102 테스트가 소유) — 이름 과장. 공유 핸들 함정은 아님 | **수용 → 6.9** 이름/주석 정직화 |
| F7 P3 | dedup 토큰 PID 부재·fallback 항상-유일 | outbox 철회로 기계 자체 제거(6.7) — 소멸 |
| F8 P3 | obs 등급표 분열 함정(미래에 Notifier로 흘리면 조용히 Normal) | D3-2가 **의도로 확정**(Normal이 설계다). 설계 주석에 명시 |
| F9 | **비-unix 위험은 존재하지 않는다** — `flock_other.go:16` ErrLockUnsupported가 1단계에서 기동 차단, 7단계 도달 불가. 원장 정정 지시 | **수용 → 6.9** 원장에서 삭제 |
| F10 | nil 라우터 panic 없음(DormantSnapshot 경로 추적, typed-nil 함정 없음 — dormant 200), 자원 누수 없음, interlock 순서 불변은 진짜 측정(M5), LIVE 접촉 없음(testenv.Isolate 확인) | 기록 |

## §3. Manager 판정 요약

- 두 P1 군의 공통 기전: **T1 쪽은 "발행이 회수만큼 전체적이지 않았다", T2 쪽은 "보고가
  게이트에 배선된 rail을 탔다."** 각각 D1-2(stage+rename)와 D3-2(outbox 철회)로 계약을
  고쳤고, D3-2는 원 D3의 명시적 **반전**이다 — 반전 사유를 설계에 남겼다.
- spec delta는 좁혀서 참으로 만들었다(형제 셋 → a109). 거짓 SHALL의 sdd-sync 유입이
  A1 F3의 실질 위험이었다.
- 2.5는 **미완**으로 남는다(핀이 실패 불가였음) — a109가 사고급 모양 핀을 소유한다.
- 사고 당시 상태의 즉시 복구(잔재 수동 이동)는 이 리뷰 시점에도 **사람 실행 대기**다.

## §4. Fix 라운드 결과

(6a·6b 완료 후 기록)
