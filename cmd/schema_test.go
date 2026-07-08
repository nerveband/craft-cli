package cmd

import "testing"

func TestBuildSchemaIncludesNestedSchemaCommands(t *testing.T) {
	rootSchema := buildSchema(rootCmd)

	var collections *CommandSchema
	for i := range rootSchema.Subcommands {
		if rootSchema.Subcommands[i].Name == "collections" {
			collections = &rootSchema.Subcommands[i]
			break
		}
	}
	if collections == nil {
		t.Fatal("expected collections command in root schema")
	}

	for _, sub := range collections.Subcommands {
		if sub.Name == "schema" {
			return
		}
	}

	t.Fatalf("expected nested collections schema command, got subcommands %#v", collections.Subcommands)
}

func TestBuildSchemaIncludesBackendAndCapabilities(t *testing.T) {
	rootSchema := buildSchema(rootCmd)

	mcp := findCommandSchema(rootSchema.Subcommands, "mcp")
	if mcp == nil {
		t.Fatal("expected mcp command in schema")
	}
	if !containsString(mcp.Backends, "mcp") {
		t.Fatalf("expected mcp backend metadata, got %#v", mcp.Backends)
	}
	if !containsString(mcp.RequiredCapabilities, "mcp") {
		t.Fatalf("expected mcp required capability, got %#v", mcp.RequiredCapabilities)
	}

	blocks := findCommandSchema(rootSchema.Subcommands, "blocks")
	if blocks == nil {
		t.Fatal("expected blocks command in schema")
	}
	add := findCommandSchema(blocks.Subcommands, "add")
	if add == nil {
		t.Fatal("expected blocks add command in schema")
	}
	if !containsString(add.Backends, "rest") {
		t.Fatalf("expected rest backend metadata, got %#v", add.Backends)
	}
	if !containsString(add.Backends, "mcp") {
		t.Fatalf("expected mcp backend metadata, got %#v", add.Backends)
	}
	if !containsString(add.RequiredCapabilities, "write") {
		t.Fatalf("expected write capability, got %#v", add.RequiredCapabilities)
	}
	if !containsString(add.OptionalCapabilities, "blocks.style") {
		t.Fatalf("expected blocks.style optional capability, got %#v", add.OptionalCapabilities)
	}

	resolveLink := findCommandSchema(findCommandSchema(rootSchema.Subcommands, "documents").Subcommands, "resolve-link")
	if resolveLink == nil {
		t.Fatal("expected documents resolve-link command in schema")
	}
	if !containsString(resolveLink.Backends, "mcp") {
		t.Fatalf("expected mcp backend for resolve-link, got %#v", resolveLink.Backends)
	}

	exploreThemes := findCommandSchema(blocks.Subcommands, "explore-themes")
	if exploreThemes == nil {
		t.Fatal("expected blocks explore-themes command in schema")
	}
	if !containsString(exploreThemes.RequiredCapabilities, "mcp") {
		t.Fatalf("expected mcp required capability for explore-themes, got %#v", exploreThemes.RequiredCapabilities)
	}
	revert := findCommandSchema(blocks.Subcommands, "revert")
	if revert == nil {
		t.Fatal("expected blocks revert command in schema")
	}
	if !containsString(revert.Backends, "mcp") {
		t.Fatalf("expected mcp backend for blocks revert, got %#v", revert.Backends)
	}

	audit := findCommandSchema(rootSchema.Subcommands, "audit")
	if audit == nil {
		t.Fatal("expected audit command in schema")
	}
	if !containsString(audit.Backends, "local") {
		t.Fatalf("expected local backend for audit, got %#v", audit.Backends)
	}
	if !containsString(audit.RequiredCapabilities, "audit") {
		t.Fatalf("expected audit required capability, got %#v", audit.RequiredCapabilities)
	}

	batch := findCommandSchema(mcp.Subcommands, "batch")
	if batch == nil {
		t.Fatal("expected mcp batch command in schema")
	}
	if !containsString(batch.Backends, "mcp") {
		t.Fatalf("expected mcp backend for mcp batch, got %#v", batch.Backends)
	}

	list := findCommandSchema(rootSchema.Subcommands, "list")
	if list == nil {
		t.Fatal("expected list command in schema")
	}
	if !containsString(list.Backends, "mcp") || !containsString(list.OptionalCapabilities, "cursor") {
		t.Fatalf("expected list to expose mcp cursor support, backends=%#v optional=%#v", list.Backends, list.OptionalCapabilities)
	}

	collections := findCommandSchema(rootSchema.Subcommands, "collections")
	if collections == nil {
		t.Fatal("expected collections command in schema")
	}
	views := findCommandSchema(collections.Subcommands, "views")
	if views == nil {
		t.Fatal("expected collections views command in schema")
	}
	if !containsString(views.Backends, "mcp") || !containsString(views.RequiredCapabilities, "collections.views") {
		t.Fatalf("expected collections views to be mcp-backed, backends=%#v required=%#v", views.Backends, views.RequiredCapabilities)
	}
}

func TestBuildSchemaIncludesContextCountFlags(t *testing.T) {
	rootSchema := buildSchema(rootCmd)

	list := findCommandSchema(rootSchema.Subcommands, "list")
	if list == nil {
		t.Fatal("expected list command in schema")
	}
	if !hasFlagSchema(list.Flags, "--count") {
		t.Fatalf("expected list --count flag, got %#v", list.Flags)
	}

	search := findCommandSchema(rootSchema.Subcommands, "search")
	if search == nil {
		t.Fatal("expected search command in schema")
	}
	if !hasFlagSchema(search.Flags, "--count") {
		t.Fatalf("expected search --count flag, got %#v", search.Flags)
	}
}

func TestBuildSchemaIncludesRawPayloadFlags(t *testing.T) {
	rootSchema := buildSchema(rootCmd)

	create := findCommandSchema(rootSchema.Subcommands, "create")
	if create == nil {
		t.Fatal("expected create command in schema")
	}
	if !hasFlagSchema(create.Flags, "--json") || !hasFlagSchema(create.Flags, "--stdin") {
		t.Fatalf("expected create raw payload flags, got %#v", create.Flags)
	}

	update := findCommandSchema(rootSchema.Subcommands, "update")
	if update == nil {
		t.Fatal("expected update command in schema")
	}
	if !hasFlagSchema(update.Flags, "--json") || !hasFlagSchema(update.Flags, "--stdin") {
		t.Fatalf("expected update raw payload flags, got %#v", update.Flags)
	}

	tasks := findCommandSchema(rootSchema.Subcommands, "tasks")
	if tasks == nil {
		t.Fatal("expected tasks command in schema")
	}
	for _, name := range []string{"add", "update", "delete"} {
		sub := findCommandSchema(tasks.Subcommands, name)
		if sub == nil {
			t.Fatalf("expected tasks %s command in schema", name)
		}
		if !hasFlagSchema(sub.Flags, "--json") || !hasFlagSchema(sub.Flags, "--stdin") {
			t.Fatalf("expected tasks %s raw payload flags, got %#v", name, sub.Flags)
		}
	}

	for _, name := range []string{"delete", "clear", "move"} {
		cmd := findCommandSchema(rootSchema.Subcommands, name)
		if cmd == nil {
			t.Fatalf("expected %s command in schema", name)
		}
		if !hasFlagSchema(cmd.Flags, "--json") || !hasFlagSchema(cmd.Flags, "--stdin") {
			t.Fatalf("expected %s raw payload flags, got %#v", name, cmd.Flags)
		}
	}

	folders := findCommandSchema(rootSchema.Subcommands, "folders")
	if folders == nil {
		t.Fatal("expected folders command in schema")
	}
	for _, name := range []string{"create", "move", "delete"} {
		sub := findCommandSchema(folders.Subcommands, name)
		if sub == nil {
			t.Fatalf("expected folders %s command in schema", name)
		}
		if !hasFlagSchema(sub.Flags, "--json") || !hasFlagSchema(sub.Flags, "--stdin") {
			t.Fatalf("expected folders %s raw payload flags, got %#v", name, sub.Flags)
		}
	}

	collections := findCommandSchema(rootSchema.Subcommands, "collections")
	if collections == nil {
		t.Fatal("expected collections command in schema")
	}
	for _, name := range []string{"add", "update", "delete"} {
		sub := findCommandSchema(collections.Subcommands, name)
		if sub == nil {
			t.Fatalf("expected collections %s command in schema", name)
		}
		if !hasFlagSchema(sub.Flags, "--json") || !hasFlagSchema(sub.Flags, "--stdin") {
			t.Fatalf("expected collections %s raw payload flags, got %#v", name, sub.Flags)
		}
	}

	comments := findCommandSchema(rootSchema.Subcommands, "comments")
	if comments == nil {
		t.Fatal("expected comments command in schema")
	}
	addComment := findCommandSchema(comments.Subcommands, "add")
	if addComment == nil {
		t.Fatal("expected comments add command in schema")
	}
	if !hasFlagSchema(addComment.Flags, "--json") || !hasFlagSchema(addComment.Flags, "--stdin") {
		t.Fatalf("expected comments add raw payload flags, got %#v", addComment.Flags)
	}

	upload := findCommandSchema(rootSchema.Subcommands, "upload")
	if upload == nil {
		t.Fatal("expected upload command in schema")
	}
	if !hasFlagSchema(upload.Flags, "--json") || !hasFlagSchema(upload.Flags, "--stdin") {
		t.Fatalf("expected upload raw payload flags, got %#v", upload.Flags)
	}

	whiteboards := findCommandSchema(rootSchema.Subcommands, "whiteboards")
	if whiteboards == nil {
		t.Fatal("expected whiteboards command in schema")
	}
	for _, name := range []string{"create", "add", "update", "delete"} {
		sub := findCommandSchema(whiteboards.Subcommands, name)
		if sub == nil {
			t.Fatalf("expected whiteboards %s command in schema", name)
		}
		if !hasFlagSchema(sub.Flags, "--json") || !hasFlagSchema(sub.Flags, "--stdin") {
			t.Fatalf("expected whiteboards %s raw payload flags, got %#v", name, sub.Flags)
		}
	}
}

func findCommandSchema(commands []CommandSchema, name string) *CommandSchema {
	for i := range commands {
		if commands[i].Name == name {
			return &commands[i]
		}
	}
	return nil
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func hasFlagSchema(flags []FlagSchema, name string) bool {
	for _, flag := range flags {
		if flag.Name == name {
			return true
		}
	}
	return false
}
