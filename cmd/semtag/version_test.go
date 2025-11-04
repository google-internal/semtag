package main

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/google-internal/semtag/internal/git"
	"github.com/stretchr/testify/require"
)

func TestIsBreaking(t *testing.T) {
	for _, commit := range []git.Commit{
		{Title: "feat!: foo"},
		{Title: "chore(lala)!: foo"},
		{Title: "docs: lalala", Body: "BREAKING CHANGE: lalal"},
		{Title: "docs: lalala", Body: "BREAKING-CHANGE: lalal"},
	} {
		t.Run(commit.String(), func(t *testing.T) {
			require.True(t, isBreaking(commit))
		})
	}

	for _, commit := range []git.Commit{
		{Title: "feat: foo"},
		{Title: "chore(lol): foo"},
		{Title: "docs: lalala"},
		{Title: "docs: BREAKING change: lalal"},
		{Title: "docs: breaking-change: aehijhk"},
		{Title: "docs: BREAKING_CHANGE: foo"},
	} {
		t.Run(commit.String(), func(t *testing.T) {
			require.False(t, isBreaking(commit))
		})
	}
}

func TestIsFeature(t *testing.T) {
	for _, commit := range []git.Commit{
		{Title: "feat: foo"},
		{Title: "feat(lalal): foobar"},
	} {
		t.Run(commit.String(), func(t *testing.T) {
			require.True(t, isFeature(commit))
		})
	}

	for _, commit := range []git.Commit{
		{Title: "fix: foo"},
		{Title: "chore: foo"},
		{Title: "docs: lalala"},
		{Title: "ci: foo"},
		{Title: "test: foo"},
		{Title: "Merge remote-tracking branch 'origin/main'"},
		{Title: "refactor: foo bar"},
	} {
		t.Run(commit.String(), func(t *testing.T) {
			require.False(t, isFeature(commit))
		})
	}
}

func TestIsPatch(t *testing.T) {
	for _, commit := range []git.Commit{
		{Title: "fix: foo"},
		{Title: "fix(lalal): lalala"},
	} {
		t.Run(commit.String(), func(t *testing.T) {
			require.True(t, isPatch(commit))
		})
	}

	for _, commit := range []git.Commit{
		{Title: "chore: foobar"},
		{Title: "docs: something"},
		{Title: "invalid commit"},
	} {
		t.Run(commit.String(), func(t *testing.T) {
			require.False(t, isPatch(commit))
		})
	}
}

func TestFindNext(t *testing.T) {
	version0a := semver.MustParse("v0.4.5")
	version0b := semver.MustParse("v0.5.5")
	version1 := semver.MustParse("v1.2.3")
	version2 := semver.MustParse("v2.4.12")
	version3 := semver.MustParse("v3.4.5-beta34+ads")

	for expected, next := range map[string]semver.Version{
		"0.4.5": findNext(version0a, []git.Commit{{Title: "chore: should do nothing"}}),
		"0.4.6": findNext(version0a, []git.Commit{{Title: "fix: inc patch"}}),
		"0.5.0": findNext(version0a, []git.Commit{{Title: "feat: inc minor"}}),
		"1.0.0": findNext(version0b, []git.Commit{{Title: "feat!: inc minor"}}),
		"1.2.3": findNext(version1, []git.Commit{{Title: "chore: should do nothing"}}),
		"1.3.0": findNext(version1, []git.Commit{{Title: "feat: inc major"}}),
		"2.0.0": findNext(version1, []git.Commit{{Title: "chore!: hashbang incs major"}}),
		"3.0.0": findNext(version2, []git.Commit{{Title: "feat: something", Body: "BREAKING CHANGE: increases major"}}),
		"3.5.0": findNext(version3, []git.Commit{{Title: "feat: inc major"}}),
	} {
		t.Run(expected, func(t *testing.T) {
			require.Equal(t, expected, next.String())
		})
	}
}

func TestGetNextPrereleaseNumber(t *testing.T) {
	t.Run("increments existing number", func(t *testing.T) {
		require.Equal(t, 3, getNextPrereleaseNumber("beta.2", "beta"))
	})

	t.Run("falls back to one", func(t *testing.T) {
		require.Equal(t, 1, getNextPrereleaseNumber("alpha", "alpha"))
	})
}
