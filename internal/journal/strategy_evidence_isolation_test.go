package journal

// v21 이 trading journal 에 **무엇을 더했는지**를 문자열 검색이 아니라 스키마 측정으로
// 잰다.
//
// 원래 근거는 schemaV21 이라는 Go 문자열 리터럴을 소문자로 바꿔 "payload", "create table"
// 같은 조각이 들어 있는지 보는 것이었다. 같은 패키지의 리터럴을 들여다보는 검사는 (a)
// 다른 migration 이 만든 표를 볼 수 없고 (b) "CREATE  TABLE"(공백 둘)에 걸리지 않는다.
// tasks.md 3.3 이 요구한 것은 그런 열·표가 **없다**는 것이므로, v20 과 v21 저장소의
// 실제 스키마를 떠서 차이를 센다.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type journalSchemaShape struct {
	objects map[string]string // "type:name" → 정규화한 SQL
	columns map[string]string // "table.column" → 선언 타입
}

func journalSchemaAt(t *testing.T, version int) journalSchemaShape {
	t.Helper()
	j := openJournalAtSchema(t, journalFileAtSchemaPath(t), version)
	defer func() {
		if err := j.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	shape := journalSchemaShape{objects: map[string]string{}, columns: map[string]string{}}
	rows, err := j.db.Query(`SELECT type,name,COALESCE(sql,'') FROM sqlite_master ORDER BY type,name`)
	if err != nil {
		t.Fatal(err)
	}
	type object struct{ kind, name, sql string }
	var objects []object
	for rows.Next() {
		var kind, name, statement string
		if err := rows.Scan(&kind, &name, &statement); err != nil {
			t.Fatal(err)
		}
		objects = append(objects, object{kind, name, statement})
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	for _, one := range objects {
		shape.objects[one.kind+":"+one.name] = strings.Join(strings.Fields(one.sql), " ")
		if one.kind != "table" {
			continue
		}
		columns, err := j.db.Query(`SELECT name,type FROM pragma_table_info(?)`, one.name)
		if err != nil {
			t.Fatal(err)
		}
		for columns.Next() {
			var name, kind string
			if err := columns.Scan(&name, &kind); err != nil {
				t.Fatal(err)
			}
			shape.columns[one.name+"."+name] = kind
		}
		if err := columns.Err(); err != nil {
			t.Fatal(err)
		}
		if err := columns.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if len(shape.objects) == 0 || len(shape.columns) == 0 {
		t.Fatalf("v%d schema shape came back empty", version)
	}
	return shape
}

func journalFileAtSchemaPath(t *testing.T) string {
	t.Helper()
	return t.TempDir() + "/" + DBFileName
}

func addedKeys(before, after map[string]string) []string {
	var added []string
	for key := range after {
		if _, existed := before[key]; !existed {
			added = append(added, key)
		}
	}
	sort.Strings(added)
	return added
}

func TestMigrationV21AddsExactlyTwoColumnsAndTwoTriggersAndNothingElse(t *testing.T) {
	before := journalSchemaAt(t, 20)
	after := journalSchemaAt(t, 21)

	wantColumns := []string{
		"strategy_decision_lineage.consumed_evidence_snapshot_digest",
		"strategy_decision_lineage.consumed_evidence_snapshot_id",
	}
	if got := addedKeys(before.columns, after.columns); !equalStringSlices(got, wantColumns) {
		t.Fatalf("v21 added columns %v, want exactly %v", got, wantColumns)
	}
	wantObjects := []string{
		"trigger:strategy_evidence_reference_insert_guard",
		"trigger:strategy_evidence_reference_update_guard",
	}
	if got := addedKeys(before.objects, after.objects); !equalStringSlices(got, wantObjects) {
		t.Fatalf("v21 added schema objects %v, want exactly %v", got, wantObjects)
	}
	for key := range before.columns {
		if _, kept := after.columns[key]; !kept {
			t.Fatalf("v21 removed column %s", key)
		}
	}
	for key := range before.objects {
		if _, kept := after.objects[key]; !kept {
			t.Fatalf("v21 removed schema object %s", key)
		}
	}

	// 새로 생긴 것 어디에도 증거 payload·원문·자격증명 저장이 없어야 한다. 이제 이
	// 판정은 Go 리터럴이 아니라 **실제 스키마**를 본다.
	forbidden := []string{"payload", "revision", "credential", "secret", "source_response", "header"}
	for _, key := range append(addedKeys(before.columns, after.columns), addedKeys(before.objects, after.objects)...) {
		lower := strings.ToLower(key)
		for _, fragment := range forbidden {
			if strings.Contains(lower, fragment) {
				t.Fatalf("v21 introduced evidence storage %q (matched %q)", key, fragment)
			}
		}
	}
	// 트리거 **본문**만 본다. 머리말의 "BEFORE UPDATE OF ..."·"BEFORE INSERT ON ..."은
	// 트리거가 언제 도는지를 말할 뿐 쓰기가 아니다. 본문이 RAISE(ABORT) 외의 무엇을
	// 하면 v21 이 저널을 조용히 고치는 것이 된다.
	for _, key := range wantObjects {
		statement := strings.ToLower(after.objects[key])
		_, body, found := strings.Cut(statement, " begin ")
		if !found {
			t.Fatalf("v21 trigger %s has no BEGIN body: %s", key, after.objects[key])
		}
		if !strings.Contains(body, "raise(abort") {
			t.Fatalf("v21 trigger %s body does not abort: %s", key, after.objects[key])
		}
		for _, fragment := range []string{"create ", "insert ", "update ", "delete ", "drop ", "alter ", "replace "} {
			if strings.Contains(body, fragment) {
				t.Fatalf("v21 trigger %s body mutates the journal (%q): %s", key, fragment, after.objects[key])
			}
		}
	}
}

// journalPackageDeclarations 는 journal 패키지의 생산 함수 전부를 이름으로 모은다.
func journalPackageDeclarations(t *testing.T) map[string][]*ast.FuncDecl {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	declarations := map[string][]*ast.FuncDecl{}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, parseErr := parser.ParseFile(token.NewFileSet(), name, nil, 0)
		if parseErr != nil {
			t.Fatalf("parsing %s: %v", name, parseErr)
		}
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			declarations[function.Name.Name] = append(declarations[function.Name.Name], function)
		}
	}
	if len(declarations) == 0 {
		t.Fatal("no journal function declarations parsed")
	}
	return declarations
}

func TestStrategyEvidenceReadBoundaryCallChainIsSelectOnly(t *testing.T) {
	declarations := journalPackageDeclarations(t)
	roots := []string{"ConsumedSnapshot", "NewStrategyEvidenceReadBoundary", "validConsumedEvidenceReference"}
	for _, root := range roots {
		if len(declarations[root]) == 0 {
			t.Fatalf("evidence read boundary entry point %s is not declared", root)
		}
	}
	reached := map[string]bool{}
	queue := append([]string(nil), roots...)
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		if reached[name] {
			continue
		}
		reached[name] = true
		for _, function := range declarations[name] {
			ast.Inspect(function.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				// 로컬 선언이 있는 이름만 따라간다. 수신자 타입은 풀지 않으므로 같은
				// 이름의 메서드를 전부 끌어온다 — 넓게 보는 쪽으로 틀린다.
				switch callee := call.Fun.(type) {
				case *ast.Ident:
					if len(declarations[callee.Name]) > 0 {
						queue = append(queue, callee.Name)
					}
				case *ast.SelectorExpr:
					if len(declarations[callee.Sel.Name]) > 0 {
						queue = append(queue, callee.Sel.Name)
					}
				}
				return true
			})
		}
	}
	if findings := journalMutationsIn(t, declarations, reached); len(findings) != 0 {
		t.Fatalf("evidence read boundary chain (%d functions) can mutate the journal: %v", len(reached), findings)
	}

	// 양성 대조군: 같은 계측기를 쓰기 경로에 대면 반드시 쓰기를 찾아야 한다.
	writing := map[string]bool{"insertExactStrategyDecision": true}
	if findings := journalMutationsIn(t, declarations, writing); len(findings) == 0 {
		t.Fatal("SELECT-only detector found no write in insertExactStrategyDecision; it cannot fail")
	}
}

func journalMutationsIn(t *testing.T, declarations map[string][]*ast.FuncDecl, reached map[string]bool) []string {
	t.Helper()
	mutatingCalls := []string{"Exec", "ExecContext", "Begin", "BeginTx"}
	mutatingSQL := []string{"INSERT ", "UPDATE ", "DELETE ", "ALTER ", "DROP ", "CREATE ", "REPLACE ", "PRAGMA "}
	var findings []string
	for name := range reached {
		for _, function := range declarations[name] {
			ast.Inspect(function.Body, func(node ast.Node) bool {
				switch value := node.(type) {
				case *ast.SelectorExpr:
					for _, forbidden := range mutatingCalls {
						if value.Sel.Name == forbidden {
							findings = append(findings, name+" calls "+forbidden)
						}
					}
				case *ast.BasicLit:
					if value.Kind != token.STRING {
						break
					}
					literal, err := strconv.Unquote(value.Value)
					if err != nil {
						break
					}
					normalized := strings.Join(strings.Fields(strings.ToUpper(literal)), " ") + " "
					for _, forbidden := range mutatingSQL {
						if strings.Contains(normalized, forbidden) {
							findings = append(findings, name+" contains SQL "+strings.TrimSpace(forbidden))
						}
					}
				}
				return true
			})
		}
	}
	return findings
}

func equalStringSlices(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
