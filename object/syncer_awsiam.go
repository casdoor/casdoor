// Copyright 2025 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package object

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/casdoor/casdoor/util"
)

// AwsIamSyncerProvider implements SyncerProvider for AWS IAM API-based syncers
type AwsIamSyncerProvider struct {
	Syncer *Syncer
	client *iam.Client
}

// InitAdapter initializes the AWS IAM syncer
func (p *AwsIamSyncerProvider) InitAdapter() error {
	// AWS IAM syncer doesn't need persistent database adapter
	// The IAM client will be created on-demand with credentials
	return nil
}

// GetOriginalUsers retrieves all users from AWS IAM API
func (p *AwsIamSyncerProvider) GetOriginalUsers() ([]*OriginalUser, error) {
	return p.getAwsIamOriginalUsers()
}

// AddUser adds a new user to AWS IAM (not supported for read-only API)
func (p *AwsIamSyncerProvider) AddUser(user *OriginalUser) (bool, error) {
	// AWS IAM syncer is typically read-only
	return false, fmt.Errorf("adding users to AWS IAM is not supported")
}

// UpdateUser updates an existing user in AWS IAM (not supported for read-only API)
func (p *AwsIamSyncerProvider) UpdateUser(user *OriginalUser) (bool, error) {
	// AWS IAM syncer is typically read-only
	return false, fmt.Errorf("updating users in AWS IAM is not supported")
}

// TestConnection tests the AWS IAM API connection
func (p *AwsIamSyncerProvider) TestConnection() error {
	client, err := p.getIamClient()
	if err != nil {
		return err
	}

	// Test connection by listing users with a limit of 1
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	input := &iam.ListUsersInput{
		MaxItems: aws.Int32(1),
	}

	_, err = client.ListUsers(ctx, input)
	return err
}

// Close closes any open connections (no-op for AWS IAM API-based syncer)
func (p *AwsIamSyncerProvider) Close() error {
	// AWS IAM syncer doesn't maintain persistent connections
	p.client = nil
	return nil
}

// getIamClient creates and returns an AWS IAM client
func (p *AwsIamSyncerProvider) getIamClient() (*iam.Client, error) {
	if p.client != nil {
		return p.client, nil
	}

	// syncer.User should be the AWS Access Key ID
	// syncer.Password should be the AWS Secret Access Key
	// syncer.Host should be the AWS region (e.g., "us-east-1")

	accessKeyId := p.Syncer.User
	if accessKeyId == "" {
		return nil, fmt.Errorf("AWS Access Key ID (user field) is required for AWS IAM syncer")
	}

	secretAccessKey := p.Syncer.Password
	if secretAccessKey == "" {
		return nil, fmt.Errorf("AWS Secret Access Key (password field) is required for AWS IAM syncer")
	}

	region := p.Syncer.Host
	if region == "" {
		// Default to us-east-1 if not specified
		region = "us-east-1"
	}

	// Create AWS config with static credentials
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyId,
			secretAccessKey,
			"", // session token (empty for IAM users)
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create IAM client
	p.client = iam.NewFromConfig(cfg)
	return p.client, nil
}

// getAwsIamUsers gets all users from AWS IAM API
func (p *AwsIamSyncerProvider) getAwsIamUsers() ([]types.User, error) {
	client, err := p.getIamClient()
	if err != nil {
		return nil, err
	}

	allUsers := []types.User{}
	var marker *string

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Paginate through all users
	for {
		input := &iam.ListUsersInput{
			Marker: marker,
		}

		output, err := client.ListUsers(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to list IAM users: %w", err)
		}

		allUsers = append(allUsers, output.Users...)

		// Check if there are more users to fetch
		if !output.IsTruncated {
			break
		}
		marker = output.Marker
	}

	return allUsers, nil
}

// awsIamUserToOriginalUser converts AWS IAM user to Casdoor OriginalUser
func (p *AwsIamSyncerProvider) awsIamUserToOriginalUser(iamUser *types.User) *OriginalUser {
	user := &OriginalUser{
		Id:          aws.ToString(iamUser.UserId),
		Name:        aws.ToString(iamUser.UserName),
		DisplayName: aws.ToString(iamUser.UserName),
		Address:     []string{},
		Properties:  map[string]string{},
		Groups:      []string{},
	}

	// Set ARN as a property for reference
	if iamUser.Arn != nil {
		user.Properties["arn"] = aws.ToString(iamUser.Arn)
	}

	// Set path as a property
	if iamUser.Path != nil {
		user.Properties["path"] = aws.ToString(iamUser.Path)
	}

	// Set creation date
	if iamUser.CreateDate != nil {
		user.CreatedTime = iamUser.CreateDate.Format(time.RFC3339)
	} else {
		user.CreatedTime = util.GetCurrentTime()
	}

	// AWS IAM users don't have built-in email/phone fields
	// These would need to be fetched from user tags if needed
	// For now, we'll leave them empty

	return user
}

// getAwsIamOriginalUsers is the main entry point for AWS IAM syncer
func (p *AwsIamSyncerProvider) getAwsIamOriginalUsers() ([]*OriginalUser, error) {
	// Get all users from AWS IAM
	iamUsers, err := p.getAwsIamUsers()
	if err != nil {
		return nil, err
	}

	// Convert AWS IAM users to Casdoor OriginalUser
	originalUsers := []*OriginalUser{}
	for _, iamUser := range iamUsers {
		originalUser := p.awsIamUserToOriginalUser(&iamUser)
		originalUsers = append(originalUsers, originalUser)
	}

	return originalUsers, nil
}
