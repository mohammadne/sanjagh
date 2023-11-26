package controllers

import (
	"github.com/mohammadne/sanjagh/controllers/apps"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func Register(mgr manager.Manager, logger *zap.Logger) error {
	executerController := apps.NewExecuter(mgr.GetClient(), mgr.GetScheme(), logger)
	if err := executerController.SetupWithManager(mgr); err != nil {
		logger.Fatal("Unable to create Executer controller", zap.Error(err))
	}

	return nil
}
