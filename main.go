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
// When still empty at runtime, readConfig uses 12, 100, 0, 0. onHours may be fractional
// (e.g. 12.5 → 12h30m on-window; off-window is the rest of 24h).
// onWindowReduceMins: after each full on+off cycle, subtract this from the next on-window;
// when the on-window reaches zero, halt with relay off.
var (
	onHours            string
	dutyPercent        string
	dutyPeriodMins     string
	onWindowReduceMins string
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

	on, reduce, dp, dperiod, ok := readConfig()
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
	startOff := 24*time.Hour - on
	redLog := "0"
	if reduce > 0 {
		redLog = durFmt(reduce)
	}
	logLine("start on-window=" + durFmt(on) + " off-window=" + durFmt(startOff) +
		" duty%=" + strconv.Itoa(dp) + " periodMins=" + strconv.Itoa(perLog) +
		" onReducePerCycle=" + redLog)

	fullOn := dperiod <= 0 || dp >= 100

	curOn := on
	for {
		if curOn <= 0 {
			logLine("on-window exhausted; relay OFF (halt)")
			relay.Set(false)
			led.Set(false)
			for {
			}
		}

		off := 24*time.Hour - curOn
		logLine("on-window start (" + durFmt(curOn) + ")")
		if fullOn {
			logLine("relay ON (holds for full on-window)")
			relay.Set(true)
			for s := 0; s < int(curOn/time.Second); s++ {
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
			totalSteps := int((curOn + dutyTick - 1) / dutyTick)
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
		logLine("on-window end; relay OFF, off-window start (" + durFmt(off) + ")")
		relay.Set(false)
		led.Set(false)
		time.Sleep(off)
		logLine("off-window end")
		if reduce > 0 {
			curOn -= reduce
			if curOn < 0 {
				curOn = 0
			}
			logLine("after reduce " + durFmt(reduce) + ", next on-window=" + durFmt(curOn))
		}
	}
}

func readConfig() (on time.Duration, reduce time.Duration, dPct int, dPer time.Duration, ok bool) {
	ohs, dps, pms, rms := onHours, dutyPercent, dutyPeriodMins, onWindowReduceMins
	if ohs == "" {
		ohs = "12"
	}
	if dps == "" {
		dps = "100"
	}
	if pms == "" {
		pms = "0"
	}
	if rms == "" {
		rms = "0"
	}
	rm, err := strconv.Atoi(rms)
	if err != nil || rm < 0 {
		return
	}
	reduce = time.Duration(rm) * time.Minute

	ohf, err := strconv.ParseFloat(ohs, 64)
	if err != nil || ohf <= 0 || ohf >= 24 {
		return
	}
	on = time.Duration(ohf * float64(time.Hour))
	dP, err := strconv.Atoi(dps)
	if err != nil || dP < 1 || dP > 100 {
		return
	}
	mins, err := strconv.Atoi(pms)
	if err != nil || mins < 0 {
		return
	}
	if dP == 100 {
		return on, reduce, 100, 0, true
	}
	if mins < 1 {
		return
	}
	period := time.Duration(mins) * time.Minute
	return on, reduce, dP, period, true
}

func logLine(msg string) {
	println(logPrefix + msg)
}

func durFmt(d time.Duration) string {
	h := d / time.Hour
	r := d % time.Hour
	m := r / time.Minute
	s := (r % time.Minute) / time.Second
	out := ""
	if h > 0 {
		out += strconv.FormatInt(int64(h), 10) + "h"
	}
	if m > 0 || (h > 0 && s > 0) {
		out += strconv.FormatInt(int64(m), 10) + "m"
	}
	if s > 0 {
		out += strconv.FormatInt(int64(s), 10) + "s"
	}
	if out == "" {
		return "0s"
	}
	return out
}
