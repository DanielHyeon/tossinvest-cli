# a109 리뷰 기록

## §0 proposal-freeze (2026-08-15)

보이스 구성: Manager 4관점(CEO/Eng/Security/QA) + 적대적 Eng 리뷰어(Opus, 별도 컨텍스트,
반증 지향). High-risk change이므로 적대 Eng 포함은 WORKFLOW 리뷰 게이트의 요구다.

### Manager 4관점 패스

**CEO/Product** — 범위는 전부 추적 가능하다: 세 endpoint의 병은 a108 review §1 F3·F4
실측, 2b 항목 셋은 a108 gstack 이관 명시분. 무근거 확장 없음(descriptor fold·콘솔 화면
신설을 선언된 생략으로 배제). 닫는 위험은 사고 재발형(잔재 → 영구 기동 루프 → 장중
손절 없음)의 잔여 표면 전부다.

**Eng** — ① T1/T2 표면 비겹침 확인: T1=internal/*, T2=cmd/tossctl/* + 각자 자기 패키지
테스트. a108 2.5 핀 파일(internal/app/engine)은 T1 소유. ② T2의 seam은
cli_testseams.go 관례로 T1 완료와 독립 — a108에서 같은 방식이 성립했다.
③ 낯선 엔트리 거부 + D3 강등의 조합은 "이물 → 그 표면만 매 부팅 강등·보고, 엔진
생존" — 사고 모양(전체 정지)을 만들지 않으면서 don't-delete-foreign을 유지한다.

**Security** — ① perm 완화(&0o077==0)와 owner-write 사망 판정의 이식 적법성: 형제도
D1 이후 최종 이름은 chmod 0600을 지난 뒤에만 나타나므로 a108의 발행-수명주기 논증이
그대로 성립. ② 탈취→거부는 보안 강화 방향(산 socket unlink 금지). ③ 강등은 표면
부재이지 접근 확대가 아니다 — 어떤 토큰·경로도 넓어지지 않는다. ④ 회수의 소유
uid·symlink·nlink 검증은 완화하지 않는다.

**QA** — ① pre-chmod 핀은 umask 조작(프로세스 전역, 병렬 테스트 오염) 대신 명시적
chmod 0700로 만들 것 — Teammate 지침에 포함. ② 재-dial rate limit은 테스트 주입
가능해야 한다(spec은 수치를 고정하지 않고 design 기본값 30s만 둔다). ③ 뮤테이션
원장은 a108 규율(생존 뮤테이션은 측정-불가 사유와 함께 선언) 유지.

### 적대적 Eng 리뷰 (별도 컨텍스트, Opus) — P0 4 · P1 9 · P2 7

요약(전체 원문은 리뷰어 보고, 근거 file:line은 전부 재검증 가능):

- **P0-1** 재부착 wrapper가 nil·sentinel만 감싸면 live 부착 후 엔진 재시작(새 socket·
  새 토큰)이 재부착 대상이 아니다 — spec 시나리오 2가 기전 없는 SHALL.
- **P0-2** `Dial`은 본문에 200ms connect probe를 포함한다(transport_unix.go:415·:29) —
  Snapshot 경로 동기 Dial과 "요청 비차단" SHALL NOT은 양립 불가.
- **P0-3** 강등 시 소비자 메시지가 원인을 오귀속한다: alerts CLI "엔진이 없다,
  engine run을 살려라"(engine_alerts_client_unix.go:33–35), 격리 해제 "control plane이
  제공하지 않는다"(exit_quarantine.go:227–229). 콘솔에 alert control 표면 자체가 없어
  (grep 0건) a108 D3-2의 "세 번째 표면" 전제가 형제에는 성립하지 않는다.
- **P0-4** policy command 강등 논증 오류: 격리 유지 = **손절 포함 미판정** 유지
  (exitloop.go:560–568, exitpolicy/recovery.go:33–46)이고 운영자 해제가 유일한 장중
  복귀 경로. "엔진 재시작" 회복 수단은 결정적 원인 앞에서 무효(D3b와 자기모순).
- **P1-1** "legacy staging"은 현행 이름(fold 생략이므로 a109도 계속 만든다).
  **P1-2** spec 시나리오 4(낯섦→거부)와 D2a(command 무시)의 모순. **P1-3**
  `statPrivateSocket` 공유로 완화가 클라이언트 검증에 누출될 수 있음. **P1-4** D4에
  필요한 파일 2개 누락(httpapi_reader.go:565 nil 검사,
  internal/httpapi/strategy_runtime.go:18) — wrapper non-nil이 dormant→unavailable
  화면 회귀를 만든다. **P1-5** "cli_testseams.go 관례"는 T1 패키지 — T2 독립성 근거
  오류(옳은 관례는 engine.go:416 패키지 var seam). **P1-6** 설계가 분기를 주장하는데
  AST 없는 함수 7개(descriptor 발행 3·보고 1·디렉터리 검증 3). **P1-7** "모든 부분
  상태" SHALL이 거짓이 되는 상태 4가지(dir perm 변형·descriptor 자리 이물·staging
  소유 미검증·Close once 비재시도). **P1-8** hex7은 EncodeToString으로 불가(짝수만) —
  절단 필요. **P1-9** proposal fatal 줄번호 오류(:294는 a108이 이미 고친 자리).
- **P2**: 103 vs 실측 107(측정 포함), 저장소 내 선례
  positionPolicyRuntimeDescriptorReader 미인용, D5b 코드 근거 병기, 재부착 실패 로그
  스팸, 이름-집합 완전성 테스트, publishPrivateDescriptor 재검증 비대칭, 콘솔
  lifecycle client 동병 미치료.
- **공격 후 생존**: staging≤최종 basename 충분조건(같은 디렉터리), pre-chmod 진단,
  탈취 기전 공통성, owner-write 사망 판정 이식 적법성, command TCP probe 불요,
  AlertControlServer listener 필드 부재 확인, defer 순서 오늘 참(AST), 13개 AST
  신선도(sha256 13/13), validate·build 통과.

### 판결 (2026-08-15, Manager)

**P0 4건 전부 수용, 계약 수정 완료:**

| 발견 | 수정 |
| --- | --- |
| P0-1 | D4 개정: wrapper는 live 포함 전체를 감싼다, 트리거 = 부재 또는 직전 Read 실패. http-api-service 요구·시나리오 2 재작성 |
| P0-2 | D4 개정: 시도는 백그라운드 single-flight, 요청 goroutine의 dial·probe를 SHALL NOT으로 명문화 |
| P0-3 | D3a-2 신설: 두 소비자 메시지 정직화(문구만, 새 상태 없음) + engine-safety 요구·시나리오 추가. T2 표면에 두 파일 추가 |
| P0-4 | D3 표 재작성: 기준선을 fatal로 명시, policy command 칸을 "격리된 포지션의 무보호 유지, fatal 대비 개선으로만 성립"으로 정정, 회복 수단을 "원인 제거 후 재시작"으로 교체 |

**P1 9건 전부 수용:** P1-1 라벨 정정(현행·구버전 공통), P1-2 시나리오 4를 socket
endpoint로 좁힘 + D2a에 무시-유지 결정 명문화(새 거부 경로를 만들지 않는다), P1-3
"완화는 회수 전용" 명문화 + 뮤테이션 원장 항목, P1-4 T2 표면에 2파일 추가 + wrapper
상태 노출 결정, P1-5 seam 명칭 정정, P1-6 AST 7개 추가 생성(총 20), P1-7 SHALL을
"control 디렉터리 안 파일의 부분 상태"로 좁히고 나머지 상태는 D3 강등이 받는 것으로
설계에 열거, P1-8 hex 절단 명시(엔트로피 28비트), P1-9 줄번호 정정.

**P2 7건:** P2-1 수용(관계식만 테스트, 103은 각주), P2-2 수용(선례 인용 + 선택 이유),
P2-3 수용(코드 근거 병기), P2-4 수용(상태 전이 1회 보고), P2-5 수용(이름-집합 완전성
테스트), P2-6 수용(선언된 생략에 기록), P2-7 수용(선언된 생략 + 후속 change 등록 후보).

**잔존 리스크(수용된 채 기록):** 운영 control 디렉터리의 낯선 엔트리 현존 여부는 이
세션에서 확인 불가(접촉 금지) — 최악의 경우가 "해당 표면 강등 + 보고"로 격하된 것이
이 change의 방어다. 배포 후 엔진 로그의 강등 보고 확인을 배포 절차에 남긴다.

freeze 종결 — T1/T2 착수 허가.

## §1 A1 적대 리뷰 (T1 — transport 회수, 2026-08-16)

구성: Opus, 별도 컨텍스트, T1 워크트리(원복 검증 — porcelain 빈 출력 + production
sha256 6/6 + MUTANT 잔재 0건). 재현: 뮤테이션 7건 원장 대조(M5·M9·M19·M27 일치,
**M1a·M22·M23 반증**) + 신규 뮤테이션 A1-N1 생존 발견, RED 3커밋 전부 축자 재현,
표면 준수·AST 신선도 20/20·시나리오×endpoint 커버리지·S4 cleanup 경로 전수·클라이언트
정확-0600 전부 생존 확인.

| id | 등급 | 발견 | 판결 → 수정 |
| --- | --- | --- | --- |
| P1-A | P1 | owner-write-없음=사망 추정이 수락 중인 0400 socket을 삭제 — 탈취 SHALL NOT의 실측 반례(두 endpoint). a108 원형에서 이식된 병 | 수용 → T1-fix F1: chmod-then-probe(추정 제거, 결정적 판정). a108 대응물은 issues I1 후속 후보 |
| P1-B | P1 | 이름-집합 완전성 테스트가 상수 대 상수 — 발행 접두 분기 뮤테이션(A1-N1)이 전 테스트 생존. gate 문서의 "생성기를 실제로 돌려 잰다"는 과장 | 수용 → T1-fix F2: go/parser 정적 핀(리터럴 접두 거부 + 상수 3-참조 계수) + 문서 정정. A1-N1 재적용 시 CAUGHT |
| P1-C | P1 | T1 단독 land는 이물→거부→fatal 영구 루프를 신설(강등은 T2 소유) | 수용 → land 단위 = T1+T2 합본, tasks 배포 절차 명문화 |
| P2-A/B | P2 | "측정 불가" 생존 M23·M22를 결정적 핀으로 사망 실증 | 수용 → T1-fix F3: 핀 편입, 원장 재분류("방법을 안 썼다") |
| P2-C | P2 | M1a 등가 반증 — 거부한 디렉터리 안 우리 파일 삭제 관측 | 수용 → T1-fix F4: 공존 시나리오 핀, 원장 재분류 |
| P2-D | P2 | 수락 중인 staging socket을 probe 없이 삭제 | 수용 → T1-fix F5: "수락 중이면 이름 무관 unlink 금지"로 규칙 통일 |
| P2-E | P2 | 거부 오류가 위반 엔트리 미지목 — D3 회복 경로의 전제 훼손 | 수용 → T1-fix F6: basename 지목 |
| P2-F | P2 | 산주인+이물 공존 시 이물만 보고 | 잔존 수용(F6로 실행 가능; 두 원인 동시 수집은 "건드리지 않고 즉시 멈춘다"를 약화) — issues I4 |

## §2 A2 적대 리뷰 (T2 — 강등·재부착, 2026-08-15)

구성: Opus, 별도 컨텍스트, T2 워크트리(원복 검증 — porcelain 빈 출력 + sha256 4/4).
재현: 뮤테이션 5건 원장 일치(M13·M10b·M21·M7·M19 계열 — M7은 원장보다 강하게 확인),
RED 2커밋 재현 일치, D5b 독립 재현(3모양) 성립, 신규 측정 1건(defer 순서 역전 → 핀
FAIL — 원장 공백 발견). 3상태 보존·비차단·single-flight·3금지·핀의 조용한-통과 없음
·표면 준수 전부 생존 확인.

| id | 등급 | 발견 | 판결 → 수정 |
| --- | --- | --- | --- |
| P1-1 | P1 | 콘솔 전략 dial(console.go:411–421)의 부팅-1회 접힘이 미선언 생략 + T2 주석의 선례 인용 반만 참 | 수용 → design 선언된 생략 명시(Manager), 주석 정정(T2-fix F9) |
| P1-2 | P1 | internal/console -race 기본 timeout 경계(582~693s/600s) — 게이트 확률적 실패 + 원장 기준선 재현 불가 | 수용 → tasks §3.1 timeout 명문화(Manager), 원장 정정(T2-fix F9) |
| P2-1 | P2 | 요청 ctx 취소를 endpoint 실패로 오독 — 창마다 탈착 보고·client 교체 | 수용 → T2-fix F1(취소+호출자 ctx 종료 시 무전이) |
| P2-2 | P2 | 늦은 read 실패가 새 부착을 탈착으로 전복 | 수용 → T2-fix F2(자리 세대 비교 — 인터페이스 == 패닉 회피 사유는 issues T2-13) |
| P2-3 | P2 | 재부착 wake가 집계 7조회 성공에 종속 | 수용 → T2-fix F3(tick 첫 줄 presence wake, 새 ticker 없음) |
| P2-4 | P2 | 주석 핀 단정 2/3 공허 | 수용 → T2-fix F4(고유 구절, 재현 뮤테이션 CAUGHT 전환) |
| P2-5 | P2 | 격리 거부 제목 3벌이 빌드 탓 + a079 핀 | 수용 → T2-fix F5(상수 1벌 + a079 동반 수정; ErrUnwired는 로그 전용 잔존 — issues T2-11) |
| P2-6 | P2 | a098 결합이 어미 하나로 유지 | 수용 → T2-fix F6(errors.Is 교체) |
| P2-7 | P2 | 부작용 있는 술어 미명문화 | 수용 → T2-fix F8 |
| P2-8 | P2 | 시도 goroutine 종료 join 부재(이론·미관측 표기) | 잔존 수용 — 기록만 |
| P2-9 | P2 | 원장 gofmt 명령이 command-not-found 모양 | 수용 → T2-fix F9(절대경로) |
| P2-10 | P2 | M7 생존 사유가 T1이 고치는 nil 가드에 의존 | 수용 → T2-fix F7(nil-safe 계약 테스트 3종 — 병합 후 교차 검증 항목) |
| P2-11 | P2 | D5b 결론 참이나 구버전 Close의 조용한 rmdir 실패 미기록 | 수용 → 배포 절차 prose 반영(Manager) |

## §3 Manager 판결 요약

- A2 P0 0건. 두 리뷰의 P1은 계약·측정의 구멍(A1 P1-A·P1-B)과 운영 절차(A1 P1-C,
  A2 P1-1·P1-2)다. 안전 불변식 위반은 발견되지 않았다.
- 수정 라운드는 전 항목 Manager 판결 후 위임 — T1-fix F1~F6, T2-fix F1~F9. 임의 확장
  없음(T2-fix F2의 자리-세대 기전은 판결 취지 내 safe-local, 사유 기록됨).
- land 단위: T1+T2 합본만 main 후보(A1 P1-C).
- 후속 change 등록 후보 3건: a108 `projectionSocketAccepts`의 동일 추정(I1), 콘솔
  lifecycle client 재부착(P2-7), 콘솔 전략 dial 접힘(A2 P1-1).

## §4 수정 라운드 정산 (2026-08-16)

- **T1-fix** 8커밋(f9bbd09d~24d654d2): F1 chmod-then-probe(RED 5 FAIL 관측→GREEN),
  F2 정적 핀 + 과장 정정, F3/F4 A1 핀 편입, F5 staging probe, F6 엔트리 지목.
  뮤테이션 7건 적용 7 사망(A1-N1·M23·M22·M1a·F1-N1·F5-N1·F6-N1). 원장 재분류:
  생존 6→3(등가 1·경합 1 잠정·비root 1), 총 32 사망/3 생존. 검증: ppr 43·engine 593
  passed(-race, 단독 순차), strategyprojectionrpc 48 passed(원형 무변경 확인),
  make lint clean, FLM 재최신화(분기 집합 변화가 새 코드에만 있음이 AST로 확인).
- **T2-fix** 12커밋: F1~F9 전부 GREEN, 원장 추가 12건(M24~M32b — defer 순서 역전
  공백 포함) 전부 정산, check_analysis T2 표면 0오류. 검증: cmd/tossctl 172s ·
  httpapi 119 passed · console 693s(-timeout 25m) 전부 ok, 회귀 0.
- 잔존(정직 기록): I2 — command endpoint의 `SweepPrivateStagingLeftovers`에는 F5
  probe가 없다(호출자가 TCP뿐이라는 문서 전제 — 가장 약한 방어). I3 — M26은 잠정
  생존. T2-12 — `publishHTTPAPISnapshots` 서명 변경은 병합 충돌 주의점.

## §5 gstack 전체 리뷰 (2b.3, 2026-08-16 — 합본 0302b1b8 대 main)

구성: Scope Check(CLEAN) + Manager critical pass + specialist 4종(testing·
maintainability sonnet / security opus / performance sonnet, 전부 읽기 전용) +
Claude 적대 패스(opus, red-team 헌장) + codex 적대 패스(gpt, 1차 5분 타임아웃 후
10분 재시도 성공). **선언된 생략**: codex 구조화 리뷰(200+ 라인 규칙)는 1차 타임아웃
후 적대 패스로 대체 — 총 8개 독립 패스(freeze·A1·A2 포함)가 같은 디프를 덮었다.
red-team 전담 패스는 Claude 적대 패스가 그 헌장(specialist가 놓친 것 찾기) 그대로
수행했으므로 별도 미발사.

### 합의 발견 → T3-fix 위임 (G1~G9)

| id | 출처 | 발견 | 판정 |
| --- | --- | --- | --- |
| G1 | Claude 적대 #1 | attempt가 nil 자리에서 unavailable sentinel을 버림 → descriptor-잔재 상태가 영구 NOT_CONFIGURED(D4-2 접힘 재발) | 수정 |
| G2 | Claude 적대 #2 | publisher wake가 `failed` 게이트 — live-였다-죽은 자리는 집계 고장 시 영원히 안 깨어남(F3 반쪽) | 수정(무조건 wake) |
| G3 | Claude 적대 #3 + testing #2 | observe 성공이 `attached` 미복원 → blip 1회 후 탈착 보고 영구 침묵 + 허위 재부착 로그 | 수정 |
| G4 | Claude 적대 #4 | 강등된 command Start가 죽은 run의 descriptor(재활용 포트+토큰) 방치 — 소비자가 무관 프로세스에 bearer 전송 가능 | 수정(flock 보유 중 stale descriptor 제거) |
| G5 | security #1(CRITICAL) + Claude 적대 #5 | 재부착마다 이전 client의 idle conn·goroutine 영구 고정(Dial transport에 Close·IdleConnTimeout 없음) — 다중 확인 | 수정(Client.Close 추가 + eviction close + DisableKeepAlives — a108 패키지의 좁은 추가, 저장소 관례 인용) |
| G6 | Claude 적대 #6 | stale-socket 거부 1경로가 F6 미적용(원인·파일 미지목) | 수정 |
| G7 | security #2·#3 + Claude 적대 #7 | staging 검증 nlink 미요구(하드링크 공유 inode chmod 관측) + chmod의 이름 재해석 창 | 수정(staging nlink==1 + chmod 후 SameFile 재검증, 불일치=생존 간주) |
| G8 | performance | 회수 probe 무상한(N×200ms 순차, flock 보유 중) | 수정(probe 상한 + 초과는 이름 지목 거부) |
| G9 | maintainability | engine 패키지 내 verbatim 중복 2블록(Mkdir→회수→Mkdir, Close 해체) — 자기 인용 원칙과 모순 | 수정(패키지-로컬 helper 추출, 기계적) |

### 기각·기록

- codex P1 "열거 후 발행 socket의 무probe unlink"(probe-remove TOCTOU): **기각 —
  문서화된 잔존.** 그 창의 방어는 journal flock(a108 transport_unix.go:317–320 명문,
  a109 design D2 동일 인용). 엔진 아닌 same-uid 점유는 freeze부터 수용된 위협 경계.
  T3-fix가 removal 창까지 주석 인용을 확인한다.
- security #4 "a108 원형의 owner-write 추정 잔존": 기지 사실 — issues I1·design 선언된
  생략·후속 change 후보로 이미 기록(A1 P1-A의 a108 대응물).
- testing #1(Read의 nil 가드 미실행): G1 테스트가 그 경로를 함께 덮는다.
- maintainability의 rtk 디프 압축 관측: 리뷰어가 `rtk proxy`로 우회해 전문 검토 —
  절차 기록.

### T3-fix 정산 (2026-08-16)

- 12커밋(866265ef~031856e5), G1~G10 전부 RED→GREEN(각 RED의 실제 FAIL 관측 기록).
  전 스위트 재검증 green(engine 910s·console 500s 포함), strategyprojectionrpc 48→50
  tests 회귀 0. 원장 추가 13건(M33~M44): 생존 2건(M38·M42)은 새 핀으로 CAUGHT 전환 —
  **M42는 A1 P1-B와 같은 병(테스트가 상수를 읽어 입력을 만든다)의 재발을 T3-fix가
  스스로 잡은 것.**
- G5의 a108 패키지 접촉은 선언대로 좁다: `Client.Close` 메서드 추가 + `Dial` transport
  `DisableKeepAlives`(저장소 관례 인용) + 테스트 1파일. 회수·발행 의례 무수정.
- 잔존 선언: G4의 RED는 정적 자리 핀(제거 지점 이후 실패 seam 부재 — 동작 2핀 + 순서
  핀으로 분할), G2의 대가(건강 자리도 간격당 재해석 1회 — eviction close·전이 1회
  로그로 무해), I5~I7.

### Manager 독립 검증 (2026-08-16)

- 합본 트리 전 스위트(build·vet·5패키지 -race 순차) rc=0 — T3-fix 이전(교차 검증
  스크립트)과 이후(T3-fix 자체 실행) 각 1회.
- 스폿체크 뮤테이션 2건 독립 재현: ① G2 wake를 `failed` 게이트로 되돌림 →
  `TestThePublisherWakesASeatThatOnlyLooksAlive` FAIL(정확히 그 핀). ② G7 SameFile
  재확인 무력화 → `TestTheProbeRefusesASocketThatChangedUnderIt` FAIL. 둘 다
  `git checkout` 원복 + porcelain 빈 출력 + 심볼 수 대조. (②의 1차 좁은 -run 생존은
  내 패턴 누락이었음을 전체 실행으로 확인 — 표본 선택의 함정 기록.)
- httpapi_strategy_attach.go 전문 판독: G1 승격의 단방향성(live 격하 금지), seat
  세대·취소 판정·전이 1회 계약이 설계 취지와 일치함을 확인.
