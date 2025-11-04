package main

import (
	"os"
	"sort"
	"strings"
)

type branchSuffixResolver struct {
	exact  map[string]string
	prefix []prefixRule
}

type prefixRule struct {
	prefix string
	suffix string
}

func newBranchSuffixResolver(overrides map[string]string) *branchSuffixResolver {
	mapping := defaultBranchSuffixMapping()

	for key, value := range loadBranchSuffixEnvOverrides() {
		mapping[key] = value
	}

	for key, value := range overrides {
		if strings.TrimSpace(key) == "" {
			continue
		}
		mapping[strings.ToLower(strings.TrimSpace(key))] = strings.TrimSpace(value)
	}

	resolver := &branchSuffixResolver{
		exact:  make(map[string]string),
		prefix: make([]prefixRule, 0),
	}

	for key, value := range mapping {
		normalizedKey := strings.ToLower(strings.TrimSpace(key))
		normalizedValue := strings.TrimSpace(value)
		if normalizedKey == "" || normalizedValue == "" {
			continue
		}

		if strings.HasSuffix(normalizedKey, "/*") {
			resolver.prefix = append(resolver.prefix, prefixRule{
				prefix: strings.TrimSuffix(normalizedKey, "/*"),
				suffix: normalizedValue,
			})
		} else {
			resolver.exact[normalizedKey] = normalizedValue
		}
	}

	sort.SliceStable(resolver.prefix, func(i, j int) bool {
		return len(resolver.prefix[i].prefix) > len(resolver.prefix[j].prefix)
	})

	return resolver
}

func (r *branchSuffixResolver) suffixForBranch(branch string) string {
	branch = strings.ToLower(strings.TrimSpace(branch))
	if branch == "" {
		return ""
	}

	if suffix, ok := r.exact[branch]; ok {
		return suffix
	}

	for _, rule := range r.prefix {
		if strings.HasPrefix(branch, rule.prefix) {
			return rule.suffix
		}
	}

	return ""
}

func defaultBranchSuffixMapping() map[string]string {
	return map[string]string{
		"beta":         "beta",
		"develop":      "beta",
		"dev":          "beta",
		"staging":      "beta",
		"alpha":        "alpha",
		"experimental": "alpha",
		"wip":          "alpha",
		"rc":           "rc",
		"release":      "rc",
		"hotfix":       "hotfix",
		"release/*":    "rc",
		"feature/*":    "alpha",
		"hotfix/*":     "beta",
	}
}

func loadBranchSuffixEnvOverrides() map[string]string {
	result := make(map[string]string)

	if envMapping := os.Getenv("SVU_BRANCH_SUFFIX_MAPPING"); envMapping != "" {
		pairs := strings.Split(envMapping, ",")
		for _, pair := range pairs {
			parts := strings.Split(strings.TrimSpace(pair), ":")
			if len(parts) != 2 {
				continue
			}
			branch := strings.ToLower(strings.TrimSpace(parts[0]))
			suffix := strings.TrimSpace(parts[1])
			if branch != "" && suffix != "" {
				result[branch] = suffix
			}
		}
	}

	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, "SVU_BRANCH_") || !strings.HasSuffix(env, "_SUFFIX") {
			continue
		}
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		branchEnv := parts[0]
		value := strings.TrimSpace(parts[1])
		if value == "" {
			continue
		}
		branchName := strings.TrimPrefix(branchEnv, "SVU_BRANCH_")
		branchName = strings.TrimSuffix(branchName, "_SUFFIX")
		branchName = strings.ToLower(branchName)
		if branchName != "" {
			result[branchName] = value
		}
	}

	return result
}
