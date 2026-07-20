package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "embed"

	"github.com/maksapakov/family-shopping-bot/internal/domain"
	"github.com/maksapakov/family-shopping-bot/internal/port"
	_ "modernc.org/sqlite"
)

var ErrNotFound = fmt.Errorf("list not found")

//go:embed migration/001_init.sql
var migrationSQL string

//go:embed migration/002_undo_actions.sql
var migrationUndoSQL string

type ListRepo struct {
	db *sql.DB
}

func (r *ListRepo) GetByChatID(ctx context.Context, chatID domain.ChatID) (*domain.ShoppingList, error) {
	var listID, messageRef string
	var activeTab int

	err := r.db.QueryRowContext(ctx, `
		SELECT list_id, active_tab, message_ref
		FROM shopping_lists
		WHERE chat_id = ?`,
		string(chatID),
	).Scan(&listID, &activeTab, &messageRef)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query shopping_lists: %w", err)
	}

	list := domain.NewShoppingList(domain.ListID(listID), chatID)
	list.ActiveTab = domain.Location(activeTab)
	list.MessageRef = messageRef

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, is_checked, checked_at, added_by, location
		FROM items
		WHERE chat_id = ?`,
		string(chatID),
	)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var id, name, addedBy string
		var isChecked, location int
		var checkedAt sql.NullString

		if err := rows.Scan(&id, &name, &isChecked, &checkedAt, &addedBy, &location); err != nil {
			return nil, fmt.Errorf("scan items: %w", err)
		}

		item := domain.NewItem(domain.ItemID(id), name, addedBy, domain.Location(location))

		if isChecked == 1 {
			var at time.Time
			if checkedAt.Valid {
				at, err = time.Parse(time.RFC3339, checkedAt.String)
				if err != nil {
					return nil, fmt.Errorf("parse time: %w", err)
				}
			}
			item.Restore(true, at)
		}

		list.AddItem(item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate items: %w", err)
	}

	undoRows, err := r.db.QueryContext(ctx, `
			SELECT item_id, was_checked, checked_at
			FROM undo_actions
			WHERE chat_id = ?
			ORDER BY position`,
		string(chatID),
	)
	if err != nil {
		return nil, fmt.Errorf("query undos: %w", err)
	}
	defer func() {
		_ = undoRows.Close()
	}()

	var snaps []domain.UndoSnapshot
	for undoRows.Next() {
		var itemID string
		var wasChecked int
		var checkedAt sql.NullString
		if err := undoRows.Scan(&itemID, &wasChecked, &checkedAt); err != nil {
			return nil, fmt.Errorf("scan undo: %w", err)
		}
		snap := domain.UndoSnapshot{
			ItemID:     domain.ItemID(itemID),
			WasChecked: wasChecked == 1,
		}
		if checkedAt.Valid {
			snap.CheckedAt, err = time.Parse(time.RFC3339, checkedAt.String)
			if err != nil {
				return nil, fmt.Errorf("parse time: %w", err)
			}
		}
		snaps = append(snaps, snap)
	}
	if err := undoRows.Err(); err != nil {
		return nil, fmt.Errorf("undo rows: %w", err)
	}
	list.RestoreUndoSnapshot(snaps)

	return list, nil
}

func Open(dbPath string) (*ListRepo, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite db: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pinging sqlite db: %w", err)
	}

	for _, q := range []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON;",
	} {
		if _, err := db.Exec(q); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("executing %s: %w", q, err)
		}
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := runMigration(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return &ListRepo{db: db}, nil
}

func runMigration(db *sql.DB) error {
	for _, sqlText := range []string{migrationSQL, migrationUndoSQL} {
		if _, err := db.Exec(sqlText); err != nil {
			return fmt.Errorf("executing %s: %w", sqlText, err)
		}
	}
	return nil
}

func (r *ListRepo) Close() error {
	return r.db.Close()
}

func (r *ListRepo) Save(ctx context.Context, list *domain.ShoppingList) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO shopping_lists (chat_id, list_id, active_tab, message_ref)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(chat_id) DO UPDATE SET 
			list_id = excluded.list_id,
			active_tab = excluded.active_tab,
			message_ref = excluded.message_ref`,
		string(list.ChatID),
		string(list.ID),
		int(list.ActiveTab),
		list.MessageRef,
	)
	if err != nil {
		return fmt.Errorf("inserting list: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		DELETE FROM items WHERE chat_id = ?`,
		string(list.ChatID),
	)
	if err != nil {
		return fmt.Errorf("deleting items: %w", err)
	}

	for _, item := range list.Items {
		checked := 0
		if item.IsChecked() {
			checked = 1
		}

		var checkedAt any
		if item.IsChecked() {
			checkedAt = item.CheckedAt.UTC().Format(time.RFC3339)
		} else {
			checkedAt = nil
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO items (id, chat_id, name, is_checked, checked_at, added_by, location)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			string(item.ID),
			string(list.ChatID),
			item.Name,
			checked,
			checkedAt,
			item.AddedBy,
			int(item.Location()),
		)
		if err != nil {
			return fmt.Errorf("inserting item: %w", err)
		}
	}

	_, err = tx.ExecContext(ctx, `
		DELETE FROM undo_actions WHERE chat_id = ?`, string(list.ChatID))
	if err != nil {
		return fmt.Errorf("deleting undo_actions: %w", err)
	}

	for pos, snap := range list.UndoSnapshot() {
		var checkedAt any
		if snap.WasChecked && !snap.CheckedAt.IsZero() {
			checkedAt = snap.CheckedAt.UTC().Format(time.RFC3339)
		}

		wasChecked := 0
		if snap.WasChecked {
			wasChecked = 1
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO undo_actions (chat_id, item_id, was_checked, checked_at, position)
			VALUES (?, ?, ?, ?, ?)`,
			string(list.ChatID),
			string(snap.ItemID),
			wasChecked,
			checkedAt,
			pos)
	}
	if err != nil {
		return fmt.Errorf("inserting undo_actions: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing items: %w", err)
	}
	return nil
}

var _ port.ListRepository = (*ListRepo)(nil)
