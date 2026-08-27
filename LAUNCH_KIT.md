# 🚀 Quant Harvest — Launch Kit (Ready to Copy-Paste)

Use this file for real publicity. All links verified 2026-08-27.

Live demo: https://quant-harvest.onrender.com
GitHub: https://github.com/Skeletor-Pirate/quant-harvest
Release: https://github.com/Skeletor-Pirate/quant-harvest/releases/tag/v0.1.0

---

## 0) Pre-flight checklist (do before posting)

- [x] GitHub description + homepage + topics set (carbon-aware, go, pqc etc.)
- [x] README SEO optimized + .gitattributes fixed
- [x] v0.1.0 release tagged
- [ ] Record 15s demo GIF: open https://quant-harvest.onrender.com, `curl -X POST /api/tasks`, show carbon intensity 260 → queue paused → drops → executes. Save as `docs/demo.gif` and uncomment README line. **This doubles engagement on HN/Reddit.**
- [ ] Add `docs/demo.gif` to repo and push — then re-post image link in comments
- [ ] Enable GitHub Discussions (already enabled) + 1 starter discussion "Ideas for real carbon data source?"

---

## 1) Hacker News — Show HN (Post Tue/Wed 08:00–10:00 ET)

**Title (max 80 chars):**
```
Show HN: Quant Harvest – Carbon-aware scheduler for post-quantum workloads (Go)
```

**URL:** `https://github.com/Skeletor-Pirate/quant-harvest`

**Text (comment after submission):**
```markdown
Hi HN,

I built Quant Harvest — a Go control-plane that defers heavy post-quantum crypto (ML-KEM) batch jobs until the grid is cleanest.

Context: ML-KEM/FIPS 203 is quantum-safe but 2-10x more CPU than RSA/ECC. Bulk signing petabytes will spike carbon. Instead of running immediately, Quant Harvest buffers tasks in SQLite (WAL mode) and polls carbon intensity (gCO2e/kWh). Above 200 → pause. Below → harvest.

Stack: pure Go + modernc.org/sqlite + Prometheus /metrics + /healthz /readyz + operator console at /.

Demo: https://quant-harvest.onrender.com — try:
curl -X POST https://quant-harvest.onrender.com/api/tasks -H "Content-Type: application/json" -H "Idempotency-Key: demo-123" -d '{"name":"pqc-batch-signing","payload":"ml-kem-768"}'

Open source Apache-2.0. I'd love feedback on scheduler heuristics and real carbon data integrations (ElectricityMap/WattTime).

GitHub: https://github.com/Skeletor-Pirate/quant-harvest
```

**HN tips:** reply to every comment in first 2 hours. Ask a question at end to boost comments.

---

## 2) Reddit

### r/golang (post Tue 09:00 ET)
**Title:** `Quant Harvest – Go + SQLite (WAL) carbon-aware queue for PQC workloads [Apache-2.0]`
```markdown
Built a Go prototype for carbon-aware scheduling of post-quantum workloads.

Go stdlib + SQLite WAL + busy timeouts to avoid locking under concurrent claims. Scheduler is a simple ticker + worker pool (concurrency 2) with idempotency via Idempotency-Key header / unique constraint. Exposes Prometheus metrics.

Live: https://quant-harvest.onrender.com | Code: https://github.com/Skeletor-Pirate/quant-harvest

Looking for feedback on claim/lock SQL and scheduler tick design. Happy to share schema!
```

### r/devops + r/climateTech + r/GreenComputing (next day)
**Title:** `Carbon-aware deferred execution for PQC workloads — Go prototype with Prometheus + K8s`
```markdown
With PQC migration, bulk signing will get heavier. I prototyped a carbon-aware queue that checks grid carbon intensity and defers execution to clean-energy windows. 

K8s manifests + Dockerfile (distroless) + /metrics included. Would love thoughts on carbon signals — right now it's simulated, next step is ElectricityMap API.

Demo: https://quant-harvest.onrender.com
GitHub: https://github.com/Skeletor-Pirate/quant-harvest
```

---

## 3) dev.to / Hashnode / Medium (SEO evergreen)

**Title:** `Why We Need to Schedule Post-Quantum Cryptography Around the Weather`

Use `blog_post.md` as-is, but add to top:
```
---
Live demo: https://quant-harvest.onrender.com
GitHub: https://github.com/Skeletor-Pirate/quant-harvest
Stack: Go, SQLite, Prometheus, PQC (ML-KEM)
---

```
Tags: `go, pqc, postquantum, greencomputing, carbonaware, devops, climateTech`

Publish same day on Hashnode + Medium (canonical URL = dev.to).

---

## 4) X / Twitter Thread (7 tweets)

```
1/ Post-quantum crypto (ML-KEM) is quantum-safe but CPU-hungry. Running batch signing on fossil-fueled grids = huge carbon spike.

I built Quant Harvest — a Go scheduler that HARVESTS clean energy windows ⚡🌿

Live: https://quant-harvest.onrender.com
GH: https://github.com/Skeletor-Pirate/quant-harvest 🧵

2/ How it works:
Enqueue → SQLite WAL → Scheduler polls grid carbon (gCO2e/kWh) every 15s → if >200 pause, if <200 execute with 2 workers. Idempotency-Key prevents double-sign.

3/ Stack: Go stdlib + modernc.org/sqlite (pure Go, WAL mode) + Prometheus /metrics + operator console. No Kafka, no Redis. Just a binary + SQLite file.

4/ Demo:
curl -X POST https://quant-harvest.onrender.com/api/tasks \
 -H "Idempotency-Key: demo-123" \
 -d '{"name":"pqc-batch-signing","payload":"ml-kem-768"}'
Watch it pause at 260 gCO2e/kWh → resume when clean.

5/ Why now? PQC migration is mandatory (NIST FIPS 203). Bulk operations will 10x compute. Scheduling to renewables is the simplest decarbonization win.

6/ Open source Apache-2.0. Next: real ElectricityMap/WattTime integration + ML-KEM libs in pqc.go.

7/ Try it, star it, roast my scheduler logic:
https://github.com/Skeletor-Pirate/quant-harvest

Feedback welcome! #golang #pqc #greencomputing #climatetech
```

---

## 5) LinkedIn Post

```
Post-quantum cryptography will make our infra quantum-safe — but also 10x more compute-intensive.

I launched Quant Harvest, an open-source Go control plane that defers heavy PQC batch jobs (ML-KEM) to run when the energy grid is cleanest.

→ Carbon-aware queue: checks gCO2e/kWh, pauses above 200, resumes when clean
→ Go + SQLite (WAL) + Prometheus + Kubernetes-ready
→ Live demo: https://quant-harvest.onrender.com

We can decarbonize security itself. Try the demo and tell me what you'd improve — especially around carbon data sources (ElectricityMap, WattTime).

GitHub (Apache-2.0): https://github.com/Skeletor-Pirate/quant-harvest

#GreenComputing #PostQuantum #Golang #SustainableTech #DevOps
```

---

## 6) Product Hunt (Week 2)

**Name:** Quant Harvest — Carbon-aware scheduler for post-quantum workloads
**Tagline:** Defer heavy PQC jobs to clean-energy windows. Go + SQLite + Prometheus.
**Description:** Use same HN text + GIF. Category: Developer Tools / Green Tech.

---

## 7) GitHub ecosystem (Day 3-7)

- Add to `awesome-go` PR: section `Queue` or `Green Computing`
- Submit issue to `Green Software Foundation` awesome list
- Enable GitHub Sponsors (optional) + add `FUNDING.yml`
- Comment helpfully on 5-10 issues: `carbon aware`, `pqc go`, `sqlite queue` — link only if relevant

---

## Calendar

| Day | Action | Link |
|-----|--------|------|
| Tue 08:00 ET | Show HN | news.ycombinator.com |
| Tue 09:00 ET | r/golang | reddit.com/r/golang |
| Wed | r/devops + r/GreenComputing | |
| Wed | dev.to + Hashnode publish | |
| Thu | X thread + LinkedIn | |
| Week 2 | Product Hunt + awesome-go PR | |

**Golden rule:** Be present for 2h after each post. Reply fast, ask follow-ups, push demo link.

---

## Metrics to track

- GitHub stars, forks, clones (`gh api repos/Skeletor-Pirate/quant-harvest/traffic/clones`)
- Render demo requests: check `GET /api/status` queue_depth
- dev.to views + HN points + Reddit upvotes

Good luck — you now have everything to copy-paste. Record that GIF and you’ll 2x conversion.
