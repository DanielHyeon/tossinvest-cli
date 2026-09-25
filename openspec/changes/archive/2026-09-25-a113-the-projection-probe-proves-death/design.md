# a113 설계 — projection probe 도 사망을 증명한다

작성 2026-09-25. base `54004f44`. 분기 주장은 전부 `analysis/function-logic/` 의 AST 번들
4개(`projectionSocketAccepts`·`reclaimStaleControlDirectory`·`verifyStaleSocketShape`·`Dial`)를
먼저 만든 뒤 그 열거를 근거로 썼다(FLM-before-claiming). 원형은 a109 §1-fix F1 + §2b.3 G7 의
`internal/positionpolicyrpc/private_staging_unix.go` `privateSocketAccepts` 다.

## 병의 재확인 (AST 열거 기준)

- `projectionSocketAccepts`(:387–400)는 분기 3개다. B3(:396)은 connect 가 실패했을 때
  `perm&0o200 == 0` 이면 **사망**으로 읽는다 — 묻지 않고 추정한다.
- 호출자는 둘이다(CodeGraph callers): 회수 `reclaimStaleControlDirectory` B12(:293)와 조회
  클라이언트 `Dial` B3(:417). 회수 쪽에서 B3 이 틀리면 **수락 중인 socket 을 지운다** — 외부 chmod
  로 owner 쓰기 비트만 깎인 산 socket 이 그 모양이다(a109 A1 P1-A 의 원형판, issues I1).
- `Dial` 은 B2(:409)에서 정확-0600 을 먼저 요구하므로 `Dial` 경로에서 B3 은 Lstat(:408)과
  connect(:417) 사이의 권한 변경 경합에서만 발동한다.

## D1 — chmod-then-probe 는 회수 전용 함수가 진다

형제(a109)는 probe 하나를 회수만 불렀으므로 그 함수 안에 chmod 를 넣었다. 이 패키지의 probe 는
`Dial` 도 부른다. 같은 함수에 chmod 를 넣으면 **조회 클라이언트(콘솔·httpapi)가 엔진 socket 의
권한을 바꾸는** 새 부작용이 생긴다 — 그래서 나눈다:

- `projectionSocketAccepts(socketPath) bool` — 순수 질문으로 남는다. **B3(owner-write 추정)을
  삭제한다.** 사망은 연결 거부·파일 부재 둘뿐이다. `Dial` 은 편집하지 않는다(0줄).
- `staleProjectionSocketAccepts(socketPath, before os.FileInfo) bool` — 새 파일
  `transport_probe_unix.go` 의 새 leaf. 회수만 부른다(freeze P1-2 반영 순서):
  ① `before == nil` 이면 생존 — 검증하지 않은 이름은 chmod 도 하지 않는다
  ② chmod **전** 재-`os.Lstat` + `os.SameFile(before, pre)` — 부재면 사망, 다른 파일·그 밖의 오류면
  생존(검증하지 않은 inode 를 chmod 하는 창을 좁힌다) ③ `os.Chmod(0o600)` — 실패면 부재
  (ErrNotExist)만 사망, 그 밖은 생존 ④ chmod **후** 재-`os.Lstat` + SameFile — 다르면 생존
  (a109 G7) ⑤ `projectionSocketAccepts` 로 묻는다.
- `verifyStaleSocketShape` 는 `(os.FileInfo, error)` 를 돌려준다 — 검증한 그 inode 를 ②④에 넘기기
  위해서다(원형 `verifyStalePrivateSocket` 과 같은 모양). freeze P2-1 반영: uid·nlink 를 **같은
  Lstat 의** `info.Sys().(*syscall.Stat_t)` 에서 읽는다(원형 `validateOwnerAndLinks(info)` 와 같다) —
  두 번 stat 하면 돌려주는 inode 와 uid·nlink 를 본 inode 가 다를 수 있다. 조건·오류 문구 무변경.
- `reclaimStaleControlDirectory` 는 B11·B12 두 줄만 바뀐다.

chmod 가 안전한 근거(원형과 동일): 회수 경로에서 이 호출에 닿으려면 이미 우리 uid 소유·0700
디렉터리 안·비symlink·nlink 1 을 통과했고(B1·B2·B11), chmod 결과가 0600 이라 **group/other 는 결코
넓어지지 않으며**, owner 권한은 발행 계약 그 자체로 **복원**된다(`listenPrivateSocket` 은 chmod
0600 뒤에만 rename). freeze P2-4 정정: 「접근이 넓어질 수 없다」는 owner 에게는 거짓이다 — 0400 산
socket 은 거부 경로에서 0600 으로 남는다. 그것은 외부 chmod 로 깎인 우리 socket 을 우리 계약으로
되돌린 것이다.

대안과 기각:

- 한 함수에 chmod 를 넣고 `Dial` 도 그것을 부른다 — 조회 클라이언트의 디스크 부작용. 기각.
- B3 만 지우고 chmod 없이 둔다 — 죽은 0500/0400 잔재(UMask=0277 배포의 pre-chmod)에 connect 가
  EACCES → 생존으로 읽혀 **매 부팅 영구 거부**가 된다(a108 이 지운 병의 재발). 기각.
  `TestStartRecoversFromUnwritableSocketLeftover` 가 이것을 잡는다.

## D2 — 뮤테이션 원형판(F1-N1)과 핀

- RED(행동): 수락 중인 socket 을 0400 으로 깎고 descriptor 를 둔 채 `Start` → 거부해야 하고,
  socket 은 그대로 있어야 하며 여전히 연결을 받아야 한다(a109
  `TestReclaimRefusesALiveSocketWhoseOwnerWriteBitWasStripped` 원형판). 편집 전 코드는 이것을
  지우고 두 번째 서버를 세운다.
- 판정 표(`TestProjectionLivenessClausesEachDecideOnTheirOwn`)에 "쓰기 비트가 깎여도 수락 중이면
  생존" 행을 더하고, 회수 전용 함수의 판정 표를 새로 둔다(부재·거부·수락·죽은 0500·산 0400·
  before nil·**바뀐 파일**(a109 `TestTheProbeRefusesASocketThatChangedUnderIt` 원형판)).
- 기존 행 "owner 쓰기 비트가 없다"(죽은 0500 → false)는 순수 probe 에서는 더 이상 참이 아니다
  (EACCES = 생존). 그 행은 **회수 전용 함수의 표로 옮긴다** — 그 사실을 보장하는 것은 이제 chmod 다.
- 뮤테이션(적용→빨강→원복→초록을 실제로 잰다): N1 B3 재도입(순수 probe) · N2 chmod 제거 ·
  N3 chmod 후 SameFile 재확인 제거 · N4 F1-N1 원형판(chmod 제거 + B3 재도입 = 병의 복원) ·
  N5 chmod 모드를 0o666 으로(freeze P1-3 — 산 socket 이 group/other 쓰기로 남는다) ·
  N6 회수가 verify 의 FileInfo 대신 새 Lstat 을 넘긴다(freeze P2-2 — 행동으로 못 가르므로 AST 핀).
- freeze P1-3: RED 테스트와 판정 표의 산 0400 행은 probe 뒤 권한이 **정확히 0600** 인지 단정한다.
- freeze P2-3: RED 테스트는 거부 사유("still alive")까지 단정한다.
- root 는 DAC 를 우회해 EACCES 가 안 나므로 해당 행은 a108 관례대로 `t.Skip` 한다.

## 비목표·선언된 생략

- staging(`.s-*`) socket 엔트리의 probe(a109 F5 의 대응물): a108 회수는 staging 을 probe 없이
  지운다. proposal 이 "그 밖의 a108 회수·발행 의례는 무변경"이라 범위 밖이다 — 수락 중인
  staging socket 은 우리 발행이 rename 전 수 μs 동안만 만들고 journal flock 이 두 번째 엔진을
  막는다. 후속 후보로 issues.md 에 적는다.
- staging 12자→11자, descriptor 발행 fold: a109 가 이미 선언한 생략 그대로.
- `Dial` 편집: 코드 0줄. 주석(:410–416 「회수와 같은 원시」)만 정정한다 — 이제 회수는 그 원시
  앞에 권한 복원을 둔다(freeze P2-6).
- `os.Chmod` 는 symlink 를 따라간다(fchmodat, NOFOLLOW 없음): 검증 뒤 이름이 symlink 로 바뀌면
  SameFile 이 알아채기 전에 대상이 chmod 된다. 같은 uid 신뢰 경계 안이며 원형(a109)이 수용한
  것과 같다 — ② 의 chmod 전 재확인이 창을 좁히고, issues.md 에 기록한다(freeze P2-5).

## 테스트 전략

`go test -race -count=1 ./internal/strategyprojectionrpc/...` 전체, 그리고 소비자 재부착 회귀로
`./cmd/tossctl/ -run 'A108|A109|Strategy|Reattach|Daemon'`. 전체 `make test` 는 change 끝에 1회.
