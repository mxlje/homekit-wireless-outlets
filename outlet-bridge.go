package main

import (
	"fmt"
	"os/exec"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/log"
)

const (
	// executable path for raspberry-remote
	remotePath = "/home/pi/raspberry-remote/send"

	// the channel that the outlets and remote share (little switches, on=1, off=0)
	remoteChannel = "01100"

	// pin to pair the bridge within HomeKit
	pin = "00102003"
)

func main() {
	// the bridge is used to combine multiple outlets behind one HomeKit device
	// so they can be paired as a group
	bridge := accessory.New(accessory.Info{
		Name: "Wireless Outlet Bridge",
	}, accessory.TypeBridge)

	// a container holds multiple acessories
	c := accessory.NewContainer()

	// create 3 outlet accessories
	for i := 1; i <= 3; i++ {
		outletNr := i

		acc := accessory.NewOutlet(accessory.Info{
			Name:         fmt.Sprintf("Wireless Outlet %d", outletNr),
			Manufacturer: "Max",
			Model:        "A",
		})

		// this function is called every time the "value" of the outlet is changed remotely
		// (through HomeKit on your phone)
		acc.Outlet.On.OnValueRemoteUpdate(func(on bool) {
			var nextState = "0"
			if on {
				nextState = "1"
			}

			output, err := exec.Command("sudo", remotePath, remoteChannel, fmt.Sprintf("%d", outletNr), nextState).Output()
			if err != nil {
				log.Printf("Error: %s (%s)", output, err)
				return
			}

			fmt.Printf("[Outlet %d] changed to %s", outletNr, nextState, output)
		})

		// add the accessory to the container so it can be used with the bridge later
		c.AddAccessory(acc.Accessory)
	}

	// combine the accessories with bridge and set the pin
	config := hc.Config{Pin: pin}
	t, err := hc.NewIPTransport(config, bridge, c.Accessories...)
	if err != nil {
		log.Fatal(err)
	}

	hc.OnTermination(func() {
		t.Stop()
	})

	t.Start()
}
