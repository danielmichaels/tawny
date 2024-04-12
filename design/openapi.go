package design

import . "goa.design/goa/v3/dsl"

var _ = Service("openapi", func() {
	Description("OpenAPI endpoints for debugging and demonstration")
	HTTP(func() {
		Path("/openapi")
	})
	Method("file", func() {
		Result(func() {
			Attribute("length", Int64, "Length is the downloaded content length in bytes.", func() {
				Example(4 * 1024 * 1024)
			})
			Attribute("encoding", String, func() {
				Example("application/json")
			})
			Required("length", "encoding")
		})

		Error("invalid_file_path", ErrorResult, "Could not locate file for download")
		Error("internal_error", ErrorResult, "Fault while processing download.")

		HTTP(func() {
			GET("/openapi3.json")
			SkipResponseBodyEncodeDecode()
			Response(func() {
				Header("length:Content-Length")
				Header("encoding:Content-Type")
			})
			Response("invalid_file_path", StatusNotFound)
			Response("internal_error", StatusInternalServerError)
		})
	})
	Method("documentation", func() {

		Result(func() {
			Attribute("length", Int64, "Length is the downloaded content length in bytes.", func() {
				Example(4 * 1024 * 1024)
			})
			Attribute("encoding", String, func() {
				Example("application/json")
			})
			Required("length", "encoding")
		})

		Error("invalid_file_path", ErrorResult, "Could not locate file for download")
		Error("internal_error", ErrorResult, "Fault while processing download.")

		HTTP(func() {
			GET("/docs")
			SkipResponseBodyEncodeDecode()
			Response(func() {
				Header("length:Content-Length")
				Header("encoding:Content-Type")
			})
			Response("invalid_file_path", StatusNotFound)
			Response("internal_error", StatusInternalServerError)
		})
	})
})
