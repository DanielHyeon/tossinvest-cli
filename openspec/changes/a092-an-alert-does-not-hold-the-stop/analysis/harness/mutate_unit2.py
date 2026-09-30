#!/usr/bin/env python3
"""a092 착지 단위 ② 변이 하네스 — RecordAlert 입구 · 기록 전용 어댑터/통지자 · K1 신원 · exit 배선을 사본에서 변이함.

사용: python3 mutate_unit2.py <scratch-dir> <own-untracked-file>... [--only REGEX]

- 사본: git 추적 파일(작업 트리 내용) + 이 로트의 미추적 파일만 복사함(남의 미추적 RED 파일이 사본에 들어오지 않게).
- 원장 첫 줄: HEAD sha 와 사본에 딸려 간 미커밋 추적 파일 목록.
- 무변이 대조군이 GREEN 이 아니면 멈춤. 변이마다 정확히 한 곳(count==1)을 바꿈.
- 판정 셋: CAUGHT(시험이 실패 — 실패한 시험 이름을 적음) · SURVIVED · BUILD-FAIL(컴파일 실패 — 닿지 않은 것이지 잡은 것이 아님).
- `--set 25.6`: a066 완화 통지 입구 이행 변이. `--set 25.7`: 단위 ③ 잠금 범위 · 원칙 E · claim-held 등급.
- 사본 디렉터리에 pid 를 붙임 — 두 판이 한 사본을 쓰지 않게. 한 번에 한 판.
"""
import os
import re
import shutil
import subprocess
import sys
from pathlib import Path

ROOT = Path(subprocess.run(["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True, check=True).stdout.strip())
RA = "internal/journal/record_alert.go"
RO = "internal/obs/record_only.go"
MODE = "internal/obs/mode.go"
EW = "internal/app/engine/exitwiring.go"
ERO = "internal/app/engine/exit_record_only.go"
ENG = "cmd/tossctl/engine.go"

MUTANTS = [
    # 원장 입구
    ("U01 RecordAlert never re-arms (remindAfter ignored)", RA,
     "id, owed, err := j.recordAlertTx(ctx, tx, a, remindAfter)", "id, owed, err := j.recordAlertTx(ctx, tx, a, 0)"),
    ("U02 RecordAlert takes a lease", RA,
     "\tif err := tx.Commit(); err != nil {\n\t\treturn 0, false, fmt.Errorf(\"journal: committing alert %s: %w\", key, err)",
     "\tif owed {\n\t\tif _, cerr := acquireAlertClaimTx(ctx, tx, id, \"mutant\", j.clk.Now(), j.alertLease); cerr != nil {\n\t\t\treturn 0, false, cerr\n\t\t}\n\t}\n\tif err := tx.Commit(); err != nil {\n\t\treturn 0, false, fmt.Errorf(\"journal: committing alert %s: %w\", key, err)"),
    # 기록 전용 어댑터
    ("U03 critical Notify goes through the sync path", RO,
     "\treturn n.recordCritical(ctx, e, n.remindAfter())\n}\n\n// AnnounceOperatingMode",
     "\treturn n.notifyCritical(ctx, e)\n}\n\n// AnnounceOperatingMode"),
    ("U04 announcer goes through the sync path", RO,
     "\tn.logEvent(e, SeverityOf(e.Type))\n\treturn n.recordCritical(ctx, e, n.remindAfter())",
     "\treturn n.Notify(ctx, e)"),
    ("U05 exit record passes a zero reminder window", RO,
     "\treturn n.recordCritical(ctx, e, n.remindAfter())\n}\n\n// AnnounceOperatingMode",
     "\treturn n.recordCritical(ctx, e, 0)\n}\n\n// AnnounceOperatingMode"),
    ("U06 no journal falls back to a publish", RO,
     "\t\t\t\tFieldDetail, \"no journal is wired, so this critical alert is neither durable nor sent\")\n\t\t}\n\t\treturn nil",
     "\t\t\t\tFieldDetail, \"no journal is wired, so this critical alert is neither durable nor sent\")\n\t\t}\n\t\tn.publishBestEffort(ctx, e, SeverityCritical)\n\t\treturn nil"),
    ("U07 record failure does not latch", RO,
     "\t\tif n.Gate != nil {\n\t\t\tn.Gate.Block(execgw.ReasonAlertUndelivered",
     "\t\tif n.Gate != nil && false {\n\t\t\tn.Gate.Block(execgw.ReasonAlertUndelivered"),
    ("U08 record failure does not escalate", RO,
     "\t\tn.escalate(ctx, e)\n\t\treturn fmt.Errorf(\"obs: recording a critical alert: %w\", err)",
     "\t\treturn fmt.Errorf(\"obs: recording a critical alert: %w\", err)"),
    ("U09 record failure is swallowed", RO,
     "\t\treturn fmt.Errorf(\"obs: recording a critical alert: %w\", err)\n\t}\n\treturn nil\n}",
     "\t\treturn nil\n\t}\n\treturn nil\n}"),
    ("U10 record failure is not logged", RO,
     "\t\t\tn.Log.Error(EventAlertUndelivered, err, FieldTriggerEvent, string(e.Type))",
     "\t\t\t_ = err"),
    ("U11 record outside the notifier lock", RO,
     "\tn.mu.Lock()\n\t_, _, err := n.Journal.RecordAlert(ctx, record, remindAfter)",
     "\t_, _, err := n.Journal.RecordAlert(ctx, record, remindAfter)\n\tn.mu.Lock()"),
    ("U12 nil notifier guard dropped (Notify)", RO,
     "func (r RecordOnly) Notify(ctx context.Context, e Event) error {\n\tn := r.N\n\tif n == nil {\n\t\treturn nil\n\t}",
     "func (r RecordOnly) Notify(ctx context.Context, e Event) error {\n\tn := r.N"),
    # K1 신원 · 사건 구성 공유
    ("U13 key drops the transition id", RO,
     "Key:   \"operating_mode:\" + rec.AccountRef + \":\" + rec.Mode + \":\" + rec.ID,",
     "Key:   \"operating_mode:\" + rec.AccountRef + \":\" + rec.Mode,"),
    ("U14 the two announcers diverge", RO,
     "\te := operatingModeEvent(previous, rec)\n",
     "\te := operatingModeEvent(previous, rec)\n\te.Title += \" (recorded)\"\n"),
    ("U15 sync announcer keeps the old key", MODE,
     "\treturn n.Notify(ctx, operatingModeEvent(previous, rec))",
     "\te := operatingModeEvent(previous, rec)\n\te.Key = \"operating_mode:\" + rec.AccountRef + \":\" + rec.Mode\n\treturn n.Notify(ctx, e)"),
    # exit 배선
    ("U16 exit loop shares the engine Retrier", ERO,
     "\texit := *shared\n\texit.Announcer = announcer\n\treturn &exit", "\t_ = announcer\n\treturn shared"),
    ("U17 exit Retrier keeps the sync announcer", ERO,
     "\texit.Announcer = announcer\n", ""),
    ("U18 exit floor keeps the shared retrier", ERO,
     "\texit.retrier = retrier\n", ""),
    ("U19 exit floor is the shared floor", EW,
     "opts.Floor = exitSideFloor(c.exitFloor, exitRetrier)", "opts.Floor = c.exitFloor"),
    ("U20 exit alerts are the sync notifier", EW,
     "\t\topts.Alerts = recordOnly\n", "\t\topts.Alerts = c.Notifier\n"),
    ("U21 exit announcer default dropped", EW,
     "\tif opts.Announcer == nil && c.Notifier != nil {\n\t\topts.Announcer = recordOnly\n\t}\n", ""),
    ("U22 exit retrier is the shared one", EW,
     "\topts.Retrier = exitRetrier\n", "\topts.Retrier = c.Retrier\n"),
    ("U23 assembly hands the exit loop the sync notifier", ENG,
     "\t\tEscalate: ectx.Journal,\n\t\t// Announcer · Alerts",
     "\t\tEscalate: ectx.Journal,\n\t\tAnnouncer: ectx.Notifier,\n\t\t// Announcer · Alerts"),
    ("U24 exit retrier copy mutates the shared one", ERO,
     "\texit := *shared\n\texit.Announcer = announcer\n\treturn &exit",
     "\tshared.Announcer = announcer\n\treturn shared"),
]

# 25.6 — a066 완화 통지의 입구 이행. `--set 25.6` 으로 고름.
RRC = "internal/app/engine/risk_relaxation_command.go"
PPC = "internal/app/engine/position_policy_command.go"
RELAX_MUTANTS = [
    ("R01 nil recorder guard dropped", RRC,
     "\tif notices == nil {\n\t\t// 알림기가 배선되지 않은 엔진", "\tif false {\n\t\t// 알림기가 배선되지 않은 엔진"),
    ("R02 no recorder reported as notified", RRC,
     "\t\tresult.NotifyError = obs.ErrAlertNotDurable.Error()\n\t\treturn result", "\t\tresult.Notified = true\n\t\treturn result"),
    ("R03 reminder window not zero", RRC, "\t}, 0)\n\tif err != nil {", "\t}, time.Hour)\n\tif err != nil {"),
    ("R04 hang-up cancels the notice", RRC, "notices.RecordCritical(context.WithoutCancel(ctx), obs.Event{", "notices.RecordCritical(ctx, obs.Event{"),
    ("R05 approval not trimmed in payload", RRC, "\t\t\t\"approval\": strings.TrimSpace(approval), \"released_at\"", "\t\t\t\"approval\": approval, \"released_at\""),
    ("R06 key drops the release seq", RRC, "Key:   EventRiskRelaxation + \"|\" + kind + \"|\" + strconv.FormatInt(seq, 10),", "Key:   EventRiskRelaxation + \"|\" + kind + \"|\" + strconv.FormatInt(0*seq, 10),"),
    ("R07 repository regains EnqueueAlert", RRC,
     "\tReleaseRiskOverageLatch(context.Context, journal.RiskOverageLatchReleaseRequest) (journal.RiskOverageLatchReleaseRecord, error)\n}",
     "\tReleaseRiskOverageLatch(context.Context, journal.RiskOverageLatchReleaseRequest) (journal.RiskOverageLatchReleaseRecord, error)\n\tEnqueueAlert(context.Context, journal.Alert) (int64, error)\n}"),
    ("R08 typed-nil notifier stored", PPC,
     "\tif ectx.Notifier != nil {\n\t\tservice.notices = ectx.Notifier\n\t}", "\tservice.notices = ectx.Notifier"),
    ("R09 constructor drops the recorder", PPC,
     "\tif ectx.Notifier != nil {\n\t\tservice.notices = ectx.Notifier\n\t}", ""),
    ("R10 entry-lock release passes no recorder", RRC, "notifyRelaxation(ctx, s.notices, \"entry_lock\"", "notifyRelaxation(ctx, nil, \"entry_lock\""),
    ("R11 latch release passes no recorder", RRC, "notifyRelaxation(ctx, s.notices, \"overage_latch\"", "notifyRelaxation(ctx, nil, \"overage_latch\""),
    ("R12 RecordCritical grades the event", RO,
     "\tn.logEvent(e, SeverityCritical)\n\treturn n.recordCritical(ctx, e, remindAfter)",
     "\tif SeverityOf(e.Type) != SeverityCritical {\n\t\tn.publishBestEffort(ctx, e, SeverityOf(e.Type))\n\t\treturn nil\n\t}\n\treturn n.recordCritical(ctx, e, remindAfter)"),
    ("R13 RecordCritical without journal returns nil", RO, "\t\treturn ErrAlertNotDurable\n", "\t\treturn nil\n"),
    ("R14 RecordCritical ignores the caller window", RO,
     "\tn.logEvent(e, SeverityCritical)\n\treturn n.recordCritical(ctx, e, remindAfter)",
     "\tn.logEvent(e, SeverityCritical)\n\treturn n.recordCritical(ctx, e, n.remindAfter())"),
    ("R15 record failure reported as notified", RRC,
     "\tif err != nil {\n\t\tresult.NotifyError = err.Error()\n\t\treturn result\n\t}",
     "\tif err != nil {\n\t\tresult.Notified = true\n\t\treturn result\n\t}"),
]
RELAX_TESTS = [
    ["go", "test", "-count=1", "-run", "TestA066|TestA092|PositionPolicyCommand", "./internal/app/engine"],
    ["go", "test", "-count=1", "-run", "TestA092", "./internal/obs"],
]

# 25.7 — 착지 단위 ③ 잠금 범위 · 원칙 E · logClaimHeld 등급. `--set 25.7`.
NOT = "internal/obs/notifier.go"
LOCK_MUTANTS = [
    ('L01 lock covers the transport again', NOT,
     '\tn.mu.Unlock()\n\n\tn.logClaimStolen(string(e.Type), claim)\n\tsent, lost, verdict := n.deliver(ctx, claim.ID, claim.Token, e)\n',
     '\n\tn.logClaimStolen(string(e.Type), claim)\n\tsent, lost, verdict := n.deliver(ctx, claim.ID, claim.Token, e)\n\tn.mu.Unlock()\n'),
    ('L02 escalating sites latch unconditionally', NOT,
     '\t\tn.Gate.BlockUnlessClearedSince(execgw.ReasonAlertUndelivered, v.epoch, v.detail)\n\t}\n\tincluded, err := n.escalate(ctx, e)',
     '\t\tn.Gate.Block(execgw.ReasonAlertUndelivered, v.detail)\n\t}\n\tincluded, err := n.escalate(ctx, e)'),
    ('L03 vanished site latches unconditionally', NOT,
     '\t\t\t\tn.Gate.BlockUnlessClearedSince(execgw.ReasonAlertUndelivered, v.epoch, v.detail)\n\t\t\t}\n\t\t\treturn false, true, latchVerdict{}',
     '\t\t\t\tn.Gate.Block(execgw.ReasonAlertUndelivered, v.detail)\n\t\t\t}\n\t\t\treturn false, true, latchVerdict{}'),
    ('L04 epoch read after the apply point', NOT,
     '\t\tepoch = n.Gate.ClearEpoch(execgw.ReasonAlertUndelivered)\n\t}\n\tn.hook("epoch:" + site)',
     '\t}\n\tn.hook("epoch:" + site)\n\tif n.Gate != nil {\n\t\tepoch = n.Gate.ClearEpoch(execgw.ReasonAlertUndelivered)\n\t}'),
    ('L05 epoch pinned before any evidence', NOT,
     '\t\tepoch = n.Gate.ClearEpoch(execgw.ReasonAlertUndelivered)\n\t}\n\tn.hook("epoch:" + site)',
     '\t\tepoch = 0 * n.Gate.ClearEpoch(execgw.ReasonAlertUndelivered)\n\t}\n\tn.hook("epoch:" + site)'),
    ('L06 failed escalation does not latch', NOT,
     '\tif included && err != nil && n.Gate != nil {',
     '\tif false && included && err != nil && n.Gate != nil {'),
    ('L07 unconditional latch without an escalation', NOT,
     '\tif included && err != nil && n.Gate != nil {',
     '\tif (!included || err != nil) && n.Gate != nil {'),
    ('L08 escalation skipped when the latch was released', NOT,
     '\t\tn.Gate.BlockUnlessClearedSince(execgw.ReasonAlertUndelivered, v.epoch, v.detail)\n\t}\n\tincluded, err := n.escalate(ctx, e)',
     '\t\tif applied, _ := n.Gate.BlockUnlessClearedSince(execgw.ReasonAlertUndelivered, v.epoch, v.detail); !applied {\n\t\t\treturn\n\t\t}\n\t}\n\tincluded, err := n.escalate(ctx, e)'),
    ('L09 escalate never reports inclusion', NOT,
     '\treturn true, err\n}',
     '\treturn false, err\n}'),
    ('L10 escalate hides its failure', NOT,
     '\treturn true, err\n}',
     '\treturn true, nil\n}'),
    ('L11 claim-held line deleted', NOT,
     '\tn.Log.Event(EventAlertClaimHeld,\n\t\tFieldTriggerEvent, eventType,',
     '\tn.Log.Event(EventAlertClaimLost,\n\t\tFieldTriggerEvent, eventType,'),
    ('L12 claim-held back to WARN', NOT,
     '\tn.Log.Event(EventAlertClaimHeld,',
     '\tn.Log.Warn(EventAlertClaimHeld,'),
    ('L13 vanished site never latches', NOT,
     '\t\t\t\tn.Gate.BlockUnlessClearedSince(execgw.ReasonAlertUndelivered, v.epoch, v.detail)\n\t\t\t}\n\t\t\treturn false, true, latchVerdict{}',
     '\t\t\t\t_ = v\n\t\t\t}\n\t\t\treturn false, true, latchVerdict{}'),
    ('L14 acknowledge counts outside the lock', NOT,
     '\tremaining, err := n.Journal.UndeliveredCount(ctx)',
     '\tn.mu.Unlock()\n\tn.mu.Lock()\n\tremaining, err := n.Journal.UndeliveredCount(ctx)'),
    ('L15 claim taken before the lock', NOT,
     '\tn.mu.Lock()\n\n\t// Before the first claim',
     '\tclaim0, _ := n.Journal.ClaimAlertForDelivery(ctx, record, n.remindAfter(), n.claimant())\n\t_ = claim0\n\tn.mu.Lock()\n\n\t// Before the first claim'),
    ('L16 unrecorded site drops its verdict', NOT,
     '\t\t\t// eventually; nothing else does.\n\t\t\treturn false, false, verdict',
     '\t\t\t// eventually; nothing else does.\n\t\t\t_ = verdict\n\t\t\treturn false, false, latchVerdict{}'),
    ('L17 exhausted site drops its verdict', NOT,
     '\t\t\t"alert_id", id)\n\t}\n\treturn false, false, verdict\n}',
     '\t\t\t"alert_id", id)\n\t}\n\t_ = verdict\n\treturn false, false, latchVerdict{}\n}'),
    ("L18 claim moved outside the lock (pin)", NOT,
     "\tn.mu.Lock()\n\n\t// Before the first claim, not after: if the lease this ledger issues cannot\n\t// cover this sender's budget, every claim below is already unsound and the\n\t// operator should read that once, here, rather than infer it from a duplicate.\n\tn.checkAlertLease()\n\n\tclaim, err := n.Journal.ClaimAlertForDelivery(ctx, record, n.remindAfter(), n.claimant())\n",
     "\n\t// Before the first claim, not after: if the lease this ledger issues cannot\n\t// cover this sender's budget, every claim below is already unsound and the\n\t// operator should read that once, here, rather than infer it from a duplicate.\n\tn.checkAlertLease()\n\n\tclaim, err := n.Journal.ClaimAlertForDelivery(ctx, record, n.remindAfter(), n.claimant())\n\tn.mu.Lock()\n"),
    # 21.5 A-4 귀속: 잠금이 좁아진 뒤 a096/a097 배제 시험을 지키는 것이 임차인지 잰다 — 임차를 무시하게 한 원장.
    ("L19 lease ignored (A-4 attribution)", "internal/journal/alert_claim.go",
     "WHERE id = ? AND state = ? AND `+alertClaimable,",
     "WHERE id = ? AND state = ? AND (1=1 OR `+alertClaimable+`)`,"),
    # 25라운드 codex P0 수리의 반증: 반납 행 없음을 다시 선점으로 되돌림.
    ("L20 release NotFound read as preemption (r25 codex P0)", NOT,
     "\tcase released.Outcome != journal.SettleAlreadySettled && released.Outcome != journal.SettleLeaseLost:",
     "\tcase false && released.Outcome != journal.SettleAlreadySettled && released.Outcome != journal.SettleLeaseLost:"),
]
LOCK_TESTS = [
    ["go", "test", "-count=1", "./internal/obs"],
]

# 25.9 — 착지 단위 ④ 모드 커밋 순서 · 투영 · 배선 · 완화. `--set 25.9`.
OM = 'internal/journal/operating_mode.go'
MG = 'internal/execgw/modegate.go'
MW = 'internal/app/engine/mode_projection_wiring.go'
GW = 'internal/app/engine/gateway.go'
MO = 'internal/app/engine/modeops.go'
MT = 'internal/app/engine/mode_control_transport_unix.go'
CE = 'cmd/tossctl/engine_mode_release.go'
CA = 'cmd/tossctl/engine_alerts.go'
MODE_MUTANTS = [
    ('M01 current mode by wall clock', OM,
     'const modeLatestOrder = " ORDER BY rowid DESC LIMIT 1"',
     'const modeLatestOrder = " ORDER BY created_at DESC, rowid DESC LIMIT 1"'),
    ('M02 history by wall clock', OM,
     '" WHERE account_ref = ? ORDER BY rowid",',
     '" WHERE account_ref = ? ORDER BY created_at, rowid",'),
    ('M03 transition carries no seq', OM,
     'CreatedAt: now, Seq: seqRow,',
     'CreatedAt: now, Seq: 0 * seqRow,'),
    ('M04 restore carries no seq', OM,
     '\t\t\tSeq:        snapshot.Seq,\n',
     ''),
    ('M05 fence removed', MG,
     '\tif rec.Seq <= g.modeSeq {\n\t\treturn\n\t}',
     '\tif false {\n\t\treturn\n\t}'),
    ('M06 fence admits equal', MG,
     '\tif rec.Seq <= g.modeSeq {',
     '\tif rec.Seq < g.modeSeq {'),
    ('M07 revision on every replacement', MG,
     '\tif !had {\n\t\tg.revision++\n\t}\n}',
     '\tg.revision++\n\t_ = had\n}'),
    ('M08 replacement leaves a gap', MG,
     '\tg.latches[ReasonOperatingModeBlocked] = detail\n',
     '\tdelete(g.latches, ReasonOperatingModeBlocked)\n\tg.mu.Unlock()\n\tg.mu.Lock()\n\tg.latches[ReasonOperatingModeBlocked] = detail\n'),
    ('M09 assembly does not bind the projection', GW,
     '\tif err := bindOperatingModeProjection(ctx, in.journal, entry, in.accountRef, in.logger); err != nil {',
     '\tif err := error(nil); err != nil {'),
    ('M10 restore failure refuses start', MW,
     '\t\tgate.Block(execgw.ReasonOperatingModeBlocked, operatingModeRestoreFailed)\n',
     '\t\treturn err\n'),
    ('M11 restore failure leaves entries open', MW,
     '\t\tgate.Block(execgw.ReasonOperatingModeBlocked, operatingModeRestoreFailed)\n',
     ''),
    ('M12 nil audit accepted', MO,
     '\tif o.auditor == nil {\n\t\treturn ModeReleaseResult{}, fmt.Errorf("%w: the engine has no audit log", ErrModeReleaseUnavailable)\n\t}',
     '\tif false {\n\t}'),
    ('M13 announcer is the sync notifier', MO,
     '\t\tannouncer: obs.RecordOnly{N: notifier},',
     '\t\tannouncer: notifier,'),
    ('M14 result not re-read', MO,
     '\tresult.Mode, result.Seq = current.Mode, current.Seq',
     '\tresult.Mode, result.Seq = to, current.Seq'),
    ('M15 operator not in the cause', MO,
     'reason + " | operator: " + operator,',
     'reason,'),
    ('M16 operator not required', MO,
     'case operator == "", approval == "", reason == "":',
     'case approval == "", reason == "":'),
    ('M17 HALT_ALL target allowed', MO,
     'case to != journal.ModeNormal && to != journal.ModeEntryBlocked:',
     'case to != journal.ModeNormal && to != journal.ModeEntryBlocked && to != journal.ModeHaltAll:'),
    ('M18 notice failure reported as an error', MO,
     '\tcase err != nil && errors.Is(err, journal.ErrModeAnnouncementFailed):',
     '\tcase false && errors.Is(err, journal.ErrModeAnnouncementFailed):'),
    ('M19 notice row state not read', MO,
     '\t\t\t\tresult.NoticePending = true\n',
     ''),
    ('M20 invalid mapped to 500', MT,
     '\t\t\t\twriteRPCError(w, http.StatusBadRequest, "invalid", err.Error())',
     '\t\t\t\twriteRPCError(w, http.StatusInternalServerError, "internal", err.Error())'),
    ('M21 mode-release not mutating', CE,
     'Annotations:  map[string]string{"source": "local", "mutating": "true"},',
     'Annotations:  map[string]string{"source": "local"},'),
    ('M22 ack not mutating', CA,
     'Annotations:  map[string]string{"source": "local", "mutating": "true"},',
     'Annotations:  map[string]string{"source": "local"},'),
    ('M23 remaining reasons not reported', MO,
     '\tresult.EntryBlocks = entryBlockReasons(o.gate)',
     '\tresult.EntryBlocks = nil'),
]
MODE_TESTS = [['go', 'test', '-count=1', '-run', 'TestA092|Mode|Transition|Escalat', './internal/journal'], ['go', 'test', '-count=1', '-run', 'TestA092|A124|Mode', './internal/execgw'], ['go', 'test', '-count=1', '-run', 'TestA092|TestTheModeProjector|TestTheLedgerModeRow|Relatch|Sibling', './internal/app/engine'], ['go', 'test', '-count=1', '-run', 'TestA092|TestTheAlertCommands|TestMutating', './cmd/tossctl']]

# 25.10 — 착지 단위 ⑤ C8 일반 등급 이관 · B#7 · A#2 · 구조 핀 반증. `--set 25.10`.
U5 = {
    'RO': 'internal/obs/record_only.go',
    'NR': 'internal/obs/normal_relay.go',
    'EW': 'internal/app/engine/exitwiring.go',
    'AX': 'internal/app/engine/auxiliary.go',
    'NW': 'internal/app/engine/normal_relay_wiring.go',
    'EN': 'cmd/tossctl/engine.go',
    'RT': 'internal/execgw/retry.go',
    'EL': 'internal/app/engine/exitloop.go',
    'AD': 'internal/app/engine/alertdelivery.go',
    'RP': 'internal/execgw/replay.go',
    'GW': 'internal/app/engine/gateway.go',
    'OM': 'internal/journal/operating_mode.go',
    'RG': 'internal/execgw/riskguardian.go',
    'FW': 'internal/app/engine/exitwiring.go',
}
NORMAL_MUTANTS = [
    ('N01 no relay falls back to a sync publish', 'internal/obs/record_only.go',
     '\t\t\tn.logNormalDrop(e, "no normal-grade relay is wired")\n\t\t\treturn nil',
     '\t\t\tn.publishBestEffort(ctx, e, severity)\n\t\t\treturn nil'),
    ('N02 hand-off blocks when full', 'internal/obs/normal_relay.go',
     '\tselect {\n\tcase r.ch <- e:\n\tdefault:\n\t\tr.n.logNormalDrop(e, "the normal-grade relay buffer is full")\n\t}',
     '\tr.ch <- e'),
    ('N03 a full drop is silent', 'internal/obs/normal_relay.go',
     '\t\tr.n.logNormalDrop(e, "the normal-grade relay buffer is full")\n',
     ''),
    ('N04 shutdown drain is silent', 'internal/obs/normal_relay.go',
     '\t\t\tr.n.logNormalDrop(e, "the engine stopped before this normal-grade alert was sent")\n',
     '\t\t\t_ = e\n'),
    ('N05 cancellation reported as nil', 'internal/obs/normal_relay.go',
     '\t\tdefault:\n\t\t\treturn ctx.Err()',
     '\t\tdefault:\n\t\t\treturn nil'),
    ('N06 hand-off skipped silently', 'internal/obs/record_only.go',
     '\t\tr.Relay.Offer(e)\n',
     ''),
    ('N07 exit observer gets no relay', 'internal/app/engine/exitwiring.go',
     'recordOnly := obs.RecordOnly{N: c.Notifier, Relay: c.NormalAlertRelay()}',
     'recordOnly := obs.RecordOnly{N: c.Notifier}'),
    ('N08 runtime does not start the relay', 'cmd/tossctl/engine.go',
     'Auxiliary: []engine.AuxiliaryExecutor{alertDelivery, ectx.NormalAlertRelayExecutor()},',
     'Auxiliary: []engine.AuxiliaryExecutor{alertDelivery},'),
    ('N09 stop event ignored', 'internal/app/engine/auxiliary.go',
     '\tevent := aux.StopEvent\n',
     '\tevent := obs.EventType("")\n'),
    ('N10 relay borrows the delivery event', 'internal/app/engine/normal_relay_wiring.go',
     '\t\tStopEvent: obs.EventNormalAlertRelayStopped,\n',
     ''),
    ('N11 caller alert path respected', 'internal/app/engine/exitwiring.go',
     '\tif c.Notifier != nil {\n\t\topts.Alerts = recordOnly\n\t\topts.Announcer = recordOnly\n\t}',
     '\tif opts.Alerts == nil && c.Notifier != nil {\n\t\topts.Alerts = recordOnly\n\t}\n\tif opts.Announcer == nil && c.Notifier != nil {\n\t\topts.Announcer = recordOnly\n\t}'),
    ('N12 credential notice failure misreported', 'internal/execgw/retry.go',
     '\t\tif errors.Is(err, journal.ErrModeAnnouncementFailed) {\n\t\t\t// 전이는 커밋됐고',
     '\t\tif false {\n\t\t\t// 전이는 커밋됐고'),
    ('N13 outage notice failure not escalated', 'internal/app/engine/exitloop.go',
     '\t\tif !errors.Is(err, journal.ErrModeAnnouncementFailed) {',
     '\t\tif true {'),
    ('N14 takeover logged as held', 'internal/app/engine/alertdelivery.go',
     '\t\td.logf(obs.EventAlertClaimStolen, nil, "an expired alert lease was taken over",',
     '\t\td.logf(obs.EventAlertClaimHeld, nil, "an expired alert lease was taken over",'),
    ('P01 outside recorder releases before inserting', 'internal/execgw/replay.go',
     '\t\tg.entry.Block(ReasonUnresolvedInDoubt, fmt.Sprintf(',
     '\t\tg.entry.Clear(ReasonUnresolvedInDoubt)\n\t\tg.entry.Block(ReasonUnresolvedInDoubt, fmt.Sprintf('),
    ('P02 notifier gets another gate', 'internal/app/engine/gateway.go',
     '\tnotifier := newNotifier(in.journal, entry, in.accountRef, in.logger, in.publisher, in.clock)',
     '\tnotifier := newNotifier(in.journal, execgw.NewEntryGate(in.clock, nil), in.accountRef, in.logger, in.publisher, in.clock)'),
    ('P03 projection in a goroutine', 'internal/journal/operating_mode.go',
     '\t\tp.ProjectOperatingMode(record)',
     '\t\tgo p.ProjectOperatingMode(record)'),
    ('P04 issuer reaches the escalation', 'internal/execgw/riskguardian.go',
     'func (g *RiskGuardian) IssueReduction(ctx context.Context, req ReductionIssuance) (Issued, error) {\n\tintent, err := g.scopedIntent(req.Intent)',
     'func (g *RiskGuardian) IssueReduction(ctx context.Context, req ReductionIssuance) (Issued, error) {\n\tif false {\n\t\t_ = g.escalateFor(ctx, risk.Decision{})\n\t}\n\tintent, err := g.scopedIntent(req.Intent)'),
    ('P05 floor reads outside the retrier', 'internal/app/engine/exitwiring.go',
     '\tif err := f.retrier.Query(ctx, execgw.QueryHoldings, func(ctx context.Context) error {\n\t\tq, err := f.official.SellableQuantity(ctx, symbol)',
     '\tif _, err := f.official.SellableQuantity(ctx, symbol); err != nil {\n\t\t_ = err\n\t}\n\tif err := f.retrier.Query(ctx, execgw.QueryHoldings, func(ctx context.Context) error {\n\t\tq, err := f.official.SellableQuantity(ctx, symbol)'),
    ('P06 a production Flush caller', 'internal/obs/record_only.go',
     '// OperatingModeEventKey 는',
     'func a092MutantFlush(n *Notifier) { _, _, _ = n.Flush(context.Background()) }\n\n// OperatingModeEventKey 는'),
    ("N15 normal grade published synchronously (C8 reverted — capped RED)", 'internal/obs/record_only.go',
     '\t\tr.Relay.Offer(e)\n',
     '\t\tn.publishBestEffort(ctx, e, severity)\n'),
]
NORMAL_TESTS = [['go', 'test', '-count=1', '-run', 'TestA092', './internal/obs'], ['go', 'test', '-count=1', '-run', 'TestA092|Outage|Auxiliar|A098', './internal/app/engine'], ['go', 'test', '-count=1', '-run', 'TestA092|Escalat|Credential', './internal/execgw'], ['go', 'test', '-count=1', '-run', 'TestProductionRuntime|TestTheAlertDeliverer|TestA092', './cmd/tossctl']]

# 25.8 — 26라운드 수리(codex P0 · P1 · P2) 반증. `--set r26`.
R26_MUTANTS = [
    ('X01 release NotFound read as preemption again', 'internal/app/engine/alertdelivery.go',
     '\tif releaseOK && released.Outcome != journal.SettleApplied && released.Outcome != journal.SettleAlreadySettled &&',
     '\tif false && releaseOK && released.Outcome != journal.SettleApplied && released.Outcome != journal.SettleAlreadySettled &&'),
    ('X02 release site latches unconditionally', 'internal/app/engine/alertdelivery.go',
     '\t\td.judge(ctx, id, d.readEpoch(id), false, alertLatchUnaccounted)\n\t}\n\tif err != nil {',
     '\t\td.Gate.Block(execgw.ReasonAlertUndelivered, alertLatchUnaccounted)\n\t}\n\tif err != nil {'),
    ('X03 release site escalates', 'internal/app/engine/alertdelivery.go',
     '\t\td.judge(ctx, id, d.readEpoch(id), false, alertLatchUnaccounted)\n\t}\n\tif err != nil {',
     '\t\td.judge(ctx, id, d.readEpoch(id), true, alertLatchUnaccounted)\n\t}\n\tif err != nil {'),
    ('X04 preemption not recorded', 'internal/app/engine/alertdelivery.go',
     '\t\td.logf(obs.EventAlertClaimLost, nil, "a failed attempt was preempted before it could be recorded",\n\t\t\t"alert_id", id, "outcome", res.Outcome.String())\n',
     ''),
    ('X05 release site returns before the limit verdict', 'internal/app/engine/alertdelivery.go',
     '\t\td.judge(ctx, id, d.readEpoch(id), false, alertLatchUnaccounted)\n\t}\n\tif err != nil {',
     '\t\td.judge(ctx, id, d.readEpoch(id), false, alertLatchUnaccounted)\n\t\treturn\n\t}\n\tif err != nil {'),
    ('X06 preemption judgement widened to not-found', 'internal/obs/settle_outcome.go',
     '\treturn outcome == journal.SettleAlreadySettled || outcome == journal.SettleLeaseLost',
     '\treturn outcome == journal.SettleAlreadySettled || outcome == journal.SettleLeaseLost || outcome == journal.SettleNotFound'),
    ('X07 hand-off after stop silently queued', 'internal/obs/normal_relay.go',
     '\tif r.stopped {\n\t\tr.n.logNormalDrop(e, "the normal-grade relay has stopped")\n\t\treturn\n\t}',
     '\tif false {\n\t}'),
    ('X08 in-flight alert at panic not recorded', 'internal/obs/normal_relay.go',
     '\t\tif inFlight != nil {\n\t\t\tr.n.logNormalDrop(*inFlight, "the normal-grade relay stopped while sending this alert")\n\t\t}\n',
     '\t\t_ = inFlight\n'),
    ('X09 publish failure not recorded', 'internal/obs/normal_relay.go',
     '\t\tr.n.logNormalDrop(e, "publishing the normal-grade alert failed: "+err.Error())',
     '\t\t_ = err'),
    ('X10 no publisher silent', 'internal/obs/normal_relay.go',
     '\t\tr.n.logNormalDrop(e, "no notification publisher is configured")\n',
     ''),
    ('X11 re-read failure returns an error again', 'internal/app/engine/modeops.go',
     '\t\tresult.ReReadError = "re-reading the operating mode after the release failed: " + err.Error()\n\t\treturn result, nil',
     '\t\treturn result, fmt.Errorf("re-read: %w", err)'),
]
R26_TESTS = [['go', 'test', '-count=1', '-run', 'TestA092', './internal/obs'], ['go', 'test', '-count=1', '-run', 'TestA092|A124|A098', './internal/app/engine']]

# 25.8 — 26라운드 2차 수리(Manager 판정 2026-09-30) 반증. `--set r26b`.
R26B_MUTANTS = [
    ('Z01 R1 release preemption not recorded', 'internal/app/engine/alertdelivery.go',
     '\t\td.logf(obs.EventAlertClaimLost, nil, "a failed attempt\'s lease was preempted before it could be handed back",\n\t\t\t"alert_id", id, "outcome", released.Outcome.String())\n',
     '\t\t_ = released\n'),
    ('Z02 R3 unknown outcome classified by ClaimedBy again', 'internal/obs/notifier.go',
     '\tcase res.Outcome != journal.SettleLeaseLost:',
     '\tcase false:'),
    ('Z03 A#1 notice attached to the request ctx', 'internal/app/engine/modeops.go',
     '\t\tannouncer = detachedAnnouncer{inner: o.announcer}',
     '\t\tannouncer = o.announcer'),
    ('Z04 A#1 re-read attached to the request ctx', 'internal/app/engine/modeops.go',
     '\tafter := context.WithoutCancel(ctx)',
     '\tafter := ctx'),
    ('Z05 R2 notice-list failure reported as mode re-read', 'internal/app/engine/modeops.go',
     '\t\t\tresult.NoticeReadError = modeReleaseNoticeReadFailed',
     '\t\t\tresult.ReReadError = modeReleaseNoticeReadFailed'),
    ('Z06 B#2 NotifyError carries raw text', 'internal/app/engine/modeops.go',
     '\t\tresult.NotifyError = modeReleaseNoticeFailed',
     '\t\tresult.NotifyError = err.Error()'),
    ('Z07 B#2 re-read error carries raw text', 'internal/app/engine/modeops.go',
     '\t\tresult.ReReadError = modeReleaseReReadFailed',
     '\t\tresult.ReReadError = err.Error()'),
    ('Z08 B#2 500 body carries raw text', 'internal/app/engine/mode_control_transport_unix.go',
     'writeRPCError(w, http.StatusInternalServerError, "internal", modeReleaseInternalFailure)',
     'writeRPCError(w, http.StatusInternalServerError, "internal", err.Error())'),
    ('Z09 B#2 gate description carries raw text', 'internal/obs/record_only.go',
     '"a critical %s alert could not be recorded in the outbox (details are in the engine log)", e.Type))',
     '"a critical %s alert could not be recorded in the outbox: %v", e.Type, err))'),
    ('Z10 B#1 Notify logs fields', 'internal/obs/record_only.go',
     '\tn.logEvent(withoutFields(e), severity)',
     '\tn.logEvent(e, severity)'),
    ('Z11 B#1 AnnounceOperatingMode logs fields', 'internal/obs/record_only.go',
     '\tn.logEvent(withoutFields(e), SeverityOf(e.Type))',
     '\tn.logEvent(e, SeverityOf(e.Type))'),
    ('Z12 A#4 drop line shadows its event key', 'internal/obs/normal_relay.go',
     '\tn.Log.Warn(EventNormalAlertDropped,\n\t\tFieldTriggerEvent,',
     '\tn.Log.Warn(EventNormalAlertDropped,\n\t\tFieldEvent,'),
    ('Z13 A#4 best-effort line shadows its severity', 'internal/obs/notifier.go',
     '\t\t\t"trigger_severity", string(severity),',
     '\t\t\tFieldSeverity, string(severity),'),
    ('Z14 C#2 takeover line demoted to INFO', 'internal/obs/notifier.go',
     '\tn.Log.Warn(EventAlertClaimStolen,',
     '\tn.Log.Event(EventAlertClaimStolen,'),
    ('Z15 C#3 mode endpoint without token check', 'internal/app/engine/mode_control_transport_unix.go',
     'mux.HandleFunc(ModeControlReleasePath, alertControlAuth(token, func(',
     'mux.HandleFunc(ModeControlReleasePath, func(h http.HandlerFunc) http.HandlerFunc { return h }(func('),
    ('Z16 C#3 mode endpoint accepts GET', 'internal/app/engine/mode_control_transport_unix.go',
     '\t\tif r.Method != http.MethodPost {',
     '\t\tif false {'),
    ('Z17 C#3 unknown fields accepted', 'internal/app/engine/mode_control_transport_unix.go',
     '\tdecoder.DisallowUnknownFields()\n',
     ''),
    ('Z18 B#7 sustained tightening reported as not persisted', 'internal/app/engine/runtime.go',
     '\tif err != nil && errors.Is(err, journal.ErrModeAnnouncementFailed) {',
     '\tif false && errors.Is(err, journal.ErrModeAnnouncementFailed) {'),
    ('Z19 B#7 daily-loss tightening reported as not persisted', 'internal/execgw/riskguardian.go',
     '\t\tif errors.Is(err, journal.ErrModeAnnouncementFailed) {\n\t\t\t// 전이는 커밋됐고 통지 기록만 실패함 — 재시작이',
     '\t\tif false && errors.Is(err, journal.ErrModeAnnouncementFailed) {\n\t\t\t// 전이는 커밋됐고 통지 기록만 실패함 — 재시작이'),
    ('Z20 R2 CLI calls not-pending delivered', 'cmd/tossctl/engine_mode_release.go',
     '\t\treturn "기록됨(대기 목록에 없음 — 전달 뒤 정산됐거나 승인됨)"',
     '\t\treturn "기록됨(이미 전달 처리됨)"'),
    ('Z21 R2 CLI guesses the notice state after a mode re-read failure', 'cmd/tossctl/engine_mode_release.go',
     '\tcase r.ReReadError != "":\n\t\treturn "기록됨(전달 상태는 재조회 실패로 확인하지 못함)"\n',
     ''),
    ('Z22 R2 CLI hides the read mode after a notice-list failure', 'cmd/tossctl/engine_mode_release.go',
     '\tif r.ReReadError != "" {\n\t\t_, err := fmt.Fprintf(',
     '\tif r.ReReadError != "" || r.NoticeReadError != "" {\n\t\t_, err := fmt.Fprintf('),
    ('Z23 R2 CLI drops the notice-list failure beside a record failure', 'cmd/tossctl/engine_mode_release.go',
     '\t\tif r.NotifyError != "" && r.NoticeReadError != "" {', '\t\tif false {'),
    ('Z24 N1 record-failure log carries the account', 'internal/obs/record_only.go',
     'MaskAccount(err, n.AccountRef), FieldTriggerEvent', 'err, FieldTriggerEvent'),
    ('Z25 mode-release failure log carries the account', 'internal/app/engine/modeops.go',
     'obs.MaskAccount(err, o.accountRef), obs.FieldDetail', 'err, obs.FieldDetail'),
    ('Z26 C-N2 bearer check compares only the length', 'internal/app/engine/alert_control_transport_unix.go',
     'subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {', 'subtle.ConstantTimeCompare([]byte(provided), []byte(provided)) != 1 {'),
]
R26B_TESTS = [['go', 'test', '-count=1', '-run', 'TestA092', './internal/obs'], ['go', 'test', '-count=1', '-run', 'TestA092|A124|A098', './internal/app/engine'], ['go', 'test', '-count=1', '-run', 'TestA092', './internal/execgw'], ['go', 'test', '-count=1', '-run', 'TestA092', './cmd/tossctl']]

TESTS = [
    ["go", "test", "-count=1", "-run", "TestA092|TestEnqueueAlert|TestClaim", "./internal/journal"],
    ["go", "test", "-count=1", "-run", "TestA092|TestA096|TestA097|Mode|Transition|Announc", "./internal/obs"],
    ["go", "test", "-count=1", "-run", "TestA092|TestProductionGuardianUsesConfiguredUSDLimitsAndReachesExitObserver", "./internal/app/engine"],
    ["go", "test", "-count=1", "-run", "TestA092|TestTheExitObserverDefersToFillDetection", "./cmd/tossctl"],
]


def run_tests(copy: Path, env: dict) -> tuple[str, str]:
    failed, notes, build = [], [], False
    for command in TESTS:
        result = subprocess.run(command, cwd=copy, env=env, capture_output=True, text=True)
        if result.returncode != 0:
            out = result.stdout + result.stderr
            if "[build failed]" in out or "[setup failed]" in out:
                build = True
                notes.append(next((l for l in out.splitlines() if ".go:" in l), "build failed")[:200])
                continue
            names = [line.strip()[len("--- FAIL: "):].split(" ")[0] for line in out.splitlines() if line.strip().startswith("--- FAIL: ")]
            leaves = [n for n in names if not any(o != n and o.startswith(n + "/") for o in names)]
            failed.extend(leaves)
            if not names:
                notes.append(out.strip().splitlines()[-1][:160] if out.strip() else "no output")
    why = f"{len(failed)} failing: " + ", ".join(failed[:8]) + (" …" if len(failed) > 8 else "") + (" | " + " | ".join(notes) if notes else "")
    if build:
        return "BUILD-FAIL", why
    if failed or notes:
        return "RED", why
    return "GREEN", ""


def main() -> None:
    args = sys.argv[1:]
    only = None
    if "--only" in args:
        i = args.index("--only")
        only = re.compile(args[i + 1])
        args = args[:i] + args[i + 2:]
    global MUTANTS, TESTS
    if "--set" in args:
        i = args.index("--set")
        if args[i + 1] == "25.6":
            MUTANTS, TESTS = RELAX_MUTANTS, RELAX_TESTS
        elif args[i + 1] == "25.7":
            MUTANTS, TESTS = LOCK_MUTANTS, LOCK_TESTS
        elif args[i + 1] == "25.9":
            MUTANTS, TESTS = MODE_MUTANTS, MODE_TESTS
        elif args[i + 1] == "25.10":
            MUTANTS, TESTS = NORMAL_MUTANTS, NORMAL_TESTS
        elif args[i + 1] == "r26":
            MUTANTS, TESTS = R26_MUTANTS, R26_TESTS
        elif args[i + 1] == "r26b":
            MUTANTS, TESTS = R26B_MUTANTS, R26B_TESTS
        args = args[:i] + args[i + 2:]
    scratch, own = Path(args[0]), args[1:]
    copy = scratch / f"mut-a092-u2-{os.getpid()}"
    copy.mkdir(parents=True)
    tracked = subprocess.run(["git", "ls-files", "--", "go.mod", "go.sum", "internal", "cmd", "tools"], cwd=ROOT,
                             capture_output=True, text=True, check=True).stdout.split()
    for rel in tracked + own:
        source = ROOT / rel
        if source.exists():
            (copy / rel).parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(source, copy / rel)
    env = dict(os.environ, GOFLAGS="-trimpath")
    head = subprocess.run(["git", "rev-parse", "HEAD"], cwd=ROOT, capture_output=True, text=True, check=True).stdout.strip()
    dirty = subprocess.run(["git", "diff", "--name-only", "HEAD", "--", "go.mod", "go.sum", "internal", "cmd", "tools"],
                           cwd=ROOT, capture_output=True, text=True, check=True).stdout.split()
    ledger = open(copy / "ledger.tsv", "w", encoding="utf-8")
    ledger.write(f"TREE\tHEAD {head}\tuncommitted tracked: {','.join(dirty) or 'none'}\town untracked: {','.join(own) or 'none'}\n")
    verdict, why = run_tests(copy, env)
    ledger.write(f"CONTROL\t{verdict}\t{why}\n")
    ledger.flush()
    if verdict != "GREEN":
        print("control not GREEN — stop:", verdict, why)
        sys.exit(2)
    for ident, rel, old, new in MUTANTS:
        if only and not only.search(ident):
            continue
        target = copy / rel
        pristine = target.read_text(encoding="utf-8")
        if pristine.count(old) != 1:
            ledger.write(f"{ident}\tNOT-APPLIED\told occurs {pristine.count(old)} times\n")
            ledger.flush()
            continue
        target.write_text(pristine.replace(old, new, 1), encoding="utf-8")
        verdict, why = run_tests(copy, env)
        label = {"RED": "CAUGHT", "GREEN": "SURVIVED"}.get(verdict, verdict)
        ledger.write(f"{ident}\t{label}\t{why}\n")
        ledger.flush()
        target.write_text(pristine, encoding="utf-8")
    ledger.close()
    print((copy / "ledger.tsv").read_text(encoding="utf-8"))


if __name__ == "__main__":
    main()
