# tiny-timer

TinyGo firmware for [Arduino Nano RP2040 Connect](https://docs.arduino.cc/hardware/nano-rp2040-connect) that drives an [IoT Relay](https://iotrelay.com/) control input on **`machine.D2`** (digital pin D2).

- **On time:** `OnHours` in `main.go` (1–23). D2 is **high** for that many hours, then **low** for the rest of each **24 h** period, repeating from power-on/reset (no wall-clock sync). While D2 is high, the onboard LED (`machine.LED`) does a **100 ms flash once per second**; it stays off during the off phase.

Patch `OnHours` and flash in one step:

```bash
./flash.sh 18
OK: flashed OnHours=18 (D2 high 18h, low 6h per 24h cycle)
```

