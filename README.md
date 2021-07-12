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

## Run through Docker
- Install Docker and Docker-compose,you see [docker](https://docs.docker.com/get-docker/) and [docker-compose](https://docs.docker.com/compose/install/)
- vi casdoor/conf/app.conf
- Modify dataSourceName = root:123@tcp(localhost:3306)/ to dataSourceName = root:123@tcp(db:3306)/
- Execute the following command
  ```shell
  docker-compose up
  ```
- Open browser:

  http://localhost:8000/

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

- Setup database (MySQL):

  Casdoor will store its users, nodes and topics informations in a MySQL database named: `casdoor`, will create it if not existed. The DB connection string can be specified at: https://github.com/casbin/casdoor/blob/master/conf/app.conf

    ```ini
  db = mysql
  dataSourceName = root:123@tcp(localhost:3306)/
  dbName = casdoor
    ```

- Setup database (Postgres):

  Since we must choose a database when opening Postgres with xorm, you should prepare a database manually before running Casdoor. Let's assume that you have already prepared a database called `casdoor`, then you should specify `app.conf` like this:

  ``` ini
  db = postgres
  dataSourceName = "user=postgres password=xxx sslmode=disable dbname="
  dbName = casdoor
  ```

  **Please notice:** You can add Postgres parameters in `dataSourceName`, but please make sure that `dataSourceName` ends with `dbname=`. Or database adapter may crash when you launch Casdoor.

  Casdoor uses XORM to connect to DB, so all DBs supported by XORM can also be used.

- Github corner

  We added a Github icon in the upper right corner, linking to your Github repository address.
  You could set `ShowGithubCorner` to hidden it.

  Configuration (`web/src/commo/Conf.js`):

    ```javascript
  export const ShowGithubCorner = true

  export const GithubRepo = "https://github.com/casbin/casdoor" //your github repository
    ```

- OSS conf

  We use an OSS to store and provide user avatars. You must modify the file `conf/oss.conf` to tell the backend your OSS info. For OSS providers, we support Aliyun(`[aliyun]`), awss3(`[s3]`) now.

  ```
  [provider]
  accessId = id
  accessKey = key
  bucket = bucket
  endpoint = endpoint
  ```

  Please fill out this conf correctly, or the avatar server won't work!

