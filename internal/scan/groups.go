package scan

import (
	"sort"
	"strings"
)

type ProjectGroup struct {
	Project    string
	Count      int
	TotalBytes int64
	Types      []string
	Candidates []Candidate
}

type TypeGroup struct {
	Name       string
	Count      int
	TotalBytes int64
}

func GroupByProject(candidates []Candidate) []ProjectGroup {
	if len(candidates) == 0 {
		return nil
	}

	groups := map[string]*ProjectGroup{}
	for _, candidate := range candidates {
		group, ok := groups[candidate.Project]
		if !ok {
			group = &ProjectGroup{Project: candidate.Project}
			groups[candidate.Project] = group
		}
		group.Count++
		group.TotalBytes += candidate.SizeBytes
		group.Candidates = append(group.Candidates, candidate)
	}

	items := make([]ProjectGroup, 0, len(groups))
	for _, group := range groups {
		typeSet := map[string]struct{}{}
		for _, candidate := range group.Candidates {
			typeSet[candidate.Name] = struct{}{}
		}
		group.Types = make([]string, 0, len(typeSet))
		for name := range typeSet {
			group.Types = append(group.Types, name)
		}
		sort.Strings(group.Types)
		items = append(items, *group)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].TotalBytes == items[j].TotalBytes {
			return items[i].Project < items[j].Project
		}
		return items[i].TotalBytes > items[j].TotalBytes
	})

	return items
}

func GroupByType(candidates []Candidate) []TypeGroup {
	if len(candidates) == 0 {
		return nil
	}

	groups := map[string]*TypeGroup{}
	for _, candidate := range candidates {
		key := strings.ToLower(candidate.Name)
		group, ok := groups[key]
		if !ok {
			group = &TypeGroup{Name: candidate.Name}
			groups[key] = group
		}
		group.Count++
		group.TotalBytes += candidate.SizeBytes
	}

	items := make([]TypeGroup, 0, len(groups))
	for _, group := range groups {
		items = append(items, *group)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].TotalBytes == items[j].TotalBytes {
			return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
		}
		return items[i].TotalBytes > items[j].TotalBytes
	})

	return items
}
