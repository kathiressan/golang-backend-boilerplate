package swagger

import (
	"fmt"
	"net/http"
	"ovmsa-be/internal/api"
	"ovmsa-be/internal/entities"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

// GenerateOpenAPI generates the OpenAPI 3.0 spec from the route registry
func GenerateOpenAPI(registries []api.Registry) OpenAPI {
	doc := OpenAPI{
		OpenAPI: "3.0.0",
		Info: Info{
			Title:       "Go Backend Boilerplate API",
			Description: "Auto-generated API documentation based on route registry",
			Version:     "1.0.0",
		},
		Paths: make(map[string]Path),
		Components: Components{
			Schemas: make(map[string]Schema),
			SecuritySchemes: map[string]SecurityScheme{
				"BearerAuth": {
					Type:         "http",
					Scheme:       "bearer",
					BearerFormat: "JWT",
				},
			},
		},
	}

	generatedSchemas := make(map[string]bool)

	for _, registry := range registries {
		basePath := fmt.Sprintf("/%s/%s", registry.Platform, registry.Version)
		for _, group := range registry.Groups {
			for _, matrix := range group.RouteMatrices {
				processRoute(&doc, basePath, group.Group, matrix, generatedSchemas)
			}
		}
	}

	return doc
}

func processRoute(doc *OpenAPI, basePath, groupPrefix string, route entities.TRouteMatrix, generatedSchemas map[string]bool) {
	// joinPath should handle slashes correctly
	// basePath: /platform/v1
	// groupPrefix: /auth
	// route.Path: /login
	
	fullPath := joinPath(basePath, groupPrefix)
	fullPath = joinPath(fullPath, route.Path)
	
	openAPIPath := convertGinPathToOpenAPI(fullPath)

	if _, exists := doc.Paths[openAPIPath]; !exists {
		doc.Paths[openAPIPath] = make(Path)
	}

	method := strings.ToLower(route.Method)
	opID := fmt.Sprintf("%s%s", method, strings.ReplaceAll(strings.ReplaceAll(openAPIPath, "/", "_"), "{", ""))
	
	operation := Operation{
		OperationID: opID,
		Tags:        []string{cleanupTag(groupPrefix)},
		Summary:     fmt.Sprintf("%s %s", route.Method, route.Path),
		Responses:   make(map[string]Response),
		Security:    []map[string][]string{},
	}

	// Request Body
	if route.Schema != nil && (method == "post" || method == "put" || method == "patch") {
		refName := generateSchema(doc, route.Schema, generatedSchemas)
		operation.RequestBody = &RequestBody{
			Required: true,
			Content: map[string]MediaType{
				"application/json": {
					Schema: &Schema{
						Ref: "#/components/schemas/" + refName,
					},
				},
			},
		}
	}

	// Responses
	successCode := route.SuccessCode
	if successCode == 0 {
		successCode = 200
	}
	successCodeStr := fmt.Sprintf("%d", successCode)

	if route.ResponseSchema != nil {
		refName := generateSchema(doc, route.ResponseSchema, generatedSchemas)
		operation.Responses[successCodeStr] = Response{
			Description: "Success",
			Content: map[string]MediaType{
				"application/json": {
					Schema: &Schema{
						Ref: "#/components/schemas/" + refName,
					},
				},
			},
		}
	} else {
		operation.Responses[successCodeStr] = Response{
			Description: "Successful operation",
		}
	}

	// Path Parameters
	pathParams := extractPathParams(fullPath)
	for _, param := range pathParams {
		operation.Parameters = append(operation.Parameters, Parameter{
			Name:     param,
			In:       "path",
			Required: true,
			Schema:   &Schema{Type: "string"},
		})
	}

	// Security
	if route.ProtectedBy == entities.JWT || route.ProtectedBy == entities.RBAC_AUTH || 
	   route.ProtectedBy == entities.ABAC_AUTH || route.ProtectedBy == entities.COMBINED_AUTH {
		operation.Security = append(operation.Security, map[string][]string{
			"BearerAuth": {},
		})
	}

	doc.Paths[openAPIPath][method] = operation
}

func generateSchema(doc *OpenAPI, model any, generatedSchemas map[string]bool) string {
	t := reflect.TypeOf(model)
	if t == nil {
		return "Unknown"
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Handle simple types (not structs) at top level
	if t.Kind() != reflect.Struct {
		// Only structs get added to Components/Schemas usually
		// But if it is a slice or map, we might want to name it?
		// For simplicity, return a name but don't add to schemas if it's primitive,
		// usually top level schema is expected to be an object.
		return t.Name()
	}

	name := t.Name()
	if name == "" {
		name = "Anonymous_" + fmt.Sprintf("%p", model)
	}

	if generatedSchemas[name] {
		return name
	}
	// Mark as generated to avoid recursion loops
	generatedSchemas[name] = true

	schema := Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		
		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		
		jsonName := strings.Split(jsonTag, ",")[0]
		if jsonName == "" {
			jsonName = field.Name
		}

		propSchema := typeToSchema(doc, field.Type, generatedSchemas)
		schema.Properties[jsonName] = propSchema

		if strings.Contains(field.Tag.Get("binding"), "required") || strings.Contains(field.Tag.Get("validate"), "required") {
			schema.Required = append(schema.Required, jsonName)
		}
	}

	doc.Components.Schemas[name] = schema
	return name
}

func typeToSchema(doc *OpenAPI, t reflect.Type, generatedSchemas map[string]bool) *Schema {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return &Schema{Type: "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Schema{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: "number"}
	case reflect.Bool:
		return &Schema{Type: "boolean"}
	case reflect.Struct:
		// Special types
		if t.PkgPath() == "time" && t.Name() == "Time" {
			return &Schema{Type: "string", Format: "date-time"}
		}
		
		// Recursively generate
		dummy := reflect.New(t).Interface()
		refName := generateSchema(doc, dummy, generatedSchemas)
		return &Schema{Ref: "#/components/schemas/" + refName}

	case reflect.Slice, reflect.Array:
		// Handle UUID
		if t.Kind() == reflect.Array && t.Name() == "UUID" && strings.Contains(t.PkgPath(), "google/uuid") {
			return &Schema{Type: "string", Format: "uuid"}
		}
		return &Schema{
			Type:  "array",
			Items: typeToSchema(doc, t.Elem(), generatedSchemas),
		}
		
	case reflect.Map:
		return &Schema{Type: "object"}
	default:
		return &Schema{Type: "string"}
	}
}

func joinPath(p1, p2 string) string {
	return "/" + strings.Trim(p1, "/") + "/" + strings.Trim(p2, "/")
}

func convertGinPathToOpenAPI(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + part[1:] + "}"
		}
	}
	return strings.Join(parts, "/")
}

func extractPathParams(path string) []string {
	var params []string
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			params = append(params, part[1:])
		}
	}
	return params
}

func cleanupTag(group string) string {
	return strings.Trim(group, "/")
}

// NewHandler creates a Gin handler to serve Swagger UI
func NewHandler(registries []api.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		// If requesting the JSON spec
		if strings.HasSuffix(c.Request.URL.Path, "doc.json") {
			spec := GenerateOpenAPI(registries)
			c.JSON(http.StatusOK, spec)
			return
		}

		// Otherwise serve the UI HTML
		// We assume the route is mounted at /swagger
		// The JSON is at /swagger/doc.json
		
		// Note: window.location.pathname might include trailing slash
		// Simple fix to ensure correct JSON path
		
		html := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css" >
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin:0; background: #fafafa; }
    </style>
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"> </script>
<script src="https://unpkg.com/swagger-ui-dist/swagger-ui-standalone-preset.js"> </script>
<script>
window.onload = function() {
    let url = window.location.pathname;
    if (url.endsWith('/')) {
        url += 'doc.json';
    } else {
        url += '/doc.json';
    }

    const ui = SwaggerUIBundle({
        url: url,
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
            SwaggerUIBundle.presets.apis,
            SwaggerUIStandalonePreset
        ],
        layout: "StandaloneLayout"
    })
    window.ui = ui
}
</script>
</body>
</html>
`
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, html)
	}
}
