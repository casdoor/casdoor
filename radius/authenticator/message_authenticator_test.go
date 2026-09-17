package authenticator

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

func newAccessRequest(t *testing.T) *radius.Packet {
	t.Helper()
	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	if err := rfc2865.UserName_SetString(packet, "alice"); err != nil {
		t.Fatal(err)
	}
	if err := rfc2865.UserPassword_SetString(packet, "password"); err != nil {
		t.Fatal(err)
	}
	return packet
}

func TestAccessRequestMessageAuthenticator(t *testing.T) {
	packet := newAccessRequest(t)
	if MessageAuthenticatorPresent(packet) {
		t.Fatal("unsigned request unexpectedly has a Message-Authenticator")
	}
	if err := SignAccessRequest(packet); err != nil {
		t.Fatal(err)
	}
	if !MessageAuthenticatorPresent(packet) {
		t.Fatal("signed request is missing Message-Authenticator")
	}
	if err := VerifyAccessRequest(packet); err != nil {
		t.Fatalf("signed request was rejected: %v", err)
	}

	if err := rfc2865.UserName_SetString(packet, "mallory"); err != nil {
		t.Fatal(err)
	}
	if err := VerifyAccessRequest(packet); !errors.Is(err, ErrInvalidMessageAuthenticator) {
		t.Fatalf("tampered request returned %v, want %v", err, ErrInvalidMessageAuthenticator)
	}
}

func TestAccessRequestRequiresOneMessageAuthenticator(t *testing.T) {
	packet := newAccessRequest(t)
	if err := VerifyAccessRequest(packet); !errors.Is(err, ErrMissingMessageAuthenticator) {
		t.Fatalf("unsigned request returned %v, want %v", err, ErrMissingMessageAuthenticator)
	}

	if err := SignAccessRequest(packet); err != nil {
		t.Fatal(err)
	}
	if err := rfc2869.MessageAuthenticator_Add(packet, make([]byte, 16)); err != nil {
		t.Fatal(err)
	}
	if err := VerifyAccessRequest(packet); !errors.Is(err, ErrInvalidMessageAuthenticator) {
		t.Fatalf("duplicate Message-Authenticator returned %v, want %v", err, ErrInvalidMessageAuthenticator)
	}
}

func TestAccessResponseMessageAuthenticator(t *testing.T) {
	request := newAccessRequest(t)
	if err := SignAccessRequest(request); err != nil {
		t.Fatal(err)
	}
	response := request.Response(radius.CodeAccessAccept)
	if err := SignAccessResponse(response, request); err != nil {
		t.Fatal(err)
	}

	wire, err := response.Encode()
	if err != nil {
		t.Fatal(err)
	}
	requestWire, err := request.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !radius.IsAuthenticResponse(wire, requestWire, request.Secret) {
		t.Fatal("response authenticator is invalid")
	}
	parsed, err := radius.Parse(wire, request.Secret)
	if err != nil {
		t.Fatal(err)
	}
	if err = VerifyAccessResponse(parsed, request); err != nil {
		t.Fatalf("signed response was rejected: %v", err)
	}

	if err = rfc2865.ReplyMessage_SetString(parsed, "tampered"); err != nil {
		t.Fatal(err)
	}
	if err = VerifyAccessResponse(parsed, request); !errors.Is(err, ErrInvalidMessageAuthenticator) {
		t.Fatalf("tampered response returned %v, want %v", err, ErrInvalidMessageAuthenticator)
	}
}

func TestAccessResponseRequiresMessageAuthenticator(t *testing.T) {
	request := newAccessRequest(t)
	response := request.Response(radius.CodeAccessReject)
	if err := VerifyAccessResponse(response, request); !errors.Is(err, ErrMissingMessageAuthenticator) {
		t.Fatalf("unsigned response returned %v, want %v", err, ErrMissingMessageAuthenticator)
	}
}

func TestAccessResponseRejectsInvalidResponseAuthenticator(t *testing.T) {
	request := newAccessRequest(t)
	if err := SignAccessRequest(request); err != nil {
		t.Fatal(err)
	}
	response := request.Response(radius.CodeAccessAccept)
	if err := SignAccessResponse(response, request); err != nil {
		t.Fatal(err)
	}
	wire, err := response.Encode()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := radius.Parse(wire, request.Secret)
	if err != nil {
		t.Fatal(err)
	}
	parsed.Authenticator[0] ^= 0xff
	if err = VerifyAccessResponse(parsed, request); !errors.Is(err, ErrInvalidResponseAuthenticator) {
		t.Fatalf("invalid Response Authenticator returned %v, want %v", err, ErrInvalidResponseAuthenticator)
	}
}

func TestExchangeWithMessageAuthenticatorRequiresSignedResponse(t *testing.T) {
	for _, test := range []struct {
		name         string
		signResponse bool
		wantError    error
	}{
		{name: "signed response", signResponse: true},
		{name: "unsigned response", wantError: ErrMissingMessageAuthenticator},
	} {
		t.Run(test.name, func(t *testing.T) {
			connection, err := net.ListenPacket("udp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer connection.Close()

			server := &radius.PacketServer{
				SecretSource: radius.StaticSecretSource([]byte("secret")),
				Handler: radius.HandlerFunc(func(writer radius.ResponseWriter, request *radius.Request) {
					if err := VerifyAccessRequest(request.Packet); err != nil {
						t.Errorf("server rejected signed request: %v", err)
						return
					}
					response := request.Response(radius.CodeAccessAccept)
					if test.signResponse {
						if err := SignAccessResponse(response, request.Packet); err != nil {
							t.Errorf("failed to sign response: %v", err)
							return
						}
					}
					if err := writer.Write(response); err != nil {
						t.Errorf("failed to write response: %v", err)
					}
				}),
			}
			go func() {
				_ = server.Serve(connection)
			}()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_, err = ExchangeWithMessageAuthenticator(ctx, newAccessRequest(t), connection.LocalAddr().String())
			if !errors.Is(err, test.wantError) {
				t.Fatalf("ExchangeWithMessageAuthenticator() returned %v, want %v", err, test.wantError)
			}
		})
	}
}

type recordingResponseWriter struct {
	wire []byte
}

func (writer *recordingResponseWriter) Write(packet *radius.Packet) error {
	wire, err := packet.Encode()
	if err != nil {
		return err
	}
	writer.wire = wire
	return nil
}

func TestMessageAuthenticatorResponseWriterPreservesRequest(t *testing.T) {
	request := newAccessRequest(t)
	if err := SignAccessRequest(request); err != nil {
		t.Fatal(err)
	}
	requestWire, err := request.Encode()
	if err != nil {
		t.Fatal(err)
	}
	originalRequest, err := radius.Parse(requestWire, request.Secret)
	if err != nil {
		t.Fatal(err)
	}

	recorder := &recordingResponseWriter{}
	writer := NewMessageAuthenticatorResponseWriter(recorder, request)
	request.Code = radius.CodeAccessChallenge
	if err = rfc2865.ReplyMessage_SetString(request, "enter OTP"); err != nil {
		t.Fatal(err)
	}
	if err = writer.Write(request); err != nil {
		t.Fatal(err)
	}

	response, err := radius.Parse(recorder.wire, request.Secret)
	if err != nil {
		t.Fatal(err)
	}
	if !radius.IsAuthenticResponse(recorder.wire, requestWire, request.Secret) {
		t.Fatal("response authenticator is invalid")
	}
	if err = VerifyAccessResponse(response, originalRequest); err != nil {
		t.Fatalf("Message-Authenticator is invalid: %v", err)
	}
}
