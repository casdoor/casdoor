package radius

import (
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"log"
)

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
	org := "built-in" // TODO
	password := rfc2865.UserPassword_GetString(r.Packet)

	var (
		code = radius.CodeAccessAccept
		lang = "en"
	)
	_, msg := object.CheckUserPassword(org, username, password, lang)
	if msg != "" {
		code = radius.CodeAccessReject
		w.Write(r.Response(code))
		return
	}
	w.Write(r.Response(code))
}
