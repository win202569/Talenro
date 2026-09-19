//go:build integration

package testinfra

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func c12PolicyFixture() *c12FixtureLedger {
	operationID := uuid.MustParse("81000000-0000-4000-8000-000000000001")
	nodeID := uuid.MustParse("81000000-0000-4000-8000-000000000002")
	certificateID := uuid.NewSHA1(operationID, []byte("certificate"))
	scope := sha256.Sum256(append([]byte("TALENRO-NODE-AUTHORITY-SCOPE-V1\x00"), nodeID[:]...))
	return &c12FixtureLedger{activationID: uuid.MustParse("79000000-0000-4000-8000-000000000001"), epoch: 31,
		certificates: map[uuid.UUID]c12FixtureCertificate{certificateID: {nodeID: nodeID, scopeDigest: scope, historySequence: 1}},
		operations:   []c12FixtureOperation{{operationID: operationID, nodeID: nodeID, certificateID: certificateID, epoch: 31, sequence: 1, scopeDigest: scope}}}
}

type c12SourceOwner struct {
	sql     []string
	argsSHA string
}

var c12SQLSourceOwners = map[string]c12SourceOwner{
	"GetNodeControlDatabaseIdentity":            {sql: []string{c12GetNodeControlDatabaseIdentitySQL}, argsSHA: "062ed05a49164fd40b25dd1873a7cfe12bb767712ad34499d5404a09cf01ec1b"},
	"GetAuthorityFenceHead":                     {sql: []string{c12GetAuthorityFenceHeadSQL}, argsSHA: "a4dabf8aa4aa5e569a169d87caf2433563e91e71c3297b8d3e54f75eefb26ac2"},
	"ListPendingAuthorityFences":                {sql: []string{c12ListPendingAuthorityFencesSQL}, argsSHA: "a1e7cf1ec779172ce4f06c70eb6876341a48e097ea0bedca5fb3a3249f51fa08"},
	"LockAuthorityFence":                        {sql: []string{c12LockAuthorityFenceSQL}, argsSHA: "4cb3429b0ee8016fd3539e11ca0d9cd9c409b93fbdf113cbfc825822746094d3"},
	"LockCertificateRevocationOutcome":          {sql: []string{c12LockCertificateRevocationOutcomeSQL}, argsSHA: "60e849ad4ea08ac7c49bf468afe1ef3f91cac922bb0c5e7719b196ab543a2a16"},
	"GetStoredAuthorityFence":                   {sql: []string{c12GetStoredAuthorityFenceSQL}, argsSHA: "e4825697aaea75b16075381e7183debc1650d27406b7de1422bae8e63540a8fe"},
	"GetAuthorityFenceForUpdate":                {sql: []string{c12GetAuthorityFenceForUpdateSQL}, argsSHA: "6122cf7c60fa91606208b206e8ac63ad4c4ac1d2707950fb2864f48ca47ac221"},
	"InsertClaimV1AuthorityFencePending":        {sql: []string{c12InsertClaimV1AuthorityFencePendingSQL}, argsSHA: "8f4202fd446a1e36a42cf38e2c76306c3486a115d52f2495570b3db46060a462"},
	"BindAuthorityFenceEffect":                  {sql: []string{c12BindAuthorityFenceEffectSQL}, argsSHA: "3818c9e23ff4df258e32e4ea981c2d71b61dafbcb170117101aae00db187e96d"},
	"ActivateCommittedAuthorityFence":           {sql: []string{c12ActivateCommittedAuthorityFenceSQL}, argsSHA: "a5f06309eb86de08b334295e77e30658435a83c7fa9a8c41bf75eb49c89ef96d"},
	"ResolveRegisteredAuthorityEffectForUpdate": {sql: []string{c12AuxResolveSQL, c12CertificateResolutionSQL}, argsSHA: "81655de1fb027843970e0dbeef207d626e38af671ccbd906828e828e1062bc02"},
	"task9ReadCertificateInput":                 {sql: []string{c12CertificateInputSQL}, argsSHA: "96d621e28abedf083da32826d7da990c42182ed48b124a733fcc57e1cb654d27"},
	"ValidatePersistedAuthorityEffect":          {sql: []string{c12PersistedEffectSQL, c12AuditOutboxCountsSQL, c12OutboxPayloadSQL}, argsSHA: "1cae982b4bfa52670939ddbd42ab6e12e5e8eef0dc593de9aad88959f0b63cce"},
	"CaptureActivationDecisionMaterial":         {sql: []string{c12CommitmentReadSQL}, argsSHA: "3952a97e390b0826173ea037434a16294967171ae4a02443cb6092a1afc77721"},
	"commitDomain":                              {sql: []string{c12CommitmentWriteSQL, c12AuxInsertSQL}, argsSHA: "3475924fe6daae04d181a13d22815059f9872641d8268b02c6aaccbf23267bec"},
	"ActivateAuthorityEffect":                   {sql: []string{c12CertificateActivateSQL, c12AuditInsertSQL, c12OutboxInsertSQL}, argsSHA: "35499d194ed2510a002f21ba3518d74198eb6df4db762e17a0b9d4c6651e2c06"},
	"task9SeedProofActivation":                  {sql: []string{c12LatchSQL, c12ReplicaSQL, c12UpgradeIntentSQL, c12RuntimeRegistrationSQL, c12UpgradeAttemptSQL, c12ProtocolActivationSQL, c12ActivationCompletionSQL, c12ActivationReleaseSQL, c12OriginSQL}, argsSHA: "e1457e4aadfd8381e3da92b7f9122d54af80c9dd5aa63ae6520b9199c2e8a26b"},
	"seedCertificate":                           {sql: []string{c12ReplicaSQL, c12NodePopSQL, c12NodeInventorySQL, c12LegacyFenceSQL, c12IssuanceSQL, c12OriginSQL, c12CertificateSeedSQL}, argsSHA: "fe7c6b191c2b7a1f044394007061c72a48ef10792d26f0989e90ade45422d459"},
	"task9InstallCrashAuxiliaryTables":          {sql: []string{c12AuxDDLSQL, ""}, argsSHA: "f6c4be6d24a7430d27dfbea67b286850a6bf98896fe472557694fe661eb89fb6"},
}

// The three snapshots are frozen here at Task2. Task5 MUST add their real
// authority bridge owner and require AST parity; this is not a missing-file skip.
func TestC12AuthorityPITRSQLRegistryMatchesSources(t *testing.T) {
	storeSource, err := os.ReadFile("../store/nodecontrol_authority.sql.go")
	if err != nil {
		t.Fatal(err)
	}
	fixtureSource, err := os.ReadFile("../nodecontrol/authority/authority_crash_integration_test.go")
	if err != nil {
		t.Fatal(err)
	}
	storeSource = bytes.ReplaceAll(storeSource, []byte("\r\n"), []byte("\n"))
	fixtureSource = bytes.ReplaceAll(fixtureSource, []byte("\r\n"), []byte("\n"))
	if err := c12CheckSQLSources(storeSource, fixtureSource, t.Logf); err != nil {
		t.Fatal(err)
	}
	t.Run("unmapped_registry_entry", func(t *testing.T) {
		c12AuthoritySQLRules["SELECT arbitrary_extra_capability"] = c12SQLRule{call: c12SQLQueryRow, modes: uint8(c12SQLPrimary), arity: 0, validate: func([]any, *c12FixtureLedger) bool { return true }}
		defer delete(c12AuthoritySQLRules, "SELECT arbitrary_extra_capability")
		if err := c12CheckSQLSources(storeSource, fixtureSource, nil); err == nil {
			t.Fatal("unmapped runtime SQL capability escaped source inventory")
		}
	})
	for _, mutation := range []struct {
		name           string
		store, fixture []byte
	}{
		{"SQL byte", bytes.Replace(storeSource, []byte("-- name: GetAuthorityFenceHead"), []byte("-- name: XetAuthorityFenceHead"), 1), fixtureSource},
		{"call kind", bytes.Replace(storeSource, []byte("q.db.QueryRow(ctx, lockAuthorityFence"), []byte("q.db.Query(ctx, lockAuthorityFence"), 1), fixtureSource},
		{"argument position", bytes.Replace(storeSource, []byte("arg.DbSystemID,\n\t\targ.DbTimeline,"), []byte("arg.DbTimeline,\n\t\targ.DbSystemID,"), 1), fixtureSource},
		{"new DBTX call", storeSource, bytes.Replace(fixtureSource, []byte("func task9SeedProofActivation(t *testing.T, tx pgx.Tx, activationID uuid.UUID, discriminator int, now time.Time) {"), []byte("func task9SeedProofActivation(t *testing.T, tx pgx.Tx, activationID uuid.UUID, discriminator int, now time.Time) { tx.Exec(t.Context(), `SELECT 1`);"), 1)},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			if err := c12CheckSQLSources(mutation.store, mutation.fixture, nil); err == nil {
				t.Fatal("source drift authorized")
			}
		})
	}
	for _, sql := range []string{pitrCertificateSnapshotSQL, pitrOperationSnapshotSQL, pitrAuxCountSQL} {
		r, ok := c12AuthoritySQLRules[sql]
		if !ok || r.call != c12SQLQueryRow || r.arity != 1 || r.modes != uint8(c12SQLPrimary|c12SQLCandidate) {
			t.Fatal("staged snapshot not frozen")
		}
	}
}

func c12CheckSQLSources(storeSource, fixtureSource []byte, log func(string, ...any)) error {
	seen := make(map[string]bool)
	var failures []string
	expectedRules := map[string]bool{pitrCertificateSnapshotSQL: true, pitrOperationSnapshotSQL: true, pitrAuxCountSQL: true}
	for _, owner := range c12SQLSourceOwners {
		for _, sql := range owner.sql {
			if sql != "" && sql != c12AuxDDLSQL {
				expectedRules[sql] = true
			}
		}
	}
	for sql := range c12AuthoritySQLRules {
		if !expectedRules[sql] {
			failures = append(failures, "unmapped runtime capability")
		}
	}
	if len(expectedRules) != len(c12AuthoritySQLRules) {
		failures = append(failures, "closed runtime inventory mismatch")
	}
	for _, src := range [][]byte{storeSource, fixtureSource} {
		fs := token.NewFileSet()
		file, err := parser.ParseFile(fs, "source.go", src, 0)
		if err != nil {
			return err
		}
		constants := make(map[string]string)
		ast.Inspect(file, func(node ast.Node) bool {
			spec, ok := node.(*ast.ValueSpec)
			if !ok {
				return true
			}
			for i, name := range spec.Names {
				if i < len(spec.Values) {
					if lit, ok := spec.Values[i].(*ast.BasicLit); ok && lit.Kind == token.STRING {
						constants[name.Name], _ = strconv.Unquote(lit.Value)
					}
				}
			}
			return true
		})
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			owner, wanted := c12SQLSourceOwners[fn.Name.Name]
			if !wanted {
				continue
			}
			seen[fn.Name.Name] = true
			count := 0
			var args bytes.Buffer
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				kind := selector.Sel.Name
				if kind != "Exec" && kind != "Query" && kind != "QueryRow" {
					return true
				}
				if len(call.Args) < 2 {
					failures = append(failures, fn.Name.Name+": missing SQL")
					return true
				}
				var sql string
				switch expr := call.Args[1].(type) {
				case *ast.BasicLit:
					sql, _ = strconv.Unquote(expr.Value)
				case *ast.Ident:
					sql = constants[expr.Name]
				}
				if count >= len(owner.sql) {
					failures = append(failures, fn.Name.Name+": new DBTX call")
					return true
				}
				expected := owner.sql[count]
				count++
				if expected != "" && (sql != expected || sha256.Sum256([]byte(sql)) != sha256.Sum256([]byte(expected))) {
					failures = append(failures, fn.Name.Name+": SQL drift")
				}
				if rule, ok := c12AuthoritySQLRules[expected]; ok {
					calls := []string{"Exec", "Query", "QueryRow"}
					if kind != calls[rule.call] || len(call.Args)-2 != rule.arity {
						failures = append(failures, fn.Name.Name+": call shape drift")
					}
				} else if fn.Name.Name != "task9InstallCrashAuxiliaryTables" {
					failures = append(failures, fn.Name.Name+": unmapped call")
				}
				args.WriteString(kind)
				args.WriteByte(0)
				// Context/SQL/argument AST fingerprints protect positional arguments and
				// the explicit excluded isolated-child outbox DDL without registering it.
				for _, arg := range call.Args {
					format.Node(&args, fs, arg)
					args.WriteByte(0)
				}
				return true
			})
			if count != len(owner.sql) {
				failures = append(failures, fn.Name.Name+": missing DBTX call")
			}
			digest := fmt.Sprintf("%x", sha256.Sum256(args.Bytes()))
			if owner.argsSHA != digest {
				if log != nil {
					log("source digest %s %s", fn.Name.Name, digest)
				}
				failures = append(failures, fn.Name.Name+": argument/source fingerprint drift")
			}
		}
	}
	for owner := range c12SQLSourceOwners {
		if !seen[owner] {
			failures = append(failures, owner+": missing owner")
		}
	}
	if len(failures) > 0 {
		return errors.New(strings.Join(failures, "; "))
	}
	return nil
}

func TestC12AuthorityPITRSQLPolicy(t *testing.T) {
	t.Run("access_transaction_ledger", c12TestPolicyAccessLedger)
	t.Run("private_initialization", c12TestPrivateInitialization)
	t.Run("all_fixed_statement_arguments", c12TestAllSQLArguments)
	t.Run("setup_commit_and_closure", c12TestSetupLifecycle)
	t.Run("historical_groups_late_binding", c12TestHistoricalGroups)
	t.Run("original_call_kind", c12TestOriginalCallKind)
	t.Run("adapter_copies_before_driver", c12TestAdapterCopies)
	t.Run("domain_separated_artifact_digests", c12TestDomainSeparatedDigests)
	t.Run("argument_copy", func(t *testing.T) {
		b := []byte{1, 2}
		n := pgtype.Numeric{Int: big.NewInt(123), Valid: true}
		args := []any{b, n}
		copied, err := c12CopySQLArguments(args)
		if err != nil {
			t.Fatal(err)
		}
		b[0] = 9
		n.Int.SetInt64(99)
		args[0] = "changed"
		if copied[0].([]byte)[0] != 1 || copied[1].(pgtype.Numeric).Int.Int64() != 123 {
			t.Fatal("caller argument alias")
		}
	})
	ledger := c12PolicyFixture()
	good := pitrAuxCountSQL
	id := ledger.operations[0].operationID
	for _, bad := range []string{good + "; DELETE FROM nodecontrol.node_certificates", good + " -- comment", " " + good, "WITH x AS (DELETE FROM nodecontrol.node_certificates RETURNING *) SELECT * FROM x", "SELECT pg_advisory_lock(1)", "SET ROLE talenro"} {
		if c12AuthorizeSQL(c12SQLCandidate, c12SQLQueryRow, bad, []any{id}, ledger) == nil {
			t.Fatal("non-registered SQL authorized")
		}
	}
	for _, args := range [][]any{{}, {id, id}, {id.String()}, {uuid.New()}, {uuid.Nil}} {
		if c12AuthorizeSQL(c12SQLCandidate, c12SQLQueryRow, good, args, ledger) == nil {
			t.Fatal("wrong binding authorized")
		}
	}
	if err := c12AuthorizeSQL(c12SQLCandidate, c12SQLQueryRow, good, []any{id}, ledger); err != nil {
		t.Fatal(err)
	}
	if c12AuthorizeSQL(c12SQLCandidate, c12SQLExec, good, []any{id}, ledger) == nil {
		t.Fatal("wrong call kind")
	}
}

func c12TestDomainSeparatedDigests(t *testing.T) {
	f := c12PolicyFixture()
	op := f.operations[0]
	// Production canonicalAuthorityArtifact hashes domain || canonical, not raw
	// canonical bytes. The SQL capability validates boundaries, not that protocol.
	body := []byte(`{"fixture":"canonical-shape"}`)
	commitment := sha256.Sum256(append([]byte("talenro.nodecontrol.authority-effect-commitment.v1\x00"), body...))
	head := sha256.Sum256(append([]byte("talenro.nodecontrol.authority-provider-head.v1\x00"), body...))
	evidence := sha256.Sum256(append([]byte("talenro.nodecontrol.activation-decision-evidence.v1\x00"), body...))
	resolution := sha256.Sum256(append([]byte("talenro.nodecontrol.authority-effect-resolution.v1\x00"), body...))
	now := time.Date(2026, 9, 19, 1, 0, 0, 0, time.UTC)
	cases := []struct {
		name, sql string
		args      []any
	}{
		{"commitment", c12CommitmentWriteSQL, []any{op.certificateID, op.operationID, int64(31), int64(1), body, commitment[:]}},
		{"head_evidence_resolution", c12CertificateActivateSQL, []any{op.operationID, now, body, head[:], "none", now.Add(10 * time.Second), now.Add(10 * time.Second), op.scopeDigest[:], body, evidence[:], body, resolution[:]}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := c12AuthorizeSQL(c12SQLPrimary, c12SQLExec, tc.sql, tc.args, f); err != nil {
				t.Fatalf("production domain-separated digest rejected: %v", err)
			}
			for i, arg := range tc.args {
				if b, ok := arg.([]byte); ok {
					bad := append([]any(nil), tc.args...)
					if len(b) == 32 {
						bad[i] = make([]byte, 32)
					} else {
						bad[i] = make([]byte, 4097)
					}
					if c12AuthorizeSQL(c12SQLPrimary, c12SQLExec, tc.sql, bad, f) == nil {
						t.Fatal("artifact/digest field boundary widened")
					}
				}
			}
		})
	}
}

type c12InitRecorder struct {
	statements []string
	committed  bool
}
type c12InitConnector struct{ r *c12InitRecorder }

func (c c12InitConnector) Connect(context.Context) (driver.Conn, error) {
	return &c12InitConnection{r: c.r}, nil
}
func (c c12InitConnector) Driver() driver.Driver { return c12InitDriver{} }

type c12InitDriver struct{}

func (c12InitDriver) Open(string) (driver.Conn, error) { return nil, errors.New("unused") }

type c12InitConnection struct{ r *c12InitRecorder }

func (c *c12InitConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected Prepare")
}
func (c *c12InitConnection) Close() error              { return nil }
func (c *c12InitConnection) Begin() (driver.Tx, error) { return c, nil }
func (c *c12InitConnection) Commit() error             { c.r.committed = true; return nil }
func (c *c12InitConnection) Rollback() error           { return nil }
func (c *c12InitConnection) ExecContext(_ context.Context, q string, _ []driver.NamedValue) (driver.Result, error) {
	c.r.statements = append(c.r.statements, q)
	return driver.RowsAffected(1), nil
}
func (c *c12InitConnection) QueryContext(_ context.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
	if q != c12GetNodeControlDatabaseIdentitySQL {
		return nil, errors.New("unexpected private identity query")
	}
	return &c12InitIdentityRow{}, nil
}

type c12InitIdentityRow struct{ done bool }

func (*c12InitIdentityRow) Columns() []string {
	return []string{"system_id", "timeline", "required_lsn"}
}
func (*c12InitIdentityRow) Close() error { return nil }
func (r *c12InitIdentityRow) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	copy(dest, []driver.Value{"123", int64(7), "0/123"})
	return nil
}
func c12TestPrivateInitialization(t *testing.T) {
	recorder := &c12InitRecorder{}
	db := sql.OpenDB(c12InitConnector{recorder})
	defer db.Close()
	state := &c12AuthorityPITRState{database: db, descriptor: c12AuthorityPITRDescriptor{RunSuffix: "0123456789abcdef0123456789abcdef"}}
	if err := state.initializeC12SQLCapabilities(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !recorder.committed || state.fixture == nil || state.fixture.systemID != 123 || state.fixture.timeline != 7 {
		t.Fatal("private physical identity/init transaction missing")
	}
	role := `"talenro_c12_0123456789abcdef01234567_candidate"`
	if len(recorder.statements) != 10 {
		t.Fatalf("private statement count%d", len(recorder.statements))
	}
	if !strings.HasPrefix(recorder.statements[0], "CREATE ROLE "+role+" LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD '") {
		t.Fatal("candidate privileges not closed")
	}
	if recorder.statements[1] != c12MarkerCreateSQL || recorder.statements[2] != c12AuxDDLSQL {
		t.Fatal("private objects missing")
	}
	want := []string{
		`GRANT CONNECT ON DATABASE "talenro_c12_0123456789abcdef0123456789abcdef" TO ` + role,
		`GRANT USAGE ON SCHEMA nodecontrol,public TO ` + role,
		`GRANT SELECT ON nodecontrol.control_plane_authority_fences,nodecontrol.node_certificates,nodecontrol.authority_task7_crash_effects,nodecontrol.node_operator_audit,public.transactional_outbox TO ` + role,
		`GRANT UPDATE(operation_id) ON nodecontrol.control_plane_authority_fences,nodecontrol.authority_task7_crash_effects TO ` + role,
		`GRANT UPDATE(certificate_id) ON nodecontrol.node_certificates TO ` + role,
		`GRANT EXECUTE ON FUNCTION pg_catalog.pg_control_system(),pg_catalog.pg_control_checkpoint(),pg_catalog.pg_current_wal_insert_lsn() TO ` + role,
	}
	for i, q := range want {
		if recorder.statements[i+3] != q {
			t.Fatalf("grant mismatch%d", i)
		}
	}
	if recorder.statements[9] != c12RequiredObjectsSQL {
		t.Fatal("required object check missing")
	}
	if len(state.candidatePassword) != 64 {
		t.Fatal("candidate password not randomized")
	}
	recorder.statements = nil
	if err := state.closeC12Setup(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !state.setupClosed || len(recorder.statements) != 1 || recorder.statements[0] != c12RequiredObjectsSQL {
		t.Fatal("backup did not close setup and verify required objects")
	}
}

func c12TestPolicyAccessLedger(t *testing.T) {
	for _, terminal := range []string{"commit", "rollback", "uncertain"} {
		t.Run(terminal, func(t *testing.T) {
			driver := &c12AccessTestFixture{}
			state := newC12AccessTestState(driver)
			state.accessPolicy = nil
			f := c12PolicyFixture()
			op := f.operations[0]
			f.operations = nil
			cert := f.certificates[op.certificateID]
			cert.issuanceID = uuid.NewSHA1(op.operationID, []byte("issuance"))
			cert.historyOperationID = uuid.NewSHA1(op.operationID, []byte("issuance-operation"))
			cert.lineageID = uuid.NewSHA1(op.operationID, []byte("lineage"))
			cert.attemptID = uuid.NewSHA1(op.operationID, []byte("attempt"))
			f.certificates[op.certificateID] = cert
			state.fixture = f
			if terminal == "uncertain" {
				driver.commitErr = errors.New("connection lost")
			}
			err := c12WithAccess(context.Background(), state, nil, func(a C12AuthorityAccess) error {
				if _, err := a.Exec(context.Background(), c12ReplicaSQL); err == nil {
					t.Fatal("direct setup write authorized")
				}
				tx, err := a.Begin(context.Background())
				if err != nil {
					return err
				}
				digest := sha256.Sum256([]byte("reservation"))
				args := []any{op.operationID, "certificate_revoke", "node", int64(31), int64(1), op.scopeDigest[:], digest[:], pgtype.Timestamptz{Time: time.Date(2026, 9, 19, 1, 0, 0, 0, time.UTC), Valid: true}, uuid.NullUUID{UUID: f.activationID, Valid: true}}
				if _, err = tx.Exec(context.Background(), c12InsertClaimV1AuthorityFencePendingSQL, args...); err != nil {
					return err
				}
				if len(state.fixture.operations) != 0 {
					t.Fatal("uncommitted identity escaped")
				}
				if _, err = tx.Query(context.Background(), c12LockAuthorityFenceSQL, op.operationID); err == nil {
					t.Fatal("Query acquired QueryRow authority")
				}
				row := tx.QueryRow(context.Background(), c12LockAuthorityFenceSQL, op.operationID)
				if _, bad := row.(c12ErrorRow); bad {
					t.Fatal("same-transaction identity not visible")
				}
				if _, err = a.Query(context.Background(), c12AuxResolveSQL, op.operationID); err == nil {
					t.Fatal("pending identity visible outside transaction")
				}
				if terminal == "rollback" {
					return tx.Rollback(context.Background())
				}
				err = tx.Commit(context.Background())
				if terminal == "uncertain" && !errors.Is(err, C12PITRIndeterminate) {
					t.Fatalf("uncertain result %v", err)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			want := 0
			if terminal == "commit" {
				want = 1
			}
			if len(state.fixture.operations) != want {
				t.Fatalf("ledger operations=%d want%d", len(state.fixture.operations), want)
			}
			if terminal == "uncertain" && !state.writesUncertain {
				t.Fatal("uncertain commit did not fail closed")
			}
			if terminal == "uncertain" {
				err = c12WithAccess(context.Background(), state, nil, func(a C12AuthorityAccess) error {
					if row := a.QueryRow(context.Background(), c12GetNodeControlDatabaseIdentitySQL); row == nil {
						t.Fatal("independent read missing")
					}
					tx, err := a.Begin(context.Background())
					if err != nil {
						return err
					}
					if _, err := tx.Exec(context.Background(), c12ReplicaSQL); err == nil {
						t.Fatal("write after uncertain commit reached driver")
					}
					return tx.Rollback(context.Background())
				})
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

type c12PolicyCaptureDriver struct {
	*c12AccessTestDriver
	onExec func(string, []any)
}

func (d *c12PolicyCaptureDriver) BeginTx(ctx context.Context, options pgx.TxOptions) (c12AccessDriver, error) {
	_, err := d.c12AccessTestDriver.BeginTx(ctx, options)
	return d, err
}
func (d *c12PolicyCaptureDriver) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	if d.onExec != nil {
		d.onExec(query, args)
	}
	return d.c12AccessTestDriver.Exec(ctx, query, args...)
}

type c12PolicyValuer struct{ called *bool }

func (v c12PolicyValuer) Value() (driver.Value, error) { *v.called = true; return "0/123", nil }
func c12TestAdapterCopies(t *testing.T) {
	d := &c12AccessTestFixture{}
	state := newC12AccessTestState(d)
	state.accessPolicy = nil
	state.fixture = c12PolicyFixture()
	state.fixture.systemID = 123
	state.fixture.timeline = 7
	op := &state.fixture.operations[0]
	op.effectDigest = sha256.Sum256([]byte("effect"))
	effect := append([]byte(nil), op.effectDigest[:]...)
	number := big.NewInt(123)
	args := []any{op.operationID, effect, pgtype.Numeric{Int: number, Valid: true}, pgtype.Int8{Int64: 7, Valid: true}, "0/123", pgtype.Timestamptz{Time: time.Date(2026, 9, 19, 1, 0, 0, 0, time.UTC), Valid: true}}
	reached := false
	state.accessOpen = func(context.Context, *c12CandidateBinding) (c12AccessBackend, error) {
		return &c12PolicyCaptureDriver{c12AccessTestDriver: &c12AccessTestDriver{fixture: d}, onExec: func(query string, received []any) {
			if query != c12BindAuthorityFenceEffectSQL {
				t.Fatal("unexpected execution")
			}
			reached = true
			args[0] = uuid.Nil
			effect[0] ^= 0xff
			number.SetInt64(99)
			if received[0] != op.operationID || !bytes.Equal(received[1].([]byte), op.effectDigest[:]) || received[2].(pgtype.Numeric).Int.Uint64() != 123 {
				t.Fatal("adapter retained caller mutable arguments")
			}
		}}, nil
	}
	err := c12WithAccess(context.Background(), state, nil, func(a C12AuthorityAccess) error {
		tx, err := a.Begin(context.Background())
		if err != nil {
			return err
		}
		if _, err := tx.Exec(context.Background(), c12BindAuthorityFenceEffectSQL, args...); err != nil {
			return err
		}
		return tx.Commit(context.Background())
	})
	if err != nil || !reached {
		t.Fatalf("copy execution reached=%v err=%v", reached, err)
	}
	called := false
	for _, bad := range []any{c12PolicyValuer{&called}, make([]byte, (1<<20)+1), pgtype.Numeric{Int: new(big.Int).Lsh(big.NewInt(1), 100), Valid: true}, map[string]string{"lsn": "0/123"}} {
		if _, err := c12CopySQLArguments([]any{bad}); err == nil {
			t.Fatal("unbounded/custom codec argument")
		}
	}
	if called {
		t.Fatal("caller Valuer invoked")
	}
}

func c12TestAllSQLArguments(t *testing.T) {
	f := c12PolicyFixture()
	f.systemID = 123
	f.timeline = 7
	op := &f.operations[0]
	cert := f.certificates[op.certificateID]
	cert.issuanceID = uuid.NewSHA1(op.operationID, []byte("issuance"))
	cert.historyOperationID = uuid.NewSHA1(op.operationID, []byte("issuance-operation"))
	cert.lineageID = uuid.NewSHA1(op.operationID, []byte("lineage"))
	cert.attemptID = uuid.NewSHA1(op.operationID, []byte("attempt"))
	f.certificates[op.certificateID] = cert
	op.reservationDigest = sha256.Sum256([]byte("reservation"))
	body := []byte(`{"fixed":"body"}`)
	op.effectDigest = sha256.Sum256(body)
	op.receiptDigest = sha256.Sum256([]byte("receipt"))
	op.requiredLSN = "0/123"
	now := time.Date(2026, 9, 19, 1, 0, 0, 0, time.UTC)
	ts := pgtype.Timestamptz{Time: now, Valid: true}
	sys := pgtype.Numeric{Int: big.NewInt(123), Valid: true}
	tl := pgtype.Int8{Int64: 7, Valid: true}
	cases := map[string][]any{
		c12GetNodeControlDatabaseIdentitySQL: {}, c12GetAuthorityFenceHeadSQL: {}, c12ListPendingAuthorityFencesSQL: {int64(31)},
		c12LockAuthorityFenceSQL: {op.operationID}, c12LockCertificateRevocationOutcomeSQL: {uuid.NullUUID{UUID: op.operationID, Valid: true}}, c12GetStoredAuthorityFenceSQL: {op.operationID}, c12GetAuthorityFenceForUpdateSQL: {op.operationID},
		c12InsertClaimV1AuthorityFencePendingSQL: {op.operationID, "certificate_revoke", "node", int64(31), int64(1), op.scopeDigest[:], op.reservationDigest[:], ts, uuid.NullUUID{UUID: f.activationID, Valid: true}},
		c12BindAuthorityFenceEffectSQL:           {op.operationID, op.effectDigest[:], sys, tl, "0/123", ts},
		c12ActivateCommittedAuthorityFenceSQL:    {op.operationID, op.receiptDigest[:], ts, op.effectDigest[:], sys, tl, "0/123"},
		c12AuxResolveSQL:                         {op.operationID}, c12CertificateResolutionSQL: {op.operationID}, c12CertificateInputSQL: {op.operationID}, c12PersistedEffectSQL: {op.operationID}, c12AuditOutboxCountsSQL: {op.operationID}, c12OutboxPayloadSQL: {op.operationID, op.nodeID, int64(1)}, c12CommitmentReadSQL: {op.operationID},
		c12CommitmentWriteSQL:      {op.certificateID, op.operationID, int64(31), int64(1), body, op.effectDigest[:]},
		c12AuxInsertSQL:            {op.operationID, "certificate_revoke", "node", op.scopeDigest[:], op.effectDigest[:], "committed"},
		c12CertificateActivateSQL:  {op.operationID, now, body, op.effectDigest[:], "none", now.Add(10 * time.Second), now.Add(10 * time.Second), op.scopeDigest[:], body, op.effectDigest[:], body, op.effectDigest[:]},
		c12AuditInsertSQL:          {op.operationID, int64(31), int64(1), op.receiptDigest[:], op.nodeID.String(), now, now.Add(181 * 24 * time.Hour)},
		c12OutboxInsertSQL:         {op.operationID, op.nodeID, int64(1), op.operationID.String(), body, now},
		pitrCertificateSnapshotSQL: {op.certificateID}, pitrOperationSnapshotSQL: {op.operationID}, pitrAuxCountSQL: {op.operationID},
	}
	for query, args := range cases {
		rule := c12AuthoritySQLRules[query]
		if err := c12AuthorizeSQL(c12SQLPrimary, rule.call, query, args, f); err != nil {
			t.Fatalf("positive statement %q: %v", query, err)
		}
		for _, mode := range []c12SQLMode{0, 8, c12SQLCandidate | c12SQLPrimary, c12SQLSetup | c12SQLCandidate} {
			if c12AuthorizeSQL(mode, rule.call, query, args, f) == nil {
				t.Fatal("invalid mode")
			}
		}
		for call := c12SQLExec; call <= c12SQLQueryRow; call++ {
			if call != rule.call && c12AuthorizeSQL(c12SQLPrimary, call, query, args, f) == nil {
				t.Fatal("wrong call")
			}
		}
		if rule.call == c12SQLExec && c12AuthorizeSQL(c12SQLCandidate, rule.call, query, args, f) == nil {
			t.Fatal("candidate write")
		}
		if c12AuthorizeSQL(c12SQLPrimary, rule.call, query, append(append([]any(nil), args...), uuid.Nil), f) == nil {
			t.Fatal("extra argument")
		}
		for i := range args {
			bad := append([]any(nil), args...)
			bad[i] = struct{}{}
			if c12AuthorizeSQL(c12SQLPrimary, rule.call, query, bad, f) == nil {
				t.Fatalf("wrong concrete type arg%d query%s", i, query)
			}
			if _, ok := args[i].(uuid.UUID); ok {
				bad[i] = uuid.MustParse("99999999-0000-4000-8000-000000000001")
				if c12AuthorizeSQL(c12SQLPrimary, rule.call, query, bad, f) == nil {
					t.Fatal("foreign fixture identity")
				}
			}
		}
	}
	for _, bad := range []any{pgtype.Numeric{Int: big.NewInt(123), Valid: false}, pgtype.Numeric{Int: big.NewInt(123), Exp: 1, Valid: true}, pgtype.Numeric{Int: big.NewInt(123), NaN: true, Valid: true}, pgtype.Numeric{Int: big.NewInt(123), InfinityModifier: pgtype.Infinity, Valid: true}, pgtype.Numeric{Int: big.NewInt(124), Valid: true}} {
		args := append([]any(nil), cases[c12BindAuthorityFenceEffectSQL]...)
		args[2] = bad
		if c12AuthorizeSQL(c12SQLPrimary, c12SQLExec, c12BindAuthorityFenceEffectSQL, args, f) == nil {
			t.Fatal("nonphysical numeric")
		}
	}
	for _, bad := range []any{pgtype.Int8{Int64: 7}, pgtype.Int8{Int64: 0, Valid: true}, pgtype.Int8{Int64: 8, Valid: true}} {
		args := append([]any(nil), cases[c12BindAuthorityFenceEffectSQL]...)
		args[3] = bad
		if c12AuthorizeSQL(c12SQLPrimary, c12SQLExec, c12BindAuthorityFenceEffectSQL, args, f) == nil {
			t.Fatal("foreign timeline")
		}
	}
	for _, bad := range []any{" 0/123", "0/123 ", "0/0", "0/100000000", uint64(123)} {
		args := append([]any(nil), cases[c12BindAuthorityFenceEffectSQL]...)
		args[4] = bad
		if c12AuthorizeSQL(c12SQLPrimary, c12SQLExec, c12BindAuthorityFenceEffectSQL, args, f) == nil {
			t.Fatal("invalid WAL value")
		}
	}
	for _, query := range []string{c12AuxDDLSQL, c12MarkerInsertSQL, c12MarkerCreateSQL, "SET ROLE talenro", "DELETE FROM nodecontrol.node_certificates"} {
		if c12AuthorizeSQL(c12SQLPrimary|c12SQLSetup, c12SQLExec, query, nil, f) == nil {
			t.Fatal("private/arbitrary statement authorized")
		}
	}
}

type c12PolicyUUIDRows struct{ *c12AccessTestRows }

func (*c12PolicyUUIDRows) FieldDescriptions() []pgconn.FieldDescription {
	return []pgconn.FieldDescription{{Name: "installation_id", DataTypeOID: pgtype.UUIDOID, Format: pgx.BinaryFormatCode}}
}
func c12TestSetupLifecycle(t *testing.T) {
	for _, commit := range []bool{false, true} {
		t.Run(fmt.Sprintf("commit_%v", commit), func(t *testing.T) {
			d := &c12AccessTestFixture{}
			state := newC12AccessTestState(d)
			state.accessPolicy = nil
			state.fixture = &c12FixtureLedger{epoch: 31, systemID: 123, timeline: 7, certificates: make(map[uuid.UUID]c12FixtureCertificate)}
			at := time.Date(2026, 9, 19, 1, 0, 0, 0, time.UTC)
			activation := uuid.MustParse("79000000-0000-4000-8000-000000000001")
			installation := uuid.MustParse("79000000-0000-4000-8000-000000000002")
			id := func(label string) uuid.UUID {
				return uuid.NewSHA1(uuid.NameSpaceOID, []byte("task8-proof-"+label+":79"))
			}
			err := c12WithAccess(context.Background(), state, nil, func(a C12AuthorityAccess) error {
				tx, err := a.Begin(context.Background())
				if err != nil {
					return err
				}
				// Supply the actual fixed binary UUID returned by the installed-latch SELECT.
				d.rows = func() pgx.Rows {
					return &c12PolicyUUIDRows{&c12AccessTestRows{fixture: d, values: [][][]byte{{installation[:]}}}}
				}
				var got uuid.UUID
				if err := tx.QueryRow(context.Background(), c12LatchSQL).Scan(&got); err != nil {
					return err
				}
				if got != installation {
					t.Fatal("latch result")
				}
				steps := []struct {
					sql  string
					args []any
				}{
					{c12ReplicaSQL, nil}, {c12UpgradeIntentSQL, []any{id("intent"), installation, activation, id("incarnation"), id("absence"), id("deployment"), at}},
					{c12RuntimeRegistrationSQL, []any{id("registration"), activation, id("runtime"), at}},
					{c12UpgradeAttemptSQL, []any{id("attempt"), activation, id("preparation"), id("completion"), id("release"), id("open"), id("deployment"), at}},
					{c12ProtocolActivationSQL, []any{activation, id("deployment"), at}},
					{c12ActivationCompletionSQL, []any{id("completion"), activation, at}},
					{c12ActivationReleaseSQL, []any{id("release"), id("open"), activation, at}}, {c12OriginSQL, nil},
				}
				for _, step := range steps {
					if c12AuthorizeSQL(c12SQLCandidate, c12SQLExec, step.sql, step.args, tx.(*c12AuthorityTx).fixture) == nil {
						t.Fatal("candidate setup write authorized")
					}
					for i, arg := range step.args {
						bad := append([]any(nil), step.args...)
						bad[i] = struct{}{}
						if c12AuthorizeSQL(c12SQLPrimary|c12SQLSetup, c12SQLExec, step.sql, bad, tx.(*c12AuthorityTx).fixture) == nil {
							t.Fatal("setup argument type")
						}
						if _, ok := arg.(uuid.UUID); ok {
							bad[i] = uuid.MustParse("99000000-0000-4000-8000-000000000001")
							if c12AuthorizeSQL(c12SQLPrimary|c12SQLSetup, c12SQLExec, step.sql, bad, tx.(*c12AuthorityTx).fixture) == nil {
								t.Fatal("foreign setup dependency")
							}
						}
					}
					if _, err := tx.Exec(context.Background(), step.sql, step.args...); err != nil {
						return err
					}
				}
				if state.fixture.activationID != uuid.Nil {
					t.Fatal("setup escaped before commit")
				}
				if commit {
					return tx.Commit(context.Background())
				}
				return tx.Rollback(context.Background())
			})
			if err != nil {
				t.Fatal(err)
			}
			if (state.fixture.activationID == activation) != commit {
				t.Fatal("setup commit binding")
			}
			state.setupClosed = true
			err = c12WithAccess(context.Background(), state, nil, func(a C12AuthorityAccess) error {
				tx, err := a.Begin(context.Background())
				if err != nil {
					return err
				}
				if _, err := tx.Exec(context.Background(), c12ReplicaSQL); err == nil {
					t.Fatal("backup allowed replica")
				}
				if _, err := tx.Exec(context.Background(), c12AuxDDLSQL); err == nil {
					t.Fatal("backup allowed DDL")
				}
				return tx.Rollback(context.Background())
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func c12HistoricalSteps(op, node uuid.UUID, seq int64, at time.Time) []struct {
	sql  string
	args []any
} {
	id := func(label string) uuid.UUID { return uuid.NewSHA1(op, []byte(label)) }
	scope := sha256.Sum256(append([]byte("TALENRO-NODE-AUTHORITY-SCOPE-V1\x00"), node[:]...))
	return []struct {
		sql  string
		args []any
	}{
		{c12ReplicaSQL, nil}, {c12NodePopSQL, []any{at}},
		{c12NodeInventorySQL, []any{node, id("lineage"), at}},
		{c12LegacyFenceSQL, []any{id("issuance-operation"), seq, scope[:], "123", int64(7), "0/123", at}},
		{c12IssuanceSQL, []any{id("issuance"), id("issuance-operation"), seq, node, id("attempt"), id("lineage"), scope[:], at, at.Add(time.Hour), at.Add(366 * 24 * time.Hour)}},
		{c12OriginSQL, nil},
		{c12CertificateSeedSQL, []any{id("certificate"), id("issuance"), id("issuance-operation"), seq, node, id("lineage"), scope[:], at, at.Add(time.Hour), at.Add(366*24*time.Hour + time.Second)}},
	}
}
func c12TestHistoricalGroups(t *testing.T) {
	at := time.Date(2026, 9, 19, 1, 0, 0, 0, time.UTC)
	f := c12PolicyFixture()
	f.systemID = 123
	f.timeline = 7
	f.operations = nil
	f.certificates = make(map[uuid.UUID]c12FixtureCertificate)
	op := uuid.MustParse("81000000-0000-4000-8000-000000000001")
	node := uuid.MustParse("81000000-0000-4000-8000-000000000002")
	steps := c12HistoricalSteps(op, node, 1, at)
	for _, step := range steps {
		if c12AuthorizeSQL(c12SQLCandidate, c12SQLExec, step.sql, step.args, f) == nil {
			t.Fatal("candidate historical write authorized")
		}
		if err := c12AuthorizeSQL(c12SQLPrimary|c12SQLSetup, c12SQLExec, step.sql, step.args, f); err != nil {
			t.Fatalf("historical step %s: %v", step.sql, err)
		}
		for i := range step.args {
			bad := append([]any(nil), step.args...)
			bad[i] = struct{}{}
			if c12AuthorizeSQL(c12SQLPrimary|c12SQLSetup, c12SQLExec, step.sql, bad, f) == nil {
				t.Fatalf("historical arg type%d", i)
			}
		}
		f.applied(step.sql, step.args, 1)
	}
	certID := uuid.NewSHA1(op, []byte("certificate"))
	cert := f.certificates[certID]
	if cert.attemptID != uuid.NewSHA1(op, []byte("attempt")) {
		t.Fatal("historical attempt binding lost")
	}
	scope := c12Scope(node)
	digest := sha256.Sum256([]byte("reservation"))
	claim := []any{op, "certificate_revoke", "node", int64(31), int64(1), scope[:], digest[:], pgtype.Timestamptz{Time: at, Valid: true}, uuid.NullUUID{UUID: f.activationID, Valid: true}}
	if err := c12AuthorizeSQL(c12SQLPrimary, c12SQLExec, c12InsertClaimV1AuthorityFencePendingSQL, claim, f); err != nil {
		t.Fatal(err)
	}
	replica := f.clone()
	replica.setup.replica = true
	if c12AuthorizeSQL(c12SQLPrimary|c12SQLSetup, c12SQLExec, c12InsertClaimV1AuthorityFencePendingSQL, claim, replica) == nil {
		t.Fatal("current claim bypassed triggers in replica mode")
	}
	for _, field := range []string{"lineage", "attempt", "issuance", "history"} {
		bad := f.clone()
		c := bad.certificates[certID]
		foreign := uuid.MustParse("99000000-0000-4000-8000-000000000001")
		switch field {
		case "lineage":
			c.lineageID = foreign
		case "attempt":
			c.attemptID = foreign
		case "issuance":
			c.issuanceID = foreign
		case "history":
			c.historyOperationID = foreign
		}
		bad.certificates[certID] = c
		if c12AuthorizeSQL(c12SQLPrimary, c12SQLExec, c12InsertClaimV1AuthorityFencePendingSQL, claim, bad) == nil {
			t.Fatalf("late-binding mismatch %s", field)
		}
	}
	cloned := f.clone()
	cloned.applied(c12InsertClaimV1AuthorityFencePendingSQL, claim, 0)
	if len(cloned.operations) != 0 {
		t.Fatal("zero-row insert minted identity")
	}
	cloned.applied(c12InsertClaimV1AuthorityFencePendingSQL, claim, 1)
	if len(f.operations) != 0 || len(cloned.operations) != 1 {
		t.Fatal("ledger clone alias")
	}
	delete(cloned.certificates, certID)
	if len(f.certificates) != 1 {
		t.Fatal("ledger map alias")
	}
	for seq := int64(2); seq <= 3; seq++ {
		id := uuid.NewSHA1(op, []byte(fmt.Sprint(seq)))
		n := uuid.NewSHA1(node, []byte(fmt.Sprint(seq)))
		for _, step := range c12HistoricalSteps(id, n, seq, at) {
			if seq == 2 {
				reused := append([]any(nil), step.args...)
				switch step.sql {
				case c12NodeInventorySQL:
					reused[1] = cert.lineageID
				case c12LegacyFenceSQL:
					reused[0] = cert.historyOperationID
				case c12IssuanceSQL:
					reused[0] = cert.issuanceID
				}
				if step.sql == c12NodeInventorySQL || step.sql == c12LegacyFenceSQL || step.sql == c12IssuanceSQL {
					if c12AuthorizeSQL(c12SQLPrimary|c12SQLSetup, c12SQLExec, step.sql, reused, f) == nil {
						t.Fatal("historical identity reused across groups")
					}
				}
				if step.sql == c12IssuanceSQL {
					reused = append([]any(nil), step.args...)
					reused[4] = cert.attemptID
					if c12AuthorizeSQL(c12SQLPrimary|c12SQLSetup, c12SQLExec, step.sql, reused, f) == nil {
						t.Fatal("historical attempt reused")
					}
				}
			}
			if err := c12AuthorizeSQL(c12SQLPrimary|c12SQLSetup, c12SQLExec, step.sql, step.args, f); err != nil {
				t.Fatalf("group%d: %v", seq, err)
			}
			affected := int64(1)
			if step.sql == c12NodePopSQL {
				affected = 0
			}
			f.applied(step.sql, step.args, affected)
		}
	}
	if len(f.certificates) != 3 {
		t.Fatal("three historical groups not retained")
	}
	f.applied(c12ReplicaSQL, nil, 0)
	if c12AuthorizeSQL(c12SQLPrimary|c12SQLSetup, c12SQLExec, c12NodePopSQL, []any{at}, f) == nil {
		t.Fatal("fourth historical group")
	}
}

func c12TestOriginalCallKind(t *testing.T) {
	for _, candidate := range []bool{false, true} {
		t.Run(fmt.Sprint(candidate), func(t *testing.T) {
			d := &c12AccessTestFixture{}
			state := newC12AccessTestState(d)
			state.accessPolicy = nil
			state.fixture = c12PolicyFixture()
			var binding *c12CandidateBinding
			if candidate {
				controller, fixture := newC12CutFixture(t)
				observed := fixture.observed("0/20", 7)
				selection, err := controller.SelectRecoveryCut(controller.state.baseBackup, observed)
				if err != nil {
					t.Fatal(err)
				}
				cut, err := controller.CrashPrimaryAtCut(context.Background(), selection, observed)
				if err != nil {
					t.Fatal(err)
				}
				dto, err := controller.RestoreAtCut(context.Background(), cut)
				if err != nil {
					t.Fatal(err)
				}
				dto, err = controller.PromoteCandidate(context.Background(), dto)
				if err != nil {
					t.Fatal(err)
				}
				state = controller.state
				state.fixture = c12PolicyFixture()
				state.candidateRole = "candidate_role"
				state.accessOpen = func(context.Context, *c12CandidateBinding) (c12AccessBackend, error) {
					d.event("open")
					return &c12PrivilegeBackend{c12AccessTestDriver: &c12AccessTestDriver{fixture: d}, values: []any{"candidate_role", "candidate_role", false, "off", true, true, false, true}}, nil
				}
				binding = &dto.binding
			}
			err := c12WithAccess(context.Background(), state, binding, func(a C12AuthorityAccess) error {
				for _, db := range []C12AuthorityAccess{a} {
					before := d.snapshot()
					if _, err := db.Query(context.Background(), pitrAuxCountSQL, state.fixture.operations[0].operationID); err == nil {
						t.Fatal("access Query admitted QueryRow")
					}
					row := db.QueryRow(context.Background(), c12AuxResolveSQL, state.fixture.operations[0].operationID)
					if _, ok := row.(c12ErrorRow); !ok {
						t.Fatal("access QueryRow admitted Query")
					}
					if !reflect.DeepEqual(before, d.snapshot()) {
						t.Fatal("authorization waited for Scan/driver")
					}
					row = db.QueryRow(context.Background(), pitrAuxCountSQL, state.fixture.operations[0].operationID)
					if _, ok := row.(c12ErrorRow); ok {
						t.Fatal("valid access QueryRow denied")
					}
				}
				tx, err := a.Begin(context.Background())
				if err != nil {
					return err
				}
				before := d.snapshot()
				if _, err := tx.Query(context.Background(), pitrAuxCountSQL, state.fixture.operations[0].operationID); err == nil {
					t.Fatal("tx Query admitted QueryRow")
				}
				row := tx.QueryRow(context.Background(), c12AuxResolveSQL, state.fixture.operations[0].operationID)
				if _, ok := row.(c12ErrorRow); !ok {
					t.Fatal("tx QueryRow admitted Query")
				}
				if !reflect.DeepEqual(before, d.snapshot()) {
					t.Fatal("invalid transaction call reached driver")
				}
				row = tx.QueryRow(context.Background(), pitrAuxCountSQL, state.fixture.operations[0].operationID)
				if _, ok := row.(c12ErrorRow); ok {
					t.Fatal("valid tx QueryRow denied")
				}
				if candidate {
					before = d.snapshot()
					for query, rule := range c12AuthoritySQLRules {
						if rule.call == c12SQLExec {
							if _, err := tx.Exec(context.Background(), query, make([]any, rule.arity)...); err == nil {
								t.Fatal("candidate write reached driver")
							}
						}
					}
					if !reflect.DeepEqual(before, d.snapshot()) {
						t.Fatal("candidate write reached driver")
					}
				}
				return tx.Rollback(context.Background())
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
