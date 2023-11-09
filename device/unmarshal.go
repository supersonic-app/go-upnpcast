package device

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/dweymouth/go-upnpcast/services"
)

type dmrSchema struct {
	XMLName xml.Name `xml:"root"`
	Device  struct {
		XMLName      xml.Name `xml:"device"`
		FriendlyName string   `xml:"friendlyName"`
		ModelName    string   `xml:"modelName"`
		ServiceList  struct {
			XMLName  xml.Name `xml:"serviceList"`
			Services []struct {
				XMLName     xml.Name `xml:"service"`
				Type        string   `xml:"serviceType"`
				ID          string   `xml:"serviceId"`
				ControlURL  string   `xml:"controlURL"`
				EventSubURL string   `xml:"eventSubURL"`
			} `xml:"service"`
		} `xml:"serviceList"`
	} `xml:"device"`
}

func mediaRendererFromDeviceURL(ctx context.Context, dmrurl string) (*MediaRenderer, error) {
	var root dmrSchema

	parsedURL, err := url.Parse(dmrurl)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("device URL parse error: %w", err)
	}

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", dmrurl, nil)
	if err != nil {
		return nil, fmt.Errorf("setup GET device manifest error: %w", err)
	}
	req.Header.Set("Connection", "close")

	xmlresp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do GET device manifest error: %w", err)
	}
	defer xmlresp.Body.Close()

	xmlbody, err := io.ReadAll(xmlresp.Body)
	if err != nil {
		return nil, fmt.Errorf("read device manifest error: %w", err)
	}

	err = xml.Unmarshal(xmlbody, &root)
	if err != nil {
		return nil, fmt.Errorf("unmarshal device manifest error: %w", err)
	}

	mr := &MediaRenderer{
		URL:          dmrurl,
		FriendlyName: root.Device.FriendlyName,
		ModelName:    root.Device.ModelName,
	}
	for i := 0; i < len(root.Device.ServiceList.Services); i++ {
		// normalize service URLs to start with leading /
		service := root.Device.ServiceList.Services[i]
		if !strings.HasPrefix(service.EventSubURL, "/") {
			service.EventSubURL = "/" + service.EventSubURL
		}
		if !strings.HasPrefix(service.ControlURL, "/") {
			service.ControlURL = "/" + service.ControlURL
		}

		switch service.Type {
		case string(services.AVTransport):
			mr.avTransportControlURL = parsedURL.Scheme + "://" + parsedURL.Host + service.ControlURL
			mr.avTransportEventSubURL = parsedURL.Scheme + "://" + parsedURL.Host + service.EventSubURL

			if _, err := url.ParseRequestURI(mr.avTransportControlURL); err != nil {
				return nil, fmt.Errorf("invalid AVTransportControlURL: %w", err)
			}

			if _, err = url.ParseRequestURI(mr.avTransportEventSubURL); err != nil {
				return nil, fmt.Errorf("invalid AVTransportEventSubURL: %w", err)
			}
		case string(services.RenderingControl):
			mr.renderingControlURL = parsedURL.Scheme + "://" + parsedURL.Host + service.ControlURL

			_, err = url.ParseRequestURI(mr.renderingControlURL)
			if err != nil {
				return nil, fmt.Errorf("invalid RenderingControlURL: %w", err)
			}
		case string(services.ConnectionManager):
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
