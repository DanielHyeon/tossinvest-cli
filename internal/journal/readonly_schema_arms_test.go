package journal

// checkSchema 의 갈래들을 실제로 실행한다.
//
// 두 종류가 한 번도 돌지 않았다.
//
//	오류 갈래  `case err != nil` 여섯 개. OpenReadOnly 는 파일 존재와 Ping 을 먼저
//	           확인하므로 정상 파일에서는 드라이버 오류가 생길 자리가 없고, 시험 16곳이
//	           전부 정상 파일을 준다. a064 가 **직접 넣은** v21 갈래도 여기 포함된다.
//	버전 관문  `r.version >= 20` 과 `>= 21` 의 거짓 쪽. v20 이하 완전한 저널을 읽기
//	           전용으로 여는 시험이 없어서, 관문을 지워도 아무것도 깨지지 않았다.
//
// 오류 갈래는 실제 sqlite 연결에 질의를 넘기다가 **고른 한 질의에서만** 실패하는
// 위임 드라이버로 연다. 정답 행은 진짜 저널에서 나오므로, 시험이 스키마를 흉내 내지 않는다.

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"
)

type schemaProbeConnector struct {
	real   *sql.DB
	failOn func(query string, args []driver.NamedValue) bool
}

func (c *schemaProbeConnector) Connect(context.Context) (driver.Conn, error) {
	return &schemaProbeConn{connector: c}, nil
}

func (c *schemaProbeConnector) Driver() driver.Driver { return schemaProbeDriver{} }

type schemaProbeDriver struct{}

func (schemaProbeDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("journal test: schema probe is connector-only")
}

type schemaProbeConn struct{ connector *schemaProbeConnector }

func (c *schemaProbeConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("journal test: schema probe does not prepare")
}

func (c *schemaProbeConn) Close() error { return nil }

func (c *schemaProbeConn) Begin() (driver.Tx, error) {
	return nil, errors.New("journal test: schema probe never writes")
}

func (c *schemaProbeConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if c.connector.failOn != nil && c.connector.failOn(query, args) {
		return nil, errors.New("journal test: injected schema-inspection failure")
	}
	values := make([]any, len(args))
	for index, arg := range args {
		values[index] = arg.Value
	}
	rows, err := c.connector.real.QueryContext(ctx, query, values...)
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		_ = rows.Close()
		return nil, err
	}
	return &schemaProbeRows{rows: rows, columns: columns}, nil
}

type schemaProbeRows struct {
	rows    *sql.Rows
	columns []string
}

func (r *schemaProbeRows) Columns() []string { return r.columns }

func (r *schemaProbeRows) Close() error { return r.rows.Close() }

func (r *schemaProbeRows) Next(dest []driver.Value) error {
	if !r.rows.Next() {
		if err := r.rows.Err(); err != nil {
			return err
		}
		return io.EOF
	}
	values := make([]any, len(dest))
	holders := make([]any, len(dest))
	for index := range holders {
		holders[index] = &values[index]
	}
	if err := r.rows.Scan(holders...); err != nil {
		return err
	}
	for index := range dest {
		dest[index] = values[index]
	}
	return nil
}

func probeReadOnly(t *testing.T, path string, failOn func(string, []driver.NamedValue) bool) *ReadOnly {
	t.Helper()
	real, err := sql.Open("sqlite", readOnlyDSN(path, DefaultBusyTimeout))
	if err != nil {
		t.Fatal(err)
	}
	real.SetMaxOpenConns(1)
	probe := sql.OpenDB(&schemaProbeConnector{real: real, failOn: failOn})
	probe.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = probe.Close()
		_ = real.Close()
	})
	return &ReadOnly{db: probe, path: path}
}

func argumentIs(wanted string) func(string, []driver.NamedValue) bool {
	return func(_ string, args []driver.NamedValue) bool {
		for _, arg := range args {
			if text, ok := arg.Value.(string); ok && text == wanted {
				return true
			}
		}
		return false
	}
}

func journalFileAtSchema(t *testing.T, version int) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), DBFileName)
	j := openJournalAtSchema(t, path, version)
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCheckSchemaReportsEveryInspectionFailureInsteadOfSwallowingIt(t *testing.T) {
	path := journalFileAtSchema(t, 21)
	for _, test := range []struct {
		name     string
		failOn   func(string, []driver.NamedValue) bool
		fragment string
	}{
		{
			name:     "schema-version",
			failOn:   func(query string, _ []driver.NamedValue) bool { return strings.Contains(query, "user_version") },
			fragment: "reading the schema version",
		},
		{name: "core-table", failOn: argumentIs("positions"), fragment: "inspecting the schema"},
		{name: "core-column", failOn: argumentIs("policy_id"), fragment: "inspecting the schema"},
		{name: "campaign-table", failOn: argumentIs("position_campaigns"), fragment: "inspecting the schema"},
		{name: "campaign-column", failOn: argumentIs("residual_quantity"), fragment: "inspecting the schema"},
		// a064 가 직접 넣은 갈래(branch-test-map 의 B27). 지금까지 한 번도 돌지 않았다.
		{name: "v21-evidence-column", failOn: argumentIs("consumed_evidence_snapshot_id"), fragment: "inspecting the schema"},
	} {
		t.Run(test.name, func(t *testing.T) {
			readonly := probeReadOnly(t, path, test.failOn)
			err := readonly.checkSchema(context.Background())
			if err == nil {
				t.Fatal("an inspection failure was swallowed and the schema reported healthy")
			}
			if !strings.Contains(err.Error(), test.fragment) {
				t.Fatalf("inspection failure reported as %v, want a %q refusal", err, test.fragment)
			}
			if errors.Is(err, ErrSchemaTooOld) || errors.Is(err, ErrSchemaTooNew) {
				t.Fatalf("a driver failure was misreported as a schema-version verdict: %v", err)
			}
		})
	}

	// 양성 대조군: 주입을 끄면 같은 위임 드라이버로 스키마가 통과한다. 이것이 없으면
	// "드라이버가 늘 깨진다"가 위의 여섯을 전부 통과시킨다.
	t.Run("no-injection", func(t *testing.T) {
		readonly := probeReadOnly(t, path, nil)
		if err := readonly.checkSchema(context.Background()); err != nil {
			t.Fatalf("complete v21 journal refused through the probe driver: %v", err)
		}
		if readonly.SchemaVersion() != 21 {
			t.Fatalf("probe read schema version %d, want 21", readonly.SchemaVersion())
		}
	})
}

func TestOpenReadOnlyAcceptsCompleteJournalsBelowTheVersionGates(t *testing.T) {
	// v19 는 campaign 관문(>=20) 아래, v20 은 evidence 관문(>=21) 아래다. 두 관문을
	// 지우면 이 두 저널이 "없는 열"을 요구받아 ErrSchemaTooOld 로 거부된다.
	for _, version := range []int{19, 20} {
		t.Run(fmt.Sprintf("v%d", version), func(t *testing.T) {
			path := journalFileAtSchema(t, version)
			readonly, err := OpenReadOnly(context.Background(), ReadOnlyOptions{Path: path})
			if err != nil {
				t.Fatalf("complete v%d journal refused read-only: %v", version, err)
			}
			if got := readonly.SchemaVersion(); got != version {
				t.Fatalf("read-only schema version = %d, want %d", got, version)
			}
			if err := readonly.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestOpenReadOnlyNamesTheMissingV21ColumnIndividually(t *testing.T) {
	// "consumed_evidence_snapshot" 를 부분 문자열로 찾으면 두 열 어느 쪽에도 걸린다.
	// 한 열만 빠뜨린 저널을 만들어, 빠진 쪽 **이름 그대로** 지목하는지 본다.
	for _, test := range []struct {
		name, present, missing string
	}{
		{name: "digest-missing", present: "consumed_evidence_snapshot_id", missing: "consumed_evidence_snapshot_digest"},
		{name: "id-missing", present: "consumed_evidence_snapshot_digest", missing: "consumed_evidence_snapshot_id"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := journalFileAtSchema(t, 20)
			raw, err := sql.Open("sqlite", "file:"+path)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := raw.Exec(`ALTER TABLE strategy_decision_lineage ADD COLUMN ` + test.present + ` TEXT`); err != nil {
				t.Fatal(err)
			}
			if _, err := raw.Exec(`PRAGMA user_version=21`); err != nil {
				t.Fatal(err)
			}
			if err := raw.Close(); err != nil {
				t.Fatal(err)
			}
			_, err = OpenReadOnly(context.Background(), ReadOnlyOptions{Path: path})
			if !errors.Is(err, ErrSchemaTooOld) {
				t.Fatalf("half-migrated v21 journal error=%v, want ErrSchemaTooOld", err)
			}
			if !strings.Contains(err.Error(), test.missing) {
				t.Fatalf("refusal does not name the missing column %s: %v", test.missing, err)
			}
			if strings.Contains(err.Error(), test.present) {
				t.Fatalf("refusal names the column that is present (%s): %v", test.present, err)
			}
		})
	}
}
