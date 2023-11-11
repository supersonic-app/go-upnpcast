package avtransport

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	HTTPClient *http.Client
	controlURL string
}

// Should not be used directly. Use device.AVTransportClient() instead.
func NewClient(controlURL, eventSubURL string) *Client {
	return &Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		controlURL: controlURL,
	}
}

func (a *Client) Play(ctx context.Context) error {
	return a.playPauseStopSoapCall(ctx, "Play")
}

func (a *Client) Pause(ctx context.Context) error {
	return a.playPauseStopSoapCall(ctx, "Pause")
}

func (a *Client) Stop(ctx context.Context) error {
	return a.playPauseStopSoapCall(ctx, "Stop")
}

type TransportInfo struct {
	Status string
	State  string
	Speed  string
}

// GetTransportInfo
func (a *Client) GetTransportInfo(ctx context.Context) (TransportInfo, error) {
	xmlbuilder, err := getTransportInfoSoapBuild()
	if err != nil {
		return TransportInfo{}, fmt.Errorf("GetTransportInfo build error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.controlURL, bytes.NewReader(xmlbuilder))
	if err != nil {
		return TransportInfo{}, fmt.Errorf("GetTransportInfo POST error: %w", err)
	}
	req.Header = http.Header{
		"SOAPAction":   []string{`"urn:schemas-upnp-org:service:AVTransport:1#GetTransportInfo"`},
		"content-type": []string{"text/xml"},
		"charset":      []string{"utf-8"},
		"Connection":   []string{"close"},
	}

	res, err := a.HTTPClient.Do(req)
	if err != nil {
		return TransportInfo{}, fmt.Errorf("GetTransportInfo Do POST error: %w", err)
	}
	defer res.Body.Close()

	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return TransportInfo{}, fmt.Errorf("GetTransportInfo Failed to read response: %w", err)
	}

	var respTransportInfo getTransportInfoResponse

	if err := xml.Unmarshal(resBytes, &respTransportInfo); err != nil {
		return TransportInfo{}, fmt.Errorf("GetTransportInfo Failed to unmarshal response: %w", err)
	}

	r := respTransportInfo.Body.GetTransportInfoResponse
	info := TransportInfo{
		Status: r.CurrentTransportStatus,
		State:  r.CurrentTransportState,
		Speed:  r.CurrentSpeed,
	}

	return info, nil
}

func (a *Client) playPauseStopSoapCall(ctx context.Context, action string) error {
	var xml []byte
	var err error

	switch action {
	case "Play":
		xml, err = playSoapBuild()
	case "Stop":
		xml, err = stopSoapBuild()
	case "Pause":
		xml, err = pauseSoapBuild()
	}
	if err != nil {
		return fmt.Errorf("AVTransportActionSoapCall action error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.controlURL, bytes.NewReader(xml))
	if err != nil {
		return fmt.Errorf("AVTransportActionSoapCall POST error: %w", err)
	}

	req.Header = http.Header{
		"SOAPAction":   []string{`"urn:schemas-upnp-org:service:AVTransport:1#` + action + `"`},
		"content-type": []string{"text/xml"},
		"charset":      []string{"utf-8"},
		"Connection":   []string{"close"},
	}

	res, err := a.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("AVTransportActionSoapCall Do POST error: %w", err)
	}
	defer res.Body.Close()

	_, err = io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("AVTransportActionSoapCall Failed to read response: %w", err)
	}

	_, err = json.Marshal(res.Header)
	if err != nil {
		return fmt.Errorf("AVTransportActionSoapCall Response Marshaling error: %w", err)
	}

	return nil
}
