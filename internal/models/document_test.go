package models

import (
	"encoding/json"
	"testing"
)

func TestCurrentResponseVariants(t *testing.T) {
	for _, input := range []string{`{"id":"p","type":"page","title":{"value":"Page","color":"#fff"},"cover":{"url":"https://example.org/a"},"content":[]}`, `{"id":"p","type":"url","title":"Link","url":"https://example.org"}`} {
		var b Block
		if err := json.Unmarshal([]byte(input), &b); err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(b)
		if err != nil {
			t.Fatal(err)
		}
		var before, after map[string]interface{}
		json.Unmarshal([]byte(input), &before)
		json.Unmarshal(data, &after)
		for key, v := range before {
			a, _ := json.Marshal(v)
			z, _ := json.Marshal(after[key])
			if string(a) != string(z) {
				t.Errorf("lost %s: %s != %s", key, a, z)
			}
		}
	}
	var tasks TaskList
	if err := json.Unmarshal([]byte(`{"items":[{"id":"t","markdown":"Task","taskInfo":{"state":"done","scheduleDate":"2026-09-26"},"repeat":{"type":"fixed","frequency":"daily"},"location":{"type":"inbox"}}]}`), &tasks); err != nil {
		t.Fatal(err)
	}
	if tasks.Total != 1 || tasks.Items[0].State != "done" || tasks.Items[0].Location["type"] != "inbox" {
		t.Fatalf("nested task lost: %+v", tasks)
	}
}

func TestRepeatConfigCurrentShape(t *testing.T) {
	input := `{"type":"fixed","frequency":"weekly","interval":2,"weekly":{"days":["monday","friday"]},"reminder":{"enabled":true,"dateOffset":540}}`
	var r RepeatConfig
	if err := json.Unmarshal([]byte(input), &r); err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var actual map[string]interface{}
	json.Unmarshal(b, &actual)
	if actual["type"] != "fixed" || actual["weekly"] == nil || actual["reminder"] == nil {
		t.Fatalf("invalid repeat: %s", b)
	}
}
