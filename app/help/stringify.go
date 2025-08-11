package help

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

type StringifyMethod int

const (
	StringifyMethodJSON       StringifyMethod = 0
	StringifyMethodJSONIndent StringifyMethod = 1
	StringifyMethodList       StringifyMethod = 2
	StringifyMethodTable      StringifyMethod = 3
)

func AsStringer(m interface{}, method ...StringifyMethod) fmt.Stringer {
	var f StringifyMethod
	if len(method) > 0 {
		f = method[0]
	}
	return Stringify{m: m, f: f}
}

type Stringify struct {
	m interface{}
	f StringifyMethod
}

func (s Stringify) urlValuesString(mp url.Values) string {
	switch s.f {
	case StringifyMethodList:
		sb := strings.Builder{}
		var mx int
		for key := range mp {
			mx = max(mx, len(key))
		}
		for key, values := range mp {
			for _, value := range values {
				sb.WriteString(strings.Repeat(` `, mx-len(key)) + key + ` : ` + value + "\n")
			}
		}
		return sb.String()
	case StringifyMethodTable:
		sb := strings.Builder{}
		var mx int
		var nx int
		for key, values := range mp {
			mx = max(mx, len(key))
			for _, value := range values {
				nx = max(nx, len(value))
			}
		}
		hd := `+` + strings.Repeat(`-`, mx+2) + `+` + strings.Repeat(`-`, nx+2) + "+\n"
		sb.WriteString(hd)
		for key, values := range mp {
			for _, value := range values {
				sb.WriteString(`| ` + strings.Repeat(` `, mx-len(key)) + key + ` | ` + value + strings.Repeat(` `, nx-len(value)) + " |\n")
			}
		}
		sb.WriteString(hd)
		return sb.String()
	}
	return ``
}

func (s Stringify) String() string {
	switch s.f {
	case StringifyMethodList, StringifyMethodTable:
		switch mp := s.m.(type) {
		case map[string][]string:
			return s.urlValuesString(mp)
		case url.Values:
			return s.urlValuesString(mp)
		}
		fallthrough
	case StringifyMethodJSONIndent:
		b, _ := json.MarshalIndent(s, ``, `  `)
		return string(b)
	default:
		b, _ := json.Marshal(s)
		return string(b)
	}
}
