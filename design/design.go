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
	userRx            = "^user_[a-zA-Z0-9]{12}$"
	teamRx            = "^team_[a-zA-Z0-9]{12}$"
	keyRx             = "^key_[a-zA-Z0-9]{12}$"
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
var _ = API(fmt.Sprintf("%s", assets.AppName), func() {
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

func listParams() {
	Param("page_size", Int32, "Number of results to display")
	Param("page_number", Int32, "Results page number")
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

// genURL joins two url strings together as a single valid URL path
func genURL(base, endpoint string) string {
	return fmt.Sprintf("%s/%s", base, endpoint)
}

var NotFound = Type("not-found", func() {
	Description("not-found indicates the resource matching the id does not exist.")
	Attribute("id", String, "ID of device", func() {
		Example("resource_12345678")
	})
	Attribute("name", String, "Name of error", func() { Example("not found") })
	Attribute("message", String, "Error message", func() {
		Example("bad request")
	})
	Attribute("detail", String, "Error details", func() {
		Example("Failed to determine machine information. Cannot continue.")
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
		Example("Failed to determine machine information. Cannot continue.")
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
