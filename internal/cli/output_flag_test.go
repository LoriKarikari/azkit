package cli_test

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunner_outputJSONFlagMatchesJSONAlias(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newRunner(&stdout, &stderr, nil)

	if code := runner.Run(t.Context(), []string{"pim", "list", "-o", "json"}); code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"assignment_id": "a1"`) {
		t.Fatalf("-o json must emit the JSON contract, got:\n%s", stdout.String())
	}
}

func TestRunner_outputJSONAppliesToCtxCurrent(t *testing.T) {
	setupContextDirs(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newRunner(&stdout, &stderr, nil)

	if code := runner.Run(t.Context(), []string{"ctx", "current", "--output", "json"}); code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"context"`) {
		t.Fatalf("--output json must emit JSON, got:\n%s", stdout.String())
	}
}

func TestRunner_outputJSONRejectedOnCtxSwitch(t *testing.T) {
	setupContextDirs(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newRunner(&stdout, &stderr, nil)

	code := runner.Run(t.Context(), []string{"ctx", "prod", "-o", "json"})
	if code != 1 {
		t.Fatalf("want exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "ctx current -o json") {
		t.Fatalf("error must point at ctx current -o json, got: %s", stderr.String())
	}
}

func TestRunner_outputRejectsUnknownFormat(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newRunner(&stdout, &stderr, nil)

	if code := runner.Run(t.Context(), []string{"pim", "list", "-o", "yaml"}); code != 2 {
		t.Fatalf("want exit 2 for unknown format, got %d", code)
	}
}
