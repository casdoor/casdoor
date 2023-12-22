package object

import (
	"fmt"
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

// https://www.cisco.com/c/en/us/td/docs/ios-xml/ios/sec_usr_radatt/configuration/xe-16/sec-usr-radatt-xe-16-book/sec-rad-ov-ietf-attr.html
type RadiusAccounting struct {
	Owner       string    `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string    `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime time.Time `json:"createdTime"`

	Username    string `xorm:"index" json:"username"`
	ServiceType int64  `json:"serviceType"` // e.g. LoginUser (1)

	NasId       string `json:"nasId"`       // String identifying the network access server originating the Access-Request.
	NasIpAddr   string `json:"nasIpAddr"`   // e.g. "192.168.0.10"
	NasPortId   string `json:"nasPortId"`   // Contains a text string which identifies the port of the NAS that is authenticating the user. e.g."eth.0"
	NasPortType int64  `json:"nasPortType"` // Indicates the type of physical port the network access server is using to authenticate the user. e.g.Ethernet（15）
	NasPort     int64  `json:"nasPort"`     // Indicates the physical port number of the network access server that is authenticating the user. e.g. 233

	FramedIpAddr    string `json:"framedIpAddr"`    // Indicates the IP address to be configured for the user by sending the IP address of a user to the RADIUS server.
	FramedIpNetmask string `json:"framedIpNetmask"` // Indicates the IP netmask to be configured for the user when the user is using a device on a network.

	AcctSessionId      string    `xorm:"index" json:"acctSessionId"`
	AcctSessionTime    int64     `json:"acctSessionTime"` // Indicates how long (in seconds) the user has received service.
	AcctInputTotal     int64     `json:"acctInputTotal"`
	AcctOutputTotal    int64     `json:"acctOutputTotal"`
	AcctInputPackets   int64     `json:"acctInputPackets"`   // Indicates how many packets have been received from the port over the course of this service being provided to a framed user.
	AcctOutputPackets  int64     `json:"acctOutputPackets"`  // Indicates how many packets have been sent to the port in the course of delivering this service to a framed user.
	AcctTerminateCause int64     `json:"acctTerminateCause"` // e.g. Lost-Carrier (2)
	LastUpdate         time.Time `json:"lastUpdate"`
	AcctStartTime      time.Time `xorm:"index" json:"acctStartTime"`
	AcctStopTime       time.Time `xorm:"index" json:"acctStopTime"`
}

func (ra *RadiusAccounting) GetId() string {
	return util.GetId(ra.Owner, ra.Name)
}

func getRadiusAccounting(owner, name string) (*RadiusAccounting, error) {
	if owner == "" || name == "" {
		return nil, nil
	}
	ra := RadiusAccounting{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&ra)
	if err != nil {
		return nil, err
	}
	if existed {
		return &ra, nil
	} else {
		return nil, nil
	}
}

func getPaginationRadiusAccounting(owner, field, value, sortField, sortOrder string, offset, limit int) ([]*RadiusAccounting, error) {
	ras := []*RadiusAccounting{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&ras)
	if err != nil {
		return ras, err
	}
	return ras, nil
}

func GetRadiusAccounting(id string) (*RadiusAccounting, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getRadiusAccounting(owner, name)
}

func GetRadiusAccountingBySessionId(sessionId string) (*RadiusAccounting, error) {
	ras, err := getPaginationRadiusAccounting("", "acct_session_id", sessionId, "created_time", "desc", 0, 1)
	if err != nil {
		return nil, err
	}
	if len(ras) == 0 {
		return nil, nil
	}
	return ras[0], nil
}

func AddRadiusAccounting(ra *RadiusAccounting) error {
	_, err := ormer.Engine.Insert(ra)
	return err
}

func DeleteRadiusAccounting(ra *RadiusAccounting) error {
	_, err := ormer.Engine.ID(core.PK{ra.Owner, ra.Name}).Delete(&RadiusAccounting{})
	return err
}

func UpdateRadiusAccounting(id string, ra *RadiusAccounting) error {
	owner, name := util.GetOwnerAndNameFromId(id)
	_, err := ormer.Engine.ID(core.PK{owner, name}).Update(ra)
	return err
}

func InterimUpdateRadiusAccounting(oldRa *RadiusAccounting, newRa *RadiusAccounting, stop bool) error {
	if oldRa.AcctSessionId != newRa.AcctSessionId {
		return fmt.Errorf("AcctSessionId is not equal, newRa = %s, oldRa = %s", newRa.AcctSessionId, oldRa.AcctSessionId)
	}
	oldRa.AcctInputTotal = newRa.AcctInputTotal
	oldRa.AcctOutputTotal = newRa.AcctOutputTotal
	oldRa.AcctInputPackets = newRa.AcctInputPackets
	oldRa.AcctOutputPackets = newRa.AcctOutputPackets
	oldRa.AcctSessionTime = newRa.AcctSessionTime
	if stop {
		oldRa.AcctStopTime = newRa.AcctStopTime
		if oldRa.AcctStopTime.IsZero() {
			oldRa.AcctStopTime = time.Now()
		}
		oldRa.AcctTerminateCause = newRa.AcctTerminateCause
	} else {
		oldRa.LastUpdate = time.Now()
	}

	return UpdateRadiusAccounting(oldRa.GetId(), oldRa)
}
