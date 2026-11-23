package cmd

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"

	"github.com/spf13/cobra"
)

var linkDatePattern = regexp.MustCompile(`^(?:\d{4}-\d{2}-\d{2}|today|yesterday|tomorrow)$`)

var linksCmd = &cobra.Command{
	Use:   "links",
	Short: "List links between documents (experimental)",
	Long: `List directed internal links between documents (GET /internal-links, experimental).

Default level is document: one row per source/target pair with referenceCount.
Use --level block for individual references. Combine --from and --to to scope
both ends; --to alone returns backlinks with source titles, --from alone returns
outgoing links with target titles. Links touching Recently Deleted or
inaccessible documents are omitted by the server.`,
	Example: "  craft links --to DOCUMENT_ID --limit 20\n  craft links --from-date today --level block\n  craft links --from DOC_A --to DOC_B --level block",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		q, err := buildLinksQuery(cmd)
		if err != nil {
			return err
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}
		result, err := client.RequestJSON("GET", "/internal-links?"+q.Encode(), nil)
		if err != nil {
			return err
		}
		return outputRawJSON(result)
	},
}

func buildLinksQuery(cmd *cobra.Command) (url.Values, error) {
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")
	fromDate, _ := cmd.Flags().GetString("from-date")
	toDate, _ := cmd.Flags().GetString("to-date")
	level, _ := cmd.Flags().GetString("level")
	self, _ := cmd.Flags().GetBool("include-self")
	limit, _ := cmd.Flags().GetInt("limit")
	cursor, _ := cmd.Flags().GetString("cursor")

	if from != "" && fromDate != "" {
		return nil, fmt.Errorf("choose --from or --from-date")
	}
	if to != "" && toDate != "" {
		return nil, fmt.Errorf("choose --to or --to-date")
	}
	if from != "" {
		if err := validateResourceID(from, "--from"); err != nil {
			return nil, err
		}
	}
	if to != "" {
		if err := validateResourceID(to, "--to"); err != nil {
			return nil, err
		}
	}
	if fromDate != "" && !linkDatePattern.MatchString(fromDate) {
		return nil, fmt.Errorf("--from-date must be YYYY-MM-DD, today, yesterday or tomorrow")
	}
	if toDate != "" && !linkDatePattern.MatchString(toDate) {
		return nil, fmt.Errorf("--to-date must be YYYY-MM-DD, today, yesterday or tomorrow")
	}
	if level != "document" && level != "block" {
		return nil, fmt.Errorf("invalid --level %q; valid: document, block", level)
	}
	if limit < 1 || limit > 200 {
		return nil, fmt.Errorf("--limit must be between 1 and 200")
	}

	q := url.Values{}
	q.Set("level", level)
	q.Set("limit", strconv.Itoa(limit))
	for name, value := range map[string]string{"fromDocumentId": from, "toDocumentId": to, "fromDate": fromDate, "toDate": toDate, "cursor": cursor} {
		if value != "" {
			q.Set(name, value)
		}
	}
	if self {
		q.Set("includeSelfLinks", "true")
	}
	return q, nil
}

func init() {
	rootCmd.AddCommand(linksCmd)
	f := linksCmd.Flags()
	f.String("from", "", "Only links from this document root block ID")
	f.String("to", "", "Only links to this document root block ID")
	f.String("from-date", "", "Source daily note: YYYY-MM-DD, today, yesterday or tomorrow")
	f.String("to-date", "", "Target daily note: YYYY-MM-DD, today, yesterday or tomorrow")
	f.String("level", "document", "document: distinct pairs with referenceCount; block: individual references")
	f.Bool("include-self", false, "Include links within one document")
	f.Int("limit", 50, "Server page size (1-200)")
	f.String("cursor", "", "Server pagination cursor; keep level and filters unchanged")
	commandEffects["links"] = CommandEffect{"read_only", "remote", false}
}
