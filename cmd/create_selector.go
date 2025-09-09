/*
Copyright © 2025 Shelton Louis

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

// Package cmd provides command-line interface implementations for the JavaScript package delegator.
package cmd

import (
	// standard library

	"fmt"

	// external
	"github.com/charmbracelet/huh"
	"github.com/samber/lo"

	// internal
	"github.com/louiss0/javascript-package-delegator/services"
)

// CreateAppSearcher defines how to search for create app packages.
// This interface is defined at the point of use (in the cmd package).
type CreateAppSearcher interface {
	SearchCreateApps(query string, size int) ([]services.PackageInfo, error)
}

// CreateAppSelector provides an interface for selecting a create app package.
// It follows Go Writing Philosophy: defined at point of use, with clean methods.
type CreateAppSelector interface {
	Run() error
	Value() string
}

// createAppSelector is a private struct implementing CreateAppSelector.
// All fields are unexported to comply with Go Writing Philosophy.
type createAppSelector struct {
	sel   *huh.Select[string]
	value string
}

// NewCreateAppSelector creates a new CreateAppSelector with pre-fetched packages and title.
// Constructor returns the interface (struct stays private) following Go Writing Philosophy.
func NewCreateAppSelector(packageInfo []services.PackageInfo, title string) (CreateAppSelector, error) {
	if len(packageInfo) == 0 {
		return nil, fmt.Errorf("no packages available for selection")
	}

	// Map []services.PackageInfo to []huh.Option[string] using lo.Map
	opts := lo.Map(packageInfo, func(p services.PackageInfo, _ int) huh.Option[string] {
		// Use Name for both label and value; add description for better UX
		return huh.NewOption(p.Name, p.Name+" - "+p.Description)
	})

	// Build the huh.Select with Title and Options
	sel := huh.NewSelect[string]().
		Title(title).
		Options(opts...)

	// Return createAppSelector as CreateAppSelector interface
	return &createAppSelector{sel: sel}, nil
}

// Run executes the interactive UI and stores the selected value.
// Uses pointer receiver since Run mutates internal state via s.value.
func (s *createAppSelector) Run() error {
	// Bind the pointer - huh.Select.Value takes *string
	s.sel.Value(&s.value)
	return s.sel.Run()
}

// Value returns the selected package name.
// No Get prefix - follows Go naming conventions.
func (s *createAppSelector) Value() string {
	return s.value
}
