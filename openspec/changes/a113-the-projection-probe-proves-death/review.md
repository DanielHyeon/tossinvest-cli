# a113 리뷰 기록

## §0 proposal-freeze (2026-09-25)

위험 등급: Normal(엔진 boot 경로의 회수 기계 — 주문·손절·원장 경로 아님). 보이스: Teammate
셀프 4관점(CEO/Eng/Security/QA) + **독립 적대 Eng 리뷰어 1**(별도 컨텍스트 서브에이전트, 읽기 전용).
`autoplan` 스킬 대신 독립 서브에이전트를 쓴 이유: autoplan 은 대화형 4보이스 파이프라인이고 이 세션은
사용자 질문이 불가한 자율 세션이다 — Manager 지시의 대체 경로(독립 서브에이전트 적대 리뷰)를 택했다.

### 셀프 4관점

- **CEO/Product** — 범위는 a109 issues I1 한 건. 확장 없음(staging probe·12자 통일은 선언된 생략).
- **Eng** — probe 가 `Dial` 과 공유된다는 사실(AST: Dial B3 :417)을 등록 문서는 몰랐다. 등록 문구
  그대로 "projectionSocketAccepts 를 chmod-then-probe 로 교체"하면 조회 클라이언트가 엔진 socket 을
  chmod 한다 → design D1 이 회수 전용 함수로 나눴다.
- **Security** — chmod 결과는 0600 뿐(넓힐 수 없음), 도달 전 소유·symlink·nlink·0700 검증,
  SameFile 재확인. 조회 클라이언트는 여전히 chmod 하지 않는다.
- **QA** — 행동 RED(Start 가 산 socket 을 지우고 두 번째 서버를 세우는 것)와 판정 표 RED 를 분리,
  root Skip, 뮤테이션 4종 실측.

### 독립 적대 Eng 리뷰 (2026-09-25, Opus 서브에이전트, 별도 컨텍스트, 읽기 전용 — go vet/test 만 실행)

P0 0 · P1 4 · P2 6.

| id | 발견 | 판결 → 반영 |
| --- | --- | --- |
| P1-1 | spec delta 가 형제 요구(:356)를 고쳤다 — projection 요구는 :1146 이고, 새 SHALL 은 probe 없는 staging 회수(`transport_unix.go:270–280`) 때문에 착지일에 거짓 | 수용 → delta 를 :1146 MODIFIED(전문 + 시나리오 5 + 신규 1)로 옮기고 SHALL 을 **최종 이름 socket** 으로 좁힘. staging 은 flock 방어 — issues R1 |
| P1-2 | design(chmod → nil 검사) 과 draft 테스트(nil 이면 chmod 안 함)의 순서 모순 | 수용 → nil → 생존을 chmod 앞에, **chmod 전 Lstat+SameFile** 추가(검증 안 한 inode chmod 창 축소), chmod 후 재확인 유지 |
| P1-3 | chmod 모드 0o666/0o777 뮤테이션이 모든 계획 테스트에서 생존 — RED 테스트가 스스로 0600 chmod 해 모드를 가림 | 수용 → RED 테스트와 산 0400 표 행이 probe 뒤 정확 0600 단정, 뮤테이션 N5 |
| P1-4 | tasks.md 가 등록 문구 그대로 | 수용 → design D1/D2 기준 재작성 |
| P2-1 | `verifyStaleSocketShape` 가 두 번 stat — 돌려줄 inode 와 uid·nlink 를 본 inode 가 다를 수 있음 | 수용 → 같은 Lstat 의 `Sys()` 에서 uid·nlink(원형 `validateOwnerAndLinks` 모양) |
| P2-2 | 회수가 verify 의 info 대신 새 Lstat 을 넘기는 배선 뮤테이션은 행동으로 못 가림 | 수용 → go/parser AST 핀 + 뮤테이션 N6 |
| P2-3 | 거부 테스트가 사유를 안 봄 | 수용 → RED 테스트가 "still alive" 단정 |
| P2-4 | "접근이 넓어질 수 없다"는 owner 에게 거짓(0400→0600) | 수용 → design 문구 정정 |
| P2-5 | `os.Chmod` 는 symlink 를 따라감 | 수용(기록) → issues R2 (fchmodat2 NOFOLLOW 가 6.6 미만에서 EOPNOTSUPP 인 것을 실측해 대안 기각) |
| P2-6 | 낡을 주석 4곳 | 수용 → proposal·`projectionSocketAccepts` doc·`Dial` 주석 정정. 형제 패키지 주석은 표면 밖 — issues R3 |

생존한 공격(리뷰어 판정): `Dial`·소비자 무회귀(정확-0600 선행으로 B3 은 경합 창에서만 발동),
죽은 0500/0400 잔재 회수 유지, 산 쓰기-깎인 socket 보존(draft RED 가 base 에서 빨강), chmod 전제
조건이 호출 지점에서 성립, nil 가드 단독 제거는 등가 변이(`SameFile(nil,x)=false`), root 에서 순수
probe 정확, staging 비목표는 비차단.

freeze 종결 — 구현 착수.

## Pre-Edit Gate (2026-09-25)

- change id / task id: a113-the-projection-probe-proves-death / 1.1–1.2
- 대상 심볼: `strategyprojectionrpc.projectionSocketAccepts`(절 삭제) · `strategyprojectionrpc.verifyStaleSocketShape`
  (반환 확장·단일 stat) · `strategyprojectionrpc.reclaimStaleControlDirectory`(B11·B12 두 줄) ·
  신규 leaf `staleProjectionSocketAccepts`(새 파일)
- CodeGraph definition/callers/callees/impact: `analysis/code-context/codegraph-baseline.md` — probe 호출자
  reclaim·Dial·판정 표, verify 호출자 reclaim 1, reclaim 호출자 Start 1, impact 5(패키지 내)
- CodeGraphContext 후보와 evidence reconciliation: CGC 심볼 없음(not-applicable), CodeGraph 의
  `conn.Close` 오해석 1건 — `evidence-reconciliation.md`
- 기존 동작 파악 근거: HEAD `54004f44` AST 4 번들, a108 테스트 3 파일, 소비자 grep(console.go:413,
  httpapi.go:299)
- Function Logic Map / Branch Test Map: `analysis/function-logic/internal-strategyprojectionrpc--{projectionsocketaccepts,
  reclaimstalecontroldirectory,verifystalesocketshape,dial}/`
- upstream 상속 테스트 영향: no — 이 패키지는 TossOS 신규(a108)
- 실패 테스트 선행 작성: yes — `a113_the_probe_proves_death_test.go` Start 수준 RED 관측(기동이 받아들여짐 =
  산 socket 탈취), 판정 표 2행 RED, 회수 전용 probe 테스트는 컴파일 RED(`undefined: staleProjectionSocketAccepts`)
- 설정·DB·journal 변경과 rollback: 없음 — 회수 판정만. rollback = 커밋 revert(디스크 형식 무변경)
- 안전 불변식 §0 위반 여부 검토: 통과 — 주문·손절·원장 무접촉, 사망 판정을 보수 방향(추정 제거)으로만 바꾼다.
  조회 클라이언트는 여전히 디스크를 바꾸지 않는다.

## §1 구현 후 리뷰 (2026-09-25 — ed3cb1d7)

구성: 독립 서브에이전트(Opus, 별도 컨텍스트, 읽기 전용 — gstack `review` 의 pre-landing 관점을
대체. codex 적대 패스는 이 자율 세션에서 쓰지 않았다). 리뷰어는 저장소 사본에서 뮤테이션을 재현했다
(대조군 rc=0, 원장 N1~N7 전부 일치).

| id | 발견 | 판결 → 반영 |
| --- | --- | --- |
| P1-1 | 순서 핀이 호출만 세서 재확인 **결과를 버리는** 변이 R1·R2 가 생존 | 수용 → chmod seam(`chmodStaleSocket`, 운영 값 `os.Chmod` 포인터 핀) + chmod 중 이름 교체 행동 테스트 + 「재확인이 반환을 가른다」 구조 핀. R1·R2 사망 |
| P2-1 | chmod ENOENT·post-check `!dead` 분기 미시험(R3 생존) | 수용 → chmod 실패 두 갈래 테스트(ENOENT=사망·EPERM=생존). R3·R3b 사망 |
| P2-2 | Accept 가 probe 의 연결을 받아 단정이 과장 | 수용 → 두 연결을 받는다 |
| P2-3 | control 디렉터리 두 번 stat | 기록 → issues R4 (a108 원본, 편집 지점 밖) |
| P2-4 | "still alive" 가 「증명 못 함」도 포함 | 수용 → 문구에 "(or its death could not be proven)" |
| P2-5 | socket uid 절은 비root 로 못 죽임 | 기록 → 원장 알려진 생존(a108 이래) |

검증 OK(리뷰어): 9개 unix GOOS `go vet`·windows build clean, `-count=5` 무flake, AST 핀의 상대 경로가
패키지 디렉터리에서 성립, 산 socket unlink 경로 없음, 죽은 잔재 영구 거부 없음, chmod 가 group/other 를
넓히지 않음, `Dial` 은 chmod 하지 않음.

2판 뮤테이션: 14/14 사망(`mutation-ledger.md`).

## 검증 기록 (2026-09-25)

- `go test -race -count=1 -v ./internal/strategyprojectionrpc/...` → ok, 최상위 47 · PASS 줄 71 · FAIL/SKIP 0 (36237b22)
- `go test -race -count=1 -run 'A108|A109|Strategy|Reattach|Daemon|Projection' ./cmd/tossctl/` ok · `internal/console -run Strategy` ok
- `make vet` · `make lint` rc 0
- `make test` 1판(ed3cb1d7): FAIL 2 — `internal/app/engine TestTheDriverFoldsAndAdoptsInOneCycle`("database or disk is
  full" — 루트 파일시스템 100%, 여유 968M→72M 실측)·`tools/a112-mb-us-source` 빌드 신원 불일치. 둘 다
  strategyprojectionrpc 에 의존하지 않는다(`go list -deps` 0). 전자는 GOCACHE 를 옮겨 단독 재실행 ok.
  2판(a114 의 1244be95 — a113 코드 포함): **ok 99 · FAIL 0**.
- `check_analysis.py` rc 0(착지 `36237b22`, 새 clone 에서 재확인).

## 완료 게이트 — Manager 실행 (2026-09-25)

- 실행자: Manager(Fable, 세션 tossos-d6). 작성자(Opus 팀메이트)와 분리된 검증 패스.
- 대상 커밋: `692adef3`(2.1 gate 줄 체크 + tracker 재생성) — 착지 기록은 `landed-commit.txt` 그대로.
- 장소: 저장소 밖 격리 워크트리 `TossOS-worktrees/archive-batch1`(detached `692adef3`, 실행 전후 `git status` 0줄) — 주 워크트리는 병행 세션·팀메이트의 미커밋 편집이 있어 판정을 오염시킨다.
- `make sdd-sync` rc 2 ×2: CodeGraph 단계는 완료(`Done`), CodeGraphContext 갱신이 kuzu `Could not set lock on file`(다른 프로세스가 DB 보유)로 실패 — advisory. fingerprint 는 기록됐다: `make sdd-check` **rc 0**(`/tmp/claude-1000/gate-sdd-check.log`).
- `make gate CHANGE=a113-the-projection-probe-proves-death` → **GATE PASS, 11/11, rc 0** (`/tmp/claude-1000/gate-a113.log`, 445줄): 1 tasks.md · 2 미완료 0 · 3 짝 없음 · 4 review.md · 5 Function Logic Map(착지 창) · 6 sdd-check · 7 make test · 8 make test-seams · 9 make test-race · 10 make vet · 11 make validate — 전부 OK.
- archive: 게이트 통과 뒤 수행 예정(사용자 지시로 Opus 팀메이트가 — 2026-09-25 Opus 주간 한도로 대기). archive 전까지 tasks 2.1 의 "후 archive" 는 미수행 상태다.
