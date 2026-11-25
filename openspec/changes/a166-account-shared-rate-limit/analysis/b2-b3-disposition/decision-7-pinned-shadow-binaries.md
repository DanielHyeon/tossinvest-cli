# a166 결정 7 — `go run` 참여자(a128 · a140) 고정 바이너리 전환 준비 (2026-11-26)

`proposal.md` 「사람 결정 기록」 결정 7(① 전환 전에 고정 빌드 바이너리로 바꾼다)의 **저장소 쪽 준비**만 기록한다. 이 커밋은 `run-probe.sh` 한 파일만 바꾸고, 운영(unit · 타이머 · 바이너리 설치)은 바꾸지 않았다. 아래 설치 · 전환 절차는 사람 승인 뒤 Manager 가 실행할 초안이다. `<HOME>` 은 홈 디렉터리, `<primary>` 는 저장소 primary 체크아웃 디렉터리다(같은 디렉터리 다른 문서와 같은 표기).

왜 필요한가: `go run` 은 실행 순간 메인 checkout 의 작업 트리에서 빌드한다(`proposal.md` 「B3-4 처분 선택지」 그 밖 #6 a128 · #8). 그래서 「게시된 정확 SHA」를 고정하지 못한다.

## 무엇을 바꿨나 — `openspec/changes/a128-jev-shadow-judgment-probe/run-probe.sh`

`cd "$ROOT"` 와 마지막 `exec go run …` 줄 사이에 6줄을 넣었다(+6 −0). 기존 줄은 한 글자도 바꾸지 않았다.

- `A128_PROBE_BIN` 이 비어 있지 않으면: 그 경로가 실행 가능한 정규 파일(`[ -f ] && [ -x ]`, 심링크는 따라감)인지 본다.
  - 맞으면 `TYPESAFE_API_KEY="$KEY" exec "$A128_PROBE_BIN" run --market "$MARKET" "$@" >>"$LOG" 2>&1`. `go run ./tools/a128-jev-shadow-probe` 자리만 바뀌고 인자 · 리다이렉트 · 키 전달 · `cd "$ROOT"` 는 같다.
  - 아니면 `run-probe.sh: A128_PROBE_BIN 이 실행 가능한 정규 파일이 아님: <경로>` 를 stderr 에 쓰고 exit 1. `go run` 으로 떨어지지 않는다(fail-closed).
- 변수가 없거나 빈 문자열이면: 기존 마지막 줄(`exec go run …`)이 그대로 실행된다.
- 판정 순서: `.env` 키 확인(기존) → `cd "$ROOT"`(기존) → 바이너리 확인(새). 상대 경로를 주면 `$ROOT` 기준으로 해석된다. unit 에는 절대 경로만 쓴다.

## 검증 (2026-11-26, 네트워크 · 실키 없음)

방법: 세션 scratchpad 아래 임시 디렉터리 `<tmp>` 에 저장소와 같은 깊이의 가짜 트리를 만들었다. `<tmp>/openspec/changes/x/run-probe.sh`(편집본 사본), `<tmp>/.env` 에 `JEV_AI_API_KEY=dummy`. PATH 맨 앞에 `go` stub(호출 흔적을 `go-trace` 에 남김)을 두었다. 바이너리 stub 은 인자 · `TYPESAFE_API_KEY` 길이 · cwd 를 찍는다(값은 안 찍음). 호출은 `cd /` 에서 `run-probe.sh KR <tmp>/out/log --session "09:00-15:30" --pool 45`. 부모 셸의 `A128_PROBE_BIN` 은 미설정 확인(`printenv` rc=1). 실제 저장소 · 실제 `.env` · 실제 probe 는 실행하지 않았다.

원문(`<tmp>` 는 실제 경로 `…/scratchpad/d7.D0KQLZ` 를 줄인 것):

```text
[a-stub] rc=0
  log: bin-stub args=[run --market KR --session 09:00-15:30 --pool 45] keylen=5 cwd=<tmp>
  stderr:
  go-trace: <none>
[b-missing] rc=1
  log: <none>
  stderr: run-probe.sh: A128_PROBE_BIN 이 실행 가능한 정규 파일이 아님: <tmp>/nope
  go-trace: <none>
[b-noexec] rc=1
  log: <none>
  stderr: run-probe.sh: A128_PROBE_BIN 이 실행 가능한 정규 파일이 아님: <tmp>/noexec
  go-trace: <none>
[b-dir] rc=1
  log: <none>
  stderr: run-probe.sh: A128_PROBE_BIN 이 실행 가능한 정규 파일이 아님: <tmp>/adir
  go-trace: <none>
[c-unset] rc=0
  log: go-stub args=[run ./tools/a128-jev-shadow-probe run --market KR --session 09:00-15:30 --pool 45] keylen=5 cwd=<tmp>
  stderr:
  go-trace: GO_CALLED args=run ./tools/a128-jev-shadow-probe run --market KR --session 09:00-15:30 --pool 45
[c-empty] rc=0
  log: go-stub args=[run ./tools/a128-jev-shadow-probe run --market KR --session 09:00-15:30 --pool 45] keylen=5 cwd=<tmp>
  stderr:
  go-trace: GO_CALLED args=run ./tools/a128-jev-shadow-probe run --market KR --session 09:00-15:30 --pool 45
bash -n rc=0
```

- (a) 바이너리 경로: 인자 `run --market KR <추가 플래그>`, 키 전달됨(길이 5 = `dummy`), cwd = `<tmp>`, `go` 미호출.
- (b) 없는 경로 · 실행 권한 없는 파일 · 디렉터리: 셋 다 rc 1 · 오류 한 줄 · `go` 미호출 · 로그 없음.
- (c) 미설정 · 빈 문자열: `go run ./tools/a128-jev-shadow-probe run --market KR …` 로 호출됨.
- `shellcheck`: PATH 에 없음 — 그 채널은 안 봤음.
- 인자 경계(`$*` 로 합쳐 찍음)는 따로 재지 않았다. `"$@"` 전달은 기존 줄과 같은 형태다. → Manager 실측으로 닫음(아래 「안 본 것」).

## 두 도구가 cwd(저장소 루트)에 기대는 상대 경로 — `rg` 문자열 · 플래그 기본값 수준

있음.

- a128: `tools/a128-jev-shadow-probe/main.go:47` `defaultLedgerDir = "openspec/changes/a128-jev-shadow-judgment-probe/analysis/shadow-ledger"`. `--out` 기본값은 `""`(`main.go:93`)이고, 이 상수를 쓰는 자리는 `profile.go:64` · `profile.go:129` · `:131` · `:133`(`-initial` · `-compare` 접미사) · `main.go:271`(`report` 의 `--ledger`) 다. `profile.go:96` 이 `filepath.Abs(path)` 로 절대화한다 → cwd 기준. 빈 `--out` 에서 어느 분기로 이 상수가 고르는지는 함수 내부라 이 문서에서 근거로 삼지 않는다(미확인).
- a140: `tools/a140-breakout-signal-shadow/main.go:44` `defaultLedgerDir = "openspec/changes/a140-breakout-signals-are-measured-in-shadow/analysis/shadow-ledger"`. `run` 의 `--out` 기본값(`main.go:79`)과 `report` 의 `--ledger` 기본값(`main.go:292`)이 이 상대 경로다. a140 unit 은 `--out` 을 주지 않는다.
- 그 밖: 두 도구의 비시험 `*.go` 에서 `os.Open/ReadFile/Create/MkdirAll/Stat("…")` 리터럴 · `//go:embed` · `os.Getwd` · `os.Executable` · `runtime.Caller` 0건. `--config-dir`(a128 · a140) · `--evidence-db`(a140)는 unit 이 절대 경로로 준다. `tools/a128-jev-shadow-probe/profile_eval.py:13` 의 `openspec/…` 는 Python 보조 스크립트라 바이너리와 무관하다.

결론: 고정 바이너리도 **cwd = 저장소 루트** 여야 원장이 지금과 같은 자리에 쌓인다. 세 unit 은 `WorkingDirectory=<primary>` 를 이미 갖고 있고, a128 은 `run-probe.sh` 가 `cd "$ROOT"` 를 한다. 전환 때 `WorkingDirectory=` 를 지우지 않는다.

## 설치 절차 (사람 승인 뒤 Manager 실행 — 이 커밋은 실행하지 않음)

1. 고정 SHA `<S>` 를 정한다(`2158b21fe` 이후 origin/main 중 하나). **primary 체크아웃**에서 `git -C <primary> rev-parse HEAD` = `<S>` 전체, `git -C <primary> status --porcelain` 빈 출력을 확인한다.
2. 디렉터리: `install -d -m 0700 <HOME>/.local/share/tossos-a128/bin <HOME>/.local/share/tossos-a140/bin`.
3. 빌드(`<primary>` 에서):
   - `go build -trimpath -buildvcs=true -o <HOME>/.local/share/tossos-a128/bin/a128-jev-shadow-probe ./tools/a128-jev-shadow-probe`
   - `go build -trimpath -buildvcs=true -o <HOME>/.local/share/tossos-a140/bin/a140-breakout-signal-shadow ./tools/a140-breakout-signal-shadow`
4. 합격 조건 · 기록(영수증에 원문으로):
   - `sha256sum` 두 파일.
   - `go version -m <파일> | grep -E 'vcs\.(revision|modified)'` 에 `vcs.revision=<S 전체>` 와 `vcs.modified=false` 가 **있어야** 한다. 줄이 없거나 값이 다르면 그 설치는 무효다. `proposal.md` 「재실측 체크리스트」 7 이 `vcs.revision` 조상 관계를 요구하기 때문이다.

**정정(2026-11-26)** — 처음 판(`2158b21fe`)은 「깨끗한 worktree 를 만든다: `git -C <primary> worktree add --detach <wt> <S>`, `git -C <wt> status --porcelain` 이 비어 있음을 확인」 뒤 `<wt>` 에서 `go build -trimpath -o …` 로 빌드하고, 끝나면 `git worktree remove <wt>` 로 지우라고 적었다. Manager 시험 빌드(2026-11-26 04:2x KST, go1.26.5, 산출물은 scratchpad 에만, 설치 0)에서 이 방법은 VCS 스탬프를 남기지 않았다.

- 연결 worktree `~/tossos-work/a166-b2-b3`(HEAD `2158b21fe`, `git status --porcelain` 빈 출력, `.git` 은 `gitdir: …/.git/worktrees/a166-b2-b3` 파일):
  - `GOFLAGS=-trimpath go build -o … ./tools/a128-jev-shadow-probe` → rc 0, `go version -m` 에 `vcs.*` 줄 0.
  - `go build -trimpath -buildvcs=true …` → rc 0, `vcs.*` 줄 0. 산출물 sha256 이 스탬프 없는 빌드와 같다(`bd0b7199…`).
  - a140 도 worktree 빌드는 rc 0, `vcs.*` 줄 0.
- primary 체크아웃(HEAD `2158b21fe`, 깨끗함)에서 `go build -trimpath -buildvcs=true -o … ./tools/a128-jev-shadow-probe` → rc 0, `vcs=git` · `vcs.revision=2158b21fedc317c1fa7643a4f2e154b815e62ce4` · `vcs.time=2026-11-25T19:18:04Z` · `vcs.modified=false`. a140 은 primary 에서 재지 않았다.
- `-buildvcs=true` 가 연결 worktree 에서 오류 없이 스탬프를 빼는 원인은 **미확인**이다. 그래서 4단계의 「줄이 있어야 한다」를 합격 조건으로 둔다.

## 전환 절차 초안 (사람 승인 뒤 Manager 실행)

창: a128-kr 이 돌지 않는 시간 — 평일 15:30 이후 ~ 08:50 이전. 다만 unit 이 `RuntimeMaxSec=9h`(08:50 기동 → 최대 17:50)이므로 시작 전에 `systemctl --user is-active tossos-a128-kr.service` 가 `inactive` 인지 본다. 이 확인은 이 문서가 덧붙인 것이다.

1. a128 두 unit(`tossos-a128-kr` · `tossos-a128-us`)에 drop-in `<HOME>/.config/systemd/user/<unit>.service.d/a166-pinned-bin.conf`:

   ```ini
   [Service]
   Environment=A128_PROBE_BIN=%h/.local/share/tossos-a128/bin/a128-jev-shadow-probe
   ```

   unit 본문은 그대로 둔다(`run-probe.sh` 가 env 를 읽는다). 이 커밋이 origin/main 에 있고 primary 체크아웃이 그 이후여야 drop-in 이 효과를 낸다. 그 전 `run-probe.sh` 는 변수를 무시하고 `go run` 을 한다. primary 는 Manager 가 a128-kr inactive 를 확인한 뒤 `2158b21fe` 로 fast-forward 했다(2026-11-26).
2. a140 unit(`tossos-a140-kr`): 먼저 `cp <unit> <unit>.bak-<UTC>`. `ExecStart` 안의 `go run ./tools/a140-breakout-signal-shadow` 를 `%h/.local/share/tossos-a140/bin/a140-breakout-signal-shadow` 로 바꾼다. 나머지 인자 · 리다이렉트 · `WorkingDirectory` · `Environment=PATH` 는 그대로 둔다.
3. `systemctl --user daemon-reload`. 기동 · 정지 · 타이머 enable 은 하지 않는다(a128-us · a140 타이머는 disabled 그대로).
4. 확인: `systemctl --user cat` 세 unit 을 영수증에 원문으로 남긴다. `systemctl --user show -p Environment tossos-a128-kr.service` 에 `A128_PROBE_BIN=` 이 있어야 한다. 첫 회차 로그(`/tmp/a128-kr-<MMDD>.log`)에서 정상 기동을 본다.

되돌리기: drop-in 파일 삭제, a140 unit 은 `.bak-<UTC>` 복원, `daemon-reload`. 실행 중인 프로세스는 다음 기동부터 바뀐다.

## 영향

- a184 짝: a128-kr unit 주석 「사용자 결정(2026-11-12): 판단 수 30 → 45(--pool 45). a184 사전등록 2판(7bfaf5ee2)의 선택 규칙 …random45… 과 짝임」. 전환은 `ExecStart` 의 `--pool 45` 를 건드리지 않는다. 다만 고정 SHA `<S>` 의 a128 코드가 그 짝을 만든 코드와 같은지는 설치 때 `<S>` 를 고르며 확인해야 한다(이 문서는 비교하지 않았다).
- 고정 빌드는 이후 main 의 a128 · a140 변경을 **자동으로 따라가지 않는다**. 그 변경을 측정에 넣으려면 설치 절차를 새 SHA 로 다시 밟고 영수증을 남긴다. 반대로 변수가 남아 있는 한 primary 체크아웃의 편집이 측정에 섞이지 않는다.
- `A128_PROBE_BIN` 이 잘못된 경로면 그 회차는 rc 1 로 끝나고 측정이 빠진다(go run 대체 없음). 첫 회차 확인(전환 4단계)이 그것을 잡는다.
- 문서 참조: `*.md` 에서 `run-probe.sh` · `tossos-a128-kr` · `tossos-a140-kr` 를 언급하는 파일 6개(archive 제외: `docs/handoff/2026-10-19-manager-handoff.md` · a063 분석 1 · a128 `review.md` · `tasks.md` · a166 `proposal.md` · `host-binding-evidence-2026-11-25.md`)는 모두 과거 기록이라 고치지 않았다.

## openspec validate (2026-11-26, 이 편집 뒤)

- `openspec validate a166-account-shared-rate-limit`: rc 0, valid, 경고 2(`order-execution/spec.md` 기존 경고 — 고치지 않음).
- `openspec validate a128-jev-shadow-judgment-probe`: rc 0, valid, 경고 0.

## not-applicable

- FLM/BTM: Go 함수 본문 수정 0(바꾼 파일은 셸 스크립트 하나).
- 변이 원장 · 다중 적대 리뷰: High-risk 경로 · 게이트 판정 변경이 아니다(비례 원칙). 그림자 측정 래퍼이며 주문 경로 밖이다.
- spec delta: 생산 동작 변경 0. 새 OpenSpec change 를 만들지 않고 a166 처분 문서에 기록한다.

## 안 본 것

- ~~실제 `go build` · `go version -m` 출력(설치하지 않았으므로).~~ → Manager 시험 빌드로 잼(설치 0). 연결 worktree 빌드는 `vcs.*` 0 줄, primary 빌드는 `vcs.revision` · `vcs.modified=false` 있음(「설치 절차」 정정 참조). a140 의 primary 빌드는 아직 안 쟀다.
- ~~`"$@"` 인자 경계~~ → Manager 실측(2026-11-26)으로 닫음. stub 이 인자를 하나씩 `[…]` 로 찍게 하고 `KR <log> --session "09:00-15:30" --wait-open --config-dir "/a b/c" --pool 45` 를 넘겼다. 결과는 `[run][--market][KR][--session][09:00-15:30][--wait-open][--config-dir][/a b/c][--pool][45]`, keylen=5, cwd=가짜 트리 루트, rc 0. 없는 경로는 rc 1 · 오류 한 줄 · 로그 파일 미생성.
- `shellcheck`(PATH 에 없음).
- 실제 systemd 에서 drop-in `%h` 확장 결과(unit 을 바꾸지 않았으므로).
