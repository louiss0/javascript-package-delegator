// Package services provides external service integrations for the JavaScript Package Delegator.
// It includes functionality for interacting with package registries and searching for packages.
package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/samber/lo" // Import samber/lo
)

// PackageInfo represents a simplified structure for package details from a registry search.
type PackageInfo struct {
	Name        string
	Version     string
	Description string
	Homepage    string // Can be repository or homepage URL
}

// NpmRegistryService defines the interface for interacting with the npm registry.
// It provides methods for searching and retrieving package information.
type NpmRegistryService interface {
	SearchPackages(pattern string) ([]PackageInfo, error)
	// Add other methods like GetPackageInfo(name string) if needed later
}

// npmRegistryServiceImpl is the concrete implementation of NpmRegistryService
// that interacts with the public npm registry API.
type npmRegistryServiceImpl struct {
	client *http.Client
	// baseSearchURL is the full base URL for the search endpoint, e.g., "https://registry.npmjs.com/-/v1/search"
	baseSearchURL string
}

// NewNpmRegistryService creates a new instance of NpmRegistryService
// with a default HTTP client suitable for production use.
func NewNpmRegistryService() *npmRegistryServiceImpl {
	return &npmRegistryServiceImpl{
		client: &http.Client{
			Timeout: 10 * time.Second, // Set a reasonable timeout for HTTP requests
		},
		baseSearchURL: "https://registry.npmjs.com/-/v1/search", // Default to real npm search API endpoint
	}
}

// NewNpmRegistryServiceWithClient allows injecting a custom HTTP client and base search URL for testing.
// This is primarily useful for unit tests where a mock server URL can be provided.
func NewNpmRegistryServiceWithClient(client *http.Client, baseSearchURL string) NpmRegistryService {
	return &npmRegistryServiceImpl{
		client:        client,
		baseSearchURL: baseSearchURL,
	}
}

// npmSearchResponse is the internal struct for decoding the npm search API response.
// It matches the structure returned by `https://registry.npmjs.com/-/v1/search`.
type npmSearchResponse struct {
	Objects []struct {
		Package struct {
			Name        string `json:"name"`
			Version     string `json:"version"`
			Description string `json:"description"`
			Links       struct {
				Repository string `json:"repository"`
				Homepage   string `json:"homepage"`
				Npm        string `json:"npm"`
			} `json:"links"`
		} `json:"package"`
	} `json:"objects"`
}

// SearchPackages searches the npm registry for packages matching the given pattern.
// It constructs the search URL, performs the HTTP GET request, and parses the JSON response
// into a slice of PackageInfo.
func (s *npmRegistryServiceImpl) SearchPackages(pattern string) ([]PackageInfo, error) {
	if pattern == "" {
		return nil, fmt.Errorf("search pattern cannot be empty")
	}

	// Construct the full URL using the service's baseSearchURL and query parameters.
	// This URL will be absolute, pointing either to the real registry or the mock server.
	url := fmt.Sprintf("%s?text=%s&size=35", s.baseSearchURL, pattern)

	// Use http.NewRequest and client.Do for more flexibility, though client.Get also works.
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request to npm registry: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body) // Read body for more info on error
		return nil, fmt.Errorf("npm registry returned status %d: %s (body: %s)", resp.StatusCode, resp.Status, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from npm registry: %w", err)
	}

	var searchResp npmSearchResponse
	if err := json.Unmarshal(bodyBytes, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse npm registry response: %w", err)
	}

	// Use samber/lo.Map to transform the response objects into PackageInfo slice.
	// This avoids a traditional for loop and makes the transformation concise.
	packages := lo.Map(searchResp.Objects, func(obj struct {
		Package struct {
			Name        string `json:"name"`
			Version     string `json:"version"`
			Description string `json:"description"`
			Links       struct {
				Repository string `json:"repository"`
				Homepage   string `json:"homepage"`
				Npm        string `json:"npm"`
			} `json:"links"`
		} `json:"package"`
	}, _ int,
	) PackageInfo {
		p := obj.Package
		homepage := p.Links.Homepage
		if homepage == "" {
			homepage = p.Links.Repository // Fallback to repository if no homepage
		}
		if homepage == "" {
			homepage = p.Links.Npm // Fallback to npm page
		}

		return PackageInfo{
			Name:        p.Name,
			Version:     p.Version,
			Description: p.Description,
			Homepage:    homepage,
		}
	})

	return packages, nil
}
