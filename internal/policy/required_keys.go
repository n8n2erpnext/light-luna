package policy

import (
	"encoding/json"
	"fmt"
)

// requireExactPolicyKeys enforces the public PolicyBundle object's declared
// field names and required/optional presence rules at the input boundary.
// encoding/json matches struct fields case-insensitively and silently defaults
// absent members; neither behavior is valid in the public JSON Schema.
//
// Caller must first run strictjson.DecodeRejectNull, which bounds size,
// rejects duplicate keys, malformed syntax and explicit null at every depth.
// This second bounded scan only checks exact names and presence, not semantics.
func requireExactPolicyKeys(raw []byte) error {
	root, err := exactObjectKeys(raw, "$", []string{"apiVersion", "kind", "metadata", "spec"})
	if err != nil {
		return err
	}
	if _, err = exactObjectKeys(root["metadata"], "metadata", []string{"name", "revision"}); err != nil {
		return err
	}
	spec, err := exactObjectKeys(root["spec"], "spec",
		[]string{"mode", "defaultInbound", "emergencyAccess", "autonomous"},
		"aliases", "rules")
	if err != nil {
		return err
	}
	if _, err = exactObjectKeys(spec["emergencyAccess"], "spec.emergencyAccess", []string{"enabled", "networks"}); err != nil {
		return err
	}
	if _, err = exactObjectKeys(spec["autonomous"], "spec.autonomous",
		[]string{"mode", "allowRealBans", "maxBanTtlSeconds"}); err != nil {
		return err
	}
	if aliases, ok := spec["aliases"]; ok {
		if err = exactArrayObjects(aliases, "spec.aliases",
			[]string{"id", "type", "values"}); err != nil {
			return err
		}
	}
	if rules, ok := spec["rules"]; ok {
		if err = exactArrayObjects(rules, "spec.rules",
			[]string{"id", "priority", "direction", "source", "protocol", "action"},
			"destinationPort", "note"); err != nil {
			return err
		}
	}
	return nil
}

func exactArrayObjects(raw json.RawMessage, path string, required []string, optional ...string) error {
	var values []json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil {
		return fmt.Errorf("%s: expected array of objects: %w", path, err)
	}
	for i, value := range values {
		if _, err := exactObjectKeys(value, fmt.Sprintf("%s[%d]", path, i), required, optional...); err != nil {
			return err
		}
	}
	return nil
}

func exactObjectKeys(raw json.RawMessage, path string, required []string, optional ...string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, fmt.Errorf("%s: expected object: %w", path, err)
	}
	if fields == nil {
		return nil, fmt.Errorf("%s: expected non-null object", path)
	}
	allowed := make(map[string]struct{}, len(required)+len(optional))
	for _, name := range required {
		allowed[name] = struct{}{}
		if _, ok := fields[name]; !ok {
			return nil, fmt.Errorf("%s: required JSON field %q is absent", path, name)
		}
	}
	for _, name := range optional {
		allowed[name] = struct{}{}
	}
	for name := range fields {
		if _, ok := allowed[name]; !ok {
			return nil, fmt.Errorf("%s: unknown or case-mismatched JSON field %q", path, name)
		}
	}
	return fields, nil
}
