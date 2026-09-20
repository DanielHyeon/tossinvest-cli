# Branch Test Map: `_read_regular` (Python, a122 task 7.5.2.2)

## task 7.5.2.2 — 스냅숏은 끝에서 디스크와 다시 대조한다 (7.5.2.1 재리뷰, 2026-09-20)

짝은 손으로 고르지 않았다 — 각 갈래를 지우는 변이를 걸어 실제로 빨개진 시험을 옮겼다(전수 95 · 생존 0: 94 는 한 판에서 CAUGHT(스위트 272 · 무변이 대조군 GREEN) · Z14 는 **도달한 채 살아남아**(`io.BytesIO(None)` 이 조용히 빈 버퍼다) 구조 시험을 더한 뒤 같은 하네스로 다시 재서 CAUGHT(273)). '—' 는 변이 대신 직접 시험이다.

| 갈래 | 변이 (`75_mut.py`) | 잡는 시험 |
|---|---|---|
| B1 `try/finally` — 어느 갈래로 나가도 연 것을 닫는다 | Z19 | `test_reading_what_is_not_a_regular_file_closes_the_descriptor` |
| B2 `S_ISDIR` → 경로를 담은 `IsADirectoryError`(건너뛰지 않는다) | Z16 · Z18 | `test_a_folder_inside_a_bundle_is_named_not_a_traceback` |
| B3 `not S_ISREG` → `None`(FIFO · 장치) | Z1 | `test_a_device_among_the_bundle_files_is_skipped_not_drained` |
| 갈래 밖 — `os.open` 자체가 거절하는 항목(소켓 ENXIO) | — | `test_a_socket_inside_a_bundle_is_named_not_a_traceback` |
