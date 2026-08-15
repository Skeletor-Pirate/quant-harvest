package main

import (
	"context"
	"log"
	"math/rand"
	"time"
)

type CarbonProvider interface {
	Intensity(context.Context) (int, error)
}

type LocalCarbonProvider struct{}

func (LocalCarbonProvider) Intensity(context.Context) (int, error) { return 115 + rand.Intn(160), nil }

func StartScheduler(ctx context.Context, store *TaskStore, metrics *Metrics, cfg Config) {
	provider := LocalCarbonProvider{}
	jobs := make(chan Task)
	for i := 0; i < cfg.MaxConcurrency; i++ {
		go worker(ctx, store, jobs, cfg)
	}
	ticker := time.NewTicker(cfg.TickInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			intensity, err := provider.Intensity(ctx)
			if err != nil {
				log.Printf("scheduler: carbon provider unavailable: %v", err)
				continue
			}
			metrics.SetCarbonIntensity(intensity)
			if intensity > cfg.CarbonThreshold {
				log.Printf("scheduler: deferring queue at %dgCO2e/kWh", intensity)
				continue
			}
			tasks, err := store.Pending(ctx, cfg.MaxConcurrency)
			if err != nil {
				log.Printf("scheduler: load pending tasks: %v", err)
				continue
			}
			for _, task := range tasks {
				claimed, err := store.Claim(ctx, task.ID)
				if err != nil {
					log.Printf("scheduler: claim %s: %v", task.ID, err)
					continue
				}
				if claimed {
					jobs <- task
				}
			}
		}
	}
}

func worker(ctx context.Context, store *TaskStore, jobs <-chan Task, cfg Config) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-jobs:
			log.Printf("scheduler: executing %s", task.Name)
			select {
			case <-ctx.Done():
				return
			case <-time.After(750 * time.Millisecond):
				if err := store.Complete(context.Background(), task.ID); err != nil {
					_ = store.Fail(context.Background(), task.ID, cfg.MaxRetries)
				}
			}
		}
	}
}
