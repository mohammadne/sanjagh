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
	"github.com/mohammadne/sanjagh/pkg/logger"
)

type ManagerCommand struct {
	config         *config.Config
	metricsPort    int
	probePort      int
	leaderElection bool
}

func NewManagerCommand(cfg *config.Config, metricsPort int, probePort int) *cobra.Command {
	managerCommand := ManagerCommand{
		config:      cfg,
		metricsPort: metricsPort,
		probePort:   probePort,
	}

	cmd := &cobra.Command{
		Use:   "manager",
		Short: "run controller-manager server",
		Run: func(_ *cobra.Command, _ []string) {
			managerCommand.main()
		},
	}

	cmd.Flags().BoolVar(&managerCommand.leaderElection, "leader-elect", true, "Enable leader election for controller manager. "+
		"Enabling this will ensure there is only one active controller manager.")

	return cmd
}

func (cmd *ManagerCommand) main() {
	logger := logger.NewZap(cmd.config.Logger)

	controllerManager, err := ctrl.NewManager(ctrl.GetConfigOrDie(), cmd.options())
	if err != nil {
		logger.Fatal("Unable to start manager", zap.Error(err))
	}

	if err := controllers.Register(controllerManager, logger); err != nil {
		logger.Fatal("Unable to register controllers", zap.Error(err))
	}

	if err := controllerManager.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		logger.Fatal("Unable to set up health check", zap.Error(err))
	}
	if err := controllerManager.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		logger.Fatal("Unable to set up ready check", zap.Error(err))
	}

	logger.Info("Starting manager")
	if err := controllerManager.Start(ctrl.SetupSignalHandler()); err != nil {
		logger.Info("Problem running manager", zap.Error(err))
	}
}

func (cmd *ManagerCommand) options() ctrl.Options {
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
