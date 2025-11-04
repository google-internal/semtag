package git

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gobwas/glob"
)

// Commit is a commit with a hash, title (first line of the message), and body
// (rest of the message, not including the title).
type Commit struct {
	SHA   string
	Title string
	Body  string
}

func (c Commit) String() string {
	return c.SHA + ": " + c.Title + "\n" + c.Body
}

const (
	TagModeAll     = "all"
	TagModeCurrent = "current"
)

// GetLatestTagWithSuffix returns the latest tag with the specified suffix
func GetLatestTagWithSuffix(suffix string, tagMode string, pattern string) (string, error) {
	args := []string{}
	if tagMode == TagModeCurrent {
		args = []string{"--merged"}
	}
	tags, err := getAllTags(args...)
	if err != nil {
		return "", err
	}

	if len(tags) == 0 {
		return "", nil
	}

	// Filter tags with the specified suffix
	var suffixTags []string
	for _, tag := range tags {
		if hasSuffix(tag, suffix) {
			suffixTags = append(suffixTags, tag)
		}
	}

	if len(suffixTags) == 0 {
		return "", nil
	}

	// Apply pattern filter if specified
	if pattern != "" {
		g, err := glob.Compile(pattern)
		if err != nil {
			return "", err
		}
		for _, tag := range suffixTags {
			if g.Match(tag) {
				return tag, nil
			}
		}
		return "", fmt.Errorf("no tags with suffix '%s' match pattern '%s'", suffix, pattern)
	}

	return suffixTags[0], nil
}

// hasSuffix checks if a tag has the specified suffix
func hasSuffix(tag, suffix string) bool {
	// Remove common prefixes like 'v'
	tag = strings.TrimPrefix(tag, "v")

	// Check if tag ends with the suffix or has the suffix followed by a number
	suffixPattern := "-" + suffix
	if strings.Contains(tag, suffixPattern) {
		// Check if it's exactly the suffix or suffix with a number
		parts := strings.Split(tag, suffixPattern)
		if len(parts) == 2 {
			remaining := parts[1]
			// If remaining is empty or starts with a number, it's a valid suffix
			if remaining == "" || (len(remaining) > 0 && (remaining[0] == '.' || isDigit(remaining[0]))) {
				return true
			}
		}
	}

	return false
}

// isDigit checks if a character is a digit
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// copied from goreleaser

// IsRepo returns true if current folder is a git repository
func IsRepo() bool {
	out, err := run("rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(out) == "true"
}

func Root() string {
	out, _ := run("rev-parse", "--show-toplevel")
	return strings.TrimSpace(out)
}

// CurrentBranch returns the current branch name
func CurrentBranch() string {
	// First try to get the current branch name
	out, err := run("rev-parse", "--abbrev-ref", "HEAD")
	if err == nil && strings.TrimSpace(out) != "HEAD" {
		return strings.TrimSpace(out)
	}

	// If we're in detached HEAD state, try to find the branch containing HEAD
	out, err = run("branch", "-r", "--contains", "HEAD")
	if err == nil && out != "" {
		lines := strings.Split(strings.TrimSpace(out), "\n")
		if len(lines) > 0 {
			// Take the first remote branch and clean it up
			branch := strings.TrimSpace(lines[0])
			// Remove "origin/" prefix if present
			branch = strings.TrimPrefix(branch, "origin/")
			return branch
		}
	}

	// Fallback: try to get branch from git symbolic-ref
	out, err = run("symbolic-ref", "--short", "HEAD")
	if err == nil {
		return strings.TrimSpace(out)
	}

	// Last resort: return "HEAD"
	return "HEAD"
}

func getAllTags(args ...string) ([]string, error) {
	tags, err := run(append([]string{"-c", "versionsort.suffix=-", "tag", "--sort=-version:refname"}, args...)...)
	if err != nil {
		return nil, err
	}
	return strings.Split(tags, "\n"), nil
}

func DescribeTag(tagMode string, pattern string) (string, error) {
	args := []string{}
	if tagMode == TagModeCurrent {
		args = []string{"--merged"}
	}
	tags, err := getAllTags(args...)
	if err != nil {
		return "", err
	}

	if len(tags) == 0 {
		return "", nil
	}
	if pattern == "" {
		return tags[0], nil
	}

	g, err := glob.Compile(pattern)
	if err != nil {
		return "", err
	}
	for _, tag := range tags {
		if g.Match(tag) {
			return tag, nil
		}
	}
	return "", fmt.Errorf("no tags match '%s'", pattern)
}

// DescribeStableTag returns the latest stable (non-prerelease) tag from the main branch
// This follows semantic-release standards where the base version should be the latest
// stable release from the main branch, not the latest tag including prereleases
func DescribeStableTag(tagMode string, pattern string) (string, error) {
	args := []string{}
	if tagMode == TagModeCurrent {
		args = []string{"--merged"}
	}
	tags, err := getAllTags(args...)
	if err != nil {
		return "", err
	}

	if len(tags) == 0 {
		return "", nil
	}

	// Filter out prerelease tags (tags with suffixes like -alpha, -beta, -rc, etc.)
	var stableTags []string
	for _, tag := range tags {
		// Check if tag is a stable release (no prerelease suffix)
		if isStableTag(tag) {
			stableTags = append(stableTags, tag)
		}
	}

	if len(stableTags) == 0 {
		return "", nil
	}

	if pattern == "" {
		return stableTags[0], nil
	}

	g, err := glob.Compile(pattern)
	if err != nil {
		return "", err
	}
	for _, tag := range stableTags {
		if g.Match(tag) {
			return tag, nil
		}
	}
	return "", fmt.Errorf("no stable tags match '%s'", pattern)
}

// isStableTag checks if a tag represents a stable release (no prerelease suffix)
func isStableTag(tag string) bool {
	// Remove common prefixes like 'v'
	tag = strings.TrimPrefix(tag, "v")

	// Check if tag contains prerelease indicators
	prereleaseIndicators := []string{"-alpha", "-beta", "-rc", "-pre", "-dev", "-snapshot"}
	for _, indicator := range prereleaseIndicators {
		if strings.Contains(tag, indicator) {
			return false
		}
	}

	// Additional check: if tag contains a dash and the part after dash is not numeric,
	// it's likely a prerelease
	parts := strings.Split(tag, "-")
	if len(parts) > 1 {
		// Check if the last part is numeric (like in 1.0.0-1 which might be a build number)
		lastPart := parts[len(parts)-1]
		if _, err := strconv.Atoi(lastPart); err != nil {
			// If last part is not numeric, it's likely a prerelease identifier
			return false
		}
	}

	return true
}

func Changelog(tag string, dirs []string) ([]Commit, error) {
	if tag == "" {
		return gitLog(dirs, "HEAD")
	} else {
		return gitLog(dirs, fmt.Sprintf("tags/%s..HEAD", tag))
	}
}

func run(args ...string) (string, error) {
	extraArgs := []string{
		"-c", "log.showSignature=false",
	}
	args = append(extraArgs, args...)
	/* #nosec */
	cmd := exec.Command("git", args...)
	bts, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New(string(bts))
	}
	return string(bts), nil
}

func gitLog(dirs []string, refs ...string) ([]Commit, error) {
	args := []string{"log", "--no-decorate", "--no-color", `--format=%H:%B<svu-commit-end>`}
	args = append(args, refs...)
	if len(dirs) > 0 {
		args = append(args, "--")
		args = append(args, dirs...)
	}
	s, err := run(args...)
	if err != nil {
		return nil, err
	}
	var result []Commit
	for _, commit := range strings.Split(s, "<svu-commit-end>") {
		commit = strings.TrimSpace(commit)
		if commit == "" { // accounts for the last split, which will be an empty line
			continue
		}

		hashEndIdx := strings.Index(commit, ":")
		titleEndIdx := strings.Index(commit, "\n")
		if titleEndIdx < 0 {
			titleEndIdx = len(commit)
		}

		result = append(result, Commit{
			commit[:hashEndIdx],
			commit[hashEndIdx+1 : titleEndIdx],
			commit[min(titleEndIdx+1, len(commit)):],
		})
	}
	return result, nil
}
