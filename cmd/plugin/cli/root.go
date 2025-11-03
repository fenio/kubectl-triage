package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/lichenglin/kubectl-triage/pkg/plugin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	KubernetesConfigFlags *genericclioptions.ConfigFlags
)

func RootCmd() *cobra.Command {
	var (
		lines         int64
		allContainers bool
		force         bool
		noColor       bool
	)

	cmd := &cobra.Command{
		Use:   "kubectl-triage <pod-name>",
		Short: "Fast triage for failed Kubernetes pods",
		Long: `kubectl-triage provides a 5-second diagnostic snapshot for failed Kubernetes pods.

It intelligently aggregates:
  - Pod status and container states
  - Critical events (Warning/Error only - filters out noise)
  - Previous crash logs (if container restarted)
  - Current container logs

Only failed/restarted containers are shown by default, keeping output focused.`,
		Example: `  # Triage a crashing pod
  kubectl triage my-failing-pod

  # Triage a pod in a specific namespace
  kubectl triage my-pod -n production

  # Show all containers, not just failed ones
  kubectl triage my-pod --all-containers

  # Inspect a healthy pod anyway
  kubectl triage my-pod --force

  # Show more log lines
  kubectl triage my-pod --lines=100`,
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ExactArgs(1), // Require exactly one argument (pod name)
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get pod name from arguments
			podName := args[0]

			// Get namespace from kubectl flags
			namespace := ""
			if KubernetesConfigFlags.Namespace != nil && *KubernetesConfigFlags.Namespace != "" {
				namespace = *KubernetesConfigFlags.Namespace
			}

			// Build triage options
			opts := &plugin.TriageOptions{
				PodName:       podName,
				Namespace:     namespace,
				Lines:         lines,
				AllContainers: allContainers,
				Force:         force,
				NoColor:       noColor,
			}

			// Run the triage
			if err := plugin.RunPlugin(KubernetesConfigFlags, opts); err != nil {
				return errors.Unwrap(err)
			}

			return nil
		},
	}

	cobra.OnInitialize(initConfig)

	// Add kubectl flags (--namespace, --context, --kubeconfig, etc.)
	KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)
	KubernetesConfigFlags.AddFlags(cmd.Flags())

	// Add custom flags
	cmd.Flags().Int64Var(&lines, "lines", 50, "Number of log lines to display (default: 50)")
	cmd.Flags().BoolVar(&allContainers, "all-containers", false, "Show all containers, not just failed/restarted ones")
	cmd.Flags().BoolVar(&force, "force", false, "Inspect pod even if it appears healthy")
	cmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	return cmd
}

func InitAndExecute() {
	if err := RootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func initConfig() {
	viper.AutomaticEnv()
}
