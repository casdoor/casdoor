// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

package idp

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type LinkedInIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

func NewLinkedInIdProvider(clientId string, clientSecret string, redirectUrl string) *LinkedInIdProvider {
	idp := &LinkedInIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

func (idp *LinkedInIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig return a point of Config, which describes a typical 3-legged OAuth2 flow
func (idp *LinkedInIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	var endpoint = oauth2.Endpoint{
		TokenURL: "https://www.linkedIn.com/oauth/v2/accessToken",
	}

	var config = &oauth2.Config{
		Scopes:       []string{"email,public_profile"},
		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

type LinkedInAccessToken struct {
	AccessToken string `json:"access_token"` //Interface call credentials
	ExpiresIn   int64  `json:"expires_in"`   //access_token interface call credential timeout time, unit (seconds)
}

// GetToken use code get access_token (*operation of getting code ought to be done in front)
// get more detail via: https://docs.microsoft.com/en-us/linkedIn/shared/authentication/authorization-code-flow?context=linkedIn%2Fcontext&tabs=HTTPS
func (idp *LinkedInIdProvider) GetToken(code string) (*oauth2.Token, error) {
	params := url.Values{}
	params.Add("grant_type", "authorization_code")
	params.Add("redirect_uri", idp.Config.RedirectURL)
	params.Add("client_id", idp.Config.ClientID)
	params.Add("client_secret", idp.Config.ClientSecret)
	params.Add("code", code)

	accessTokenUrl := fmt.Sprintf("%s?%s", idp.Config.Endpoint.TokenURL, params.Encode())
	bs, _ := json.Marshal(params.Encode())
	req, _ := http.NewRequest("POST", accessTokenUrl, strings.NewReader(string(bs)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	rbs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	tokenResp := LinkedInAccessToken{}
	if err = json.Unmarshal(rbs, &tokenResp); err != nil {
		return nil, err
	}

	token := &oauth2.Token{
		AccessToken: tokenResp.AccessToken,
		TokenType:   "Bearer",
		Expiry:      time.Unix(time.Now().Unix()+tokenResp.ExpiresIn, 0),
	}

	return token, nil
}

/*
{
    "firstName": {
        "localized": {
            "zh_CN": "继坤"
        },
        "preferredLocale": {
            "country": "CN",
            "language": "zh"
        }
    },
    "lastName": {
        "localized": {
            "zh_CN": "刘"
        },
        "preferredLocale": {
            "country": "CN",
            "language": "zh"
        }
    },
    "profilePicture": {
        "displayImage": "urn:li:digitalmediaAsset:C5603AQHbdR8RkG62yg",
        "displayImage~": {
            "paging": {
                "count": 10,
                "start": 0,
                "links": []
            },
            "elements": [
                {
                    "artifact": "urn:li:digitalmediaMediaArtifact:(urn:li:digitalmediaAsset:C5603AQHbdR8RkG62yg,urn:li:digitalmediaMediaArtifactClass:profile-displayphoto-shrink_100_100)",
                    "authorizationMethod": "PUBLIC",
                    "data": {
                        "com.linkedin.digitalmedia.mediaartifact.StillImage": {
                            "mediaType": "image/jpeg",
                            "rawCodecSpec": {
                                "name": "jpeg",
                                "type": "image"
                            },
                            "displaySize": {
                                "width": 100.0,
                                "uom": "PX",
                                "height": 100.0
                            },
                            "storageSize": {
                                "width": 100,
                                "height": 100
                            },
                            "storageAspectRatio": {
                                "widthAspect": 1.0,
                                "heightAspect": 1.0,
                                "formatted": "1.00:1.00"
                            },
                            "displayAspectRatio": {
                                "widthAspect": 1.0,
                                "heightAspect": 1.0,
                                "formatted": "1.00:1.00"
                            }
                        }
                    },
                    "identifiers": [
                        {
                            "identifier": "https://media.licdn.cn/dms/image/C5603AQHbdR8RkG62yg/profile-displayphoto-shrink_100_100/0/1625279434135?e=1630540800&v=beta&t=Z-bQKf_jFv8L1uwr6X5AJLoTQRWZrueT7qrITDSvxWM",
                            "index": 0,
                            "mediaType": "image/jpeg",
                            "file": "urn:li:digitalmediaFile:(urn:li:digitalmediaAsset:C5603AQHbdR8RkG62yg,urn:li:digitalmediaMediaArtifactClass:profile-displayphoto-shrink_100_100,0)",
                            "identifierType": "EXTERNAL_URL",
                            "identifierExpiresInSeconds": 1630540800
                        }
                    ]
                },
				// ...
                }
            ]
        }
    },
    "id": "vvMfLsLIRs"
}
*/

type LinkedInUserInfo struct {
	FirstName struct {
		Localized       map[string]string `json:"localized"`
		PreferredLocale struct {
			Country  string `json:"country"`
			Language string `json:"language"`
		} `json:"preferredLocale"`
	} `json:"firstName"`
	LastName struct {
		Localized       map[string]string `json:"localized"`
		PreferredLocale struct {
			Country  string `json:"country"`
			Language string `json:"language"`
		} `json:"preferredLocale"`
	} `json:"lastName"`
	ProfilePicture struct {
		DisplayImage  string `json:"displayImage"`
		DisplayImage1 struct {
			Paging struct {
				Count int           `json:"count"`
				Start int           `json:"start"`
				Links []interface{} `json:"links"`
			} `json:"paging"`
			Elements []struct {
				Artifact            string `json:"artifact"`
				AuthorizationMethod string `json:"authorizationMethod"`
				Data                struct {
					ComLinkedinDigitalmediaMediaartifactStillImage struct {
						MediaType    string `json:"mediaType"`
						RawCodecSpec struct {
							Name string `json:"name"`
							Type string `json:"type"`
						} `json:"rawCodecSpec"`
						DisplaySize struct {
							Width  float64 `json:"width"`
							Uom    string  `json:"uom"`
							Height float64 `json:"height"`
						} `json:"displaySize"`
						StorageSize struct {
							Width  int `json:"width"`
							Height int `json:"height"`
						} `json:"storageSize"`
						StorageAspectRatio struct {
							WidthAspect  float64 `json:"widthAspect"`
							HeightAspect float64 `json:"heightAspect"`
							Formatted    string  `json:"formatted"`
						} `json:"storageAspectRatio"`
						DisplayAspectRatio struct {
							WidthAspect  float64 `json:"widthAspect"`
							HeightAspect float64 `json:"heightAspect"`
							Formatted    string  `json:"formatted"`
						} `json:"displayAspectRatio"`
					} `json:"com.linkedin.digitalmedia.mediaartifact.StillImage"`
				} `json:"data"`
				Identifiers []struct {
					Identifier                 string `json:"identifier"`
					Index                      int    `json:"index"`
					MediaType                  string `json:"mediaType"`
					File                       string `json:"file"`
					IdentifierType             string `json:"identifierType"`
					IdentifierExpiresInSeconds int    `json:"identifierExpiresInSeconds"`
				} `json:"identifiers"`
			} `json:"elements"`
		} `json:"displayImage~"`
	} `json:"profilePicture"`
	Id string `json:"id"`
}

/*
{
    "handle": "urn:li:emailAddress:3775708763",
    "handle~": {
        "emailAddress": "hsimpson@linkedin.com"
    }
}
*/

type LinkedInUserEmail struct {
	Elements []struct {
		Handle struct {
			EmailAddress string `json:"emailAddress"`
		} `json:"handle~"`
		Handle1 string `json:"handle"`
	} `json:"elements"`
}

// GetUserInfo use LinkedInAccessToken gotten before return LinkedInUserInfo
// get more detail via: https://docs.microsoft.com/en-us/linkedin/consumer/integrations/self-serve/sign-in-with-linkedin?context=linkedin/consumer/context
func (idp *LinkedInIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	var linkedInUserInfo LinkedInUserInfo
	bs, err := idp.GetUrlRespWithAuthorization("https://api.linkedIn.com/v2/me?projection=(id,firstName,lastName,profilePicture(displayImage~:playableStreams))", token.AccessToken)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(bs, &linkedInUserInfo); err != nil {
		return nil, err
	}

	var linkedInUserEmail LinkedInUserEmail
	bs, err = idp.GetUrlRespWithAuthorization("https://api.linkedIn.com/v2/emailAddress?q=members&projection=(elements*(handle~))", token.AccessToken)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(bs, &linkedInUserEmail); err != nil {
		return nil, err
	}

	username := ""
	for _, name := range linkedInUserInfo.FirstName.Localized {
		username += name
	}
	for _, name := range linkedInUserInfo.LastName.Localized {
		username += name
	}
	userInfo := UserInfo{
		Id:          linkedInUserInfo.Id,
		DisplayName: username,
		Username:    username,
		Email:       linkedInUserEmail.Elements[0].Handle.EmailAddress,
		AvatarUrl:   linkedInUserInfo.ProfilePicture.DisplayImage1.Elements[0].Identifiers[0].Identifier,
	}
	return &userInfo, nil
}

func (idp *LinkedInIdProvider) GetUrlRespWithAuthorization(url, token string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bs, nil
}
