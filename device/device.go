package device

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/dweymouth/go-upnpcast/services"
	"github.com/koron/go-ssdp"
)

var (
	ErrNoDeviceAvailable = errors.New("no available Media Renderers")
)

// MediaRenderer represents a Digital Media Renderer (DMR) device discovered on the LAN
type MediaRenderer struct {
	// URL for the device's service descrption manifest
	URL string

	// Friendly name of the device
	FriendlyName string

	// Model name of the device
	ModelName string

	avTransportControlURL  string
	avTransportEventSubURL string
	renderingControlURL    string
	connectionManagerURL   string
}

// SearchMediaRenderers searches for MediaRenderer devices on the LAN
// that implement all the service types specified in `requiredServices`
// - waitSec is how many seconds to wait for device responses to the SSDP search
// If passing a context with deadline/expiration, it should be longer than waitSec
func SearchMediaRenderers(ctx context.Context, waitSec int, requiredServices ...services.ServiceType) ([]*MediaRenderer, error) {
	deviceLocations, err := getSSDPAVTransportDeviceLocations(waitSec)
	if err != nil {
		return nil, err
	}

	devices := make([]*MediaRenderer, 0, len(deviceLocations))
	for _, l := range deviceLocations {
		mr, err := mediaRendererFromDeviceURL(ctx, l)
		if err != nil {
			// TODO: surface error to caller
			log.Printf("skipping bad device: %v", err)
			continue
		}
		devices = append(devices, mr)
	}

	if len(devices) > 0 {
		return devices, nil
	}

	return nil, ErrNoDeviceAvailable
}

// Gets the list of DMR schema URLs for all found devices that support the AVTransport service
func getSSDPAVTransportDeviceLocations(waitSec int) ([]string, error) {
	ssdpServices, err := ssdp.Search(ssdp.All, waitSec, "")
	if err != nil {
		return nil, fmt.Errorf("SSDP search error: %w", err)
	}

	var deviceLocations listSet
	for _, srv := range ssdpServices {
		// All DMRs we care about must support the AVTransport service
		if srv.Type == string(services.AVTransport) {
			deviceLocations.add(srv.Location)
		}
	}
	return deviceLocations, nil
}

type listSet []string

func (l *listSet) add(s string) {
	for _, x := range *l {
		if s == x {
			return
		}
	}
	*l = append(*l, s)
}
