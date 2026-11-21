// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package ee

import (
	"time"

	"github.com/casdoor/casdoor/audit"
	"github.com/casdoor/casdoor/object"
)

func init() {
	object.RegisterTables(
		new(SigninTrace),
		new(McpApproval),
		new(ProvisionedAccount),
		new(ProvisioningTask),
		new(ConsentLog),
		new(CertLoginTicket),
		new(PrivilegedAccount),
		new(AccessReview),
		new(PrivilegedRequest),
		new(AccessReviewItem),
	)

	object.EncryptUserFieldsHook = EncryptUserFields
	object.DecryptUserFieldsHook = DecryptUserFields
	object.KeepMaskedAccountSecretsHook = keepMaskedAccountSecrets
	object.PopulateHsmContentHook = populateHsmContent
	object.GetHsmSecretHook = getHsmSecret
	object.GetJwtSigningKeyHook = getJwtSigningKey
	object.GetJwtVerifyKeyHook = getJwtVerifyKey
	object.SignJwtTokenHook = signJwtToken
	object.NewCertXmlSigningContextHook = newCertXmlSigningContext
	object.GetPreviousJsonWebKeyHook = getPreviousJsonWebKey
	object.GetCachedSecurityPolicyHook = GetCachedSecurityPolicy
	object.GetEnterprisePolicyHook = GetEnterprisePolicy
	object.CheckPasswordSecurityPolicyHook = checkPasswordSecurityPolicy
	object.CheckSmsSecurityPolicyHook = checkSmsSecurityPolicy
	object.CheckSmsVolumeHook = checkSmsVolume
	object.AddChainedRecordHook = addChainedRecord
	object.ConsumePushChallengeHook = consumePushChallenge
	object.CheckProviderEditionHook = CheckProviderEdition
	object.MergePaidIntegrationsHook = mergePaidIntegrations
	object.NewPluginSyncerProviderHook = func(syncer *object.Syncer) object.SyncerProvider {
		return &PluginSyncerProvider{Syncer: syncer}
	}
	object.GetUserByWechatUnionIdHook = GetUserByWechatUnionId
	object.CheckProvisioningConfigHook = CheckProvisioningConfig
	object.ClearProvisioningApplicationCacheHook = ClearProvisioningApplicationCache
	object.IsProvisionedUserUpdateHook = isProvisionedUserUpdate
	object.ProvisionUserHook = provisionUser
	object.RequestOrganizationProvisioningHook = requestOrganizationProvisioning
	object.IsTrustedOrganizationHook = isTrustedOrganization

	audit.FormatSiemEventHook = func(format string, event *audit.Event, eventTime time.Time) string {
		if format == "CEF" {
			return formatCef(event, eventTime)
		}
		return formatLeef(event, eventTime)
	}
}
