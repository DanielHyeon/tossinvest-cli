#!/usr/bin/env python3
"""a196 변이 원장 하네스 — 저장소가 아니라 사본에서 한 번에 한 판.

사용(저장소 루트에서): python3 openspec/changes/<change>/analysis/mutation/run_mutations.py
- 사본: ~/tossos-work/a196-mut-<pid>/ (go.mod·go.sum·internal/ 만 복사). 끝나면 지움.
- 시작 sha 단언: 저장소 HEAD 와 대상 파일 sha256 을 기록하고, 사본 파일이 저장소 파일과 같은지 확인.
- 무변이 대조군이 GREEN 이 아니면 멈춤.
- 변이마다 「닿음」 = 대상 문자열이 사본 파일에 정확히 1번 있음(0·2+ 이면 NOT-REACHED 로 기록하고 시험 안 돌림).
- 원복은 원본 바이트로 쓰고 sha256 으로 확인(심볼 세기 대신 파일 전체 대조).
"""
import hashlib, os, shutil, subprocess, sys, time
from pathlib import Path

repo = Path(subprocess.run(["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True, check=True).stdout.strip())
head = subprocess.run(["git", "-C", str(repo), "rev-parse", "HEAD"], capture_output=True, text=True, check=True).stdout.strip()
copy = Path.home() / "tossos-work" / f"a196-mut-{os.getpid()}"
assert not copy.exists(), copy

W = "internal/app/engine/attestation_watch.go"
S = "internal/soak/attest.go"
A = "internal/attest/attest.go"
ENGINE = ["go", "test", "-count=1", "-run", "A196|A135|Attestation", "./internal/app/engine"]
SOAK = ["go", "test", "-count=1", "-run", "A196|Supervised|BuildAttestation", "./internal/soak"]

MUTANTS = [
    ("M01", "① 기한을 만료로만 계산", W, "!proofAt.IsZero() && validity > 0", "false && !proofAt.IsZero() && validity > 0"),
    ("M02", "② min 을 max 로", W, "due.Before(deadline)", "due.After(deadline)"),
    ("M03", "③ 날수 신원에 발급 시각 되돌리기", W, "identity := v.Deadline.UTC().Format(time.RFC3339Nano)",
     'identity := v.IssuedAt.UTC().Format(time.RFC3339Nano) + "|" + v.ExpiresAt.UTC().Format(time.RFC3339Nano)'),
    ("M04", "④ 거절 key(outbox)에 발급 신원 되돌리기", W, 'Key:    "attestation-watch:" + key,',
     'Key:    "attestation-watch:" + key + ":" + v.IssuedAt.UTC().Format(time.RFC3339Nano),'),
    ("M05", "⑤ 누락 집합을 key 에서 빼기", W, 'return string(v.Stage) + ":" + strings.Join(v.Missing, ",") + ":" + day',
     'return string(v.Stage) + ":" + day'),
    ("M06", "⑥ D-2 경계 한 칸 밀기", W, "case v.DaysLeft > attestationCriticalDays:", "case v.DaysLeft >= attestationCriticalDays:"),
    ("M07", "⑦ 수용 경계 >= → >(soak)", S, "if age < 0 || age >= validity {", "if age < 0 || age > validity {"),
    ("M08a", "⑧ 기한을 가장 늦은 증거로(binding 최소 → 최대)", W, "if earliest.IsZero() || at.Before(earliest) {",
     "if earliest.IsZero() || at.After(earliest) {"),
    ("M08b", "⑧ binding 우회 — 목록 첫 증거로 기한 계산", W, "restartDeadline(att, v.ProofAt)",
     "restartDeadline(att, func() time.Time { if len(att.SupervisedBy) > 0 { return att.SupervisedBy[0].At }; return time.Time{} }())"),
    ("M09", "⑨ 0 이하 잔여를 imminent 로 접기", W, "if remaining <= 0 {", "if false {"),
    ("M10", "⑩ KST 날짜를 UTC 로", W, 'day := now.In(kst).Format("2006-01-02")', 'day := now.UTC().Format("2006-01-02")'),
    ("M11", "⑪ 거절 단계 key 에서 KST 날짜 삭제", W, '	return string(v.Stage) + ":" + day\n}', '	return string(v.Stage)\n}'),
    ("M12", "⑫ 누락 집합 정렬 삭제(MissingEndpoints)", A, "	sort.Strings(missing)\n", "	_ = sort.Strings\n"),
    # 구현 적대 리뷰 r1 처분(2026-11-23)으로 더함 — run-2 부터.
    ("N1", "P1-1 imminent 를 만료 문구로(감독 증거 갈래를 notice 에만)", W, "if v.DeadlineKind == deadlineProof {",
     "if v.DeadlineKind == deadlineProof && v.Stage == stageNotice {"),
    ("N2", "P2-1 「보냄」 재설정을 첫 호출에만(기한 비교 삭제)", W, "if identity != w.identity || w.sent == nil {", "if w.sent == nil {"),
    ("N3", "P2-2 유효 기간 > 0 가드 삭제", W, "!proofAt.IsZero() && validity > 0", "!proofAt.IsZero()"),
    ("N4", "R1 만료 무시(증거가 있으면 늘 감독 증거를 기한으로)", W, "due.Before(deadline)", "true || due.Before(deadline)"),
    ("N5", "P2-4 soak 수용 기간을 30일로 박기", S, "acceptSupervised(s.AccountRef, now, c.Validity, supervised)",
     "acceptSupervised(s.AccountRef, now, 30*24*time.Hour, supervised)"),
    ("N6", "P1-2 범위 상실을 늘 주문 실행 갈래로", W, "if mutations == len(v.Missing) {", "if true {"),
    ("N7", "P2-3 overdue key 에서 기한 삭제", W, 'return string(v.Stage) + ":" + deadline + ":" + day', 'return string(v.Stage) + ":" + day'),
]

def sha(p): return hashlib.sha256(Path(p).read_bytes()).hexdigest()

def run(cmd):
    env = dict(os.environ, GOFLAGS="-trimpath")
    p = subprocess.run(cmd, cwd=copy, capture_output=True, text=True, env=env)
    return p.returncode, p.stdout + p.stderr

def failed_tests(out):
    return sorted({l.split()[2] for l in out.splitlines() if l.startswith("--- FAIL:")})

def main():
    print(f"# a196 mutation run — repo HEAD {head}, copy {copy}, {time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime())}")
    copy.mkdir(parents=True)
    try:
        for f in ("go.mod", "go.sum"):
            shutil.copy2(repo / f, copy / f)
        shutil.copytree(repo / "internal", copy / "internal")
        for f in (W, S, A):
            assert sha(repo / f) == sha(copy / f), f"copy differs from repo: {f}"
            print(f"start sha256 {f} {sha(copy / f)}")
        rc_e, out_e = run(ENGINE); rc_s, out_s = run(SOAK)
        print(f"CONTROL engine rc={rc_e} soak rc={rc_s}")
        if rc_e or rc_s:
            print(out_e[-3000:], out_s[-3000:]); sys.exit("control is not GREEN — stop")
        for mid, desc, f, old, new in MUTANTS:
            path = copy / f
            orig = path.read_bytes(); text = orig.decode()
            n = text.count(old)
            if n != 1:
                print(f"{mid} NOT-REACHED (target occurs {n}x) — {desc}"); continue
            path.write_text(text.replace(old, new))
            rc_e, out_e = run(ENGINE); rc_s, out_s = run(SOAK)
            path.write_bytes(orig)
            assert sha(path) == sha(repo / f), f"revert failed for {mid}"
            build_fail = "[build failed]" in out_e + out_s or "cannot use" in out_e + out_s
            verdict = "BUILD-FAIL" if build_fail else ("CAUGHT" if (rc_e or rc_s) else "SURVIVED")
            killers = failed_tests(out_e) + failed_tests(out_s)
            print(f"{mid} {verdict} engine_rc={rc_e} soak_rc={rc_s} — {desc}")
            print(f"    killed by: {', '.join(killers) if killers else '-'}")
            if build_fail:
                print("    " + (out_e + out_s)[-800:].replace("\n", "\n    "))
        for f in (W, S, A):
            assert sha(copy / f) == sha(repo / f), f"final revert check {f}"
        print("final: copy files equal repo files (sha256)")
    finally:
        shutil.rmtree(copy, ignore_errors=True)

main()
