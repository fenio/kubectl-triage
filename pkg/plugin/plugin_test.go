package plugin

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
)

// TestIsPodTrulyHealthy tests the health check logic
func TestIsPodTrulyHealthy(t *testing.T) {
	tests := []struct {
		name     string
		pod      *corev1.Pod
		expected bool
	}{
		{
			name: "Healthy pod - Running, Ready, No restarts",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						},
					},
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:         "app",
							RestartCount: 0,
							Ready:        true,
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "Unhealthy - Pod not Running",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						},
					},
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:         "app",
							RestartCount: 0,
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "Unhealthy - Pod not Ready",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionFalse,
						},
					},
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:         "app",
							RestartCount: 0,
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "Unhealthy - Has restarts (key test!)",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						},
					},
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:         "app",
							RestartCount: 5, // Has restarted!
							Ready:        true,
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "Multi-container - One has restarts",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						},
					},
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:         "app",
							RestartCount: 0,
							Ready:        true,
						},
						{
							Name:         "sidecar",
							RestartCount: 2, // Sidecar restarted
							Ready:        true,
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPodTrulyHealthy(tt.pod)
			if result != tt.expected {
				t.Errorf("isPodTrulyHealthy() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetReadyCount tests the ready count formatting
func TestGetReadyCount(t *testing.T) {
	tests := []struct {
		name     string
		pod      *corev1.Pod
		expected string
	}{
		{
			name: "All ready - 1/1",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{Name: "app", Ready: true},
					},
				},
			},
			expected: "1/1",
		},
		{
			name: "None ready - 0/2",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{Name: "app", Ready: false},
						{Name: "sidecar", Ready: false},
					},
				},
			},
			expected: "0/2",
		},
		{
			name: "Partial ready - 1/3",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{Name: "app", Ready: true},
						{Name: "sidecar1", Ready: false},
						{Name: "sidecar2", Ready: false},
					},
				},
			},
			expected: "1/3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getReadyCount(tt.pod)
			if result != tt.expected {
				t.Errorf("getReadyCount() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestIdentifyFailedContainers tests the container failure detection
func TestIdentifyFailedContainers(t *testing.T) {
	tests := []struct {
		name               string
		pod                *corev1.Pod
		allContainers      bool
		expectedFailedLen  int
		expectedHealthyLen int
		expectedFailedName string
	}{
		{
			name: "CrashLoopBackOff container",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:         "app",
							RestartCount: 5,
							State: corev1.ContainerState{
								Waiting: &corev1.ContainerStateWaiting{
									Reason: "CrashLoopBackOff",
								},
							},
						},
					},
				},
			},
			allContainers:      false,
			expectedFailedLen:  1,
			expectedHealthyLen: 0,
			expectedFailedName: "app",
		},
		{
			name: "OOMKilled container",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:         "worker",
							RestartCount: 2,
							State: corev1.ContainerState{
								Terminated: &corev1.ContainerStateTerminated{
									Reason: "OOMKilled",
								},
							},
						},
					},
				},
			},
			allContainers:      false,
			expectedFailedLen:  1,
			expectedHealthyLen: 0,
			expectedFailedName: "worker",
		},
		{
			name: "Running with restarts (golden indicator)",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:         "app",
							RestartCount: 3,
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{},
							},
						},
					},
				},
			},
			allContainers:      false,
			expectedFailedLen:  1, // Should be marked failed due to restarts
			expectedHealthyLen: 0,
			expectedFailedName: "app",
		},
		{
			name: "Multi-container - one failed, one healthy",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:         "app",
							RestartCount: 5,
							State: corev1.ContainerState{
								Waiting: &corev1.ContainerStateWaiting{
									Reason: "CrashLoopBackOff",
								},
							},
						},
						{
							Name:         "istio-proxy",
							RestartCount: 0,
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{},
							},
						},
					},
				},
			},
			allContainers:      false,
			expectedFailedLen:  1,
			expectedHealthyLen: 1,
			expectedFailedName: "app",
		},
		{
			name: "All containers healthy",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:         "app",
							RestartCount: 0,
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{},
							},
						},
						{
							Name:         "sidecar",
							RestartCount: 0,
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{},
							},
						},
					},
				},
			},
			allContainers:      false,
			expectedFailedLen:  0,
			expectedHealthyLen: 2,
		},
		{
			name: "All containers flag - shows all even if healthy",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:         "app",
							RestartCount: 0,
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{},
							},
						},
					},
				},
			},
			allContainers:      true, // Force show all
			expectedFailedLen:  1,    // Marked as "failed" list but not actually failed
			expectedHealthyLen: 0,
			expectedFailedName: "app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			failed, healthy := identifyFailedContainers(tt.pod, tt.allContainers)

			if len(failed) != tt.expectedFailedLen {
				t.Errorf("identifyFailedContainers() failed len = %v, want %v", len(failed), tt.expectedFailedLen)
			}

			if len(healthy) != tt.expectedHealthyLen {
				t.Errorf("identifyFailedContainers() healthy len = %v, want %v", len(healthy), tt.expectedHealthyLen)
			}

			if tt.expectedFailedLen > 0 && failed[0].Name != tt.expectedFailedName {
				t.Errorf("identifyFailedContainers() failed name = %v, want %v", failed[0].Name, tt.expectedFailedName)
			}
		})
	}
}

// TestFormatDuration tests the duration formatting
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "Seconds",
			duration: 30 * time.Second,
			expected: "30s",
		},
		{
			name:     "Minutes",
			duration: 5 * time.Minute,
			expected: "5m",
		},
		{
			name:     "Hours",
			duration: 3 * time.Hour,
			expected: "3h",
		},
		{
			name:     "Days",
			duration: 48 * time.Hour,
			expected: "2d",
		},
		{
			name:     "Edge - just under 1 minute",
			duration: 59 * time.Second,
			expected: "59s",
		},
		{
			name:     "Edge - exactly 1 minute",
			duration: 60 * time.Second,
			expected: "1m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestEventInfo tests event information structure
func TestEventInfoFiltering(t *testing.T) {
	// This is a conceptual test - in real implementation you'd mock the Kubernetes client
	// Here we just test the EventInfo struct can be properly constructed
	now := time.Now()

	event := EventInfo{
		Type:      corev1.EventTypeWarning,
		Reason:    "BackOff",
		Message:   "Back-off restarting failed container",
		Timestamp: now,
	}

	if event.Type != corev1.EventTypeWarning {
		t.Errorf("EventInfo Type = %v, want %v", event.Type, corev1.EventTypeWarning)
	}

	if event.Reason != "BackOff" {
		t.Errorf("EventInfo Reason = %v, want %v", event.Reason, "BackOff")
	}
}

// TestTriageOptions tests the options structure
func TestTriageOptions(t *testing.T) {
	opts := &TriageOptions{
		PodName:       "test-pod",
		Namespace:     "production",
		Lines:         50,
		AllContainers: false,
		Force:         false,
		NoColor:       false,
	}

	if opts.PodName != "test-pod" {
		t.Errorf("TriageOptions PodName = %v, want %v", opts.PodName, "test-pod")
	}

	if opts.Lines != 50 {
		t.Errorf("TriageOptions Lines = %v, want %v", opts.Lines, 50)
	}
}
