package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var patchJSONBody string

var patchCmd = &cobra.Command{
	Use:   "patch [url]",
	Short: "Send a PATCH request",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Error: URL required")
			return
		}

		url := args[0]
		reqBody := bytes.NewBuffer([]byte(patchJSONBody))
		start := time.Now()

		req, _ := http.NewRequest("PATCH", url, reqBody)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		res, err := client.Do(req)

		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		defer res.Body.Close()

		duration := time.Since(start)
		body, _ := io.ReadAll(res.Body)

		fmt.Printf("\nStatus: %d (%s)\n\n", res.StatusCode, duration)

		var pretty map[string]interface{}
		if json.Unmarshal(body, &pretty) == nil {
			b, _ := json.MarshalIndent(pretty, "", "  ")
			fmt.Println(string(b))
		} else {
			fmt.Println(string(body))
		}
	},
}

func init() {
	patchCmd.Flags().StringVar(&patchJSONBody, "json", "", "JSON payload")
	rootCmd.AddCommand(patchCmd)
}
