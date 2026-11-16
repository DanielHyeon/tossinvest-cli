#!/usr/bin/env python3
"""a189 범위 확장 2(낡은 시험 가드 3건) 사본 변이 하네스.

앞 하네스(analysis/mutation/mutation.py)의 규칙을 그대로 따름:
- HEAD 를 git archive 로 $HOME/tossos-work 아래 pid 붙은 사본에 풀고 사본에서만 변이함(작업 트리 무변경).
- 작업 트리의 변이 대상 파일이 HEAD 와 같은지 먼저 단언(커밋 뒤 돌릴 것).
- 무변이 대조군(판본별)이 GREEN 이 아니면 멈춤.
- 변이마다: 대상 문자열이 정확히 1회인지 단언 → 적용 → 변이 문자열이 사본에 들어갔는지 단언 → 패키지를 시험 파일까지
  컴파일(-run '^$')해 빌드 실패를 CAUGHT 로 세지 않음 → 시험 → 원복(새 파일은 삭제) 후 SHA-256 단언.
- 판정은 「기대」 목록으로만 냄: (패키지, 시험 이름, FAIL|PASS). FAIL 기대는 `--- FAIL: <이름> (` 가, PASS 기대는
  `--- PASS: <이름> (` 가 출력에 있어야 함. 하나라도 어긋나면 MISMATCH.
- HOME·XDG_* 격리, 전용 GOCACHE(끝에 지움). -trimpath 는 쓰지 않음(시험이 상대 경로로 소스를 읽음).
- 「판본」: 같은 변이를 시험 쪽 판본을 바꿔 다시 돎 — S(#1 문자열만 고친 판본), OLD(#3 앞 판본 admitSites). 좁힘·선택이
  무엇을 얻고 잃는지 같은 변이로 비교하려는 것임. 판본 자체도 대조군을 먼저 돎(OLD 판본 대조군은 base 실패가 기대값).
사용: python3 -I mutation2.py [ID…]
"""
from __future__ import annotations

import hashlib
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
from collections import Counter
from pathlib import Path


class Fail(Exception):
    pass


def check(cond: bool, msg: str) -> None:
    if not cond:
        raise Fail(msg)


MOD = "github.com/JungHoonGhae/tossinvest-cli/"
CONSOLE = "cmd/tossctl/console.go"
CONSOLE_TEST = "cmd/tossctl/console_test.go"
EXITLOOP = "internal/app/engine/exitloop.go"
HANDOFF = "internal/app/engine/strategy_dispatch_handoff.go"
GUARD = "internal/app/engine/strategy_dispatch_handoff_guard_test.go"
NEWFILE = "internal/app/engine/a189_mutant.go"
WRAP = "internal/a189wrap/wrap.go"

T1 = "TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient"
T2 = "TestA163TheJudgementFunctionsRenderExactlyAsBeforeTheChange"
T3 = "TestExactlyOneProductionSiteAdmitsIntoTheSeam"
T3_SPELL = "TestAdmitCensusSeesEverySpellingItClaims"
T3_OUTSIDE = "TestA189AdmitCensusIgnoresSameNamedMembersOutsideTheSeam"
T3_UNALIASED = "TestA189AdmitCensusCountsTheUnaliasedSeamName"
T3_DOT = "TestA189AdmitCensusIsConservativeInADotImportedFile"
T_ONLY_ENGINE = "TestOnlyTheEngineImportsThisSeam"
T3_LITERAL = "TestA189AdmitCensusDecodesTheImportPathLiteral"
# T3_UNSAFE·T3_NOLINK 시험은 Manager 결정 2(2026-11-16)로 삭제됨 — a9~a14 는 「잡지 못함(기존 맹점, 범위 밖)」 으로 재분류
LINKPKG = "internal/a189link/link.go"
RELAY = "internal/app/engine/testdata/a189relay/relay.go"
SEAMFILE = "internal/strategyhandoff/handoff.go"
PUSHPKG = "internal/a189push/push.go"
SYMDIR = "internal/a189sym"
HIDDEN = "internal/.a189hidden/h.go"

PKG1 = ("./cmd/tossctl/", f"^{T1}$")
PKG2 = ("./internal/app/engine/", f"^{T2}$")
PKG3 = ("./internal/app/engine/", f"^({T3}|{T3_SPELL}|{T3_OUTSIDE}|{T3_UNALIASED}|{T3_DOT}|{T3_LITERAL})$")
# Manager 결정 1(2026-11-16): 두 그림자 도구의 import 시험이 official·accountbudget·accountbudgetruntime 의 별칭·dot·blank 를 금지함
A128_DIR, A140_DIR = "tools/a128-jev-shadow-probe", "tools/a140-breakout-signal-shadow"
T_IMP = "TestImportsStayInsideTheAllowlist"
T_A128_NAMES, T_A140_NAMES = "TestNoMutatingOrAccountSurfaceIsNamed", "TestNoMutatingAccountOrWriteSurfaceIsNamed"
PKG_A128 = ("./" + A128_DIR + "/", f"^({T_IMP}|{T_A128_NAMES})$")
PKG_A140 = ("./" + A140_DIR + "/", f"^({T_IMP}|{T_A140_NAMES})$")

PKG_SEAM = ("./internal/strategyhandoff/", f"^{T_ONLY_ENGINE}$")

RESOLVE = "\tbroker, accountRef, err := l.shared.resolve(ctx)\n"
SEAM_IMPORT = f'"{MOD}internal/strategyhandoff"'

# ---- 판본(시험 쪽 대체) --------------------------------------------------------------------------------------------

# S: #1 의 AST 단언 블록을 「철자만 고친」 문자열 검사로 바꾼 판본(선택 근거 비교용).
S_START = "\t// 공유 resolver 사용은 (*lazyOrders).Orders 본문 안에서만 셈"
S_END = "read per session\", found, resolves, feeds, rebinds, asserted)\n\t}\n"
S_BODY = ('\tif !strings.Contains(src, "l.shared.resolve(ctx)") {\n'
          '\t\tt.Error("the orders seam no longer reads through the console\'s shared client; a resolver " +\n'
          '\t\t\t"of its own is a second /api/v1/accounts read per session")\n\t}\n')

# OLD: #3 의 admitSites 를 앞 판본(이름 기반) 본문으로 되돌린 판본.
OLD_ADMIT = '''func admitSites(file *ast.File, doors map[string]bool) []string {
	sites := make([]string, 0, 1)
	for _, decl := range file.Decls {
		ast.Inspect(decl, func(node ast.Node) bool {
			if ident, ok := node.(*ast.Ident); ok && doors[ident.Name] {
				sites = append(sites, declName(decl)+":"+ident.Name)
			}
			return true
		})
	}
	return sites
}
'''


PRE1_A128 = """			if ownNameOnly[path] && spec.Name != nil {
				t.Errorf("import %q is renamed to %q; official/accountbudget/accountbudgetruntime must be imported under their own name so the name pins see every use", path, spec.Name.Name)
			}
"""
PRE1_A140 = """			if strings.HasPrefix(path, module) && ownNameOnly[strings.TrimPrefix(path, module)] && spec.Name != nil {
				t.Errorf("import %q is renamed to %q; official/accountbudget/accountbudgetruntime must be imported under their own name so the name pins see every use", path, spec.Name.Name)
			}
"""


def variant_edits(name: str, text_of) -> list:
    if name == "S":
        src = text_of(CONSOLE_TEST)
        i, j = src.find(S_START), src.find(S_END)
        check(i > 0 and j > i, "S 판본 표식을 못 찾음")
        return [(CONSOLE_TEST, "once", src[i:j + len(S_END)], S_BODY)]
    if name == "PRE1":
        # Manager 결정 1 이전 판본(별칭 금지 갈래 없음) — 같은 변이를 결정 전 가드가 못 잡았음을 잼
        return [(A128_DIR + "/static_test.go", "once", PRE1_A128, "\t\t\t_ = ownNameOnly // PRE1: 금지 갈래만 뺌(미사용 변수 컴파일 오류 방지)\n"),
                (A140_DIR + "/static_test.go", "once", PRE1_A140, "\t\t\t_ = ownNameOnly // PRE1: 금지 갈래만 뺌(미사용 변수 컴파일 오류 방지)\n")]
    if name == "OLD":
        src = text_of(GUARD)
        i = src.find("func admitSites(")
        j = src.find("\n}\n", i)
        check(i > 0 and j > i, "OLD 판본 표식을 못 찾음")
        return [(GUARD, "once", src[i:j + 3], OLD_ADMIT)]
    return []


# ---- 변이 ---------------------------------------------------------------------------------------------------------
# (ID, 판본, 편집 목록[(파일, 종류, 옛, 새)], 기대[(패키지묶음, 시험, FAIL|PASS)], 설명)
# 종류: once(정확히 1회 치환) · new(새 파일 — 옛 값은 무시)

def newfile(body: str):
    return (NEWFILE, "new", "", "package engine\n\n" + body)


def seam_file(body: str, imp: str = SEAM_IMPORT):
    return newfile(f"import {imp}\n\n{body}")


O1 = [(CONSOLE, "once", RESOLVE, RESOLVE + "\t_, _, _ = l.shared.resolve(ctx)\n")]
O2 = [(CONSOLE, "once", RESOLVE, "\tbroker, accountRef, err := verifyBrokerFactory(ctx, l.shared.root)\n")]
O3 = [(CONSOLE, "once", RESOLVE, "\tbroker, accountRef, err := newConsoleBroker(l.shared.root).resolve(ctx)\n")]
O4 = [(CONSOLE, "once", RESOLVE,
       "\tif l.shared.client == nil {\n\t\tb, ref, ferr := verifyBrokerFactory(ctx, l.shared.root)\n"
       "\t\tif ferr != nil {\n\t\t\treturn console.OrdersReading{}, ferr\n\t\t}\n"
       "\t\tl.shared.client, l.shared.accountRef = b, strings.TrimSpace(ref)\n\t}\n"
       "\tbroker, accountRef, err := l.shared.client, l.shared.accountRef, error(nil)\n")]
O5 = [(CONSOLE, "once", RESOLVE, "\tresolve := l.shared.resolve\n\tbroker, accountRef, err := resolve(ctx)\n")]
# 2차 리뷰(2026-11-16): 참조 수만 세던 판본을 뚫던 모양들
O7 = [(CONSOLE, "once", RESOLVE, O4[0][3] + "\t_ = l.shared.resolve\n")]
O8 = [(CONSOLE, "once", RESOLVE, O4[0][3].replace("\tbroker, accountRef, err :=", "\t_, _, _ = l.shared.resolve(ctx)\n\tbroker, accountRef, err :="))]
O9 = [(CONSOLE, "once", RESOLVE, RESOLVE + "\tbroker = l.shared.client\n")]
O10 = [(CONSOLE, "once", "\treads, ok := broker.(consoleOrdersReader)\n",
        "\tvar reads consoleOrdersReader\n\tok := false\n\t{\n\t\tvar broker verifylive.Broker = l.shared.client\n"
        "\t\treads, ok = broker.(consoleOrdersReader)\n\t}\n")]
ASSERT = "\treads, ok := broker.(consoleOrdersReader)\n"
O11 = [(CONSOLE, "once", ASSERT, "\tvar reads consoleOrdersReader\n\tok := false\n"
        "\tfor _, broker := range []verifylive.Broker{l.shared.client} {\n\t\treads, ok = broker.(consoleOrdersReader)\n\t}\n")]
O12 = [(CONSOLE, "once", ASSERT, "\treads, ok := func(broker verifylive.Broker) (consoleOrdersReader, bool) {\n"
        "\t\tr, ok := broker.(consoleOrdersReader)\n\t\treturn r, ok\n\t}(l.shared.client)\n")]
O13 = [(CONSOLE, "once", ASSERT, ASSERT + "\tif r2, ok2 := verifylive.Broker(l.shared.client).(consoleOrdersReader); ok2 {\n\t\treads = r2\n\t}\n")]
O14 = [(CONSOLE, "once", ASSERT, "\talias := &broker\n\t*alias = verifylive.Broker(l.shared.client)\n" + ASSERT)]
O6 = [(CONSOLE, "once", RESOLVE, RESOLVE + "\t_ = official.New(official.Credentials{}, \"\")\n")]

PRIO_CRIT = "\t\t\t\tclearCtx = accountbudget.WithPriority(ctx, accountbudget.Critical)\n"
R1 = [(EXITLOOP, "once", PRIO_CRIT, PRIO_CRIT.replace("Critical", "Normal"))]
R2 = [(EXITLOOP, "once",
       "\t\t\tclearCtx := accountbudget.WithPriority(ctx, accountbudget.Normal)\n\t\t\tif isProtective(proposal) {\n"
       + PRIO_CRIT + "\t\t\t}\n\t\t\tcleared, err := o.clearTheSymbol(clearCtx, m, snapshot.CancelPendingFirst)\n",
       "\t\t\tcleared, err := o.clearTheSymbol(ctx, m, snapshot.CancelPendingFirst)\n")]

ADMIT_RETURN = "\treturn strategyhandoff.Admit(authority.snapshot.Ready, selected)\n"
FLOW_SEAM = f'(\n\t"{MOD}internal/strategyflow"\n\t{SEAM_IMPORT}\n)'

ONE = lambda name, verdict="FAIL": [(PKG1, name, verdict)]

def tool_alias(dir_: str, imp: str, body: str):
    return (dir_ + "/zz_a189_decision1.go", "new", "", "package main\n\nimport (\n\t\"context\"\n\n\t" + imp + "\n)\n\n" + body)


AB = '"' + MOD + 'internal/accountbudget"'
D1_ALIAS = "func a189Escalate(ctx context.Context) context.Context { return ab.WithPriority(ctx, ab.Critical) }\n"
D1_DOT = "func a189Escalate(ctx context.Context) context.Context { return WithPriority(ctx, Critical) }\n"
D1_BLANK = "func a189Touch(ctx context.Context) context.Context { return ctx }\n"


MUTANTS = [
    # Manager 결정 1 — 둘째 파일의 별칭·dot·blank import(전부 FAIL), PRE1 판본은 별칭·dot 을 못 잡았음
    ("d1a-a128-alias-critical", "", [tool_alias(A128_DIR, "ab " + AB, D1_ALIAS)], [(PKG_A128, T_IMP, "FAIL")], "a128 둘째 파일에서 별칭 import + WithPriority(ctx, ab.Critical)"),
    ("d1a-a128-alias-critical-PRE1", "PRE1", [tool_alias(A128_DIR, "ab " + AB, D1_ALIAS)], [(PKG_A128, T_IMP, "PASS"), (PKG_A128, T_A128_NAMES, "PASS")], "[PRE1 판본] 같은 변이 — 결정 전 가드는 import·이름 시험 둘 다 통과(못 잡음)"),
    ("d1b-a128-dot-critical", "", [tool_alias(A128_DIR, ". " + AB, D1_DOT)], [(PKG_A128, T_IMP, "FAIL")], "a128 둘째 파일에서 dot import + WithPriority(ctx, Critical)"),
    ("d1b-a128-dot-critical-PRE1", "PRE1", [tool_alias(A128_DIR, ". " + AB, D1_DOT)], [(PKG_A128, T_IMP, "PASS"), (PKG_A128, T_A128_NAMES, "PASS")], "[PRE1 판본] 같은 변이 — 못 잡음"),
    ("d1c-a128-blank", "", [tool_alias(A128_DIR, "_ " + AB, D1_BLANK)], [(PKG_A128, T_IMP, "FAIL")], "a128 둘째 파일에서 blank import"),
    ("d1a-a140-alias-critical", "", [tool_alias(A140_DIR, "ab " + AB, D1_ALIAS)], [(PKG_A140, T_IMP, "FAIL")], "a140 둘째 파일에서 별칭 import + WithPriority(ctx, ab.Critical)"),
    ("d1a-a140-alias-critical-PRE1", "PRE1", [tool_alias(A140_DIR, "ab " + AB, D1_ALIAS)], [(PKG_A140, T_IMP, "PASS"), (PKG_A140, T_A140_NAMES, "PASS")], "[PRE1 판본] 같은 변이 — 못 잡음"),
    ("d1b-a140-dot-critical", "", [tool_alias(A140_DIR, ". " + AB, D1_DOT)], [(PKG_A140, T_IMP, "FAIL")], "a140 둘째 파일에서 dot import + WithPriority(ctx, Critical)"),
    ("d1b-a140-dot-critical-PRE1", "PRE1", [tool_alias(A140_DIR, ". " + AB, D1_DOT)], [(PKG_A140, T_IMP, "PASS"), (PKG_A140, T_A140_NAMES, "PASS")], "[PRE1 판본] 같은 변이 — 못 잡음"),
    ("d1c-a140-blank-official", "", [tool_alias(A140_DIR, '_ "' + MOD + 'internal/official"', D1_BLANK)], [(PKG_A140, T_IMP, "FAIL")], "a140 둘째 파일에서 official blank import"),
    # #1 콘솔 orders seam
    ("o1-second-resolve", "", O1, ONE(T1), "Orders 가 공유 resolver 를 두 번 부름(캐시라 빌드는 1회)"),
    ("o1-second-resolve-S", "S", O1, ONE(T1, "PASS"), "[S 판본] 같은 변이 — 철자만 고친 판본은 못 잡음(기대 PASS)"),
    ("o2-own-factory", "", O2, ONE(T1), "Orders 가 공유 resolver 대신 verifyBrokerFactory 를 직접 부름"),
    ("o2-own-factory-S", "S", O2, ONE(T1), "[S 판본] 같은 변이 — 빌드 횟수 절반이 잡음"),
    ("o3-own-broker", "", O3, ONE(T1), "Orders 가 자기 consoleBroker 를 새로 만들어 resolve"),
    ("o4-unlocked-cache", "", O4, ONE(T1), "Orders 가 잠금을 건너뛰고 공유 캐시 필드를 직접 채움(빌드 1회 유지)"),
    ("o4-unlocked-cache-S", "S", O4, ONE(T1, "PASS"), "[S 판본] 같은 변이 — Holdings 의 철자로 통과(기대 PASS)"),
    ("o5-method-value", "", O5, ONE(T1), "resolver 를 메서드 값으로 빼서 한 번 부름(2차: 직접 호출만 받으므로 기대 FAIL)"),
    ("o7-dead-reference", "", O7, ONE(T1), "o4 에 안 쓰는 참조 `_ = l.shared.resolve` 한 줄을 더함(1차 판본을 뚫던 모양)"),
    ("o8-discarded-call", "", O8, ONE(T1), "o4 에 결과를 버리는 직접 호출 `_, _, _ = l.shared.resolve(ctx)` 를 더함"),
    ("o10-var-shadow", "", O10, ONE(T1), "안쪽 블록의 var broker 로 가린 뒤 그 값을 주문 읽기로 단언"),
    ("o11-range-shadow", "", O11, ONE(T1), "range 의 broker 로 가린 값을 단언(2회차 리뷰)"),
    ("o12-funclit-param", "", O12, ONE(T1), "함수 리터럴 인자 broker 로 가린 값을 단언(2회차 리뷰)"),
    ("o13-second-assertion", "", O13, ONE(T1), "둘째 단언으로 reads 를 다른 값으로 바꿈(2회차 리뷰)"),
    ("o14-pointer-write", "", O14, ONE(T1), "&broker 포인터로 broker 를 바꿔 씀(3회차 리뷰)"),
    ("o9-rebind", "", O9, ONE(T1), "공유 resolver 결과를 받은 뒤 broker 를 공유 캐시 필드로 다시 대입"),
    ("o6-official-new", "", O6, ONE(T1), "console.go 가 official.New 로 둘째 클라이언트를 만듦"),
    # #2 a163 렌더 해시
    ("r1-record-priority", "", R1, [(PKG2, T2, "FAIL")], "record 의 보호 청산 우선순위 Critical → Normal"),
    ("r2-record-revert-9c4bdf4de", "", R2, [(PKG2, T2, "FAIL")], "record 를 9c4bdf4de 이전 모양으로 되돌림(옛 해시를 새 고정값이 거절)"),
    # #3 dispatch seam — 진짜 둘째 자리(전부 잡혀야 함)
    ("a1-second-call", "", [seam_file("func a189Second() { _ = strategyhandoff.Admit(true, nil) }\n")],
     [(PKG3, T3, "FAIL")], "새 생산 파일에 strategyhandoff.Admit 둘째 호출"),
    ("a2-alias-call", "", [seam_file("func a189Alias() { _ = sh.Admit(true, nil) }\n", "sh " + SEAM_IMPORT)],
     [(PKG3, T3, "FAIL")], "별칭 import(sh)로 Admit 호출"),
    ("a3-dot-call", "", [seam_file("func a189Dot() { _ = Admit(true, nil) }\n", ". " + SEAM_IMPORT)],
     [(PKG3, T3, "FAIL")], "dot import 로 맨 이름 Admit 호출"),
    ("a4-each-scope", "", [seam_file("func a189Each() { _ = strategyhandoff.AdmitEachOwnerScope(true, nil) }\n")],
     [(PKG3, T3, "FAIL")], "AdmitEachOwnerScope 둘째 호출"),
    ("a5-func-value", "", [seam_file("func a189Value() { f := strategyhandoff.Admit; _ = f(true, nil) }\n")],
     [(PKG3, T3, "FAIL")], "함수 값으로 빼돌려 부름(f := strategyhandoff.Admit; f(…))"),
    ("a6-package-var", "", [seam_file("var a189Mint = strategyhandoff.Admit\n")],
     [(PKG3, T3, "FAIL")], "패키지 수준 함수 값 var a189Mint = strategyhandoff.Admit"),
    ("a7-raw-string-import", "", [seam_file("func a189Raw() { _ = sh.Admit(true, nil) }\n", "sh `" + MOD + "internal/strategyhandoff`")],
     [(PKG3, T3, "FAIL")], "원시 문자열 import 경로(`…`)로 별칭 Admit 호출(2차 리뷰)"),
    ("a8-escaped-import", "", [seam_file("func a189Esc() { _ = sh.Admit(true, nil) }\n", 'sh "' + MOD + 'internal/strategy\\x68andoff"')],
     [(PKG3, T3, "FAIL")], "이스케이프 import 경로(\\x68)로 별칭 Admit 호출(2차 리뷰)"),
    ("a9-linkname", "", [newfile(f'import (\n\t_ "unsafe"\n\n\t"{MOD}internal/strategyflow"\n\t{SEAM_IMPORT}\n)\n\n'
                                 f'//go:linkname a189Mint {MOD}internal/strategyhandoff.Admit\n'
                                 'func a189Mint(bool, []strategyflow.Result) strategyhandoff.Handoff\n\n'
                                 'func a189ViaLink() { _ = a189Mint(true, nil) }\n')],
     [(PKG3, T3, "PASS")], "[잡지 못함 — 기존 맹점, 범위 밖] 엔진 안 //go:linkname 몸체 없는 다른 이름 함수(census PASS)"),
    ("a11-linkname-testdata-relay", "", [(RELAY, "new", "", f'package a189relay\n\nimport _ "unsafe"\n\n'
                                         f'//go:linkname door {MOD}internal/strategyhandoff.AdmitEachOwnerScope\n'
                                         'func door(bool, []struct{}) []struct{}\n\nvar Door = door\n')],
     [(PKG3, T3, "PASS")], "[잡지 못함 — 기존 맹점, 범위 밖] testdata 아래 중계 패키지의 linkname(census PASS)"),
    ("a12-linkname-dispatch-target", "", [(LINKPKG, "new", "", 'package a189link\n\nimport _ "unsafe"\n\n'
                                          f'//go:linkname hand {MOD}internal/app/engine.(*strategyDispatchCycle).dispatch\n'
                                          'func hand()\n\nvar Hand = hand\n')],
     [(PKG3, T3, "PASS")], "[잡지 못함 — 기존 맹점, 범위 밖] 엔진 사설 dispatch 대상 linkname(census PASS)"),
    ("a13-push-linkname", "", [("internal/strategyhandoff/zz_a189push.go", "new", "", 'package strategyhandoff\n\nimport _ "unsafe"\n\n'
                                f'//go:linkname Admit {MOD}internal/a189push.door\n'),
                               (PUSHPKG, "new", "", 'package a189push\n\nfunc door()\n\nvar Door = door\n'),
                               ("internal/a189push/empty.s", "new", "", "")],
     [(PKG3, T3, "PASS")], "[잡지 못함 — 기존 맹점, 범위 밖] 경계 패키지 안 push 형태 linkname(census PASS)"),
    ("a14-symlinked-dir", "", [(".a189hiddenreal/h.go", "new", "", "package a189hidden\n"),
                               (SYMDIR, "symlink", "../.a189hiddenreal", "")],
     [(PKG3, T3, "PASS")], "[잡지 못함 — 범위 밖] 디렉터리 심링크 뒤 패키지(census 는 엔진 파일만 봄 — PASS)"),
    ("a10-linkname-other-package", "", [(LINKPKG, "new", "", f'package a189link\n\nimport (\n\t_ "unsafe"\n\n\t"{MOD}internal/strategyflow"\n)\n\n'
                                       'type look struct{ a, b uintptr }\n\n'
                                       f'//go:linkname door {MOD}internal/strategyhandoff.Admit\n'
                                       'func door(bool, []strategyflow.Result) look\n\nvar Door = door\n')],
     [(PKG3, T3, "PASS")], "[잡지 못함 — 기존 맹점, 범위 밖] 엔진 밖 패키지가 linkname 으로 문을 다시 내보냄(census PASS)"),
    # #3 무해한 같은 이름(안 잡혀야 함)
    ("b1-benign-method-call", "", [newfile('import "context"\n\nfunc a189Benign(ctx context.Context, b *exitLowerLimitBoard) error {\n'
                                           '\treturn b.admission.Admit(ctx)\n}\n')],
     [(PKG3, T3, "PASS")], "다른 타입의 같은 이름 메서드 호출 b.admission.Admit(ctx)"),
    ("b1-benign-method-call-OLD", "OLD", [newfile('import "context"\n\nfunc a189Benign(ctx context.Context, b *exitLowerLimitBoard) error {\n'
                                                  '\treturn b.admission.Admit(ctx)\n}\n')],
     [(PKG3, T3, "EXTRA:a189Benign:Admit")], "[OLD 판본] 같은 변이 — 앞 판본은 이것도 경계 자리로 셈(대조군 목록 + a189Benign:Admit)"),
    ("b2-benign-interface", "", [newfile("type a189Gate interface{ Admit() }\n")],
     [(PKG3, T3, "PASS")], "엔진 사설 인터페이스의 Admit 메서드 선언 하나 더"),
    # #3 좁힘이 잃는 모양(잃음을 세려고 둠 — 새 판본 PASS, OLD 판본 FAIL 이면 「잃은 검출」)
    ("l1-door-named-field", "", [(HANDOFF, "once", ADMIT_RETURN,
                                  "\ta189Doors.Admit = strategyhandoff.Admit\n\treturn a189Doors.Admit(authority.snapshot.Ready, selected)\n"),
                                 newfile(f"import {FLOW_SEAM}\n\nvar a189Doors struct {{\n\tAdmit func(bool, []strategyflow.Result) strategyhandoff.Handoff\n}}\n\n"
                                         "func a189Elsewhere() strategyhandoff.Handoff { return a189Doors.Admit(true, nil) }\n")],
     [(PKG3, T3, "PASS")], "허용 자리 안에서 문을 문 이름 필드로 빼돌리고 다른 자리에서 부름(잃음 — 기대 PASS)"),
    ("l1-door-named-field-OLD", "OLD", [(HANDOFF, "once", ADMIT_RETURN,
                                         "\ta189Doors.Admit = strategyhandoff.Admit\n\treturn a189Doors.Admit(authority.snapshot.Ready, selected)\n"),
                                        newfile(f"import {FLOW_SEAM}\n\nvar a189Doors struct {{\n\tAdmit func(bool, []strategyflow.Result) strategyhandoff.Handoff\n}}\n\n"
                                                "func a189Elsewhere() strategyhandoff.Handoff { return a189Doors.Admit(true, nil) }\n")],
     [(PKG3, T3, "EXTRA:a189Elsewhere:Admit")], "[OLD 판본] 같은 변이 — 앞 판본은 이름이 같아 우연히 잡음(대조군 목록 + a189Elsewhere:Admit)"),
    ("l2-other-named-field", "", [(HANDOFF, "once", ADMIT_RETURN,
                                   "\ta189Doors.mint = strategyhandoff.Admit\n\treturn a189Doors.mint(authority.snapshot.Ready, selected)\n"),
                                  newfile(f"import {FLOW_SEAM}\n\nvar a189Doors struct {{\n\tmint func(bool, []strategyflow.Result) strategyhandoff.Handoff\n}}\n\n"
                                          "func a189Elsewhere() strategyhandoff.Handoff { return a189Doors.mint(true, nil) }\n")],
     [(PKG3, T3, "PASS")], "같은 빼돌리기를 다른 이름(mint)으로 — 새 판본도 못 잡음(기대 PASS)"),
    ("l2-other-named-field-OLD", "OLD", [(HANDOFF, "once", ADMIT_RETURN,
                                          "\ta189Doors.mint = strategyhandoff.Admit\n\treturn a189Doors.mint(authority.snapshot.Ready, selected)\n"),
                                         newfile(f"import {FLOW_SEAM}\n\nvar a189Doors struct {{\n\tmint func(bool, []strategyflow.Result) strategyhandoff.Handoff\n}}\n\n"
                                                 "func a189Elsewhere() strategyhandoff.Handoff { return a189Doors.mint(true, nil) }\n")],
     [(PKG3, T3, "SAME")], "[OLD 판본] 같은 변이 — 앞 판본도 못 잡음(자리 목록이 OLD 대조군과 같음: 둘 다 같은 한계)"),
    ("l3-wrapper-package", "", [(WRAP, "new", "", f"package a189wrap\n\nimport {SEAM_IMPORT}\n\nvar Admit = strategyhandoff.Admit\n"),
                                newfile(f'import "{MOD}internal/a189wrap"\n\nfunc a189ViaWrap() {{ _ = a189wrap.Admit(true, nil) }}\n')],
     [(PKG3, T3, "PASS"), (PKG_SEAM, T_ONLY_ENGINE, "FAIL")],
     "엔진 밖 래퍼 패키지가 문을 다시 내보내고 엔진이 a189wrap.Admit 을 부름(엔진 census 는 잃음, strategyhandoff 시험이 막음)"),
    ("l3-wrapper-package-OLD", "OLD", [(WRAP, "new", "", f"package a189wrap\n\nimport {SEAM_IMPORT}\n\nvar Admit = strategyhandoff.Admit\n"),
                                       newfile(f'import "{MOD}internal/a189wrap"\n\nfunc a189ViaWrap() {{ _ = a189wrap.Admit(true, nil) }}\n')],
     [(PKG3, T3, "EXTRA:a189ViaWrap:Admit")], "[OLD 판본] 같은 변이 — 앞 판본은 이름으로 잡음(대조군 목록 + a189ViaWrap:Admit)"),
    # #3 인식기 자체의 변이(fixture 시험이 살아 있는지)
    ("g1-ignore-alias", "", [(GUARD, "once", "\t\t\t\tlocal = spec.Name.Name\n\t\t\t}\n\t\t\tseamNames[local] = true\n",
                              "\t\t\t\t_ = spec.Name.Name\n\t\t\t}\n\t\t\tseamNames[local] = true\n")],
     [(PKG3, T3_SPELL, "FAIL")], "별칭을 무시하고 기본 이름만 셈"),
    ("g2-old-rule", "OLD", [], [(PKG3, T3_OUTSIDE, "FAIL"), (PKG3, T3, "FAIL")], "앞 판본(이름 기반)으로 되돌림"),
    ("g3-no-dot", "", [(GUARD, "once", "\t\t\tdotted = true\n", "\t\t\tdotted = false\n")],
     [(PKG3, T3_SPELL, "FAIL")], "dot import 갈래를 끔"),
    ("g4-count-selector-names-in-dot-file", "", [(GUARD, "once", "\t\t\t\tqualified[value.Sel] = true\n", "")],
     [(PKG3, T3_DOT, "FAIL")], "dot import 파일에서 선택자의 Sel 도 맨 이름으로 셈"),
    ("g5-exact-module-path", "", [(GUARD, "once",
                                   'importPath == "strategyhandoff" || strings.HasSuffix(importPath, "/strategyhandoff")',
                                   f'importPath == "{MOD}internal/strategyhandoff"')],
     [(PKG3, T3_SPELL, "FAIL")], "경로 끝 원소 비교를 정확한 모듈 경로 비교로 좁힘"),
    ("g6-no-selector-count", "", [(GUARD, "once",
                                   "\t\t\t\t\tsites = append(sites, declName(decl)+\":\"+value.Sel.Name)\n", "")],
     [(PKG3, T3, "FAIL")], "선택자 세기를 지움"),
    ("g8-trim-instead-of-unquote", "", [(GUARD, "once", "\t\timportPath, err := strconv.Unquote(spec.Path.Value)\n",
                                         "\t\timportPath, err := strings.Trim(spec.Path.Value, `\"`), error(nil)\n")],
     [(PKG3, T3_LITERAL, "FAIL")], "경로 해독을 따옴표 벗기기로 되돌림"),
    ("g7-unaliased-default", "", [(GUARD, "once", '\t\t\tlocal := "strategyhandoff"\n', '\t\t\tlocal := ""\n')],
     [(PKG3, T3, "FAIL")], "별칭 없는 import 의 기본 지역 이름을 잃음"),
]

CONTROLS = {
    "": [(PKG1, T1, "PASS"), (PKG2, T2, "PASS"), (PKG3, T3, "PASS"), (PKG3, T3_SPELL, "PASS"), (PKG3, T3_OUTSIDE, "PASS"),
         (PKG3, T3_UNALIASED, "PASS"), (PKG3, T3_DOT, "PASS"), (PKG3, T3_LITERAL, "PASS"),
         (PKG_A128, T_IMP, "PASS"), (PKG_A128, T_A128_NAMES, "PASS"), (PKG_A140, T_IMP, "PASS"), (PKG_A140, T_A140_NAMES, "PASS"),
         (PKG_SEAM, T_ONLY_ENGINE, "PASS")],
    "S": [(PKG1, T1, "PASS")],
    "PRE1": [(PKG_A128, T_IMP, "PASS"), (PKG_A140, T_IMP, "PASS")],
    # OLD 판본 대조군은 base 실패가 기대값(type 블록:Admit) — 이 변경이 없으면 생산 census 가 빨갛다는 것을 다시 잼
    "OLD": [(PKG3, T3, "FAIL"), (PKG3, T3_SPELL, "PASS")],
}


def sha(p: Path) -> str:
    return hashlib.sha256(p.read_bytes()).hexdigest()


_GO_ENV: dict[str, str] = {}


def isolated_env(cache: Path) -> dict[str, str]:
    if not _GO_ENV:
        _GO_ENV.update(json.loads(subprocess.check_output(["go", "env", "-json", "GOPATH", "GOMODCACHE", "GOENV"], text=True)))
    home = cache / "home"
    home.mkdir(exist_ok=True)
    env = {k: v for k, v in os.environ.items() if k != "GOFLAGS" and not k.startswith("XDG_")}
    env.update(_GO_ENV)
    env.update(HOME=str(home), XDG_CACHE_HOME=str(home / ".cache"), XDG_CONFIG_HOME=str(home / ".config"),
               XDG_DATA_HOME=str(home / ".local/share"), GOCACHE=str(cache / "go-build"))
    return env


def go(base: Path, args: list[str], cache: Path) -> tuple[int, str]:
    proc = subprocess.run(["go", *args], cwd=base, env=isolated_env(cache), stdout=subprocess.PIPE,
                          stderr=subprocess.STDOUT, text=True)
    return proc.returncode, proc.stdout


SITES = re.compile(r"경계에 값을 싣는 생산 자리=\[(.*?)\]")
OLD_BASELINE: list[str] = []  # OLD 판본 대조군의 생산 자리 목록(앞 판본은 base 에서 이미 type 블록:Admit 으로 빨갛다)


def sites_of(out: str) -> list[str] | None:
    m = SITES.search(out)
    # 자리 이름 「type 블록:Admit」 은 공백을 품으므로 split 대신 자리 모양으로 뽑음
    return re.findall(r"(type 블록:\w+|[^\s\[\]]+:\w+)", m.group(1)) if m else None


def judge(base: Path, expects, cache: Path) -> tuple[bool, list[str]]:
    groups: dict = {}
    for tests, name, verdict in expects:
        groups.setdefault(tests, []).append((name, verdict))
    ok, lines = True, []
    for (pkg, run), wants in groups.items():
        _, out = go(base, ["test", "-count=1", "-v", "-run", run, pkg], cache)
        for name, verdict in wants:
            if verdict.startswith("EXTRA:") or verdict == "SAME":
                # 앞 판본 비교: PASS/FAIL 이 아니라 OLD 대조군 자리 목록과의 차이로 판정함(대조군이 이미 FAIL 이므로)
                got = sites_of(out)
                check(bool(OLD_BASELINE), "OLD 대조군 자리 목록이 없음")
                # 다중 집합 차이(같은 자리가 여러 번 나오는 것도 셈)
                extra = sorted((Counter(got) - Counter(OLD_BASELINE)).elements()) if got is not None else None
                if verdict == "SAME":
                    hit = got is not None and sorted(got) == sorted(OLD_BASELINE)
                else:
                    hit = bool(extra) and verdict[6:] in extra
                ok &= hit
                lines.append(f"    {'ok ' if hit else 'MISS'} 기대 {verdict} {name} — 대조군 대비 더해진 자리 {extra}")
                continue
            hit = f"--- {verdict}: {name} (" in out
            ok &= hit
            lines.append(f"    {'ok ' if hit else 'MISS'} 기대 {verdict} {name} ({pkg})")
        for l in out.splitlines():
            if l.startswith("--- FAIL") or re.match(r"^\s+\S+_test\.go:\d+: ", l):
                lines.append("      " + l.strip()[:400])
    return ok, lines


def write(base: Path, rel: str, text: str | None) -> None:
    p = base / rel
    if text == "\0symlink":
        p.unlink()
        return
    if text is None:
        p.unlink()
        if not any(p.parent.iterdir()):
            p.parent.rmdir()
    else:
        p.parent.mkdir(parents=True, exist_ok=True)
        p.write_text(text)


def apply(base: Path, edits) -> dict[str, str | None]:
    """편집을 적용하고, 원복용 원래 내용(새 파일이면 None)을 돌려줌. 중간 편집이 실패하면 앞 편집을 되돌린 뒤 다시 던짐
    (gstack 리뷰 2회차: 앞 편집이 디스크에 남아 다음 변이의 기준이 되던 결함)."""
    saved: dict[str, str | None] = {}
    try:
        _apply_into(base, edits, saved)
    except Fail:
        restore(base, saved)
        raise
    return saved


def _apply_into(base: Path, edits, saved: dict[str, str | None]) -> None:
    for rel, kind, old, new in edits:
        p = base / rel
        if kind == "symlink":
            # 디렉터리 심링크(3회차 리뷰): old = 링크가 가리킬 상대 경로
            check(not p.exists() and not p.is_symlink(), f"심링크 자리가 이미 있음: {rel}")
            saved.setdefault(rel, "\0symlink")
            p.parent.mkdir(parents=True, exist_ok=True)
            os.symlink(old, p)
            continue
        if kind == "new":
            check(not p.exists() or rel in saved, f"새 파일 자리가 이미 있음: {rel}")
            saved.setdefault(rel, None)
            write(base, rel, new)
            continue
        text = p.read_text()
        saved.setdefault(rel, text)
        n = text.count(old)
        check(n == 1, f"TARGET-COUNT {n} in {rel}: {old[:80]!r}")
        text = text.replace(old, new, 1)
        if new:
            check(new in text, f"변이 문자열이 사본에 없음: {new[:80]!r}")
        p.write_text(text)


def restore(base: Path, saved: dict[str, str | None]) -> None:
    for rel, text in saved.items():
        write(base, rel, text)


def main() -> int:
    root = Path(subprocess.check_output(["git", "rev-parse", "--show-toplevel"], text=True).strip())
    head = subprocess.check_output(["git", "-C", str(root), "rev-parse", "HEAD"], text=True).strip()
    wanted = set(sys.argv[1:])
    unknown = sorted(wanted - {m[0] for m in MUTANTS})
    if unknown:
        print(f"알 수 없는 변이 ID: {unknown}")
        return 2
    selected = [m for m in MUTANTS if not wanted or m[0] in wanted]
    work = Path(os.environ["HOME"]) / "tossos-work"
    base = Path(tempfile.mkdtemp(prefix=f"a189-mut2-{os.getpid()}-", dir=work))
    cache = Path(tempfile.mkdtemp(prefix=f"a189-mut2-gocache-{os.getpid()}-", dir=work))
    try:
        archive = subprocess.run(["git", "-C", str(root), "archive", "HEAD", "go.mod", "go.sum", "cmd", "internal", "tools", "docs/migration"],
                                 check=True, stdout=subprocess.PIPE).stdout
        subprocess.run(["tar", "-x", "-C", str(base)], input=archive, check=True)
        check((base / "go.mod").is_file() and not (base / ".git").exists(), "사본 모양이 다름")
        tracked = [CONSOLE, CONSOLE_TEST, EXITLOOP, HANDOFF, GUARD, A128_DIR + "/static_test.go", A140_DIR + "/static_test.go"]
        starts = {f: sha(base / f) for f in tracked}
        for f in tracked:
            check(starts[f] == sha(root / f), f"작업 트리 {f} 가 HEAD 와 다름 — 커밋 뒤 돌릴 것")
        check(not (base / NEWFILE).exists() and not (base / WRAP).exists() and not (base / LINKPKG).exists(), "새 파일 자리가 사본에 이미 있음")
        print(f"기준 HEAD {head}, 사본 {base.name}($HOME/tossos-work 아래), 전용 GOCACHE, HOME·XDG 격리")
        for f in tracked:
            print(f"  {f} sha256={starts[f]}")
        text_of = lambda rel: (base / rel).read_text()
        for name in sorted({m[1] for m in selected} | {""}):
            saved = apply(base, variant_edits(name, text_of))
            ok, lines = judge(base, CONTROLS[name], cache)
            if name == "OLD":
                _, out = go(base, ["test", "-count=1", "-run", f"^{T3}$", PKG3[0]], cache)
                OLD_BASELINE.extend(sites_of(out) or [])
                lines.append(f"    OLD 대조군 생산 자리 목록: {OLD_BASELINE}")
                ok &= "type 블록:Admit" in OLD_BASELINE
            restore(base, saved)
            print(f"control[{name or '현재'}]: {'기대대로' if ok else 'MISMATCH — 변이 판정 중단'}")
            print("\n".join(lines))
            if not ok:
                return 1
        issues = 0
        for mid, variant, edits, expects, label in selected:
            v_edits = variant_edits(variant, text_of)
            saved, saved_m = apply(base, v_edits), {}
            try:
                saved_m = apply(base, edits)
            except Fail as exc:
                print(f"{mid}: TARGET — {label}: {exc}")
                issues += 1
                restore(base, saved)
                continue
            # 변이·판본이 닿은 패키지를 전부 시험 파일까지 컴파일함(빌드 실패를 CAUGHT 로 세지 않게)
            # 점으로 시작하는 경로(go 도구가 무시)와 심링크 편집은 컴파일 대상이 아님 — 그 둘은 소스 훑기 시험만 잼
            touched = sorted({"./" + str(Path(e[0]).parent) + "/" for e in [*v_edits, *edits]
                              if e[1] != "symlink" and not any(part.startswith(".") for part in Path(e[0]).parts)})
            build_ok, bout = True, ""
            for pkg in touched:
                s, o = go(base, ["test", "-count=1", "-run", "^$", pkg], cache)
                if s != 0:
                    build_ok, bout = False, o
                    break
            if build_ok:
                ok, lines = judge(base, expects, cache)
            restore(base, saved_m)
            restore(base, saved)
            if not build_ok:
                print(f"{mid}: INVALID(빌드 실패) — {label}")
                print("    " + "\n    ".join(bout.splitlines()[:6]))
                issues += 1
                continue
            print(f"{mid}: {'기대대로' if ok else 'MISMATCH'} — {label}")
            print("\n".join(lines))
            issues += 0 if ok else 1
        for f in tracked:
            check(sha(base / f) == starts[f], f"{f} 끝 원복 sha 불일치")
        check(not (base / NEWFILE).exists() and not (base / WRAP).exists() and not (base / LINKPKG).exists(), "새 파일이 남음")
        print(f"끝: 원복 sha 전부 일치, 새 파일 0, 판정 이상 {issues}건")
        return 1 if issues else 0
    finally:
        shutil.rmtree(base, ignore_errors=True)
        shutil.rmtree(cache, ignore_errors=True)


if __name__ == "__main__":
    sys.exit(main())
