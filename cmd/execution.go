package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var writeResponses []json.RawMessage
var lastRESTResponse []byte

// executeCaptured preserves every upstream write response, including earlier
// successes when a later request fails. Human output remains human output.
func executeCaptured() error {
	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		return err
	}
	os.Stdout = writer
	finished := make(chan []byte, 1)
	go func() { data, _ := io.ReadAll(reader); reader.Close(); finished <- data }()
	err = rootCmd.Execute()
	writer.Close()
	os.Stdout = original
	output := <-finished
	if errors.Is(err, errPreviewComplete) {
		err = nil
	}
	if len(writeResponses) > 0 && getOutputFormat() == FormatJSON && getOutputOnly() == "" {
		var result interface{}
		if len(writeResponses) == 1 {
			result = writeResponses[0]
		} else {
			result = map[string]interface{}{"operations": writeResponses}
		}
		r, w, pipeErr := os.Pipe()
		if pipeErr != nil {
			return pipeErr
		}
		os.Stdout = w
		done := make(chan []byte, 1)
		go func() { b, _ := io.ReadAll(r); r.Close(); done <- b }()
		renderErr := outputJSON(result)
		w.Close()
		os.Stdout = original
		output = <-done
		if renderErr != nil {
			return renderErr
		}

	}
	if activeCommand != nil && err == nil && isJSONFormat(getOutputFormat()) && getOutputOnly() == "" && !rawOutput && !json.Valid(bytes.TrimSpace(output)) {
		path := canonicalCommandPath(activeCommand)
		switch path {
		case "local append", "local new", "local open", "local search", "local space", "local today", "local tomorrow", "local yesterday", "setup", "config add", "config add-rest", "config add-mcp", "config use", "config remove", "config reset", "profiles add-rest", "profiles add-mcp", "profiles use", "profiles remove":
			output, _ = json.Marshal(map[string]interface{}{"status": "ok", "message": strings.TrimSpace(string(output)), "_provenance": map[string]interface{}{"scope": "local", "source": "local", "backend": "local"}})
			output = append(output, '\n')
		}
	}
	if outputFormat == "raw" && len(lastRESTResponse) > 0 && transformExpr == "" && getOutputOnly() == "" {
		output = lastRESTResponse
	}
	if len(output) > 0 {
		if !isJSONFormat(getOutputFormat()) && outputFormat != "raw" {
			output = []byte(safeTerminalText(string(output)))
		}
		if stringsHasFileDelivery() && !isDryRun() && err == nil {
			return writeAtomicOutput(deliverTarget[5:], output)
		}
		if _, writeErr := io.Copy(original, bytes.NewReader(output)); writeErr != nil {
			return fmt.Errorf("write output: %w", writeErr)
		}
	}
	return err
}
func stringsHasFileDelivery() bool { return len(deliverTarget) > 5 && deliverTarget[:5] == "file:" }

var ansiSequence = regexp.MustCompile("\x1b\\[[0-?]*[ -/]*[@-~]|\x1b\\][^\x07]*(?:\x07|\x1b\\\\)")

func safeTerminalText(text string) string {
	text = ansiSequence.ReplaceAllString(text, "")
	return strings.Map(func(r rune) rune {
		if (r < 32 && r != '\n' && r != '\t') || (r >= 127 && r <= 159) {
			return -1
		}
		return r
	}, text)
}
