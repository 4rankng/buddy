package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ToMySQLJSONObjectExpr converts a JSON *object* string into a MySQL expression like:
//
//	{"key":"value"} -> JSON_OBJECT('key','value')
//
// It also handles nested objects/arrays via JSON_OBJECT / JSON_ARRAY.
// Returns an error if the top-level JSON is not an object.
func ToMySQLJSONObjectExpr(jsonStr string) (string, error) {
	dec := json.NewDecoder(strings.NewReader(jsonStr))
	dec.UseNumber()

	var v any
	if err := dec.Decode(&v); err != nil {
		return "", fmt.Errorf("invalid json: %w", err)
	}
	if _, err := dec.Token(); err == nil {
		// If we can still read tokens, there's trailing data.
		return "", fmt.Errorf("invalid json: trailing content")
	}

	obj, ok := v.(map[string]any)
	if !ok {
		return "", fmt.Errorf("top-level JSON must be an object")
	}

	return buildValue(obj)
}

func buildValue(v any) (string, error) {
	switch x := v.(type) {
	case map[string]any:
		// Stable output (nice for tests/logging)
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var parts []string
		for _, k := range keys {
			valExpr, err := buildValue(x[k])
			if err != nil {
				return "", err
			}
			parts = append(parts, mysqlStringLiteral(k), valExpr)
		}
		return "JSON_OBJECT(" + strings.Join(parts, ", ") + ")", nil

	case []any:
		var parts []string
		for _, it := range x {
			valExpr, err := buildValue(it)
			if err != nil {
				return "", err
			}
			parts = append(parts, valExpr)
		}
		return "JSON_ARRAY(" + strings.Join(parts, ", ") + ")", nil

	case string:
		return mysqlStringLiteral(x), nil

	case json.Number:
		// Could be int/float; MySQL will interpret it numerically.
		return x.String(), nil

	case float64:
		return strconv.FormatFloat(x, 'g', -1, 64), nil

	case bool:
		if x {
			return "TRUE", nil
		}
		return "FALSE", nil

	case nil:
		return "NULL", nil

	default:
		// Fallback: serialize and CAST as JSON
		b, err := json.Marshal(x)
		if err != nil {
			return "", fmt.Errorf("marshal fallback: %w", err)
		}
		return "CAST(" + mysqlStringLiteral(string(b)) + " AS JSON)", nil
	}
}

func mysqlStringLiteral(s string) string {
	// MySQL single-quoted literal with basic escaping.
	// Escapes: \, ', NUL, newline, carriage return, tab, backspace, formfeed.
	var b bytes.Buffer
	b.WriteByte('\'')
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case 0:
			b.WriteString(`\0`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\\':
			b.WriteString(`\\`)
		case '\'':
			b.WriteString(`\'`)
		default:
			b.WriteByte(s[i])
		}
	}
	b.WriteByte('\'')
	return b.String()
}

// GetRollbackStreamMessage extracts and formats the StreamMessage for rollback
func GetRollbackStreamMessage(data string) string {
	// Default success object
	streamMessageExpr := "JSON_OBJECT('TxID','', 'Status','SUCCESS', 'ErrorCode','', 'ExternalID','', 'ReferenceID','', 'ErrorMessage','', 'ValueTimestamp','')"

	if data != "" {
		var dataMap map[string]interface{}
		if err := json.Unmarshal([]byte(data), &dataMap); err == nil {
			if sm, ok := dataMap["StreamMessage"]; ok {
				if smBytes, err := json.Marshal(sm); err == nil {
					smJSON := string(smBytes)
					if expr, err := ToMySQLJSONObjectExpr(smJSON); err == nil {
						streamMessageExpr = expr
						// Update status to SUCCESS and clear errors
						streamMessageExpr = strings.ReplaceAll(streamMessageExpr,
							"'Status','FAILED'", "'Status','SUCCESS'")
						streamMessageExpr = strings.ReplaceAll(streamMessageExpr,
							"'ErrorCode','ADAPTER_ERROR'", "'ErrorCode',''")
						streamMessageExpr = strings.ReplaceAll(streamMessageExpr,
							"'ErrorMessage','Manual Rejected'", "'ErrorMessage',''")
					}
				}
			}
		}
	}

	return streamMessageExpr
}
