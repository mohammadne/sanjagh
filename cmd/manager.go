package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	appsv1alpha1 "github.com/mohammadne/sanjagh/api/v1alpha1"
	"github.com/mohammadne/sanjagh/config"
	"github.com/mohammadne/sanjagh/controllers"
	"github.com/mohammadne/sanjagh/pkg/k8s"
	"github.com/mohammadne/sanjagh/pkg/logger"
)

type Manager struct {
	config         *config.Config
	metricsPort    int
	probePort      int
	leaderElection bool
	kubeconfig     string
}

func NewManager(cfg *config.Config) *cobra.Command {
	manager := Manager{config: cfg}

	cmd := &cobra.Command{
		Use:   "manager",
		Short: "run controller-manager server",
		Run: func(_ *cobra.Command, _ []string) {
			manager.main()
		},
	}

	cmd.Flags().IntVar(&manager.metricsPort, "metrics-bind-address", 8080, "The port the metric endpoint binds to")
	cmd.Flags().IntVar(&manager.probePort, "health-probe-bind-address", 8081, "The port the probe endpoint binds to")
	cmd.Flags().StringVar(&manager.kubeconfig, "kubeconfig", "", "The kubeconfig file path")
	cmd.Flags().BoolVar(&manager.leaderElection, "leader-elect", false, "Enable leader election for controller manager. "+
		"Enabling this will ensure there is only one active controller manager.")

	return cmd
}

func (cmd *Manager) main() {
	logger := logger.NewZap(cmd.config.Logger)

	kubeConfig, err := k8s.KubeConfig(cmd.kubeconfig)
	if err != nil {
		logger.Fatal("Unable to create kubernetes configuration", zap.Error(err))
	}

	manager, err := ctrl.NewManager(kubeConfig, cmd.options())
	if err != nil {
		logger.Fatal("Unable to start manager", zap.Error(err))
	}

	if err := controllers.Register(manager, logger); err != nil {
		logger.Fatal("Unable to register controllers", zap.Error(err))
	}

	// TODO: add your custom metrics here

	if err := manager.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		logger.Fatal("Unable to set up health check", zap.Error(err))
	}
	if err := manager.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		logger.Fatal("Unable to set up ready check", zap.Error(err))
	}

	logger.Info("Starting manager")
	if err := manager.Start(ctrl.SetupSignalHandler()); err != nil {
		logger.Info("Problem running manager", zap.Error(err))
	}
}

func (cmd *Manager) options() ctrl.Options {
	var scheme = runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(appsv1alpha1.AddToScheme(scheme))

	return ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     fmt.Sprintf(":%d", cmd.metricsPort),
		HealthProbeBindAddress: fmt.Sprintf(":%d", cmd.probePort),
		LeaderElection:         cmd.leaderElection,
		LeaderElectionID:       "eca9d324.mohammadne.me",
	}
}
