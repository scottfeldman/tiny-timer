# tiny-timer

TinyGo firmware for [Arduino Nano RP2040 Connect](https://docs.arduino.cc/hardware/nano-rp2040-connect) that drives an [IoT Relay](https://iotrelay.com/) control input on **`machine.D2`** (digital pin D2).

- **On time:** `OnHours` in `main.go` (1–23). D2 is **high** for that many hours, then **low** for the rest of each **24 h** period, repeating from power-on/reset (no wall-clock sync). While D2 is high, the onboard LED (`machine.LED`) does a **100 ms flash once per second**; it stays off during the off phase.

Patch `OnHours` and flash in one step:

```bash
chmod +x flash.sh   # once
./flash.sh 12
```

This rewrites the `const OnHours` line in `main.go`, then runs `GOWORK=off tinygo flash -target=nano-rp2040`.

Build (disable `go.work` if your environment errors on workspace modules):

```bash
GOWORK=off tinygo build -target=nano-rp2040 -o tiny-timer.uf2 .
```

Flash (double-tap reset for UF2 boot if needed):

```bash
GOWORK=off tinygo flash -target=nano-rp2040 .
```

