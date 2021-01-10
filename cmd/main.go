package main

import (
	"github.com/womat/debug"
	"os"
	"os/signal"
	"syscall"

	"powerusage/global"
	"powerusage/pkg/powerusage"
)

type pm struct {
	data *powerusage.Measurements
}

func main() {
	debug.SetDebug(global.Config.Debug.File, global.Config.Debug.Flag)

	global.Measurements = powerusage.New()
	global.Measurements.SetInverterMeterURL(global.Config.InverterMeterURL)
	global.Measurements.SetPowerMeterURL(global.Config.PowerMeterURL)

	runtime := &pm{
		data: global.Measurements,
	}

	go runtime.calcRuntime(global.Config.DataCollectionInterval)

	// capture exit signals to ensure resources are released on exit.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	// wait for am os.Interrupt signal (CTRL C)
	sig := <-quit
	debug.InfoLog.Printf("Got %s signal. Aborting...\n", sig)
	os.Exit(1)
}
