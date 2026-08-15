# Why We Need to Schedule Post-Quantum Cryptography Around the Weather

The security industry is currently undergoing one of the largest migrations in history: moving from classical cryptography (like RSA and ECC) to **Post-Quantum Cryptography (PQC)**. 

While algorithms like **ML-KEM** (FIPS 203) are quantum-resistant, they come with a hidden cost: **they are incredibly CPU-intensive**. Compared to classical algorithms, key generation, signing, and verification consume significantly more compute cycles.

More compute cycles mean more electricity. And in an era where tech companies are under pressure to reach net-zero carbon goals, running massive PQC operations (like bulk document signing or encrypting petabytes of backup data) could significantly spike your carbon footprint.

But what if we could schedule these workloads to run **only when the energy grid is cleanest**?

Enter **Quant Harvest**.

---

## ⚡ What is Quant Harvest?

**Quant Harvest** is an open-source, cloud-native control-plane prototype written in Go. It acts as a carbon-aware deferred-execution queue for PQC workloads. 

Instead of executing computationally heavy cryptographic signing immediately, Quant Harvest buffers workloads and monitors the local energy grid's carbon intensity (gCO2e/kWh). When the grid is running on fossil fuels (dirty energy), the scheduler pauses the queue. The moment solar or wind energy kicks in (clean energy), the scheduler automatically resumes and harvests the workloads.

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

---

## 🛠️ The Architecture

Quant Harvest is designed as a lightweight, production-ready Go daemon:

*   **SQLite in WAL Mode**: The task queue is backed by an embedded SQLite database running in Write-Ahead Logging mode for concurrency, durability, and atomic idempotency.
*   **Dynamic Scheduler**: A background ticker routinely queries local grid carbon intensity metrics and manages a worker pool that processes claims.
*   **Prometheus Metrics**: Exposes native metrics (e.g., `qharvest_task_queue_size`, `qharvest_carbon_intensity`) for real-time monitoring.
*   **Console UI**: A self-contained, responsive HTML operator console to queue and watch workloads.

---

## 💻 Applying the Code

Adding a workload is as simple as hitting a hardened HTTP REST endpoint. The API enforces strict validation and supports `Idempotency-Key` headers to prevent double-scheduling.

Here's a sample request to enqueue a signing job:

```bash
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: task-uuid-9988-7766" \
  -d '{"name": "pqc-bulk-signing", "payload": "ml-kem-768-batch-payload"}'
```

And the scheduler handles the deferral log:
```log
2026/08/15 12:01:45 Q-HARVEST listening on :8080
2026/08/15 12:02:00 scheduler: deferring queue at 270gCO2e/kWh
2026/08/15 12:02:15 scheduler: carbon intensity dropped to 145gCO2e/kWh
2026/08/15 12:02:15 scheduler: executing pqc-bulk-signing
```

---

## 🔮 The Future: Post-Quantum Integration

The project currently has an explicit integration boundary inside `pqc.go` representing PQC envelopes:

```go
// EnvelopeIdentifier is a non-cryptographic integration placeholder.
func EnvelopeIdentifier(data []byte) [32]byte {
	return sha256.Sum256(data)
}
```

The next milestone is integrating vetted Go implementations of standard PQC algorithms (like `crypto/hybrid` or standard Go PQC packages) to run actual cryptographic operations under carbon constraints.

---

## 🤝 Get Involved!

We believe green computing and quantum-resistant security should go hand-in-hand. 

Quant Harvest is open-source under the **Apache License 2.0**. 

If you are a Go developer, a DevOps engineer, or a cryptography enthusiast, check out the repository, run the tests, and help us build the future of sustainable security!

👉 **GitHub Repository**: [https://github.com/Skeletor-Pirate/quant-harvest](https://github.com/Skeletor-Pirate/quant-harvest)
