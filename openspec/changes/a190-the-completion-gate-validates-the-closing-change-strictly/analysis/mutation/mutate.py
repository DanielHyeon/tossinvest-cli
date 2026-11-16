#!/usr/bin/env python3
"""a190 변이 원장 하네스.

- 대상은 **사본**이다: `$HOME/tossos-work/a190-mut-<pid>/` 에 필요한 파일만 같은 상대 경로로 복사한다.
  원본 저장소에는 쓰지 않는다. 시작할 때 원본 HEAD sha, 대상 파일의 미커밋 여부, **파일별 sha256**(이 판이 잰 바이트)을 기록한다.
- 무변이 대조군이 GREEN 이 아니면 멈춘다(형제 파일이 빠진 사본에서 변이가 전부 CAUGHT 로 찍힌 전례).
- 변이는 문자열 치환이고 원본에서 정확히 1번 나와야 한다(안 닿는 변이 금지). 적용 뒤 내용이 바뀌었는지도 단언한다.
- CAUGHT = 기대한 시험 이름 중 하나 이상이 FAIL/ERROR. 그 밖의 실패만 있으면 OTHER, 아무것도 안 실패하면 SURVIVED.
- 변이 사이에는 사본을 원본 바이트로 되돌리고 되돌린 바이트를 대조한다.

사용법: python3 mutate.py <원본 저장소 루트>
"""

from __future__ import annotations

import hashlib
import os
import re
import shutil
import subprocess
import sys
from pathlib import Path

FILES = (
    "tools/gate.sh",
    "tools/sdd/check_openspec_validate.py",
    "tools/sdd/test_gate_validate_step.py",
    "tools/sdd/test_race_detector_actually_runs.py",
    "Makefile",
    ".github/workflows/ci.yml",
)
# test_race_detector_actually_runs 의 나머지 시험은 Go 파일·패키지 디렉터리를 보므로 사본에서 빨갛다(첫 판 대조군 RED,
# history-run1-control-red-missing-go-files.txt). gate.sh 를 읽는 두 시험만 고른다.
RACE = "tools.sdd.test_race_detector_actually_runs.RaceDetectorActuallyRunsTests."
MODULES = (
    "tools.sdd.test_gate_validate_step",
    RACE + "test_the_gate_step_count_matches_the_targets_it_runs",
    RACE + "test_the_completion_gate_runs_it",
)
SCRIPT = "tools/sdd/check_openspec_validate.py"
GATE = "tools/gate.sh"
A189 = "a189-shadow-tool-guards-admit-the-account-budget"

# (이름, 파일, 찾을 문자열, 바꿀 문자열, 기대한 시험 이름들)
MUTANTS = (
    ("M1 (가)에서 --strict 제거", SCRIPT,
     '["validate", change_id, "--strict", "--no-interactive", "--json"]',
     '["validate", change_id, "--no-interactive", "--json"]',
     {"test_2_a_strict_warning_in_the_closing_change_fails_at_the_closing_check",
      "test_all_ok_passes_and_the_three_commands_have_the_decided_shape",
      "test_real_openspec_fails_at_step_eleven_on_a_closing_warning"}),
    ("M2 (가) 호출 삭제", SCRIPT,
     '("(가)", f"{change_id} strict valid", closing_verdict(change_id, ROOT)),',
     '("(가)", f"{change_id} strict valid", (True, [])),',
     {"test_2_a_strict_warning_in_the_closing_change_fails_at_the_closing_check",
      "test_5_a_change_openspec_cannot_find_fails",
      "test_all_ok_passes_and_the_three_commands_have_the_decided_shape",
      "test_stubbed_closing_failure_stops_the_gate_at_step_eleven"}),
    ("M3 (나) 호출 삭제", SCRIPT,
     '("(나)", "저장소 전체(strict 없음) 오류 0", repository_verdict(change_id, ROOT)),',
     '("(나)", "저장소 전체(strict 없음) 오류 0", (True, [])),',
     {"test_3_an_error_anywhere_fails_at_the_repository_check",
      "test_6_a_repository_check_over_zero_items_fails",
      "test_4_openspec_absent_fails_both_checks",
      "test_v10_a_repository_report_that_never_saw_the_closing_change_fails"}),
    ("M4 (나)에 --strict 추가", SCRIPT,
     'run(["validate", "--all", "--no-interactive", "--json"], root, timeout)',
     'run(["validate", "--all", "--strict", "--no-interactive", "--json"], root, timeout)',
     {"test_1_strict_warnings_elsewhere_do_not_fail_the_step",
      "test_all_ok_passes_and_the_three_commands_have_the_decided_shape",
      "test_real_openspec_passes_with_strict_warnings_elsewhere"}),
    ("M5 표본 0 검사(V9) 삭제", SCRIPT,
     '    if not items:\n        return False, ["[V9] 저장소 전체 검증이 0 항목을 검사함"]\n',
     '',
     {"test_6_a_repository_check_over_zero_items_fails",
      "test_v9_zero_items_fails_even_when_the_closing_check_passes"}),
    ("M5b V9 와 V10 둘 다 삭제", SCRIPT,
     '    if not items:\n        return False, ["[V9] 저장소 전체 검증이 0 항목을 검사함"]\n'
     '    if not any(i.get("id") == change_id and i.get("type") == "change" for i in items):\n'
     '        return False, [f"[V10] 저장소 전체 검증 {len(items)} 항목에 닫는 change {change_id} 가 없음"]\n',
     '',
     {"test_6_a_repository_check_over_zero_items_fails",
      "test_v9_zero_items_fails_even_when_the_closing_check_passes",
      "test_v10_a_repository_report_that_never_saw_the_closing_change_fails"}),
    ("M6 advisory 실패를 판정으로 승격", SCRIPT,
     'return 0 if all(ok for _, _, (ok, _) in verdicts) else 1',
     'return 0 if all(ok for _, _, (ok, _) in verdicts) and " 0 failed" in advisory(ROOT) else 1',
     {"test_1_strict_warnings_elsewhere_do_not_fail_the_step",
      "test_all_ok_passes_and_the_three_commands_have_the_decided_shape",
      "test_7_an_advisory_that_crashes_never_changes_the_verdict",
      "test_stubbed_pass_reaches_gate_pass_with_this_change_id",
      "test_real_openspec_passes_with_strict_warnings_elsewhere"}),
    ("M7a gate.sh: CHANGE_ID 대신 고정 문자열", GATE,
     'python3 tools/sdd/check_openspec_validate.py "$CHANGE_ID"',
     f'python3 tools/sdd/check_openspec_validate.py "{A189}"',
     {"test_stubbed_pass_reaches_gate_pass_with_this_change_id",
      "test_real_openspec_passes_with_strict_warnings_elsewhere"}),
    ("M7b 스크립트: (가)에 change_id 대신 고정 문자열", SCRIPT,
     '["validate", change_id, "--strict", "--no-interactive", "--json"]',
     f'["validate", "{A189}", "--strict", "--no-interactive", "--json"]',
     {"test_all_ok_passes_and_the_three_commands_have_the_decided_shape",
      "test_1_strict_warnings_elsewhere_do_not_fail_the_step"}),
    ("M8 gate.sh: 옛 make validate 로 되돌림(11단계 블록 삭제 + 순회에 validate)", GATE,
     None, None,  # 아래 OLD_GATE 함수가 만든다
     {"test_stubbed_pass_reaches_gate_pass_with_this_change_id",
      "test_stubbed_closing_failure_stops_the_gate_at_step_eleven"}),
    ("M9 gate.sh: 11단계 블록 삭제 + TOTAL_STEPS=10 (리뷰 #4 시나리오)", GATE,
     None, None,
     {"test_stubbed_pass_reaches_gate_pass_with_this_change_id",
      "test_stubbed_closing_failure_stops_the_gate_at_step_eleven"}),
    ("M10 root 대조(V6) 삭제", SCRIPT,
     '    if reported != expected:\n',
     '    if False:\n',
     {"test_v6_a_report_about_another_root_fails", "test_v6_a_report_without_a_root_fails"}),
    ("M11 시한 초과(V2) 처리 삭제", SCRIPT,
     '    except subprocess.TimeoutExpired:\n        return None, f"[V2] openspec 이 {timeout}초 안에 끝나지 않음"\n',
     '',
     {"test_v2_a_command_that_outlives_its_deadline_fails"}),
    ("M12 (가) id·type 대조(V8) 삭제", SCRIPT,
     'if len(items) != 1 or items[0].get("id") != change_id or items[0].get("type") != "change":',
     'if len(items) != 1:',
     {"test_v8_the_closing_report_must_name_exactly_this_change"}),
    # 리뷰 R1 뒤 추가 — testing 전문가: (나)의 V11 갈래가 CI(openspec 없음)에서 무방비였다.
    ("M13 (나) rc·JSON 대조 삭제", SCRIPT,
     "    faults = agreement_faults(rc, items)\n    return (not faults), faults\n\n\ndef advisory",
     "    return True, []\n\n\ndef advisory",
     {"test_3_an_error_anywhere_fails_at_the_repository_check",
      "test_v11_an_invalid_item_anywhere_fails_the_repository_check",
      "test_v12_rc_and_json_must_agree"}),
    ("M14 root 를 해석해 비교(R1 이전·R1 중간의 resolve 방식, Codex R1 #2·적대 R2 F3)", SCRIPT,
     "    if reported != expected:\n",
     "    if not isinstance(reported, str) or str(Path(reported).resolve()) != expected:\n",
     {"test_v6_an_empty_relative_or_unreadable_root_is_not_this_repository"}),
    ("M15 issues 목록 아님 처리 삭제(adversarial R1 #2)", SCRIPT,
     '    if not isinstance(issues, list):\n        return f"issues 를 읽을 수 없음: {issues!r}"\n',
     "",
     {"test_null_issues_still_report_both_checks_and_the_advisory"}),
    ("M16 엄격 디코딩을 errors=replace 로 되돌림(Codex R2 #5)", SCRIPT,
     'json.loads(proc.stdout.decode("utf-8"))',
     'json.loads(proc.stdout.decode("utf-8", errors="replace"))',
     {"test_broken_utf8_inside_a_json_string_is_v3_not_a_pass"}),
    ("M16b 디코딩 예외 처리 삭제(Codex R1 #1)", SCRIPT,
     "    except (ValueError, RecursionError) as exc:\n",
     "    except (RecursionError,) as exc:\n",
     {"test_undecodable_output_is_v3_and_never_breaks_the_advisory_contract",
      "test_v3_output_that_is_not_json_fails"}),
    ("M18 RecursionError 처리 삭제(적대 R2 F2)", SCRIPT,
     "    except (ValueError, RecursionError) as exc:\n",
     "    except ValueError as exc:\n",
     {"test_deeply_nested_json_is_v3_not_a_traceback"}),
    ("M17 advisory 예외 처리를 옛 (TypeError, KeyError) 로 좁힘", SCRIPT,
     "    except Exception as exc:  # 판정이 아니므로",
     "    except (TypeError, KeyError) as exc:  # 판정이 아니므로",
     {"test_the_advisory_turns_any_exception_into_a_line"}),
)

STEP11_BLOCK = re.compile(r"\n# ---- 11\. openspec validate \(a190\).*?echo \"OK: openspec validate\"\n", re.S)


def gate_text_for(name: str, text: str) -> str:
    """M8·M9 는 블록 단위 편집이라 정규식으로 만든다. 블록이 정확히 1번 맞아야 한다."""
    assert len(STEP11_BLOCK.findall(text)) == 1, "11단계 블록을 정확히 1번 찾지 못함"
    cut = STEP11_BLOCK.sub("\n", text)
    if name.startswith("M8"):
        loop = "for target in sdd-check test test-seams test-race vet; do"
        assert cut.count(loop) == 1
        return cut.replace(loop, "for target in sdd-check test test-seams test-race vet validate; do")
    assert cut.count("TOTAL_STEPS=11\n") == 1
    return cut.replace("TOTAL_STEPS=11\n", "TOTAL_STEPS=10\n")


def run_tests(copy: Path, env: dict) -> tuple[int, set[str], str]:
    proc = subprocess.run(
        [sys.executable, "-m", "unittest", "-v", *MODULES], cwd=copy, capture_output=True, text=True, env=env, timeout=900
    )
    failed = set(re.findall(r"^(?:FAIL|ERROR): (\w+) \(", proc.stderr, re.M))
    return proc.returncode, failed, proc.stderr


def main() -> int:
    try:
        return _main()
    finally:
        # 실패로 끝나도 사본·임시 디렉터리를 남기지 않음(첫 판들이 남겼다).
        # 이 실행이 **만든** 경로만 지움 — 미리 있던 같은 이름 경로는 건드리지 않음(Codex R2 #6).
        for leftover in CREATED:
            shutil.rmtree(leftover, ignore_errors=True)


CREATED: list[Path] = []


def _main() -> int:
    source = Path(sys.argv[1]).resolve()
    work = Path.home() / "tossos-work"
    copy = work / f"a190-mut-{os.getpid()}"
    tmp = work / f"a190-mut-tmp-{os.getpid()}"
    assert not copy.exists() and not tmp.exists()
    copy.mkdir(parents=True)
    CREATED.append(copy)
    head = subprocess.run(["git", "-C", str(source), "rev-parse", "HEAD"], capture_output=True, text=True, check=True).stdout.strip()
    dirty = subprocess.run(["git", "-C", str(source), "status", "--porcelain", "--", *FILES], capture_output=True, text=True, check=True).stdout
    print(f"# 원본 {source} HEAD {head}")
    print(f"# 대상 파일 미커밋 변경: {'없음' if not dirty else dirty.strip()}")
    print(f"# 사본 {copy} · TMPDIR {tmp} · OPENSPEC_TELEMETRY=0 · openspec {shutil.which('openspec')}")
    originals = {}
    for rel in FILES:
        (copy / rel).parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(source / rel, copy / rel)
        originals[rel] = (source / rel).read_bytes()
        print(f"# sha256 {hashlib.sha256(originals[rel]).hexdigest()}  {rel}")
    tmp.mkdir()
    CREATED.append(tmp)
    env = dict(os.environ, TMPDIR=str(tmp), OPENSPEC_TELEMETRY="0", DO_NOT_TRACK="1")

    rc, failed, log = run_tests(copy, env)
    print(f"\n## 대조군(무변이): rc={rc} 실패={sorted(failed)}")
    print(log.strip().splitlines()[-1])
    if rc != 0:
        print(log)
        print("대조군이 GREEN 이 아니다 — 멈춘다")
        return 1

    rows = []
    for name, rel, find, replace, expected in MUTANTS:
        target = copy / rel
        text = originals[rel].decode("utf-8")
        if find is None:
            mutated = gate_text_for(name, text)
        else:
            assert text.count(find) == 1, f"{name}: 찾을 문자열이 {text.count(find)}번 나온다"
            mutated = text.replace(find, replace)
        assert mutated != text, f"{name}: 변이가 내용을 바꾸지 않았다"
        target.write_text(mutated, encoding="utf-8")
        rc, failed, log = run_tests(copy, env)
        extra = ""
        if name.startswith("M4"):
            # 오늘의 실제 저장소에서 (나)에 strict 를 넣으면 다시 막히는가 — 원본 저장소를 cwd·ROOT 로 두고 변이 스크립트를 부른다.
            probe = copy / "probe"
            (probe / "tools" / "sdd").mkdir(parents=True, exist_ok=True)
            shutil.copy2(target, probe / "tools" / "sdd" / Path(SCRIPT).name)
            real = subprocess.run(
                [sys.executable, "-c",
                 "import importlib.util,sys,pathlib;"
                 f"s=importlib.util.spec_from_file_location('m',{str(probe / 'tools' / 'sdd' / Path(SCRIPT).name)!r});"
                 "m=importlib.util.module_from_spec(s);s.loader.exec_module(m);"
                 f"m.ROOT=pathlib.Path({str(source)!r});sys.exit(m.main([{A189!r}]))"],
                capture_output=True, text=True, env=env, timeout=900)
            extra = f" · 실제 저장소 a189 판정 rc={real.returncode}: " + " | ".join(
                line for line in (real.stdout + real.stderr).splitlines() if line.startswith("(나)"))[:300]
            shutil.rmtree(probe)
        hit = expected & failed
        verdict = "CAUGHT" if hit else ("OTHER" if failed else "SURVIVED")
        rows.append((name, verdict, sorted(hit), sorted(expected - failed), sorted(failed - expected), extra))
        print(f"\n## {name}: {verdict} (unittest rc={rc}){extra}")
        print(f"   기대 중 실패: {sorted(hit)}")
        print(f"   기대했으나 통과: {sorted(expected - failed)}")
        print(f"   기대 밖 실패: {sorted(failed - expected)}")
        target.write_bytes(originals[rel])
        assert target.read_bytes() == originals[rel], f"{name}: 되돌림 실패"

    rc, failed, log = run_tests(copy, env)
    print(f"\n## 대조군(끝, 되돌린 뒤): rc={rc} 실패={sorted(failed)}")
    for rel in FILES:
        assert (source / rel).read_bytes() == originals[rel], f"원본 {rel} 이 하네스 도중 바뀌었다"
    print("# 원본 대상 파일 바이트 불변 확인")
    print("\n| 변이 | 판정 | 잡은 시험 |")
    print("|---|---|---|")
    for name, verdict, hit, _, _, _ in rows:
        print(f"| {name} | {verdict} | {', '.join(hit) or '-'} |")
    shutil.rmtree(copy)
    shutil.rmtree(tmp)
    return 0 if rc == 0 and all(r[1] == "CAUGHT" for r in rows) else 1


if __name__ == "__main__":
    sys.exit(main())
