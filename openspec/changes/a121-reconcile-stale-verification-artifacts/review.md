# Proposal-freeze review

## Status: freeze 재리뷰 수리 라운드 (2026-10-05 — 아래 「freeze 적대 리뷰 1차」)

(아래 BLOCKED 는 2026-09-06 당시 기록이다. Q1~Q6 결정과 로트 1 착지로 전제가 바뀌었고,
현 상태는 freeze 적대 리뷰 1차의 P0 1·P1 7 수리 반영 → 같은 리뷰어 재검 대기다.
RED·구현·계좌 읽기·라이브 명령은 여전히 재검 APPROVE 전에 열리지 않는다.)

## (역사) Status: BLOCKED before implementation

The independent adversarial review at
[`analysis/proposal-freeze-adversarial-review.md`](analysis/proposal-freeze-adversarial-review.md)
blocked this proposal on 2026-09-06. No RED work, implementation, account read,
or live command is authorized.

The current official OPEN/CLOSED pagination has no proven consistent snapshot
cut; therefore absence of a parent conditional cannot safely exclude a
replacement or triggered child. The contract also lacks a versioned non-raw
identity digest and executable profile/account/market equality checks. These
must be specified, including fail-closed behavior when the official API cannot
prove a consistent read, before an adversarial re-review and subsequent gstack
review can run.

`make sdd-check` remains blocked by stale/missing CodeGraph hard evidence even
after `make sdd-sync`; it is an implementation blocker. PM generation and its
check now pass. a063 remains operationally blocked and unarchived.

## Revision 1 (2026-09-27) — 설계 갭 해소 초안, 재리뷰 전

초안 3cf64d5d 뒤 gstack 문서 리뷰 1회(읽기 전용, HEAD cb378a63, `analysis/ast-evidence/` 12개의 source_sha256 이
HEAD blob 과 일치)를 반영했다. 판정은 "적대 재리뷰 전 개정" — P0 1 · P1 8 · P2 6.

| # | 등급 | 지적 | 처리 |
|---|---|---|---|
| 1 | P0 | 정정으로 id 가 바뀐 후속이 발동하면 다른 id 로 CLOSED 에 남아 규칙이 통과시킨다; resting child 가 안 보인다; Q1(a) 가 과대 | G1-2 를 "심볼의 CLOSED 에 어떤 id 든 `triggeredOrderId`·`COMPLETED` 가 있으면 거절" 로, G1-3(OPEN 일반 주문 0건) 신설, Q1(a)(b) 문구 수정(보존 기간 측정 필요 · `Artifact` 에 방향·수량 없음), RED 2.2.1 추가 |
| 2 | P1 | "HeldUntil 이 걸린 artifact 제외" 가 a063 의 artifact 자체를 뺀다; M0 수동 대사 제외 누락 | 후보 = `PendingCleanup` 의 조건주문 − `M0Unsettled` 대상. spec·RED 2.2.2 동반 수정 |
| 3 | P1 | 현재 참조는 원문 계좌번호(`reads.go:93`), 기록은 `attest.Mask` — 문자열 비교는 늘 거절 | `attest.Mask(TrimSpace(ref)) == entry.AccountRef`, 끝 4자리 한계를 Q5 에 묶음 |
| 4 | P1 | 환경 변수 자격 증명·기본 경로가 달라 "구성으로 프로필 결속" 이 거짓 | `--config-dir` 필수 + 환경 변수 자격 증명 거절 + override 거절, RED 추가 |
| 5 | P1 | 시장 대조가 수락 경로(행 0건)에서 비어 있다; 파일 이름 대조는 동어반복 | 시장 결속 = 파일 이름 + 심볼뿐임을 명시, 행의 심볼 불일치 거절 추가 |
| 6 | P1 | proposal 이 막는 기전을 trigger 모드의 B9 로 잘못 적었다 | 보통 `verify run` 의 재개 정리 재선택 + 조건주문 상한(`MaxLiveConditionals = 1`)으로 고치고 B9 는 trigger 모드 한정으로 |
| 7 | P1 | 대사 줄의 `Calls` 가 성공 endpoint 가 된다(구 바이너리 포함) | 대사 줄은 `Calls` 를 싣지 않는다(design·spec·RED 2.4.1) |
| 8 | P1 | 잠금·동시성 규칙 없음 | 실행 flock + rate-budget lease, 추가 직전 기록 재대조(design·spec·RED) |
| 9 | P1 | spec 옛 수락 시나리오가 부재만으로 수락; Q4 미표기; task 1.2 에 report/status 누락; "Q1 미답 → 거절" RED 없음 | 시나리오 교체 + "Q1 is unanswered" 시나리오, `[비움 — Q4]`, 1.2 에 `BuildReport`·`BuildProgress`, RED 추가 |
| 10 | P2 | 구 바이너리 주장은 맞음 — 다만 다시 보유 상태도 됨; 시뮬레이션 방법 | design 에 "다시 보유" 추가, 2.4.1 에 방법 명시 |
| 11 | P2 | "B12·B13 빈 커서" 오인용 | B13 만(README·design) |
| 12 | P2 | `Market` 줄·kinds 범위 인용 오차 | `:111`, `record.go:65-88` |
| 13 | P2 | 읽기 도구 미지정 | `ProtectionConditionalOrdersRaw` + `OrdersPageRaw` 로 지정 |
| 14 | P2 | 두 읽기를 `Digest` 로 비교하면 순서 의존 | 정렬 집합 비교, 지문은 정규 순서 직렬화 |
| 15 | P2 | `EXPIRED` 거절·페이지 상한의 결과 미기재 | 「치르는 값」 절 + Q6 신설 |
| 16 | P2 | "required commit ancestry is absent" 출처 없음 | `tools/logic-map/execution_baseline.py:100` 인용 |

반영 뒤 분기 근거 AST 4개를 더했다(`checkConditionalCap`·`holdGate`·`cleanupFrom`·`SucceededEndpoints`, HEAD `0004536c`).
남은 사람 결정은 Q1~Q6 이다(design 「열린 질문」).

상태는 그대로 **BLOCKED** 다 — 이 초안은 RED·구현·계좌 읽기·라이브 명령을 허가하지 않는다. 적대 재리뷰는
Manager 지시 뒤에 돈다. `make sdd-check` 차단(위 절)도 그대로 구현 차단 조건이다.

## Manager 결정 (2026-09-27) — Revision 1 열린 질문

| Q | 판정 | 내용 |
|---|---|---|
| Q4 | **Manager 확정** | 같은 종목의 다른 조건주문 = 거절. 후속 증명 예외는 만들지 않는다(범위 확장 금지). |
| Q5 | **Manager 확정** | 복수 계좌·끝 4자리 충돌 = 거절. |
| Q6 | **Manager 확정** | EXPIRED = 거절. 귀결 — a063 의 artifact 가 EXPIRED 면 이 경로로는 영구히 대사되지 않는다. 거절 메시지와 문서(design G1-2·Q6, spec) 양쪽에 적는다. 그 경우 사용자 재결정 항목이 된다. |
| Q1 | 사용자행 | 발동 후 CLOSED 잔존의 실측(계좌 읽기 승인 필요) / 보유 기준값 확장 / 영구 거절 수용 중 선택. |
| Q2 | 사용자행 | 원문 id 경계 문구. Manager 추천: 기존 키 재사용 + 경계 문구 수정. |
| Q3 | 구현 로트로 이연 | 상수 없음 — 구현 시 보수 형태로 제안. |

spec 의 `[비움 — Q4]`·`[비움 — Q6]` 는 위 판정으로 채웠고 Q5 거절을 계좌 결속 요구에 넣었다.

## 로트 1 — 2026-10-05 (1.0 집행 · base 재고정 · FLM 증거, 커밋 de147cc2·b81380dd·29a0d745)

- **1.0(Manager, de147cc2)**: Q1~Q6 기결정을 design 3자리·spec 3자리에 반영, placeholder 0.
- **재고정(Manager, b81380dd)**: 9408fc95 → de147cc2, WORKFLOW 세 요건 전부 — 귀속 실측(자기 Go
  커밋 0 · 대상 소스 0파일 · 옛 required 242 전부 형제 몫), 승인 기록(tasks 1.1 · 이 절),
  단독 커밋+영수증(커밋 메시지). 사이 582커밋.
- **로트 1(Terra Opus, 29a0d745)**: FLM 번들 4 · 호출자 증거 · ast-evidence 16/16 재검증 +2.
  Manager 배터리: 경로 전수 22/22(허용 밖 0), check_analysis rc=0(HEAD 29a0d745), 공유 트리 잔여
  기지 untracked 2뿐. 로트가 보고한 어긋남 D1(분기 귀속)·D2(범위 끝행)·D4(env 우선 조건)는 design
  정정, D3(Mask 경계)·R1(StepID)은 tasks 2.2.2 RED 로, S1(WriteText 편집 대상 추가·콘솔 비편집)·
  S2(mutating=true)·STORY 판독(제약)은 design 「로트 1 처분」으로 닫았다.
- 다음: 1.3 proposal-freeze 적대 리뷰(+gstack) → RED 로트.

## freeze 적대 리뷰 1차 — 2026-10-05 (독립 Terra Opus, 58870829 고정 사본)

**판정: P0·P1 수리 후 재검**(P0=1 · P1=7 · P2=11). 격리 준수, AST 22/22 sha 일치·표본 3 바이트 동일,
인용 전수 재검(깨진 것 P2-1 하나). Manager 수리 반영(이 커밋):

- **P0-1** Q1 한도의 형태 부재 → design G1-4: 기간·기준점 outstanding 줄 `CreatedAt`(영이면 거절)·
  생산 nil 비공개 값(시험만 주입, 실값은 측정 후 리뷰된 상수 커밋, 설정 경로 금지)·AST 핀. spec 동문.
- **P1-1** no-live-mutation 봉인 불가 → 읽기 전용 좁은 인터페이스+전용 생성자(Broker 비반환),
  대사 파일 AST census(쓰기 7이름+type assertion 금지), 토큰 POST 는 경계 밖 명시(P2-9).
- **P1-2** G3-2 미배선 → 대사 경로 자체 `Accounts()` 호출(resolveVerifyAccount·buildVerifyBroker 비편집).
- **P1-3** 공허한 수락 → 양성 대조 2(심볼 바이트=기록 원문·종목 조회 GET 에코), 잔여 위험 명기.
- **P1-4** 정렬 집합 → multiset·한 읽기 내 (그룹,id) 중복 거절·읽기 순서 고정.
- **P1-5** a063 영구 거절 공산 → 치르는 값·proposal 에 기록, **Q1 표본은 a063 심볼 금지**(tasks 1.0).
- **P1-6** RedoSet 부활 → design 명시 + tasks 2.3 RED 핀.
- **P1-7** 삼면 불일치 → spec(env 반쪽·복수 계좌 시나리오·S10 multiset 키·proving 문구),
  tasks 2.2.1/2.2.2 RED 추가(한도 초과·Q3 부재·Q6 문구·중복 신원·flock/lease·찢긴 꼬리 등).
- **P2** 11건 전부 반영: 인용 2건 정정(P2-1·2), 낡은 상태 표기 주석(P2-3), R1 소비자 4(P2-4),
  terminal 직접 리더 3 명기(P2-5), 대사 줄 필드 확정(P2-6), 지문=파일 바이트 sha256+`\n` 꼬리(P2-7),
  OCO 잔여(P2-8), 경계 문구(P2-9), spec proving(P2-10), runCleanup FLM 유보(P2-11).

재검(같은 리뷰어, 표적)은 수리 커밋 좌표로 돈다. gstack 리뷰는 재검 APPROVE 뒤.

**파킹.** a121 은 BLOCKED 로 둔다. 적대 재리뷰는 사용자가 Q1·Q2 에 답한 뒤에 돈다.
(2026-10-05 해소 — Q1·Q2 는 2026-09-28 사용자 결정, 적대 재리뷰는 아래 freeze 1차로 실행됨.)
