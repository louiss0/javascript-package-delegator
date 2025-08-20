package services

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNpmRegistryService_SearchPackages_HappyPath(t *testing.T) {
	// Create a mock server that returns a valid npm registry response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		assert.Equal(t, "/search", r.URL.Path)
		assert.Equal(t, "react", r.URL.Query().Get("text"))
		assert.Equal(t, "35", r.URL.Query().Get("size"))

		// Return a mock npm registry response
		response := `{
			"objects": [
				{
					"package": {
						"name": "react",
						"version": "18.2.0",
						"description": "React is a JavaScript library for building user interfaces.",
						"links": {
							"repository": "https://github.com/facebook/react",
							"homepage": "https://reactjs.org/",
							"npm": "https://www.npmjs.com/package/react"
						}
					}
				},
				{
					"package": {
						"name": "react-dom",
						"version": "18.2.0",
						"description": "React package for working with the DOM.",
						"links": {
							"repository": "https://github.com/facebook/react",
							"homepage": "",
							"npm": "https://www.npmjs.com/package/react-dom"
						}
					}
				}
			]
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	// Create service with the mock server
	service := NewNpmRegistryServiceWithClient(http.DefaultClient, server.URL+"/search")

	// Test the search
	packages, err := service.SearchPackages("react")

	assert.NoError(t, err)
	assert.Len(t, packages, 2)

	// Verify first package
	assert.Equal(t, "react", packages[0].Name)
	assert.Equal(t, "18.2.0", packages[0].Version)
	assert.Equal(t, "React is a JavaScript library for building user interfaces.", packages[0].Description)
	assert.Equal(t, "https://reactjs.org/", packages[0].Homepage)

	// Verify second package (homepage fallback logic)
	assert.Equal(t, "react-dom", packages[1].Name)
	assert.Equal(t, "18.2.0", packages[1].Version)
	assert.Equal(t, "React package for working with the DOM.", packages[1].Description)
	assert.Equal(t, "https://github.com/facebook/react", packages[1].Homepage) // Should fallback to repository
}

func TestNpmRegistryService_SearchPackages_EmptyResults(t *testing.T) {
	// Create a mock server that returns empty results
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{"objects": []}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	service := NewNpmRegistryServiceWithClient(http.DefaultClient, server.URL+"/search")

	packages, err := service.SearchPackages("nonexistentpackage123456")

	assert.NoError(t, err)
	assert.Len(t, packages, 0)
}

func TestNpmRegistryService_SearchPackages_MalformedJSON(t *testing.T) {
	// Create a mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"objects": [invalid json`))
	}))
	defer server.Close()

	service := NewNpmRegistryServiceWithClient(http.DefaultClient, server.URL+"/search")

	packages, err := service.SearchPackages("test")

	assert.Error(t, err)
	assert.Nil(t, packages)
	assert.Contains(t, err.Error(), "failed to parse npm registry response")
}

func TestNpmRegistryService_SearchPackages_HTTPError(t *testing.T) {
	// Create a mock server that returns HTTP 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	service := NewNpmRegistryServiceWithClient(http.DefaultClient, server.URL+"/search")

	packages, err := service.SearchPackages("test")

	assert.Error(t, err)
	assert.Nil(t, packages)
	assert.Contains(t, err.Error(), "npm registry returned status 500")
	assert.Contains(t, err.Error(), "Internal Server Error")
}

func TestNpmRegistryService_SearchPackages_EmptyPattern(t *testing.T) {
	service := NewNpmRegistryService()

	packages, err := service.SearchPackages("")

	assert.Error(t, err)
	assert.Nil(t, packages)
	assert.Contains(t, err.Error(), "search pattern cannot be empty")
}

func TestNpmRegistryService_SearchPackages_NetworkError(t *testing.T) {
	// Use an invalid URL to simulate network error
	service := NewNpmRegistryServiceWithClient(http.DefaultClient, "http://invalid-url-that-should-not-exist.local/search")

	packages, err := service.SearchPackages("test")

	assert.Error(t, err)
	assert.Nil(t, packages)
	assert.Contains(t, err.Error(), "failed to make HTTP request to npm registry")
}

func TestNpmRegistryService_SearchPackages_HomepageFallbackLogic(t *testing.T) {
	// Test the homepage fallback logic: homepage -> repository -> npm
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"objects": [
				{
					"package": {
						"name": "package-with-homepage",
						"version": "1.0.0",
						"description": "Package with homepage",
						"links": {
							"repository": "https://github.com/user/repo",
							"homepage": "https://example.com",
							"npm": "https://www.npmjs.com/package/test"
						}
					}
				},
				{
					"package": {
						"name": "package-with-repo-only",
						"version": "1.0.0",
						"description": "Package with repo only",
						"links": {
							"repository": "https://github.com/user/repo2",
							"homepage": "",
							"npm": "https://www.npmjs.com/package/test2"
						}
					}
				},
				{
					"package": {
						"name": "package-with-npm-only",
						"version": "1.0.0",
						"description": "Package with npm only",
						"links": {
							"repository": "",
							"homepage": "",
							"npm": "https://www.npmjs.com/package/test3"
						}
					}
				}
			]
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	service := NewNpmRegistryServiceWithClient(http.DefaultClient, server.URL+"/search")

	packages, err := service.SearchPackages("test")

	assert.NoError(t, err)
	assert.Len(t, packages, 3)

	// First package should use homepage
	assert.Equal(t, "https://example.com", packages[0].Homepage)

	// Second package should fallback to repository
	assert.Equal(t, "https://github.com/user/repo2", packages[1].Homepage)

	// Third package should fallback to npm
	assert.Equal(t, "https://www.npmjs.com/package/test3", packages[2].Homepage)
}

func TestNewNpmRegistryService(t *testing.T) {
	service := NewNpmRegistryService()

	// Type assert to access private fields for testing
	impl, ok := service.(*npmRegistryServiceImpl)
	assert.True(t, ok, "Service should be of type *npmRegistryServiceImpl")
	assert.NotNil(t, impl.client, "HTTP client should be initialized")
	assert.Equal(t, "https://registry.npmjs.com/-/v1/search", impl.baseSearchURL, "Base URL should be set to npm registry")
}

func TestNewNpmRegistryServiceWithClient(t *testing.T) {
	customClient := &http.Client{}
	customURL := "https://custom-registry.example.com/search"

	service := NewNpmRegistryServiceWithClient(customClient, customURL)

	// Type assert to access private fields for testing
	impl, ok := service.(*npmRegistryServiceImpl)
	assert.True(t, ok, "Service should be of type *npmRegistryServiceImpl")
	assert.Equal(t, customClient, impl.client, "Custom HTTP client should be used")
	assert.Equal(t, customURL, impl.baseSearchURL, "Custom base URL should be used")
}
