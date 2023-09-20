package radius

import (
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"log"
)

// https://support.huawei.com/enterprise/zh/doc/EDOC1000178159/35071f9a#tab_3
func StartRadiusServer() {
	server := radius.PacketServer{
		Addr:         "0.0.0.0:" + conf.GetConfigString("radiusServerPort"),
		Handler:      radius.HandlerFunc(handlerRadius),
		SecretSource: radius.StaticSecretSource([]byte(`secret`)),
	}
	log.Printf("Starting Radius server on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("StartRadiusServer() failed, err = %v", err)
	}
}

func handlerRadius(w radius.ResponseWriter, r *radius.Request) {
	switch r.Code {
	case radius.CodeAccessRequest:
		handleAccessRequest(w, r)
	default:
		log.Printf("radius message, code = %d", r.Code)
	}
}

func handleAccessRequest(w radius.ResponseWriter, r *radius.Request) {
	username := rfc2865.UserName_GetString(r.Packet)
	password := rfc2865.UserPassword_GetString(r.Packet)
	organization := parseOrganization(r.Packet)
	code := radius.CodeAccessAccept

	log.Printf("username=%v, password=%v, code=%v, org=%v", username, password, code, organization)
	if organization == "" {
		code = radius.CodeAccessReject
		w.Write(r.Response(code))
		return
	}
	_, msg := object.CheckUserPassword(organization, username, password, "en")
	if msg != "" {
		code = radius.CodeAccessReject
		w.Write(r.Response(code))
		return
	}
	w.Write(r.Response(code))
}
