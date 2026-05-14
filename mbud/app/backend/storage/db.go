package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func Open(dataDir string) (*sql.DB, error) {
	inMemory := dataDir == ":memory:"
	var dsn string
	if inMemory {
		dsn = ":memory:"
	} else {
		if err := os.MkdirAll(dataDir, 0o755); err != nil {
			return nil, err
		}
		dsn = filepath.Join(dataDir, "mbud.db")
	}
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA busy_timeout = 5000",
	}
	if !inMemory {
		// WAL needed because the host's PluginDBManager may keep its own *sql.DB
		// open against the same file; without WAL we hit "database is locked".
		pragmas = append(pragmas, "PRAGMA journal_mode = WAL")
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("%s: %w", p, err)
		}
	}
	if err := InitSchema(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func InitSchema(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS businesses (
			id          TEXT    PRIMARY KEY,
			name        TEXT    NOT NULL,
			tax_id      TEXT    NOT NULL DEFAULT '',
			email       TEXT    NOT NULL DEFAULT '',
			address     TEXT    NOT NULL DEFAULT '',
			notes       TEXT    NOT NULL DEFAULT '',
			logo_type   TEXT    NOT NULL DEFAULT '',
			created_at  INTEGER NOT NULL,
			updated_at  INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS invoices (
			id           TEXT    PRIMARY KEY,
			business_id  TEXT    REFERENCES businesses(id) ON DELETE CASCADE,
			amount       REAL    NOT NULL,
			currency     TEXT    NOT NULL,
			description  TEXT    NOT NULL DEFAULT '',
			issued_at    INTEGER NOT NULL,
			due_at       INTEGER NOT NULL,
			paid         INTEGER NOT NULL DEFAULT 0,
			paid_at      INTEGER NOT NULL DEFAULT 0,
			created_at   INTEGER NOT NULL,
			updated_at   INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_invoices_business_id ON invoices(business_id)`,
		`CREATE INDEX IF NOT EXISTS idx_invoices_due_at ON invoices(due_at)`,
		`CREATE INDEX IF NOT EXISTS idx_invoices_paid_due_at ON invoices(paid, due_at)`,
		`CREATE TABLE IF NOT EXISTS recurring_invoices (
			id                  TEXT    PRIMARY KEY,
			business_id         TEXT    REFERENCES businesses(id) ON DELETE CASCADE,
			amount              REAL    NOT NULL,
			currency            TEXT    NOT NULL,
			description         TEXT    NOT NULL DEFAULT '',
			frequency           TEXT    NOT NULL,
			start_at            INTEGER NOT NULL,
			end_at              INTEGER NOT NULL DEFAULT 0,
			active              INTEGER NOT NULL DEFAULT 1,
			issue_day_of_week   INTEGER NOT NULL DEFAULT 0,
			issue_day_of_month  INTEGER NOT NULL DEFAULT 0,
			issue_month_of_year INTEGER NOT NULL DEFAULT 0,
			created_at          INTEGER NOT NULL,
			updated_at          INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_recurring_business_id ON recurring_invoices(business_id)`,
		`CREATE INDEX IF NOT EXISTS idx_recurring_active ON recurring_invoices(active)`,
		`CREATE TABLE IF NOT EXISTS upcoming_invoices (
			id           TEXT    PRIMARY KEY,
			business_id  TEXT    REFERENCES businesses(id) ON DELETE CASCADE,
			amount       REAL    NOT NULL,
			currency     TEXT    NOT NULL,
			description  TEXT    NOT NULL DEFAULT '',
			due_at       INTEGER NOT NULL,
			created_at   INTEGER NOT NULL,
			updated_at   INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_upcoming_business_id ON upcoming_invoices(business_id)`,
		`CREATE INDEX IF NOT EXISTS idx_upcoming_due_at ON upcoming_invoices(due_at)`,
		// No FK on invoice_id: when a user deletes a generated invoice, the link row
		// must survive so the reconciler treats the period as already emitted.
		`CREATE TABLE IF NOT EXISTS recurring_invoice_links (
			invoice_id   TEXT    NOT NULL,
			recurring_id TEXT    NOT NULL,
			period_index INTEGER NOT NULL,
			PRIMARY KEY (invoice_id, recurring_id),
			FOREIGN KEY (recurring_id) REFERENCES recurring_invoices(id) ON DELETE CASCADE
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_links_rule_period ON recurring_invoice_links(recurring_id, period_index)`,
		`CREATE INDEX IF NOT EXISTS idx_links_invoice_id ON recurring_invoice_links(invoice_id)`,
		// No FK on invoice_id: when a user deletes a materialised invoice, the link row
		// must survive so the reconciler treats this upcoming as already emitted.
		`CREATE TABLE IF NOT EXISTS upcoming_invoice_links (
			invoice_id  TEXT NOT NULL,
			upcoming_id TEXT NOT NULL,
			PRIMARY KEY (invoice_id, upcoming_id),
			FOREIGN KEY (upcoming_id) REFERENCES upcoming_invoices(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_upcoming_links_upcoming_id ON upcoming_invoice_links(upcoming_id)`,
		`CREATE INDEX IF NOT EXISTS idx_upcoming_links_invoice_id ON upcoming_invoice_links(invoice_id)`,
		`CREATE TABLE IF NOT EXISTS users (
			id          TEXT    PRIMARY KEY,
			name        TEXT    NOT NULL,
			email       TEXT    NOT NULL DEFAULT '',
			notes       TEXT    NOT NULL DEFAULT '',
			created_at  INTEGER NOT NULL,
			updated_at  INTEGER NOT NULL
		)`,
		// Both FKs cascade — unlike recurring/upcoming link tables, an invoice-user
		// link has no value once either side is deleted.
		`CREATE TABLE IF NOT EXISTS invoice_users (
			invoice_id  TEXT NOT NULL,
			user_id     TEXT NOT NULL,
			PRIMARY KEY (invoice_id, user_id),
			FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id)    REFERENCES users(id)    ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_invoice_users_user_id    ON invoice_users(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_invoice_users_invoice_id ON invoice_users(invoice_id)`,
		`CREATE TABLE IF NOT EXISTS recurring_users (
			recurring_id TEXT NOT NULL,
			user_id      TEXT NOT NULL,
			PRIMARY KEY (recurring_id, user_id),
			FOREIGN KEY (recurring_id) REFERENCES recurring_invoices(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id)      REFERENCES users(id)              ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_recurring_users_user_id ON recurring_users(user_id)`,
		`CREATE TABLE IF NOT EXISTS upcoming_users (
			upcoming_id TEXT NOT NULL,
			user_id     TEXT NOT NULL,
			PRIMARY KEY (upcoming_id, user_id),
			FOREIGN KEY (upcoming_id) REFERENCES upcoming_invoices(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id)     REFERENCES users(id)             ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_upcoming_users_user_id ON upcoming_users(user_id)`,
		`CREATE TABLE IF NOT EXISTS tags (
			id          TEXT    PRIMARY KEY,
			name        TEXT    NOT NULL UNIQUE,
			created_at  INTEGER NOT NULL,
			updated_at  INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS business_tags (
			business_id  TEXT NOT NULL,
			tag_id       TEXT NOT NULL,
			PRIMARY KEY (business_id, tag_id),
			FOREIGN KEY (business_id) REFERENCES businesses(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id)      REFERENCES tags(id)       ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_business_tags_tag_id ON business_tags(tag_id)`,
		`CREATE TABLE IF NOT EXISTS invoice_tags (
			invoice_id  TEXT NOT NULL,
			tag_id      TEXT NOT NULL,
			PRIMARY KEY (invoice_id, tag_id),
			FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id)     REFERENCES tags(id)     ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_invoice_tags_tag_id ON invoice_tags(tag_id)`,
		`CREATE TABLE IF NOT EXISTS recurring_tags (
			recurring_id TEXT NOT NULL,
			tag_id       TEXT NOT NULL,
			PRIMARY KEY (recurring_id, tag_id),
			FOREIGN KEY (recurring_id) REFERENCES recurring_invoices(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id)       REFERENCES tags(id)                ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_recurring_tags_tag_id ON recurring_tags(tag_id)`,
		`CREATE TABLE IF NOT EXISTS upcoming_tags (
			upcoming_id TEXT NOT NULL,
			tag_id      TEXT NOT NULL,
			PRIMARY KEY (upcoming_id, tag_id),
			FOREIGN KEY (upcoming_id) REFERENCES upcoming_invoices(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id)      REFERENCES tags(id)              ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_upcoming_tags_tag_id ON upcoming_tags(tag_id)`,
		`CREATE TABLE IF NOT EXISTS upload_sessions (
			id          TEXT    PRIMARY KEY,
			invoice_id  TEXT,
			status      TEXT    NOT NULL,
			created_at  INTEGER NOT NULL,
			expires_at  INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_upload_sessions_status_expires ON upload_sessions(status, expires_at)`,
		`CREATE TABLE IF NOT EXISTS invoice_attachments (
			id                TEXT    PRIMARY KEY,
			invoice_id        TEXT,
			session_id        TEXT    REFERENCES upload_sessions(id) ON DELETE SET NULL,
			mime              TEXT    NOT NULL,
			original_filename TEXT    NOT NULL,
			size_bytes        INTEGER NOT NULL,
			created_at        INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_invoice_attachments_invoice ON invoice_attachments(invoice_id)`,
		`CREATE INDEX IF NOT EXISTS idx_invoice_attachments_session ON invoice_attachments(session_id)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	alters := []string{
		`ALTER TABLE businesses ADD COLUMN logo_type TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE recurring_invoices ADD COLUMN issue_day_of_week INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE recurring_invoices ADD COLUMN issue_day_of_month INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE recurring_invoices ADD COLUMN issue_month_of_year INTEGER NOT NULL DEFAULT 0`,
	}
	for _, q := range alters {
		if _, err := db.Exec(q); err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			return fmt.Errorf("%s: %w", q, err)
		}
	}
	return migrateNullableBusinessID(db)
}

// SQLite has no ALTER COLUMN DROP NOT NULL, so dropping the constraint requires
// a full table rebuild. Idempotent via pragma_table_info.notnull check.
func migrateNullableBusinessID(db *sql.DB) error {
	type rebuild struct {
		table     string
		createSQL string
		indexSQLs []string
	}
	plans := []rebuild{
		{
			table: "invoices",
			createSQL: `CREATE TABLE invoices_new (
				id           TEXT    PRIMARY KEY,
				business_id  TEXT    REFERENCES businesses(id) ON DELETE CASCADE,
				amount       REAL    NOT NULL,
				currency     TEXT    NOT NULL,
				description  TEXT    NOT NULL DEFAULT '',
				issued_at    INTEGER NOT NULL,
				due_at       INTEGER NOT NULL,
				paid         INTEGER NOT NULL DEFAULT 0,
				paid_at      INTEGER NOT NULL DEFAULT 0,
				created_at   INTEGER NOT NULL,
				updated_at   INTEGER NOT NULL
			)`,
			indexSQLs: []string{
				`CREATE INDEX IF NOT EXISTS idx_invoices_business_id ON invoices(business_id)`,
				`CREATE INDEX IF NOT EXISTS idx_invoices_due_at ON invoices(due_at)`,
				`CREATE INDEX IF NOT EXISTS idx_invoices_paid_due_at ON invoices(paid, due_at)`,
			},
		},
		{
			table: "recurring_invoices",
			createSQL: `CREATE TABLE recurring_invoices_new (
				id                  TEXT    PRIMARY KEY,
				business_id         TEXT    REFERENCES businesses(id) ON DELETE CASCADE,
				amount              REAL    NOT NULL,
				currency            TEXT    NOT NULL,
				description         TEXT    NOT NULL DEFAULT '',
				frequency           TEXT    NOT NULL,
				start_at            INTEGER NOT NULL,
				end_at              INTEGER NOT NULL DEFAULT 0,
				active              INTEGER NOT NULL DEFAULT 1,
				issue_day_of_week   INTEGER NOT NULL DEFAULT 0,
				issue_day_of_month  INTEGER NOT NULL DEFAULT 0,
				issue_month_of_year INTEGER NOT NULL DEFAULT 0,
				created_at          INTEGER NOT NULL,
				updated_at          INTEGER NOT NULL
			)`,
			indexSQLs: []string{
				`CREATE INDEX IF NOT EXISTS idx_recurring_business_id ON recurring_invoices(business_id)`,
				`CREATE INDEX IF NOT EXISTS idx_recurring_active ON recurring_invoices(active)`,
			},
		},
		{
			table: "upcoming_invoices",
			createSQL: `CREATE TABLE upcoming_invoices_new (
				id           TEXT    PRIMARY KEY,
				business_id  TEXT    REFERENCES businesses(id) ON DELETE CASCADE,
				amount       REAL    NOT NULL,
				currency     TEXT    NOT NULL,
				description  TEXT    NOT NULL DEFAULT '',
				due_at       INTEGER NOT NULL,
				created_at   INTEGER NOT NULL,
				updated_at   INTEGER NOT NULL
			)`,
			indexSQLs: []string{
				`CREATE INDEX IF NOT EXISTS idx_upcoming_business_id ON upcoming_invoices(business_id)`,
				`CREATE INDEX IF NOT EXISTS idx_upcoming_due_at ON upcoming_invoices(due_at)`,
			},
		},
	}
	for _, p := range plans {
		var notNull int
		err := db.QueryRow(`SELECT "notnull" FROM pragma_table_info(?) WHERE name = 'business_id'`, p.table).Scan(&notNull)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return err
		}
		if notNull == 0 {
			continue
		}
		// PRAGMA foreign_keys cannot change inside a transaction; we must toggle it
		// here so the rebuild's DROP/RENAME doesn't trigger FK checks mid-flight.
		if _, err := db.Exec(`PRAGMA foreign_keys = OFF`); err != nil {
			return err
		}
		if err := rebuildTable(db, p.table, p.createSQL, p.indexSQLs); err != nil {
			db.Exec(`PRAGMA foreign_keys = ON`)
			return err
		}
		if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
			return err
		}
	}
	return nil
}

func rebuildTable(db *sql.DB, table, createSQL string, indexSQLs []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(createSQL); err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO ` + table + `_new SELECT * FROM ` + table); err != nil {
		return err
	}
	if _, err := tx.Exec(`DROP TABLE ` + table); err != nil {
		return err
	}
	if _, err := tx.Exec(`ALTER TABLE ` + table + `_new RENAME TO ` + table); err != nil {
		return err
	}
	for _, idx := range indexSQLs {
		if _, err := tx.Exec(idx); err != nil {
			return err
		}
	}
	return tx.Commit()
}
