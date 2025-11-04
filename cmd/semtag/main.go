package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/google-internal/semtag/internal/git"
)

type cli struct {
	Next    nextCommand    `cmd:"next" help:"Calculate the next semantic version based on commits and branch" default:"1"`
	Current currentCommand `cmd:"current" help:"Get the highest version tag reachable from the current branch"`
}

type nextCommand struct {
	Prefix       string   `help:"Version prefix" short:"p"`
	BranchSuffix []string `help:"Custom branch to pre-release suffix mapping in branch:suffix format" name:"branch-suffix"`
}

type currentCommand struct {
	Prefix string `help:"Version prefix" short:"p"`
}

func main() {
	var root cli
	app := kong.Parse(&root,
		kong.Name("semtag"),
		kong.Description("Semantic version tagging helper"),
	)

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func (cmd *nextCommand) Run() error {
	if err := ensureRepo(); err != nil {
		return err
	}

	overrides, err := parseBranchSuffixPairs(cmd.BranchSuffix)
	if err != nil {
		return err
	}

	version, err := NextVersion(cmd.Prefix, overrides)
	if err != nil {
		return err
	}

	fmt.Println(formatVersion(cmd.Prefix, version))
	return nil
}

func (cmd *currentCommand) Run() error {
	if err := ensureRepo(); err != nil {
		return err
	}

	version, err := CurrentVersion(cmd.Prefix)
	if err != nil {
		return err
	}

	fmt.Println(formatVersion(cmd.Prefix, version))
	return nil
}

func ensureRepo() error {
	if git.IsRepo() {
		return nil
	}
	return errors.New("current directory is not a git repository")
}

func parseBranchSuffixPairs(values []string) (map[string]string, error) {
	if len(values) == 0 {
		return nil, nil
	}

	mapping := make(map[string]string)
	for _, entry := range values {
		candidate := strings.TrimSpace(entry)
		if candidate == "" {
			continue
		}

		parts := strings.SplitN(candidate, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid branch-suffix format: %s", entry)
		}

		branch := strings.ToLower(strings.TrimSpace(parts[0]))
		suffix := strings.TrimSpace(parts[1])
		if branch == "" || suffix == "" {
			return nil, fmt.Errorf("invalid branch-suffix format: %s", entry)
		}

		mapping[branch] = suffix
	}

	if len(mapping) == 0 {
		return nil, nil
	}

	return mapping, nil
}

func formatVersion(prefix, version string) string {
	if prefix == "" {
		return version
	}
	if strings.HasPrefix(version, prefix) {
		return version
	}
	return prefix + version
}
