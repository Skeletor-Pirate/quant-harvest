---
title: "Why We Need to Schedule Post-Quantum Cryptography Around the Weather"
published: false
tags: go, pqc, postquantum, greencomputing, devops, climateTech, opensource
cover_image: https://quant-harvest.onrender.com/static/media/og-image.png
canonical_url: 
description: "ML-KEM is quantum-safe but CPU-hungry. Quant Harvest is a Go + SQLite WAL scheduler that defers batch PQC jobs until the grid is clean — live demo inside."
---

Live demo: **https://quant-harvest.onrender.com** | GitHub: **https://github.com/Skeletor-Pirate/quant-harvest** (Apache-2.0) | Go + SQLite WAL + Prometheus

> Copy everything below into dev.to editor (Markdown). Replace cover with screenshot from live demo if needed.

# Why We Need to Schedule Post-Quantum Cryptography Around the Weather

The security industry is undergoing one of the largest migrations in history: moving from classical cryptography (RSA/ECC) to **Post-Quantum Cryptography (PQC)**.

While algorithms like **ML-KEM** (FIPS 203) are quantum-resistant, they have a hidden cost: they are **CPU-intensive**. Compared to classical algorithms, key generation, signing, and verification consume significantly more compute.

More compute = more electricity. For bulk operations (signing millions of docs, encrypting petabytes of backups), this spikes carbon footprint — bad for net-zero goals.

What if we schedule these workloads to run **only when the grid is cleanest**?

Enter **Quant Harvest**.

---

## What is Quant Harvest?

**Quant Harvest** is an open-source Go control plane — a carbon-aware deferred-execution queue for PQC workloads.

Instead of executing heavy crypto immediately, it buffers workloads and monitors grid carbon intensity (gCO2e/kWh). Dirty grid (fossil) → pause queue. Clean grid (solar/wind) → resume and harvest.

```
                  ┌──────────────────────┐
                  │   Workload Enqueued  │
                  └──────────┬───────────┘
                             ▼
                  ┌──────────────────────┐
                  │    Task Buffered     │
                  └──────────┬───────────┘
                             ▼
               No  /───────────────────\
             ┌────/  Grid Clean enough? \
             │    \   (Under threshold) /
             │     \───────────────────/
             ▼               │ Yes
     ┌──────────────┐        ▼
     │ Queue Paused │ ┌──────────────┐
     └──────────────┘ │ Execute Task │
                      └──────────────┘
```

**Try it live:** https://quant-harvest.onrender.com

```bash
curl -X POST https://quant-harvest.onrender.com/api/tasks \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: demo-123" \
  -d '{"name":"pqc-batch-signing","payload":"ml-kem-768"}'

# Watch /api/status — when carbon >200 it defers, when <200 it runs
curl https://quant-harvest.onrender.com/api/status | jq
```

---

## Architecture

- **Go stdlib + SQLite WAL** (`modernc.org/sqlite` pure-Go driver) — busy timeouts, atomic claims, idempotency via `Idempotency-Key`
- **Ticker scheduler** — polls every 15s, worker pool concurrency 2, threshold 200 gCO2e/kWh
- **Prometheus** — `/metrics` (`qharvest_task_queue_size`, `qharvest_carbon_intensity`) + `/healthz` `/readyz`
- **Operator console** at `/` — real-time grid chart + queue + WAL logs (see live demo)
- **Docker + Kubernetes** manifests in `deploy/`

```go
// pqc.go — integration boundary ready for ML-KEM
func EnvelopeIdentifier(data []byte) [32]byte {
    return sha256.Sum256(data) // placeholder for vetted ML-KEM/FIPS203
}
```

---

## How logging looks

```
2026/08/15 12:01:45 Q-HARVEST listening on :8080
2026/08/15 12:02:00 scheduler: deferring queue at 270gCO2e/kWh
2026/08/15 12:02:15 scheduler: carbon intensity dropped to 145gCO2e/kWh
2026/08/15 12:02:15 scheduler: executing pqc-bulk-signing
```

---

## Why "harvest"?

Green computing is usually "use less." Quant Harvest is "use **when** clean." Same jobs, shifted to renewable windows — simplest decarbonization win for PQC migration.

---

## Get involved

- **GitHub:** https://github.com/Skeletor-Pirate/quant-harvest — star + open issues on carbon data sources (ElectricityMap/WattTime)
- **Release v0.1.0:** https://github.com/Skeletor-Pirate/quant-harvest/releases/tag/v0.1.0
- **Awesome-go PR:** https://github.com/avelino/awesome-go/pull/6626

Apache-2.0. If you're Go/DevOps/crypto, try the live demo and roast the scheduler logic!

---

*Keywords: green computing, carbon-aware, sustainable devops, pqc, post-quantum cryptography, ml-kem, climate-tech, golang scheduler*
