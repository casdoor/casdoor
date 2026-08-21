<div align="center">
  <a href="https://casdoor.ai">
    <img src="https://cdn.casbin.org/img/casdoor-logo_1185x256.png" alt="Casdoor" width="500">
  </a>

  <h3>An open-source, self-hosted identity and access management platform</h3>

  <p>
    Casdoor is a single sign-on (SSO) and authentication server with a web console.<br>
    It speaks <strong>OAuth&nbsp;2.0</strong>, <strong>OIDC</strong>, <strong>SAML&nbsp;2.0</strong>, <strong>CAS</strong>, <strong>LDAP</strong>, <strong>SCIM&nbsp;2.0</strong>, <strong>WebAuthn</strong>, <strong>TOTP/MFA</strong> and <strong>MCP</strong>,<br>
    and connects to Google Workspace, Microsoft Entra ID (Azure AD), GitHub and many other identity providers.
  </p>

  <p>
    <a href="https://casdoor.ai"><strong>Website</strong></a> &middot;
    <a href="https://casdoor.ai/docs/overview"><strong>Documentation</strong></a> &middot;
    <a href="https://demo.casdoor.com"><strong>Live demo</strong></a> &middot;
    <a href="https://discord.gg/5rPsrAzK7S"><strong>Discord</strong></a>
  </p>

  <p>
    <a href="https://github.com/casdoor/casdoor/releases/latest">
      <img src="https://img.shields.io/github/v/release/casdoor/casdoor?style=flat-square&color=blue" alt="Release">
    </a>
    <a href="https://hub.docker.com/r/casbin/casdoor">
      <img src="https://img.shields.io/docker/pulls/casbin/casdoor?style=flat-square&color=brightgreen" alt="Docker Pulls">
    </a>
    <a href="https://github.com/casdoor/casdoor/actions/workflows/build.yml">
      <img src="https://img.shields.io/github/actions/workflow/status/casdoor/casdoor/build.yml?style=flat-square&label=build" alt="Build Status">
    </a>
    <a href="https://github.com/casdoor/casdoor/actions/workflows/golangci-lint.yml">
      <img src="https://img.shields.io/github/actions/workflow/status/casdoor/casdoor/golangci-lint.yml?style=flat-square&label=golangci-lint&logo=go&logoColor=white" alt="golangci-lint">
    </a>
    <a href="https://discord.gg/5rPsrAzK7S">
      <img src="https://img.shields.io/discord/1022748306096537660?style=flat-square&logo=discord&label=Discord&color=5865F2" alt="Discord">
    </a>
    <a href="https://github.com/casdoor/casdoor/blob/master/LICENSE">
      <img src="https://img.shields.io/github/license/casdoor/casdoor?style=flat-square&color=orange" alt="License">
    </a>
  </p>
</div>

<div align="center">
  <a href="https://door.casdoor.net">
    <img src="https://cdn.casbin.org/img/casdoor-signin.png" alt="Casdoor sign-in page with password, code, WebAuthn and Face ID tabs and social login icons" width="900">
  </a>
  <p><sub>The sign-in page your users see: password, email/SMS code, WebAuthn and Face ID, plus every social provider you enable.</sub></p>
</div>

<table>
  <tr>
    <td width="33%" valign="top" align="center">
      <a href="https://door.casdoor.net"><img src="https://cdn.casbin.org/img/casdoor-console.png" alt="Casdoor admin console dashboard with user, application and provider statistics"></a>
      <sub><b>Admin console.</b> Users, tokens, organizations and providers at a glance.</sub>
    </td>
    <td width="33%" valign="top" align="center">
      <a href="https://door.casdoor.net/applications"><img src="https://cdn.casbin.org/img/casdoor-applications.png" alt="Casdoor applications list showing several applications with their organizations and providers"></a>
      <sub><b>Applications.</b> Every app that delegates login to Casdoor, across all organizations.</sub>
    </td>
    <td width="33%" valign="top" align="center">
      <a href="https://door.casdoor.net/applications/admin/app-built-in"><img src="https://cdn.casbin.org/img/casdoor-application-providers.png" alt="Casdoor application settings, Providers tab, with per-provider signup, signin and unlink toggles"></a>
      <sub><b>Application settings.</b> OAuth, SAML, providers and branding — no redeploy, no config file.</sub>
    </td>
  </tr>
</table>

## 🚀 Try it in 30 seconds

No database and no config file needed. This runs Casdoor on SQLite with sample data:

```bash
docker run -p 8000:8000 casbin/casdoor-all-in-one
```

Open <http://localhost:8000> and sign in:

| Field | Value |
|-------|-------|
| Organization | `built-in` |
| Username | `admin` |
| Password | `123` |

> The sign-in form has separate **organization** and **username** fields. Docs sometimes write this pair as `built-in/admin` — that is the same thing, not a username containing a slash.

Prefer not to install anything? Use the hosted demos:

| Demo | URL | Notes |
|------|-----|-------|
| **Writable** | [demo.casdoor.com](https://demo.casdoor.com) | Full access, so you can click through everything. **All data resets about every 5 minutes.** |
| **Read-only** | [door.casdoor.net](https://door.casdoor.net) | Stable global demo. **Every write operation fails by design.** |

Both accept the same `built-in` / `admin` / `123` credentials.

## 🤔 Why Casdoor

Casdoor is a **complete identity provider**, not an authentication proxy and not a library you embed. It stores your users, issues the tokens, and gives you an admin console to manage all of it — so your applications can delegate login entirely and never handle a password themselves.

- **One server, many protocols.** The same user directory is reachable over OAuth 2.0, OIDC, SAML 2.0, CAS, LDAP and SCIM, so a modern SPA and a legacy CAS-only app can share one set of accounts.
- **Everything is editable in the UI.** Organizations, applications, providers, sign-in methods, email and SMS templates, and login-page branding are configured in the web console instead of in files you have to redeploy.
- **Policy-based authorization built in.** Access rules are expressed with [Casbin](https://casbin.org/) — ACL, RBAC, ABAC and custom models — rather than a fixed permission scheme.
- **Straightforward to self-host.** A single Go binary plus a database. No JVM, no operator, no cluster required.

If all you need is a login screen in front of an existing reverse proxy, a smaller tool may suit you better. Casdoor is for when you want to own the user directory itself.

## 📦 Installation

Four supported paths, fastest first. All of them end up at <http://localhost:8000>.

### Docker — all-in-one (evaluation)

```bash
docker run -p 8000:8000 casbin/casdoor-all-in-one
```

Bundles SQLite and demo data into a single container. Ideal for a first look, but **not intended for production**: the data lives inside the container and disappears with it.

Guide: [Try with Docker](https://casdoor.ai/docs/basic/try-with-docker)

### Docker Compose — Casdoor with MySQL

[`docker-compose.yml`](docker-compose.yml) starts Casdoor next to a MySQL 8 container.

> **Two things to know before running it:**
>
> 1. Compose **builds the image from source** (Go backend plus React frontend). The first `docker compose up` takes several minutes, so it is not the quick-trial path — use the all-in-one image above for that.
> 2. You have to point Casdoor at the bundled database first.

Set the MySQL settings in [`conf/app.conf`](conf/app.conf) to match the `db` service:

```ini
driverName = mysql
dataSourceName = root:123456@tcp(localhost:3306)/
dbName = casdoor
```

Use `localhost` here even though MySQL runs in a separate container: the compose file sets `RUNNING_IN_DOCKER=true`, and Casdoor rewrites `localhost` to the Docker host address at startup (see [`conf/conf.go`](conf/conf.go)). Then start everything:

```bash
docker compose up
```

The compose entrypoint already passes `--createDatabase=true`, so the `casdoor` database is created for you.

Guide: [Try with Docker](https://casdoor.ai/docs/basic/try-with-docker)

### Kubernetes — Helm

Requires Helm v3 and a running cluster:

```bash
helm install casdoor oci://registry-1.docker.io/casbin/casdoor-helm-charts
```

The chart does not expose Casdoor outside the cluster by default. To reach it, find the service and forward a port:

```bash
kubectl get svc
```

```bash
kubectl port-forward svc/<service-name-from-above> 8000:8000
```

For a real deployment, configure an Ingress and an external database through the chart's values. [`k8s.yaml`](k8s.yaml) in this repo is a minimal plain-manifest example if you would rather not use Helm.

Guide: [Try with Helm](https://casdoor.ai/docs/basic/try-with-helm)

### From source — for development

Use this if you intend to modify Casdoor. Prerequisites: **Go 1.25+** (see [`go.mod`](go.mod)), **Node.js 20 LTS**, **Yarn 1.x**, and a supported database (MySQL, PostgreSQL, SQLite, SQL Server and others).

```bash
git clone https://github.com/casdoor/casdoor.git
cd casdoor
```

Set `driverName`, `dataSourceName` and `dbName` in [`conf/app.conf`](conf/app.conf). For MySQL, create the `casdoor` database first, or start the server with `--createDatabase=true`. Then build the frontend and run the server:

```bash
cd web && yarn install && yarn build && cd .. && go run main.go
```

While working on the frontend, run `yarn start` in [`web/`](web) instead of `yarn build` to get hot reload on port 7001, with `go run main.go` serving the API from a second terminal.

Guide: [Server installation](https://casdoor.ai/docs/basic/server-installation)

## 👉 After you sign in

At this point you have a running identity provider with nothing connected to it yet. Next:

1. **Change the `admin` password.** `123` is a demo credential and must not survive contact with production.
2. **[Connect your first application](https://casdoor.ai/docs/how-to-connect/overview)** — create an Application in the console, copy its Client ID and Client Secret, and point your app's OAuth/OIDC client at Casdoor.
3. **[Add an identity provider](https://casdoor.ai/docs/provider/overview)** if you want Google, GitHub or Entra ID sign-in.
4. **[Pick an SDK](https://casdoor.ai/docs/category/integrations)** for your language, or call the [Public API](https://casdoor.ai/docs/basic/public-api) directly.

## ✨ Features

**🔐 Authentication**

- **OAuth 2.0 / OIDC** — full authorization server and OpenID Connect provider
- **SAML 2.0** — enterprise SSO, as both IdP and SP
- **CAS** — Central Authentication Service for legacy applications
- **LDAP** — sync from a directory, or serve as one
- **WebAuthn / passkeys** — passwordless sign-in
- **TOTP / MFA** — multi-factor authentication, including email and SMS codes
- **Face ID** — biometric sign-in

**🏢 Organizations and access control**

- **Multi-tenancy** — independent organizations, each with its own users and branding
- **RBAC and beyond** — roles, permissions and Casbin policy models
- **SCIM 2.0** — automated user provisioning and de-provisioning
- **Social login** — Google, GitHub, Entra ID (Azure AD) and many more
- **Custom providers** — plug in your own identity, email, SMS, storage or payment backends
- **Audit logs** — a record of sign-ins and administrative changes

**🤖 AI and agents**

- **MCP gateway** — expose Model Context Protocol servers and control access to them
- **A2A** — agent-to-agent communication support

**🛠️ Developer experience**

- **REST API** — every console action is also an API call
- **SDKs** — Go, Java, Python, Node.js, .NET, PHP, Rust and more
- **Swagger UI** — [live API explorer](https://door.casdoor.net/swagger)
- **Webhooks** — push user and sign-in events into your own systems
- **Customizable UI** — theme the login page and console per organization

## 🧱 Technology stack

Casdoor is a frontend–backend separated application:

- **Backend** — Go with the [Beego](https://github.com/beego/beego) framework, exposing REST APIs ([repository root](https://github.com/casdoor/casdoor))
- **Frontend** — React 18 with Ant Design ([`web/`](web))
- **Database** — MySQL, PostgreSQL, SQLite, SQL Server and others through [XORM](https://xorm.io/)
- **Cache** — Redis, optional; needed if you run more than one Casdoor replica

## 📖 Documentation

The full documentation lives at **[casdoor.ai/docs](https://casdoor.ai/docs/overview)**. Common starting points:

| I want to… | Go to |
|------------|-------|
| Install Casdoor | [From source](https://casdoor.ai/docs/basic/server-installation) &middot; [Docker](https://casdoor.ai/docs/basic/try-with-docker) &middot; [Helm](https://casdoor.ai/docs/basic/try-with-helm) |
| Connect my application | [How to connect to Casdoor](https://casdoor.ai/docs/how-to-connect/overview) |
| Use the API | [Public API](https://casdoor.ai/docs/basic/public-api) &middot; [Swagger UI](https://door.casdoor.net/swagger) |
| Choose an SDK | [Integrations](https://casdoor.ai/docs/category/integrations) |
| Deploy to production | [Deployment](https://casdoor.ai/docs/category/deployment) |

## 🔌 SDKs and integrations

Official SDKs and framework integrations, by language:

- **Go** — [casdoor-go-sdk](https://github.com/casdoor/casdoor-go-sdk)
- **Java** — [casdoor-java-sdk](https://github.com/casdoor/casdoor-java-sdk) &middot; [Spring Boot starter](https://github.com/casdoor/casdoor-spring-boot-starter)
- **Python** — [casdoor-python-sdk](https://github.com/casdoor/casdoor-python-sdk)
- **Node.js** — [casdoor-nodejs-sdk](https://github.com/casdoor/casdoor-nodejs-sdk)
- **JavaScript** — [casdoor-js-sdk](https://github.com/casdoor/casdoor-js-sdk) &middot; [React](https://github.com/casdoor/casdoor-react-sdk) &middot; [Vue](https://github.com/casdoor/casdoor-vue-sdk) &middot; [Angular](https://github.com/casdoor/casdoor-angular-sdk)
- **.NET** — [casdoor-dotnet-sdk](https://github.com/casdoor-net/casdoor-dotnet-sdk)
- **PHP** — [casdoor-php-sdk](https://github.com/casdoor/casdoor-php-sdk)
- **Rust** — [casdoor-rust-sdk](https://github.com/casdoor/casdoor-rust-sdk)

The complete list, including reverse proxies and third-party applications, is in the [Integrations](https://casdoor.ai/docs/category/integrations) documentation.

## 🔒 Security

**Please do not report security vulnerabilities in public GitHub issues.** Email <admin@casdoor.org> instead — [SECURITY.md](SECURITY.md) has the full policy and disclosure process.

Before exposing a Casdoor instance to the internet:

- Change the built-in `admin` password. Never ship the demo credential `123`.
- Serve Casdoor over HTTPS only, and set `origin` in [`conf/app.conf`](conf/app.conf) to your public URL.
- Review [`conf/app.conf`](conf/app.conf) for values inherited from the sample file, especially `dataSourceName` and any provider secrets.
- Set `runmode = prod` and keep `showSql = false` in production.

## 🤝 Community and support

- **Discord** — [join the community](https://discord.gg/5rPsrAzK7S) for questions and help
- **GitHub Discussions** — [ask and search here](https://github.com/casdoor/casdoor/discussions)
- **GitHub Issues** — [bug reports and feature requests](https://github.com/casdoor/casdoor/issues)
- **Commercial support** — [casdoor.ai/help](https://casdoor.ai/help)

## 🌍 Contributing

Contributions are welcome. For anything larger than a small fix, **please open an issue first** so you can agree on the approach with the maintainers before writing code.

Read the [contribution guidelines](https://casdoor.ai/docs/contributing/) before you start.

**Translations.** User-facing strings in the web console go through [i18next](https://www.i18next.com/). When you add or change one under [`web/`](web), update the English catalog at [`web/src/locales/en/data.json`](web/src/locales/en/data.json). The other languages are translated on [Crowdin](https://crowdin.com/project/casdoor-site) and should not be edited by hand.

## ❤️ Sponsors

Casdoor is free and open source. If it saves you time, consider supporting its development on [Open Collective](https://opencollective.com/casdoor).

<a href="https://opencollective.com/casdoor#sponsor"><img src="https://opencollective.com/casdoor/tiers/sponsor.svg?avatarHeight=74" alt="Sponsors on Open Collective"></a>

<a href="https://opencollective.com/casdoor#backer"><img src="https://opencollective.com/casdoor/tiers/backer.svg?avatarHeight=36" alt="Backers on Open Collective"></a>

## 📄 License

Casdoor is licensed under the [Apache License 2.0](LICENSE).

---

<div align="center">

If Casdoor is useful to you, a star helps other people find it.

<a href="https://github.com/casdoor/casdoor/stargazers"><img src="https://img.shields.io/github/stars/casdoor/casdoor?style=social&logo=github&label=Star" alt="GitHub Stars"></a>

<sub>© 2026 <a href="https://casdoor.ai">Casdoor</a> &middot; <a href="LICENSE">Apache License 2.0</a></sub>

</div>
