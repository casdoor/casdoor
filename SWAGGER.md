# Swagger Documentation

This document explains how to update and regenerate the Swagger API documentation for Casdoor.

## Overview

Casdoor uses [Beego](https://beego.me/) framework's annotation-based approach to generate Swagger documentation. The swagger files are located in the `swagger/` directory and include:

- `swagger.json` - JSON format OpenAPI 2.0 specification
- `swagger.yml` - YAML format OpenAPI 2.0 specification  
- `index.html` - Swagger UI interface

## Viewing the Documentation

The Swagger UI is available at `/swagger` when running Casdoor:

- Live demo: https://door.casdoor.com/swagger
- Local development: http://localhost:8000/swagger

## Regenerating Swagger Documentation

When you add or modify API endpoints in the controllers, you need to regenerate the Swagger documentation. Follow these steps:

### Prerequisites

Install the Bee tool (Beego's CLI):

```bash
go install github.com/beego/bee/v2@latest
```

### Steps to Update

1. **Add/Update Swagger Annotations**

   When creating or modifying API endpoints, add swagger annotations to your controller methods:

   ```go
   // GetSubscriptions
   // @Title GetSubscriptions
   // @Tag Subscription API
   // @Description get subscriptions
   // @Param   owner     query    string  true        "The owner of subscriptions"
   // @Success 200 {array} object.Subscription The Response object
   // @router /get-subscriptions [get]
   func (c *ApiController) GetSubscriptions() {
       // ... implementation
   }
   ```

2. **Generate Documentation**

   Run the bee tool to generate swagger documentation:

   ```bash
   bee generate docs
   ```

   Or use the Makefile target:

   ```bash
   make swagger
   ```

3. **Fix Metadata and Tags**

   The bee tool may not preserve all metadata correctly. Run the fix script to restore proper metadata:

   ```bash
   python3 scripts/fix_swagger.py
   ```

   This script will:
   - Restore the API title, description, version, and contact info
   - Clean up tag names from package paths to readable names
   - Remove any unnecessary HTML tags from descriptions

4. **Verify Changes**

   Start the Casdoor server and visit http://localhost:8000/swagger to verify the documentation looks correct.

5. **Commit the Changes**

   ```bash
   git add swagger/swagger.json swagger/swagger.yml
   git commit -m "Update swagger documentation"
   ```

## Swagger Annotation Format

Casdoor uses Beego v1 annotation format. Here are the common annotations:

- `@Title` - Short title for the endpoint
- `@Tag` - Group/category for the endpoint
- `@Description` - Detailed description of what the endpoint does
- `@Param` - Define input parameters (query, body, path, header)
- `@Success` - Define successful response with status code and type
- `@Failure` - Define error responses
- `@router` - Define the route path and HTTP method

### Parameter Format

```
@Param <name> <location> <type> <required> "<description>"
```

Where:
- `location` can be: `query`, `path`, `body`, `header`, `formData`
- `type` can be: `string`, `int`, `bool`, `object.TypeName`, etc.
- `required` is either `true` or `false`

### Example

```go
// UpdateUser
// @Title UpdateUser
// @Tag User API
// @Description update user
// @Param   id     query    string  true        "The id ( owner/name ) of the user"
// @Param   body    body   object.User  true        "The details of the user"
// @Success 200 {object} controllers.Response The Response object
// @router /update-user [post]
func (c *ApiController) UpdateUser() {
    // implementation
}
```

## Package-Level Documentation

The main API metadata is defined in `routers/router.go`:

```go
// Package routers
// @APIVersion 1.503.0
// @Title Casdoor RESTful API
// @Description Swagger Docs of Casdoor Backend API
// @Contact casbin@googlegroups.com
// @SecurityDefinition AccessToken apiKey Authorization header
// @Schemes https,http
// @ExternalDocs Find out more about Casdoor
// @ExternalDocsUrl https://casdoor.org/
package routers
```

## Troubleshooting

### Bee tool not found

Make sure `~/go/bin` is in your PATH:

```bash
export PATH="$HOME/go/bin:$PATH"
```

### Generated docs have wrong metadata

This is expected with bee v2 and beego v1 combination. Run the fix script after generation:

```bash
python3 scripts/fix_swagger.py
```

### Missing API endpoints

1. Check that your controller method has proper swagger annotations
2. Verify the route is registered in `routers/router.go`
3. Regenerate the documentation

## References

- [Beego Documentation](https://beego.me/docs/advantage/docs.md)
- [Swagger/OpenAPI Specification](https://swagger.io/specification/v2/)
- [Casdoor API Documentation](https://casdoor.org/docs/basic/public-api)
