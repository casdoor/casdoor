Casdoor
====

[![Go Report Card](https://goreportcard.com/badge/github.com/casbin/casdoor)](https://goreportcard.com/report/github.com/casbin/casdoor) <img src="https://img.shields.io/github/license/casbin/casdoor?style=flat-square" alt="license"> [![GitHub issues](https://img.shields.io/github/issues/casbin/casdoor?style=flat-square)](https://github.com/casbin/casdoor/issues) [![GitHub stars](https://img.shields.io/github/stars/casbin/casdoor?style=flat-square)](https://github.com/casbin/casdoor/stargazers) [![GitHub forks](https://img.shields.io/github/forks/casbin/casdoor?style=flat-square)](https://github.com/casbin/casdoor/network) 

Casdoor is a UI-first centralized authentication / Single-Sign-On (SSO) platform based on OAuth 2.0 / OIDC.

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

Just execute:

```bash
docker-compose up
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

For casdoor, if you have any questions, you can give Issues, and you can also directly Pull Requests(but we recommend give issues first to communicate with the community).

### I18n notice

If you are contributing to casdoor, please note that we use [Crowdin](https://crowdin.com/project/casdoor-web) as translating platform and i18next as translating tool. When you add some words using i18next in the ```web/``` directory, please remember to add what you have added to the ```web/src/locales/en/data.json``` file.

## License

 [Apache-2.0](https://github.com/casbin/casdoor/blob/master/LICENSE)

