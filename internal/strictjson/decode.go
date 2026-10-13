// Package strictjson provides bounded, one-value JSON decoding with duplicate key rejection.
// It is used at local API and policy file trust boundaries; it is never a shell executor.
package strictjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// Decode rejects duplicate keys at every nested object depth and unknown struct fields.
// maxBytes is a hard input limit, including trailing data.
func Decode(r io.Reader, maxBytes int64, dst any) error {
	if maxBytes < 1 || maxBytes > 16<<20 {
		return errors.New("invalid JSON byte budget")
	}
	b, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return fmt.Errorf("read JSON: %w", err)
	}
	if int64(len(b)) > maxBytes {
		return errors.New("JSON input too large")
	}
	if err := uniqueObjectKeys(b); err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(dst); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("multiple JSON values are forbidden")
		}
		return fmt.Errorf("trailing JSON: %w", err)
	}
	return nil
}

// uniqueObjectKeys traverses objects and arrays via the JSON token parser.
// The bounded reader above limits memory; depth is capped to guard recursion.
func uniqueObjectKeys(b []byte) error {
	d := json.NewDecoder(bytes.NewReader(b))
	if err := walk(d, 0); err != nil {
		return err
	}
	var tok any
	if err := d.Decode(&tok); err != io.EOF {
		if err == nil {
			return errors.New("multiple JSON values are forbidden")
		}
		return fmt.Errorf("trailing JSON: %w", err)
	}
	return nil
}
func walk(d *json.Decoder, depth int) error {
	if depth > 64 {
		return errors.New("JSON nesting exceeds 64 levels")
	}
	tok, err := d.Token()
	if err != nil {
		return fmt.Errorf("JSON token: %w", err)
	}
	delimiter, ok := tok.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := make(map[string]struct{})
		for d.More() {
			raw, err := d.Token()
			if err != nil {
				return err
			}
			k, ok := raw.(string)
			if !ok {
				return errors.New("object key not a string")
			}
			if _, dup := seen[k]; dup {
				return fmt.Errorf("duplicate JSON object key %q", k)
			}
			seen[k] = struct{}{}
			if err := walk(d, depth+1); err != nil {
				return err
			}
		}
		end, err := d.Token()
		if err != nil || end != json.Delim('}') {
			return errors.New("unterminated JSON object")
		}
	case '[':
		for d.More() {
			if err := walk(d, depth+1); err != nil {
				return err
			}
		}
		end, err := d.Token()
		if err != nil || end != json.Delim(']') {
			return errors.New("unterminated JSON array")
		}
	default:
		return errors.New("unexpected JSON delimiter")
	}
	return nil
}
