# Function Logic Map: `write`

- Source: `internal/enginelock/enginelock.go` (405-447)
- AST evidence: `ast.json` — AST 기준 branches **8** / returns 9 / calls 26
- Risk scan: `risk-pattern-report.md` (매치 없음)
- source SHA-256: `e615b0c66ce3ecb71540cb96f5bcf81ed0240530af5c1aac430a577e0c582b38`
- 작성 사유: a102 §3.9(D4c, A2 F3) — 파일 교체를 `os.WriteFile`에서 tmp+rename으로 바꾼다.
  **기존 함수의 내부를 편집하므로 편집 전에 만든다.** 마커는 콘솔·서베이·자동시작이 읽는
  신호이고 a102 이후로는 "보호가 서 있다"까지 담으므로 High-risk다.

## 두 판

| | 1판 (편집 전, `6cd643ca`) | **2판 (이 문서)** |
|---|---|---|
| 위치 | `:305-324` | **`:405-447`** |
| 분기 | 3 | **8** |
| 이탈 | 4 | **9** |
| 호출 | 8 | **26** |
| 교체 방식 | `os.WriteFile`(O_TRUNC) → `os.Chtimes` | **tmp 생성 → 쓰기 → chmod → chtimes → rename** |
| source SHA-256 | `b4ee6b24394a…` | **`e615b0c66ce3…`** |

분기가 다섯 늘어난 것은 전부 **원자 교체의 실패 경로**다. 각 단계가 실패하면 임시 파일을
지우고 오류를 돌려준다 — 저널 디렉터리에 쓰레기를 남기지 않기 위해서다.

## 이 함수가 하는 일

마커 파일 하나를 **원자적으로 교체**한다. 내용은 JSON이고, 사실은 mtime이다.

**편집 이유는 측정이다.** A2가 편집 전 형태를 동시 부하로 재현했다: `os.WriteFile`은 자르고
나서 채우므로 그 창에 들어온 독자가 **12259회 중 3617회 찢어진 파일**을 읽었다. 관대한
reader는 파싱 실패를 "running, build unknown"으로 강등하는데(패키지 주석), a102 이후로는
그 강등이 **ready_at도 함께 사라진 것**을 뜻한다. 그리고 그 독자는 대시보드와 자동시작
스크립트다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `path` | 저널 디렉터리의 `engine-run.json` | `MarkerPath(dir)` | 각 단계가 오류를 돌려준다 |
| `m Marker` | 값 복사본 | 호출자 (`Hold`·`Ready`·`refresh`) | — |
| `at` | 임의 | 주입된 시계 | mtime이 되고 그것이 freshness의 사실 |
| 임시 파일 | `dir` 안의 `.engine-run-*.json` | `os.CreateTemp` | 실패 시 **반드시 지운다** |

> **불변식 셋.**
> (1) 독자는 **이전 마커 아니면 다음 마커**를 본다. 그 중간은 없다 — rename이 원자다.
> (2) 임시 파일은 마커와 **같은 디렉터리**에 만든다. rename이 파일시스템을 건너지 않게.
> (3) mtime은 **rename 전에** 찍는다. 잘못된 시각의 마커가 한 순간도 보이지 않게.
> (4) 모드는 0600이다. 마커는 pid와 빌드 지문을 담는다.

## Branches and early returns

| Branch | 위치 | Condition | Mutation/side effect | Return/이탈 |
|---|---|---|---|---|
| B1 | `:407` | `os.MkdirAll(dir)` 오류 | 없음 | `:408` |
| B2 | `:412` | `json.MarshalIndent` 오류 | 없음 | `:413` |
| B3 | `:417` | `os.CreateTemp` 오류 | 없음 | `:418` |
| B4 | `:421` | 임시 파일 쓰기 오류 | `Close` + **`Remove`** | `:424` |
| B5 | `:426` | `Close` 오류 | **`Remove`** | `:428` |
| B6 | `:430` | `Chmod` 오류 | **`Remove`** | `:432` |
| B7 | `:438` | `Chtimes` 오류 | **`Remove`** | `:440` |
| B8 | `:442` | `Rename` 오류 | **`Remove`** | `:444` |

정상 이탈: `:446` `return nil`.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `filepath.Dir(path)` `:406` | 임시 파일을 같은 디렉터리에 | — | ast.json |
| `os.MkdirAll` `:407` | 첫 실행의 저널 디렉터리 | B1 | ast.json |
| `json.MarshalIndent` `:412` | 사람이 `cat`으로 읽는 형식 | B2 | ast.json |
| `os.CreateTemp` `:417` | 원자 교체의 준비 (0600으로 생성) | B3 | ast.json |
| `tmp.Write` `:421` | 본문 | B4 — 지우고 반환 | ast.json |
| `tmp.Close` `:426` | flush | B5 | ast.json |
| `os.Chmod` `:430` | 0600 **명시** (umask와 무관하게) | B6 | ast.json |
| `os.Chtimes` `:438` | **rename 전에** mtime 고정 | B7 | ast.json |
| `os.Rename` `:442` | **원자 교체** | B8 | ast.json |
| `os.Remove` (실패 경로 5회) | 쓰레기 제거 | 반환값은 무시한다 — 이미 오류를 들고 나간다 | ast.json |

live binding — 호출자는 셋이고 전부 이 패키지 안이다: `Hold`(`:366`, 첫 쓰기) ·
`(*Held).Ready`(`:325`) · `(*Held).refresh`(`:349`) · `(*Held).Identify`(`:296`). **뒤의 둘은 `h.mu`를 잡은 채 부른다**
— A2 F3의 나머지 절반(뮤텍스가 메모리만 덮고 파일은 안 덮던 것)이 그 이동으로 닫힌다.

## State mutations and fallbacks

- 프로세스 밖 상태: 임시 파일 하나(만들고 지우거나 rename으로 사라짐)와 마커 파일 하나.
- **fallback이 없다.** 여덟 분기 전부 즉시 오류 반환이다. 그 오류를 호출자가 어떻게 다루는지가
  마커의 규율이다: `Hold`는 돌려주고(거절은 아니다), `Ready`·`refresh`는 삼킨다.
- 부분 상태를 남기지 않는다 — 실패한 경로는 임시 파일을 지우므로 **마커는 편집 전 상태 그대로**다.

## Safety conclusion

- Safe edit boundary: **파일 교체 방식 하나.** 직렬화 형식(JSON 필드·들여쓰기·끝 개행)과
  mtime 규율, 0600 모드는 전부 편집 전과 같다. `Read`는 무편집이다.
- High-risk impact: **yes** — 이 파일이 "엔진이 살아 있다"와 "보호가 서 있다"를 동시에 말한다.
  방향은 보수적이다: 실패하면 **마커가 갱신되지 않을 뿐**이고, 그것은 StaleAfter로 늙어
  "엔진 없음"이 된다 — 서베이는 그때 오늘의 동작(즉시 시작)으로 돌아간다.
- 남은 공백: 여덟 분기 중 **여섯의 본문이 count=0**이다(§3.9c가 B8을 잡았다)(`branch-test-map.md`). 실패 경로를
  만들려면 파일시스템을 고장 내야 하고, 이 change는 그 주입 지점을 만들지 않는다.
  **침묵하지 않고 이름을 붙여 남긴다** — 대신 성공 경로의 *성질*(원자성·모드·쓰레기 없음)은
  전부 측정으로 고정했다.
