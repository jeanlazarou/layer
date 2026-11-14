package interpreter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
)

// Plain data is what can cross the boundary between a Layer program and its
// runtime — an app's model and its messages: nil, booleans, integers,
// floats, strings, and lists and records of plain data. Functions and UI
// components are not plain data.

// CheckData reports whether v is plain data. root names the value in the
// error ("model", "message", ...), which then points at the offending part,
// e.g. "model.items[2].on_done is a function".
func CheckData(v interface{}, root string) error {
	var b strings.Builder
	return encodeData(&b, v, root)
}

// EncodeData renders plain data as JSON. Floats always carry a decimal point
// or an exponent ("3.0"), so DecodeData restores integers and floats exactly.
func EncodeData(v interface{}) (string, error) {
	var b strings.Builder
	if err := encodeData(&b, v, "value"); err != nil {
		return "", err
	}
	return b.String(), nil
}

func encodeData(b *strings.Builder, v interface{}, path string) error {
	switch x := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		b.WriteString(strconv.FormatBool(x))
	case int64:
		b.WriteString(strconv.FormatInt(x, 10))
	case int:
		b.WriteString(strconv.Itoa(x))
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return fmt.Errorf("%s is %v, which plain data cannot represent", path, x)
		}
		s := strconv.FormatFloat(x, 'g', -1, 64)
		if !strings.ContainsAny(s, ".eE") {
			s += ".0"
		}
		b.WriteString(s)
	case string:
		b.WriteString(quoteJSON(x))
	case []interface{}:
		b.WriteByte('[')
		for idx, elem := range x {
			if idx > 0 {
				b.WriteByte(',')
			}
			if err := encodeData(b, elem, fmt.Sprintf("%s[%d]", path, idx)); err != nil {
				return err
			}
		}
		b.WriteByte(']')
	case map[string]interface{}:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		b.WriteByte('{')
		for idx, k := range keys {
			if idx > 0 {
				b.WriteByte(',')
			}
			b.WriteString(quoteJSON(k))
			b.WriteByte(':')
			if err := encodeData(b, x[k], path+"."+k); err != nil {
				return err
			}
		}
		b.WriteByte('}')
	default:
		return fmt.Errorf("%s is a %s, but only plain data (records, lists, strings, numbers, booleans) is allowed", path, typeName(v))
	}
	return nil
}

// quoteJSON quotes s as a JSON string without json.Marshal's HTML escaping
// (which would turn "<" into "\u003c" in messages and saved models).
func quoteJSON(s string) string {
	var b strings.Builder
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s)
	return strings.TrimSuffix(b.String(), "\n")
}

// DecodeData parses JSON produced by EncodeData (or by the editor) back into
// Layer values. Numbers with a decimal point or exponent become floats;
// others become integers.
func DecodeData(s string) (interface{}, error) {
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	var raw interface{}
	if err := dec.Decode(&raw); err != nil {
		return nil, fmt.Errorf("invalid data: %v", err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("invalid data: unexpected content after the value")
	}
	return fromJSON(raw)
}

func fromJSON(v interface{}) (interface{}, error) {
	switch x := v.(type) {
	case json.Number:
		if strings.ContainsAny(string(x), ".eE") {
			return x.Float64()
		}
		if n, err := x.Int64(); err == nil {
			return n, nil
		}
		return x.Float64()
	case []interface{}:
		out := make([]interface{}, len(x))
		for idx, elem := range x {
			conv, err := fromJSON(elem)
			if err != nil {
				return nil, err
			}
			out[idx] = conv
		}
		return out, nil
	case map[string]interface{}:
		out := make(map[string]interface{}, len(x))
		for k, elem := range x {
			conv, err := fromJSON(elem)
			if err != nil {
				return nil, err
			}
			out[k] = conv
		}
		return out, nil
	default: // nil, bool, string
		return x, nil
	}
}
