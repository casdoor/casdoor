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
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

// uploadedFilesPrefix is the URL prefix registered by
// web.SetStaticPath("/files", "files") in main.go.
const uploadedFilesPrefix = "/files"

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

// lookupUploadedFile resolves urlPath to the file that serverStaticRouter()
// would serve for it, mirroring beego's searchFile(): the request path is
// cleaned before it is matched against the "/files" prefix, so a path like
// "/files/../index.html" is not treated as an uploaded file here either.
func lookupUploadedFile(urlPath string) (string, bool) {
	requestPath := filepath.ToSlash(filepath.Clean(urlPath))
	if requestPath != uploadedFilesPrefix && !strings.HasPrefix(requestPath, uploadedFilesPrefix+"/") {
		return "", false
	}

	staticDir, ok := web.BConfig.WebConfig.StaticDir[uploadedFilesPrefix]
	if !ok {
		return "", false
	}

	return path.Join(staticDir, requestPath[len(uploadedFilesPrefix):]), true
}

// UploadedFileFilter hardens the response headers of the "/files" static path,
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
func UploadedFileFilter(ctx *context.Context) {
	filePath, ok := lookupUploadedFile(ctx.Request.URL.Path)
	if !ok {
		return
	}

	// Only harden the responses that serverStaticRouter() actually serves.
	// When the file does not exist, beego skips the static router and the
	// request falls through to StaticFilter(), which serves the frontend
	// "index.html". That response has to keep its original headers, otherwise
	// requesting a missing "/files/xxx.html" would download the Casdoor
	// frontend as an attachment instead of rendering it.
	if _, err := os.Stat(filePath); err != nil {
		return
	}

	header := ctx.ResponseWriter.Header()
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("Content-Security-Policy", "default-src 'none'; sandbox")

	if activeContentExts[strings.ToLower(path.Ext(filePath))] {
		header.Set("Content-Disposition", "attachment")
	}
}
