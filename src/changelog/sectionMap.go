package changelog

import (
	"sort"
)

// ComponentMap is a map whose keys are the components changed, and values are
// any associated commits
type ComponentMap map[string]Commits

// SectionMap is a map whose keys are the commit type, and values are a
// ComponentMap
type SectionMap struct {
	Sections map[string]ComponentMap
	order    []string
}

// Order returns the order the sections will appear in the change log
func (s SectionMap) Order() []string {
	return s.order
}

// SetOrder sets the order of the sections in the changelog. Any non-specified
// sections will be appended to the end alphabetically. Any sections that don't
// exist will be discarded.
//
// "Unknown" section is always last
func (s *SectionMap) SetOrder(order []string) {
	newOrder := []string{}

	// used for keeping track of what we've already seen
	sectionsUsed := map[string]bool{}
	for _, section := range order {
		if _, ok := s.Sections[section]; !ok || section == "Unknown" {
			continue
		}
		sectionsUsed[section] = true
		newOrder = append(newOrder, section)
	}

	additionalSections := []string{}
	for section := range s.Sections {
		if _, ok := sectionsUsed[section]; ok || section == "Unknown" {
			// We've already added it
			continue
		}
		additionalSections = append(additionalSections, section)
	}
	sort.Strings(additionalSections)
	additionalSections = append(additionalSections, "Unknown")
	s.order = append(newOrder, additionalSections...)
}

// DefaultOrder defines the default order of sections in the change log
var DefaultOrder = []string{
	"Features",
	"Bug Fixes",
	"Performance",
	"Breaking Changes",
	"Unknown",
}

// NewSectionMap returns a new SectionMap for the given commits
func NewSectionMap(commits Commits) SectionMap {
	sections := make(map[string]ComponentMap)

	for _, entry := range commits {
		if len(entry.Breaks) != 0 {
			breakKey := "Breaking Changes"
			if _, ok := sections[breakKey]; !ok {
				sections[breakKey] = make(ComponentMap)
			}

			// Get the ComponentMap for the commit's component
			secMap, ok := sections[breakKey][entry.Component]
			if !ok {
				secMap = make(Commits, 0)
			}
			sections[breakKey][entry.Component] = append(secMap, entry)
		}

		// Check if the map for the given commit type exists
		// If it doesn't exist, create it and set componentMap to the proper
		// value (an empty ComponentMap)
		if _, ok := sections[entry.CommitType]; !ok {
			sections[entry.CommitType] = make(ComponentMap)
		}

		// Get the ComponentMap for the commit's component
		secMap, ok := sections[entry.CommitType][entry.Component]
		if !ok {
			secMap = make(Commits, 0)
		}
		sections[entry.CommitType][entry.Component] = append(secMap, entry)
	}

	s := SectionMap{Sections: sections}
	s.SetOrder(DefaultOrder)
	return s
}
