// Copyright 2026 Skeletor-Pirate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

var ErrTaskNotFound = errors.New("task not found")

type TaskStore struct {
	db *sql.DB
}

func OpenTaskStore(path string) (*TaskStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	store := &TaskStore{db: db}
	if err := store.init(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *TaskStore) init() error {
	_, err := s.db.Exec(`
		PRAGMA journal_mode = WAL;
		PRAGMA busy_timeout = 5000;
		CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			payload TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL,
			idempotency_key TEXT UNIQUE,
			created_at TEXT NOT NULL,
			started_at TEXT,
			completed_at TEXT,
			retries INTEGER NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS idx_tasks_status_created ON tasks(status, created_at);
	`)
	return err
}

func (s *TaskStore) Close() error { return s.db.Close() }

func (s *TaskStore) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }

func (s *TaskStore) Enqueue(ctx context.Context, task Task, idempotencyKey string) (Task, bool, error) {
	if idempotencyKey != "" {
		var existing Task
		var createdAt string
		err := s.db.QueryRowContext(ctx, `SELECT id, name, payload, status, created_at FROM tasks WHERE idempotency_key = ?`, idempotencyKey).
			Scan(&existing.ID, &existing.Name, &existing.Payload, &existing.Status, &createdAt)
		if err == nil {
			existing.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
			if err != nil {
				return Task{}, false, err
			}
			return existing, true, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return Task{}, false, err
		}
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO tasks (id, name, payload, status, idempotency_key, created_at) VALUES (?, ?, ?, 'queued', NULLIF(?, ''), ?)`, task.ID, task.Name, task.Payload, idempotencyKey, task.CreatedAt.Format(time.RFC3339Nano))
	if err != nil {
		return Task{}, false, err
	}
	task.Status = "queued"
	return task, false, nil
}

func (s *TaskStore) Pending(ctx context.Context, limit int) ([]Task, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, payload, status, created_at FROM tasks WHERE status = 'queued' ORDER BY created_at LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []Task
	for rows.Next() {
		var task Task
		var createdAt string
		if err := rows.Scan(&task.ID, &task.Name, &task.Payload, &task.Status, &createdAt); err != nil {
			return nil, err
		}
		task.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (s *TaskStore) Claim(ctx context.Context, id string) (bool, error) {
	result, err := s.db.ExecContext(ctx, `UPDATE tasks SET status = 'running', started_at = ? WHERE id = ? AND status = 'queued'`, time.Now().UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	return count == 1, err
}

func (s *TaskStore) Complete(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `UPDATE tasks SET status = 'completed', completed_at = ? WHERE id = ? AND status = 'running'`, time.Now().UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return err
	}
	if count, _ := result.RowsAffected(); count != 1 {
		return fmt.Errorf("%w: %s", ErrTaskNotFound, id)
	}
	return nil
}

func (s *TaskStore) Fail(ctx context.Context, id string, maxRetries int) error {
	_, err := s.db.ExecContext(ctx, `UPDATE tasks SET retries = retries + 1, status = CASE WHEN retries + 1 >= ? THEN 'failed' ELSE 'queued' END WHERE id = ? AND status = 'running'`, maxRetries, id)
	return err
}

func (s *TaskStore) Counts(ctx context.Context) (pending, running, completed, failed int, err error) {
	rows, err := s.db.QueryContext(ctx, `SELECT status, COUNT(*) FROM tasks GROUP BY status`)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return 0, 0, 0, 0, err
		}
		switch status {
		case "queued":
			pending = count
		case "running":
			running = count
		case "completed":
			completed = count
		case "failed":
			failed = count
		}
	}
	return pending, running, completed, failed, rows.Err()
}
