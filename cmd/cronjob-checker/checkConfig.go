package main

import (
	"os"
	"time"

	"github.com/kuberhealthy/kuberhealthy/v3/pkg/checkclient"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// namespaceEnvVar is the environment variable for scoping cronjobs.
	namespaceEnvVar = "NAMESPACE"

	// defaultCheckTimeLimit is used when KH_CHECK_RUN_DEADLINE is unavailable.
	defaultCheckTimeLimit = time.Minute * 5
	// deadlinePadding keeps a small buffer before the Kuberhealthy deadline.
	deadlinePadding = time.Second * 5
)

// CheckConfig stores configuration for the cronjob checker.
type CheckConfig struct {
	// Namespace scopes cronjob queries. Use NamespaceAll for all namespaces.
	Namespace string
	// CheckTimeLimit sets the timeout for the run.
	CheckTimeLimit time.Duration
}

// parseConfig loads environment variables into a CheckConfig.
func parseConfig() (*CheckConfig, error) {
	// Start with defaults.
	cfg := &CheckConfig{}
	cfg.Namespace = metav1.NamespaceAll
	cfg.CheckTimeLimit = defaultCheckTimeLimit

	// Read namespace configuration.
	namespaceEnv := os.Getenv(namespaceEnvVar)
	if len(namespaceEnv) != 0 {
		cfg.Namespace = namespaceEnv
	}

	// Determine the deadline from Kuberhealthy when available.
	deadline, err := checkclient.GetDeadline()
	if err == nil {
		remaining := deadline.Sub(time.Now().Add(deadlinePadding))
		if remaining > 0 {
			cfg.CheckTimeLimit = remaining / 2
		}
	}
	if err != nil {
		log.Debugln("Using default check time limit.")
	}

	return cfg, nil
}
