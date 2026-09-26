package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// renderData is the common structured-output pipeline. Selection happens before
// serialization, so formats and delivery have the same payload contract.
func renderData(data interface{}) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	var value interface{}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	if transformExpr != "" {
		value, err = selectPath(value, transformExpr)
		if err != nil {
			return err
		}
	}
	var buf bytes.Buffer
	if field := getOutputOnly(); field != "" {
		values := []interface{}{value}
		if obj, ok := value.(map[string]interface{}); ok {
			if items, ok := obj["items"].([]interface{}); ok {
				values = items
			}
		}
		if items, ok := value.([]interface{}); ok {
			values = items
		}
		for _, item := range values {
			v, e := selectPath(item, field)
			if e != nil {
				return e
			}
			fmt.Fprintln(&buf, v)
		}
	} else {
		switch outputFormat {
		case "yaml":
			// Decode through ordinary numbers so YAML does not quote json.Number.
			normalized, e := json.Marshal(value)
			if e != nil {
				return e
			}
			var y interface{}
			if e = json.Unmarshal(normalized, &y); e != nil {
				return e
			}
			b, e := yaml.Marshal(y)
			if e != nil {
				return e
			}
			buf.Write(b)
		case "jsonl":
			rows := []interface{}{value}
			if obj, ok := value.(map[string]interface{}); ok {
				if items, ok := obj["items"].([]interface{}); ok {
					rows = items
				}
			} else if items, ok := value.([]interface{}); ok {
				rows = items
			}
			enc := json.NewEncoder(&buf)
			for _, row := range rows {
				if err := enc.Encode(row); err != nil {
					return err
				}
			}
		default:
			enc := json.NewEncoder(&buf)
			if outputFormat != "raw" {
				enc.SetIndent("", "  ")
			}
			if err := enc.Encode(value); err != nil {
				return err
			}
		}
	}

	_, err = os.Stdout.Write(buf.Bytes())
	return err
}

func selectPath(value interface{}, path string) (interface{}, error) {
	for _, part := range strings.Split(path, ".") {
		switch v := value.(type) {
		case map[string]interface{}:
			next, ok := v[part]
			if !ok {
				return nil, fmt.Errorf("unknown field %q in %q", part, path)
			}
			value = next
		case []interface{}:
			i, err := strconv.Atoi(part)
			if err != nil || i < 0 || i >= len(v) {
				return nil, fmt.Errorf("invalid array index %q in %q", part, path)
			}
			value = v[i]
		default:
			return nil, fmt.Errorf("cannot select %q from scalar", path)
		}
	}
	return value, nil
}

func writeAtomicOutput(path string, data []byte) error {
	if path == "" {
		return fmt.Errorf("--deliver file: requires a path")
	}
	if err := validateOutputPath(path, allowOutsideCWD); err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil && !yesFlag {
		return fmt.Errorf("output exists; use --yes to replace %s", path)
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".craft-output-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if !yesFlag { // Atomic no-clobber publication; hard link fails if destination appeared.
		return os.Link(tmp, path)
	}
	return os.Rename(tmp, path)
}
