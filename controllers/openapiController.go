package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"finalproject/docs"

	"github.com/gin-gonic/gin"
)

var publicOpenAPIPathPrefixes = []string{
	"/api/v1/auth",
	"/api/v1/me",
	"/api/v1/photos",
	"/api/v1/uploads",
	"/api/v1/comments",
	"/api/v1/social-media",
}

// PublicOpenAPISpec godoc
// @Summary Get public OpenAPI spec
// @Description Get a public OpenAPI spec that excludes admin and legacy endpoints
// @Tags docs
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Public OpenAPI spec"
// @Failure 500 {object} ErrorResponse "Failed to build public spec"
// @Router /openapi/public.json [get]
func PublicOpenAPISpec(c *gin.Context) {
	spec, err := buildPublicOpenAPISpec()
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to build public OpenAPI spec")
		return
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, spec)
}

func buildPublicOpenAPISpec() (map[string]interface{}, error) {
	spec := map[string]interface{}{}
	if err := json.Unmarshal([]byte(docs.SwaggerInfo.ReadDoc()), &spec); err != nil {
		return nil, err
	}

	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		spec["paths"] = map[string]interface{}{}
		return spec, nil
	}

	publicPaths := map[string]interface{}{}
	for path, routeSpec := range paths {
		if isPublicOpenAPIPath(path) {
			publicPaths[path] = routeSpec
		}
	}
	spec["paths"] = publicPaths
	pruneOpenAPIDefinitions(spec)

	return spec, nil
}

func isPublicOpenAPIPath(path string) bool {
	for _, prefix := range publicOpenAPIPathPrefixes {
		if path == prefix || strings.HasPrefix(path, prefix+"/") || strings.HasPrefix(path, prefix+"{") {
			return true
		}
	}

	return false
}

func pruneOpenAPIDefinitions(spec map[string]interface{}) {
	definitions, ok := spec["definitions"].(map[string]interface{})
	if !ok {
		return
	}

	referenced := map[string]bool{}
	collectOpenAPIRefs(spec["paths"], referenced)

	for {
		changed := false
		for name := range referenced {
			definition, exists := definitions[name]
			if !exists {
				continue
			}

			before := len(referenced)
			collectOpenAPIRefs(definition, referenced)
			if len(referenced) > before {
				changed = true
			}
		}

		if !changed {
			break
		}
	}

	publicDefinitions := map[string]interface{}{}
	for name := range referenced {
		if definition, exists := definitions[name]; exists {
			publicDefinitions[name] = definition
		}
	}
	spec["definitions"] = publicDefinitions
}

func collectOpenAPIRefs(value interface{}, refs map[string]bool) {
	switch typed := value.(type) {
	case map[string]interface{}:
		if ref, ok := typed["$ref"].(string); ok {
			if name, ok := strings.CutPrefix(ref, "#/definitions/"); ok && name != "" {
				refs[name] = true
			}
		}
		for _, child := range typed {
			collectOpenAPIRefs(child, refs)
		}
	case []interface{}:
		for _, child := range typed {
			collectOpenAPIRefs(child, refs)
		}
	}
}
