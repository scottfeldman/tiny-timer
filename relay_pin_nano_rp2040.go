//go:build nano_rp2040

package main

import "machine"

func boardRelayPin() machine.Pin { return machine.D2 }
