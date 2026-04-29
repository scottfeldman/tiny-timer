//go:build pico

package main

import "machine"

func boardRelayPin() machine.Pin { return machine.GP16 }
