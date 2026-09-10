# Branch Test Map: `resolve_referenced_change`

번호는 같은 디렉터리 `ast.json` 의 열거 순서다. **번호는 위치이므로**
함수를 편집하면 편집 지점 뒤가 다른 분기를 가리킨다 —
[[positional-branch-ids-break-hand-renumbering]]. 그래서 아래는 번호만이 아니라
**소스 한 줄**을 같이 적어 편집 뒤에도 대조할 수 있게 한다.

## 편집 전 (HEAD)

| ID | 소스 | 덮는 테스트 | 덮이나 |
|---|---|---|---|
| B1 | `if direct.is_dir():` | `test_real_reference_resolves_an_open_change`(간접, 다수) | yes |
| B2 | 아카이브 순회 | `test_real_reference_resolves_an_archived_change` | yes |
| B3 | `archive.is_dir() else ()` | — 아카이브 디렉터리 자체가 없는 저장소 | **no** |
| B4 | 날짜 접두사 전부 일치 | `test_real_archived_reference_rejects_a_suffix_collision` | yes |
| B5 | `if not matches:` | 같은 위 | yes |
| B6 | `if len(matches) > 1:` | `test_real_archived_reference_rejects_two_copies` | yes |
| — | **활성 + 아카이브 동시 존재** | **없다 — 이것이 3.2.4 다** | **no** |

## 편집 후 (`ast.worktree.json` 열거 확인)

B1 의 early return 을 없애고 활성·아카이브를 **한 목록으로 모은 뒤** 센다.
열거가 그것을 구조로 확인한다 — 분기 6→10, raise 2→3, 그리고 **반환이 2→1** 이다.

| | 편집 전 (`ast.json`, revision `HEAD`) | 편집 후 (`ast.worktree.json`) |
|---|---|---|
| 반환 | L240 `direct` · L256 `archive/matches[0]` | **L279 `found[0]` 하나뿐** |
| raise | L250 없음 · L253 아카이브 중복 | L267 없음 · **L269 `AmbiguousChange`** · L275 아카이브 중복 |
| 활성 판정 | `B1 if direct.is_dir(): return direct` | `B4 [direct] if open_here else []` — **돌려주지 않고 목록에 넣는다** |
| 중복 판정 범위 | `B6 len(matches) > 1` (아카이브 산출만) | `B7 open_here and archived` + `B9 len(archived) > 1` |

**반환이 하나로 줄어든 것이 이 편집의 구조적 증거다.** early return 이 남아 있으면
반환이 둘이고, 그 둘 중 하나는 세는 코드를 건너뛴다. 하나면 건너뛸 길이 없다.

호출자 쪽도 열거로 확인했다 — `check` 의 편집 후 열거에서
`B2 L714 except AmbiguousChange` 가 `B3 L718 except ValueError` **앞**에 선다.
순서가 뒤집히면 상속 관계 때문에 넓은 쪽이 먼저 잡아 새 실패가 다시 삼켜진다.

| 판정 | 덮는 테스트 (신규) |
|---|---|
| 활성만 1개 → 활성을 돌려준다 | 기존 통과 테스트 전부(회귀) |
| 아카이브만 1개 → 그것을 돌려준다 | `test_real_reference_resolves_an_archived_change`(회귀) |
| 활성 1 + 아카이브 1 → **멈춘다** | `test_real_reference_refuses_when_open_and_archived_collide` |
| 아카이브만 2개 → 멈춘다(문구 불변) | `test_real_archived_reference_rejects_two_copies`(회귀) |
| 어디에도 없음 → 멈춘다(문구 불변) | `test_real_archived_reference_rejects_a_suffix_collision`(회귀) |
| 게이트 대상 자신이 활성+아카이브 → **멈춘다** | `test_real_gate_target_refuses_when_open_and_archived_collide` |

마지막 줄이 이 태스크의 핵심이다. 호출자 `check:689-692` 가 `except ValueError` 로
되돌아가므로, 해소기만 고치면 **게이트 대상 경로에서는 예외가 삼켜져** 여전히
활성이 조용히 이긴다. 그래서 시험이 둘이다 — 빌린 증거 경로(`:715`)와
게이트 대상 경로(`:690`).

## 거부하게 될 정상 입력 (fail-closed 는 무엇을 죽이는지 말해야 한다)

[[fail-closed-must-name-what-it-rejects]].

- **아카이브된 change 와 같은 id 로 활성 디렉터리를 다시 만드는 것.** 후속 작업을
  같은 id 로 여는 관행이 있다면 이 규칙이 그것을 죽인다. 오늘 저장소에는 **0건**이다
  (2026-09-10 측정: 활성 27 · 아카이브 id 99 · 교집합 0 · 아카이브 내 중복 0).
  대신 새 id 를 쓰라는 것이 [[tossos-merge-main-fallout]] 이 적은 renumber 관행이다.
- 그 외에는 없다. 활성만·아카이브만인 경우는 전부 지금과 같이 통과한다.
