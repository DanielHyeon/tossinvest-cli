# Function Logic Map: `_read_regular` (Python, a122 task 7.5.2.2)

## task 7.5.2.2 — 스냅숏은 끝에서 디스크와 다시 대조한다 (7.5.2.1 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:649-675` · 분기 3 · 반환 2 · raise 1 (새 함수).

새 함수. 정규 파일이면 바이트, 아니면(FIFO · 장치) `None`, 없거나 못 열면 · **디렉터리면** `OSError`. `O_NONBLOCK` 으로 열고 **연 것의** 종류를 `fstat` 으로 보므로 확인과 읽기 사이 틈이 없다(쓰는 쪽 없는 FIFO 도 바로 열리고 곧 닫힌다). 심링크는 따라간다. **디렉터리는 건너뛰지 않는다** — 건너뛰면 번들 안에 폴더를 만들어 그 안에 열거형 호출 표를 넣는 것으로 감사가 꺼진다(`_bundle_text` 가 막으려는 바로 그것). 종류 검사를 파일 객체로 감싸기 **앞에** 둔 까닭도 그것이다: 먼저 감싸면 `open(fd)` 가 디렉터리에서 스스로 터지고 그 예외의 `filename` 이 정수 fd 라서 이름을 대려던 `_verdict` 가 `TypeError`(GATE_FAULTS 밖)로 죽었다 — **이 로트가 만든 회귀이고 소켓을 재다 찾았다**.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 666 | Try | `try:` |
| B2 | 668 | If | `if stat.S_ISDIR(mode):` |
| B3 | 670 | If | `if not stat.S_ISREG(mode):` |

| raise 줄 | 소스 |
|---|---|
| 669 | `raise IsADirectoryError(errno.EISDIR, os.strerror(errno.EISDIR), str(path))` |

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:792-802` · 분기 1 · 반환 1 · raise 1 (편집 전 L649-675 · 분기 3 · 반환 2 · raise 1, `ast.before-7.5.2.3.json` = revision `1d12520c`).

몸이 `_opened_bytes` 로 가고 여기는 **원장에 적는 깔때기**가 됐다. 반환형이 `bytes | None` → `bytes` 다: 종류가 틀린 것은 이제 예외이고, 호출자는 "못 읽었다" 를 예외 한 가지 모양으로만 만난다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 800 | If | `if isinstance(value, OSError):` |
