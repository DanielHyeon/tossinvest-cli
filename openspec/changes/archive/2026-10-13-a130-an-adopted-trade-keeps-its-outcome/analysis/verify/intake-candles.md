# a130 4판 — 편입 묶음 1분봉 실측 (수 · 사유 · 플래그만, F-6)

- 모집단: 운영 저널 읽기 전용 사본(2026-10-12T21:34:39Z, 스키마 32)의 backfill 모집단 12(체결로 닫힌 편입 포지션 · 행 없음).
- 브로커 읽기: 2026-10-13, `tossctl ops call candles --backend openapi`(읽기 전용 GET, 1.2초 간격 — 첫 측정은 기본 토큰 캐시를 썼음: 리뷰 4차 R4-1 사고 기록;
  커밋된 하네스 `analysis/harness/a130_intake_candles.py` 는 `--config-dir` 로 엔진 캐시를 씀, 재실행 결과 같은 수 · 토큰 파일 재기록 없음), interval 1m, count 1,
  `before` = 봉 이름(= floor_minute(조정 created_at) + 1분, 봉 종료 시각), `adjusted=false`. 그 뒤 도구 미리보기가 같은 읽기를 다시 함.

| 항목 | 수 |
|---|---|
| 편입 뒤 증가 묶음 | 23 (포지션 11 — 12 중 1 은 묶음 없음) |
| 조정 시각이 정규장 안 | 17 |
| 정규장 밖 | 6 (KR 장 전 1 · KR 장 후 3 · US 장 후 2) |
| 그 분의 봉 있음 | **21** (정규장 17 + 정규장 밖 4 — NXT 장 전 · 장 후, US 장 후 체결이 있던 분) |
| 그 분의 봉 없음 | **2** (KRX 마감 뒤 10분 · NXT 마감 뒤 2분 — API 는 더 이른 봉을 돌려줌) |

| 미리보기 판정(모집단 12) | 수 |
|---|---|
| fill | 9 |
| refused `no_intake_candle` | 2 |
| refused `sell_unpriced` | 1 |
| refused `foreign_instance_sell` · `unattributed_local_sell` | 0 |

- 미리보기 전후 사본 파일 SHA-256 동일(쓰기 0). GREEN 커밋 dbb0681c 의 도구(`--config-dir ~/.config/tossctl`)로 재실행: 같은 판정 9 · 2 · 1, 토큰 캐시 파일 재기록 없음.
- 쓰기 잠금 시간: 사본을 따로 하나 더 복사해 스키마 35 로 이관한 뒤 실행(토큰 일치) — 9 항목, 항목당 9.8 ~ 22.2 ms(최대 22.17 ms). 디스크는 /tmp(운영과 다름).
- 봉 이름 규약 확인: 사건 조정 00:18:54Z 는 00:19:00Z(09:19 KST) 봉에 듦 — 공식 `timestamp` 는 봉 종료 시각(a112 결정 30 · 31).
