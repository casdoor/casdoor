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
	"net/http"
	"os"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/casdoor/casdoor/util"
)

var (
	oldStaticBaseUrl = "https://cdn.casbin.org"
	newStaticBaseUrl = beego.AppConfig.String("staticBaseUrl")
)

func StaticFilter(ctx *context.Context) {
	urlPath := ctx.Request.URL.Path
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

	if oldStaticBaseUrl == newStaticBaseUrl {
		http.ServeFile(ctx.ResponseWriter, ctx.Request, path)
	} else {
		serveFileWithReplace(ctx.ResponseWriter, ctx.Request, path, oldStaticBaseUrl, newStaticBaseUrl)
	}
}

func serveFileWithReplace(w http.ResponseWriter, r *http.Request, name string, old string, new string) {
	f, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		panic(err)
	}

	oldContent := util.ReadStringFromPath(name)
	newContent := strings.ReplaceAll(oldContent, old, new)

	http.ServeContent(w, r, d.Name(), d.ModTime(), strings.NewReader(newContent))
	_, err = w.Write([]byte(newContent))
	if err != nil {
		panic(err)
	}
}
