package main

import (
	"context"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/dweymouth/go-upnpcast/device"
)

func main() {
	if devices, err := device.SearchMediaRenderers(context.Background(), 5); err != nil {
		log.Printf("Error loading devices: %v", err)
	} else {
		spew.Dump(devices)
	}
}
