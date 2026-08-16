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
	_ "embed"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	store   *TaskStore
	metrics *Metrics
	config  Config
}

func NewServer(store *TaskStore, metrics *Metrics, config Config) *Server {
	return &Server{store: store, metrics: metrics, config: config}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleConsole)
	mux.HandleFunc("/api/tasks", s.handleTasks)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/readyz", s.handleReady)
	mux.HandleFunc("/metrics", s.handleMetrics)
	return withHeaders(mux)
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tasks, err := s.store.Pending(r.Context(), 100)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read task queue"})
			return
		}
		pending, running, completed, failed, err := s.store.Counts(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read task counts"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks, "pending": pending, "running": running, "completed": completed, "failed": failed})
	case http.MethodPost:
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			writeJSON(w, http.StatusUnsupportedMediaType, map[string]string{"error": "content type must be application/json"})
			return
		}
		task, err := readTask(r, s.config.MaxBodyBytes)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		stored, duplicate, err := s.store.Enqueue(r.Context(), task, r.Header.Get("Idempotency-Key"))
		if err != nil {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "task already exists or could not be queued"})
			return
		}
		if duplicate {
			writeJSON(w, http.StatusOK, stored)
			return
		}
		writeJSON(w, http.StatusAccepted, stored)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	pending, running, completed, failed, err := s.store.Counts(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read status"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"service": "q-harvest", "state": "operational", "carbon_intensity": s.metrics.CarbonIntensity(), "carbon_threshold": s.config.CarbonThreshold, "queue_depth": pending, "running": running, "completed": completed, "failed": failed, "pqc": "integration boundary only", "updated_at": time.Now().UTC()})
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.store.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "store unavailable"})
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	pending, running, completed, failed, err := s.store.Counts(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "metrics unavailable"})
		return
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	fmt.Fprintf(w, "# HELP qharvest_task_queue_size Tasks waiting for execution.\n# TYPE qharvest_task_queue_size gauge\nqharvest_task_queue_size %d\n", pending)
	fmt.Fprintf(w, "qharvest_tasks_running %d\nqharvest_tasks_completed_total %d\nqharvest_tasks_failed %d\nqharvest_carbon_intensity %d\n", running, completed, failed, s.metrics.CarbonIntensity())
}

func (s *Server) handleConsole(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(consoleHTML))
}

func withHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src https://fonts.gstatic.com; script-src 'self' 'unsafe-inline'")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}

//go:embed console.html
var consoleHTML string
