#!/usr/bin/env python3
"""a200 1.9 — 정본 delta 생성.

정본 Requirement 블록을 그대로 복사한 뒤 exact-string 치환만 한다. 치환마다 원문에 정확히 1회 있어야 하고,
결과에 역치환을 걸면 원문과 바이트가 같아야 한다(의도 외 차이 0). Scenario 이름 집합은 그대로여야 한다.
ADDED 본문(제목 · Scenario 제외)은 500자 이하.

    python3 <이 파일>       # 저장소 어디서든
"""
import pathlib
import re
import subprocess
import sys

HERE = pathlib.Path(__file__).resolve().parent
CHANGE = HERE.parent.parent
ROOT = pathlib.Path(subprocess.check_output(["git", "-C", str(HERE), "rev-parse", "--show-toplevel"], text=True).strip())

ADDED_NAME_LEDGER = "수렴 조정이 닫은 인스턴스에 늦게 닿은 엔진 청산 체결은 인정된다"

LEDGER_ADDED = f"""### Requirement: {ADDED_NAME_LEDGER}
수렴 조정이 수량 0 으로 닫은 인스턴스에 매도 체결이 와 전이표가 닫힌 인스턴스의 매도로 거절하면, 다음이 모두 참일 때 그 체결을 그 조정이 먼저 반영한 감소의 설명으로 인정해야 한다(SHALL): 그 인스턴스를 닫은 것이 수렴이 쓰는 종류의 수량 0 조정 하나이고, 체결 주문의 소유 intent 가 그 인스턴스의 무장된 청산 발의 intent 이며, 그 intent 의 주문이 정정 없는 접수 확정 주문 하나이고 그 주문 수량이 그 조정의 감소분 이하다. 인정된 체결은 투영 수량을 다시 빼지 않고(SHALL NOT) 대사 차단을 만들지 않으며(SHALL NOT) exit 이력에 인정 표지 한 줄을 남긴다(SHALL). 하나라도 어긋나면 지금처럼 거절하고 RECONCILE 로 전이한다(SHALL).

#### Scenario: 수렴 조정 뒤 같은 초에 온 손절 체결
- **WHEN** 수렴 조정이 관리 포지션을 수량 0 으로 닫은 뒤 그 인스턴스의 손절 발의 주문이 같은 초에 전량 체결되면
- **THEN** 투영은 CLOSED · 0 그대로이고 차단 행은 생기지 않으며 인정 표지 한 줄이 남고, exit 상태 · 성과 기록은 조정으로 닫힌 포지션과 같다

#### Scenario: 조정 뒤 부분 체결 여러 번
- **WHEN** 주문 수량이 조정 감소분과 같은 그 손절 주문이 조정 뒤에 여러 번 나뉘어 체결되면
- **THEN** 각 체결이 인정되어 체결마다 표지가 남고 차단 행은 생기지 않는다

#### Scenario: 조건이 어긋난 체결
- **WHEN** 체결로 닫힌 인스턴스에 엔진 매도 체결이 오거나, 체결 주문이 그 인스턴스의 청산 발의가 아니거나(발의가 이미 해소된 뒤의 체결 포함), 주문 수량이 조정 감소분을 넘거나, 그 발의에 정정 · 재제출된 주문이 있거나, 닫은 조정의 종류가 수렴이 쓰는 값이 아니면
- **THEN** 지금처럼 거절되어 귀속 실패 RECONCILE 이 되고 표지는 남지 않는다

#### Scenario: 조정을 사이에 둔 부분 체결
- **WHEN** 손절 주문의 체결 일부가 조정 전에 투영에 반영되고 나머지가 조정 뒤에 오면
- **THEN** 조정이 닫은 수량이 주문 수량보다 작으므로 조정 뒤 첫 체결부터 인정하지 않고 귀속 실패 RECONCILE 이 된다
"""

RECON_ADDED = """### Requirement: 인정된 엔진 청산 체결은 귀속 실패 차단을 만들지 않는다
position-ledger 의 「수렴 조정이 닫은 인스턴스에 늦게 닿은 엔진 청산 체결은 인정된다」 가 인정한 체결은 귀속 실패 차단을 만들어서는 안 된다(SHALL NOT). 그 대사 주기에 수렴이 세운 수량 불일치 차단은 지금의 해제 규칙 그대로 풀린다(SHALL — 조정 반영 근거가 있는 다음 일치 비교에서, 그 근거를 잃는 재기동이 끼면 지금처럼 조정 대기). 인정 조건이 하나라도 어긋난 체결은 지금과 같이 다룬다(SHALL) — 그 종목에 활성 차단이 없으면 귀속 실패 차단을 만들고 그 차단은 스스로 풀리지 않으며, 이미 활성 차단이 있으면 새 행을 쓰지 않는다. 귀속 실패 차단의 해제 규칙(운영자 해제)은 바꾸지 않는다(SHALL NOT).

#### Scenario: 수렴과 같은 주기에 손절 체결이 닿는다
- **WHEN** 대사의 수렴 조정이 커밋된 뒤 그 주기의 비교 기록 전에 인정되는 손절 체결이 커밋되면
- **THEN** 그 종목에 귀속 실패 행이 생기지 않고, 수량 불일치 차단은 다음 일치 비교에서 조정 반영으로 풀려 그 종목 진입이 다시 열린다

#### Scenario: 인정 조건이 어긋난 체결
- **WHEN** 그 종목에 활성 차단이 없을 때 수렴 조정이 닫은 인스턴스에 온 매도 체결이 인정 조건 중 하나를 어기면
- **THEN** 그 종목에 귀속 실패 차단이 생기고 차이 없는 비교가 반복돼도 풀리지 않는다
"""

# (정본 파일, Requirement 이름, [(old, new)], 덧붙일 Scenario 텍스트)
MODIFIED = [
    ("openspec/specs/position-ledger/spec.md", "포지션 상태기계", [
        ("허용되지 않은 전이는 오류이며 RECONCILE로 전이한다(SHALL — 산식 보정 금지).",
         "허용되지 않은 전이는 오류이며 RECONCILE로 전이한다(SHALL — 산식 보정 금지). 예외는 "
         f"「{ADDED_NAME_LEDGER}」 가 인정한 체결 하나뿐이며 그때도 투영 수량은 바뀌지 않는다(SHALL)."),
    ], ""),
]


def requirement_block(text: str, name: str) -> str:
    start = text.index(f"### Requirement: {name}\n")
    nxt = text.find("\n### Requirement: ", start + 1)
    end = len(text) if nxt < 0 else nxt + 1
    return text[start:end]


def scenarios(block: str) -> list[str]:
    return re.findall(r"^#### Scenario: (.+)$", block, re.M)


def body_len(block: str) -> int:
    lines = block.splitlines()[1:]
    body = []
    for line in lines:
        if line.startswith("#### Scenario"):
            break
        body.append(line)
    return len("\n".join(body).strip())


def main() -> int:
    out: dict[str, list[str]] = {}
    for rel, name, subs, appended in MODIFIED:
        original = requirement_block((ROOT / rel).read_text(encoding="utf-8"), name)
        block = original
        for old, new in subs:
            if block.count(old) != 1:
                print(f"치환 원문이 {rel} 「{name}」 에 정확히 1회가 아님: {block.count(old)}")
                return 1
            block = block.replace(old, new)
        block = block.rstrip("\n") + "\n" + appended if appended else block
        reverse = block[: len(block) - len(appended)] if appended else block
        if appended:
            reverse = reverse.rstrip("\n") + "\n"
        for old, new in subs:
            if reverse.count(new) != 1:
                print("역치환 대상이 1회가 아님")
                return 1
            reverse = reverse.replace(new, old)
        if reverse.rstrip("\n") != original.rstrip("\n"):
            print(f"역치환이 원문과 다름: {rel} 「{name}」")
            return 1
        if not set(scenarios(original)) <= set(scenarios(block)):
            print("Scenario 이름이 빠짐")
            return 1
        out.setdefault(rel.split("/")[2], []).append(block.rstrip("\n") + "\n")
        print(f"MODIFIED {rel} 「{name}」: 치환 {len(subs)} · 역치환 원문 일치 · Scenario {len(scenarios(block))}")

    for label, text in (("position-ledger ADDED", LEDGER_ADDED), ("reconciliation ADDED", RECON_ADDED)):
        n = body_len(text)
        print(f"{label} 본문 {n}자")
        if n > 500:
            print("ADDED 본문 500자 초과")
            return 1

    ledger = CHANGE / "specs/position-ledger/spec.md"
    ledger.parent.mkdir(parents=True, exist_ok=True)
    ledger.write_text("## ADDED Requirements\n\n" + LEDGER_ADDED + "\n## MODIFIED Requirements\n\n"
                      + "\n".join(out["position-ledger"]), encoding="utf-8")
    recon = CHANGE / "specs/reconciliation/spec.md"
    recon.parent.mkdir(parents=True, exist_ok=True)
    recon.write_text("## ADDED Requirements\n\n" + RECON_ADDED, encoding="utf-8")
    print("wrote", ledger.relative_to(ROOT), recon.relative_to(ROOT))
    return 0


if __name__ == "__main__":
    sys.exit(main())
