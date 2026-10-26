# 착수 증거 정합성

| 출처 | 확인 결과 | 채택 |
|---|---|---|
| a175 design/tasks/spec | 읽기 전용 재배치, 값·핸들러 불변, 시장 독립 표시 요구 | 채택 |
| 현재 템플릿·레지스트리 | `table-scroll`, `details.tech`, 정책·후보·runtime read model이 이미 존재 | 채택 |
| ai-memory 검색 | 기존 Go 시험 함수도 FLM 대상이며 미측정 값을 기본값으로 보이지 않게 해야 함 | 현재 원문과 일치하여 채택 |
| codegraph | 이전 기준선에서 `/optimization` 단일 라우트와 handler 연결을 확인함 | 원문 검색으로 재확인 |
| CodeGraphContext | 실행 파일 부재 | 미사용, 누락으로 기록 |
| GBrain | `No brain configured` | 미사용, 새 구성·업그레이드 금지 |
| research-first | 외부 라이브러리 없이 기존 `html/template`·DOM helper·레지스트리로 충족 | 신규 의존성 0 |

결론: 새 로컬 view-model과 템플릿 재배치만 구현함. 외부 패키지·공식 문서·레지스트리 검색은 현재 저장소에 완결된 구현 근거가 있어 추가로 필요하지 않았음.
