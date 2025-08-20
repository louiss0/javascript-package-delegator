package services_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"github.com/louiss0/javascript-package-delegator/services"
)

// This mockRoundTripper is a helper function to create a custom http.RoundTripper for mocking HTTP responses.
type mockRoundTripper func(req *http.Request) (*http.Response, error)

func (m mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m(req)
}

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
