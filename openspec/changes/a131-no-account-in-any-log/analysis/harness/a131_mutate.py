#!/usr/bin/env python3
"""a131 변이 하네스 — 사본에서만 변이하고, 무변이 대조군이 GREEN 이어야 판을 돎(a129 하네스와 같은 틀).

사용: python3 a131_mutate.py [--keep]
- 저장소 루트는 이 파일 위치에서 유도함. 사본 = `git archive HEAD` + 워킹트리의 a131 파일 덮어쓰기, 사본 경로에 pid.
- 각 변이: 원문 조각이 사본 파일에 **정확히 한 번** 있어야 함(닿음 단언) → 바꿈 → 대상 시험 → 시험 실패면 CAUGHT → 원복.
- 컴파일 실패는 BUILD-FAILED 로 따로 셈(시험이 잡은 것이 아님).
"""

from __future__ import annotations

import os
import shutil
import subprocess
import sys
import tarfile
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[5]
CHANGE = "openspec/changes/a131-no-account-in-any-log"
PACKAGES = ["./internal/redact/", "./internal/obs/", "./internal/official/", "./internal/audit/",
            "./internal/soak/", "./internal/verifylive/", "./cmd/tossctl/"]
RUN = "A131|TestSoakAttestWritesAVerifiableAttestation|TestReconcileLineMatchesTheDesignShape|TestRecordCarriesNoUnmaskedAccountNumber|TestTheTransitionLogLineIsCountable"


def changed_files() -> list[str]:
    """a131 이 바꾼 Go 파일 — 작업 트리의 수정·새 파일 중 a131 표면(병행 a130 의 journal·콘솔 파일은 제외)."""
    out = subprocess.run(["git", "-C", str(ROOT), "status", "--porcelain", "--untracked-files=all"],
                         capture_output=True, text=True, check=True).stdout
    surface = ("internal/redact/", "internal/obs/", "internal/official/", "internal/audit/", "internal/soak/",
               "internal/verifylive/", "cmd/tossctl/")
    return [line[3:] for line in out.splitlines() if line[3:].endswith(".go") and line[3:].startswith(surface)]


MUTANTS = [
    ("M1 NewLogger does not wrap the redacting handler", "internal/obs/log.go",
     "\thandler = redact.Handler(handler)\n", "\t_ = redact.Handler\n"),
    ("M2 the handler keeps account keys", "internal/redact/sinks.go",
     "\tif IsAccountKey(a.Key) {\n\t\treturn slog.Attr{}, false\n\t}\n", ""),
    ("M3 the mask rule is removed", "internal/redact/redact.go",
     "\ts = maskPattern.ReplaceAllString(s, Placeholder)\n", ""),
    ("M4 the official accounts read registers nothing", "internal/official/reads.go",
     "\tregisterAccountNumbers(raw)\n", ""),
    ("M5 engine run writes stdout unwrapped", "cmd/tossctl/engine.go",
     "\t\t\tcmd.SetOut(redact.Writer(cmd.OutOrStdout()))\n", ""),
    ("M6 audit Record stores the entry unredacted", "internal/audit/audit.go",
     "\te = RedactEntry(e)\n", ""),
    ("M7 the soak record writes the account in the clear", "internal/soak/record.go",
     "\tline, err = redact.SealRecordLine(line, r.key)\n\tif err != nil {\n\t\treturn fmt.Errorf(\"soak: sealing the record line: %w\", err)\n\t}\n", ""),
    ("M8 the soak reader skips the seal on interior lines", "internal/soak/record.go",
     "\t\tif c, err = openCycleAccount(c, opener); err != nil {\n\t\t\treturn fmt.Errorf(\"soak: %s line %d: %w\", path, pendingLine, err)\n\t\t}\n", ""),
    ("M9 the position-id rule is removed", "internal/redact/redact.go",
     "\treturn derivedIDPattern.ReplaceAllString(s, \"${1}${2}-[redacted]\")\n", "\treturn s\n"),
    ("M10 bound attributes are handed to the inner handler at bind time", "internal/redact/sinks.go",
     "\tbound = append(append(bound, h.bound...), attrs...)\n\treturn &handler{inner: h.inner, bound: bound}\n",
     "\t_ = bound\n\treturn &handler{inner: h.inner.WithAttrs(attrs), bound: h.bound}\n"),
    ("M11 the preview seals (writes the key)", "cmd/tossctl/scrub_account_records.go",
     "\tif !apply {\n\t\tchanged, err := redact.RecordLineNeedsRewrite(line)\n\t\treturn line, changed, err\n\t}\n",
     "\tif !apply {\n\t\tk, _ := redact.PrepareSealKey(t.path)\n\t\tnext, err := redact.SealRecordLine(line, k)\n\t\treturn line, !bytes.Equal(next, line), err\n\t}\n"),
    ("M12 the verify record writes the account in the clear", "internal/verifylive/record.go",
     "\tline, err = redact.SealRecordLine(line, r.key)\n\tif err != nil {\n\t\treturn fmt.Errorf(\"verifylive: sealing the record line: %w\", err)\n\t}\n", ""),
    ("M13 the reconcile line is written unsealed", "internal/verifylive/record.go",
     "\tif line, err = redact.SealRecordLine(line, key); err != nil {\n\t\treturn fmt.Errorf(\"verifylive: sealing the reconcile line: %w\", err)\n\t}\n", "\t_ = key\n"),
    ("M14 the strict reconcile reader does not open the seal", "internal/verifylive/reconcile_record.go",
     "\tif err := openRecordEntries(entries, path); err != nil {\n\t\treturn nil, nil, refuse(RefuseRecordUndecodable, err.Error())\n\t}\n", ""),
    ("M15 the digit-run rule is removed", "internal/redact/redact.go",
     "\t\t\tif reg.digits[onlyDigits(run)] {\n", "\t\t\tif false && reg.digits[onlyDigits(run)] {\n"),
    ("M16 main leaves cobra's stderr unwrapped", "cmd/tossctl/main.go",
     "\troot.SetErr(redact.Writer(os.Stderr))\n", ""),
    ("M17 the scrub replaces a file that moved", "cmd/tossctl/scrub_account_records.go",
     "\tif now.Size() != plan.info.Size() || !now.ModTime().Equal(plan.info.ModTime()) {\n",
     "\tif false && (now.Size() != plan.info.Size() || !now.ModTime().Equal(plan.info.ModTime())) {\n"),
    ("M18 the attest output prints the masked account", "cmd/tossctl/soak.go",
     "\tfmt.Fprintf(out, \"  soak days    %d\\n\", attestation.SoakDays)\n",
     "\tfmt.Fprintf(out, \"  account      %s\\n\", attest.Mask(attestation.AccountRef))\n\tfmt.Fprintf(out, \"  soak days    %d\\n\", attestation.SoakDays)\n"),
    ("M19 an unopenable seal reads as an empty account", "internal/soak/account_seal.go",
     "\tif !ok {\n\t\treturn c, errSealUnopenable\n\t}\n", "\tif !ok {\n\t\treturn c, nil\n\t}\n"),
    ("M20 a missing key is recreated for a sealed record", "internal/redact/record.go",
     "\t\tif data, rerr := os.ReadFile(recordPath); rerr == nil && bytes.Contains(data, []byte(`\"`+SealedField+`\"`)) {\n",
     "\t\tif data, rerr := os.ReadFile(recordPath); false && rerr == nil && bytes.Contains(data, []byte(`\"`+SealedField+`\"`)) {\n"),
    ("M21 Append does not reopen a replaced record", "internal/soak/record.go",
     "\tif err := reopenIfReplaced(r); err != nil {\n\t\treturn err\n\t}\n", ""),
    ("M22 the scrub drops the replaced tail", "cmd/tossctl/scrub_account_records.go",
     "\treturn drainReplacedTail(plan)\n", "\treturn nil\n"),
    ("M23 the handler lets a panicking value kill the line", "internal/redact/sinks.go",
     "\t\tif recover() != nil {\n", "\t\tif false {\n"),
    ("M24 identity keys are scrubbed", "internal/redact/record.go",
     "\t\tif k == SealedField || identityKey(k) {\n", "\t\tif k == SealedField {\n"),
    ("M25 the sealed line keeps format 1", "internal/redact/record.go",
     "\t\t\t\tobj[\"format_version\"] = json.Number(fmt.Sprint(SealedFormatVersion))\n", ""),
]


def run(cmd: list[str], cwd: Path) -> tuple[int, str]:
    env = dict(os.environ, GOFLAGS="-trimpath")
    p = subprocess.run(cmd, cwd=cwd, capture_output=True, text=True, env=env)
    return p.returncode, p.stdout + p.stderr


def tests(copy: Path) -> tuple[str, str]:
    rc, out = run(["go", "test", "-count=1", "-run", RUN, *PACKAGES], copy)
    failed = [l.strip() for l in out.splitlines() if l.strip().startswith("--- FAIL") or "panic:" in l]
    if "[build failed]" in out or "[setup failed]" in out:
        return "BUILD", "; ".join([l for l in out.splitlines() if ".go:" in l][:4])
    return ("GREEN" if rc == 0 else "RED"), "; ".join(failed[:6])


def main() -> int:
    keep = "--keep" in sys.argv
    head = subprocess.run(["git", "-C", str(ROOT), "rev-parse", "--short", "HEAD"],
                          capture_output=True, text=True, check=True).stdout.strip()
    files = changed_files()
    copy = Path(tempfile.mkdtemp(prefix=f"a131-mutate-{os.getpid()}-"))
    try:
        archive = copy / "head.tar"
        subprocess.run(["git", "-C", str(ROOT), "archive", "--format=tar", "-o", str(archive), "HEAD"], check=True)
        with tarfile.open(archive) as tf:
            tf.extractall(copy / "repo", filter="data")
        archive.unlink()
        repo = copy / "repo"
        for rel in files:
            src = ROOT / rel
            if src.exists():
                (repo / rel).parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(src, repo / rel)
        print(f"# a131 mutation — HEAD {head} + {len(files)} working-tree a131 files · copy {repo}")
        state, why = tests(repo)
        print(f"CONTROL {state} {why}")
        if state != "GREEN":
            print("# control is not GREEN — no mutant is judged")
            return 2
        caught = 0
        for name, rel, old, new in MUTANTS:
            path = repo / rel
            text = path.read_text()
            if text.count(old) != 1:
                print(f"UNREACHED {name} — fragment count {text.count(old)} in {rel}")
                continue
            path.write_text(text.replace(old, new, 1))
            try:
                state, why = tests(repo)
            finally:
                path.write_text(text)
            verdict = {"GREEN": "SURVIVED", "RED": "CAUGHT", "BUILD": "BUILD-FAILED"}[state]
            caught += 1 if state == "RED" else 0
            print(f"{verdict} {name} — {why}", flush=True)
        print(f"# {caught}/{len(MUTANTS)} caught")
        return 0 if caught == len(MUTANTS) else 1
    finally:
        if not keep:
            shutil.rmtree(copy, ignore_errors=True)


if __name__ == "__main__":
    raise SystemExit(main())
