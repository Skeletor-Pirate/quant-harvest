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
	"testing"
	"time"
)

func TestTaskStorePersistsAndDeduplicates(t *testing.T) {
	store, err := OpenTaskStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	task := Task{ID: "qh-1", Name: "batch", CreatedAt: time.Now().UTC()}
	first, duplicate, err := store.Enqueue(ctx, task, "request-1")
	if err != nil || duplicate || first.Status != "queued" {
		t.Fatalf("first enqueue = %#v duplicate=%v err=%v", first, duplicate, err)
	}
	second, duplicate, err := store.Enqueue(ctx, Task{ID: "qh-2", Name: "other", CreatedAt: time.Now().UTC()}, "request-1")
	if err != nil || !duplicate || second.ID != task.ID {
		t.Fatalf("duplicate enqueue = %#v duplicate=%v err=%v", second, duplicate, err)
	}
	claimed, err := store.Claim(ctx, task.ID)
	if err != nil || !claimed {
		t.Fatalf("claim = %v, %v", claimed, err)
	}
	if err := store.Complete(ctx, task.ID); err != nil {
		t.Fatal(err)
	}
	_, _, completed, _, err := store.Counts(ctx)
	if err != nil || completed != 1 {
		t.Fatalf("completed=%d err=%v", completed, err)
	}
}

func TestConfigRejectsUnsafeValues(t *testing.T) {
	if err := (Config{TickInterval: 0, MaxConcurrency: 1, MaxRetries: 1, MaxBodyBytes: 1, StorePath: "x"}).Validate(); err == nil {
		t.Fatal("expected invalid interval")
	}
	if err := (Config{TickInterval: time.Second, MaxConcurrency: 0, MaxRetries: 1, MaxBodyBytes: 1, StorePath: "x"}).Validate(); err == nil {
		t.Fatal("expected invalid concurrency")
	}
}
