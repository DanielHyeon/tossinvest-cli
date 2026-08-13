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

## §4. Fix 라운드 결과 (2026-08-14)

**T1-fix (커밋 39ec4feb·19744358·bec6e133·217c8813)** — §6a 전부 이행. RED 5건(생산자가
만드는 영구 거부) 선행 커밋 실재. `listenPrivateSocket`(임시 bind→chmod→rename,
`SetUnlinkOnClose(false)`)·descriptor stage+rename·staging 잔재 회수·`perm&0o077==0`
완화·형식/내용 분리·`Dial` connect probe·`Close`의 listener 소유. 패키지 테스트 23→35.
뮤테이션 §B 15건 중 14 사망 + §A 7건 재적용 전부 재사망(**회수 함수를 다시 썼으므로 옛
관측을 새 코드에 다시 돌렸다** — 좋은 규율). 생존 1(M17, rmdir ENOENT 경합 분기 —
seam 없이 결정 불가, 명시 선언). M7 해석 정정 기록. 이행 편차 1건(0o644 케이스 비이동)은
design 정합으로 Manager 수용(tasks 6.5 주석).

**T2-fix (커밋 d8b27021·aecc03e0·e3b30841)** — §6b 전부 이행. **이번엔 RED 커밋이
히스토리에 실재**(A2 F5 교훈 이행). outbox 철회 + 핀 2종(미전달 행 0 / journal 재개방
후 entry gate 비잠금 — 대조군 EnqueueAlert로 harness의 측정 능력까지 증명). stat 콘솔
패리티·reader의 strategy Read 실패 dormant 흡수(NOT_CONFIGURED과 RUNTIME_UNAVAILABLE을
접지 않음). 뮤테이션 12/12 — M2b/M3b는 "부르지 않는다"가 아니라 "미전달 행이 없다"를
재서 제3의 경로도 잡는다. 원장 정정 3건(execgw 선례·비-unix 항목 삭제·ready 테스트 개명).

**교차 검증**: T2-fix가 의도된 RED로 남긴 `TestASocketFileWithNoOwnerDegradesTheDaemon`
(S3에서 httpapi 강등)이 T1-fix의 Dial probe 병합만으로 무접촉 GREEN — 두 소유 경계에
걸친 계약(D4-2)이 실제로 맞물렸다는 실측.

**병합 후 검증(Manager)**: `strategyprojectionrpc`+`cmd/tossctl` -race 625 passed,
`check_analysis.py` evidence complete(오류 0).

## §5. 선언된 생략 (not-applicable)

- T1-fix: `Start`/`writeDescriptor` 실패 분기 12+2건 무테스트 — 파일시스템 오류 주입
  seam이 발행 경로에 새 실패 모드를 더한다. 성질 핀 3개로 대체 측정.
- T1-fix: chown 불가(비root)·`Lstat` symlink mode 두 절은 디스크 상태로 단독 실패 불가 —
  원장 §B5.
- T2-fix: `httpAPIReader.Snapshot` B1~B7(기존 fail-closed 읽기) 미핀 — fixture가 7개
  전부 성공해야 한다는 대조 조건으로 대체.
- staging socket 이름의 sun_path 108 한도(실측 78바이트, 여유) — 길이 가드 미도입, 관찰만
  기록. 극단적으로 긴 `$TOSSOS_DATA_DIR` 배포는 한도에 걸릴 수 있다.
- tasks 2.5는 **미완**(실패-불가 핀) — a109가 소유.

## §6. gstack 독립 리뷰 (2026-08-14, 7패스)

패스: 스페셜리스트 4(testing·maintainability·security·performance) + Red Team +
Claude 적대(3라운드째) + Codex 적대·구조화(모델 교차). Codex 구조화 GATE: **FAIL**(P1 1)
→ Fix-First(T3-fix) 후 재검증. Scope check: CLEAN(선언된 a109 등록 포함, drift 없음).

### 수렴 지도

| 발견 | 소스 | 판정 |
| --- | --- | --- |
| 동기 Notify가 부팅을 최대 10s 블록(rt.Run 전, ntfy 배선 시) — "관측이 보호를 막지 않는다" 계약을 fix 라운드 코드가 위반 | **performance + Claude 적대 + Codex 구조화 (3소스)** | **T3-fix A1 (P1)** |
| dial-실패 강등이 NOT_CONFIGURED로 렌더 — 엔진 장애가 기능-미사용과 같은 신호 | **security + maintainability + Claude 적대 (3소스)** | **T3-fix A2** — sentinel reader로 RUNTIME_UNAVAILABLE |
| staging 이름이 최종 이름보다 길어 sun_path 직전 배포에서 매 부팅 소실 — T1-fix의 "관찰"을 수정으로 승격 | **Codex 구조화 + Red Team (2소스)** | **T3-fix B1** |
| probe 절 미핀 + owner-write 없는 socket EACCES→생존 오판(영구 거부 잔류) | **testing + Codex 적대 (2소스)** | **T3-fix B2** |
| spec delta "가동 중 사망 → dormant"가 코드·핀(RUNTIME_UNAVAILABLE)과 모순 — archive 시 정본에 거짓 문장 | Red Team (CRITICAL) | **Manager 즉시 수정** (spec delta) |
| 롤백이 a108을 가로지르면 구 바이너리가 새 잔재를 영구 거부 — 사고 재생산 | Claude 적대 | **Manager D5-3** + tasks 5.2 + a109 2b.3 |
| .staging 디렉터리가 회수를 영구 wedge / 발행 실패 잔재 당회 잔존 / 무경고 강등(engineJournalDir) / token 스코프 / stale RED 주석 / line-range 인용 rot | 단일 소스 P3들 | **T3-fix B3·B4·A3·A4·A5** |

### 기각·이관 (사유 명시)

- **소비자측 lazy 재-dial**(Red Team): 회복 부팅마다 화면 소실이 기본값이 되는 지적은
  타당하나 reader에 재시도 상태를 넣는 설계라 **a109 2b.1로 이관**. A2의 sentinel이
  소실을 최소한 "보이게"는 만든다(unavailable ≠ 미사용).
- **same-UID 위조 endpoint·pathname TOCTOU/dirfd**(Codex 적대 1·4, security 2):
  전제가 journal·config를 이미 장악한 same-uid 공격자다 — 신뢰 앵커 밖. 하드닝 후보로
  기록만.
- **콘솔·httpapi 강등 블록 공용 helper**(maintainability): 콘솔은 a108 무접촉
  원칙 — 정리 change 후보로 기록.
- **wedged-engine 5s 스냅샷 스톨**(performance): 변경 전과 동일한 상한, 죽음의 일반
  형태(ECONNREFUSED)는 µs — 선택 개선으로 기록.
- **Codex 검증 한계**: 샌드박스 /tmp 읽기 전용으로 go test 미실행(발견은 정적 추적) —
  Claude 적대·Red Team이 같은 트리에서 테스트 실행으로 보완.

### Red Team이 닫은 각도

obs Normal 경로가 outbox에 닿지 않음을 notifier 코드로 재검증(A2 P1의 entry-gate 위험
종결 확인), 형제 endpoint 잔존 결함은 D5-2 스코프 선언과 일치, datagram 스쿼팅·ENOTDIR은
안전 방향(Claude 적대 실측).
