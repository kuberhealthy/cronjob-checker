package main

import (
	"os"
	"strconv"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// deadlineEnvVar mirrors the checkclient deadline environment variable.
	deadlineEnvVar = "KH_CHECK_RUN_DEADLINE"
)

// setEnv sets an environment variable for tests.
func setEnv(key string, value string) {
	// Apply the environment variable for the test.
	_ = os.Setenv(key, value)
}

// unsetEnv removes an environment variable for tests.
func unsetEnv(key string) {
	// Clear the environment variable for the test.
	_ = os.Unsetenv(key)
}

// TestParseConfigDefaultNamespace verifies default namespace handling.
func TestParseConfigDefaultNamespace(t *testing.T) {
	// Ensure environment variables are cleared.
	unsetEnv(namespaceEnvVar)
	unsetEnv(deadlineEnvVar)

	// Parse the config.
	cfg, err := parseConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify default namespace behavior.
	if cfg.Namespace != metav1.NamespaceAll {
		t.Fatalf("expected namespace %s, got %s", metav1.NamespaceAll, cfg.Namespace)
	}
}

// TestParseConfigCustomNamespace verifies custom namespace handling.
func TestParseConfigCustomNamespace(t *testing.T) {
	// Set the namespace environment variable.
	setEnv(namespaceEnvVar, "custom")
	unsetEnv(deadlineEnvVar)

	// Parse the config.
	cfg, err := parseConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Validate namespace override.
	if cfg.Namespace != "custom" {
		t.Fatalf("expected namespace custom, got %s", cfg.Namespace)
	}
}

// TestParseConfigDeadline verifies deadline handling uses half the remaining time.
func TestParseConfigDeadline(t *testing.T) {
	// Reset environment to avoid bleeding from other tests.
	unsetEnv(namespaceEnvVar)
	unsetEnv(deadlineEnvVar)

	// Set a deadline far enough in the future.
	deadline := time.Now().Add(20 * time.Minute).Unix()
	deadlineValue := strconv.FormatInt(deadline, 10)
	setEnv(deadlineEnvVar, deadlineValue)

	// Parse the config.
	cfg, err := parseConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Clean up the deadline variable after parsing.
	unsetEnv(deadlineEnvVar)

	// Verify the resulting time limit is roughly half of the deadline window.
	if cfg.CheckTimeLimit > 10*time.Minute {
		t.Fatalf("expected check time limit to be <= 10m, got %s", cfg.CheckTimeLimit)
	}
	if cfg.CheckTimeLimit < 9*time.Minute {
		t.Fatalf("expected check time limit to be >= 9m, got %s", cfg.CheckTimeLimit)
	}
}
