package oidc

import (
	"context"
	"github.com/casdoor/casdoor/object"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
)

type ClientStore struct {}

func NewClientStore() *ClientStore {
	return &ClientStore{}
}


// GetByID according to the ID for the client information
func (cs *ClientStore) GetByID(ctx context.Context, id string) (oauth2.ClientInfo, error) {
	client := object.GetClientByID(id)
	return &models.Client{
		ID: client.ID,
		Secret: client.Secret,
		Domain: client.Domain,
		UserID: client.UserID,
	}, nil
}