package dbkp

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// pathFilter stores the include/exclude settings for walking a directory tree.
type pathFilter struct {
	only     map[string]struct{}
	excludes []*regexp.Regexp
}

func newPathFilter(file File) (pathFilter, error) {
	pf := pathFilter{}

	if len(file.Only) > 0 {
		pf.only = make(map[string]struct{}, len(file.Only))
		for _, entry := range file.Only {
			if entry == "" {
				continue
			}
			pf.only[entry] = struct{}{}
		}
	}

	if len(file.Exclude) > 0 {
		pf.excludes = make([]*regexp.Regexp, 0, len(file.Exclude))
		for _, pattern := range file.Exclude {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return pf, fmt.Errorf("invalid exclude pattern %q: %w", pattern, err)
			}
			pf.excludes = append(pf.excludes, re)
		}
	}

	return pf, nil
}

func (pf pathFilter) shouldSkip(rel string, isDir bool) (skip bool, skipDir bool) {
	if rel == "" {
		return false, false
	}

	if len(pf.only) > 0 {
		head := rel
		if idx := strings.IndexRune(rel, filepath.Separator); idx != -1 {
			head = rel[:idx]
		}
		if _, ok := pf.only[head]; !ok {
			if isDir {
				return true, true
			}
			return true, false
		}
	}

	if len(pf.excludes) > 0 {
		normalized := filepath.ToSlash(rel)
		for _, re := range pf.excludes {
			if re.MatchString(normalized) {
				if isDir {
					return true, true
				}
				return true, false
			}
		}
	}

	return false, false
}
