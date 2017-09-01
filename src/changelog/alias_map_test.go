package changelog_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/skuid/changelog/src/changelog"
)

func TestSectionFor(t *testing.T) {
	cases := []struct {
		alias string
		want  string
	}{
		{"ft", "Features"},
		{"feat", "Features"},
		{"fx", "Bug Fixes"},
		{"fix", "Bug Fixes"},
		{"perf", "Performance"},
		{"unk", "Unknown"},
		{"breaks", "Breaking Changes"},
		{"anything", "Unknown"},
		{"", "Unknown"},
	}

	aMap := changelog.NewSectionAliasMap()
	for i := range cases {
		if got := aMap.SectionFor(cases[i].alias); got != cases[i].want {
			t.Errorf("Section title doesn't match! Expected '%s', got '%s'", cases[i].want, got)
		}
	}
}

func TestMergeSectionAliasMaps(t *testing.T) {
	cases := []struct {
		original changelog.SectionAliasMap
		addition changelog.SectionAliasMap
		want     changelog.SectionAliasMap
	}{
		{
			changelog.SectionAliasMap{"Features": []string{"new"}},
			changelog.SectionAliasMap{"features": []string{"feat"}},
			changelog.SectionAliasMap{"Features": []string{"feat", "new"}},
		},
		{
			changelog.SectionAliasMap{"Features": []string{"feat"}},
			changelog.SectionAliasMap{"bug fixes": []string{"fix"}},
			changelog.SectionAliasMap{"Features": []string{"feat"}, "Bug Fixes": []string{"fix"}},
		},
	}
	for _, c := range cases {
		got := changelog.MergeSectionAliasMaps(c.original, c.addition)
		if !reflect.DeepEqual(got, c.want) {
			errorDiff(t, "AliasMaps not equal!", fmt.Sprintf("%v", c.want), fmt.Sprintf("%v", got))
		}
	}
}
func TestGrep(t *testing.T) {
	cases := []struct {
		aMap changelog.SectionAliasMap
		want string
	}{
		{
			changelog.SectionAliasMap{"Features": []string{"feat", "new", ""}},
			"BREAKING|^feat|^new",
		},
		{
			changelog.SectionAliasMap{"Features": []string{"feat"}, "Bug Fixes": []string{"fix"}},
			"BREAKING|^feat|^fix",
		},
		{
			changelog.NewSectionAliasMap(),
			"BREAKING|^breaks|^feat|^fix|^ft|^fx|^perf|^unk",
		},
	}
	for _, c := range cases {
		if got := c.aMap.Grep(); got != c.want {
			errorDiff(t, "SectionAliasMap{}.Grep() not equal!", c.want, got)
		}
	}
}
