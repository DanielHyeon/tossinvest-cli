# a103 — 롤백 핀은 기억하는 게 아니라 만들어지는 것이다

## Why

2026-08-13 11:03 배포에서 `docker tag`가 또 빠졌다. 새 이미지 `66553dba92d8`은
`tossos:local` 하나만 달고 있었고, 다음 빌드가 그 태그를 가져가면 schema 31을
배포한 이미지가 이름 없이 사라질 참이었다.

### 원인은 잊어버림이 아니라 구조다

저장소 전체에서 `docker tag`가 나오는 곳은 **정확히 한 군데**다.

```
docs/operations.md:276:  docker tag <그 id> tossos:<change>-<commit>   # ← 이것이 빠지면 다음 빌드에 사라진다
```

산문 안의 코드블록이다. Makefile에도, `tools/*.sh`에도, `scripts/*.sh`에도 이 일을
하는 것이 **없다**. `make build`는 Go 바이너리만 만들고 이미지와 무관하다.

즉 사람이 실제로 실행하는 두 명령 사이에 **실행되지 않는 한 줄**이 끼어 있다.

```
docker compose build          ← 실행된다
docker tag <id> tossos:<...>  ← 문서에만 있다. 사람이 기억해야 한다
docker compose up -d          ← 실행된다
```

### 문서는 이미 한 번 실패한 처방이다

그 문서 절이 언제 쓰였는지가 결정적이다.

```
c1411dc7  2026-08-12 09:03:00  docs(operations): 롤백 대상은 태그가 없으면 사라진다
```

**이미지 두 개(`86c6e4d2`, `e3cc6244`)를 같은 방식으로 잃은 직후**, 바로 그 재발을
막으려고 쓴 문서다. 그리고 **26시간 뒤에 똑같이 빠졌다.** 같은 처방을 한 번 더
쓰는 것은 이미 결과가 나온 실험을 반복하는 것이다.

### 왜 이게 위험한가

`docker compose build`는 `tossos:local`을 새 이미지로 옮기고 **직전 이미지는 태그를
잃는다**. 태그 없는 이미지는 다음 prune에 사라진다. 그리고 이번 배포는 저널을
**30 → 31로 편도 마이그레이션**했으므로, 롤백은 "이전 이미지를 띄운다"가 아니라
"pre-migration 백업을 복원하고 이전 이미지를 띄운다"다. 이전 이미지가 없으면
그 절차의 절반이 없다.

## What Changes

**태그를 파괴가 일어나는 명령 안으로 옮긴다.** 문서 강화가 아니라 구조 변경이다.

### 1. build와 tag는 한 명령이다

`make image CHANGE=<change-id>`가 빌드하고 그 자리에서 `tossos:<change>-<commit>`을
박는다. 사이에 사람이 기억할 단계가 없다.

### 2. 파괴하는 쪽이 파괴 대상을 먼저 이름 부른다

빌드 **전에** 지금 `tossos:local`이 가리키는 이미지가 다른 태그를 갖고 있는지 본다.
없으면 이 빌드가 그것을 이름 없는 상태로 만든다는 뜻이므로, 그 사실을 말하고
멈춘다. 계속하려면 그 이미지에 먼저 이름을 준다.

이것이 손으로 `docker compose build`를 치는 경로까지 닫지는 못한다. 그래서
문서가 그 두 줄 대신 **한 명령만** 보여준다.

### 3. 문서는 실행되는 것을 가리킨다

`docs/operations.md`의 교체 절차에서 `docker compose build` + 기억해야 하는
`docker tag` 조합을 지우고 `make image CHANGE=…`로 바꾼다. 잃는 방식에 대한 설명
(`259-281`)은 남긴다 — 그것은 **왜**이고, 여전히 참이다.

## 범위 밖

- release digest-pinned 교체 절차(`internal/deployguard`)는 건드리지 않는다.
  그것은 계획 값만 만들고 Docker를 실행하지 않는다는 계약이 있다
  (`docs/operations.md:255`). 이 change는 **개발 build 경로**의 태그 유실만 다룬다.
- 이미 이름 없이 사라진 이미지(`86c6e4d2`, `e3cc6244`)는 복구 대상이 아니다.
- 롤백 핀의 schema 기록(`operations.md:280`)은 이미 문서에 있고 그대로 둔다.

## Impact

- `Makefile` — 새 타겟. production trading code 아님.
- `docs/operations.md` — 교체 절차.
- 주문·손절·사이징 경로 **무관**. High-risk 아님.
