// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/casdoor/casdoor/util"
	"golang.org/x/net/html"
)

type Link struct {
	Rel   string
	Sizes string
	Href  string
}

func GetFaviconUrl(htmlStr string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return "", err
	}

	var links []Link
	findLinks(doc, &links)

	if len(links) == 0 {
		return "", fmt.Errorf("no Favicon links found")
	}

	chosenLink := chooseFaviconLink(links)
	if chosenLink == nil {
		return "", fmt.Errorf("unable to determine favicon URL")
	}

	return chosenLink.Href, nil
}

func findLinks(n *html.Node, links *[]Link) {
	if n.Type == html.ElementNode && n.Data == "link" {
		link := parseLink(n)
		if link != nil {
			*links = append(*links, *link)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		findLinks(c, links)
	}
}

func parseLink(n *html.Node) *Link {
	var link Link

	for _, attr := range n.Attr {
		switch attr.Key {
		case "rel":
			link.Rel = attr.Val
		case "sizes":
			link.Sizes = attr.Val
		case "href":
			link.Href = attr.Val
		}
	}

	if link.Href != "" {
		return &link
	}

	return nil
}

func chooseFaviconLink(links []Link) *Link {
	var appleTouchLinks []Link
	var shortcutLinks []Link
	var iconLinks []Link

	for _, link := range links {
		switch link.Rel {
		case "apple-touch-icon":
			appleTouchLinks = append(appleTouchLinks, link)
		case "shortcut icon":
			shortcutLinks = append(shortcutLinks, link)
		case "icon":
			iconLinks = append(iconLinks, link)
		}
	}

	if len(appleTouchLinks) > 0 {
		return chooseFaviconLinkBySizes(appleTouchLinks)
	}

	if len(shortcutLinks) > 0 {
		return chooseFaviconLinkBySizes(shortcutLinks)
	}

	if len(iconLinks) > 0 {
		return chooseFaviconLinkBySizes(iconLinks)
	}

	return nil
}

func chooseFaviconLinkBySizes(links []Link) *Link {
	if len(links) == 1 {
		return &links[0]
	}

	var chosenLink *Link

	for _, link := range links {
		link := link
		if chosenLink == nil || compareSizes(link.Sizes, chosenLink.Sizes) > 0 {
			chosenLink = &link
		}
	}

	return chosenLink
}

func compareSizes(sizes1, sizes2 string) int {
	if sizes1 == sizes2 {
		return 0
	}

	size1 := parseSize(sizes1)
	size2 := parseSize(sizes2)

	if size1 == nil {
		return -1
	}

	if size2 == nil {
		return 1
	}

	if size1[0] == size2[0] {
		return size1[1] - size2[1]
	}

	return size1[0] - size2[0]
}

func parseSize(sizes string) []int {
	size := strings.Split(sizes, "x")
	if len(size) != 2 {
		return nil
	}

	var result []int

	for _, s := range size {
		val := strings.TrimSpace(s)
		if len(val) > 0 {
			num := 0
			for i := 0; i < len(val); i++ {
				if val[i] >= '0' && val[i] <= '9' {
					num = num*10 + int(val[i]-'0')
				} else {
					break
				}
			}
			result = append(result, num)
		}
	}

	if len(result) == 2 {
		return result
	}

	return nil
}

var publicEmailDomains = map[string]int{
	"gmail.com":      1,
	"163.com":        1,
	"qq.com":         1,
	"yahoo.com":      1,
	"hotmail.com":    1,
	"outlook.com":    1,
	"icloud.com":     1,
	"mail.com":       1,
	"aol.com":        1,
	"live.com":       1,
	"yandex.com":     1,
	"yahoo.co.jp":    1,
	"yahoo.co.in":    1,
	"yahoo.co.uk":    1,
	"me.com":         1,
	"msn.com":        1,
	"comcast.net":    1,
	"sbcglobal.net":  1,
	"verizon.net":    1,
	"earthlink.net":  1,
	"cox.net":        1,
	"rediffmail.com": 1,
	"in.com":         1,
	"hotmail.co.uk":  1,
	"hotmail.fr":     1,
	"zoho.com":       1,
	"gmx.com":        1,
	"gmx.de":         1,
	"gmx.net":        1,
}

func isPublicEmailDomain(domain string) bool {
	_, exists := publicEmailDomains[domain]
	return exists
}

func getFaviconFileBuffer(client *http.Client, email string) (*bytes.Buffer, string, error) {
	tokens := strings.Split(email, "@")
	domain := tokens[1]
	if isPublicEmailDomain(domain) {
		return nil, "", nil
	}

	htmlUrl := fmt.Sprintf("https://%s", domain)
	buffer, _, err := downloadImage(client, htmlUrl)
	if err != nil {
		return nil, "", err
	}

	faviconUrl := ""
	if buffer != nil {
		faviconUrl, err = GetFaviconUrl(buffer.String())
		if err != nil {
			fmt.Printf("getFaviconFileBuffer() error, faviconUrl is empty, error = %s\n", err.Error())
		} else {
			if !strings.HasPrefix(faviconUrl, "http") {
				faviconUrl = util.UrlJoin(htmlUrl, faviconUrl)
			}
		}
	}

	if faviconUrl == "" {
		faviconUrl = fmt.Sprintf("https://%s/favicon.ico", domain)
	}
	return downloadImage(client, faviconUrl)
}
