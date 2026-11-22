package parser

import "slices"

type Filter struct {
	ID   []int
	Type []string
	Name []string
}

func (f *Filter) filterID(target int) bool {
	return shouldExclude(f.ID, target)
}

func (f *Filter) filterType(target string) bool {
	return shouldExclude(f.Type, target)
}

func (f *Filter) filterName(target string) bool {
	return shouldExclude(f.Name, target)
}

func shouldExclude[T comparable](slice []T, target T) bool {
	return len(slice) > 0 && !slices.Contains(slice, target)
}
