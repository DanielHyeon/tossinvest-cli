# Function Logic Map: `SetA087LowerLimitForTest`

시험 전용 helper. options nil이면 반환; 아니면 non-nil 시험 admission을 가진 판에 지정 종목·가격·시각을 발행해 options에 설정한다. 기존 nil admission은 휴면 상태이므로 활성 fixture의 목적을 보존하는 값만 변경했다. 생산 Go에 포함되지 않음. 후행 AST `ast.json`, 선행 AST `analysis/20261030-daily-refresh/export-pre.ast.json`.
