package main

import (
	"machine"
	"time"
)

// OnHours is how long D2 stays high each 24 h cycle (1–23).
// Change this before flashing. The first cycle begins at power-on/reset.
const OnHours = 18

func main() {
	relay := machine.D2
	relay.Configure(machine.PinConfig{Mode: machine.PinOutput})
	relay.Set(false)

	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.Set(false)

	if OnHours < 1 || OnHours > 23 {
		for {
		}
	}

	on := time.Duration(OnHours) * time.Hour
	off := 24*time.Hour - on

	const flashOn = 100 * time.Millisecond

	for {
		relay.Set(true)
		for s := 0; s < int(on/time.Second); s++ {
			led.Set(true)
			time.Sleep(flashOn)
			led.Set(false)
			time.Sleep(time.Second - flashOn)
		}
		relay.Set(false)
		led.Set(false)
		time.Sleep(off)
	}
}
