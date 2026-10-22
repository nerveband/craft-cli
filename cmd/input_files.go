package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

var expandFiles bool

const maxEmbeddedBytes = 8 << 20

// Expansion is explicitly opted into, applies only to structured input, and
// never follows network URLs. @@ escapes one leading @ without recursion.
func decodeInputJSON(data []byte, target interface{}) error {
	if !expandFiles {
		return json.Unmarshal(data, target)
	}
	if len(data) > maxEmbeddedBytes {
		return fmt.Errorf("JSON input exceeds 8 MiB expansion budget")
	}
	var value interface{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	if err := decoder.Decode(new(interface{})); err != io.EOF {
		return fmt.Errorf("expected one JSON value")
	}
	budget := maxEmbeddedBytes - len(data)
	expanded, err := expandInputValue(value, &budget)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(expanded)
	if err != nil {
		return err
	}
	if len(encoded) > maxEmbeddedBytes {
		return fmt.Errorf("expanded JSON exceeds 8 MiB budget")
	}
	return json.Unmarshal(encoded, target)
}

func expandInputValue(value interface{}, budget *int) (interface{}, error) {
	switch v := value.(type) {
	case map[string]interface{}:
		for k, item := range v {
			expanded, err := expandInputValue(item, budget)
			if err != nil {
				return nil, fmt.Errorf("field %s: %w", k, err)
			}
			v[k] = expanded
		}
		return v, nil
	case []interface{}:
		for i, item := range v {
			expanded, err := expandInputValue(item, budget)
			if err != nil {
				return nil, err
			}
			v[i] = expanded
		}
		return v, nil
	case string:
		if strings.HasPrefix(v, "@@") {
			return v[1:], nil
		}
		mode, path := "", ""
		for _, prefix := range []string{"@file://", "@text://", "@data://"} {
			if strings.HasPrefix(v, prefix) {
				mode, path = prefix, strings.TrimPrefix(v, prefix)
				break
			}
		}
		if mode == "" {
			return v, nil
		}
		if path == "" {
			return nil, fmt.Errorf("embedded file path is empty")
		}
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open embedded file: %w", err)
		}
		defer f.Close()
		info, err := f.Stat()
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("embedded input must be a regular file")
		}
		data, err := io.ReadAll(io.LimitReader(f, int64(*budget)+1))
		if err != nil {
			return nil, err
		}
		cost := len(data)
		if mode == "@data://" {
			cost = base64.StdEncoding.EncodedLen(len(data))
		}
		*budget -= cost
		if *budget < 0 {
			return nil, fmt.Errorf("embedded input exceeds 8 MiB expansion budget")
		}
		if mode == "@text://" {
			if !utf8.Valid(data) {
				return nil, fmt.Errorf("embedded text is not UTF-8; use @data:// for binary input")
			}
			return string(data), nil
		}
		if mode == "@data://" {
			return base64.StdEncoding.EncodeToString(data), nil
		}
		var nested interface{}
		d := json.NewDecoder(bytes.NewReader(data))
		d.UseNumber()
		if err = d.Decode(&nested); err != nil {
			return nil, fmt.Errorf("embedded JSON: %w", err)
		}
		if err = d.Decode(new(interface{})); err != io.EOF {
			return nil, fmt.Errorf("embedded JSON must contain one value")
		}
		// Files are data; nested expansion must be expressed in the original payload.
		return nested, nil
	default:
		return value, nil
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&expandFiles, "expand-files", false, "Expand @file:// JSON, @text:// text and @data:// base64 in structured input; @@ escapes @ (8 MiB total)")
}

// Bound reads before allocation when expansion is enabled, including the source
// JSON itself. Plain upload bytes use a separate binary input path.
func readStructuredBytes(reader io.Reader) ([]byte, error) {
	if expandFiles {
		reader = io.LimitReader(reader, maxEmbeddedBytes+1)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if expandFiles && len(data) > maxEmbeddedBytes {
		return nil, fmt.Errorf("structured input exceeds 8 MiB expansion budget")
	}
	return data, nil
}
func readStructuredFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return readStructuredBytes(f)
}
