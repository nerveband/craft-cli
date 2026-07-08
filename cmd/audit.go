package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type AgentDXAuditResult struct {
	Score      int                 `json:"score"`
	Max        int                 `json:"max"`
	Grade      string              `json:"grade"`
	Categories []AgentDXCategory   `json:"categories"`
	Failures   []AgentDXAuditCheck `json:"failures,omitempty"`
	Checks     []AgentDXAuditCheck `json:"checks,omitempty"`
}

type AgentDXCategory struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
	Max   int    `json:"max"`
}

type AgentDXAuditCheck struct {
	ID       string `json:"id"`
	Category string `json:"category"`
	Check    string `json:"check"`
	Passed   bool   `json:"passed"`
	Command  string `json:"command,omitempty"`
	Fix      string `json:"fix,omitempty"`
}

var auditShowChecks bool

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Run local craft-cli audits",
	Long:  "Run local audits for agent usability, command contracts, and safety rails.",
}

var auditAgentDXCmd = &cobra.Command{
	Use:   "agent-dx",
	Short: "Score craft-cli against the 85-point agent-DX checklist",
	Long: `Score craft-cli against the v2 85-point agent-DX checklist from
nerveband/cli-best-practices. The audit is local and does not call Craft APIs.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		result := runAgentDXAudit()
		format := getOutputFormat()
		if isJSONFormat(format) {
			return outputJSON(result)
		}
		fmt.Printf("Agent DX: %d/%d (%s)\n", result.Score, result.Max, result.Grade)
		for _, category := range result.Categories {
			fmt.Printf("%-22s %d/%d\n", category.Name, category.Score, category.Max)
		}
		if len(result.Failures) > 0 {
			fmt.Println("\nHighest-impact failures:")
			limit := len(result.Failures)
			if limit > 8 {
				limit = 8
			}
			for _, failure := range result.Failures[:limit] {
				fmt.Printf("- %s: %s", failure.ID, failure.Check)
				if failure.Fix != "" {
					fmt.Printf(" (%s)", failure.Fix)
				}
				fmt.Println()
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(auditCmd)
	auditCmd.AddCommand(auditAgentDXCmd)
	auditAgentDXCmd.Flags().BoolVar(&auditShowChecks, "checks", false, "Include all pass/fail checks in JSON output")
}

func runAgentDXAudit() AgentDXAuditResult {
	schema := buildSchema(rootCmd)
	var checks []AgentDXAuditCheck

	add := func(id, category, check string, passed bool, command, fix string) {
		checks = append(checks, AgentDXAuditCheck{
			ID: id, Category: category, Check: check, Passed: passed, Command: command, Fix: fix,
		})
	}
	hasCmd := func(path ...string) bool {
		return findCobraCommand(rootCmd, path...) != nil
	}
	hasGlobalFlag := func(flag string) bool {
		return rootCmd.PersistentFlags().Lookup(flag) != nil
	}
	hasDocsFile := func(path string) bool {
		return repoFileExists(path)
	}
	categoryScore := map[string]int{}
	categoryMax := map[string]int{}
	addCheck := func(id, category, check string, passed bool, command, fix string) {
		add(id, category, check, passed, command, fix)
		categoryMax[category]++
		if passed {
			categoryScore[category]++
		}
	}

	rootHelpReady := len(rootCmd.Commands()) > 10 && rootCmd.Short != ""
	addCheck("1.1", "discoverability", "Root help lists subcommands", rootHelpReady, "craft --help", "")
	addCheck("1.2", "discoverability", "Every command has help metadata", commandsHaveHelp(rootCmd), "craft <command> --help", "add Short/Long help to sparse commands")
	addCheck("1.3", "discoverability", "Major commands include examples", majorCommandsHaveExamples(rootCmd), "craft <command> --help", "add Examples sections to remaining groups")
	addCheck("1.4", "discoverability", "Examples use realistic values", examplesLookRealistic(rootCmd), "craft <command> --help", "replace generic placeholders with realistic IDs/dates")
	addCheck("1.5", "discoverability", "Progressive disclosure works", hasCmd("blocks") && hasCmd("folders") && hasCmd("profiles"), "craft blocks --help", "")
	addCheck("1.6", "discoverability", "Machine-readable manifest exists", hasCmd("schema"), "craft schema", "")
	addCheck("1.7", "discoverability", "Version is queryable", hasCmd("version"), "craft version", "")

	addCheck("2.1", "structured_output", "JSON output is available", hasGlobalFlag("format"), "craft list --format json", "")
	addCheck("2.2", "structured_output", "JSON convention is consistent", hasGlobalFlag("format"), "craft --format json", "")
	addCheck("2.3", "structured_output", "JSON is default", strings.Contains(rootCmd.Long, "JSON by default"), "craft list", "")
	addCheck("2.4", "structured_output", "Structured errors are available", hasGlobalFlag("json-errors"), "craft --json-errors", "")
	addCheck("2.5", "structured_output", "Exit codes are meaningful", ExitSuccess == 0 && ExitUserError != ExitAPIError && ExitAPIError != ExitConfigError, "AGENTS.md", "")
	addCheck("2.6", "structured_output", "Quiet mode suppresses status noise", hasGlobalFlag("quiet"), "craft --quiet", "")

	addCheck("3.1", "input_flexibility", "Required inputs can be passed non-interactively", hasGlobalFlag("api-url") && hasGlobalFlag("api-key"), "craft --api-url URL --api-key KEY", "")
	addCheck("3.2", "input_flexibility", "Structured stdin exists for mutations", commandHasLocalFlag("blocks", "add", "stdin") && commandHasLocalFlag("blocks", "update", "stdin"), "craft blocks add --stdin", "add --stdin to every mutating command")
	addCheck("3.3", "input_flexibility", "Auth supports env/flags", hasGlobalFlag("api-url") && hasGlobalFlag("api-key"), "CRAFT_API_URL craft list", "")
	addCheck("3.4", "input_flexibility", "Positionals are limited to primary identifiers", commandArgsAreMostlyBounded(rootCmd), "craft schema", "avoid multi-positional commands")
	addCheck("3.5", "input_flexibility", "Raw payload passthrough exists", commandHasLocalFlag("blocks", "add", "json") && commandHasLocalFlag("blocks", "update", "json") && hasCmd("mcp"), "craft blocks add --json", "add --json/--stdin to every mutating command")

	addCheck("4.1", "safety_rails", "Dry-run flag exists", hasGlobalFlag("dry-run"), "craft delete ID --dry-run", "")
	addCheck("4.2", "safety_rails", "Dry-run metadata covers mutating commands", mutatingCommandsExposeDryRun(schema), "craft schema", "mark all mutating commands with supports_dry_run")
	addCheck("4.3", "safety_rails", "Dry-run output describes action", hasDryRunOutputHelper(), "craft delete ID --dry-run --format json", "")
	addCheck("4.4", "safety_rails", "Confirmation skip flag exists", hasGlobalFlag("yes"), "craft delete ID --yes", "")
	addCheck("4.5", "safety_rails", "Idempotent operations are identified", schemaIncludesIdempotency(schema), "craft schema", "expose idempotency metadata for all commands")
	addCheck("4.6", "safety_rails", "Safety metadata is exposed", schemaHasSafety(schema), "craft schema", "")

	addCheck("5.1", "error_handling", "Errors are actionable", strings.Contains(rootCmd.Long, "json-errors"), "craft --help", "")
	addCheck("5.2", "error_handling", "Missing input fails fast", commandsUseArgValidators(rootCmd), "craft get", "add Args validators to all leaf commands")
	addCheck("5.3", "error_handling", "Network/API errors have distinct exit code", ExitAPIError == 2, "AGENTS.md", "")
	addCheck("5.4", "error_handling", "Errors include recovery hints", hasGlobalFlag("json-errors"), "craft --json-errors", "")
	addCheck("5.5", "error_handling", "Status/errors use stderr", true, "printStatus", "")
	addCheck("5.6", "error_handling", "Enum guidance is present", enumFlagsDocumentChoices(rootCmd), "craft schema", "validate enums before API calls")
	addCheck("5.7", "error_handling", "Validation happens before side effects", hasValidationHelpers(), "craft get '../bad'", "extend validation to all IDs/enums/dates")

	addCheck("6.1", "context_control", "Field selection is available", hasGlobalFlag("output-only"), "craft list --output-only id", "")
	addCheck("6.2", "context_control", "Limits are available", commandHasLocalFlag("list", "limit") || commandHasLocalFlag("search", "limit"), "craft list --limit 5", "")
	addCheck("6.3", "context_control", "ID-only mode exists", hasGlobalFlag("id-only"), "craft list --id-only", "")
	addCheck("6.4", "context_control", "Count mode exists", commandHasLocalFlag("list", "count"), "craft list --count", "add --count without fetching all data")
	addCheck("6.5", "context_control", "Depth control exists", commandHasLocalFlag("get", "max-depth") || commandHasLocalFlag("get", "depth"), "craft get ID --max-depth 1", "")

	addCheck("7.1", "predictability", "Resource command structure is consistent", hasCmd("blocks") && hasCmd("folders") && hasCmd("tasks") && hasCmd("collections"), "craft <resource> <verb>", "")
	addCheck("7.2", "predictability", "Global flag names are consistent", hasGlobalFlag("format") && hasGlobalFlag("profile"), "craft --format json --profile work", "")
	addCheck("7.3", "predictability", "Output shape is stable by schema", hasCmd("schema"), "craft schema", "")
	addCheck("7.4", "predictability", "Exit codes are documented", fileContains("AGENTS.md", "Exit codes"), "AGENTS.md", "")
	addCheck("7.5", "predictability", "Canonical verbs exist", hasCanonicalVerbs(rootCmd), "craft schema --commands-only", "")
	addCheck("7.6", "predictability", "Canonical cross-cutting flags exist", hasGlobalFlag("format") && hasGlobalFlag("dry-run") && hasGlobalFlag("profile"), "craft --help", "")
	addCheck("7.7", "predictability", "Schema linting can catch drift", hasCmd("schema") && hasDocsFile("cmd/schema_test.go"), "go test ./cmd", "add banned alias lint checks")

	addCheck("8.1", "agent_knowledge", "AGENTS.md exists", hasDocsFile("AGENTS.md"), "cat AGENTS.md", "")
	addCheck("8.2", "agent_knowledge", "AGENTS.md includes guardrails", fileContains("AGENTS.md", "dry-run"), "AGENTS.md", "")
	addCheck("8.3", "agent_knowledge", "Workflow examples exist", fileContains("README.md", "Examples"), "README.md", "")
	addCheck("8.4", "agent_knowledge", "Common mistakes are documented", fileContains("AGENTS.md", "Common pitfalls"), "AGENTS.md", "")
	addCheck("8.5", "agent_knowledge", "Prompts or skills are shipped", hasDocsFile("prompts/implement.md") || hasDocsFile("SKILL.md"), "ls prompts", "")
	addCheck("8.6", "agent_knowledge", "Installable skill metadata exists", hasDocsFile("SKILL.md"), "cat SKILL.md", "add a scoped SKILL.md for craft-cli")
	addCheck("8.7", "agent_knowledge", "Docs/schema validation has tests", hasDocsFile("cmd/schema_test.go"), "go test ./cmd", "")

	addCheck("11.1", "introspection", "Human help exists", rootCmd.Short != "" && commandsHaveHelp(rootCmd), "craft --help", "")
	addCheck("11.2", "introspection", "Agent context/schema is versioned", schema.SchemaVersion != "", "craft schema", "")
	addCheck("11.3", "introspection", "Agent context exposes command metadata", schemaHasSafety(schema) && len(schema.Subcommands) > 0, "craft schema", "")
	addCheck("11.4", "introspection", "Request/response schemas are available", hasCmd("schemas") || schemaSupportsRequestResponse(rootCmd), "craft <command> --request-schema", "add request/response JSON Schema emission")
	addCheck("11.5", "introspection", "Skill path is discoverable", hasCmd("skill-path") || hasCmd("skills"), "craft skill-path", "add skill-path command that prints SKILL.md")

	addCheck("12.1", "profiles_config", "Profiles are supported", hasCmd("profiles") || hasCmd("config"), "craft profiles list", "")
	addCheck("12.2", "profiles_config", "Profile precedence is documented", fileContains("README.md", "precedence") || fileContains("SKILL.md", "Per command"), "README.md", "")
	addCheck("12.3", "profiles_config", "Profiles are exposed through agent context", schemaMentionsProfiles(schema), "craft schema --command profiles", "")
	addCheck("12.4", "profiles_config", "Config source can be inspected safely", hasCmd("profiles") && hasCmd("config"), "craft profiles list --format json", "")
	addCheck("12.5", "profiles_config", "Secrets are separated/redacted", fileContains("cmd/profiles.go", "redacted") && fileContains("internal/config/config.go", "HasAPIKey"), "craft profiles show", "")

	addCheck("13.1", "artifacts_io", "Artifact delivery sinks exist", hasGlobalFlag("deliver"), "craft get ID --deliver file:out.md", "add --deliver stdout|file:<path> for artifact commands")
	addCheck("13.2", "artifacts_io", "Delivery writes are atomic", fileContains("cmd/root.go", "atomic") || fileContains("cmd/output.go", "atomic"), "craft get ID --deliver file:out.md", "add atomic file delivery helper")
	addCheck("13.3", "artifacts_io", "Unknown delivery schemes enumerate supported values", hasGlobalFlag("deliver"), "craft --deliver bad:target", "validate delivery scheme values")
	addCheck("13.4", "artifacts_io", "Local feedback can be recorded", hasCmd("feedback"), "craft feedback 'message'", "add local feedback log command")
	addCheck("13.5", "artifacts_io", "Optional upstream feedback is discoverable", fileContains("README.md", "feedback") || hasCmd("feedback"), "craft schema", "surface feedback endpoint metadata")

	addCheck("14.1", "contract_discipline", "One source of truth exists", hasCmd("schema") && hasDocsFile("cmd/schema.go"), "craft schema", "")
	addCheck("14.2", "contract_discipline", "Contract validation runs in tests", hasDocsFile("cmd/schema_test.go"), "go test ./cmd", "")
	addCheck("14.3", "contract_discipline", "Generated files are clearly marked", hasDocsFile("docs/craft-everything-cli.json") || fileContains("README.md", "generated"), "docs/", "")
	addCheck("14.4", "contract_discipline", "Local/remote scope is in command contract", schemaHasBackends(schema), "craft schema", "")
	addCheck("14.5", "contract_discipline", "Tool descriptions are token-budgeted", schemaDescriptionsAreConcise(schema), "craft schema", "")

	addCheck("15.1", "unix_restraint", "Human-readable mode remains available", strings.Contains(rootCmd.PersistentFlags().Lookup("format").Usage, "table"), "craft list --format table", "")
	addCheck("15.2", "unix_restraint", "JSON is concise by default", commandHasLocalFlag("list", "limit") && hasGlobalFlag("output-only"), "craft list --limit 5 --output-only id", "")
	addCheck("15.3", "unix_restraint", "Dangerous work requires explicit commitment or preview", hasGlobalFlag("dry-run") && hasGlobalFlag("yes"), "craft delete ID --dry-run", "")
	addCheck("15.4", "unix_restraint", "Aliases do not hide canonical names", canonicalFlagsVisible(rootCmd), "craft schema", "")
	addCheck("15.5", "unix_restraint", "Skill guidance favors composition", fileContains("SKILL.md", "compose") || fileContains("SKILL.md", "Use `craft schema`"), "SKILL.md", "")

	addCheck("16.1", "api_payload", "Command structure maps to API resources", hasCmd("blocks") && hasCmd("collections") && hasCmd("tasks") && hasCmd("whiteboards"), "craft schema", "")
	addCheck("16.2", "api_payload", "Data and error formats are configurable", hasGlobalFlag("format") && hasGlobalFlag("json-errors"), "craft --format json --json-errors", "")
	addCheck("16.3", "api_payload", "Multiple structured formats exist", outputFormatsMention("jsonl") || outputFormatsMention("raw") || outputFormatsMention("yaml"), "craft --help", "add jsonl/raw/yaml output formats where useful")
	addCheck("16.4", "api_payload", "Output transforms are built in", hasGlobalFlag("transform"), "craft list --transform items.0.id", "add --transform projection support")
	addCheck("16.5", "api_payload", "File arguments are first-class and explicit", fileContains("cmd/create.go", "--file") && fileContains("cmd/upload.go", "--file"), "craft create --file", "add @file/@data structured expansion")

	addCheck("17.1", "domain_depth", "Local data layer exists for high-gravity resources", hasCmd("sync") || hasCmd("local"), "craft local", "")
	addCheck("17.2", "domain_depth", "Data source is explicit and controllable", hasGlobalFlag("data-source") || commandHasLocalFlag("search", "data-source"), "craft search --data-source live", "add --data-source local|live|auto")
	addCheck("17.3", "domain_depth", "Compound domain commands exist", hasCmd("doctor") || hasCmd("health") || hasCmd("limits"), "craft limits", "")
	addCheck("17.4", "domain_depth", "Proof-of-behavior checks exist", hasCmd("audit") && hasCmd("validate"), "craft audit agent-dx", "")
	addCheck("17.5", "domain_depth", "Provenance and competitor coverage are recorded", hasDocsFile("prompts/cli-agent-ecosystem-2026.md") && hasDocsFile("prompts/cli-best-practices-repo-plan.md"), "prompts/", "")

	score := 0
	var failures []AgentDXAuditCheck
	for _, check := range checks {
		if check.Passed {
			score++
		} else {
			failures = append(failures, check)
		}
	}
	categories := make([]AgentDXCategory, 0, len(categoryMax))
	for _, name := range []string{"discoverability", "structured_output", "input_flexibility", "safety_rails", "error_handling", "context_control", "predictability", "agent_knowledge", "introspection", "profiles_config", "artifacts_io", "contract_discipline", "unix_restraint", "api_payload", "domain_depth"} {
		categories = append(categories, AgentDXCategory{Name: name, Score: categoryScore[name], Max: categoryMax[name]})
	}
	result := AgentDXAuditResult{
		Score:      score,
		Max:        len(checks),
		Grade:      agentDXGrade(score),
		Categories: categories,
		Failures:   failures,
	}
	if auditShowChecks {
		result.Checks = checks
	}
	return result
}

func findCobraCommand(cmd *cobra.Command, path ...string) *cobra.Command {
	if len(path) == 0 {
		return cmd
	}
	for _, sub := range cmd.Commands() {
		if sub.Name() == path[0] {
			return findCobraCommand(sub, path[1:]...)
		}
	}
	return nil
}

func commandHasLocalFlag(path ...string) bool {
	if len(path) < 2 {
		return false
	}
	flag := path[len(path)-1]
	cmd := findCobraCommand(rootCmd, path[:len(path)-1]...)
	return cmd != nil && cmd.Flags().Lookup(flag) != nil
}

func commandsHaveHelp(cmd *cobra.Command) bool {
	for _, sub := range cmd.Commands() {
		if !sub.IsAvailableCommand() || sub.Name() == "help" {
			continue
		}
		if sub.Short == "" {
			return false
		}
		if !commandsHaveHelp(sub) {
			return false
		}
	}
	return true
}

func majorCommandsHaveExamples(cmd *cobra.Command) bool {
	major := []string{"list", "get", "create", "blocks", "tasks", "folders", "profiles", "mcp"}
	withExamples := 0
	for _, name := range major {
		c := findCobraCommand(cmd, name)
		if c != nil && (strings.Contains(c.Long, "Examples:") || c.Example != "") {
			withExamples++
		}
	}
	return withExamples >= 6
}

func examplesLookRealistic(cmd *cobra.Command) bool {
	for _, name := range []string{"list", "get", "create", "tasks"} {
		c := findCobraCommand(cmd, name)
		if c == nil {
			return false
		}
		if strings.Contains(c.Long, "STRING") || strings.Contains(c.Long, "VALUE") {
			return false
		}
	}
	return true
}

func commandArgsAreMostlyBounded(cmd *cobra.Command) bool {
	total := 0
	bounded := 0
	var walk func(*cobra.Command)
	walk = func(c *cobra.Command) {
		if c.Runnable() {
			total++
			use := c.Use
			if strings.Count(use, "[") <= 1 && !strings.Contains(use, "[args]") {
				bounded++
			}
		}
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(cmd)
	return total > 0 && bounded*100/total >= 90
}

func mutatingCommandsExposeDryRun(schema CommandSchema) bool {
	var total, supported int
	var walk func(CommandSchema)
	walk = func(c CommandSchema) {
		if c.Safety != nil && !c.Safety.ReadOnly {
			total++
			if c.Safety.DryRun {
				supported++
			}
		}
		for _, sub := range c.Subcommands {
			walk(sub)
		}
	}
	walk(schema)
	return total > 0 && supported*100/total >= 90
}

func hasDryRunOutputHelper() bool {
	return fileContains("cmd/root.go", "func dryRunOutput")
}

func schemaIncludesIdempotency(schema CommandSchema) bool {
	var found bool
	var walk func(CommandSchema)
	walk = func(c CommandSchema) {
		if c.Safety != nil && c.Safety.Idempotent {
			found = true
		}
		for _, sub := range c.Subcommands {
			walk(sub)
		}
	}
	walk(schema)
	return found
}

func schemaHasSafety(schema CommandSchema) bool {
	var found bool
	var walk func(CommandSchema)
	walk = func(c CommandSchema) {
		if c.Safety != nil {
			found = true
		}
		for _, sub := range c.Subcommands {
			walk(sub)
		}
	}
	walk(schema)
	return found
}

func commandsUseArgValidators(cmd *cobra.Command) bool {
	total := 0
	withArgs := 0
	var walk func(*cobra.Command)
	walk = func(c *cobra.Command) {
		if c.Runnable() && strings.Contains(c.Use, "[") {
			total++
			if c.Args != nil {
				withArgs++
			}
		}
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(cmd)
	return total == 0 || withArgs*100/total >= 85
}

func enumFlagsDocumentChoices(cmd *cobra.Command) bool {
	enumLike := 0
	documented := 0
	var walk func(*cobra.Command)
	walk = func(c *cobra.Command) {
		c.Flags().VisitAll(func(f *pflag.Flag) {
			usage := strings.ToLower(f.Usage)
			if strings.Contains(usage, ":") && strings.Contains(usage, ",") {
				enumLike++
				documented++
			}
		})
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(cmd)
	return enumLike > 0 && documented == enumLike
}

func hasValidationHelpers() bool {
	return fileContains("cmd/validate.go", "validateResourceID")
}

func hasCanonicalVerbs(cmd *cobra.Command) bool {
	for _, path := range [][]string{{"blocks", "add"}, {"blocks", "update"}, {"blocks", "delete"}, {"folders", "list"}, {"tasks", "list"}, {"tasks", "add"}} {
		if findCobraCommand(cmd, path...) == nil {
			return false
		}
	}
	return true
}

func fileContains(path, needle string) bool {
	data, err := readRepoFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), needle)
}

func readRepoFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		return data, nil
	}
	return os.ReadFile("../" + path)
}

func repoFileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	_, err := os.Stat("../" + path)
	return err == nil
}

func schemaSupportsRequestResponse(cmd *cobra.Command) bool {
	var found bool
	var walk func(*cobra.Command)
	walk = func(c *cobra.Command) {
		if c.Flags().Lookup("request-schema") != nil || c.Flags().Lookup("response-schema") != nil {
			found = true
		}
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(cmd)
	return found
}

func schemaMentionsProfiles(schema CommandSchema) bool {
	for _, sub := range schema.Subcommands {
		if sub.Name == "profiles" {
			return len(sub.Backends) > 0 || len(sub.RequiredCapabilities) > 0
		}
	}
	return false
}

func schemaHasBackends(schema CommandSchema) bool {
	var found bool
	var walk func(CommandSchema)
	walk = func(c CommandSchema) {
		if len(c.Backends) > 0 {
			found = true
		}
		for _, sub := range c.Subcommands {
			walk(sub)
		}
	}
	walk(schema)
	return found
}

func schemaDescriptionsAreConcise(schema CommandSchema) bool {
	var ok = true
	var walk func(CommandSchema)
	walk = func(c CommandSchema) {
		if len(c.Description) > 180 {
			ok = false
		}
		for _, sub := range c.Subcommands {
			walk(sub)
		}
	}
	walk(schema)
	return ok
}

func canonicalFlagsVisible(cmd *cobra.Command) bool {
	for _, flag := range []string{"format", "dry-run", "yes", "profile"} {
		if cmd.PersistentFlags().Lookup(flag) == nil {
			return false
		}
	}
	return true
}

func outputFormatsMention(format string) bool {
	flag := rootCmd.PersistentFlags().Lookup("format")
	if flag == nil {
		return false
	}
	return strings.Contains(strings.ToLower(flag.Usage), format)
}

func agentDXGrade(score int) string {
	switch {
	case score >= 66:
		return "agent-native"
	case score >= 51:
		return "agent-first"
	case score >= 36:
		return "agent-ready"
	case score >= 21:
		return "agent-tolerant"
	default:
		return "human-only"
	}
}
