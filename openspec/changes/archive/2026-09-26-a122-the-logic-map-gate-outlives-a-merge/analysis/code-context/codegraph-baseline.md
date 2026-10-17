# CodeGraph baseline — gate.sh 아카이브 인식 (task 1.11)

Date: 2026-09-10
HEAD: `96a073697ab7806269576504fd5f753c6d95a42f`
Base commit (a122): `7f8b1ae7cb2e9a539dcb9ff5696645a47cac145a`

## 질의와 결과

```
codegraph query   "resolve_referenced_change"  → tools/logic-map/check_analysis.py:231 (function)
codegraph callers "resolve_referenced_change"  → check (check_analysis.py:684) — 1건
codegraph query   "PAIR_DIR"                   → No results found
codegraph query   "gate.sh"                    → 무관한 Go 심볼만 (Block, engineRecoveryObserver …)
```

## Hard evidence

- `resolve_referenced_change` 의 호출자는 `check` 하나다. 파일 안에서는 두 자리
  (`check_analysis.py:690`, `:715`) 이고 둘 다 `check` 본문이다 — CodeGraph 의 1건과
  `rg` 의 2건은 모순이 아니라 함수 단위 대 호출 자리 단위의 차이다.
- 이번 편집 대상인 `tools/gate.sh` 는 **CodeGraph 인덱스에 없다.** shell 파일은
  색인되지 않는다(`PAIR_DIR` 무결과로 확인). 따라서 이 파일의 호출 관계는 CodeGraph 가
  아니라 직접 열거로 확정해야 한다.

## Blast radius (직접 열거)

`tools/gate.sh` 의 호출자는 하나다.

| 자리 | 내용 |
|---|---|
| `Makefile:169` | `bash tools/gate.sh $(CHANGE)` — `make gate` 의 유일한 몸통 |

`.github/` 에는 `gate.sh` 참조가 없다. CI 는 `sdd-check-ci` 만 부른다.
즉 이 편집의 전파 범위는 **사람이 손으로 부르는 완료 게이트 하나**이며, 엔진·주문·
원장·콘솔 런타임 어디에도 닿지 않는다. Go 파일은 한 줄도 바뀌지 않는다.
