package main

import (
	"machine"
	"strconv"
	"time"
)

// println goes to the default serial (USB on Nano RP2040 / Pico) — use
// `tinygo flash -monitor` or `tinygo monitor` to watch.
const logPrefix = "[tiny-timer] "

// Build-time strings via -ldflags -X (e.g. -X 'main.onHours=4' …). Do not initialize
// these here or -X is ignored; see
// https://tinygo.org/docs/guides/tips-n-tricks/#how-to-set-build-time-variables
// When still empty at runtime, readConfig uses 12, 100, 0.
var (
	onHours        string
	dutyPercent    string
	dutyPeriodMins string
)

// Internal tick for duty mode (100ms) — balance timing resolution vs. CPU wakeups.
const dutyTick = 100 * time.Millisecond

const flashOn = 100 * time.Millisecond

func main() {
	relay := boardRelayPin()
	relay.Configure(machine.PinConfig{Mode: machine.PinOutput})
	relay.Set(false)

	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.Set(false)

	oh, dp, dperiod, ok := readConfig()
	if !ok {
		logLine("invalid config; halting")
		for {
		}
	}

	time.Sleep(2 * time.Second) // help USB CDC so early lines show in -monitor
	perLog := 0
	if dperiod > 0 {
		perLog = int(dperiod / time.Minute)
	}
	logLine("start onHours=" + strconv.Itoa(oh) + "h offHours=" + strconv.Itoa(24-oh) +
		"h duty%=" + strconv.Itoa(dp) + " periodMins=" + strconv.Itoa(perLog))

	on := time.Duration(oh) * time.Hour
	off := 24*time.Hour - on

	fullOn := dperiod <= 0 || dp >= 100

	for {
		logLine("on-window start (" + durHours(on) + ")")
		if fullOn {
			logLine("relay ON (holds for full on-window)")
			relay.Set(true)
			for s := 0; s < int(on/time.Second); s++ {
				led.Set(true)
				time.Sleep(flashOn)
				led.Set(false)
				time.Sleep(time.Second - flashOn)
			}
		} else {
			pm := int(dperiod / time.Minute)
			logLine("duty cycle: relay toggles, period " + strconv.Itoa(pm) + "m, duty " + strconv.Itoa(dp) + "%")
			period := dperiod
			if period < dutyTick {
				period = dutyTick
			}
			onPart := time.Duration((int64(period) * int64(dp)) / 100)
			if onPart < dutyTick {
				onPart = dutyTick
			}
			if onPart > period {
				onPart = period
			}
			periodSteps := int(period / dutyTick)
			if periodSteps < 1 {
				periodSteps = 1
			}
			onSteps := int(onPart / dutyTick)
			if onSteps < 1 {
				onSteps = 1
			}
			if onSteps > periodSteps {
				onSteps = periodSteps
			}
			totalSteps := int((on + dutyTick - 1) / dutyTick)
			var lastRel bool
			for n := 0; n < totalSteps; n++ {
				posInPeriod := n % periodSteps
				rel := posInPeriod < onSteps
				relay.Set(rel)
				led.Set(rel)
				if n == 0 || rel != lastRel {
					if rel {
						logLine("relay ON")
					} else {
						logLine("relay OFF")
					}
					lastRel = rel
				}
				time.Sleep(dutyTick)
			}
		}
		logLine("on-window end; relay OFF, off-window start (" + durHours(off) + ")")
		relay.Set(false)
		led.Set(false)
		time.Sleep(off)
		logLine("off-window end")
	}
}

func readConfig() (onH, dPct int, dPer time.Duration, ok bool) {
	ohs, dps, pms := onHours, dutyPercent, dutyPeriodMins
	if ohs == "" {
		ohs = "12"
	}
	if dps == "" {
		dps = "100"
	}
	if pms == "" {
		pms = "0"
	}
	oh, err := strconv.Atoi(ohs)
	if err != nil || oh < 1 || oh > 23 {
		return
	}
	dP, err := strconv.Atoi(dps)
	if err != nil || dP < 1 || dP > 100 {
		return
	}
	mins, err := strconv.Atoi(pms)
	if err != nil || mins < 0 {
		return
	}
	if dP == 100 {
		return oh, 100, 0, true
	}
	if mins < 1 {
		return
	}
	period := time.Duration(mins) * time.Minute
	return oh, dP, period, true
}

func logLine(msg string) {
	println(logPrefix + msg)
}

func durHours(d time.Duration) string {
	// on/off windows are always a whole number of hours in this app
	return strconv.FormatInt(int64(d/time.Hour), 10) + "h"
}
