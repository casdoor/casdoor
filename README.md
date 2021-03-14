Casdoor
====

Casdoor is a UI-first centralized authentication / Single-Sign-On (SSO) platform based on OAuth 2.0 / OIDC.

## Online demo

### Casdoor

Casdoor is the authentication server. It serves both the web UI and the login requests from the application users.

- Deployed site: https://door.casbin.com/
- Source code: https://github.com/casbin/casdoor (this repo)

Global admin login:

- Username: `admin`
- Password: `123`

### Web application

Casbin-OA is one of our applications that use Casdoor as authentication.

- Deployed site: https://oa.casbin.com/
- Source code: https://github.com/casbin/casbin-oa

## Architecture

Casdoor contains 2 parts:

Name | Description | Language | Source code
----|------|----|----
Frontend | Web frontend UI for Casdoor | Javascript + React | https://github.com/casbin/casdoor/tree/master/web
Backend | RESTful API backend for Casdoor | Golang + Beego + MySQL | https://github.com/casbin/casdoor

## Installation

- Get code via `go get`:

    ```shell
    go get github.com/casbin/casdoor
    ```

  or `git clone`:

    ```shell
    git clone https://github.com/casbin/casdoor
    ```

## Run (Dev Environment)

- Run backend (in port 8000):

    ```shell
    go run main.go
    ```

- Run frontend (in the same machine's port 7001):

    ```shell
    cd web
    ## npm
    npm install
    npm run start
    ## yarn
    yarn install
    yarn run start
    ```

- Open browser:

  http://localhost:7001/

## Run (Production Environment)

- build static pages:

  ```
  cd web
  ## npm
  npm run build
  ## yarn
  yarn run build
  ## back to casdoor directory
  cd ..
  ```

- build and run go code:

  ```
  go build
  ./casdoor
  ```

Now, Casdoor is running on port 8000. You can access Casdoor pages directly in your browser, or you can setup a reverse proxy to hold your domain name, SSL, etc.

## Config

- Setup database:

  Casdoor will store its users, nodes and topics informations in a MySQL database named: `casdoor`, will create it if not existed. The DB connection string can be specified at: https://github.com/casbin/casdoor/blob/master/conf/app.conf

    ```ini
    dataSourceName = root:123@tcp(localhost:3306)/
    ```

  Casdoor uses XORM to connect to DB, so all DBs supported by XORM can also be used.

- Setup your Casdoor to enable some third-party login platform:

  Casdoor provide a way to sign up using Google account, Github account, WeChat account and so on,  so you may have to get your own  ClientID and ClientSecret first.

    1. Google

       You could get them by clicking on this url: https://console.developers.google.com/apis
       You should set `Authorized JavaScript origins` to fit your own domain address, for local testing, set`http://localhost:3000`. And set the `Authorized redirect URIs`, the same domain address as before, add `/callback/google/signup` and `/callback/google/link` after that, for local testing, set`http://localhost:3000/callback/google/signup` + `http://localhost:3000/callback/google/link`.

    2. Github

       You could get them by clicking on this url: https://github.com/settings/developers
       You should set `Homepage URL` to fit your own domain address, for local testing, set`http://localhost:3000`. And set the `Authorization callback URL`, the same domain address as before, add `/callback/github` after that, for local testing, set`http://localhost:3000/callback/github`.

  And to improve security, you could set a `state` value determined by **yourself** to make sure the request is requesting by yourself, such as "random".
  Those information strings can be specified at: https://github.com/casbin/casdoor/blob/master/conf/app.conf

    ```ini
    GoogleAuthClientID = "xxx" //your own client id
    GoogleAuthClientSecret = "xxx" //your own client secret
    GoogleAuthState = "xxx" //set by yourself
    GithubAuthClientID = "xxx" //your own client id
    GithubAuthClientSecret = "xxx" //your own client secret
    GithubAuthState = "xx" //set by yourself, we may change this to a random word in the future
    ```

  You may also have to fill in the **same** information at: https://github.com/casbin/casdoor/blob/master/web/src/Conf.js. By the way, you could change the value of `scope` to get different user information form them if you need, we just take `profile` and `email`.

    ```javascript
    export const GoogleClientId  = "xxx"

    export const GoogleAuthState  = "xxx"

    export const GoogleAuthScope  = "profile+email"

    export const GithubClientId  = "xxx"

    export const GithubAuthState  = "xxx"

    export const GithubAuthScope  = "user:email+read:user"
    ```

    3. QQ

       Before you begin to use QQ login services, you should make sure that you have applied the application at [QQ-connect](https://connect.qq.com/manage.html#/)

  Configuration:

    ```javascript
    export const QQClientId  = ""
  
    export const QQAuthState  = ""
  
    export const QQAuthScope  = "get_user_info"
  
    export const QQOauthUri = "https://graph.qq.com/oauth2.0/authorize"
    ```

    ```ini
    QQAPPID = ""
    QQAPPKey = ""
    QQAuthState = ""
    ```

    4. WeChat

       Similar to QQ login service, before using WeChat to log in, you need to apply for OAuth2.0 service fee on the WeChat open platform [open weixin](https://open.weixin.qq.com/cgi-bin/frame?t=home/web_tmpl). After completing the configuration, you can log in via WeChat QR code.

  Configuration:

    ```javascript
    export const WechatClientId  = ""

    export const WeChatAuthState = ""

    export const WeChatAuthScope = "snsapi_login"

    export const WeChatOauthUri = "https://open.weixin.qq.com/connect/qrconnect"
    ```

    ```ini
    WeChatAPPID = ""
    WeChatKey = ""
    WeChatAuthState = ""
    ```

  We would show different login/signup methods depending on your configuration.

- Github corner

  We added a Github icon in the upper right corner, linking to your Github repository address.
  You could set `ShowGithubCorner` to hidden it.

  Configuration:

    ```javascript
  export const ShowGithubCorner = true

  export const GithubRepo = "https://github.com/casbin/casdoor" //your github repository
    ```

- OSS conf

  We use an OSS to store and provide user avatars. You must modify the file `conf/oss.conf` to tell the backend your OSS info. For OSS providers, we support Aliyun(`[aliyun]`), awss3(`[s3]`) now.

  ```
  [provider]
  accessid = id
  accesskey = key
  bucket = bucket
  endpoint = endpoint
  ```

  Please fill out this conf correctly, or the avatar server won't work!

