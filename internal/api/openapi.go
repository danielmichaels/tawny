package api

import (
	"context"
	"github.com/danielmichaels/tawny"
	"github.com/danielmichaels/tawny/gen/openapi"
	"github.com/danielmichaels/tawny/internal/logger"
	"io"
)

// openapi service example implementation.
// The example methods log the requests and return zero values.
type openapisrvc struct {
	logger *logger.Logger
}

// NewOpenapi returns the openapi service implementation.
func NewOpenapi(logger *logger.Logger) openapi.Service {
	return &openapisrvc{logger}
}

func (o openapisrvc) File(ctx context.Context) (res *openapi.FileResult, body io.ReadCloser, err error) {
	f, err := assets.EmbeddedFiles.Open("gen/http/openapi3.json")
	if err != nil {
		return nil, nil, openapi.MakeInvalidFilePath(err)
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, nil, openapi.MakeInternalError(err)
	}
	return &openapi.FileResult{
		Length:   fi.Size(),
		Encoding: "application/json",
	}, f, nil
}

func (o openapisrvc) Documentation(ctx context.Context) (res *openapi.DocumentationResult, body io.ReadCloser, err error) {
	f, err := assets.EmbeddedFiles.Open("assets/static/docs.html")
	if err != nil {
		return nil, nil, openapi.MakeInvalidFilePath(err)
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, nil, openapi.MakeInternalError(err)
	}
	return &openapi.DocumentationResult{
		Length:   fi.Size(),
		Encoding: "text/html",
	}, f, nil
}
