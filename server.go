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

const consoleHTML = `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Q-HARVEST / Control Plane</title><style>body{font-family:system-ui;background:#0b0d0e;color:#eef3ef;max-width:1000px;margin:40px auto;padding:20px}h1{font-size:clamp(38px,7vw,76px);letter-spacing:-.07em}.grid{display:grid;grid-template-columns:repeat(4,1fr);gap:12px}.card{background:#151a1b;border:1px solid #293031;border-radius:14px;padding:18px}.label{color:#8e9b98;font-size:11px;text-transform:uppercase;letter-spacing:.12em}.value{font-size:30px;margin-top:20px;color:#b7f397}.panel{margin-top:28px;border:1px solid #293031;border-radius:14px;overflow:hidden}.row{padding:16px;border-bottom:1px solid #293031;display:flex;justify-content:space-between}.empty{padding:25px;color:#8e9b98}form{display:flex;gap:10px;margin-top:16px}input,button{padding:12px;border-radius:9px;border:1px solid #293031;background:#151a1b;color:#eef3ef}input{flex:1}button{background:#b7f397;color:#142016;font-weight:700}@media(max-width:700px){.grid{grid-template-columns:repeat(2,1fr)}} </style></head><body><div class="label">Q-HARVEST / CONTROL PLANE</div><h1>Compute when the grid is clean.</h1><p style="color:#8e9b98;max-width:650px;line-height:1.7">A durable, carbon-aware workload queue with bounded execution and an explicit post-quantum integration boundary.</p><section class="grid"><div class="card"><div class="label">carbon intensity</div><div id="carbon" class="value">—</div></div><div class="card"><div class="label">queue depth</div><div id="queue" class="value">—</div></div><div class="card"><div class="label">completed</div><div id="completed" class="value">—</div></div><div class="card"><div class="label">workers</div><div id="running" class="value">—</div></div></section><div class="panel" id="tasks"><div class="empty">Loading queue…</div></div><form onsubmit="enqueue(event)"><input id="name" maxlength="120" placeholder="Workload name" required><button>Queue workload</button></form><script>async function refresh(){const[s,t]=await Promise.all([fetch('/api/status').then(r=>r.json()),fetch('/api/tasks').then(r=>r.json())]);carbon.textContent=s.carbon_intensity+' gCO₂e/kWh';queue.textContent=s.queue_depth;completed.textContent=s.completed;running.textContent=s.running;tasks.innerHTML=t.tasks.length?t.tasks.map(x=>'<div class="row"><strong>'+esc(x.name)+'</strong><span>'+x.status+'</span></div>').join(''):'<div class="empty">No workloads waiting.</div>'}function esc(x){return x.replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]))}async function enqueue(e){e.preventDefault();const input=document.getElementById('name');await fetch('/api/tasks',{method:'POST',headers:{'Content-Type':'application/json','Idempotency-Key':crypto.randomUUID()},body:JSON.stringify({name:input.value})});input.value='';refresh()}refresh();setInterval(refresh,5000)</script></body></html>`
