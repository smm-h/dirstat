package scan

import (
	"sort"
	"strings"
)

// SingletonsGroup is the format name of the pseudo-group produced by
// CollapseSingletons.
const SingletonsGroup = "(singletons)"

// Stat names accepted by --stats (canonical column order).
var StatNames = []string{
	"count",
	"total-size", "min-size", "max-size", "avg-size",
	"total-loc", "min-loc", "max-loc", "avg-loc",
}

// IsLOCStat reports whether name is one of the LOC stat names.
func IsLOCStat(name string) bool {
	return strings.HasSuffix(name, "-loc")
}

// sortValue returns the numeric sort value of group g for a stat key. For
// LOC keys, groups without LOC (binary groups) sort as 0.
func sortValue(g *Group, key string) int64 {
	switch key {
	case "count":
		return int64(g.Count)
	case "total-size":
		return g.TotalSize
	case "min-size":
		return g.MinSize
	case "max-size":
		return g.MaxSize
	case "avg-size":
		return g.AvgSize
	}
	if !g.HasLOC {
		return 0
	}
	switch key {
	case "total-loc":
		return g.TotalLOC
	case "min-loc":
		return g.MinLOC
	case "max-loc":
		return g.MaxLOC
	case "avg-loc":
		return g.AvgLOC
	}
	return 0
}

// SortGroups sorts groups in place by the given keys ("format" or a stat
// name) in precedence order, all in the same direction (desc when desc is
// true). Ties are broken by format name ascending (case-insensitive, then
// case-sensitive) so the output is deterministic.
func SortGroups(groups []Group, keys []string, desc bool) {
	sort.Slice(groups, func(i, j int) bool {
		gi, gj := &groups[i], &groups[j]
		for _, key := range keys {
			var c int
			if key == "format" {
				c = strings.Compare(strings.ToLower(gi.Format), strings.ToLower(gj.Format))
			} else {
				vi, vj := sortValue(gi, key), sortValue(gj, key)
				switch {
				case vi < vj:
					c = -1
				case vi > vj:
					c = 1
				}
			}
			if c != 0 {
				if desc {
					return c > 0
				}
				return c < 0
			}
		}
		li, lj := strings.ToLower(gi.Format), strings.ToLower(gj.Format)
		if li != lj {
			return li < lj
		}
		return gi.Format < gj.Format
	})
}

// CollapseSingletons merges all groups with exactly one file into a single
// "(singletons)" pseudo-group: count is the number of merged groups, sizes
// are aggregated over all merged groups, and LOC values are aggregated over
// the text merged groups only. The input is not modified.
func CollapseSingletons(groups []Group) []Group {
	var out []Group
	var singles []Group
	for _, g := range groups {
		if g.Count == 1 && !g.Pseudo {
			singles = append(singles, g)
		} else {
			out = append(out, g)
		}
	}
	if len(singles) == 0 {
		return groups
	}

	merged := Group{Format: SingletonsGroup, Pseudo: true, Count: len(singles)}
	textCount := int64(0)
	for i, s := range singles {
		if i == 0 {
			merged.MinSize, merged.MaxSize = s.MinSize, s.MaxSize
		} else {
			merged.MinSize = min(merged.MinSize, s.MinSize)
			merged.MaxSize = max(merged.MaxSize, s.MaxSize)
		}
		merged.TotalSize += s.TotalSize
		if s.HasLOC {
			if textCount == 0 {
				merged.MinLOC, merged.MaxLOC = s.MinLOC, s.MaxLOC
			} else {
				merged.MinLOC = min(merged.MinLOC, s.MinLOC)
				merged.MaxLOC = max(merged.MaxLOC, s.MaxLOC)
			}
			merged.TotalLOC += s.TotalLOC
			textCount++
		}
	}
	merged.AvgSize = roundDiv(merged.TotalSize, int64(merged.Count))
	if textCount > 0 {
		merged.HasLOC = true
		merged.AvgLOC = roundDiv(merged.TotalLOC, textCount)
	}
	return append(out, merged)
}
