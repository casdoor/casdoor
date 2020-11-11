# Casdoor

## Development

### Prerequisites

Go (1.15 or later)
MySQL (5.8 or higher)
Node.js (version 8 or higher)
Yarn

### Initialize project

Create a database named casdoor.

```sql
CREATE DATABASE IF NOT EXISTS casdoor default charset utf8 COLLATE utf8_general_ci
```

Configure database source in the `.env` file. If you have not copied the `.env.template` file to a new file named `.env`, which should now be performed.

### Run project

#### backend

```shell
go run cmd/main.go
```

### frontend

```shell
cd web
yarn
yarn start
```
