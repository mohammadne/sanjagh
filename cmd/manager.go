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

	appsv1alpha1 "github.com/mohammadne/sanjagh/internal/api/v1alpha1"
	"github.com/mohammadne/sanjagh/internal/config"
	"github.com/mohammadne/sanjagh/internal/controllers"
	"github.com/mohammadne/sanjagh/pkg/logger"
)

type Manager struct{}

func (manager Manager) Command() *cobra.Command {
	config := config.Load(true)

	cmd := &cobra.Command{
		Use:   "manager",
		Short: "run controller-manager server",
		Run:   func(_ *cobra.Command, _ []string) { manager.main(config) },
	}

	cmd.Flags().IntVar(&config.MetricsPort, "metrics-bind-address", 8080, "The port the metric endpoint binds to")
	cmd.Flags().IntVar(&config.ProbePort, "health-probe-bind-address", 8081, "The port the probe endpoint binds to")
	cmd.Flags().BoolVar(&config.LeaderElection, "leader-elect", true, "Enable leader election for controller manager. "+
		"Enabling this will ensure there is only one active controller manager.")

	return cmd
}

func (*Manager) main(cfg *config.Config) {
	logger := logger.NewZap(cfg.Logger)

	var scheme = runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(appsv1alpha1.AddToScheme(scheme))

	manager, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     fmt.Sprintf(":%d", cfg.MetricsPort),
		HealthProbeBindAddress: fmt.Sprintf(":%d", cfg.ProbePort),
		LeaderElection:         cfg.LeaderElection,
		LeaderElectionID:       "eca9d324.mohammadne.me",
	})

	if err != nil {
		logger.Fatal("Unable to start manager", zap.Error(err))
	}

	executerController := controllers.NewExecuter(manager.GetClient(), manager.GetScheme(), logger)
	if err := executerController.SetupWithManager(manager); err != nil {
		logger.Fatal("Unable to create Executer controller", zap.Error(err))
	}

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
