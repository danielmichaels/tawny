package design

import . "goa.design/goa/v3/dsl"

var _ = Service("identity", func() {
	Description("The identity service")
	HTTP(func() {
		Path("/identity")
	})
	Security(APIKeyAuth)
	commonErrors()
	// Users
	Method("createUser", func() {
		Description("Create a new user")
		Payload(func() {
			Attribute("user", UserIn)
			APIKey(
				apiKeyScheme,
				apiKeyName,
				String,
				func() { Description("API key"); Example("key_000000000000") },
			)
			Required("user", apiKeyName)
		})
		Result(UserResult)
		HTTP(func() {
			POST("/users")
			Response(StatusCreated)
			Header(apiKeyHeader)
			commonResponses()
		})
	})
	Method("retrieveUser", func() {
		Description("Retrieve a single User")
		Payload(func() {
			Field(1, "id", String, "ID of the user", func() {
				Pattern(userRx)
				Example("user_123456789012")
			})
			APIKey(
				apiKeyScheme,
				apiKeyName,
				String,
				func() { Description("API key"); Example("key_000000000000") },
			)
			Required("id", apiKeyName)
		})
		Result(UserResult)
		HTTP(func() {
			GET("/users/{id}")
			Response(StatusOK)
			Header(apiKeyHeader)
			commonResponses()
		})
	})
	// Teams
	Method("createTeam", func() {
		Description("Create a new team")
		Payload(func() {
			Attribute("team", TeamIn)
			APIKey(
				apiKeyScheme,
				apiKeyName,
				String,
				func() { Description("API key"); Example("key_000000000000") },
			)
			Required("team", apiKeyName)
		})
		Result(TeamResult)
		HTTP(func() {
			POST("/teams")
			Response(StatusCreated)
			Header(apiKeyHeader)
			commonResponses()
		})
	})
	Method("retrieveTeam", func() {
		Description("Retrieve a single Team")
		Payload(func() {
			Field(1, "id", String, "ID of the team", func() {
				Pattern(teamRx)
				Example("team_123456789012")
			})
			APIKey(
				apiKeyScheme,
				apiKeyName,
				String,
				func() { Description("API key"); Example("key_000000000000") },
			)
			Required("id", apiKeyName)
		})
		Result(TeamResult)
		HTTP(func() {
			GET("/teams/{id}")
			Response(StatusOK)
			Header(apiKeyHeader)
			commonResponses()
		})
	})
})

var UserIn = Type("User", func() {
	Description("User object")
	Attribute("username", String, "ID of the user", func() { Example("my-username") })
	Attribute("password", String, "ID of the user", func() { Example("fakePassword") })
	Attribute("email", String, "ID of the user", func() { Example("email@example.com") })
})
var UserResult = ResultType("application/vnd.tawny.user", func() {
	TypeName("UserResult")
	Description("A single user")

	Attribute("id", String, "User ID", func() { Example("123") })
	Attribute("username", String, "Username", func() { Example("username1") })
	Attribute("email", String, "Email", func() { Example("me@gmail.com") })
	Attribute("verified", String, "Verified", func() { Example("true") })
	CreatedAndUpdateAtResult()
	Required("username", "email")

	View(viewDefault, func() {
		Attribute("id")
		Attribute("username")
		Attribute("email")
		Attribute("verified")
	})
})
var UsersResult = ResultType("application/vnd.tawny.users", func() {
	TypeName("Users")
	Attributes(func() {
		Attribute("users", CollectionOf(UserResult))
		Attribute("total", Int32)
		Attribute("page", Int32)
		Required("users", "total", "page")
	})
})

var TeamIn = Type("Team", func() {
	Description("Team object")
	Attribute("name", String, "Name", func() { Example("Dream Team") })
	Attribute("email", String, "Email", func() { Example("my-team@teamsters.com") })
})
var TeamResult = ResultType("application/vnd.tawny.team", func() {
	TypeName("Team")
	Description("A single team")
	Attribute("team_id", String, "Team ID", func() { Example("team_123456789012") })
	Attribute("name", String, "Name", func() { Example("Dream Team") })
	Attribute("email", String, "Email", func() { Example("my-team@teamsters.com") })
	Required("team_id", "name", "email")

	View(viewDefault, func() {
		Attribute("team_id")
		Attribute("name")
		Attribute("email")
	})
})
