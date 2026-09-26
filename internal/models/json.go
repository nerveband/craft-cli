package models

import "encoding/json"

func mergeJSON(raw map[string]json.RawMessage, value interface{}) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	fields := map[string]json.RawMessage{}
	for k, v := range raw {
		fields[k] = v
	}
	var typed map[string]json.RawMessage
	if err = json.Unmarshal(data, &typed); err != nil {
		return nil, err
	}
	for k, v := range typed {
		fields[k] = v
	}
	return json.Marshal(fields)
}

func (b *Block) UnmarshalJSON(data []byte) error {
	type plain Block
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	title := fields["title"]
	delete(fields, "title")
	rest, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	var value plain
	if err = json.Unmarshal(rest, &value); err != nil {
		return err
	}
	*b = Block(value)
	b.Extra = fields
	b.TitleValue = title
	if len(title) > 0 {
		if err := json.Unmarshal(title, &b.Title); err != nil {
			var rich struct {
				Value string `json:"value"`
			}
			if err = json.Unmarshal(title, &rich); err != nil {
				return err
			}
			b.Title = rich.Value
		}
	}
	return nil
}
func (b Block) MarshalJSON() ([]byte, error) {
	type plain Block
	data, err := mergeJSON(b.Extra, plain(b))
	if err != nil {
		return nil, err
	}
	if len(b.TitleValue) == 0 {
		return data, nil
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	fields["title"] = b.TitleValue
	return json.Marshal(fields)
}
func (b *BlocksResponse) UnmarshalJSON(data []byte) error {
	type plain BlocksResponse
	var value plain
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*b = BlocksResponse(value)
	return json.Unmarshal(data, &b.Extra)
}
func (b BlocksResponse) MarshalJSON() ([]byte, error) {
	type plain BlocksResponse
	return mergeJSON(b.Extra, plain(b))
}

func (t *Task) UnmarshalJSON(data []byte) error {
	type plain Task
	var value struct {
		plain
		TaskInfo *TaskInfo `json:"taskInfo"`
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*t = Task(value.plain)
	if value.TaskInfo != nil {
		t.State = value.TaskInfo.State
		t.ScheduleDate = value.TaskInfo.ScheduleDate
		t.DeadlineDate = value.TaskInfo.DeadlineDate
	}
	return nil
}
func (d Document) MarshalJSON() ([]byte, error) {
	type plain Document
	data, err := json.Marshal(plain(d))
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	if d.CreatedAt.IsZero() {
		delete(fields, "createdAt")
	}
	if d.LastModifiedAt.IsZero() {
		delete(fields, "lastModifiedAt")
	}
	if d.SpaceID == "" {
		delete(fields, "spaceId")
	}
	if !d.HasChildren {
		delete(fields, "hasChildren")
	}
	return json.Marshal(fields)
}
func (s *CollectionSchema) UnmarshalJSON(data []byte) error {
	type plain CollectionSchema
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	// JSON Schema uses a property map; Craft's editable schema uses an array.
	if p := fields["properties"]; len(p) > 0 && p[0] == '{' {
		delete(fields, "properties")
	}
	partial, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	var value plain
	if err = json.Unmarshal(partial, &value); err != nil {
		return err
	}
	*s = CollectionSchema(value)
	s.Raw = append([]byte(nil), data...)
	return nil
}
func (s CollectionSchema) MarshalJSON() ([]byte, error) {
	if len(s.Raw) > 0 {
		return s.Raw, nil
	}
	type plain CollectionSchema
	return json.Marshal(plain(s))
}

func (v *DocumentList) UnmarshalJSON(data []byte) error {
	type plain DocumentList
	var value plain
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*v = DocumentList(value)
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if _, ok := fields["total"]; !ok {
		v.Total = len(v.Items)
	}
	return nil
}

func (v *FolderList) UnmarshalJSON(data []byte) error {
	type plain FolderList
	var value plain
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*v = FolderList(value)
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if _, ok := fields["total"]; !ok {
		v.Total = len(v.Items)
	}
	return nil
}

func (v *TaskList) UnmarshalJSON(data []byte) error {
	type plain TaskList
	var value plain
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*v = TaskList(value)
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if _, ok := fields["total"]; !ok {
		v.Total = len(v.Items)
	}
	return nil
}

func (v *SearchResult) UnmarshalJSON(data []byte) error {
	type plain SearchResult
	var value plain
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*v = SearchResult(value)
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if _, ok := fields["total"]; !ok {
		v.Total = len(v.Items)
		v.Metadata = map[string]interface{}{"count_scope": "returned_results", "server_limit": 20, "possibly_truncated": len(v.Items) >= 20}
	}
	return nil
}
