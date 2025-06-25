<h1 align="center" style="border-bottom: none;">📦⚡️ Casdoor</h1>
<h3 align="center">An open-source UI-first Identity and Access Management (IAM) / Single-Sign-On (SSO) platform with web UI supporting OAuth 2.0, OIDC, SAML, CAS, LDAP, SCIM, WebAuthn, TOTP, MFA and RADIUS</h3>
<p align="center">
  <a href="#badge">
    <img alt="semantic-release" src="https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg">
  </a>
  <a href="https://hub.docker.com/r/casbin/casdoor">
    <img alt="docker pull casbin/casdoor" src="https://img.shields.io/docker/pulls/casbin/casdoor.svg">
  </a>
  <a href="https://github.com/casdoor/casdoor/actions/workflows/build.yml">
    <img alt="GitHub Workflow Status (branch)" src="https://github.com/casdoor/casdoor/workflows/Build/badge.svg?style=flat-square">
  </a>
  <a href="https://github.com/casdoor/casdoor/releases/latest">
    <img alt="GitHub Release" src="https://img.shields.io/github/v/release/casdoor/casdoor.svg">
  </a>
  <a href="https://hub.docker.com/r/casbin/casdoor">
    <img alt="Docker Image Version (latest semver)" src="https://img.shields.io/badge/Docker%20Hub-latest-brightgreen">
  </a>
</p>

<p align="center">
  <a href="https://goreportcard.com/report/github.com/casdoor/casdoor">
    <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/casdoor/casdoor?style=flat-square">
  </a>
  <a href="https://github.com/casdoor/casdoor/blob/master/LICENSE">
    <img src="https://img.shields.io/github/license/casdoor/casdoor?style=flat-square" alt="license">
  </a>
  <a href="https://github.com/casdoor/casdoor/issues">
    <img alt="GitHub issues" src="https://img.shields.io/github/issues/casdoor/casdoor?style=flat-square">
  </a>
  <a href="#">
    <img alt="GitHub stars" src="https://img.shields.io/github/stars/casdoor/casdoor?style=flat-square">
  </a>
  <a href="https://github.com/casdoor/casdoor/network">
    <img alt="GitHub forks" src="https://img.shields.io/github/forks/casdoor/casdoor?style=flat-square">
  </a>
  <a href="https://crowdin.com/project/casdoor-site">
    <img alt="Crowdin" src="https://badges.crowdin.net/casdoor-site/localized.svg">
  </a>
  <a href="https://discord.gg/5rPsrAzK7S">
    <img alt="Discord" src="https://img.shields.io/discord/1022748306096537660?style=flat-square&logo=discord&label=discord&color=5865F2">
  </a>
</p>

<p align="center">
  <sup>Sponsored by</sup>
  <br>
  <a href="https://stytch.com/docs?utm_source=oss-sponsorship&utm_medium=paid_sponsorship&utm_campaign=casbin">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="https://cdn.casbin.org/img/stytch-white.png">
      <source media="(prefers-color-scheme: light)" srcset="https://cdn.casbin.org/img/stytch-charcoal.png">
      <img src="https://cdn.casbin.org/img/stytch-charcoal.png" width="275">
    </picture>
  </a><br/>
  <a href="https://stytch.com/docs?utm_source=oss-sponsorship&utm_medium=paid_sponsorship&utm_campaign=casbin"><b>Build auth with fraud prevention, faster.</b><br/> Try Stytch for API-first authentication, user & org management, multi-tenant SSO, MFA, device fingerprinting, and more.</a>
  <br>
</p>

## Online demo

- Read-only site: https://door.casdoor.com (any modification operation will fail)
- Writable site: https://demo.casdoor.com (original data will be restored for every 5 minutes)

## Documentation

https://casdoor.org

## Install

- By source code: https://casdoor.org/docs/basic/server-installation
- By Docker: https://casdoor.org/docs/basic/try-with-docker
- By Kubernetes Helm: https://casdoor.org/docs/basic/try-with-helm

## How to connect to Casdoor?

https://casdoor.org/docs/how-to-connect/overview

## Casdoor Public API

- Docs: https://casdoor.org/docs/basic/public-api
- Swagger: https://door.casdoor.com/swagger

## Integrations

https://casdoor.org/docs/category/integrations

## How to contact?

- Discord: https://discord.gg/5rPsrAzK7S
- Contact: https://casdoor.org/help

## Contribute

For casdoor, if you have any questions, you can give Issues, or you can also directly start Pull Requests(but we recommend giving issues first to communicate with the community).

### I18n translation

If you are contributing to casdoor, please note that we use [Crowdin](https://crowdin.com/project/casdoor-site) as translating platform and i18next as translating tool. When you add some words using i18next in the `web/` directory, please remember to add what you have added to the `web/src/locales/en/data.json` file.

## License

[Apache-2.0](https://github.com/casdoor/casdoor/blob/master/LICENSE)
