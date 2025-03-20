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

	"github.com/supersonic-app/go-upnpcast/internal/utils"
)

type Client struct {
	HTTPClient *http.Client
	controlURL string
}

// MediaItem represents a media item to be rendered by the device.
type MediaItem struct {
	// URL of the media item. Required.
	URL string

	SubtitlesURL string
	Title        string
	ContentType  string
	Seekable     bool
	Duration     time.Duration
}

// TransportInfo is the information returned by GetTransportInfo
type TransportInfo struct {
	Status string
	State  string
	Speed  string
}

// PositionInfo is the duration and current playback position of the current media item,
// returned by GetPositionInfo
type PositionInfo struct {
	Duration time.Duration
	RelTime  time.Duration
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

func (a *Client) Seek(ctx context.Context, relSecs int) error {
	time, err := utils.SecondsToClockTime(relSecs)
	if err != nil {
		return err
	}
	xml, err := seekSoapBuild(time)
	if err != nil {
		return fmt.Errorf("SeekSoapCall action error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.controlURL, bytes.NewReader(xml))
	if err != nil {
		return fmt.Errorf("SeekSoapCall POST error: %w", err)
	}
	req.Header = utils.BuildRequestHeader(`"urn:schemas-upnp-org:service:AVTransport:1#Seek"`)

	res, err := a.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("SeekSoapCall Do POST error: %w", err)
	}
	defer res.Body.Close()

	_, err = io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("SeekSoapCall Failed to read response: %w", err)
	}
	_, err = json.Marshal(res.Header)
	if err != nil {
		return fmt.Errorf("SeekSoapCall Response Marshaling error: %w", err)
	}

	return nil

}

func (a *Client) SetAVTransportMedia(ctx context.Context, media *MediaItem) error {
	soapCall, err := setAVTransportSoapBuild(media)
	if err != nil {
		return fmt.Errorf("SetAVTransportMedia build error: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", a.controlURL, bytes.NewReader(soapCall))
	if err != nil {
		return fmt.Errorf("SetAVTransportMedia POST error: %w", err)
	}
	req.Header = utils.BuildRequestHeader(`"urn:schemas-upnp-org:service:AVTransport:1#SetAVTransportURI"`)
	res, err := a.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("SetAVTransportMedia Do POST error: %w", err)
	}
	defer res.Body.Close()

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("SetAVTransportMedia Failed to read response: %w", err)
	}

	var resp setAVTransportURIResponse
	if err := xml.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("SetAVTransportMedia Failed to unmarshal response: %w", err)
	}

	return nil
}

func (a *Client) SetNextAVTransportMedia(ctx context.Context, media *MediaItem) error {
	soapCall, err := setNextAVTransportSoapBuild(media)
	if err != nil {
		return fmt.Errorf("SetNextAVTransportMedia build error: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", a.controlURL, bytes.NewReader(soapCall))
	if err != nil {
		return fmt.Errorf("SetNextAVTransportMedia POST error: %w", err)
	}

	req.Header = utils.BuildRequestHeader(`"urn:schemas-upnp-org:service:AVTransport:1#SetNextAVTransportURI"`)
	res, err := a.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("SetNextAVTransportMedia Do POST error: %w", err)
	}
	defer res.Body.Close()

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("SetNextAVTransportMedia Failed to read response: %w", err)
	}

	var resp setNextAVTransportURIResponse
	if err := xml.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("SetNextAVTransportMedia Failed to unmarshal response: %w", err)
	}

	return nil
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
	req.Header = utils.BuildRequestHeader(`"urn:schemas-upnp-org:service:AVTransport:1#GetTransportInfo"`)

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

func (a *Client) GetPositionInfo(ctx context.Context) (PositionInfo, error) {
	xmlRequest, err := getPositionInfoSoapBuild()
	if err != nil {
		return PositionInfo{}, fmt.Errorf("GetPositionInfo build error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.controlURL, bytes.NewReader(xmlRequest))
	if err != nil {
		return PositionInfo{}, fmt.Errorf("GetPositionInfo POST error: %w", err)
	}
	req.Header = utils.BuildRequestHeader(`"urn:schemas-upnp-org:service:AVTransport:1#GetPositionInfo"`)

	res, err := a.HTTPClient.Do(req)
	if err != nil {
		return PositionInfo{}, fmt.Errorf("GetPositionInfo Do POST error: %w", err)
	}
	defer res.Body.Close()

	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return PositionInfo{}, fmt.Errorf("GetPositionInfo Failed to read response: %w", err)
	}

	var respPositionInfo getPositionInfoResponse
	if err := xml.Unmarshal(resBytes, &respPositionInfo); err != nil {
		return PositionInfo{}, fmt.Errorf("GetPositionInfo Failed to unmarshal response: %w", err)
	}

	r := respPositionInfo.Body.GetPositionInfoResponse
	dur, err := utils.ParseDuration(r.TrackDuration)
	rel, err2 := utils.ParseDuration(r.RelTime)
	if err2 != nil && err == nil {
		err = err2
	}

	return PositionInfo{Duration: dur, RelTime: rel}, err
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

	req.Header = utils.BuildRequestHeader(`"urn:schemas-upnp-org:service:AVTransport:1#` + action + `"`)

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
