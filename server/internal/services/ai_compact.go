package services

import (
	"bytes"
	"encoding/json"

	"the-slot-finder/server/internal/constants"
)

// aiOmittedSuffix names the sibling key that counts the items a trimmed list lost; a trimmed list
// that has no parent object is wrapped as {"items": [...], "_omitted_items": n} instead.
const (
	aiOmittedSuffix = "_omitted_items"
	aiItemsKey      = "items"
)

// compactJSON shrinks a JSON result to at most limit bytes and keeps it valid JSON: it drops nulls
// and empty values, then keeps fewer items in long lists (marking how many were left out). Only when
// that does not fit does it cut the text and add the cut marker.
func compactJSON(data []byte, limit int) string {
	if len(data) <= limit {
		return string(data)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber() // keep numbers as written
	var v any
	if err := dec.Decode(&v); err != nil {
		return cutText(data, limit)
	}
	pruned, _ := pruneEmpty(v)
	out, err := json.Marshal(pruned)
	if err != nil {
		return cutText(data, limit)
	}
	if len(out) <= limit {
		return string(out)
	}
	// The largest list size that fits; fits is monotonic, so binary search it.
	lo, hi, best := 0, longestList(pruned), []byte(nil)
	for lo <= hi {
		mid := (lo + hi) / 2
		trimmed, err := json.Marshal(trimLists(pruned, mid))
		if err == nil && len(trimmed) <= limit {
			best, lo = trimmed, mid+1
		} else {
			hi = mid - 1
		}
	}
	if best != nil {
		return string(best)
	}
	return cutText(out, limit)
}

func cutText(data []byte, limit int) string {
	return string(data[:limit]) + constants.LLMToolCutMarker
}

// pruneEmpty removes nulls, empty strings, and lists and objects that end up empty. The bool says
// whether the value itself is empty.
func pruneEmpty(v any) (any, bool) {
	switch t := v.(type) {
	case nil:
		return nil, true
	case string:
		return t, t == ""
	case []any:
		out := make([]any, 0, len(t))
		for _, item := range t {
			if p, empty := pruneEmpty(item); !empty {
				out = append(out, p)
			}
		}
		return out, len(out) == 0
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, item := range t {
			if p, empty := pruneEmpty(item); !empty {
				out[k] = p
			}
		}
		return out, len(out) == 0
	default:
		return v, false
	}
}

func longestList(v any) int {
	longest := 0
	switch t := v.(type) {
	case []any:
		longest = len(t)
		for _, item := range t {
			longest = max(longest, longestList(item))
		}
	case map[string]any:
		for _, item := range t {
			longest = max(longest, longestList(item))
		}
	}
	return longest
}

// trimLists keeps the first n items of every list and records how many were left out.
func trimLists(v any, n int) any {
	switch t := v.(type) {
	case []any:
		kept := trimItems(t[:min(n, len(t))], n)
		if len(t) <= n {
			return kept
		}
		return map[string]any{aiItemsKey: kept, aiOmittedSuffix: len(t) - n}
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, item := range t {
			if list, ok := item.([]any); ok && len(list) > n {
				out[k] = trimItems(list[:n], n)
				out[k+aiOmittedSuffix] = len(list) - n
				continue
			}
			out[k] = trimLists(item, n)
		}
		return out
	default:
		return v
	}
}

func trimItems(items []any, n int) []any {
	out := make([]any, len(items))
	for i, item := range items {
		out[i] = trimLists(item, n)
	}
	return out
}
