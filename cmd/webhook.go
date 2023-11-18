package cmd

import (
	"context"
	"errors"
	"os"

	"github.com/mohammadne/sanjagh/internal/config"
	"github.com/mohammadne/sanjagh/internal/webhook/server"
	"github.com/mohammadne/sanjagh/internal/webhook/validation"
	"github.com/spf13/cobra"
	"gorm.io/gorm/logger"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

type Webhook struct{}

func (manager Webhook) Command() *cobra.Command {
	config := config.Load(true)

	return &cobra.Command{
		Use:   "webhook",
		Short: "run webhook server",
		Run:   func(_ *cobra.Command, _ []string) { manager.main(config) },
	}
}

func (*Webhook) main(cfg *config.Config) {
	// initialize kubernetes client
	var conf *rest.Config
	kubeconfigFile := projectconfig.KubeConfig()
	if kubeconfigFile != "" {
		conf, err = clientcmd.BuildConfigFromFlags("", kubeconfigFile)
		if err != nil {
			setupLog.Error(err, "unable to create kubernetes configuration")
			os.Exit(1)
		}
	} else {
		conf = ctrl.GetConfigOrDie()
	}

	client, err := NewCachedClient(conf)
	if err != nil {
		setupLog.Error(err, "couldn't create cached client")
		os.Exit(1)
	}

	validation := validation.NewValidation(webhookConfig.Validation, client)

	s, err := server.New(webhookConfig.Server, validation)
	if err != nil {
		logger.Error(err, "could not create new server")
		os.Exit(1)
	}
	if err := s.Run(); err != nil {
		logger.Error(err, "could not run server")
		os.Exit(1)
	}
}

func NewCachedClient(restConfig *rest.Config) (crclient.Reader, error) {
	ctx := context.TODO()

	client, err := crclient.New(restConfig, crclient.Options{})
	if err != nil {
		return nil, err
	}

	cache, err := cache.New(restConfig, cache.Options{})
	if err != nil {
		return nil, err
	}

	// add indexers here

	go cache.Start(ctx)

	cachedClient, err := crclient.NewDelegatingClient(crclient.NewDelegatingClientInput{CacheReader: cache, Client: client})
	if err != nil {
		return nil, err
	}

	if successful := cache.WaitForCacheSync(ctx); !successful {
		return nil, errors.New("could not sync cache")
	}

	return cachedClient, nil
}
