package cmd

import (
	"errors"
	"strings"
	"testing"
)

func TestSkillInstallDefaultsToProjectScope(t *testing.T) {
	originalLookPath := lookPath
	originalRunExternal := runExternalCommand
	defer func() {
		lookPath = originalLookPath
		runExternalCommand = originalRunExternal
	}()

	var called []string
	var cwd string
	lookPath = func(file string) (string, error) {
		if file != "npx" {
			t.Fatalf("unexpected lookup target: %s", file)
		}
		return "/usr/bin/npx", nil
	}
	runExternalCommand = func(runtime *Runtime, args []string) error {
		called = append([]string(nil), args...)
		cwd = runtime.Cwd
		return nil
	}

	exitCode, stdout, stderr := executeTestCommand(t, nil, nil, "skill", "install")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if strings.TrimSpace(stdout) != "Running: npx skills add itodca/marketo-cli" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if strings.Join(called, " ") != "npx skills add itodca/marketo-cli" {
		t.Fatalf("unexpected command: %#v", called)
	}
	if cwd == "" {
		t.Fatalf("expected project-local cwd to be set")
	}
}

func TestSkillUninstallSupportsGlobalScope(t *testing.T) {
	originalLookPath := lookPath
	originalRunExternal := runExternalCommand
	defer func() {
		lookPath = originalLookPath
		runExternalCommand = originalRunExternal
	}()

	var called []string
	lookPath = func(file string) (string, error) {
		return "/usr/bin/npx", nil
	}
	runExternalCommand = func(runtime *Runtime, args []string) error {
		called = append([]string(nil), args...)
		return nil
	}

	exitCode, stdout, stderr := executeTestCommand(t, nil, nil, "skill", "uninstall", "--global")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if strings.TrimSpace(stdout) != "Running: npx skills remove itodca/marketo-cli --global" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if strings.Join(called, " ") != "npx skills remove itodca/marketo-cli --global" {
		t.Fatalf("unexpected command: %#v", called)
	}
}

func TestSkillInstallReportsMissingNpx(t *testing.T) {
	originalLookPath := lookPath
	defer func() {
		lookPath = originalLookPath
	}()

	lookPath = func(file string) (string, error) {
		return "", errors.New("not found")
	}

	exitCode, stdout, stderr := executeTestCommand(t, nil, nil, "skill", "install")

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if stdout != "" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}
	if !strings.Contains(stderr, "npx not found. Install Node.js or run manually:") {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stderr, "npx skills add itodca/marketo-cli") {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}
