package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/google-internal/semtag/internal/git"
)

var (
	breakingBody = regexp.MustCompile("(?m).*BREAKING[ -]CHANGE:.*")
	breaking     = regexp.MustCompile(`(?im).*(\w+)(\(.*\))?!:.*`)
	feature      = regexp.MustCompile(`(?im).*feat(\(.*\))?:.*`)
	patch        = regexp.MustCompile(`(?im).*fix(\(.*\))?:.*`)
)

func CurrentVersion(prefix string) (string, error) {
	currentTag, err := git.DescribeTag(git.TagModeCurrent, "")
	if err != nil {
		return "", fmt.Errorf("failed to get current tag: %w", err)
	}

	if currentTag == "" {
		return "0.0.0", nil
	}

	return strings.TrimPrefix(currentTag, prefix), nil
}

func NextVersion(prefix string, overrides map[string]string) (string, error) {
	stableTag, err := git.DescribeStableTag(git.TagModeCurrent, "")
	if err != nil {
		return "", fmt.Errorf("failed to get stable tag: %w", err)
	}

	current, err := versionFromTag(stableTag, prefix)
	if err != nil {
		return "", fmt.Errorf("could not parse stable tag '%s': %w", stableTag, err)
	}

	next, err := nextVersionFromGit(current, stableTag)
	if err != nil {
		return "", err
	}

	nextWithSuffix, err := applyBranchSuffix(next, prefix, overrides)
	if err != nil {
		return "", err
	}

	return nextWithSuffix.String(), nil
}

func versionFromTag(tag, prefix string) (*semver.Version, error) {
	if tag == "" {
		return semver.NewVersion(strings.TrimPrefix("0.0.0", prefix))
	}
	return semver.NewVersion(strings.TrimPrefix(tag, prefix))
}

func nextVersionFromGit(current *semver.Version, stableTag string) (semver.Version, error) {
	commits, err := git.Changelog(stableTag, nil)
	if err != nil {
		return semver.Version{}, fmt.Errorf("failed to get changelog: %w", err)
	}

	return findNext(current, commits), nil
}

func applyBranchSuffix(version semver.Version, prefix string, overrides map[string]string) (semver.Version, error) {
	resolver := newBranchSuffixResolver(overrides)
	branchSuffix := resolver.suffixForBranch(git.CurrentBranch())
	if branchSuffix == "" {
		return version, nil
	}

	existingTag, err := git.GetLatestTagWithSuffix(branchSuffix, git.TagModeCurrent, "")
	if err != nil {
		return version, fmt.Errorf("failed to get latest tag with suffix '%s': %w", branchSuffix, err)
	}

	var nextSuffix string
	if existingTag == "" {
		nextSuffix = branchSuffix + ".1"
	} else {
		existingVersion, err := versionFromTag(existingTag, prefix)
		if err != nil {
			return version, fmt.Errorf("failed to parse existing suffix tag '%s': %w", existingTag, err)
		}

		if existingVersion.Major() == version.Major() &&
			existingVersion.Minor() == version.Minor() &&
			existingVersion.Patch() == version.Patch() {
			prerelease := existingVersion.Prerelease()
			if prerelease == "" {
				nextSuffix = branchSuffix + ".1"
			} else {
				nextNumber := getNextPrereleaseNumber(prerelease, branchSuffix)
				nextSuffix = branchSuffix + "." + strconv.Itoa(nextNumber)
			}
		} else {
			nextSuffix = branchSuffix + ".1"
		}
	}

	withSuffix, err := version.SetPrerelease(nextSuffix)
	if err != nil {
		return version, fmt.Errorf("failed to set prerelease suffix '%s': %w", nextSuffix, err)
	}

	return withSuffix, nil
}

func getNextPrereleaseNumber(prerelease, suffix string) int {
	pattern := suffix + "."
	if strings.HasPrefix(prerelease, pattern) {
		numberStr := strings.TrimPrefix(prerelease, pattern)
		if number, err := strconv.Atoi(numberStr); err == nil {
			return number + 1
		}
	}

	return 1
}

func findNext(current *semver.Version, changes []git.Commit) semver.Version {
	var major, minor, patchCommit *git.Commit
	for i := range changes {
		if isBreaking(changes[i]) {
			major = &changes[i]
			break
		}

		if minor == nil && isFeature(changes[i]) {
			minor = &changes[i]
		}

		if patchCommit == nil && isPatch(changes[i]) {
			patchCommit = &changes[i]
		}
	}

	if major != nil {
		log.Printf("detected breaking change: %s %s", major.SHA, major.Title)
		return current.IncMajor()
	}

	if minor != nil {
		log.Printf("detected feature: %s %s", minor.SHA, minor.Title)
		return current.IncMinor()
	}

	if patchCommit != nil {
		log.Printf("detected fix: %s %s", patchCommit.SHA, patchCommit.Title)
		return current.IncPatch()
	}

	return *current
}

func isBreaking(commit git.Commit) bool {
	return breakingBody.MatchString(commit.Body) || breaking.MatchString(commit.Title)
}

func isFeature(commit git.Commit) bool {
	return feature.MatchString(commit.Title)
}

func isPatch(commit git.Commit) bool {
	return patch.MatchString(commit.Title)
}
