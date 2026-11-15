// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
	"fmt"
	"io"
	"strings"

	"github.com/casdoor/casdoor/conf"
	"golang.org/x/net/html"
)

// The custom HTML of an organization admin is for layout and branding: text, links, pictures, inline
// SVG and styles. The frontend already strips script from it; what is checked here is the rest an
// admin could turn against the users signing in: fields and forms that collect what they type, and
// files loaded from places the admin can watch.
var safeHtmlTags = map[string]bool{
	"div": true, "span": true, "p": true, "a": true, "img": true, "br": true, "hr": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"ul": true, "ol": true, "li": true, "dl": true, "dt": true, "dd": true,
	"strong": true, "b": true, "em": true, "i": true, "u": true, "s": true, "small": true, "sub": true, "sup": true, "mark": true,
	"blockquote": true, "q": true, "cite": true, "pre": true, "code": true, "kbd": true, "abbr": true, "time": true, "address": true,
	"table": true, "thead": true, "tbody": true, "tfoot": true, "tr": true, "td": true, "th": true, "caption": true, "colgroup": true, "col": true,
	"section": true, "article": true, "header": true, "footer": true, "nav": true, "main": true, "aside": true, "figure": true, "figcaption": true,
	"details": true, "summary": true, "center": true, "font": true, "style": true,
	"svg": true, "g": true, "path": true, "circle": true, "rect": true, "line": true, "polyline": true, "polygon": true, "ellipse": true,
	"defs": true, "lineargradient": true, "radialgradient": true, "stop": true,
}

var unsafeHtmlAttributes = map[string]bool{
	"srcset": true, "ping": true, "action": true, "formaction": true, "srcdoc": true, "lowsrc": true, "dynsrc": true,
	"loading": true, "usemap": true, "ismap": true, "form": true, "autofocus": true, "contenteditable": true,
	"http-equiv": true, "manifest": true, "xlink:href": true, "xml:base": true, "is": true,
}

func isHtmlLinkSafe(link string) bool {
	link = strings.ToLower(strings.TrimSpace(link))
	for _, r := range link {
		if r <= ' ' || r == 0x7f {
			return false
		}
	}
	colon := strings.IndexByte(link, ':')
	if colon == -1 || strings.ContainsAny(link[:colon], "/?#") {
		return true
	}
	scheme := link[:colon]
	return scheme == "http" || scheme == "https" || scheme == "mailto" || scheme == "tel"
}

func isHtmlAttributeSafe(tag string, key string, value string, prefixes []string, extraTags map[string]bool) bool {
	if strings.HasPrefix(key, "on") || unsafeHtmlAttributes[key] {
		return false
	}

	lowerValue := strings.ToLower(strings.TrimSpace(value))
	switch key {
	case "href":
		return tag == "a" && isHtmlLinkSafe(value)
	case "src":
		return (tag == "img" || extraTags[tag]) && isCssUrlAllowed(lowerValue, prefixes, true)
	case "data":
		return tag != "object" || isCssUrlAllowed(lowerValue, prefixes, false)
	case "background", "poster":
		return isCssUrlAllowed(lowerValue, prefixes, true)
	case "style":
		return isCssSafeFor(value, prefixes)
	}
	if strings.Contains(lowerValue, "url(") {
		return isCssSafeFor(value, prefixes)
	}
	return true
}

// getExtraHtmlTags returns the elements a global admin lets organization admins use besides safeHtmlTags,
// set in the global settings
func getExtraHtmlTags() map[string]bool {
	res := map[string]bool{}
	for _, tag := range conf.GetCustomHtmlAllowedTags() {
		res[tag] = true
	}
	return res
}

func isHtmlSafeFor(text string, prefixes []string) bool {
	return getUnsafeHtml(text, prefixes) == ""
}

// getUnsafeHtml names what in text an organization admin may not use, or returns "" when there is nothing
func getUnsafeHtml(text string, prefixes []string) string {
	extraTags := getExtraHtmlTags()
	tokenizer := html.NewTokenizer(strings.NewReader(text))
	inStyle := false
	for {
		switch tokenizer.Next() {
		case html.ErrorToken:
			if tokenizer.Err() == io.EOF {
				return ""
			}
			return "malformed HTML"
		case html.TextToken:
			// the browser reads a style element inside svg as markup, so its text must not hold any
			if inStyle && !isCssSafeFor(string(tokenizer.Raw()), prefixes) {
				return "the CSS in <style>"
			}
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			if !safeHtmlTags[token.Data] && !extraTags[token.Data] {
				return fmt.Sprintf("<%s>", token.Data)
			}
			for _, attribute := range token.Attr {
				key := strings.ToLower(attribute.Key)
				if !isHtmlAttributeSafe(token.Data, key, attribute.Val, prefixes, extraTags) {
					return fmt.Sprintf("the %s attribute of <%s>", key, token.Data)
				}
			}
			inStyle = token.Data == "style"
		case html.EndTagToken:
			token := tokenizer.Token()
			if !safeHtmlTags[token.Data] && !extraTags[token.Data] {
				return fmt.Sprintf("</%s>", token.Data)
			}
			inStyle = false
		default:
			// comments and doctypes end at different places in different parsers, which could hide markup
			return "an HTML comment or doctype"
		}
	}
}
