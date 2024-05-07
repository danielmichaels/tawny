package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = Service("domains", func() {
	Description("The domain service")
	HTTP(func() {
		Path("/domains")
	})
	Security(APIKeyAuth)
	commonErrors()
	Method("listDomains", func() {
		Description("List all domains which this user has access to manage")
		Payload(func() {
			apiKeyAuth()
			Attribute("app_id", String, func() { Example("my-app") })
			paginationPayload()
			Required(apiKeyName, "app_id")
		})
		Result(DomainsResult)
		HTTP(func() {
			GET("/{app_id}")
			Response(StatusOK)
			Header(apiKeyHeader)
			paginationParams()
			commonResponses()
		})
	})
	Method("createDomain", func() {
		Description("Create a new domain")
		Payload(func() {
			apiKeyAuth()
			Attribute("domain", String, func() { Example("example.com") })
			Attribute("app_id", String, func() { Example("my-app") })
			Required(apiKeyName, "domain", "app_id")
		})
		Result(DomainResult)
		HTTP(func() {
			POST("/{app_id}")
			Response(StatusCreated)
			Header(apiKeyHeader)
			commonResponses()
		})
	})
})

var DomainIn = Type("Domain", func() {
	Description("Domain type")
	Attribute("domain_name", String, func() { Example("tawny.com", "123.123.123.123") })
	Attribute("project", String, func() {
		Example(
			"my-project. Unique identifier provided by tawny once a project or application is created.",
		)
	})
	Attribute("email_address", String, func() { Example("me@tawny.com") })
	Attribute("protocol", String, func() { Example("http"); Example("https"); Default("https") })
	Attribute("port", String, func() {
		Example("8080")
		Description(
			"Port to listen on. Must match the port of the service which is being exposed. " +
				"Mandatory for IP only addressing and cannot be 80 or 443 unless a domain name is specified.",
		)
	})
	Attribute("certificate_type", String, func() {
		Default("production")
		Example("production")
		Description("Let's Encrypt server type. Accepts staging or production. " +
			"Only use staging for testing and making sure your domain is accessible by Let's Encrypt. " +
			"Rate limits are applied by Let's Encrypt on the production instance and can result in significant locks out if the limits are abused.")
	})
	Required("domain_name", "project")
})

var DomainResult = ResultType("application/vnd.tawny.domain", func() {
	TypeName("DomainResult")
	Description("A single domain result")
	Attribute("domain_name", String, func() { Example("tawny.com", "123.123.123.123") })
	Attribute("project", String, func() {
		Example(
			"my-project. Unique identifier provided by tawny once a project or application is created.",
		)
	})
	Attribute("email_address", String, func() { Example("me@tawny.com") })
	Attribute("protocol", String, func() { Example("http"); Example("https"); Default("https") })
	Attribute("port", String, func() { Example("8080") })
	Attribute("certificate_type", String, func() { Example("production") })

	View(viewDefault, func() {
		Attribute("domain_name")
		Attribute("project")
		Attribute("email_address")
		Attribute("protocol")
		Attribute("port")
		Attribute("certificate_type")
	})
})

var DomainsResult = ResultType("application/vnd.tawny.domains", func() {
	TypeName("DomainsResult")
	Attribute("domains", CollectionOf(DomainResult))
	Attribute("metadata", PaginationMetadata)
	Required("domains", "metadata")
})
