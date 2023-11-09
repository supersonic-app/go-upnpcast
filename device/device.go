package device

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"

	"github.com/koron/go-ssdp"
)

var (
	ErrNoDeviceAvailable  = errors.New("loadSSDPservices: No available Media Renderers")
	ErrDeviceNotAvailable = errors.New("devicePicker: Requested device not available")
	ErrSomethingWentWrong = errors.New("devicePicker: Something went terribly wrong")
)

type AVTransportDevice struct {
	// URL of the device
	URL string

	FriendlyName string
}

// GetFriendlyName returns the friendly name value for a the specific DMR url.
func GetFriendlyName(ctx context.Context, dmr string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dmr, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create NewRequest for GetFriendlyName: %w", err)
	}

	req.Header.Set("Connection", "close")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send HTTP request for GetFriendlyName: %w", err)
	}
	defer resp.Body.Close()

	var fn struct {
		FriendlyName string `xml:"device>friendlyName"`
	}

	if err = xml.NewDecoder(resp.Body).Decode(&fn); err != nil {
		return "", fmt.Errorf("failed to read response body for GetFriendlyName: %w", err)
	}

	return fn.FriendlyName, nil
}

// LoadAVTransportDevices returns a map with all the devices that support the
// AVTransport service.
func LoadAVTransportDevices(waitSec int) ([]AVTransportDevice, error) {
	// Reset device list every time we call this.
	var devices []AVTransportDevice
	list, err := ssdp.Search(ssdp.All, waitSec, "")
	if err != nil {
		return nil, fmt.Errorf("LoadSSDPservices search error: %w", err)
	}

	for _, srv := range list {
		// We only care about the AVTransport services for basic actions
		// (stop,play,pause). If we need support other functionalities
		// like volume control we need to use the RenderingControl service.
		if srv.Type == "urn:schemas-upnp-org:service:AVTransport:1" {
			friendlyName, err := GetFriendlyName(context.Background(), srv.Location)
			if err != nil {
				continue
			}

			devices = append(devices, AVTransportDevice{
				URL:          srv.Location,
				FriendlyName: friendlyName,
			})
		}
	}

	/*
		deviceList := make(map[string]string)
		dupNames := make(map[string]int)
		for loc, fn := range urlList {
			_, exists := dupNames[fn]
			dupNames[fn]++
			if exists {
				fn = fn + " (" + loc + ")"
			}

			deviceList[fn] = loc
		}

		for fn, c := range dupNames {
			if c > 1 {
				loc := deviceList[fn]
				delete(deviceList, fn)
				fn = fn + " (" + loc + ")"
				deviceList[fn] = loc
			}
		}
	*/

	if len(devices) > 0 {
		return devices, nil
	}

	return nil, ErrNoDeviceAvailable
}

/*
// DevicePicker will pick the nth device from the devices input map.
func DevicePicker(devices map[string]string, n int) (string, error) {
	if n > len(devices) || len(devices) == 0 || n <= 0 {
		return "", ErrDeviceNotAvailable
	}

	var keys []string
	for k := range devices {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for q, k := range keys {
		if n == q+1 {
			return devices[k], nil
		}
	}

	return "", ErrSomethingWentWrong
}
*/
