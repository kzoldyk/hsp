package cmd

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var deleteHeaders []string

var delCmd = &cobra.Command{
	Use:   "delete [url]",
	Short: "Send a DELETE request",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Error: URL required")
			return
		}

		url := args[0]

		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			fmt.Println("Request Build Error:", err)
			return
		}

		for _, h := range deleteHeaders {
			parts := strings.SplitN(h, ":", 2)
			if len(parts) == 2 {
				req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}

		client := &http.Client{}
		start := time.Now()
		res, err := client.Do(req)
		if err != nil {
			fmt.Println("Request Error:", err)
			return
		}
		defer res.Body.Close()

		duration := time.Since(start)
		body, _ := io.ReadAll(res.Body)

		statusColor := color.New(color.FgGreen).Add(color.Bold)
		if res.StatusCode >= 400 {
			statusColor = color.New(color.FgRed).Add(color.Bold)
		}

		statusColor.Printf("\nStatus: %d (%s)\n\n", res.StatusCode, duration)
		fmt.Println(string(body))
	},
}

func init() {
	delCmd.Flags().StringArrayVarP(&deleteHeaders, "header", "H", []string{}, "Custom request headers")
	rootCmd.AddCommand(delCmd)
}
