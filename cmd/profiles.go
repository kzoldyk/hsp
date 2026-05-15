package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type Profile struct {
	Name        string            `json:"name"`
	URL         string            `json:"url"`
	Method     string            `json:"method"`
	Headers    map[string]string `json:"headers"`
	QueryParams map[string]string `json:"params"`
	Body        string            `json:"body"`
	BodyFormat  string           `json:"body_format"`
	CreatedAt   string           `json:"created_at"`
	UpdatedAt   string           `json:"updated_at"`
}

func ProfileDir() string {
	homeDir := os.ExpandEnv("$HOME")
	return filepath.Join(homeDir, ".hsp", "profiles")
}

func profilePath(name string) string {
	return filepath.Join(ProfileDir(), name+".json")
}

func LoadProfile(name string) (*Profile, error) {
	path := profilePath(name)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var profile Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

func SaveProfile(p *Profile) error {
	dir := ProfileDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := profilePath(p.Name)
	p.UpdatedAt = time.Now().Format(time.RFC3339)

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	return nil
}

func DeleteProfile(name string) error {
	path := profilePath(name)
	return os.Remove(path)
}

func ListProfiles() []string {
	dir := ProfileDir()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) > 5 && name[len(name)-5:] == ".json" {
			names = append(names, name[:len(name)-5])
		}
	}

	return names
}

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage request profiles",
	Long:  "Save, load, run, edit, and delete named request profiles",
}

var profileSaveCmd = &cobra.Command{
	Use:   "save <name>",
	Short: "Save current request as a named profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			if _, err := LoadProfile(name); err == nil {
				color.Red("✗ Profile '%s' already exists. Use --force to overwrite.", name)
				return
			}
		}

		lastReq, err := MustLoadLastRequest()
		if err != nil {
			color.Red("✗ No request to save. Run 'hsp request' first.")
			return
		}

		profile := &Profile{
			Name:        name,
			URL:         lastReq.URL,
			Method:      lastReq.Method,
			Headers:     lastReq.Headers,
			QueryParams: lastReq.QueryParams,
			Body:        lastReq.Body,
			BodyFormat:  lastReq.BodyFormat,
			CreatedAt:   time.Now().Format(time.RFC3339),
			UpdatedAt:   time.Now().Format(time.RFC3339),
		}

		if err := SaveProfile(profile); err != nil {
			color.Red("✗ Failed to save profile: %v", err)
			return
		}

		color.Green("✓ Saved profile '%s'", name)
	},
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved profiles",
	Run: func(cmd *cobra.Command, args []string) {
		names := ListProfiles()

		if len(names) == 0 {
			fmt.Println("No profiles saved yet.")
			return
		}

		for _, name := range names {
			profile, err := LoadProfile(name)
			if err != nil {
				continue
			}

			fmt.Printf("%s - %s to %s", color.CyanString(name), profile.Method, profile.URL)

			if profile.UpdatedAt != "" {
				parsed, err := time.Parse(time.RFC3339, profile.UpdatedAt)
				if err == nil {
					fmt.Printf(" (updated: %s)", parsed.Format("2006-01-02"))
				}
			}

			fmt.Println()
		}
	},
}

var profileRunCmd = &cobra.Command{
	Use:   "run <name>",
	Short: "Run a saved profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		profile, err := LoadProfile(name)
		if err != nil {
			color.Red("✗ Profile '%s' not found", name)
			return
		}

		builder := &RequestBuilder{
			URL:         profile.URL,
			Method:      profile.Method,
			Headers:     profile.Headers,
			QueryParams: profile.QueryParams,
			Body:        profile.Body,
			BodyFormat:  profile.BodyFormat,
		}

		fmt.Printf("Running profile '%s': %s %s\n", name, builder.Method, builder.URL)
		builder.resolveVars()
		builder.SendRequest()
	},
}

var profileDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a saved profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if err := DeleteProfile(name); err != nil {
			color.Red("✗ Failed to delete profile: %v", err)
			return
		}

		color.Green("✓ Deleted profile '%s'", name)
	},
}

var profileEditCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit a profile in $EDITOR",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if _, err := LoadProfile(name); err != nil {
			color.Red("✗ Profile '%s' not found", name)
			return
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}

		path := profilePath(name)

		editCmd := exec.Command(editor, path)
		editCmd.Stdin = os.Stdin
		editCmd.Stdout = os.Stdout
		editCmd.Stderr = os.Stderr

		if err := editCmd.Run(); err != nil {
			color.Red("✗ Editor error: %v", err)
			return
		}

		color.Green("✓ Updated profile '%s'", name)
	},
}

func init() {
	profileCmd.AddCommand(profileSaveCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileRunCmd)
	profileCmd.AddCommand(profileDeleteCmd)
	profileCmd.AddCommand(profileEditCmd)

	profileSaveCmd.Flags().BoolP("force", "f", false, "Overwrite if profile exists")

	rootCmd.AddCommand(profileCmd)
}