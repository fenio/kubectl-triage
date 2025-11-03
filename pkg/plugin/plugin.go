package plugin

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/lichenglin/kubectl-triage/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

// TriageOptions holds configuration for the triage command
type TriageOptions struct {
	PodName       string
	Namespace     string
	Lines         int64
	AllContainers bool
	Force         bool
	NoColor       bool
}

// ContainerInfo holds information about a container's state
type ContainerInfo struct {
	Name         string
	State        string
	Reason       string
	RestartCount int32
	Failed       bool
}

// LogResult holds log output for a container
type LogResult struct {
	ContainerName string
	Previous      string
	Current       string
	PreviousError error
	CurrentError  error
}

// EventInfo holds simplified event information
type EventInfo struct {
	Type      string
	Reason    string
	Message   string
	Timestamp time.Time
}

// RunPlugin is the main entry point for the triage command
func RunPlugin(configFlags *genericclioptions.ConfigFlags, opts *TriageOptions) error {
	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return fmt.Errorf("failed to read kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	// Determine namespace
	namespace := opts.Namespace
	if namespace == "" {
		if configFlags.Namespace != nil && *configFlags.Namespace != "" {
			namespace = *configFlags.Namespace
		} else {
			namespace = "default"
		}
	}

	// Fetch the pod
	pod, err := clientset.CoreV1().Pods(namespace).Get(opts.PodName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get pod %s in namespace %s: %w", opts.PodName, namespace, err)
	}

	// Check if pod is truly healthy
	if !opts.Force && isPodTrulyHealthy(pod) {
		log := logger.NewLogger()
		readyCount := getReadyCount(pod)
		log.Info(fmt.Sprintf("✅ Pod '%s' is healthy (Ready %s, 0 restarts).", pod.Name, readyCount))
		log.Info("Use --force to inspect anyway.")
		return nil
	}

	// Identify failed containers
	failedContainers, healthyContainers := identifyFailedContainers(pod, opts.AllContainers)

	// Get relevant events (Warning/Error only)
	events, err := getRelevantEvents(clientset, namespace, pod.Name)
	if err != nil {
		return fmt.Errorf("failed to get events: %w", err)
	}

	// Collect logs for failed containers in parallel
	logResults := collectLogs(clientset, namespace, pod.Name, failedContainers, opts.Lines)

	// Display the triage output
	displayTriage(pod, failedContainers, healthyContainers, events, logResults, opts)

	return nil
}

// isPodTrulyHealthy checks if a pod is actually healthy
// A pod is considered healthy if:
// 1. Phase is Running
// 2. All containers are Ready
// 3. RestartCount is 0 for all containers
func isPodTrulyHealthy(pod *corev1.Pod) bool {
	// Check 1: Phase must be Running
	if pod.Status.Phase != corev1.PodRunning {
		return false
	}

	// Check 2: All containers must be Ready
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			if condition.Status != corev1.ConditionTrue {
				return false
			}
		}
	}

	// Check 3: RestartCount must be 0 for all containers
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.RestartCount > 0 {
			return false
		}
	}

	return true
}

// getReadyCount returns the ready count in format "1/1", "0/2", etc.
func getReadyCount(pod *corev1.Pod) string {
	total := len(pod.Status.ContainerStatuses)
	ready := 0
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Ready {
			ready++
		}
	}
	return fmt.Sprintf("%d/%d", ready, total)
}

// identifyFailedContainers analyzes container statuses and categorizes them
func identifyFailedContainers(pod *corev1.Pod, allContainers bool) ([]ContainerInfo, []ContainerInfo) {
	var failed []ContainerInfo
	var healthy []ContainerInfo

	for _, cs := range pod.Status.ContainerStatuses {
		info := ContainerInfo{
			Name:         cs.Name,
			RestartCount: cs.RestartCount,
			Failed:       false,
		}

		// Check if container has restarted (golden indicator)
		if cs.RestartCount > 0 {
			info.Failed = true
		}

		// Check current state
		if cs.State.Waiting != nil {
			info.State = "Waiting"
			info.Reason = cs.State.Waiting.Reason
			// Check for failure reasons
			failureReasons := []string{
				"CrashLoopBackOff", "Error", "ImagePullBackOff", "ErrImagePull",
				"CreateContainerError", "InvalidImageName",
			}
			for _, reason := range failureReasons {
				if cs.State.Waiting.Reason == reason {
					info.Failed = true
					break
				}
			}
		} else if cs.State.Terminated != nil {
			info.State = "Terminated"
			info.Reason = cs.State.Terminated.Reason
			// Check for failure reasons
			failureReasons := []string{
				"Error", "OOMKilled", "ContainerCannotRun", "DeadlineExceeded",
			}
			for _, reason := range failureReasons {
				if cs.State.Terminated.Reason == reason {
					info.Failed = true
					break
				}
			}
		} else if cs.State.Running != nil {
			info.State = "Running"
			info.Reason = ""
		}

		if info.Failed || allContainers {
			failed = append(failed, info)
		} else {
			healthy = append(healthy, info)
		}
	}

	return failed, healthy
}

// getRelevantEvents fetches only Warning and Error events for the pod
func getRelevantEvents(clientset *kubernetes.Clientset, namespace, podName string) ([]EventInfo, error) {
	eventList, err := clientset.CoreV1().Events(namespace).List(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s", podName),
	})
	if err != nil {
		return nil, err
	}

	var relevantEvents []EventInfo
	for _, event := range eventList.Items {
		// Only include Warning and Error type events (filter out Normal events)
		if event.Type == corev1.EventTypeWarning || event.Type == "Error" {
			relevantEvents = append(relevantEvents, EventInfo{
				Type:      event.Type,
				Reason:    event.Reason,
				Message:   event.Message,
				Timestamp: event.LastTimestamp.Time,
			})
		}
	}

	// Sort by timestamp (most recent first)
	// Simple bubble sort for small lists
	for i := 0; i < len(relevantEvents)-1; i++ {
		for j := i + 1; j < len(relevantEvents); j++ {
			if relevantEvents[i].Timestamp.Before(relevantEvents[j].Timestamp) {
				relevantEvents[i], relevantEvents[j] = relevantEvents[j], relevantEvents[i]
			}
		}
	}

	return relevantEvents, nil
}

// collectLogs fetches logs for failed containers in parallel
func collectLogs(clientset *kubernetes.Clientset, namespace, podName string, containers []ContainerInfo, tailLines int64) []LogResult {
	var wg sync.WaitGroup
	results := make([]LogResult, len(containers))

	for i, container := range containers {
		wg.Add(1)
		go func(idx int, containerName string) {
			defer wg.Done()

			result := LogResult{ContainerName: containerName}

			// Fetch previous logs
			prevLogOpts := &corev1.PodLogOptions{
				Container: containerName,
				Previous:  true,
				TailLines: &tailLines,
			}
			result.Previous, result.PreviousError = fetchLog(clientset, namespace, podName, prevLogOpts)

			// Fetch current logs
			currLogOpts := &corev1.PodLogOptions{
				Container: containerName,
				TailLines: &tailLines,
			}
			result.Current, result.CurrentError = fetchLog(clientset, namespace, podName, currLogOpts)

			results[idx] = result
		}(i, container.Name)
	}

	wg.Wait()
	return results
}

// fetchLog retrieves log content from a pod
func fetchLog(clientset *kubernetes.Clientset, namespace, podName string, opts *corev1.PodLogOptions) (string, error) {
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, opts)
	podLogs, err := req.Stream()
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// displayTriage outputs the formatted triage information
func displayTriage(pod *corev1.Pod, failedContainers, healthyContainers []ContainerInfo, events []EventInfo, logResults []LogResult, opts *TriageOptions) {
	log := logger.NewLogger()

	// Display each failed container
	for i, container := range failedContainers {
		log.Info("")
		log.Info(strings.Repeat("=", 80))
		if container.Failed {
			log.ErrorMsg(fmt.Sprintf("🚨 TRIAGE FOR FAILED CONTAINER: '%s' (Reason: %s)", container.Name, container.Reason))
		} else {
			log.Info(fmt.Sprintf("🔍 TRIAGE FOR CONTAINER: '%s'", container.Name))
		}
		log.Info(strings.Repeat("=", 80))
		log.Info("")

		// Pod Status Section
		log.Info("📋 POD STATUS")
		log.Info(fmt.Sprintf("  Phase: %s | Restarts: %d | Ready: %s",
			pod.Status.Phase,
			container.RestartCount,
			getReadyCount(pod)))
		log.Info("")

		// Events Section (only show once for first container)
		if i == 0 && len(events) > 0 {
			log.Info("⚠️  CRITICAL EVENTS (Warning/Error only - Last 10)")
			count := 0
			for _, event := range events {
				if count >= 10 {
					break
				}
				ago := time.Since(event.Timestamp).Round(time.Second)
				eventLine := fmt.Sprintf("  %s ago | %-7s | %-15s | %s",
					formatDuration(ago),
					event.Type,
					event.Reason,
					event.Message)

				if event.Type == "Error" || strings.Contains(event.Reason, "Failed") {
					log.ErrorMsg(eventLine)
				} else {
					log.Info(eventLine)
				}
				count++
			}
			log.Info("")
		}

		// Logs Section
		if i < len(logResults) {
			logResult := logResults[i]

			// Previous logs
			if logResult.PreviousError == nil && logResult.Previous != "" {
				log.Info("🔥 PREVIOUS LOGS (Last Crash) - Last 50 lines")
				log.Info(strings.Repeat("-", 80))
				printHighlightedLogs(logResult.Previous, opts.NoColor)
				log.Info(strings.Repeat("-", 80))
				log.Info("")
			} else if logResult.PreviousError != nil && !strings.Contains(logResult.PreviousError.Error(), "previous terminated container") {
				log.Info("🔥 PREVIOUS LOGS")
				log.Info(fmt.Sprintf("  (No previous logs: %v)", logResult.PreviousError))
				log.Info("")
			}

			// Current logs
			if logResult.CurrentError == nil && logResult.Current != "" {
				log.Info("🔄 CURRENT LOGS - Last 50 lines")
				log.Info(strings.Repeat("-", 80))
				printHighlightedLogs(logResult.Current, opts.NoColor)
				log.Info(strings.Repeat("-", 80))
				log.Info("")
			} else if logResult.CurrentError != nil {
				log.Info("🔄 CURRENT LOGS")
				log.Info(fmt.Sprintf("  (No current logs: %v)", logResult.CurrentError))
				log.Info("")
			}
		}
	}

	// Healthy containers summary
	if len(healthyContainers) > 0 {
		log.Info(strings.Repeat("-", 80))
		var healthyNames []string
		for _, c := range healthyContainers {
			healthyNames = append(healthyNames, fmt.Sprintf("%s (%s, %d restarts)", c.Name, c.State, c.RestartCount))
		}
		log.Info(fmt.Sprintf("ℹ️  %d other container(s) running normally: [%s]",
			len(healthyContainers),
			strings.Join(healthyNames, ", ")))
		log.Info("")
	}
}

// printHighlightedLogs outputs logs with keyword highlighting
func printHighlightedLogs(logs string, noColor bool) {
	lines := strings.Split(logs, "\n")
	keywords := []string{"ERROR", "error", "Error", "panic", "PANIC", "Panic",
		"fatal", "FATAL", "Fatal", "exception", "Exception", "EXCEPTION",
		"failed", "Failed", "FAILED", "killed", "Killed", "KILLED", "OOMKilled"}

	log := logger.NewLogger()
	for _, line := range lines {
		if line == "" {
			continue
		}

		shouldHighlight := false
		if !noColor {
			for _, keyword := range keywords {
				if strings.Contains(line, keyword) {
					shouldHighlight = true
					break
				}
			}
		}

		if shouldHighlight {
			log.ErrorMsg("  " + line)
		} else {
			fmt.Println("  " + line)
		}
	}
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	} else {
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}
