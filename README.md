<h1 align="center" style="border-bottom: none;">📦⚡️ Casdoor</h1>
<h3 align="center">A UI-first centralized authentication / Single-Sign-On (SSO) platform based on OAuth 2.0 / OIDC.</h3>
<p align="center">
  <a href="#badge">
    <img alt="semantic-release" src="https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg">
  </a>
  <a href="https://hub.docker.com/r/casbin/casdoor">
    <img alt="docker pull casbin/casdoor" src="https://img.shields.io/docker/pulls/casbin/casdoor.svg">
  </a>
  <a href="https://github.com/casdoor/casdoor/actions/workflows/build.yml">
    <img alt="GitHub Workflow Status (branch)" src="https://github.com/casbin/jcasbin/workflows/build/badge.svg?style=flat-square">
  </a>
  <a href="https://github.com/casdoor/casdoor/releases/latest">
    <img alt="GitHub Release" src="https://img.shields.io/github/v/release/casbin/casdoor.svg">
  </a>
  <a href="https://hub.docker.com/repository/docker/casbin/casdoor">
    <img alt="Docker Image Version (latest semver)" src="https://img.shields.io/badge/Docker%20Hub-latest-brightgreen">
  </a>
</p>

<p align="center">
  <a href="https://goreportcard.com/report/github.com/casdoor/casdoor">
    <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/casdoor/casdoor?style=flat-square">
  </a>
  <a href="https://github.com/casdoor/casdoor/blob/master/LICENSE">
    <img src="https://img.shields.io/github/license/casbin/casdoor?style=flat-square" alt="license">
  </a>
  <a href="https://github.com/casdoor/casdoor/issues">
    <img alt="GitHub issues" src="https://img.shields.io/github/issues/casbin/casdoor?style=flat-square">
  </a>
  <a href="#">
    <img alt="GitHub stars" src="https://img.shields.io/github/stars/casbin/casdoor?style=flat-square">
  </a>
  <a href="https://github.com/casdoor/casdoor/network">
    <img alt="GitHub forks" src="https://img.shields.io/github/forks/casbin/casdoor?style=flat-square">
  </a>
  <a href="https://crowdin.com/project/casdoor-site">
    <img alt="Crowdin" src="https://badges.crowdin.net/casdoor-site/localized.svg">
  </a>
  <a href="https://gitter.im/casbin/casdoor">
    <img alt="Gitter" src="https://badges.gitter.im/casbin/casdoor.svg">
  </a>
</p>

## Online demo

Deployed site: https://door.casdoor.com/

## Quick Start
Run your own casdoor program in a few minutes.

### Download

There are two methods, get code via go subcommand `get`:

```shell
go get github.com/casdoor/casdoor
```

  or `git`:

```bash
git clone https://github.com/casdoor/casdoor
```

Finally, change directory:

```bash
cd casdoor/
```

We provide two start up methods for all kinds of users.

### Manual

#### Simple configuration
Casdoor requires a running Relational database to be operational.Thus you need to modify configuration to point out the location of database.

Edit `conf/app.conf`, modify `dataSourceName` to correct database info, which follows this format:

```bash
username:password@tcp(database_ip:database_port)/
```

Then create an empty schema (database) named `casdoor` in your relational database. After the program runs for the first time, it will automatically create tables in this schema.

You can also edit `main.go`, modify `false` to `true`. It will automatically create the schema (database) named `casdoor` in this database.

```bash
createDatabase := flag.Bool("createDatabase", false, "true if you need casdoor to create database")
```

#### Run

Casdoor provides two run modes, the difference is binary size and user prompt.

##### Dev Mode

Edit `conf/app.conf`, set `runmode=dev`. Firstly build front-end files:

```bash
cd web/ && yarn && yarn run start
```
*❗ A word of caution ❗: Casdoor's front-end is built using yarn. You should use `yarn` instead of `npm`. It has a potential failure during building the files if you use `npm`.*

Then build back-end binary file, change directory to root(Relative to casdoor):

```bash
go run main.go
```

That's it! Try to visit http://127.0.0.1:7001/. :small_airplane:  
**But make sure you always request the backend port 8000 when you are using SDKs.**

##### Production Mode

Edit `conf/app.conf`, set `runmode=prod`. Firstly build front-end files:

```bash
cd web/ && yarn && yarn run build
```

Then build back-end binary file, change directory to root(Relative to casdoor):

```bash
go build main.go && sudo ./main
```

> Notice, you should visit back-end port, default 8000. Now try to visit **http://SERVER_IP:8000/**

### Docker

Casdoor provide 2 kinds of image: 
- casbin/casdoor-all-in-one, in which casdoor binary, a mysql database and all necessary configurations are packed up. This image is for new user to have a trial on casdoor quickly. **With this image you can start a casdoor immediately with one single command (or two) without any complex configuration**. **Note: we DO NOT recommend you to use this image in productive environment**

- casbin/casdoor: normal & graceful casdoor image with only casdoor and environment installed. 

This method requires [docker](https://docs.docker.com/get-docker/) and [docker-compose](https://docs.docker.com/compose/install/) to be installed first.

### Start casdoor with casbin/casdoor-all-in-one
if the image is not pulled, pull it from dockerhub
```shell
docker pull casbin/casdoor-all-in-one
```
Start it with
```shell
docker run -p 8000:8000 casbin/casdoor-all-in-one
```
Now you can visit http://localhost:8000 and have a try. Default account and password is 'admin' and '123'. Go for it!

### Start casdoor with casbin/casdoor
#### modify the configurations
For the convenience of your first attempt, docker-compose.yml contains commands to start a database via docker.

Thus edit `conf/app.conf` to point out the location of database(db:3306), modify `dataSourceName` to the fixed content:

```bash
dataSourceName = root:123456@tcp(db:3306)/
```

> If you need to modify `conf/app.conf`, you need to re-run `docker-compose up`.

#### Run

```bash
docker-compose up
```

### K8S
You could use helm to deploy casdoor in k8s. At first, you should modify the [configmap](./manifests/casdoor/templates/configmap.yaml) for your application.
And then run bellow command to deploy it.

```bash
IMG_TAG=latest make deploy 
```

And undeploy it with:
```bash
make undeploy
```

That's it! Try to visit http://localhost:8000/. :small_airplane:

## Detailed documentation

We also provide a complete [document](https://casdoor.org/) as a reference.

## Other examples

These all use casdoor as a centralized authentication platform.

- [Casnode](https://github.com/casbin/casnode): Next-generation forum software based on React + Golang.
- [Casbin-OA](https://github.com/casbin/casbin-oa): A full-featured OA(Office Assistant) system.
- ......

## Contribute

For casdoor, if you have any questions, you can give Issues, or you can also directly start Pull Requests(but we recommend giving issues first to communicate with the community).

### I18n notice

If you are contributing to casdoor, please note that we use [Crowdin](https://crowdin.com/project/casdoor-web) as translating platform and i18next as translating tool. When you add some words using i18next in the ```web/``` directory, please remember to add what you have added to the ```web/src/locales/en/data.json``` file.

## License

 [Apache-2.0](https://github.com/casdoor/casdoor/blob/master/LICENSE)

