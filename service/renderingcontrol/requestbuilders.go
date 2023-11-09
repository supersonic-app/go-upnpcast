package renderingcontrol

import (
	"encoding/xml"
	"fmt"
)

type setMuteEnvelope struct {
	XMLName     xml.Name    `xml:"s:Envelope"`
	Schema      string      `xml:"xmlns:s,attr"`
	Encoding    string      `xml:"s:encodingStyle,attr"`
	SetMuteBody setMuteBody `xml:"s:Body"`
}

type setMuteBody struct {
	XMLName       xml.Name      `xml:"s:Body"`
	SetMuteAction setMuteAction `xml:"u:SetMute"`
}

type setMuteAction struct {
	XMLName          xml.Name `xml:"u:SetMute"`
	RenderingControl string   `xml:"xmlns:u,attr"`
	InstanceID       string
	Channel          string
	DesiredMute      string
}

type getMuteEnvelope struct {
	XMLName     xml.Name    `xml:"s:Envelope"`
	Schema      string      `xml:"xmlns:s,attr"`
	Encoding    string      `xml:"s:encodingStyle,attr"`
	GetMuteBody getMuteBody `xml:"s:Body"`
}

type getMuteBody struct {
	XMLName       xml.Name      `xml:"s:Body"`
	GetMuteAction getMuteAction `xml:"u:GetMute"`
}

type getMuteAction struct {
	XMLName          xml.Name `xml:"u:GetMute"`
	RenderingControl string   `xml:"xmlns:u,attr"`
	InstanceID       string
	Channel          string
}

type getVolumeEnvelope struct {
	XMLName       xml.Name      `xml:"s:Envelope"`
	Schema        string        `xml:"xmlns:s,attr"`
	Encoding      string        `xml:"s:encodingStyle,attr"`
	GetVolumeBody getVolumeBody `xml:"s:Body"`
}

type getVolumeBody struct {
	XMLName         xml.Name        `xml:"s:Body"`
	GetVolumeAction getVolumeAction `xml:"u:GetVolume"`
}

type getVolumeAction struct {
	XMLName          xml.Name `xml:"u:GetVolume"`
	RenderingControl string   `xml:"xmlns:u,attr"`
	InstanceID       string
	Channel          string
}

type setVolumeEnvelope struct {
	XMLName       xml.Name      `xml:"s:Envelope"`
	Schema        string        `xml:"xmlns:s,attr"`
	Encoding      string        `xml:"s:encodingStyle,attr"`
	SetVolumeBody setVolumeBody `xml:"s:Body"`
}

type setVolumeBody struct {
	XMLName         xml.Name        `xml:"s:Body"`
	SetVolumeAction setVolumeAction `xml:"u:SetVolume"`
}

type setVolumeAction struct {
	XMLName          xml.Name `xml:"u:SetVolume"`
	RenderingControl string   `xml:"xmlns:u,attr"`
	InstanceID       string
	Channel          string
	DesiredVolume    string
}

func setMuteSoapBuild(muted bool) ([]byte, error) {
	m := "0"
	if muted {
		m = "1"
	}

	d := setMuteEnvelope{
		XMLName:  xml.Name{},
		Schema:   "http://schemas.xmlsoap.org/soap/envelope/",
		Encoding: "http://schemas.xmlsoap.org/soap/encoding/",
		SetMuteBody: setMuteBody{
			XMLName: xml.Name{},
			SetMuteAction: setMuteAction{
				XMLName:          xml.Name{},
				RenderingControl: "urn:schemas-upnp-org:service:RenderingControl:1",
				InstanceID:       "0",
				Channel:          "Master",
				DesiredMute:      m,
			},
		},
	}
	xmlStart := []byte(`<?xml version="1.0" encoding="utf-8"?>`)
	b, err := xml.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("setMuteSoapBuild Marshal error: %w", err)
	}

	return append(xmlStart, b...), nil
}

func getMuteSoapBuild() ([]byte, error) {
	d := getMuteEnvelope{
		XMLName:  xml.Name{},
		Schema:   "http://schemas.xmlsoap.org/soap/envelope/",
		Encoding: "http://schemas.xmlsoap.org/soap/encoding/",
		GetMuteBody: getMuteBody{
			XMLName: xml.Name{},
			GetMuteAction: getMuteAction{
				XMLName:          xml.Name{},
				RenderingControl: "urn:schemas-upnp-org:service:RenderingControl:1",
				InstanceID:       "0",
				Channel:          "Master",
			},
		},
	}
	xmlStart := []byte(`<?xml version="1.0" encoding="utf-8"?>`)
	b, err := xml.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("getMuteSoapBuild Marshal error: %w", err)
	}

	return append(xmlStart, b...), nil
}

func getVolumeSoapBuild() ([]byte, error) {
	d := getVolumeEnvelope{
		XMLName:  xml.Name{},
		Schema:   "http://schemas.xmlsoap.org/soap/envelope/",
		Encoding: "http://schemas.xmlsoap.org/soap/encoding/",
		GetVolumeBody: getVolumeBody{
			XMLName: xml.Name{},
			GetVolumeAction: getVolumeAction{
				XMLName:          xml.Name{},
				RenderingControl: "urn:schemas-upnp-org:service:RenderingControl:1",
				InstanceID:       "0",
				Channel:          "Master",
			},
		},
	}
	xmlStart := []byte(`<?xml version="1.0" encoding="utf-8"?>`)
	b, err := xml.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("getVolumeSoapBuild Marshal error: %w", err)
	}

	return append(xmlStart, b...), nil
}

func setVolumeSoapBuild(v string) ([]byte, error) {
	d := setVolumeEnvelope{
		XMLName:  xml.Name{},
		Schema:   "http://schemas.xmlsoap.org/soap/envelope/",
		Encoding: "http://schemas.xmlsoap.org/soap/encoding/",
		SetVolumeBody: setVolumeBody{
			XMLName: xml.Name{},
			SetVolumeAction: setVolumeAction{
				XMLName:          xml.Name{},
				RenderingControl: "urn:schemas-upnp-org:service:RenderingControl:1",
				InstanceID:       "0",
				Channel:          "Master",
				DesiredVolume:    v,
			},
		},
	}
	xmlStart := []byte(`<?xml version="1.0" encoding="utf-8"?>`)
	b, err := xml.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("setVolumeSoapBuild Marshal error: %w", err)
	}

	return append(xmlStart, b...), nil
}
