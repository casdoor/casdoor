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

package routers

import (
	"path"
	"strings"

	"github.com/beego/beego/v2/server/web/context"
)

const uploadedFilesPrefix = "/files/"

// activeContentExts lists the extensions a browser renders as an *active
// document* (one that can run script) when the URL is opened directly.
// A resource uploaded through /api/upload-resource is served from the public
// "/files" static path, so without these headers an .svg containing
// <script> executes in the origin of the Casdoor server itself and can drive
// the whole API with the session cookie of whoever opens the link.
var activeContentExts = map[string]bool{
	".htm":   true,
	".html":  true,
	".mht":   true,
	".mhtml": true,
	".svg":   true,
	".svgz":  true,
	".swf":   true,
	".xht":   true,
	".xhtml": true,
	".xml":   true,
	".xsl":   true,
	".xslt":  true,
}

// StaticFilesFilter hardens the response headers of the "/files" static path,
// which serves user-uploaded resources.
//
// It has to be registered at web.BeforeStatic, i.e. before serverStaticRouter()
// hands the file to http.ServeContent(): ServeContent only infers a
// Content-Type when the header is not set yet, so whatever is set here wins.
//
// The headers are chosen so that uploads keep working as images (<img src>
// ignores both CSP and Content-Disposition, so avatars still render) while a
// direct navigation to an uploaded file can no longer touch the Casdoor origin:
//
//   - "sandbox" puts the document in an opaque origin and, without
//     "allow-scripts", disables script execution altogether.
//   - "nosniff" stops the browser from upgrading e.g. a .txt upload to
//     text/html by content sniffing.
//   - "Content-Disposition: attachment" makes the active content types above
//     download instead of render, for browsers that do not support CSP sandbox.
func StaticFilesFilter(ctx *context.Context) {
	urlPath := ctx.Request.URL.Path
	if !strings.HasPrefix(urlPath, uploadedFilesPrefix) {
		return
	}

	header := ctx.ResponseWriter.Header()
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("Content-Security-Policy", "default-src 'none'; sandbox")

	if activeContentExts[strings.ToLower(path.Ext(urlPath))] {
		header.Set("Content-Disposition", "attachment")
	}
}
