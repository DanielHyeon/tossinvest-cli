# 증거 조정 — a169 2.1

- 기준: base `2bbee345f`, 판정 HEAD `566e1cf7c`(+ 작업 트리 편집).
- 변경 표면: `domain.Quote.SecurityType`(새 칸, `security_type,omitempty`) · `adaptStocks` 리터럴 대입 한 줄.
- CodeGraph callers 와 `rg 'adaptStocks|\.Stocks\('` 결과 일치(소비자 3 · 시험 3). 불일치 0.
- 직렬화 · digest 영향: `domain.Quote` 를 통째로 JSON 으로 내는 곳은 `output.WriteQuote` · `ops call` 반환뿐이고 둘 다 omitempty 라
  값 없는 생산자(WTS `GetQuote` · `/prices`)는 바이트 동일. a128 도구는 Quote 를 통째로 직렬화하지 않아 상태 digest 불변.
  `reflect` 로 Quote 칸을 세는 시험 0(`rg`).
- 결론: 편집 허용 — 읽기 전용 경로, High-risk 밖.
