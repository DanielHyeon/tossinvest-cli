# 영속 세션 컴포넌트 Manager 한정 검증

2026-11-06. 판정: **P10 구현 범위 PASS, P11 게시 진행 승인**. 전체 a169 완료·측정·활성화·archive 승인이 아니다.

## 검증 기준과 직접 확인

- 계약/연구/독립 설계 검토/Pre-Edit는 `1a53b8917`에서 소스보다 먼저 고정됐다. 구현은 `ec254d8ca`, 독립 재리뷰 기준은 `8e8b6db24`, 최종 시험 보강은 `80ca5dbfb46b480a6148d281f6e2878846d95e8c`다.
- Manager는 `session_linux.go` 전체와 unsupported/types, 초기·보강 시험, 상태 계약, 실제 diff 및 원문 결과를 읽었다. 코딩·시험 실행은 Terra에 위임했다. 운영 소스 SHA-256은 `ce79892b33822c1f670eab66bf9d2bffc239c9e972b43a64cc9e1491eed8d29a`로 보강 전후 동일하다.
- base `232d6b86c` 대비 Go 변경은 신규 `internal/shadowsession` 파일뿐이다. 기존 Go 함수 수정과 외부 production caller는 0이다. HTTP/주문/관측 원장/운영 설정은 바꾸지 않았다.
- 허용 전이는 부재 또는 paused에서 active, active에서 paused/aborted/completed다. active·aborted·completed 및 손상 기록은 Begin에서 거절한다. Begin 오류는 nil handle, 종료 전이 시도는 결과와 관계없이 handle 폐기다.
- root/파일 identity·권한·ext4 검증, lock 수명, append 및 file/directory Sync, canonical JSONL의 전체 프레임 검사, context·cleanup 오류 의미를 설계와 대조했다. 최초 lock 생성은 flock 전, state 접근은 flock 후다. 오류가 반환돼도 이미 저장된 paused는 다음 실행에서 재개될 수 있다는 계약을 유지한다.

## 시험과 리뷰 판정

- 최초 RED는 API 미정의 컴파일 오류이며 행동 RED로 주장하지 않는다. 행동 검증은 GREEN·fault·subprocess·변이 근거를 구분해 보존한다.
- 독립 초기 리뷰의 개인 홈 경로 의존 P1과 kill 뒤 잠금 미해제도 통과할 수 있던 시험 문제를 수정했다. 명시 ext4 fixture parent가 잘못되면 실패하고, 기본 temp가 지원되지 않아 skip된 결과를 coverage로 사용하지 않는다.
- `8e8b6db24`에 대한 독립 normal 및 `tossos_testseams` tagged race 패키지 검증 PASS, skip 0. 원문 `/tmp/a169-session-rereview.md`와 `/tmp/a169-rereview-5eP3xeos/{normal.log,race.log}`를 직접 확인했다. 소스/시험/실행 파일 해시는 검증 전후 일치했다.
- Manager가 계약 §5에서 빠진 pause 후 child kill 및 동시 재개 시나리오를 추가 요구했다. `80ca5dbfb`의 두 시험을 독립 normal/tagged race로 증분 검증하여 PASS·skip 0. 원문 `/tmp/a169-session-final-increment-review.md`, `/tmp/a169-final-increment-xAZOKcBG/{normal.log,race.log}`를 확인했다. 이전 패키지 PASS와 최종 증분 PASS의 합이며 최종 SHA에서 패키지 전체를 다시 돌렸다는 주장이 아니다.
- 최초 동시 시작/paused 재개는 두 child의 lock-open barrier로 경합시킨다. kill 뒤 직접 flock을 획득해 잠금 해제와 active 거절을 분리한다. Begin/terminal 오류 전후 실제 기록 상태와 재시작, inode 치환, 정상 paused prefix 뒤 partial EOF, 실제 non-ext4 거절도 시험했다.
- 선택 변이는 state Sync 생략, active/aborted 재개, flock 생략, inode 무시, checksum 무시를 검출했다. checksum 변이 최초 두 번 생존 기록도 보존하며, paused 손상 fixture 보완 뒤 검출로 처리했다. 새 함수 AST와 `related_tests` 연결은 분기 전체 실행 증거가 아니며 `test:null`은 미확인으로 남긴다.
- 구현자와 독립 리뷰어의 gstack skill/checklist 적용은 한정 수동 리뷰다. 전체 자동 gstack workflow 또는 전체 gate PASS로 확대하지 않는다. 최종 blocker P0/P1 발견 없음.

## 남은 범위와 게시 조건

- impeccable: N/A — backend/documentation-only. grill-me/grilling에서 기존 Q2 정상 재개와 사용자 과거 gap 제외 결정을 보존했다. search-first 근거와 dependency 선택은 `daily-session-research.md`에 있다. 새로운 외부 의존성은 없다.
- private fault seam과 process kill은 실제 저장장치 전원 손실 증거가 아니다. Pause 도중 kill 전 단계를 열거하지 않았다. root 보존·동일 identity·협조 writer·성공한 Sync 내구성 전제는 그대로다.
- 실제 HTTP drain 및 wire 연결, KST 거래일 권위, 엔진 identity, universe 복원, runtime 호출자, 실제 그림자측정과 운영 저장장치 검증은 후속이다. 전체 시험/전체 gate는 실행하지 않았다. 전체 change의 미완 체크를 닫거나 archive하지 않는다.
- Terra는 독립 보고서와 원시 로그를 새 evidence 경로에 원문 보존하고 P10/P11 기록·PM 파생 파일을 갱신한 뒤 scoped strict/PM/diff를 확인한다. 기존 소스/시험이 바뀌면 이 승인은 재검토한다.
- 원격 main은 검증 당시 `232d6b86c024dfd71cc8ffe6fd0fbba342b3e9ec`. 별건 로컬 main의 미게시 `ca7a3f4f7` 및 사용자 변경은 이번 게시에 포함하지 않는다. 원격 상태를 다시 확인해 fast-forward 가능한 a169 브랜치만 게시한다. 로컬 main이 갈라져 있으면 reset/rebase/force하지 않고 보존·보고한다.
- `git -c sdd.captureCommits=false`는 커밋 관측 이벤트와 파생 비동기 색인 전체를 생략하는 지원 옵션으로 사용했다. 검증용 pre-commit 훅은 유지했다. 전역 설정·훅 파일 변경은 없다.
