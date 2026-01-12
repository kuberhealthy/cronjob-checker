IMAGE := "kuberhealthy/cronjob-checker"
TAG := "latest"

# Build the cronjob checker container locally.
build:
	podman build -f Containerfile -t {{IMAGE}}:{{TAG}} .

# Run the unit tests for the cronjob checker.
test:
	go test ./...

# Build the cronjob checker binary locally.
binary:
	go build -o bin/cronjob-checker ./cmd/cronjob-checker
