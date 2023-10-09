package scim

import (
	"fmt"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/errors"
	"net/http"
)

type UserResourceHandler struct{}

// https://github.com/elimity-com/scim/blob/master/resource_handler_test.go Example in-memory resource handler
// https://datatracker.ietf.org/doc/html/rfc7644#section-3.4 How to query/update resources

func (h UserResourceHandler) Create(r *http.Request, attrs scim.ResourceAttributes) (scim.Resource, error) {
	resource := &scim.Resource{Attributes: attrs}
	err := AddScimUser(resource)
	return *resource, err
}

func (h UserResourceHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	resource, err := GetScimUser(id)
	if err != nil {
		return scim.Resource{}, err
	}
	return *resource, nil
}

func (h UserResourceHandler) Delete(r *http.Request, id string) error {
	owner, name := util.GetOwnerAndNameFromId(id)
	_, err := object.DeleteUser(&object.User{Owner: owner, Name: name})
	return err
}

func (h UserResourceHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	if params.Count == 0 {
		count, err := object.GetGlobalUserCount("", "")
		if err != nil {
			return scim.Page{}, err
		}
		return scim.Page{TotalResults: int(count)}, nil
	}

	resources := make([]scim.Resource, 0)
	// startIndex is 1-based index
	users, err := object.GetPaginationGlobalUsers(params.StartIndex-1, params.Count, "", "", "", "")
	if err != nil {
		return scim.Page{}, err
	}
	for _, user := range users {
		resources = append(resources, *user2resource(user))
	}
	return scim.Page{
		TotalResults: len(resources),
		Resources:    resources,
	}, nil
}

func (h UserResourceHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	user, err := object.GetUser(id)
	if err != nil {
		return scim.Resource{}, err
	}
	if user == nil {
		return scim.Resource{}, errors.ScimErrorResourceNotFound(id)
	}
	return UpdateScimUserByPatchOperation(id, operations)
}

func (h UserResourceHandler) Replace(r *http.Request, id string, attrs scim.ResourceAttributes) (scim.Resource, error) {
	user, err := object.GetUser(id)
	if err != nil {
		return scim.Resource{}, err
	}
	if user == nil {
		return scim.Resource{}, errors.ScimErrorResourceNotFound(id)
	}
	resource := &scim.Resource{Attributes: attrs}
	err = UpdateScimUser(id, resource)
	return *resource, err
}

func GetScimUser(id string) (*scim.Resource, error) {
	user, err := object.GetUser(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	r := user2resource(user)
	return r, nil
}

func AddScimUser(r *scim.Resource) error {
	user, err := resource2user(r.Attributes)
	if err != nil {
		return err
	}
	affect, err := object.AddUser(user)
	if err != nil {
		return err
	}
	if !affect {
		return fmt.Errorf("add user failed")
	}

	r.ID = user.GetId()
	r.ExternalID = buildExternalId(user)
	r.Meta = buildMeta(user)
	return nil
}

func UpdateScimUser(id string, r *scim.Resource) error {
	user, err := resource2user(r.Attributes)
	if err != nil {
		return err
	}
	affect, err := object.UpdateUser(id, user, nil, true)
	if err != nil {
		return err
	}
	if !affect {
		return fmt.Errorf("update user failed")
	}

	r.ID = user.GetId()
	r.ExternalID = buildExternalId(user)
	r.Meta = buildMeta(user)
	return nil
}

func UpdateScimUserByPatchOperation(id string, ops []scim.PatchOperation) (r scim.Resource, err error) {
	user, err := object.GetUser(id)
	if err != nil {
		return scim.Resource{}, err
	}
	if user == nil {
		return scim.Resource{}, errors.ScimErrorResourceNotFound(id)
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("invalid patch op value: %v", r)
		}

	}()

	getValue := func(v interface{}, defaultV interface{}) interface{} {
		if v == nil {
			return defaultV
		}
		return v
	}
	for _, op := range ops {
		value := op.Value
		if op.Op == scim.PatchOperationRemove {
			value = nil
		}
		switch op.Path.String() {
		case "userName":
			user.Name = getValue(value, "").(string)
		case "externalId":
			user.ExternalId = getValue(value, "").(string)
		case "displayName":
			user.DisplayName = getValue(value, "").(string)
		case "profileUrl":
			user.Homepage = getValue(value, "").(string)
		case "userType":
			user.Type = getValue(value, "").(string)
		case "name.givenName":
			user.FirstName = getValue(value, "").(string)
		case "name.familyName":
			user.LastName = getValue(value, "").(string)
		case "name":
			defaultV := map[string]interface{}{"givenName": "", "familyName": ""}
			v := getValue(value, defaultV).(map[string]interface{}) // e.g. {"givenName": "AA", "familyName": "BB"}
			if v["givenName"] != nil {
				user.FirstName = v["givenName"].(string)
			}
			if v["familyName"] != nil {
				user.LastName = v["familyName"].(string)
			}
		case "emails":
			defaultV := []map[string]interface{}{{"value": ""}}
			vs := getValue(value, defaultV).([]map[string]interface{}) // e.g. [{"value": "test@casdoor"}]
			if len(vs) > 0 && vs[0]["value"] != nil {
				user.Email = vs[0]["value"].(string)
			}
		case "phoneNumbers":
			defaultV := []map[string]interface{}{{"value": ""}}
			vs := getValue(value, defaultV).([]map[string]interface{}) // e.g. [{"value": "18750004417"}]
			if len(vs) > 0 && vs[0]["value"] != nil {
				user.Phone = vs[0]["value"].(string)
			}
		case "photos":
			defaultV := []map[string]interface{}{{"value": ""}}
			vs := getValue(value, defaultV).([]map[string]interface{}) // e.g. [{"value": "https://cdn.casbin.org/img/casbin.svg"}]
			if len(vs) > 0 && vs[0]["value"] != nil {
				user.Avatar = vs[0]["value"].(string)
			}
		case "addresses":
			defaultV := []map[string]interface{}{{"locality": "", "region": "", "country": ""}}
			vs := getValue(value, defaultV).([]map[string]interface{}) // e.g. [{"locality": "Hollywood", "region": "CN", "country": "USA"}]
			if len(vs) > 0 && vs[0]["locality"] != nil {
				user.Location = vs[0]["locality"].(string)
			}
			if len(vs) > 0 && vs[0]["region"] != nil {
				user.Region = vs[0]["region"].(string)
			}
			if len(vs) > 0 && vs[0]["country"] != nil {
				user.CountryCode = vs[0]["country"].(string)
			}
		}
	}
	return scim.Resource{}, nil
}
