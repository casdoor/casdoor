<h1 align="center" style="border-bottom: none;">📦⚡️ Casdoor</h1>
<h3 align="center">A UI-first centralized authentication / Single-Sign-On (SSO) platform based on OAuth 2.0 / OIDC.</h3>
<p align="center">
  <a href="#badge">
    <img alt="semantic-release" src="https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg">
  </a>
  <a href="https://hub.docker.com/r/casbin/casdoor">
    <img alt="docker pull casbin/casdoor" src="https://img.shields.io/docker/pulls/casbin/casdoor.svg">
  </a>
  <a href="https://github.com/casbin/casdoor/actions/workflows/build.yml">
    <img alt="GitHub Workflow Status (branch)" src="https://github.com/casbin/jcasbin/workflows/build/badge.svg?style=flat-square">
  </a>
  <a href="https://github.com/casbin/casdoor/releases/latest">
    <img alt="GitHub Release" src="https://img.shields.io/github/v/release/casbin/casdoor.svg">
  </a>
  <a href="https://hub.docker.com/repository/docker/casbin/casdoor">
    <img alt="Docker Image Version (latest semver)" src="https://img.shields.io/badge/Docker%20Hub-latest-brightgreen">
  </a>
</p>

<p align="center">
  <a href="https://goreportcard.com/report/github.com/casbin/casdoor">
    <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/casbin/casdoor?style=flat-square">
  </a>
  <a href="https://github.com/casbin/casdoor/blob/master/LICENSE">
    <img src="https://img.shields.io/github/license/casbin/casdoor?style=flat-square" alt="license">
  </a>
  <a href="https://github.com/casbin/casdoor/issues">
    <img alt="GitHub issues" src="https://img.shields.io/github/issues/casbin/casdoor?style=flat-square">
  </a>
  <a href="#">
    <img alt="GitHub stars" src="https://img.shields.io/github/stars/casbin/casdoor?style=flat-square">
  </a>
  <a href="https://github.com/casbin/casdoor/network">
    <img alt="GitHub forks" src="https://img.shields.io/github/forks/casbin/casdoor?style=flat-square">
  </a>
</p>

## Online demo

Deployed site: https://door.casbin.com/

## Quick Start

Run your own casdoor program in a few minutes:smiley:

### Download

There are two methods, get code via go subcommand `get`:

```shell
go get github.com/casbin/casdoor
```

  or `git`:

```bash
git clone https://github.com/casbin/casdoor
```

Finally, change directory:

```bash
cd casdoor/
```

We provide two start up methods for all kinds of users.

### Manual

#### Simple configuration

Edit `conf/app.conf`, modify `dataSourceName` to correct database info, which follows this format:

```bash
username:password@tcp(database_ip:database_port)/
```

#### Run

Casdoor provides two run modes, the difference is binary size and user prompt.

##### Dev Mode

Edit `conf/app.conf`, set `runmode=dev`. Firstly build front-end files:

```bash
cd web/ && npm install && npm run start
```

Then build back-end binary file, change directory to root(Relative to casdoor):

```bash
go run main.go
```

That's it! Try to visit http://127.0.0.1:7001/. :small_airplane:

##### Production Mode

Edit `conf/app.conf`, set `runmode=prod`. Firstly build front-end files:

```bash
cd web/ && npm install && npm run build
```

Then build back-end binary file, change directory to root(Relative to casdoor):

```bash
go build main.go && sudo ./main
```

> Notice, you should visit back-end port, default 8000. Now try to visit **http://SERVER_IP:8000/**

### Docker

This method requires [docker](https://docs.docker.com/get-docker/) and [docker-compose](https://docs.docker.com/compose/install/) to be installed first.

#### Simple configuration

Edit `conf/app.conf`, modify `dataSourceName` to the fixed content:

```bash
dataSourceName = root:123@tcp(db:3306)/
```

> If you need to modify `conf/app.conf`, you need to re-run `docker-compose up`.

#### Run

```bash
docker-compose up
```

That's it! Try to visit http://localhost:8000/. :small_airplane:

### Docker Hub

This method requires [docker](https://docs.docker.com/get-docker/) and [docker-compose](https://docs.docker.com/compose/install/) to be installed first.

```bash
docker pull casbin/casdoor
```

## Detailed documentation

We also provide a complete [document](https://casdoor.org/) as a reference.

## Other examples

These all use casdoor as a centralized authentication platform.

- [Casnode](https://github.com/casbin/casnode): Next-generation forum software based on React + Golang.
- [Casbin-OA](https://github.com/casbin/casbin-oa): A full-featured OA(Office Assistant) system.
- ......

## Contribute

For casdoor, if you have any questions, you can give Issues, and you can also directly Pull Requests(but we recommend give issues first to communicate with the community).

### I18n notice

If you are contributing to casdoor, please note that we use [Crowdin](https://crowdin.com/project/casdoor-web) as translating platform and i18next as translating tool. When you add some words using i18next in the ```web/``` directory, please remember to add what you have added to the ```web/src/locales/en/data.json``` file.

## License

 [Apache-2.0](https://github.com/casbin/casdoor/blob/master/LICENSE)

