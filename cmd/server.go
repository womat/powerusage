package main

import (
	"time"

	"github.com/womat/debug"
)

func (r *pm) calcRuntime(p time.Duration) {

	ticker := time.NewTicker(p)
	defer ticker.Stop()

	for ; true; <-ticker.C {
		debug.DebugLog.Println("get data")

		if err := r.data.Read(); err != nil {
			debug.ErrorLog.Printf("get power usage data: %v", err)
			continue
		}
	}
}
