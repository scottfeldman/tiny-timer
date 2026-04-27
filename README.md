# tiny-timer

TinyGo firmware that drives an [IoT Relay](https://iotrelay.com/)–style **active-high** control input.

It runs a **repeating 24h schedule** from reset (no real-time clock): you set how long the **on-window** is each cycle; the remainder of the day is the **off-window**. During the on-window the relay can stay on continuously or follow an optional **duty cycle** (on/off within a repeating period). The onboard LED blinks or follows that output.

| Target            | `tinygo -target=…` | Relay GPIO   | Onboard LED   |
| ----------------- | -------------------- | ------------------------ | ------------- |
| Arduino Nano RP2040 Connect | `nano-rp2040` | `machine.D2` | `machine.LED` |
| Raspberry Pi Pico         | `pico`         | `machine.GP2`             | `machine.LED` |

## Configuration (build time)

`main.go` declares **`onHours`**, **`dutyPercent`**, and **`dutyPeriodMins`** as **uninitialized** `string` package variables. TinyGo will **ignore** `-ldflags -X` if you assign defaults in the `var` block ([How to set build-time variables](https://tinygo.org/docs/guides/tips-n-tricks/#how-to-set-build-time-variables)). When nothing is set at link time, `readConfig` uses **12h / 100% / 0** period-mins. **Override at build** with **`-ldflags -X 'main.onHours=…' …`** (see `flash.sh`).

**`dutyPeriodMins`:** `0` means the relay stays on for the full on-window (same as 100% duty). If **`dutyPercent` &lt; 100**, set **`dutyPeriodMins` ≥ 1** so the relay can cycle within each period.

**`onHours`:** 1–23 — length of the **on-window** each 24h cycle; **off-window** is `24 - onHours` hours. No wall clock; timing starts at reset.

With full on (100% duty or `dutyPeriodMins` 0), the onboard LED does a **100ms blink once per second** while the relay is on. With sub-100% duty, the **LED follows the relay**.

## Build / flash

```bash
tinygo build -target=nano-rp2040 -o firm.uf2 .
tinygo build -target=pico -o firm.uf2 .
```

Flash with `flash.sh` (no edits to `main.go`):

```bash
./flash.sh 12                    # onHours=12, default target, 100% duty, no cycling
./flash.sh 12 pico
./flash.sh 12 nano-rp2040 50 60 # 50% on, 60 min period, during the on-window
```

Equivalent by hand (note **one** pair of quotes around the entire `main.onHours=…` / `-X` list — if `-ldflags` is not quoted, the shell passes only the first word and you get default values in the firmware):

```bash
tinygo flash -ldflags='-X main.onHours=12 -X main.dutyPercent=100 -X main.dutyPeriodMins=0' -target pico -monitor .
```

Omit `-monitor` if you do not need the serial log after programming.

## Pin files

- `relay_pin_nano_rp2040.go` (`//go:build nano_rp2040`) — `D2`
- `relay_pin_pico.go` (`//go:build pico`) — `GP2`

Add another board by copying a pin file, setting the build tag to that [TinyGo target name](https://tinygo.org/docs/reference/microcontrollers/), and mapping the GPIO you wire to the IoT relay input.
