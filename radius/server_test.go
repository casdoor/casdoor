package radius

import (
	"context"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"testing"
)

func TestAccessRequestRejected(t *testing.T) {
	packet := radius.New(radius.CodeAccessRequest, []byte(`secret`))
	rfc2865.UserName_SetString(packet, "tim")
	rfc2865.UserPassword_SetString(packet, "12345")
	response, err := radius.Exchange(context.Background(), packet, "localhost:1812")
	if err != nil {
		t.Fatal(err)
	}
	if response.Code != radius.CodeAccessReject {
		t.Fatalf("Expected %v, got %v", radius.CodeAccessReject, response.Code)
	}
}

func TestAccessRequestAccepted(t *testing.T) {
	packet := radius.New(radius.CodeAccessRequest, []byte(`secret`))
	rfc2865.UserName_SetString(packet, "admin")
	rfc2865.UserPassword_SetString(packet, "123")
	response, err := radius.Exchange(context.Background(), packet, "localhost:1812")
	if err != nil {
		t.Fatal(err)
	}
	if response.Code != radius.CodeAccessAccept {
		t.Fatalf("Expected %v, got %v", radius.CodeAccessAccept, response.Code)
	}
}
