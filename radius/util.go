package radius

import (
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

const (
	OrganizationVendorID = uint32(100)
)

func parseOrganization(p *radius.Packet) string {
	for _, avp := range p.Attributes {
		if avp.Type == rfc2865.VendorSpecific_Type {
			attr := avp.Attribute
			vendorId, value, err := radius.VendorSpecific(attr)
			if err != nil {
				return ""
			}
			if vendorId == OrganizationVendorID {
				return string(value)
			}
		}
	}
	return ""
}
