// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package ee

import (
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// isTrustedOrganization lets the providers of an organization that a global admin trusts reach the intranet
func isTrustedOrganization(name string) bool {
	if !edition.Has(edition.Commercial) {
		return false
	}
	organization, err := object.GetOrganization(util.GetId("admin", name))
	return err == nil && organization != nil && organization.IsTrusted
}
