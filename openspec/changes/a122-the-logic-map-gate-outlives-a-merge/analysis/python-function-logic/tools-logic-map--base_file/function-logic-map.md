# Function Logic Map: `base_file` (Python — a122 task 7.5.34 에서 지웠다)

`ast.before-7.5.34.json` — 분기 1(`if process.returncode:`) · 반환 2(`None` · 임시 경로). 지우기 직전의 열거다.

`git show <rev>:<경로>` 로 옛 쪽(그리고 7.5.31 까지 커밋 대상의 현재 쪽) 바이트를 읽어 임시 `.go` 로 썼다. git 이 oid 로
찾아 주는 바이트를 그대로 믿었으므로 base blob 을 위조하면 옛 쪽이 편집 내용이 되어 요구가 빈다(7.5.25 적대 재리뷰 F5 ·
7.5.33). 실패는 `None` 이었고 호출자가 `cannot load existing base file` 로 멈췄다 — 그 문장과 멈춤은 몸통 B5 가 이어받는다
(비교에 옛 쪽 바이트가 없으면). 두 쪽 바이트를 `Comparison.old` · `.new` 가 들고 오므로 이 함수는 호출자가 없다.
