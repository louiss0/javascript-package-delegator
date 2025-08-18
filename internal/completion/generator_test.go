package completion_test

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/louiss0/javascript-package-delegator/internal/completion"
)

func newGenCmdBuf() (completion.Generator, *cobra.Command, *bytes.Buffer) {
	return completion.NewGenerator(), &cobra.Command{Use: "test"}, &bytes.Buffer{}
}

func TestGenerator_GetSupportedShells(t *testing.T) {
	g := completion.NewGenerator()
	shells := g.GetSupportedShells()

	// Membership and count
	assert.Len(t, shells, 6)
	assert.Contains(t, shells, "bash")
	assert.Contains(t, shells, "carapace")
	assert.Contains(t, shells, "fish")
	assert.Contains(t, shells, "nushell")
	assert.Contains(t, shells, "powershell")
	assert.Contains(t, shells, "zsh")

	// Order
	expected := []string{"bash", "carapace", "fish", "nushell", "powershell", "zsh"}
	assert.Equal(t, expected, shells)
}

func TestGenerator_GetDefaultAliasMapping(t *testing.T) {
	g := completion.NewGenerator()
	aliasMap := g.GetDefaultAliasMapping()

	// Keys present
	assert.Contains(t, aliasMap, "install")
	assert.Contains(t, aliasMap, "run")
	assert.Contains(t, aliasMap, "exec")
	assert.Contains(t, aliasMap, "dlx")
	assert.Contains(t, aliasMap, "update")
	assert.Contains(t, aliasMap, "uninstall")
	assert.Contains(t, aliasMap, "clean-install")
	assert.Contains(t, aliasMap, "agent")

	// Specific aliases
	assert.Contains(t, aliasMap["exec"], "jpe")
	assert.Contains(t, aliasMap["dlx"], "jpx")

	// Long-form aliases
	assert.Contains(t, aliasMap["install"], "jpd-install")
	assert.Contains(t, aliasMap["run"], "jpd-run")
	assert.Contains(t, aliasMap["exec"], "jpd-exec")
	assert.Contains(t, aliasMap["dlx"], "jpd-dlx")
}

func TestGenerator_GenerateCompletion_BaseScripts(t *testing.T) {
	type tc struct {
		name         string
		shell        string
		substrings   []string
		nonEmptyOnly bool
	}
	cases := []tc{
		{name: "bash", shell: "bash", substrings: []string{"bash completion"}},
		{name: "zsh", shell: "zsh", substrings: []string{"zsh completion"}},
		{name: "fish", shell: "fish", substrings: []string{"fish completion"}},
		{name: "nushell", shell: "nushell", nonEmptyOnly: true},
		{name: "powershell", shell: "powershell", substrings: []string{"PowerShell"}},
		{name: "carapace", shell: "carapace", substrings: []string{"carapace completion bridge", "https://github.com/rsteube/carapace-bin"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g, cmd, buf := newGenCmdBuf()
			err := g.GenerateCompletion(cmd, c.shell, buf, false)
			assert.NoError(t, err)

			out := buf.String()
			if c.nonEmptyOnly {
				assert.NotEmpty(t, out)
				return
			}
			for _, s := range c.substrings {
				assert.Contains(t, out, s)
			}
		})
	}
}

func TestGenerator_GenerateCompletion_WithShorthand(t *testing.T) {
	type tc struct {
		name       string
		shell      string
		substrings []string
	}
	cases := []tc{
		{name: "bash", shell: "bash", substrings: []string{"function jpe()", "function jpx()", "function jpi()"}},
		{name: "zsh", shell: "zsh", substrings: []string{"jpe() { jpd exec", "jpx() { jpd dlx"}},
		{name: "fish", shell: "fish", substrings: []string{"function jpe", "function jpx", "jpd exec $argv"}},
		{name: "nushell", shell: "nushell", substrings: []string{"export def jpe", "export def jpx"}},
		{name: "powershell", shell: "powershell", substrings: []string{"function jpe {", "function jpx {", "jpd exec @args", "Register-ArgumentCompleter"}},
		{name: "carapace", shell: "carapace", substrings: []string{"function jpe()", "function jpx()"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g, cmd, buf := newGenCmdBuf()
			err := g.GenerateCompletion(cmd, c.shell, buf, true)
			assert.NoError(t, err)

			out := buf.String()
			for _, s := range c.substrings {
				assert.Contains(t, out, s)
			}
		})
	}
}

func TestGenerator_GenerateCompletion_UnsupportedShell(t *testing.T) {
	g, cmd, buf := newGenCmdBuf()
	err := g.GenerateCompletion(cmd, "unsupported", buf, false)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "unsupported shell: unsupported")
}

func Test_GetNushellCompletionScript(t *testing.T) {
	s := completion.GetNushellCompletionScript()
	assert.NotEmpty(t, s)
}

func Test_GenerateCarapaceBridge(t *testing.T) {
	b := completion.GenerateCarapaceBridge()
	assert.Contains(t, b, "carapace completion bridge")
	assert.Contains(t, b, "Setup Instructions")
	assert.Contains(t, b, "Bash/Zsh")
	assert.Contains(t, b, "Fish")
	assert.Contains(t, b, "Nushell")
	assert.Contains(t, b, "https://rsteube.github.io/carapace/")
}
