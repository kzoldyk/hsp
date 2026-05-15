/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var aliasMap = map[string]string{
	"r":  "request",
	"g":  "get",
	"p":  "post",
	"pu": "put",
	"pa": "patch",
	"d":  "delete",
}

func normalizeAlias(f *pflag.FlagSet, name string) pflag.NormalizedName {
	if replacement, ok := aliasMap[name]; ok {
		return pflag.NormalizedName(replacement)
	}
	return pflag.NormalizedName(name)
}

func aliasCmd(use string, target *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: "Alias for " + target.Name(),
		Run:   target.Run,
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hsp",
	Short: "HTTP Superpowers - Easiest HTTP client in the terminal",
	Long: `HSP is an interactive HTTP client that makes API testing as easy as Postman, but in your terminal.

No need to remember curl syntax - just run 'hsp request' and answer simple prompts!

Features:
  • Interactive request builder - step-by-step guided flow
  • Auto-format JSON bodies and set Content-Type headers
  • Easy header and query parameter management
  • Request preview before sending
  • Automatic request history
  • Pretty-printed JSON responses

Examples:
  hsp request          - Start interactive request builder
  hsp get <url>        - Quick GET request
  hsp post <url>       - Quick POST request
  hsp put <url>        - Quick PUT request
  hsp patch <url>      - Quick PATCH request
  hsp delete <url>     - Quick DELETE request`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetGlobalNormalizationFunc(normalizeAlias)
	rootCmd.AddCommand(varCmd)
	rootCmd.AddCommand(envCmd)
	rootCmd.AddCommand(aliasCmd("r", requestCmd))
	rootCmd.AddCommand(aliasCmd("g", getCmd))
	rootCmd.AddCommand(aliasCmd("p", postCmd))
	rootCmd.AddCommand(aliasCmd("pu", putCmd))
	rootCmd.AddCommand(aliasCmd("pa", patchCmd))
	rootCmd.AddCommand(aliasCmd("d", delCmd))
}
