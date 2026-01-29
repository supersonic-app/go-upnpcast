package device

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/supersonic-app/go-upnpcast/services"
)

type device struct {
	DeviceType   string `xml:"deviceType"`
	FriendlyName string `xml:"friendlyName"`
	ModelName    string `xml:"modelName"`

	ServiceList struct {
		Services []struct {
			Type        services.Type `xml:"serviceType"`
			ID          string        `xml:"serviceId"`
			ControlURL  string        `xml:"controlURL"`
			EventSubURL string        `xml:"eventSubURL"`
		} `xml:"service"`
	} `xml:"serviceList"`

	DeviceList struct {
		Devices []device `xml:"device"`
	} `xml:"deviceList"`
}

type dmrSchema struct {
	XMLName xml.Name `xml:"root"`
	Device  device   `xml:"device"`
}

func mediaRendererFromDeviceURL(ctx context.Context, dmrurl string) (*MediaRenderer, error) {
	parsedURL, err := url.Parse(dmrurl)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("device URL parse error: %w", err)
	}

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dmrurl, nil)
	if err != nil {
		return nil, fmt.Errorf("setup GET device manifest error: %w", err)
	}
	req.Header.Set("Connection", "close")

	xmlresp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do GET device manifest error: %w", err)
	}
	defer xmlresp.Body.Close()

	var root dmrSchema
	err = xml.NewDecoder(xmlresp.Body).Decode(&root)
	if err != nil {
		return nil, fmt.Errorf("unmarshal device manifest error: %w", err)
	}

	mrDevice := findMediaRenderer(root.Device)
	if mrDevice == nil {
		return nil, errors.New("no MediaRenderer device found")
	}

	mr := &MediaRenderer{
		URL:          dmrurl,
		FriendlyName: mrDevice.FriendlyName,
		ModelName:    mrDevice.ModelName,
	}
	for i := 0; i < len(mrDevice.ServiceList.Services); i++ {
		// normalize service URLs to start with leading /
		service := mrDevice.ServiceList.Services[i]
		if !strings.HasPrefix(service.EventSubURL, "/") {
			service.EventSubURL = "/" + service.EventSubURL
		}
		if !strings.HasPrefix(service.ControlURL, "/") {
			service.ControlURL = "/" + service.ControlURL
		}

		switch service.Type {
		case services.AVTransport:
			mr.avTransportControlURL = parsedURL.Scheme + "://" + parsedURL.Host + service.ControlURL
			mr.avTransportEventSubURL = parsedURL.Scheme + "://" + parsedURL.Host + service.EventSubURL

			if _, err := url.ParseRequestURI(mr.avTransportControlURL); err != nil {
				return nil, fmt.Errorf("invalid AVTransportControlURL: %w", err)
			}

			if _, err = url.ParseRequestURI(mr.avTransportEventSubURL); err != nil {
				return nil, fmt.Errorf("invalid AVTransportEventSubURL: %w", err)
			}
		case services.RenderingControl:
			mr.renderingControlURL = parsedURL.Scheme + "://" + parsedURL.Host + service.ControlURL

			_, err = url.ParseRequestURI(mr.renderingControlURL)
			if err != nil {
				return nil, fmt.Errorf("invalid RenderingControlURL: %w", err)
			}
		case services.ConnectionManager:
			mr.connectionManagerURL = parsedURL.Scheme + "://" + parsedURL.Host + service.ControlURL
			if err != nil {
				return nil, fmt.Errorf("invalid ConnectionManagerURL: %w", err)
			}
		}
	}

	if mr.avTransportControlURL != "" {
		return mr, nil
	}

	return nil, errors.New("wrong DMR")
}

func findMediaRenderer(d device) *device {
	// MediaRenderer can be either at root or nested
	if d.DeviceType == "urn:schemas-upnp-org:device:MediaRenderer:1" {
		return &d
	}

	for i := range d.DeviceList.Devices {
		if mr := findMediaRenderer(d.DeviceList.Devices[i]); mr != nil {
			return mr
		}
	}

	return nil
}
