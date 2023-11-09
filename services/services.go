package services

type ServiceType string

const (
	AVTransport       ServiceType = "urn:schemas-upnp-org:service:AVTransport:1"
	ConnectionManager ServiceType = "urn:schemas-upnp-org:service:ConnectionManager:1"
	RenderingControl  ServiceType = "urn:schemas-upnp-org:service:RenderingControl:1"
)
