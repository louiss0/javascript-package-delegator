package services_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"

	"github.com/louiss0/javascript-package-delegator/services"
)

// Test Suite setup
func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Services Suite")
}

// This mockRoundTripper is a helper function to create a custom http.RoundTripper for mocking HTTP responses.
type mockRoundTripper func(req *http.Request) (*http.Response, error)

func (m mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m(req)
}

// Ginkgo BDD Tests
var _ = Describe("NpmRegistryService", Label("slow", "integration"), func() {
	var (
		service    services.NpmRegistryService // Use interface type for service
		mockServer *httptest.Server
		assertT    = assert.New(GinkgoT()) // Initialize assert with GinkgoT()
	)

	AfterEach(func() {
		if mockServer != nil {
			mockServer.Close()
		}
	})

	Describe("SearchPackages", func() {
		Context("when the search is successful", func() {
			BeforeEach(func() {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// In the handler, the path is relative to the server's base URL
					assertT.Equal("/-/v1/search", r.URL.Path)
					assertT.Equal("react", r.URL.Query().Get("text"))
					assertT.Equal("35", r.URL.Query().Get("size"))

					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte(`
					{
						"objects": [
							{
								"package": {
									"name": "react",
									"version": "18.2.0",
									"description": "React is a JavaScript library for building user interfaces.",
									"links": {
										"repository": "https://github.com/facebook/react",
										"homepage": "https://react.dev/"
									}
								}
							},
							{
								"package": {
									"name": "react-dom",
									"version": "18.2.0",
									"description": "React package for working with the DOM.",
									"links": {
										"npm": "https://www.npmjs.com/package/react-dom"
									}
								}
							}
						]
					}`))
					assertT.NoError(err)
				}))
				// Use NewNpmRegistryServiceWithClient to inject the mock server's client and its URL
				// The baseSearchURL must be the full path including the API endpoint
				service = services.NewNpmRegistryServiceWithClient(mockServer.Client(), fmt.Sprintf("%s/-/v1/search", mockServer.URL))
			})

			It("should return a list of packages", func() {
				packages, err := service.SearchPackages("react")
				assertT.NoError(err)
				assertT.Len(packages, 2) // Corrected assertion
				assertT.Equal("react", packages[0].Name)
				assertT.Equal("18.2.0", packages[0].Version) // Corrected assertion to match mock data
				assertT.Contains(packages[0].Description, "JavaScript library")
				assertT.Equal("https://react.dev/", packages[0].Homepage) // Prefer homepage

				assertT.Equal("react-dom", packages[1].Name)
				assertT.Equal("https://www.npmjs.com/package/react-dom", packages[1].Homepage) // Fallback to npm link
			})
		})

		Context("when the search pattern is empty", func() {
			BeforeEach(func() {
				// For this test case, we don't need a mock server,
				// but initialize a default service to avoid nil pointer.
				service = services.NewNpmRegistryService()
			})

			It("should return an error", func() {
				packages, err := service.SearchPackages("")
				assertT.Error(err)
				assertT.Contains(err.Error(), "search pattern cannot be empty")
				assertT.Nil(packages)
			})
		})

		Context("when the HTTP request fails", func() {
			BeforeEach(func() {
				// Inject a client with a custom transport that always returns an error
				service = services.NewNpmRegistryServiceWithClient(&http.Client{
					Transport: mockRoundTripper(func(req *http.Request) (*http.Response, error) {
						return nil, fmt.Errorf("network error")
					}),
				}, "http://localhost:12345/-/v1/search") // Provide a dummy URL since it won't be hit anyway
			})

			It("should return an error", func() {
				packages, err := service.SearchPackages("fail")
				assertT.Error(err)
				assertT.Contains(err.Error(), "failed to make HTTP request")
				assertT.Nil(packages)
			})
		})

		Context("when the registry returns a non-200 status code", func() {
			BeforeEach(func() {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					_, err := w.Write([]byte("Server Error"))
					assertT.NoError(err)
				}))
				service = services.NewNpmRegistryServiceWithClient(mockServer.Client(), fmt.Sprintf("%s/-/v1/search", mockServer.URL))
			})

			It("should return an error with status and body", func() {
				packages, err := service.SearchPackages("nice")
				assertT.Error(err)
				assertT.Nil(packages)
				assertT.Contains(err.Error(), "npm registry returned status 500: 500 Internal Server Error (body: Server Error)")
			})
		})

		Context("when the response body is invalid JSON", func() {
			BeforeEach(func() {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte("invalid json"))
					assertT.NoError(err)
				}))
				service = services.NewNpmRegistryServiceWithClient(mockServer.Client(), fmt.Sprintf("%s/-/v1/search", mockServer.URL))
			})

			It("should return an error", func() {
				packages, err := service.SearchPackages("badjson")
				assertT.Error(err)
				assertT.Contains(err.Error(), "failed to parse npm registry response")
				assertT.Nil(packages)
			})
		})

		Context("when no packages are found", func() {
			BeforeEach(func() {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte(`{"objects": []}`))
					assertT.NoError(err)
				}))
				service = services.NewNpmRegistryServiceWithClient(mockServer.Client(), fmt.Sprintf("%s/-/v1/search", mockServer.URL))
			})

			It("should return an empty slice", func() {
				packages, err := service.SearchPackages("nonexistent")
				assertT.NoError(err)
				assertT.Empty(packages)
			})
		})
	})
})

// Standard Go Tests (from npm_registry_test.go)

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
	service := services.NewNpmRegistryServiceWithClient(http.DefaultClient, server.URL+"/search")

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

	service := services.NewNpmRegistryServiceWithClient(http.DefaultClient, server.URL+"/search")

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

	service := services.NewNpmRegistryServiceWithClient(http.DefaultClient, server.URL+"/search")

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

	service := services.NewNpmRegistryServiceWithClient(http.DefaultClient, server.URL+"/search")

	packages, err := service.SearchPackages("test")

	assert.Error(t, err)
	assert.Nil(t, packages)
	assert.Contains(t, err.Error(), "npm registry returned status 500")
	assert.Contains(t, err.Error(), "Internal Server Error")
}

func TestNpmRegistryService_SearchPackages_EmptyPattern(t *testing.T) {
	service := services.NewNpmRegistryService()

	packages, err := service.SearchPackages("")

	assert.Error(t, err)
	assert.Nil(t, packages)
	assert.Contains(t, err.Error(), "search pattern cannot be empty")
}

func TestNpmRegistryService_SearchPackages_NetworkError(t *testing.T) {
	// Use an invalid URL to simulate network error
	service := services.NewNpmRegistryServiceWithClient(http.DefaultClient, "http://invalid-url-that-should-not-exist.local/search")

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

	service := services.NewNpmRegistryServiceWithClient(http.DefaultClient, server.URL+"/search")

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
	service := services.NewNpmRegistryService()

	// Test that the service was created and is functional by doing a basic check
	assert.NotNil(t, service, "Service should be created")

	// Test with empty pattern to verify the service works
	packages, err := service.SearchPackages("")
	assert.Error(t, err)
	assert.Nil(t, packages)
	assert.Contains(t, err.Error(), "search pattern cannot be empty")
}

func TestNewNpmRegistryServiceWithClient(t *testing.T) {
	customClient := &http.Client{}
	customURL := "https://custom-registry.example.com/search"

	service := services.NewNpmRegistryServiceWithClient(customClient, customURL)

	// Test that the service was created and is functional
	assert.NotNil(t, service, "Service should be created")

	// Test with empty pattern to verify the service works
	packages, err := service.SearchPackages("")
	assert.Error(t, err)
	assert.Nil(t, packages)
	assert.Contains(t, err.Error(), "search pattern cannot be empty")
}
