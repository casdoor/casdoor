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

package routers

import (
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/beego/beego/context"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
)

var (
	oldStaticBaseUrl = "https://cdn.casbin.org"
	newStaticBaseUrl = conf.GetConfigString("staticBaseUrl")
	enableGzip       = conf.GetConfigBool("enableGzip")
)

func StaticFilter(ctx *context.Context) {
	urlPath := ctx.Request.URL.Path

	if urlPath == "/.well-known/acme-challenge/filename" {
		http.ServeContent(ctx.ResponseWriter, ctx.Request, "acme-challenge", time.Now(), strings.NewReader("content"))
	}

	if strings.HasPrefix(urlPath, "/api/") || strings.HasPrefix(urlPath, "/.well-known/") {
		return
	}
	if strings.HasPrefix(urlPath, "/cas") && (strings.HasSuffix(urlPath, "/serviceValidate") || strings.HasSuffix(urlPath, "/proxy") || strings.HasSuffix(urlPath, "/proxyValidate") || strings.HasSuffix(urlPath, "/validate") || strings.HasSuffix(urlPath, "/p3/serviceValidate") || strings.HasSuffix(urlPath, "/p3/proxyValidate") || strings.HasSuffix(urlPath, "/samlValidate")) {
		return
	}

	path := "web/build"
	if urlPath == "/" {
		path += "/index.html"
	} else {
		path += urlPath
	}

	if !util.FileExist(path) {
		path = "web/build/index.html"
	}
	if !util.FileExist(path) {
		return
	}

	if oldStaticBaseUrl == newStaticBaseUrl {
		makeGzipResponse(ctx.ResponseWriter, ctx.Request, path)
	} else {
		serveFileWithReplace(ctx.ResponseWriter, ctx.Request, path)
	}
}

func serveFileWithReplace(w http.ResponseWriter, r *http.Request, name string) {
	f, err := os.Open(filepath.Clean(name))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		panic(err)
	}

	oldContent := util.ReadStringFromPath(name)
	newContent := strings.ReplaceAll(oldContent, oldStaticBaseUrl, newStaticBaseUrl)

	http.ServeContent(w, r, d.Name(), d.ModTime(), strings.NewReader(newContent))
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func makeGzipResponse(w http.ResponseWriter, r *http.Request, path string) {
	if !enableGzip || !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		serveFileWithReplace(w, r, path)
		return
	}
	w.Header().Set("Content-Encoding", "gzip")
	gz := gzip.NewWriter(w)
	defer gz.Close()
	gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
	serveFileWithReplace(gzw, r, path)
}
