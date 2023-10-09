package scim

import (
	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
)

/*
Example json user resource
{
    "schemas": [
        "urn:ietf:params:scim:schemas:core:2.0:User",
        "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"
    ],
    "addresses": [
        {
            "country": "US",
            "locality": "Shanghai",
            "region": "CN"
        }
    ],
    "displayName": "Hello, Scim",
    "name": {
        "familyName": "Bob",
        "givenName": "Alice"
    },
    "phoneNumbers": [
        {
            "value": "46407568879"
        }
    ],
    "photos": [
        {
            "value": "https://cdn.casbin.org/img/casbin.svg"
        }
    ],
    "emails": [
        {
            "value": "cbvdho@example.com"
        }
    ],
    "profileUrl": "https://door.casdoor.com/users/build-in/scim_test_user2",
    "userName": "scim_test_user2",
    "userType": "normal-user",
    "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": {
        "organization": "built-in"
    }
}
*/

const (
	UserExtensionKey = "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"
)

var (
	UserStringField = []schema.SimpleParams{
		newStringParams("externalId", false, true),
		newStringParams("userName", true, true),
		newStringParams("displayName", false, false),
		newStringParams("profileUrl", false, false),
		newStringParams("userType", false, false),
		//newStringParams("preferredLanguage", false, false),
		//newStringParams("password", false, false),
	}
	UserComplexField = []schema.ComplexParams{
		newComplexParams("name", false, false, []schema.SimpleParams{
			newStringParams("givenName", false, false),
			newStringParams("familyName", false, false),
		}),
		newComplexParams("emails", false, true, []schema.SimpleParams{
			newStringParams("value", true, false),
		}),
		newComplexParams("phoneNumbers", false, true, []schema.SimpleParams{
			newStringParams("value", true, false),
		}),
		newComplexParams("photos", false, true, []schema.SimpleParams{
			newStringParams("value", true, false),
		}),
		newComplexParams("addresses", false, true, []schema.SimpleParams{
			newStringParams("locality", false, false),
			newStringParams("region", false, false),
			newStringParams("country", false, false),
		}),
	}
	Server = GetScimServer()
)

func GetScimServer() scim.Server {
	config := scim.ServiceProviderConfig{
		//DocumentationURI: optional.NewString("www.example.com/scim"),
		SupportPatch: true,
	}

	codeAttrs := make([]schema.CoreAttribute, 0, len(UserStringField)+len(UserComplexField))
	for _, field := range UserStringField {
		codeAttrs = append(codeAttrs, schema.SimpleCoreAttribute(field))
	}
	for _, field := range UserComplexField {
		codeAttrs = append(codeAttrs, schema.ComplexCoreAttribute(field))
	}

	userSchema := schema.Schema{
		ID:          schema.UserSchema,
		Name:        optional.NewString("User"),
		Description: optional.NewString("User Account"),
		Attributes:  codeAttrs,
	}

	extension := schema.Schema{
		ID:          UserExtensionKey,
		Name:        optional.NewString("EnterpriseUser"),
		Description: optional.NewString("Enterprise User"),
		Attributes: []schema.CoreAttribute{
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name:     "organization",
				Required: true,
			})),
		},
	}

	resourceTypes := []scim.ResourceType{
		{
			ID:          optional.NewString("User"),
			Name:        "User",
			Endpoint:    "/Users",
			Description: optional.NewString("User Account in Casdoor"),
			Schema:      userSchema,
			SchemaExtensions: []scim.SchemaExtension{
				{Schema: extension},
			},
			Handler: UserResourceHandler{},
		},
	}

	server := scim.Server{
		Config:        config,
		ResourceTypes: resourceTypes,
	}
	return server
}
