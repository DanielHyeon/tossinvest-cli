# CodeGraphContext context — a133

- 2026-10-13 `codegraphcontext stats .` → Files 5023 · Functions 15268(색인 있음).
- `codegraphcontext analyze callers Publish`(`raw/cgc-callers-Publish.txt`) — 현재 경로에서 `publishHTTPAPISnapshots` · `publishNotificationTest` · **`deliverOne`** · `Beat` ·
  `publishBestEffort` · `deliver` · `Flush`. CodeGraph 가 놓친 `deliverOne` 을 찾음. 대신 `NormalRelay.publish` 를 놓침. 옛 체크아웃 경로(`/mnt/D/project/axipient/TossOS`)의
  행이 섞여 있음(이름을 바꾼 체크아웃의 잔재 — 기억 「이름을 바꾼 체크아웃은 절대경로 상태를 고아로 만든다」).
- 누락 후보: 없음(두 도구의 합집합 + HEAD grep 이 같은 7 자리).
