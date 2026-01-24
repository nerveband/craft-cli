package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const craftLogo = `
    ╭──────────────╮
    │              │
    │   ╭──────╮   │
    │   │ ──── │   │
    │   │ ──── │   │
    │   │ ───  │   │
    │   ╰──────╯   │
    │              │
    ╰──────────────╯
`

const craftLogoAlt = `
    ┌─────────────┐
    │  ┌───────┐  │
    │  │ ────  │  │
    │  │ ────  │  │
    │  │ ──    │  │
    │  └───────┘  │
    └─────────────┘
`

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup for first-time users",
	Long: `Interactive setup wizard to configure Craft CLI.

Guides you through:
  - Getting your API URL from the Craft app
  - Creating your first profile
  - Verifying the connection works

Use --quiet to skip interactive prompts (for scripts).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if isQuiet() {
			return fmt.Errorf("setup requires interactive mode. Remove --quiet flag")
		}
		return runSetup()
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup() error {
	reader := bufio.NewReader(os.Stdin)

	// Show welcome
	printWelcome()

	// Check if already configured
	profiles, _ := cfgManager.ListProfiles()
	if len(profiles) > 0 {
		fmt.Println("You already have profiles configured:")
		for _, p := range profiles {
			marker := "  "
			if p.Active {
				marker = "* "
			}
			fmt.Printf("  %s%s\n", marker, p.Name)
		}
		fmt.Println()

		if !promptYesNo(reader, "Do you want to add another profile?", false) {
			fmt.Println("\nRun 'craft config list' to see your profiles.")
			return nil
		}
		fmt.Println()
	}

	// Show instructions
	printInstructions()

	// Offer to open Craft app (macOS only)
	if runtime.GOOS == "darwin" && isCraftInstalled() {
		fmt.Println()
		if promptYesNo(reader, "Would you like to open the Craft app now?", true) {
			openCraftApp()
			fmt.Println("  Craft is opening...")
			fmt.Println()
		}
	}

	// Wait for user to have URL ready
	fmt.Println()
	if !promptYesNo(reader, "Do you have your API URL ready?", true) {
		printLaterInstructions()
		return nil
	}

	// Get API URL
	fmt.Println()
	fmt.Print("Paste your Craft API URL: ")
	apiURL, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	apiURL = strings.TrimSpace(apiURL)

	if apiURL == "" {
		return fmt.Errorf("API URL cannot be empty")
	}

	// Get profile name
	fmt.Println()
	fmt.Print("Give this profile a name (e.g., work, personal): ")
	profileName, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	profileName = strings.TrimSpace(profileName)

	if profileName == "" {
		profileName = "default"
		fmt.Printf("  Using name: %s\n", profileName)
	}

	// Save profile
	if err := cfgManager.AddProfile(profileName, apiURL); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	// Success!
	printSuccess(profileName)

	return nil
}

func printWelcome() {
	fmt.Println()
	fmt.Print(craftLogoAlt)
	fmt.Println("    C R A F T   C L I")
	fmt.Println()
	fmt.Println("  Welcome! Let's get you set up.")
	fmt.Println()
}

func printInstructions() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("  HOW TO GET YOUR API URL:")
	fmt.Println()
	fmt.Println("  1. Open Craft and go to the \"Imagine\" tab (sidebar)")
	fmt.Println("  2. Click \"Add Your First API Connection\"")
	fmt.Println("  3. Give it a name (e.g., \"CLI Access\")")
	fmt.Println("  4. Click \"Add Document\" to select docs to share")
	fmt.Println("  5. Copy the API URL shown at the top")
	fmt.Println()
	fmt.Println("  Full guide: https://support.craft.do/hc/en-us/articles/23702897811612")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func printLaterInstructions() {
	fmt.Println()
	fmt.Println("No problem! When you're ready, run one of these:")
	fmt.Println()
	fmt.Println("  craft setup                    # Run this wizard again")
	fmt.Println("  craft config add <name> <url>  # Add profile directly")
	fmt.Println()
	fmt.Println("Guide: https://support.craft.do/hc/en-us/articles/23702897811612")
	fmt.Println()
}

func printSuccess(profileName string) {
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Printf("  ✓ Profile '%s' created and set as active!\n", profileName)
	fmt.Println()
	fmt.Println("  You're all set. Try these commands:")
	fmt.Println()
	fmt.Println("    craft list           List your documents")
	fmt.Println("    craft info           See connection info")
	fmt.Println("    craft config list    Manage profiles")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}

func promptYesNo(reader *bufio.Reader, question string, defaultYes bool) bool {
	defaultHint := "Y/n"
	if !defaultYes {
		defaultHint = "y/N"
	}

	fmt.Printf("%s (%s): ", question, defaultHint)

	answer, err := reader.ReadString('\n')
	if err != nil {
		return defaultYes
	}

	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer == "" {
		return defaultYes
	}

	return answer == "y" || answer == "yes"
}

func isCraftInstalled() bool {
	// Check if Craft.app exists in Applications
	_, err := os.Stat("/Applications/Craft.app")
	return err == nil
}

func openCraftApp() {
	exec.Command("open", "-a", "Craft").Start()
}

// checkFirstRun checks if this is a first run and offers setup
// Returns true if setup was run successfully, false otherwise
func checkFirstRun() bool {
	// Don't interrupt if quiet mode
	if isQuiet() {
		return false
	}

	// Check if we have any profiles
	profiles, err := cfgManager.ListProfiles()
	if err != nil || len(profiles) > 0 {
		return false
	}

	// First run - offer setup
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Print(craftLogoAlt)
	fmt.Println("    C R A F T   C L I")
	fmt.Println()
	fmt.Println("  Welcome! It looks like this is your first time.")
	fmt.Println()

	if promptYesNo(reader, "Would you like to set up Craft CLI now?", true) {
		fmt.Println()
		if err := runSetup(); err == nil {
			return true
		}
	}

	fmt.Println()
	fmt.Println("No problem! Run 'craft setup' when you're ready.")
	fmt.Println()
	os.Exit(0) // Exit cleanly instead of showing error
	return false
}
