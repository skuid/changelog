package changelog

import (
	"fmt"
	"sort"
	"strings"
)

// SectionAliasMap is for associating commit prefixes to a section of the
// changelog
type SectionAliasMap map[string][]string

// NewSectionAliasMap returns the default map
func NewSectionAliasMap() SectionAliasMap {
	sectionAliasMap := make(SectionAliasMap)
	sectionAliasMap["Features"] = []string{"ft", "feat"}
	sectionAliasMap["Bug Fixes"] = []string{"fx", "fix"}
	sectionAliasMap["Performance"] = []string{"perf"}
	sectionAliasMap["Breaking Changes"] = []string{"breaks"}
	sectionAliasMap["Unknown"] = []string{"unk"}
	return sectionAliasMap
}

// SectionFor returns the section title for a given alias
func (s SectionAliasMap) SectionFor(alias string) string {
	for title, aliases := range s {
		for i := range aliases {
			if aliases[i] == alias {
				return strings.Title(title)
			}
		}
	}
	return "Unknown"
}

// MergeSectionAliasMaps merges multiple maps into the first and returns the first
func MergeSectionAliasMaps(first SectionAliasMap, additional ...SectionAliasMap) SectionAliasMap {
	for _, successive := range additional {
		for title, aliases := range successive {
			title = strings.Title(title)
			// key doesn't exist in the first map, just add it
			if _, ok := first[title]; !ok {
				first[title] = aliases
			}
			// key already exists, union the values
			first[title] = mergeStringSlices(first[title], aliases)
		}
	}
	return first
}

func mergeStringSlices(first []string, successive ...[]string) []string {
	values := map[string]bool{}
	for _, word := range first {
		values[word] = true
	}
	for _, next := range successive {
		for _, word := range next {
			values[word] = true
		}
	}

	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Grep produces a regex to search for lines starting with each section key
func (s SectionAliasMap) Grep() string {
	prefixes := []string{"BREAKING"}
	for _, items := range s {
		for _, item := range items {
			if item == "" {
				continue
			}
			prefixes = append(prefixes, fmt.Sprintf("^%s", item))
		}
	}
	sort.Strings(prefixes)
	return strings.Join(prefixes, "|")
}
