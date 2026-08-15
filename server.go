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

const consoleHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>Quant Harvest / Carbon-Aware PQC</title>
  <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%2310B981' stroke-width='2'%3E%3Ccircle cx='12' cy='12' r='10' stroke='%233F3F46'/%3E%3Cpath d='M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z' stroke='%2310B981'/%3E%3Cpath d='M12 6c0 0-3 3-3 5.5s2 4.5 3 4.5 3-2 3-4.5S12 6 12 6z' fill='%2310B981' opacity='0.8'/%3E%3C/svg%3E">
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&family=Outfit:wght@600;700;800&display=swap" rel="stylesheet">
  <style>
    :root {
      --bg: #030303;
      --card: rgba(10, 10, 12, 0.7);
      --border: rgba(255, 255, 255, 0.05);
      --border-hover: rgba(255, 255, 255, 0.12);
      --text: #F4F4F5;
      --text-soft: #A1A1AA;
      
      /* Pure HSL harmony (No AI-slop neon) */
      --brand: hsl(142, 70%, 50%); /* Clean Green */
      --brand-soft: rgba(16, 185, 129, 0.1);
      --accent: hsl(200, 95%, 45%); /* Slate Blue */
      --warn: hsl(38, 92%, 50%); /* Amber */
      --danger: hsl(0, 72%, 51%); /* Soft Red */
      
      --easing: cubic-bezier(0.16, 1, 0.3, 1);
    }
    
    * { box-sizing: border-box; margin: 0; padding: 0; }
    
    body {
      font-family: 'Inter', system-ui, sans-serif;
      background: var(--bg);
      color: var(--text);
      max-width: 1040px;
      margin: 0 auto;
      padding: 60px 24px;
      line-height: 1.6;
      background-image: 
        radial-gradient(circle at top left, rgba(16, 185, 129, 0.04) 0%, transparent 35%),
        radial-gradient(circle at bottom right, rgba(99, 102, 241, 0.03) 0%, transparent 40%);
      background-attachment: fixed;
      -webkit-font-smoothing: antialiased;
    }
    
    header {
      margin-bottom: 48px;
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      gap: 24px;
    }
    
    .brand-group {
      display: flex;
      flex-direction: column;
      gap: 12px;
    }
    
    .brand-logo-title {
      display: flex;
      align-items: center;
      gap: 16px;
    }
    
    /* Handcrafted SVG Logo */
    .brand-svg {
      width: 44px;
      height: 44px;
      filter: drop-shadow(0 0 12px rgba(16, 185, 129, 0.25));
    }
    
    h1 {
      font-family: 'Outfit', sans-serif;
      font-size: 32px;
      font-weight: 800;
      letter-spacing: -0.03em;
      line-height: 1.15;
      color: #FFFFFF;
    }
    
    .desc {
      color: var(--text-soft);
      max-width: 600px;
      font-size: 14px;
      margin-top: 4px;
    }
    
    .badge {
      font-size: 10px;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.14em;
      color: var(--brand);
      background: var(--brand-soft);
      border: 1px solid rgba(16, 185, 129, 0.15);
      padding: 6px 14px;
      border-radius: 30px;
      height: fit-content;
      margin-top: 4px;
    }
    
    .grid {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 16px;
      margin-bottom: 32px;
    }
    
    .card {
      background: var(--card);
      border: 1px solid var(--border);
      border-radius: 16px;
      padding: 20px;
      transition: border-color 0.3s var(--easing), transform 0.3s var(--easing), box-shadow 0.3s var(--easing);
      backdrop-filter: blur(16px);
      position: relative;
      overflow: hidden;
    }
    
    .card::before {
      content: '';
      position: absolute;
      top: 0; left: 0; width: 100%; height: 100%;
      background: linear-gradient(180deg, rgba(255,255,255,0.01) 0%, transparent 100%);
      pointer-events: none;
    }
    
    .card:hover {
      border-color: var(--border-hover);
      transform: translateY(-4px);
      box-shadow: 0 12px 30px rgba(0, 0, 0, 0.6);
    }
    
    .card-label {
      font-size: 10px;
      text-transform: uppercase;
      letter-spacing: 0.1em;
      color: var(--text-soft);
      font-weight: 600;
    }
    
    .card-value {
      font-family: 'Outfit', sans-serif;
      font-size: 26px;
      font-weight: 700;
      margin-top: 10px;
      display: flex;
      align-items: baseline;
    }
    
    .layout-main {
      display: grid;
      grid-template-columns: 1fr 340px;
      gap: 24px;
    }
    
    .panel {
      background: var(--card);
      border: 1px solid var(--border);
      border-radius: 20px;
      padding: 28px;
      backdrop-filter: blur(16px);
    }
    
    .panel-title {
      font-family: 'Outfit', sans-serif;
      font-size: 16px;
      font-weight: 700;
      margin-bottom: 20px;
      display: flex;
      align-items: center;
      gap: 10px;
      color: #FFFFFF;
      letter-spacing: -0.01em;
    }
    
    .panel-title svg {
      color: var(--text-soft);
    }
    
    /* Live Chart styling */
    .chart-container {
      height: 150px;
      margin-bottom: 32px;
      background: rgba(0, 0, 0, 0.2);
      border-radius: 14px;
      border: 1px solid var(--border);
      overflow: hidden;
      padding: 10px;
    }
    
    .chart-svg {
      width: 100%;
      height: 100%;
      overflow: visible;
    }
    
    .threshold-line {
      stroke: var(--danger);
      stroke-dasharray: 6 4;
      stroke-width: 1.2;
    }
    
    .chart-line {
      fill: none;
      stroke: var(--brand);
      stroke-width: 2.5;
      stroke-linecap: round;
      stroke-linejoin: round;
      filter: drop-shadow(0 4px 10px rgba(16, 185, 129, 0.3));
    }
    
    .chart-area {
      fill: url(#chart-gradient);
      opacity: 0.12;
    }
    
    /* Logger Panel */
    .logger-panel {
      display: flex;
      flex-direction: column;
      height: 100%;
    }
    
    .console-win {
      flex: 1;
      background: #050507;
      border: 1px solid var(--border);
      border-radius: 14px;
      padding: 18px;
      font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
      font-size: 11.5px;
      color: var(--text-soft);
      overflow-y: auto;
      max-height: 380px;
      min-height: 250px;
      box-shadow: inset 0 4px 20px rgba(0,0,0,0.8);
      line-height: 1.6;
    }
    
    .log-line {
      margin-bottom: 8px;
      display: flex;
      gap: 10px;
      animation: log-fade 0.25s var(--easing) forwards;
    }
    
    @keyframes log-fade {
      from { opacity: 0; transform: translateY(2px); }
      to { opacity: 1; transform: translateY(0); }
    }
    
    .log-time { color: rgba(255,255,255,0.25); }
    .log-text-info { color: var(--text); }
    .log-text-warning { color: var(--warn); }
    .log-text-success { color: var(--brand); }
    .log-text-system { color: var(--accent); }
    
    /* Task Queue Row styling */
    .queue-container {
      margin-top: 16px;
      border: 1px solid var(--border);
      border-radius: 14px;
      overflow: hidden;
      background: #050507;
    }
    
    .queue-row {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 16px 20px;
      border-bottom: 1px solid var(--border);
      transition: background 0.2s var(--easing);
    }
    
    .queue-row:last-child { border-bottom: none; }
    .queue-row:hover { background: rgba(255,255,255,0.015); }
    
    .status-badge {
      font-size: 10px;
      font-weight: 700;
      padding: 4px 10px;
      border-radius: 8px;
      text-transform: uppercase;
      letter-spacing: 0.08em;
    }
    
    .status-queued { background: rgba(245, 158, 11, 0.1); color: var(--warn); border: 1px solid rgba(245, 158, 11, 0.2); }
    .status-running { background: rgba(6, 182, 212, 0.1); color: var(--cyan); border: 1px solid rgba(6, 182, 212, 0.2); }
    .status-completed { background: rgba(16, 185, 129, 0.1); color: var(--brand); border: 1px solid rgba(16, 185, 129, 0.2); }
    
    .empty-state {
      padding: 48px;
      text-align: center;
      color: var(--text-soft);
      font-size: 13.5px;
    }
    
    /* Flowchart / Explanation Cards */
    .flow-steps {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 16px;
      margin-top: 40px;
      border-top: 1px solid var(--border);
      padding-top: 32px;
    }
    
    .step-card {
      background: rgba(255,255,255,0.01);
      border: 1px dashed var(--border);
      border-radius: 16px;
      padding: 20px;
      transition: border-color 0.3s var(--easing);
    }
    
    .step-card:hover {
      border-color: rgba(255,255,255,0.15);
    }
    
    .step-num {
      width: 24px;
      height: 24px;
      border-radius: 50%;
      background: var(--brand-soft);
      border: 1px solid rgba(16, 185, 129, 0.3);
      display: grid;
      place-items: center;
      font-size: 11px;
      font-weight: 700;
      margin-bottom: 12px;
      color: var(--brand);
    }
    
    .step-title {
      font-size: 13px;
      font-weight: 600;
      margin-bottom: 6px;
      color: #FFFFFF;
    }
    
    .step-desc {
      font-size: 12px;
      color: var(--text-soft);
      line-height: 1.5;
    }
    
    /* Form input styling (Anti-AI-slop custom feel) */
    form {
      display: flex;
      gap: 12px;
      margin-top: 20px;
    }
    
    input {
      flex: 1;
      background: #050507;
      border: 1px solid var(--border);
      border-radius: 12px;
      padding: 16px;
      color: white;
      font-family: inherit;
      font-size: 13.5px;
      transition: border-color 0.3s var(--easing), box-shadow 0.3s var(--easing);
    }
    
    input:focus {
      outline: none;
      border-color: var(--brand);
      box-shadow: 0 0 0 4px rgba(16, 185, 129, 0.1);
    }
    
    button {
      background: #FFFFFF;
      border: none;
      color: #030303;
      font-family: 'Outfit', sans-serif;
      font-weight: 700;
      font-size: 13.5px;
      padding: 0 28px;
      border-radius: 12px;
      cursor: pointer;
      transition: transform 0.2s var(--easing), opacity 0.2s var(--easing);
    }
    
    button:hover { opacity: 0.95; }
    button:active { transform: scale(0.97); }
    
    @media(max-width: 900px) {
      .layout-main { grid-template-columns: 1fr; }
      .grid { grid-template-columns: repeat(2, 1fr); }
      .flow-steps { grid-template-columns: 1fr; }
      header { flex-direction: column; align-items: flex-start; gap: 16px; }
    }
  </style>
</head>
<body>
  <header>
    <div class="brand-group">
      <div class="brand-logo-title">
        <!-- Stylized PQC orbital atomic leaf logo -->
        <svg class="brand-svg" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg">
          <circle cx="16" cy="16" r="14" stroke="#1F1F23" stroke-width="2"/>
          <ellipse cx="16" cy="16" rx="14" ry="4" stroke="#7C3AED" stroke-width="1.5" transform="rotate(45 16 16)"/>
          <ellipse cx="16" cy="16" rx="14" ry="4" stroke="#06B6D4" stroke-width="1.5" transform="rotate(-45 16 16)"/>
          <path d="M16 8C16 8 12 12 12 15.5C12 17.98 13.79 20 16 20C18.21 20 20 17.98 20 15.5C20 12 16 8 16 8Z" fill="#10B981" opacity="0.9"/>
        </svg>
        <h1>Quant Harvest</h1>
      </div>
      <p class="desc">A cloud-native carbon-aware control plane that defers heavy post-quantum cryptographic (PQC) workloads when grid carbon intensity exceeds local limits.</p>
    </div>
    <div class="badge">Operational Control Plane</div>
  </header>

  <section class="grid">
    <div class="card">
      <div class="card-label">Grid Carbon Intensity</div>
      <div class="card-value" style="color: var(--brand)" id="carbon">—</div>
    </div>
    <div class="card">
      <div class="card-label">Execution Cutoff</div>
      <div class="card-value" style="color: var(--danger)" id="threshold">—</div>
    </div>
    <div class="card">
      <div class="card-label">Queue Depth</div>
      <div class="card-value" style="color: var(--warn)" id="queue">—</div>
    </div>
    <div class="card">
      <div class="card-label">Completed Jobs</div>
      <div class="card-value" style="color: #FFFFFF" id="completed">—</div>
    </div>
  </section>

  <div class="layout-main">
    <div class="panel">
      <div class="panel-title">
        <!-- SVG Icon Grid Chart -->
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 3v18h18"/><path d="m19 9-5 5-4-4-3 3"/></svg>
        Grid Carbon Intensity History
      </div>
      
      <div class="chart-container">
        <svg class="chart-svg" id="chart" viewBox="0 0 500 120" preserveAspectRatio="none">
          <defs>
            <linearGradient id="chart-gradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="var(--brand)" />
              <stop offset="100%" stop-color="transparent" />
            </linearGradient>
          </defs>
          <!-- Threshold line -->
          <line id="thresh-line" class="threshold-line" x1="0" y1="60" x2="500" y2="60" />
          <!-- Area under curve -->
          <path id="chart-area-path" class="chart-area" d="" />
          <!-- Carbon line -->
          <path id="chart-line-path" class="chart-line" d="" />
        </svg>
      </div>

      <div class="panel-title">
        <!-- SVG Icon Queue -->
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" height="18" width="18" y="3" rx="2"/><path d="M21 9H3"/><path d="M21 15H3"/></svg>
        Workload Dispatch Queue
      </div>
      
      <div class="queue-container" id="tasks">
        <div class="empty-state">Loading workloads...</div>
      </div>

      <form onsubmit="enqueue(event)">
        <input id="name" maxlength="120" placeholder="Workload name (e.g., pqc-key-rotation-job)" required>
        <button>Queue Job</button>
      </form>
    </div>

    <div class="panel logger-panel">
      <div class="panel-title">
        <!-- SVG Icon Console -->
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/></svg>
        System Logs
      </div>
      <div class="console-win" id="console">
        <div class="log-line">
          <span class="log-time">[System]</span>
          <span class="log-text-system">Initializing Quant Harvest operator console...</span>
        </div>
      </div>
    </div>
  </div>

  <section class="flow-steps">
    <div class="step-card">
      <div class="step-num">1</div>
      <div class="step-title">Queue PQC Job</div>
      <div class="step-desc">Enter a workload to schedule computationally heavy signing or key exchanges.</div>
    </div>
    <div class="step-card">
      <div class="step-num">2</div>
      <div class="step-title">Real-Time Grid Check</div>
      <div class="step-desc">Scheduler continuously queries local carbon intensity metrics of the energy grid.</div>
    </div>
    <div class="step-card">
      <div class="step-num">3</div>
      <div class="step-title">Carbon Threshold Guard</div>
      <div class="step-desc">If the grid is dirty (exceeds threshold), execution is dynamically deferred.</div>
    </div>
    <div class="step-card">
      <div class="step-num">4</div>
      <div class="step-title">Execute on Clean Energy</div>
      <div class="step-desc">When carbon drops, scheduler claims tasks and processes them safely in SQLite WAL.</div>
    </div>
  </section>

  <script>
    let carbonHistory = [];
    const maxHistoryPoints = 25;
    let lastCarbonVal = null;
    let lastQueueVal = null;

    function addLog(text, type = 'info') {
      const win = document.getElementById('console');
      const timeStr = new Date().toLocaleTimeString();
      const line = document.createElement('div');
      line.className = 'log-line';
      line.innerHTML = '<span class="log-time">['+timeStr+']</span> <span class="log-text-'+type+'">'+text+'</span>';
      win.appendChild(line);
      win.scrollTop = win.scrollHeight;
    }

    function updateChart(history, threshold) {
      if (history.length < 2) return;
      const width = 500;
      const height = 120;
      
      const minVal = 80;
      const maxVal = 300;
      const range = maxVal - minVal;

      // Map threshold Y coordinate
      const threshY = height - ((threshold - minVal) / range) * height;
      document.getElementById('thresh-line').setAttribute('y1', threshY);
      document.getElementById('thresh-line').setAttribute('y2', threshY);

      // Construct points
      let points = [];
      const step = width / (maxHistoryPoints - 1);
      
      history.forEach((pt, i) => {
        const offset = maxHistoryPoints - history.length;
        const x = (i + offset) * step;
        const y = height - ((pt.value - minVal) / range) * height;
        points.push({x, y});
      });

      // Line Path
      let pathD = 'M ' + points[0].x + ' ' + points[0].y;
      for (let i = 1; i < points.length; i++) {
        pathD += ' L ' + points[i].x + ' ' + points[i].y;
      }
      document.getElementById('chart-line-path').setAttribute('d', pathD);

      // Area Path
      let areaD = pathD + ' L ' + points[points.length - 1].x + ' ' + height + ' L ' + points[0].x + ' ' + height + ' Z';
      document.getElementById('chart-area-path').setAttribute('d', areaD);
    }

    async function refresh() {
      try {
        const [s, t] = await Promise.all([
          fetch('/api/status').then(r => r.json()),
          fetch('/api/tasks').then(r => r.json())
        ]);

        carbon.textContent = s.carbon_intensity + ' g';
        threshold.textContent = s.carbon_threshold + ' g';
        queue.textContent = s.queue_depth;
        completed.textContent = s.completed;

        // Log grid checks
        if (s.carbon_intensity !== lastCarbonVal) {
          const stateStr = s.carbon_intensity > s.carbon_threshold ? 'DIRTY (Deferred)' : 'CLEAN (Ready)';
          const logType = s.carbon_intensity > s.carbon_threshold ? 'warning' : 'success';
          addLog('Grid check: ' + s.carbon_intensity + ' gCO2e/kWh - Status: ' + stateStr, logType);
          
          carbonHistory.push({ value: s.carbon_intensity });
          if (carbonHistory.length > maxHistoryPoints) carbonHistory.shift();
          updateChart(carbonHistory, s.carbon_threshold);
          lastCarbonVal = s.carbon_intensity;
        }

        if (s.queue_depth !== lastQueueVal && lastQueueVal !== null) {
          if (s.queue_depth > lastQueueVal) {
            addLog('New task enqueued. Queue depth: ' + s.queue_depth, 'system');
          } else {
            addLog('Workload dispatched successfully. Queue depth: ' + s.queue_depth, 'success');
          }
          lastQueueVal = s.queue_depth;
        } else if (lastQueueVal === null) {
          lastQueueVal = s.queue_depth;
        }

        // Render Queue
        const tasksContainer = document.getElementById('tasks');
        if (t.tasks && t.tasks.length) {
          tasksContainer.innerHTML = t.tasks.map(x => {
            const statusClass = x.status === 'queued' ? 'status-queued' : (x.status === 'running' ? 'status-running' : 'status-completed');
            return '<div class="queue-row"><strong>' + esc(x.name) + '</strong><span class="status-badge ' + statusClass + '">' + x.status + '</span></div>';
          }).join('');
        } else {
          tasksContainer.innerHTML = '<div class="empty-state">No workloads waiting in the queue.</div>';
        }

      } catch (err) {
        addLog('Error reading control plane status: ' + err.message, 'warning');
      }
    }

    function esc(x) {
      return x.replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
    }

    async function enqueue(e) {
      e.preventDefault();
      const input = document.getElementById('name');
      const taskName = input.value.trim();
      if (!taskName) return;

      try {
        addLog('Submitting workload "' + taskName + '" with idempotency token...', 'system');
        const res = await fetch('/api/tasks', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Idempotency-Key': crypto.randomUUID()
          },
          body: JSON.stringify({ name: taskName })
        });
        
        if (res.ok) {
          addLog('Workload "' + taskName + '" enqueued successfully.', 'info');
        } else {
          const data = await res.json();
          addLog('Failed to queue: ' + (data.error || 'Server error'), 'warning');
        }
        input.value = '';
        refresh();
      } catch (err) {
        addLog('Network error enqueuing task: ' + err.message, 'warning');
      }
    }

    // Seed initial historical chart data
    for (let i = 0; i < maxHistoryPoints; i++) {
      carbonHistory.push({ value: 150 + Math.floor(Math.random() * 80) });
    }

    refresh();
    setInterval(refresh, 5000);
  </script>
</body>
</html>`
