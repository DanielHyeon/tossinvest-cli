# a113 구현 중 발견 — 분류와 처리

**blocking 0건.**

## ② safe local

### S1. probe 는 `Dial` 과 공유된다 — chmod 는 회수 전용 새 함수에 둔다 (2026-09-25)

등록 문서는 "projectionSocketAccepts 를 chmod-then-probe 로 교체"라고 적었다. AST 열거
(`Dial` B3 :417)가 그 함수를 조회 클라이언트도 부른다는 것을 보였다. 그대로 교체하면 콘솔·httpapi
가 엔진 socket 을 chmod 한다. 의도(회수의 사망 증명)는 명백하므로 회수 전용 함수로 나눴다(design D1).

## 잔존·후속 후보

### R1. staging(`.s-*`) socket 은 probe 없이 지운다 (a109 F5 의 대응물)

a108 회수는 staging 엔트리를 probe 없이 지운다. 이 change 의 범위(proposal "그 밖의 a108 의례
무변경") 밖이다. 방어는 journal flock(엔진 싱글턴)이고, spec 새 SHALL 은 최종 이름 socket 으로
좁혔다(freeze P1-1). 후속 change 후보.

### R2. `os.Chmod` 는 symlink 를 따라간다 (freeze P2-5)

검증 뒤 이름이 symlink 로 바뀌면 SameFile 재확인 전에 대상이 chmod 된다. 같은 uid 신뢰 경계 안이고
원형(a109)이 수용한 것과 같다. chmod **전** SameFile 재확인이 창을 좁힌다.
`unix.Fchmodat(…, AT_SYMLINK_NOFOLLOW)` 는 대안이 아니다: x/sys 는 flag 가 있으면 `fchmodat2`
(Linux ≥ 6.6)를 부르고, 없으면 **EOPNOTSUPP** 를 돌려준다(`golang.org/x/sys/unix/syscall_linux.go`
`Fchmodat` :67–81 실측). 그 오류를 생존으로 읽으면 6.6 미만 커널에서 모든 죽은 잔재가 매 부팅
영구 거부가 된다 — 이 change 가 막으려는 모양의 반대편 사고다.

### R3. 형제 패키지의 이식 원형 주석 (freeze P2-6)

`internal/positionpolicyrpc/private_staging_unix.go:5-8` 는 "a108 확정 코드 — 원형은 수정하지
않는다"고 적는다. a109 시점의 진술로는 참이지만 이제 원형도 chmod-then-probe 다. 파일 표면 밖이라
고치지 않았다 — 다음에 그 파일을 만지는 change 가 한 줄 정정한다.
