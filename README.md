# tiny-timer

TinyGo firmware that drives an [IoT Relay](https://iotrelay.com/)–style **active-high** control input.

It runs a **repeating 24h schedule** from reset (no real-time clock): you set how long the **on-window** is each cycle; the remainder of the day is the **off-window**. Optionally, after each full cycle the **on-window** can shrink by a fixed **per-cycle reduction** until it reaches zero; then the firmware **halts** with the relay off. During each on-window the relay can stay on continuously or follow an optional **duty cycle**. The onboard LED blinks or follows that output.

| Target            | `tinygo -target=…` | Relay GPIO   | Onboard LED   |
| ----------------- | -------------------- | ------------------------ | ------------- |
| Arduino Nano RP2040 Connect | `nano-rp2040` | `machine.D2` | `machine.LED` |
| Raspberry Pi Pico         | `pico`         | `machine.GP2`             | `machine.LED` |

## Configuration (build time)

`main.go` declares **`onHours`**, **`dutyPercent`**, **`dutyPeriodMins`**, and **`onWindowReduceMins`** as **uninitialized** `string` package variables. TinyGo will **ignore** `-ldflags -X` if you assign defaults in the `var` block ([How to set build-time variables](https://tinygo.org/docs/guides/tips-n-tricks/#how-to-set-build-time-variables)). When nothing is set at link time, `readConfig` uses **12h / 100% / 0 / 0** (no per-cycle reduction). **Override at build** with **`-ldflags -X 'main.onHours=…' …`** (see `flash.sh`).

**`dutyPeriodMins`:** `0` means the relay stays on for the full on-window (same as 100% duty). If **`dutyPercent` &lt; 100**, set **`dutyPeriodMins` ≥ 1** so the relay can cycle within each period.

**`onWindowReduceMins`:** Non‑negative integer minutes **subtracted from the next on-window** after each complete **on + off** cycle (e.g. **`15`** with **`onHours=17.5`**: first cycle **17.5h** on, second **17.25h** on, …). **`0`** disables reduction (repeat the same on-window forever). When the on-window would become **≤ 0**, the device **stops**, relay **off** (infinite idle).

**`onHours`:** Decimal **hours** for the **starting** on-window (`0 < onHours < 24`), e.g. **`12.5`** → 12h30m. **Off-window** each cycle is **24h − current on-window**. No wall clock; timing starts at reset.

With full on (100% duty or `dutyPeriodMins` 0), the onboard LED does a **100ms blink once per second** while the relay is on. With sub-100% duty, the **LED follows the relay**.

## Build / flash

```bash
tinygo build -target=nano-rp2040 -o firm.uf2 .
tinygo build -target=pico -o firm.uf2 .
```

Flash with `flash.sh` (no edits to `main.go`):

```bash
./flash.sh 12                    # 12h on-window, no reduction
./flash.sh 12.5 pico             # 12h30m on-window every cycle
./flash.sh 17.5 pico 100 0 15    # 17.5h then 17.25h… (−15m/cycle); then halt
./flash.sh 12 nano-rp2040 50 60  # duty cycle; no reduction (omit 5th arg)
```

Equivalent by hand (note **one** pair of quotes around the entire `main.…=…` / `-X` list — if `-ldflags` is not quoted, the shell passes only the first word and you get default values in the firmware):

```bash
tinygo flash -ldflags='-X main.onHours=17.5 -X main.dutyPercent=100 -X main.dutyPeriodMins=0 -X main.onWindowReduceMins=15' -target pico -monitor .
```

Omit `-monitor` if you do not need the serial log after programming.

## Pin files

- `relay_pin_nano_rp2040.go` (`//go:build nano_rp2040`) — `D2`
- `relay_pin_pico.go` (`//go:build pico`) — `GP2`

Add another board by copying a pin file, setting the build tag to that [TinyGo target name](https://tinygo.org/docs/reference/microcontrollers/), and mapping the GPIO you wire to the IoT relay input.
