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
