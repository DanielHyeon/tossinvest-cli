#!/usr/bin/env python3
"""a135 경량 변이 — 판정 · 실행자 핵심 줄을 하나씩 바꿔 TestA135 가 잡는지 잼(비례 원칙: 의무 아님, 시험 근거 확인용).

원본은 실행 전에 메모리에 두고, 각 변이 뒤 바이트 그대로 되돌린 뒤 sha256 을 대조함. 무변이 대조군이 GREEN 이 아니면 멈춤.
사용: python3 openspec/changes/a135-attestation-expiry-is-announced/analysis/mutation/run_mutations.py (저장소 루트에서)
"""
import hashlib, pathlib, subprocess, sys

ROOT = pathlib.Path(__file__).resolve().parents[5]
TARGET = ROOT / "internal/app/engine/attestation_watch.go"
MUTATIONS = [
    ("M1 critical 경계 >= 로", "case v.DaysLeft > attestationCriticalDays:", "case v.DaysLeft >= attestationCriticalDays:"),
    ("M2 보낸 단계 기록 삭제", "\tw.sent[key] = true\n}\n\n// notifyOutsideRecovery", "}\n\n// notifyOutsideRecovery"),
    ("M3 알림 꺼짐 무시", "if v.critical() && w.critical {", "if v.critical() {"),
    ("M4 판정 패닉 회수 삭제", "if rec := recover(); rec != nil {", "if rec := false; rec {"),
    ("M5 키에서 날수 삭제", 'return fmt.Sprintf("%s:%d", v.Stage, v.DaysLeft)', "return string(v.Stage)"),
    ("M6 범위 상실을 normal 로", "case stageHealthy, stageNotice:", "case stageHealthy, stageNotice, stageScopeLost:"),
    ("M7 신원 재설정 삭제(리뷰 P2)", "if identity != w.identity || w.sent == nil {", "if w.sent == nil {"),
    ("M8 outbox 키에서 신원 삭제(리뷰 P2)", 'Key:    "attestation-watch:" + key + ":" + w.identity,', 'Key:    "attestation-watch:" + key,'),
    ("M9 Notify 를 회수 경계 안에서 직접 부름(리뷰 P1)", "go func() { done <- w.notifier.Notify(ctx, event) }()", "done <- w.notifier.Notify(ctx, event)"),
    ("M10 무효 상태 날짜 리마인드 삭제", 'return string(v.Stage) + ":" + now.In(kst).Format("2006-01-02")', "return string(v.Stage)"),
    ("M11 endpoint 별 최신 대신 전체 최소", "if p.At.After(newest[key]) {", "if newest[key].IsZero() || p.At.Before(newest[key]) {"),
]

def run() -> int:
    p = subprocess.run(["go", "test", "-count=1", "-run", "TestA135", "./internal/app/engine/"],
                       cwd=ROOT, capture_output=True, text=True, timeout=600)
    return p.returncode

original = TARGET.read_bytes()
digest = hashlib.sha256(original).hexdigest()
if run() != 0:
    print("CONTROL RED — 멈춤"); sys.exit(2)
print("CONTROL GREEN")
caught = 0
try:
    for name, old, new in MUTATIONS:
        text = original.decode()
        if text.count(old) != 1:
            print(f"{name}: 변이 자리 {text.count(old)}곳 — 하네스 결함"); sys.exit(3)
        TARGET.write_text(text.replace(old, new))
        rc = run()
        TARGET.write_bytes(original)
        verdict = "CAUGHT" if rc != 0 else "SURVIVED"
        caught += rc != 0
        print(f"{name}: {verdict} (rc={rc})")
finally:
    TARGET.write_bytes(original)
assert hashlib.sha256(TARGET.read_bytes()).hexdigest() == digest, "원복 실패"
print(f"{caught}/{len(MUTATIONS)} CAUGHT, 원복 sha256 {digest[:16]} 일치")
