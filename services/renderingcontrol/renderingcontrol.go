package renderingcontrol

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/dweymouth/go-upnpcast/internal/utils"
)

// Client is a client to the device's RenderingControl service
type Client struct {
	HTTPClient *http.Client
	controlURL string
}

// Should not be used directly. Use device.RenderingControlClient() instead.
func NewClient(controlURL string) *Client {
	return &Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		controlURL: controlURL,
	}
}

// GetMute returns the mute status for our device
func (c *Client) GetMute(ctx context.Context) (string, error) {
	xmlbuilder, err := getMuteSoapBuild()
	if err != nil {
		return "", fmt.Errorf("GetMuteSoapCall build error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.controlURL, bytes.NewReader(xmlbuilder))
	if err != nil {
		return "", fmt.Errorf("GetMuteSoapCall POST error: %w", err)
	}

	req.Header = http.Header{
		"SOAPAction":   []string{`"urn:schemas-upnp-org:service:RenderingControl:1#GetMute"`},
		"content-type": []string{"text/xml"},
		"charset":      []string{"utf-8"},
		"Connection":   []string{"close"},
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("GetMuteSoapCall Do POST error: %w", err)
	}
	defer res.Body.Close()

	var buf bytes.Buffer
	tresp := io.TeeReader(res.Body, &buf)

	var respGetMute getMuteRespBody
	if err = xml.NewDecoder(tresp).Decode(&respGetMute); err != nil {
		return "", fmt.Errorf("GetMuteSoapCall XML Decode error: %w", err)
	}

	return respGetMute.Body.GetMuteResponse.CurrentMute, nil
}

// GetVolume returns the volume level for our device.
func (c *Client) GetVolume(ctx context.Context) (int, error) {
	xmlbuilder, err := getVolumeSoapBuild()
	if err != nil {
		return 0, fmt.Errorf("GetVolumeSoapCall build error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.controlURL, bytes.NewReader(xmlbuilder))
	if err != nil {
		return 0, fmt.Errorf("GetVolumeSoapCall POST error: %w", err)
	}

	req.Header = utils.BuildRequestHeader(`"urn:schemas-upnp-org:service:RenderingControl:1#GetVolume"`)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("GetVolumeSoapCall Do POST error: %w", err)
	}
	defer res.Body.Close()

	var buf bytes.Buffer

	tresp := io.TeeReader(res.Body, &buf)

	var respGetVolume getVolumeRespBody
	if err = xml.NewDecoder(tresp).Decode(&respGetVolume); err != nil {
		return 0, fmt.Errorf("GetVolumeSoapCall XML Decode error: %w", err)
	}

	intVolume, err := strconv.Atoi(respGetVolume.Body.GetVolumeResponse.CurrentVolume)
	if err != nil {
		return 0, fmt.Errorf("GetVolumeSoapCall failed to parse volume value: %w", err)
	}

	if intVolume < 0 {
		intVolume = 0
	}

	return intVolume, nil
}

// SetMute sets the mute status of the device
func (c *Client) SetMute(ctx context.Context, muted bool) error {
	xmlbuilder, err := setMuteSoapBuild(muted)
	if err != nil {
		return fmt.Errorf("SetMuteSoapCall build error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.controlURL, bytes.NewReader(xmlbuilder))
	if err != nil {
		return fmt.Errorf("SetMuteSoapCall POST error: %w", err)
	}

	req.Header = utils.BuildRequestHeader(`"urn:schemas-upnp-org:service:RenderingControl:1#SetMute"`)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("SetMuteSoapCall Do POST error: %w", err)
	}
	defer res.Body.Close()

	if _, err := io.ReadAll(res.Body); err != nil {
		return fmt.Errorf("SetMuteSoapCall Failed to read response: %w", err)
	}

	return nil
}

// SetVolume sets the desired volume level.
func (c *Client) SetVolume(ctx context.Context, vol int) error {
	v := strconv.Itoa(vol)
	xmlbuilder, err := setVolumeSoapBuild(v)
	if err != nil {
		return fmt.Errorf("SetVolumeSoapCall build error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.controlURL, bytes.NewReader(xmlbuilder))
	if err != nil {
		return fmt.Errorf("SetVolumeSoapCall POST error: %w", err)
	}

	req.Header = utils.BuildRequestHeader(`"urn:schemas-upnp-org:service:RenderingControl:1#SetVolume"`)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("SetVolumeSoapCall Do POST error: %w", err)
	}
	defer res.Body.Close()

	return nil
}
