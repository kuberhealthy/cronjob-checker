package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/kuberhealthy/kuberhealthy/v3/pkg/checkclient"
	nodecheck "github.com/kuberhealthy/kuberhealthy/v3/pkg/nodecheck"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// init enables debug output for the nodecheck package.
func init() {
	// Keep nodecheck logging behavior aligned with v2.
	nodecheck.EnableDebugOutput()
}

// main wires configuration and executes the cronjob checker.
func main() {
	// Parse configuration.
	cfg, err := parseConfig()
	if err != nil {
		reportFailureAndExit(err)
		return
	}

	// Create a deadline-aware context.
	ctx, cancel := context.WithTimeout(context.Background(), cfg.CheckTimeLimit)
	defer cancel()

	// Wait for the Kuberhealthy endpoint to be reachable.
	err = nodecheck.WaitForKuberhealthy(ctx)
	if err != nil {
		log.Errorln("Error waiting for kuberhealthy endpoint to be contactable by checker pod with error:", err.Error())
	}

	// Build a Kubernetes client.
	client, err := createKubeClient()
	if err != nil {
		reportFailureAndExit(err)
		return
	}

	// Evaluate cronjob schedules.
	err = checkCronJobs(ctx, client, cfg.Namespace)
	if err != nil {
		reportFailureAndExit(err)
		return
	}

	// Report success when no errors were detected.
	reportSuccess()
}

// checkCronJobs verifies cronjob schedules against their expected windows.
func checkCronJobs(ctx context.Context, client *kubernetes.Clientset, namespace string) error {
	// Fetch the list of cronjobs in the namespace.
	log.Infoln("Fetching cronjobs in namespace", namespace)
	cronList, err := client.BatchV1().CronJobs(namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to fetch cronjobs: %w", err)
	}

	log.Infoln("Found", len(cronList.Items), "cronjob(s) in namespace", namespace)

	// Track failures to determine overall check status.
	problemCount := 0
	goodCount := 0

	// Inspect each cronjob for schedule compliance.
	for _, cronJob := range cronList.Items {
		// Skip cronjobs that have never scheduled.
		if cronJob.Status.LastScheduleTime == nil {
			continue
		}

		// Fetch the latest cronjob resource for schedule and status data.
		cronGet, err := client.BatchV1().CronJobs(namespace).Get(ctx, cronJob.Name, v1.GetOptions{})
		if err != nil {
			log.Errorln("Error retrieving cronjob status for cronjob", cronJob.Name, "with error:", err)
			continue
		}

		// Compute the expected schedule window.
		schedule := cronGet.Spec.Schedule
		lastRunTime := cronGet.Status.LastScheduleTime.Time
		shouldRun := findLastCronRunTime(schedule)
		earliestRunTime, latestRunTime := scheduleWindow(shouldRun, scheduleWindowSize)

		log.Infoln("Cronjob", cronJob.Name, "was last scheduled at", lastRunTime)

		// Validate the last schedule time falls within the window.
		if lastRunTime.After(earliestRunTime) && lastRunTime.Before(latestRunTime) {
			log.Infoln("Cronjob", cronJob.Name, "is scheduling correctly")
			goodCount++
			continue
		}

		log.Infoln("Cronjob", cronJob.Name, "has not scheduled a job in scheduled window. Please confirm there are no issues with cronjob in namespace", cronJob.Namespace)
		problemCount++
	}

	// Report an error if any cronjobs are outside their schedule window.
	if problemCount != 0 {
		message := "There were " + strconv.Itoa(problemCount) + " cronjob(s) that had a last schedule time outside of scheduled window in namespace " + namespace
		return fmt.Errorf("%s", message)
	}

	log.Infoln("All cronjobs in namespace", namespace, "scheduled jobs in schedule window")
	log.Debugln("Cronjobs scheduling correctly:", goodCount)
	return nil
}

// reportFailureAndExit reports an error to Kuberhealthy and exits the process.
func reportFailureAndExit(err error) {
	// Report the failure to Kuberhealthy.
	reportErr := checkclient.ReportFailure([]string{err.Error()})
	if reportErr != nil {
		log.Fatalln("error when reporting to kuberhealthy:", reportErr.Error())
	}

	log.Infoln("Succesfully reported error to kuberhealthy")
	os.Exit(0)
}

// reportSuccess reports successful check results to Kuberhealthy.
func reportSuccess() {
	// Send success to Kuberhealthy.
	err := checkclient.ReportSuccess()
	if err != nil {
		log.Fatalln("error when reporting to kuberhealthy:", err.Error())
	}

	log.Infoln("Successfully reported to Kuberhealthy")
}
