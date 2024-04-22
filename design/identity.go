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
		Description("Create a new user. This will also generate a new team for that user.")
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
		Description("Retrieve a single user. Can only retrieve users from an associated team.")
		Payload(func() {
			Attribute("id", String, "ID of the user", func() {
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
	Method("listUsers", func() {
		Description("Retrieve all users that this user can see from associated teams.")
		Payload(func() {
			APIKey(
				apiKeyScheme,
				apiKeyName,
				String,
				func() { Description("API key"); Example("key_000000000000") },
			)
			paginationPayload()
			Required(apiKeyName)
		})
		Result(UsersResult)
		HTTP(func() {
			GET("/users")
			Response(StatusOK)
			Header(apiKeyHeader)
			paginationParams()
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
	Method("addTeamMember", func() {
		Description("Add a user to a team")
		Payload(func() {
			Attribute("user_id", String)
			Attribute("team_id", String)
			APIKey(
				apiKeyScheme,
				apiKeyName,
				String,
				func() { Description("API key"); Example("key_000000000000") },
			)
			Required("user_id", "team_id", apiKeyName)
		})
		Result(TeamResult)
		HTTP(func() {
			POST("/teams/{team_id}/users")
			Response(StatusCreated)
			Header(apiKeyHeader)
			commonResponses()
		})
	})
	Method("removeTeamMember", func() {
		Description("Remove a team member from a team")
		Payload(func() {
			Attribute("user_id", String)
			Attribute("team_id", String)
			APIKey(
				apiKeyScheme,
				apiKeyName,
				String,
				func() { Description("API key"); Example("key_000000000000") },
			)
			Required("user_id", "team_id", apiKeyName)
		})
		Result(TeamResult)
		HTTP(func() {
			DELETE("/teams/{team_id}/users")
			Response(StatusOK)
			Header(apiKeyHeader)
			commonResponses()
		})
	})
})

var PaginationMetadata = Type("PaginationMetadata", func() {
	Attribute("total", Int32, func() { Example(25) })
	Attribute("current_page", Int32, func() { Example(1) })
	Attribute("first_page", Int32, func() { Example(1) })
	Attribute("last_page", Int32, func() { Example(10) })
	Attribute("page_size", Int32, func() { Example(20) })
	Required("total", "page_size", "first_page", "current_page", "last_page")
})
var UserIn = Type("User", func() {
	Description("User object")
	Attribute("username", String, "ID of the user", func() { Example("my-username") })
	Attribute("password", String, "ID of the user", func() { Example("fakePassword") })
	Attribute("email", String, "ID of the user", func() { Example("email@example.com") })
	Required("username", "password", "email")
})
var UserResult = ResultType("application/vnd.tawny.user", func() {
	TypeName("UserResult")
	Description("A single user")

	Attribute("id", String, "User ID", func() { Example("123") })
	Attribute("username", String, "Username", func() { Example("username1") })
	Attribute("email", String, "Email", func() { Example("me@gmail.com") })
	Attribute("role", String, "Role", func() { Example("admin") })
	Attribute("verified", Boolean, "Verified", func() { Example(true) })
	createdAndUpdateAtResult()
	Required("username", "email", "role")

	View(viewDefault, func() {
		Attribute("id")
		Attribute("username")
		Attribute("email")
		Attribute("role")
		Attribute("verified")
		Attribute("created_at")
		Attribute("updated_at")
	})
})
var UsersResult = ResultType("application/vnd.tawny.users", func() {
	TypeName("Users")
	Attributes(func() {
		Attribute("users", CollectionOf(UserResult))
		Attribute("metadata", PaginationMetadata)
		Required("users", "metadata")
	})
})
var TeamIn = Type("Team", func() {
	Description("Team object")
	Attribute("name", String, "Name", func() { Example("Dream Team") })
	Attribute("email", String, "Email", func() { Example("my-team@teamsters.com") })
	Required("name", "email")
})
var TeamResult = ResultType("application/vnd.tawny.team", func() {
	TypeName("Team")
	Description("A single team")
	Attribute("team_id", String, "Team ID", func() { Example("team_123456789012") })
	Attribute("name", String, "Name", func() { Example("Dream Team") })
	Attribute("email", String, "Email", func() { Example("my-team@teamsters.com") })
	createdAndUpdateAtResult()
	Required("team_id", "name", "email")

	View(viewDefault, func() {
		Attribute("team_id")
		Attribute("name")
		Attribute("email")
	})
})
