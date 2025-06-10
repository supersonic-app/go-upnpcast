package connectionmanager

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/supersonic-app/go-upnpcast/internal/utils"
)

type getProtocolInfoEnvelope struct {
	XMLName             xml.Name            `xml:"s:Envelope"`
	Schema              string              `xml:"xmlns:s,attr"`
	Encoding            string              `xml:"s:encodingStyle,attr"`
	GetProtocolInfoBody getProtocolInfoBody `xml:"s:Body"`
}

type getProtocolInfoBody struct {
	XMLName               xml.Name              `xml:"s:Body"`
	GetProtocolInfoAction getProtocolInfoAction `xml:"u:GetProtocolInfo"`
}

type getProtocolInfoAction struct {
	XMLName           xml.Name `xml:"u:GetProtocolInfo"`
	ConnectionManager string   `xml:"xmlns:u,attr"`
}

func getProtocolInfoSoapBuild() (io.Reader, error) {
	d := getProtocolInfoEnvelope{
		XMLName:  xml.Name{},
		Schema:   "http://schemas.xmlsoap.org/soap/envelope/",
		Encoding: "http://schemas.xmlsoap.org/soap/encoding/",
		GetProtocolInfoBody: getProtocolInfoBody{
			XMLName: xml.Name{},
			GetProtocolInfoAction: getProtocolInfoAction{
				XMLName:           xml.Name{},
				ConnectionManager: "urn:schemas-upnp-org:service:ConnectionManager:1",
			},
		},
	}
	b, err := xml.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("getProtocolInfoSoapBuild Marshal error: %w", err)
	}

	return io.MultiReader(strings.NewReader(utils.XMLStart), bytes.NewReader(b)), nil
}
