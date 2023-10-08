package controllers

import (
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
	"net/http"
	"strings"
)

var (
	scimServer = GetScimServer()
)

func newStringParams(name string, required, unique bool) schema.SimpleParams {
	uniqueness := schema.AttributeUniquenessNone()
	if unique {
		uniqueness = schema.AttributeUniquenessServer()
	}
	return schema.SimpleStringParams(schema.StringParams{
		Name:       name,
		Required:   required,
		Uniqueness: uniqueness,
	})
}

func newComplexParams(name string, required bool, multi bool, subAttributes []schema.SimpleParams) schema.ComplexParams {
	return schema.ComplexParams{
		Name:          name,
		Required:      required,
		MultiValued:   multi,
		SubAttributes: subAttributes,
	}
}

func GetScimServer() scim.Server {
	config := scim.ServiceProviderConfig{
		//DocumentationURI: optional.NewString("www.example.com/scim"),
	}

	stringField := []schema.SimpleParams{
		newStringParams("externalId", false, true),
		newStringParams("userName", true, true),
		newStringParams("displayName", false, false),
		newStringParams("profileUrl", false, false),
		newStringParams("userType", false, false),
		newStringParams("preferredLanguage", false, false),
	}
	complexField := []schema.ComplexParams{
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
	codeAttrs := make([]schema.CoreAttribute, 0, len(stringField)+len(complexField))
	for _, field := range stringField {
		codeAttrs = append(codeAttrs, schema.SimpleCoreAttribute(field))
	}
	for _, field := range complexField {
		codeAttrs = append(codeAttrs, schema.ComplexCoreAttribute(field))
	}

	userSchema := schema.Schema{
		ID:          schema.UserSchema,
		Name:        optional.NewString("User"),
		Description: optional.NewString("User Account"),
		Attributes:  codeAttrs,
	}

	extension := schema.Schema{
		ID:          object.UserExtensionKey,
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

func (c *RootController) HandleScim() {
	path := c.Ctx.Request.URL.Path
	c.Ctx.Request.URL.Path = strings.TrimPrefix(path, "/scim")
	scimServer.ServeHTTP(c.Ctx.ResponseWriter, c.Ctx.Request)
}

type UserResourceHandler struct{}

func (h UserResourceHandler) Create(r *http.Request, attrs scim.ResourceAttributes) (scim.Resource, error) {
	resource := &scim.Resource{Attributes: attrs}
	err := object.AddScimUser(resource)
	return *resource, err
}

func (h UserResourceHandler) Delete(r *http.Request, id string) error {
	owner, name := util.GetOwnerAndNameFromId(id)
	_, err := object.DeleteUser(&object.User{Owner: owner, Name: name})
	return err
}

func (h UserResourceHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	resource, err := object.GetScimUser(id)
	if err != nil {
		return scim.Resource{}, err
	}
	return *resource, nil
}

func (h UserResourceHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	return scim.Page{}, nil
}

func (h UserResourceHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	return scim.Resource{}, nil
}

func (h UserResourceHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	return scim.Resource{}, nil
}
