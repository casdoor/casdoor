package object

import (
	"fmt"
	"github.com/casdoor/casdoor/util"
	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/optional"
	"log"
)

/*
Example json user resource

	{
	  "schemas":[
	     "urn:ietf:params:scim:schemas:core:2.0:User"
	  ],
	  "id":"3cc032f5-2361-417f-9e2f-bc80adddf4a3",
	  "meta":{
	     "resourceType":"User",
	     "created":"2019-11-20T13:09:00",
	     "lastModified":"2019-11-20T13:09:00",
	     "location":"https://identity.imulab.io/Users/3cc032f5-2361-417f-9e2f-bc80adddf4a3",
	     "version":"W/\"1\""
	  },
	  "userName":"imulab",
	  "name":{
	     "formatted":"Mr. Weinan Qiu",
	     "familyName":"Qiu",
	     "givenName":"Weinan",
	     "honorificPrefix":"Mr."
	  },
	  "displayName":"Weinan",
	  "profileUrl":"https://identity.imulab.io/profiles/3cc032f5-2361-417f-9e2f-bc80adddf4a3",
	  "userType":"Employee",
	  "preferredLanguage":"zh_CN",
	  "locale":"zh_CN",
	  "timezone":"Asia/Shanghai",
	  "active":true,
	  "emails":[
	     {
	        "value":"imulab@foo.com",
	        "type":"work",
	        "primary":true,
	        "display":"imulab@foo.com"
	     },
	     {
	        "value":"imulab@bar.com",
	        "type":"home",
	        "display":"imulab@bar.com"
	     }
	  ],
	  "phoneNumbers":[
	     {
	        "value":"123-45678",
	        "type":"work",
	        "primary":true,
	        "display":"123-45678"
	     },
	     {
	        "value":"123-45679",
	        "type":"work",
	        "display":"123-45679"
	     }
	  ]
	}
*/

const (
	UserExtensionKey = "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"
)

func buildExternalId(user *User) optional.String {
	if user.ExternalId != "" {
		return optional.NewString(user.ExternalId)
	} else {
		return optional.String{}
	}
}

func buildMeta(user *User) scim.Meta {
	createdTime := util.String2Time(user.CreatedTime)
	updatedTime := util.String2Time(user.UpdatedTime)
	return scim.Meta{
		Created:      &createdTime,
		LastModified: &updatedTime,
		Version:      user.GetId(),
	}
}

func getAttrString(attrs scim.ResourceAttributes, key string) string {
	if attrs[key] == nil {
		return ""
	} else {
		return attrs[key].(string)
	}
}

func getAttrJson(attrs scim.ResourceAttributes, key string) scim.ResourceAttributes {
	if attrs[key] == nil {
		return nil
	} else {
		if v, ok := attrs[key].(map[string]interface{}); ok {
			return v
		} else if v, ok := attrs[key].([]interface{}); ok {
			if len(v) > 0 {
				return v[0].(map[string]interface{})
			} else {
				return nil
			}
		} else {
			panic("invalid attribute type")
		}
	}
}

func getAttrJsonValue(attrs scim.ResourceAttributes, key1 string, key2 string) string {
	attr := getAttrJson(attrs, key1)
	if attr == nil {
		return ""
	} else {
		return getAttrString(attr, key2)
	}
}

func user2resource(user *User) *scim.Resource {
	attrs := make(map[string]interface{})
	// Singular attributes
	attrs["userName"] = user.Name
	attrs["name"] = scim.ResourceAttributes{
		"formatted":  fmt.Sprintf("%v %v", user.FirstName, user.LastName),
		"familyName": user.LastName,
		"givenName":  user.FirstName,
	}
	attrs["displayName"] = user.DisplayName
	attrs["nickName"] = user.DisplayName
	attrs["userType"] = user.Type
	attrs["profileUrl"] = user.Homepage
	attrs["preferredLanguage"] = user.Language
	//attrs["locale"] = language.Make(user.Region).String() // e.g. zh_CN
	attrs["active"] = !user.IsForbidden && !user.IsDeleted
	// Multi-Valued attributes
	attrs["emails"] = []scim.ResourceAttributes{
		{
			"value": user.Email,
		},
	}
	attrs["phoneNumbers"] = []scim.ResourceAttributes{
		{
			"value": user.Phone,
		},
	}
	attrs["photos"] = []scim.ResourceAttributes{
		{
			"value": user.Avatar,
		},
	}
	attrs["addresses"] = []scim.ResourceAttributes{
		{
			"locality": user.Location, // City of residence
			"region":   user.Region,   // e.g. CN
			"country":  user.CountryCode,
		},
	}
	// Enterprise user schema extension
	attrs[UserExtensionKey] = scim.ResourceAttributes{
		"organization": user.Owner,
	}

	return &scim.Resource{
		ID:         user.GetId(),
		ExternalID: buildExternalId(user),
		Attributes: attrs,
		Meta:       buildMeta(user),
	}
}

func resource2user(attrs scim.ResourceAttributes) (user *User, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic in resource2user(): %v", r)
			err = fmt.Errorf("%v", r)
		}
	}()

	user = &User{
		ExternalId:  getAttrString(attrs, "externalId"),
		Name:        getAttrString(attrs, "userName"),
		DisplayName: getAttrString(attrs, "displayName"),
		Homepage:    getAttrString(attrs, "profileUrl"),
		Language:    getAttrString(attrs, "preferredLanguage"),
		Type:        getAttrString(attrs, "userType"),

		Owner:       getAttrJsonValue(attrs, UserExtensionKey, "organization"),
		FirstName:   getAttrJsonValue(attrs, "name", "givenName"),
		LastName:    getAttrJsonValue(attrs, "name", "familyName"),
		Email:       getAttrJsonValue(attrs, "emails", "value"),
		Phone:       getAttrJsonValue(attrs, "phoneNumbers", "value"),
		Avatar:      getAttrJsonValue(attrs, "photos", "value"),
		Location:    getAttrJsonValue(attrs, "addresses", "locality"),
		Region:      getAttrJsonValue(attrs, "addresses", "region"),
		CountryCode: getAttrJsonValue(attrs, "addresses", "country"),

		CreatedTime: util.GetCurrentTime(),
	}

	return
}

func GetScimUser(id string) (*scim.Resource, error) {
	user, err := GetUser(id)
	if err != nil {
		return nil, err
	}
	r := user2resource(user)
	return r, nil
}

func AddScimUser(r *scim.Resource) error {
	user, err := resource2user(r.Attributes)
	if err != nil {
		return err
	}
	affect, err := AddUser(user)
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
