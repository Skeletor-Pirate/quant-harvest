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
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

type Config struct {
	CarbonThreshold int
	TickInterval    time.Duration
	MaxConcurrency  int
	MaxRetries      int
	MaxBodyBytes    int64
	StorePath       string
}

type Task struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Payload   string    `json:"payload,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	addr := flag.String("addr", envOr("QHARVEST_ADDR", ":8080"), "HTTP service address")
	threshold := flag.Int("carbon-threshold", envInt("QHARVEST_CARBON_THRESHOLD", 200), "maximum carbon intensity for execution")
	interval := flag.Duration("tick", envDuration("QHARVEST_TICK", 15*time.Second), "scheduler polling interval")
	concurrency := flag.Int("max-concurrency", envInt("QHARVEST_MAX_CONCURRENCY", 2), "maximum concurrent workloads")
	storePath := flag.String("store", envOr("QHARVEST_STORE", "data/qharvest.db"), "SQLite database path")
	flag.Parse()

	cfg := Config{
		CarbonThreshold: *threshold,
		TickInterval:    *interval,
		MaxConcurrency:  *concurrency,
		MaxRetries:      envInt("QHARVEST_MAX_RETRIES", 3),
		MaxBodyBytes:    int64(envInt("QHARVEST_MAX_BODY_BYTES", 1<<20)),
		StorePath:       *storePath,
	}
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(cfg.StorePath), 0750); err != nil {
		log.Fatalf("create store directory: %v", err)
	}
	store, err := OpenTaskStore(cfg.StorePath)
	if err != nil {
		log.Fatalf("open task store: %v", err)
	}
	defer store.Close()

	metrics := NewMetrics()
	server := NewServer(store, metrics, cfg)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go StartScheduler(ctx, store, metrics, cfg)

	httpServer := &http.Server{
		Addr:              *addr,
		Handler:           server.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		log.Printf("Q-HARVEST listening on %s", *addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server failed: %v", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
}

func (c Config) Validate() error {
	if c.TickInterval <= 0 {
		return fmt.Errorf("tick interval must be positive")
	}
	if c.MaxConcurrency <= 0 {
		return fmt.Errorf("max concurrency must be positive")
	}
	if c.MaxRetries <= 0 {
		return fmt.Errorf("max retries must be positive")
	}
	if c.MaxBodyBytes <= 0 {
		return fmt.Errorf("max body bytes must be positive")
	}
	if c.StorePath == "" {
		return fmt.Errorf("store path must not be empty")
	}
	return nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return value
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return value
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func readTask(r *http.Request, maxBytes int64) (Task, error) {
	r.Body = http.MaxBytesReader(nil, r.Body, maxBytes)
	var input struct {
		Name    string `json:"name"`
		Payload string `json:"payload"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&input); err != nil {
		return Task{}, err
	}
	if input.Name == "" || len(input.Name) > 120 {
		return Task{}, fmt.Errorf("name is required and must be at most 120 characters")
	}
	if len(input.Payload) > 64*1024 {
		return Task{}, fmt.Errorf("payload must be at most 65536 characters")
	}
	return Task{ID: fmt.Sprintf("qh-%d", time.Now().UnixNano()), Name: input.Name, Payload: input.Payload, CreatedAt: time.Now().UTC()}, nil
}
