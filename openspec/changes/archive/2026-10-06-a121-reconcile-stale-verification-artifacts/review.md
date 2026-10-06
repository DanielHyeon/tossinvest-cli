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

## freeze 재검 — 2026-10-05 (같은 리뷰어, 3447441f 표적)

원 발견 19/19 CLOSED. 신규 **P0-R1**: spec 「refuses before any official read」 가 P1-2 의 자체
`Accounts()` 호출(그 자체가 공식 읽기)과 모순 — 1차가 놓친 기존 모순을 수리가 다섯째 사본으로
노출. 수리(이 커밋): 요구·시나리오 5곳·design :292 를 「주문/조건주문 목록 읽기 전(계좌 목록·종목
조회 GET 만 선행 허용)」으로 정정. 동기화 P2 4건 반영 — (a) sorted set→multiset·Q3 「RED 로트가
확정」, (b) tasks 2.4.1 에 census 명세·생산 nil AST 핀·토큰 예외 명시, (c) 대사 줄 `CreatedAt` =
원본 outstanding 줄 값(추가 시각은 `ReconciledAt`), (d) 양성 대조 spec 요구+시나리오·객체 유형
0-후보 RED. 리뷰어 선언: 이 수리 뒤 추가 재검 불요 — 최종 이진 확인만 수행.

**1차 최종(2026-10-05): FREEZE-APPROVE — P0=0, P1=0** (같은 리뷰어, ed33a732 이진 확인 3/3 YES —
P0-R1 닫힘·P2 a~d 선언대로·새 모순 없음. 잔여: "same" 중복 오타 1 — 이 커밋에서 수리).
수용된 no-live-mutation 경계(tasks 1.3 요구 기록): **대사 경로에서 주문·조건주문 변이 도달 0** —
읽기 전용 좁은 인터페이스(목록 2 + 계좌 + 종목 조회)·Broker 비반환 생성자·대사 파일 AST census
(쓰기 7이름·type assertion 금지)·토큰 갱신 POST 만 경계 밖(auth 기반). 명령은 mutating=true 로
등재하고 실행은 사람 승인 뒤(4.3).

## gstack 리뷰 + codex 외부 적대 — 2026-10-05 (tasks 1.3 둘째 반쪽)

- gstack /review 실행(대상 d028db60..a56b8481, 문서 28파일 +1585/−49): scope CLEAN(전부 a121),
  체크리스트 코드 범주 not-applicable(코드 0), 기계 재검(placeholder·모순 문자열 잔존 0),
  strict validate 통과. Claude 적대 반쪽은 위 freeze 3라운드(독립 에이전트)가 실질 수행 — 재탕
  대신 그 기록을 매핑.
- **codex 외부 적대 r1 (gpt-6-astra): 형식상 무효** — 출력 첫 줄에 `~/.codex/memories/MEMORY.md`
  접근 자가 선언(하네스 메모리가 프롬프트 금지보다 상위로 동작). 선례(a112 8.5)대로 r1 은 패스로
  셈하지 않되, **발견 7건은 Manager 가 실코드로 검증 후 수용**: F2(`unwrapAndDecode` 가
  `{"result":null}`/`{}` 를 무오류 빈 페이지로 — client.go:213-227 실측 확인),
  F3(`LoadEntries` 가 깨진 꼬리 줄을 개행 무관 묵살 — record.go:423-430 실측 확인, 1차 P2-7 의
  `\n` 규칙 **불충분 판명**), F6(API `Second` 존재·어댑터 탈락 — conditional_reads.go:38-39 확인),
  F1(마스크 결속 우회: 접미 충돌 자격 교체)·F4(CLOSED denylist)·F5(행 수 축출 미배제)·F7(Q3 창이
  추가 전 종료) 논증 수용.
- **수리(이 커밋)**: F2 스키마 존재 검증 / F3 전 줄 엄격 해독(선택 전+추가 직전) / F4 status
  allowlist(골든 전사)+결측·미지 거절 / F5 Q1 측정에 축출 모형 요구(행 수 기반·판별 불능 =
  영구 거절) / F6 OCO 행 전면 거절 / F7 Q3 창을 추가 승인까지 연장+직전 재검사 / F1 대사 줄
  계좌 digest+승인 출력 두-마스크 표시+잔여 기록(작성기 내구 결속은 후속 change — proposal).
  spec·tasks RED 동기화. 치르는 값 ④⑤⑥ 추가.
- 다음: 수리 좌표로 **clean codex r2**(격리 사본·메모리 미부착 환경) — r2 가 이 수리의 유효
  외부 패스다. 그 뒤 freeze 리뷰어 이진 확인.

## codex 외부 적대 r2 — 2026-10-05 (clean, 유효 패스 — /tmp 격리 사본 f6c8272f, gpt-6-astra)

메모리 접근 선언 없음(격리 사본 + 최우선 금지 문장). **판정 BLOCK — P1 4·P2 1, 전부 FIXABLE.**
F1~F7 수리 자체는 7/7 설계 반영 확인(표 제출), 발견은 수리의 2차 결함:

- **R2-1**: 존재 검증만으로 부족 — null 값 필드(`hasNext:null`→false 접힘·컬렉션 null→빈 목록)가
  남는다. 수리: non-null 배열·non-null 불리언·커서 일관성, 일반 주문 페이지 포함.
- **R2-2**: 검증한 seq ≠ transport 캐시 seq — `Accounts()` 가 첫 양수 seq 를 자동 캐시
  (`reads.go:38-54`). 수리: 좁은 생성자의 명시 재결속(`WithAccountSeq` 선례) + 기형 계좌 행 거절.
- **R2-3**: F3 가 개행 규칙을 「대체」해 과수리 — `Append` 는 뒤에만 개행을 붙여 개행 없는 완전
  JSON 꼬리에 이어 붙으면 양쪽이 깨진다. 수리: 엄격 해독 ∧ 개행 꼬리 검사 **병존**.
- **R2-4**: F5 의 영구 거절이 spec 규범문·시나리오·RED 에 미전파. 수리: spec 두 곳 + RED 3사례.
- **R2-5**: 「keyed digest」 표기 vs 무키 공식 모순 + 계좌번호 추측 대조 가능(저엔트로피) —
  **digest 철회**(마스크 외 계좌 신원 비저장, 안전 불변식 8), 내구 결속은 작성기 후속 change 로.
  부수: Q5 의 「충돌은 계좌 둘 이상일 때만」 반증 문장 정정.

수리는 이 커밋. 다음: freeze 리뷰어(Claude)의 codex 라운드 수리 이진 확인 → 1.3 종결.

## freeze 최종 — 2026-10-05 (리뷰어 이진 확인 + 2줄 수리)

리뷰어 판정(adaebf69 표적): P0=0·P1=0 유지, 원 19건·P0-R1 재개 없음(R2-3 은 P2-7 과 병존 확인),
유일 잔존 = **R2-5 digest 철회의 proposal·design 잔존 문구 2줄(P2)** — 리뷰어 선언: 그 2줄 수리
커밋으로 **FREEZE-APPROVE (P0=0/P1=0), 추가 재검 불요**. 이 커밋이 그 수리다: proposal 의
「대사 줄 digest + …」 → 「마스크 참조 + 두-마스크 표시 + seq 재결속」, design R2-2 절의
「표시·digest」 → 「표시」. 비차단 P2 3건도 반영 — spec :178 「without any list read」, 스키마
검증 null 값·커서 일관성 spec 명문화, 승인 순서(목록 읽기 전) design 핀.
**∴ tasks 1.3 종결 좌표 = 이 커밋. 유효 리뷰 사슬: Claude 적대 3라운드(19건) + P0-R1 +
codex r1(무효·발견 7 수용) + clean codex r2(5건) + 이진 확인 3회 — 전부 CLOSED.**

## A-RED — 2026-10-05 (RED 로트 적대 리뷰, 착지 ffa8eb0d)

- **구현**: Terra-RED(Opus), 2c6ef1ef 고정. RED 82·골격 4·봉인 census·FLM 번들 3·Q3 메모.
- **A-RED 1차**(freeze 리뷰어 — 구현자와 별개): P0=0 / P1=4 / P2=6. P1-1 봉인 census 우회
  (Broker 필드·New 참조 변이 생존 + glob 탈출) / P1-2 testseam census 미등재(sdd-test FAIL —
  가드 설계대로 격발) / P1-3 기본 승인 미핀(auto-approve GREEN 이 전판 통과) / P1-4 이름만 있는
  거절 모양 3. 리뷰어가 구현자 주장 재실측: 29파일 전부 신규·회귀 0·롤백 사본 핀 14/14,
  계수 정정(RED 74·코드 45).
- **수리 r2**: 봉인 3층(이름 금지 AST·타입 핀·go/types 도달 census+allowlist+import 금지)·
  EXPECTED 26·승인 4시험·모양 4시험·Q3 상수 15s production-var 핀·basis digest 값 핀·allowlist
  주입 수락 시험. 리뷰어 재검: P1-1·2·4·P2 전부 CLOSED, **P1-3 반쪽**(실제 terminal 기본값이
  tty 미판정 — `echo y |` 자가 승인 변이 생존).
- **마감**: tty 실판정 RED(os.Pipe·/dev/null 모두 비대화형 단언 — /dev/null 은 문자 장치라
  기존 isTerminal 로는 오판, GREEN 은 x/term 필요) + import 금지 목록 + 주석. TERMINAL·
  chardevice·net/http·os/exec 변이 CAUGHT 전환 실측. 리뷰어 선언에 따라 추가 재검 생략.
- **Manager 배터리**: 경로 전수 31/31(코드 17 신규·openspec 13·tools/sdd 1), 트레일러 정상,
  RED 수 재실측 일치, 비-RED FAIL 0 재확인(수리 전 라운드), check_analysis rc=0·sdd-test rc=0
  (커밋 뒤 공유 트리). **Manager acceptance: ffa8eb0d 푸시.** 구체화 판정 7건·가드 순서·Q3
  승인·allowlist 출처는 design 「RED 로트 처분」.

## A-GREEN — 2026-10-05 (GREEN 로트 적대 리뷰, 착지 c9889cb2)

- **구현**(같은 Terra, 리뷰어 분리): RED 82 전부 GREEN, 생산 신규 verifylive 4 + record/report
  확장·official 해독 3·cmd 명령(등록 포함), 기존 함수 편집 6(Pre-Edit·FLM·분기 수 불변),
  `Artifact.MarshalJSON`(대사 줄만 확정형·비대사는 base 바이트 동일 핀), x/term direct.
- **판정 경과**: GREEN 중간 보고에서 census 읽기 통로 누락(RED 로트 결함)이 census FAIL 로 격발 —
  Manager 승인으로 `readRecordRaw` 한 줄 허용(사유 기록). 변이 원장 과정에서 생존 8 을 새 시험으로
  잡고 중복 분기 1 제거.
- **A-GREEN(같은 freeze 리뷰어): APPROVE — P0=0/P1=0, P2=3.** 실측: 시험 약화 0(허용 2줄뿐),
  MarshalJSON 에 자체 변이 10(8 CAUGHT·1 빌드실패·1 → P2-b), 가드 순서 AST 전수 일치(추가 재검은
  강화 방향), no-live-mutation 코드 추적(GET 3·seq 재결속 실재), 원장 표본 10/10 재현·동등 2 논증
  성립, 기존 함수 6 분기 수 불변·AST 바이트 동일, §0 불변식 영향 0, go.mod 선언대로.
- **P2 처분**: (a) 정적 거절 승인 선행 — 계약 개정 수용(GO2 변이 CAUGHT 핀), (b) 중복 줄 삭제
  (MJ10 을 새 시험이 잡음 — 기존 fixture 는 HeldUntil 공란이라 못 잡았음), (c) 주석. 최종 원장
  110 중 CAUGHT 108·동등 2(V-tail-append — 지문이 개행 함의 / C-env-creds — preflight 선행 핀).
- **Manager 배터리**: 경로 전수 45/45(허용 밖 0), 트레일러 정상, 커밋 뒤 check_analysis·sdd-test·
  3패키지 무태그 전부 rc=0(구현자 실측). **Manager acceptance: c9889cb2 푸시.**
  잔여: 4.1 의 sdd-sync/sdd-check(Manager), 4.2 gstack+codex 코드 diff 패스, 4.3(사람)·4.4 gate.

## codex GREEN 외부 패스 + CG 수리 — 2026-10-05/06 (착지 cfdf219a)

- **codex(clean, /tmp 격리·메모리 접근 0)**: GREEN 코드에서 FIXABLE 5·INVESTIGATE 1 —
  CG-1 둘째 스냅숏 미검증(OCO 다리·심볼·시장 탈출, 최강 발견) / CG-2 기록 쓰기 직렬화 부재
  (inode alias·--record override) / CG-3 자기 부분 쓰기 창 / CG-4 승인 읽기 오류 무시·취소
  불가·표시 채널 분리 / CG-5 최종 게이트가 쓰기 경계 앞 아님 / CG-6 토큰 캐시 신원 교체.
  전부 미래 활성 경로(현 생산은 Q1 nil 거절 전용).
- **수리**: 양 스냅숏 독립 검증·기록 파일 flock+잠근 fd 바이트 판정+크기 재검·단일 Write+fsync+
  read-back·승인 강건화(완전한 줄·ctx·같은 채널)·게이트 재배치(AST 핀)·읽기 클라이언트 신원
  재확인. RED 시험 수정 2건은 Manager 승인(census 추가 통로 교체·/accounts 기대 1→2 — 리뷰어가
  비약화 확인). 원장 130: CAUGHT 127·동등 3(V-tail-append·C-rebind·C-env-creds, 전부 핀 결속).
- **CG-APPROVE(P0=0/P1=0, 같은 리뷰어)** — 6건 전부 counterexample 재구성 이진으로 CLOSED 확인,
  flock 플랫폼 거절(비-unix fail-closed)·OpenRecorder 비편집(record.go 삭제 0줄) 확인.
  P2 2(찢긴 꼬리 제자리 복구 가능성·승인 고루틴 누수 — CLI 수명 내 무해) 기록.
- **Manager 배터리·acceptance**: 경로 전수 20/20·트레일러 정상·커밋 뒤 3검사 rc=0 — cfdf219a 푸시.
- **에이전트 트랙 종결.** 잔여 = 4.3(사람 승인 조회 전용 관측 — 현 상태에선 정적 거절
  retention-unmeasured 가 정상 결과) → 4.4 gate·PM·archive. Q1 사람 실측(CLOSED 어휘 영수증 포함,
  a063 심볼 금지)이 활성화의 유일 관문.

**파킹.** a121 은 BLOCKED 로 둔다. 적대 재리뷰는 사용자가 Q1·Q2 에 답한 뒤에 돈다.
(2026-10-05 해소 — Q1·Q2 는 2026-09-28 사용자 결정, 적대 재리뷰는 아래 freeze 1차로 실행됨.)

## 4.3 — 2026-10-06 (사람 승인 조회 전용 관측)

- **승인**: 사용자 "1번 즉시"(2026-10-06) — Manager 가 제시한 정확한 명령·예상 결과에 대한 명시 승인.
- **실행**: `go run ./cmd/tossctl verify reconcile --config-dir ~/.config/tossctl --market KR`
  (브랜치 bccae0bb, 2026-10-06 13:28:20 KST).
- **관측 결과(원문)**: `verify reconcile: refused (execution-lock): verify: engine, system update,
  or another verification owns the execution exclusion: engine: another engine instance already
  holds this journal directory (/home/daniel/.config/tossctl/engine.lock)` — exit 1.
- **부작용 0 증거**: 거절 지점은 `runVerifyReconcile` 의 flock 단계(`cmd/tossctl/verify_reconcile.go`
  84~86행) — rate budget·reader 생성자·승인 프롬프트·네트워크 읽기·기록 쓰기 전부 **이전**.
  주문 side effect 0, API 호출 0, 기록 파일 무변경.
- **lock 보유자 실재**: 라이브 엔진 컨테이너(이미지 5c77491d93d4, Up 2 days healthy)가 저널
  디렉터리를 보유 — stale lock 아님.
- **예측 정정(침묵 생략 금지)**: 직전 기록은 "현 상태 정상 결과 = retention-unmeasured 정적 거절"
  로 예측했으나, 그 정적 거절은 verifylive.Reconcile **안**(execution lock 뒤)이라 라이브
  프로필에선 엔진이 저널을 쥐는 한 도달 불가다. 실측된 execution-lock 거절이 더 바깥의 설계된
  fail-closed 층이고(검증이 엔진과 경합하지 않는다는 계약), retention-unmeasured 거절·가드
  순서는 커밋된 시험과 변이 원장 130(CAUGHT 127)이 핀한다. 라이브에서 retention-unmeasured 를
  직접 보려면 엔진이 저널을 놓은 창(KST 05:00~09:00)이 필요하나 4.3 의 요구("read-only, redacted
  observation with explicit human approval; no live order mutation")는 본 관측으로 충족된다.
