package locale

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

var (
	translations map[string]string
	once         sync.Once
)

// Init loads the locale file (default: locales/id_ID.json)
func Init(filePath string) error {
	if filePath == "" {
		filePath = "locales/id_ID.json"
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read locale file %s: %w", filePath, err)
	}

	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("failed to parse locale JSON: %w", err)
	}

	translations = m
	return nil
}

// T translates a key and substitutes positional placeholders ({0}, {1}, etc.)
// Also supports {name} fallback if the first argument is a string.
func T(key string, args ...any) string {
	val, ok := translations[key]
	if !ok {
		val = key
	}

	if len(args) == 0 {
		return val
	}

	// Flatten if first arg is a slice of any or slice of string
	var flatArgs []any
	if len(args) == 1 {
		switch v := args[0].(type) {
		case []any:
			flatArgs = v
		case []string:
			for _, s := range v {
				flatArgs = append(flatArgs, s)
			}
		default:
			flatArgs = args
		}
	} else {
		flatArgs = args
	}

	// Positional {0}, {1}, ...
	for i, arg := range flatArgs {
		placeholder := fmt.Sprintf("{%d}", i)
		val = strings.ReplaceAll(val, placeholder, fmt.Sprintf("%v", arg))
	}

	// Key-value pairs if even count: "name", "Budi", "place", "Bangka" -> {name}=Budi, {place}=Bangka
	if len(flatArgs) >= 2 && len(flatArgs)%2 == 0 {
		for i := 0; i < len(flatArgs); i += 2 {
			k := fmt.Sprintf("%v", flatArgs[i])
			v := fmt.Sprintf("%v", flatArgs[i+1])
			val = strings.ReplaceAll(val, fmt.Sprintf("{%s}", k), v)
			val = strings.ReplaceAll(val, fmt.Sprintf("{{%s}}", k), v)
		}
	}

	// Direct {name} fallback if first arg is string
	if len(flatArgs) > 0 {
		val = strings.ReplaceAll(val, "{name}", fmt.Sprintf("%v", flatArgs[0]))
	}

	// Sprintf %d, %s format fallback if string contains % and has matching verb
	if strings.Contains(val, "%") && !strings.Contains(val, "{") {
		formatted := fmt.Sprintf(val, flatArgs...)
		return formatted
	}

	return val
}
