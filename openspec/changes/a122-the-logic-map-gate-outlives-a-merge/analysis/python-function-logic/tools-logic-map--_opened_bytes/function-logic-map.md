# Function Logic Map: `_opened_bytes` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:749-776` · 분기 4 · 반환 1 · raise 3 (새 함수).

새 함수. 파일을 **여는 유일한 자리** — `_read_regular` 의 몸에서 떼어 냈다. 종류 검사는 그대로(`O_NONBLOCK` + `fstat`, 디렉터리는 경로를 담은 `IsADirectoryError`)이고 둘이 바뀌었다: 정규 파일이 아니면 이제 `None` 이 아니라 **이름 댄 예외**(`NotRegularFile`)다 — 조용한 건너뛰기가 타이밍 공격의 문이었다(실측 6/14) — 그리고 상한보다 **한 바이트 더** 읽어 넘치면 `FileTooLarge` 다(옛 판본은 큰 정규 파일에서 `MemoryError` 로 판정 줄이 0 이었다). 어느 갈래로 나가도 서술자를 닫는다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 764 | Try | `try:` |
| B2 | 766 | If | `if stat.S_ISDIR(mode):` |
| B3 | 768 | If | `if not stat.S_ISREG(mode):` |
| B4 | 774 | If | `if len(raw) > READ_CAP:` |
