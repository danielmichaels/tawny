// design
// common rules:
// GET:
// params are used to filter and to provide user,team or device context
// POST,PUT,PATCH:
// body is used to provide user,team or device context

package design

import (
	"fmt"

	assets "github.com/danielmichaels/tawny"
	. "goa.design/goa/v3/dsl"
	cors "goa.design/plugins/v3/cors/dsl"
)

const (
	userRx            = "^user_[a-zA-Z0-9]{7}$"
	teamRx            = "^team_[a-zA-Z0-9]{7}$"
	keyRx             = "^key_[a-zA-Z0-9]{20}$"
	apiKeyScheme      = "api_key"
	apiKeyName        = "key"
	apiKeyHeaderValue = "X-API-KEY"
	viewDefault       = "default"
	viewShort         = "short"
)

var (
	apiKeyHeader = fmt.Sprintf("%s:%s", apiKeyName, apiKeyHeaderValue)
)

// API describes the global properties of the API server.
//
// I think of this as the "headings" to a OpenAPI spec for lack of a better analogy
var _ = API(assets.AppName, func() {
	Title(fmt.Sprintf("%s Service", assets.AppName))
	Description("API service for tawny")
	HTTP(func() {
		Path("/v1")
	})
	Server("server", func() {
		Host("localhost", func() { URI("http://localhost:9090") })
	})
})

// APIKeyAuth defines a security scheme that uses API keys.
var APIKeyAuth = APIKeySecurity("api_key", func() {
	Description("Secures endpoint by requiring an API key.")
})

var Origins = []string{
	"/127.0.0.1:\\d+/",             // Dev
	"/localhost:\\d+/",             // Dev
	"/192.168.(\\d+).(\\d+):\\d+/", // Dev
}

func commonErrors() {
	Error("unauthorized", Unauthorized, "unauthorized")
	Error("forbidden", Forbidden, "forbidden")
	Error("not-found", NotFound, "not found")
	Error("bad-request", BadRequest, "bad request")
	Error("server-error", ServerError, "server error")
	Error("unauthorized", String, "Credentials are invalid")
	Error("invalid-scopes", String, "Token scopes are invalid")
}
func commonResponses() {
	Response("unauthorized", StatusUnauthorized)
	Response("forbidden", StatusForbidden)
	Response("not-found", StatusNotFound)
	Response("bad-request", StatusBadRequest)
	Response("server-error", StatusInternalServerError)
}

// commonOptions provides a range of dsl schema applicable to all services.
func commonCors() {
	corsRules := func() {
		cors.Headers("X-API-TOKEN", "Content-Type")
		cors.Expose("X-API-TOKEN", "Content-Type")
		cors.Methods("GET", "OPTIONS", "POST", "DELETE", "PATCH", "PUT")
		cors.Credentials()
	}
	for _, origin := range Origins {
		cors.Origin(origin, corsRules)
	}
}

func createdAndUpdateAtResult() {
	Attribute("created_at", String, "Created at", func() { Example("2024-04-18 01:18:43 +0000") })
	Attribute("updated_at", String, "Updated at", func() { Example("2024-04-21 11:43:02 +0000") })
}

// PaginationQueryParams returns the current_page and page_size values
func PaginationQueryParams(page, pageSize int) (int32, int32) {
	return int32(page), int32((pageSize - 1) * page)
}
func paginationParams() {
	Param("page_size", Int)
	Param("page_number", Int)
}
func paginationPayload() {
	Attribute("page_size", Int, func() {
		Description("The maximum number of results to return.")
		Default(20)
		Example(20)
	})
	Attribute("page_number", Int, func() {
		Description("The page number to view")
		Default(1)
		Example(1)
	})
}
func apiKeyAuth() {
	APIKey(
		apiKeyScheme,
		apiKeyName,
		String,
		func() { Description("API key"); Example("key_00000000000000000000"); Pattern(keyRx) },
	)
	Required(apiKeyName)
}

var PaginationMetadata = Type("PaginationMetadata", func() {
	Attribute("total", Int32, func() { Example(25) })
	Attribute("current_page", Int32, func() { Example(1) })
	Attribute("first_page", Int32, func() { Example(1) })
	Attribute("last_page", Int32, func() { Example(10) })
	Attribute("page_size", Int32, func() { Example(20) })
	Required("total", "page_size", "first_page", "current_page", "last_page")
})

var NotFound = Type("not-found", func() {
	Description("not-found indicates the resource matching the id does not exist.")
	Attribute("id", String, "ID of device", func() {
		Example("resource_1234567")
	})
	Attribute("name", String, "Name of error", func() { Example("not found") })
	Attribute("message", String, "Error message", func() {
		Example("bad request")
	})
	Attribute("detail", String, "Error details", func() {
		Example("Resource not found")
	})
	Required("message", "detail", "name")

})
var BadRequest = Type("bad-request", func() {
	Description("bad-request indicates the values provided are invalid")
	Attribute("name", String, "Name of error", func() { Example("bad request") })
	Attribute("message", String, "Error message", func() {
		Example("bad request")
	})
	Attribute("detail", String, "Error details", func() {
		Example("Failed to validate information. Cannot continue.")
	})
	Required("message", "detail", "name")
})
var ServerError = Type("server-error", func() {
	Description("server-error indicates the server encountered an error.")
	Attribute("name", String, "Name of error", func() { Example("internal server error") })
	Attribute("message", String, "Error message", func() {
		Example("bad request")
	})
	Attribute("detail", String, "Error details", func() {
		Example("Failed to determine machine information. Cannot continue.")
	})
	Required("message", "detail", "name")
})
var Forbidden = Type("forbidden", func() {
	Description("forbidden indicates access to a resource was denied")
	Attribute("name", String, "Name of error", func() { Example("forbidden") })
	Attribute("message", String, "Error message", func() {
		Example("bad request")
	})
	Attribute("detail", String, "Error details", func() {
		Example("Failed to determine machine information. Cannot continue.")
	})
	Required("message", "detail", "name")
})

var Unauthorized = Type("unauthorized", func() {
	Description("unauthorized indicates authentication failed")
	Attribute("message", String, "Message of the error", func() {
		Example("unauthorized")
	})
	Required("message")
})

var _ = Service("monitoring", func() {
	Description("Monitoring endpoints for external services")
	HTTP(func() {
		Path("/")
	})
	commonCors()
	Method("healthz", func() {
		Description("Health check endpoint")
		HTTP(func() {
			GET("/healthz")
			Response(StatusOK)
		})
		Result(Empty)
	})
	Method("version", func() {
		Description("Application version information endpoint")
		HTTP(func() {
			GET("/version")
			Response(StatusOK)
		})
		Result(AppVersion)
	})
})

var AppVersion = Type("version", func() {
	Description("Application version information")
	Attribute("version", String, "Application version", func() {
		Example("1.0")
	})
	Attribute("build_time", String, "Application build time", func() {
		Example("")
	})
})
