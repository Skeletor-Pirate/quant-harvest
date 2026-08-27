# ⚡ Quant Harvest — Carbon-Aware Scheduler for Post-Quantum Workloads

[![Go Report Card](https://goreportcard.com/badge/github.com/Skeletor-Pirate/quant-harvest)](https://goreportcard.com/report/github.com/Skeletor-Pirate/quant-harvest)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Live Demo](https://img.shields.io/badge/Demo-Live_on_Render-purple.svg)](https://quant-harvest.onrender.com/)
[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8.svg)](https://go.dev)
[![Deploy to Render](https://img.shields.io/badge/Deploy-Render-46E3B7.svg)](https://quant-harvest.onrender.com)

> **Defer heavy ML-KEM / PQC batch jobs until the grid is cleanest. Go + SQLite (WAL) + Prometheus.**

**Quant Harvest** is a cloud-native, **carbon-aware control-plane** for scheduling computationally heavy **post-quantum cryptography (PQC)** workloads — ML-KEM (FIPS 203), bulk signing, encryption batches — when energy grid carbon intensity (gCO2e/kWh) is lowest.

It implements a bounded, deferred-execution queue that polls real-time carbon intensity against a user-defined threshold (`200 gCO2e/kWh` default), dynamically **pausing and resuming** queue processing. Ideal for green computing, sustainable DevOps, and climate-tech infrastructure.

**🎥 Live Demo:** https://quant-harvest.onrender.com — try `POST /api/tasks` and watch the carbon-aware scheduler defer at 260 gCO2e/kWh.

<!-- Demo GIF: record 15s of console at https://quant-harvest.onrender.com and replace this -->
<!-- ![Demo](docs/demo.gif) -->

---

## 🚀 Key Features

*   **Carbon-Aware Scheduling**: Defers queue processing automatically when the grid's carbon intensity (gCO2e/kWh) exceeds the configured threshold.
*   **Idempotency & Reliability**: SQLite-backed task store running in WAL (Write-Ahead Logging) mode with robust transaction handling and automatic retries.
*   **Built-in Operator Console**: A responsive, self-contained dashboard served at the root path (`/`) to monitor grid status and manage workloads in real-time.
*   **Observability-Ready**: Native Prometheus `/metrics` endpoint, along with standard `/healthz` (liveness) and `/readyz` (readiness) check endpoints.
*   **Explicit Cryptographic Boundary**: Includes a structured boundary (`pqc.go`) ready for ML-KEM/PQC library integration.

---

## 🛠️ Tech Stack & Architecture

*   **Language**: Go 1.20+ (Standard library focused — zero heavy framework)
*   **Database**: Embedded SQLite (`modernc.org/sqlite` pure-Go driver) in **WAL mode** + busy timeouts for concurrent claim safety
*   **Observability**: Prometheus (`/metrics` — `qharvest_task_queue_size`, `qharvest_carbon_intensity`) + `/healthz` `/readyz` + hardened headers (CSP, nosniff)
*   **Deploy**: Dockerfile (multi-stage, distroless) + `deploy/kubernetes.yaml` + Render

```
Enqueue → SQLite WAL → Scheduler (tick 15s) → Carbon check (<200?) → Workers (concurrency 2) → Prometheus
```

---

## 📋 API Reference

| Endpoint | Method | Description | Headers / Payload |
| :--- | :--- | :--- | :--- |
| `/` | `GET` | Serves the Operator Console dashboard. | — |
| `/healthz` | `GET` | Liveness check. Returns `200 OK`. | — |
| `/readyz` | `GET` | Readiness check. Verifies SQLite connection. | — |
| `/metrics` | `GET` | Exposes Prometheus metrics (`qharvest_task_queue_size`, etc.). | — |
| `/api/status` | `GET` | Fetches scheduler status and current grid intensity. | — |
| `/api/tasks` | `GET` | Lists all pending/queued tasks. | — |
| `/api/tasks` | `POST` | Enqueues a new PQC workload. | JSON body: `{"name": "...", "payload": "..."}`<br>Supports optional `Idempotency-Key` header. |

### Enqueue Workload Example

**Request:**
```bash
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: unique-uuid-12345" \
  -d '{"name": "pqc-batch-signing", "payload": "ml-kem-768-request-block"}'
```

**Response:**
```json
{
  "id": "qh-1786772642674",
  "name": "pqc-batch-signing",
  "payload": "ml-kem-768-request-block",
  "status": "queued",
  "created_at": "2026-08-15T06:30:00Z"
}
```

---

## ⚙️ Configuration Options

Configuration is managed via environment variables or command-line flags (flags take precedence).

| Environment Variable | CLI Flag | Default | Description |
| :--- | :--- | :--- | :--- |
| `QHARVEST_ADDR` | `-addr` | `:8080` | Bind address for the HTTP service. |
| `QHARVEST_CARBON_THRESHOLD` | `-carbon-threshold` | `200` | Maximum carbon intensity (gCO2e/kWh) before execution is paused. |
| `QHARVEST_TICK` | `-tick` | `15s` | Scheduler tick interval for grid polling. |
| `QHARVEST_MAX_CONCURRENCY` | `-max-concurrency` | `2` | Maximum concurrent workers running tasks. |
| `QHARVEST_STORE` | `-store` | `data/qharvest.db` | Location of the SQLite storage file. |

---

## 🏃 Getting Started

### Local Setup

Ensure you have [Go](https://go.dev/doc/install) installed.

1. **Run the project locally**:
   ```bash
   go run .
   ```
2. Open [http://localhost:8080](http://localhost:8080) in your browser to view the operator console.
3. **Run unit tests**:
   ```bash
   go test -v ./...
   ```

### Docker Container

Build and launch the lightweight multi-stage Docker container locally:

```bash
# Build the image
docker build -t qharvest:latest .

# Run the container
docker run --rm -p 8080:8080 -v qharvest-data:/app/data qharvest:latest
```

### Kubernetes Deployment

Kubernetes manifests are located in the `deploy/` folder:

```bash
kubectl apply -f deploy/config.yaml
kubectl apply -f deploy/kubernetes.yaml
```

---

## 🔒 Security & Post-Quantum Boundary

> [!WARNING]
> The current cryptographic boundary defined in `pqc.go` is an envelope representation for scheduling testing. Vetted implementations of ML-KEM/FIPS 203 standards must be imported into `pqc.go` before using this system in production cryptographic environments.

---

## 🌿 Why Carbon-Aware PQC?

Post-quantum algorithms (ML-KEM, Dilithium) are quantum-safe but **2-10x more CPU-intensive** than RSA/ECC. Bulk operations (document signing, backups) can spike carbon footprints. Quant Harvest lets you **harvest clean energy windows** instead of burning fossil-fueled compute — green computing meets quantum-safe security.

Keywords: `green-computing`, `carbon-aware-computing`, `sustainable-devops`, `pqc`, `post-quantum-cryptography`, `ml-kem`, `climate-tech`, `golang-scheduler`

## 🤝 Contributing & Community

Star the repo if this is useful! Feedback on scheduler heuristics, carbon data sources (ElectricityMap, WattTime), and PQC integration (`pqc.go`) welcome.

- **Issues:** https://github.com/Skeletor-Pirate/quant-harvest/issues
- **Live Demo:** https://quant-harvest.onrender.com
- **Blog post template:** `blog_post.md` | **Community kit:** `community_posts.md`

## 📄 License

This project is licensed under the Apache License, Version 2.0 - see the [LICENSE](LICENSE) file for details.

