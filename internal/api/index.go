package api

import (
	"net/http"
	"ovmsa-be/internal/entities"
	"ovmsa-be/pkg/logger"
	"ovmsa-be/pkg/response"
	validatorHelper "ovmsa-be/pkg/validator"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// formatValidationErrors converts validator errors into a user-friendly format
func formatValidationErrors(err error) map[string]string {
	validationErrors := make(map[string]string)

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			fieldName := e.Field()

			switch e.Tag() {
			case "required":
				validationErrors[fieldName] = "This field is required"
			case "email":
				validationErrors[fieldName] = "Must be a valid email address"
			case "min":
				if e.Kind() == reflect.String {
					validationErrors[fieldName] = "Must be at least " + e.Param() + " characters"
				} else {
					validationErrors[fieldName] = "Value must be at least " + e.Param()
				}
			case "max":
				if e.Kind() == reflect.String {
					validationErrors[fieldName] = "Must be no more than " + e.Param() + " characters"
				} else {
					validationErrors[fieldName] = "Value must be no more than " + e.Param()
				}
			case "url":
				validationErrors[fieldName] = "Must be a valid URL"
			case "oneof":
				validationErrors[fieldName] = "Must be one of [" + strings.ReplaceAll(e.Param(), " ", " | ") + "]"
			case "len":
				validationErrors[fieldName] = "Length must be exactly " + e.Param()
			case "alphanum":
				validationErrors[fieldName] = "Must contain only letters and numbers"
			case "numeric":
				validationErrors[fieldName] = "Must be a valid number"
			default:
				validationErrors[fieldName] = "Invalid value (failed " + e.Tag() + ")"
			}
		}
	} else {
		validationErrors["general"] = err.Error()
	}

	return validationErrors
}

// endPointFunc is a factory function that creates a Gin handler function for each route
func endPointFunc(routeMatrix entities.TRouteMatrix) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if Handler is nil
		if routeMatrix.Controller.Handler == nil {
			logger.Error("Handler not configured for route",
				"path", c.FullPath(),
				"method", c.Request.Method,
			)
			response.InternalServerErrorResponse(c, nil, "Internal server error: route misconfigured")
			return
		}

		// Create an empty struct to hold request parameters
		var routeMatrixSchema any = &struct{}{}
		if routeMatrix.Schema != nil {
			routeMatrixSchema = routeMatrix.Schema
		}

		// Use reflection to create a new instance of the schema type
		/*
			- reflect.ValueOf(routeMatrixSchema) takes the routeMatrixSchema variable and turns it into a reflect.Value object so the code can inspect its structure.
			- .Elem() is crucial here: it assumes routeMatrixSchema is a pointer. .Elem() "dereferences" that pointer to get to the actual value/struct underneath. If it weren't a pointer, this line would panic.
		*/
		schema := reflect.ValueOf(routeMatrixSchema).Elem()
		/*
			- schema.Type() gets the type definition (the "blueprint") of the original object.
			- reflect.New(...) allocates new memory for that type. It’s equivalent to calling new(T) in standard Go. This returns a pointer to the new, zeroed-out value.
			- .Elem() dereferences that new pointer immediately, so the variable clone represents the actual struct value rather than a pointer to it.
		*/
		clone := reflect.New(schema.Type()).Elem()
		/*
			- clone.Addr() takes the address of our new struct, turning it back into a reflect.Value representing a pointer.
			- .Interface() converts that reflection object back into a standard interface{}.
			- The Result: payload is now a pointer to a brand-new, empty instance of whatever type routeMatrixSchema was.
		*/
		payload := clone.Addr().Interface()

		/*
			Stage 1: Payload Binding
			- Translates the raw request (JSON, Form, etc.) into our Go struct.
			- Catches structural/syntax issues like malformed JSON or incorrect data types.
		*/
		err := c.ShouldBind(payload)
		if err != nil {
			validationErrors := formatValidationErrors(err)
			response.ValidationErrorWithDetailsResponse(c, err, "Payload binding error", validationErrors)
			return
		}

		/*
			Stage 2: Schema Validation
			- Performs semantic validation based on business rules (e.g., min length, email format).
			- We execute this manually to support the standard 'validate' tag and ensure.
			- Granular error reporting for business rule violations.
		*/
		if routeMatrix.Schema != nil {
			validate := validatorHelper.GetValidator()
			err := validate.Struct(payload)
			if err != nil {
				validationErrors := formatValidationErrors(err)
				logger.Debug("Validation failed",
					"error", err,
					"details", validationErrors,
				)
				response.ValidationErrorWithDetailsResponse(c, err, "Payload validation error", validationErrors)
				return
			}
		}

		// Create a JWT data object
		jwtData := &entities.TJwtData{}

		// Call the actual handler function
		value, err, validationError := routeMatrix.Controller.Handler(c, payload, jwtData, routeMatrix.Params)
		if err != nil {
			logger.Error("Handler error", "error", err.Error())
			response.InternalServerErrorResponse(c, err, "Internal Server Error")
			return
		} else if validationError != nil {
			validationErrors := formatValidationErrors(validationError)
			response.ValidationErrorWithDetailsResponse(c, validationError, "Validation error", validationErrors)
			return
		} else if value != nil {
			c.JSON(http.StatusOK, response.Success("Success", value))
			return
		}
	}
}

// registerRoute registers a route with the appropriate HTTP method
func registerRoute(router *gin.RouterGroup, path string, method string, handler gin.HandlerFunc) {
	switch strings.ToUpper(method) {
	case "GET":
		router.GET(path, handler)
	case "POST":
		router.POST(path, handler)
	case "PUT":
		router.PUT(path, handler)
	case "DELETE":
		router.DELETE(path, handler)
	case "PATCH":
		router.PATCH(path, handler)
	case "OPTIONS":
		router.OPTIONS(path, handler)
	default:
		router.GET(path, handler)
	}
}

// ApiHandler is the main function that sets up all API routes
func ApiHandler(router *gin.Engine) {
	// Register base routes
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.Success("success", "server up and running"))
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.Success("success", "server healthy"))
	})

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.Success("success", "pong"))
	})

	// Silent handler for browser favicon requests to keep logs clean
	router.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	// Register API routes hierarchical: Platform -> Version -> Module
	for _, registry := range RouteRegistry {
		// Create platform group (e.g., /ovmsa)
		platformRouter := router.Group(registry.Platform)

		// Create version group (e.g., /v1.0)
		versionRouter := platformRouter.Group(registry.Version)

		for _, group := range registry.Groups {
			// Create module group (e.g., /health)
			groupRouter := versionRouter.Group(group.Group)

			for _, routeMatrix := range group.RouteMatrices {
				endpoint := endPointFunc(routeMatrix)

				// Apply protection strategy based on route configuration
				registerRoute(groupRouter, routeMatrix.Path, routeMatrix.Method, endpoint)
			}
		}
	}

	logger.Info("API routes registered successfully")
}
