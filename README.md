# Kuberhealthy Cronjob Checker

This check validates that Kubernetes CronJobs are scheduling on time. It compares the last schedule time for each CronJob against a calculated schedule window and reports failures when CronJobs fall outside that window.

## What It Does

1. Lists CronJobs in the configured namespace.
2. Computes the expected last run time from the CronJob schedule.
3. Validates the last schedule time is within a 10-minute window.
4. Reports failure if any CronJobs are outside the window.

## Configuration

All configuration is controlled via environment variables.

- `NAMESPACE`: Namespace to inspect. Default is all namespaces.

Kuberhealthy injects these variables automatically into the check pod:

- `KH_REPORTING_URL`
- `KH_RUN_UUID`
- `KH_CHECK_RUN_DEADLINE`

## Build

Use the `Justfile` to build or test the check:

```bash
just build
just test
```

## Example HealthCheck

```yaml
apiVersion: kuberhealthy.github.io/v2
kind: HealthCheck
metadata:
  name: cronjob-checker
  namespace: kuberhealthy
spec:
  runInterval: 5m
  timeout: 10m
  podSpec:
    spec:
      serviceAccountName: cronjob-checker
      containers:
        - name: cronjob-checker
          image: kuberhealthy/cronjob-checker:sha-<short-sha>
          imagePullPolicy: IfNotPresent
          env:
            - name: NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
          resources:
            requests:
              cpu: 15m
              memory: 15Mi
            limits:
              cpu: 25m
      restartPolicy: Never
      terminationGracePeriodSeconds: 5
```

A full install bundle with RBAC is available in `healthcheck.yaml`.

## Image Tags

- `sha-<short-sha>` tags are published on every push to `main`.
- `vX.Y.Z` tags are published when a matching Git tag is pushed.
