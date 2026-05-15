// Package tokens converts parsed CSS variables into DTCG design token trees.
package tokens

import (
	"sort"
	"strings"

	"t-f/internal/domain"
)

type tokenEntry struct {
	name      string
	lightVal  string
	darkVal   string
	tokenType domain.TokenType
}

func categorizeAndGroup(vars map[string]map[string]string) map[string][]tokenEntry {
	lightMap := vars["light"]
	darkMap := vars["dark"]
	themeMap := vars["theme"]

	allKeys := make(map[string]bool)
	for k := range lightMap {
		allKeys[k] = true
	}
	for k := range darkMap {
		allKeys[k] = true
	}
	for k := range themeMap {
		allKeys[k] = true
	}

	groups := make(map[string][]tokenEntry)

	for key := range allKeys {
		lv := lightMap[key]
		dv := darkMap[key]
		if dv == "" {
			dv = themeMap[key]
		}
		if lv == "" {
			lv = themeMap[key]
		}

		resolvedVal := lv
		if resolvedVal == "" {
			resolvedVal = dv
		}

		category, ttype := domain.InferCategory(key, resolvedVal)
		tn := tokenName(key)

		groups[category] = append(groups[category], tokenEntry{
			name:      tn,
			lightVal:  lv,
			darkVal:   dv,
			tokenType: ttype,
		})
	}

	for _, entries := range groups {
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].name < entries[j].name
		})
	}

	return groups
}

func tokenName(varName string) string {
	cleaned := strings.TrimPrefix(varName, "--")
	parts := strings.SplitN(cleaned, "-", 2)
	if len(parts) < 2 {
		return cleaned
	}
	return parts[1]
}

func buildTree(entries []tokenEntry) domain.DTCGGroup {
	root := make(domain.DTCGGroup)

	for _, e := range entries {
		parts := strings.Split(e.name, "-")
		insertIntoTree(root, parts, e)
	}

	return root
}

func insertIntoTree(group domain.DTCGGroup, parts []string, e tokenEntry) {
	if len(parts) == 1 {
		leaf := buildLeaf(e)
		if existing, ok := group[parts[0]]; ok {
			if existingMap, ok := existing.(domain.DTCGGroup); ok {
				for k, v := range leaf {
					existingMap[k] = v
				}
				return
			}
		}
		group[parts[0]] = leaf
		return
	}

	sub, ok := group[parts[0]]
	if !ok {
		sub = make(domain.DTCGGroup)
		group[parts[0]] = sub
	}
	subGroup := sub.(domain.DTCGGroup)

	if _, hasValue := subGroup["$value"]; hasValue {
		base := make(domain.DTCGGroup)
		for k, v := range subGroup {
			base[k] = v
		}
		for k := range subGroup {
			delete(subGroup, k)
		}
		subGroup["base"] = base
	}

	insertIntoTree(subGroup, parts[1:], e)
}

func flattenTree(entries []tokenEntry) domain.DTCGGroup {
	root := make(domain.DTCGGroup)
	for _, e := range entries {
		name := e.name
		hasDark := e.darkVal != "" && e.darkVal != e.lightVal

		leaf := make(domain.DTCGGroup)
		leaf["$value"] = e.lightVal
		leaf["$type"] = e.tokenType
		root[name] = leaf

		if hasDark {
			darkLeaf := make(domain.DTCGGroup)
			darkLeaf["$value"] = e.darkVal
			darkLeaf["$type"] = e.tokenType
			root[name+"-dark"] = darkLeaf
		}
	}
	return root
}

func buildLeaf(e tokenEntry) domain.DTCGGroup {
	leaf := make(domain.DTCGGroup)

	if e.darkVal != "" && e.darkVal != e.lightVal {
		leaf["$value"] = e.lightVal
		leaf["$type"] = e.tokenType

		darkLeaf := make(domain.DTCGGroup)
		darkLeaf["$value"] = e.darkVal
		darkLeaf["$type"] = e.tokenType
		leaf["dark"] = darkLeaf
	} else {
		leaf["$value"] = e.lightVal
		leaf["$type"] = e.tokenType
	}

	return leaf
}
