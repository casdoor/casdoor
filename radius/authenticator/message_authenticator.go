package authenticator

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"errors"
	"fmt"

	layehradius "layeh.com/radius"
	"layeh.com/radius/rfc2869"
)

var (
	ErrMissingMessageAuthenticator  = errors.New("RADIUS Message-Authenticator is missing")
	ErrInvalidMessageAuthenticator  = errors.New("RADIUS Message-Authenticator is invalid")
	ErrInvalidResponseAuthenticator = errors.New("RADIUS Response Authenticator is invalid")
)

func MessageAuthenticatorPresent(packet *layehradius.Packet) bool {
	if packet == nil {
		return false
	}
	for _, attribute := range packet.Attributes {
		if attribute.Type == rfc2869.MessageAuthenticator_Type {
			return true
		}
	}
	return false
}

func SignAccessRequest(packet *layehradius.Packet) error {
	if packet == nil {
		return errors.New("RADIUS packet is nil")
	}
	if packet.Code != layehradius.CodeAccessRequest {
		return fmt.Errorf("cannot sign RADIUS packet with code %d as an Access-Request", packet.Code)
	}
	return setMessageAuthenticator(packet)
}

func VerifyAccessRequest(packet *layehradius.Packet) error {
	if packet == nil {
		return errors.New("RADIUS packet is nil")
	}
	if packet.Code != layehradius.CodeAccessRequest {
		return fmt.Errorf("cannot verify RADIUS packet with code %d as an Access-Request", packet.Code)
	}
	return verifyMessageAuthenticator(packet, packet.Authenticator, packet.Secret)
}

func SignAccessResponse(response, request *layehradius.Packet) error {
	if err := validateAccessResponse(response, request); err != nil {
		return err
	}
	response.Secret = request.Secret
	response.Authenticator = request.Authenticator
	return setMessageAuthenticator(response)
}

func VerifyAccessResponse(response, request *layehradius.Packet) error {
	if err := validateAccessResponse(response, request); err != nil {
		return err
	}
	if err := verifyMessageAuthenticator(response, request.Authenticator, request.Secret); err != nil {
		return err
	}

	responseWire, err := response.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to encode RADIUS response: %w", err)
	}
	requestWire, err := request.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode RADIUS request: %w", err)
	}
	if !layehradius.IsAuthenticResponse(responseWire, requestWire, request.Secret) {
		return ErrInvalidResponseAuthenticator
	}
	return nil
}

func ExchangeWithMessageAuthenticator(ctx context.Context, packet *layehradius.Packet, address string) (*layehradius.Packet, error) {
	if err := SignAccessRequest(packet); err != nil {
		return nil, err
	}

	response, err := layehradius.Exchange(ctx, packet, address)
	if err != nil {
		return nil, err
	}
	if err = VerifyAccessResponse(response, packet); err != nil {
		return nil, err
	}
	return response, nil
}

type messageAuthenticatorResponseWriter struct {
	layehradius.ResponseWriter
	request *layehradius.Packet
}

func NewMessageAuthenticatorResponseWriter(writer layehradius.ResponseWriter, request *layehradius.Packet) layehradius.ResponseWriter {
	requestCopy := *request
	requestCopy.Secret = append([]byte(nil), request.Secret...)
	return &messageAuthenticatorResponseWriter{ResponseWriter: writer, request: &requestCopy}
}

func (writer *messageAuthenticatorResponseWriter) Write(response *layehradius.Packet) error {
	if err := SignAccessResponse(response, writer.request); err != nil {
		return err
	}
	return writer.ResponseWriter.Write(response)
}

func setMessageAuthenticator(packet *layehradius.Packet) error {
	if len(packet.Secret) == 0 {
		return errors.New("RADIUS shared secret is empty")
	}
	if err := rfc2869.MessageAuthenticator_Set(packet, make([]byte, md5.Size)); err != nil {
		return fmt.Errorf("failed to initialize RADIUS Message-Authenticator: %w", err)
	}

	wire, err := packet.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to encode RADIUS packet: %w", err)
	}
	messageAuthenticator := calculateMessageAuthenticator(wire, packet.Secret)
	if err = rfc2869.MessageAuthenticator_Set(packet, messageAuthenticator); err != nil {
		return fmt.Errorf("failed to set RADIUS Message-Authenticator: %w", err)
	}
	return nil
}

func verifyMessageAuthenticator(packet *layehradius.Packet, authenticator [16]byte, secret []byte) error {
	messageAuthenticator, err := getMessageAuthenticator(packet)
	if err != nil {
		return err
	}

	copyPacket := clonePacket(packet)
	copyPacket.Authenticator = authenticator
	copyPacket.Secret = secret
	if err = rfc2869.MessageAuthenticator_Set(copyPacket, make([]byte, md5.Size)); err != nil {
		return fmt.Errorf("failed to clear RADIUS Message-Authenticator: %w", err)
	}
	wire, err := copyPacket.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to encode RADIUS packet: %w", err)
	}
	expected := calculateMessageAuthenticator(wire, secret)
	if !hmac.Equal(messageAuthenticator, expected) {
		return ErrInvalidMessageAuthenticator
	}
	return nil
}

func getMessageAuthenticator(packet *layehradius.Packet) ([]byte, error) {
	var messageAuthenticator []byte
	count := 0
	for _, attribute := range packet.Attributes {
		if attribute.Type != rfc2869.MessageAuthenticator_Type {
			continue
		}
		count++
		messageAuthenticator = layehradius.Bytes(attribute.Attribute)
	}
	if count == 0 {
		return nil, ErrMissingMessageAuthenticator
	}
	if count != 1 || len(messageAuthenticator) != md5.Size {
		return nil, ErrInvalidMessageAuthenticator
	}
	return messageAuthenticator, nil
}

func calculateMessageAuthenticator(packet, secret []byte) []byte {
	mac := hmac.New(md5.New, secret)
	_, _ = mac.Write(packet)
	return mac.Sum(nil)
}

func clonePacket(packet *layehradius.Packet) *layehradius.Packet {
	copyPacket := *packet
	copyPacket.Secret = append([]byte(nil), packet.Secret...)
	copyPacket.Attributes = make(layehradius.Attributes, len(packet.Attributes))
	for i, attribute := range packet.Attributes {
		copyAttribute := *attribute
		copyAttribute.Attribute = append(layehradius.Attribute(nil), attribute.Attribute...)
		copyPacket.Attributes[i] = &copyAttribute
	}
	return &copyPacket
}

func validateAccessResponse(response, request *layehradius.Packet) error {
	if response == nil || request == nil {
		return errors.New("RADIUS request and response are required")
	}
	if request.Code != layehradius.CodeAccessRequest {
		return fmt.Errorf("RADIUS request has unexpected code %d", request.Code)
	}
	switch response.Code {
	case layehradius.CodeAccessAccept, layehradius.CodeAccessReject, layehradius.CodeAccessChallenge:
	default:
		return fmt.Errorf("RADIUS response has unexpected code %d", response.Code)
	}
	if response.Identifier != request.Identifier {
		return errors.New("RADIUS response identifier does not match request")
	}
	if len(request.Secret) == 0 {
		return errors.New("RADIUS shared secret is empty")
	}
	return nil
}
