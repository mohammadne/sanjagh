package k8s

import (
	"context"
	"errors"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func NewCachedClient(kubeConfig *rest.Config, indexer func(cache cache.Cache)) (crclient.Reader, error) {
	ctx := context.TODO()

	client, err := crclient.New(kubeConfig, crclient.Options{})
	if err != nil {
		return nil, err
	}

	cache, err := cache.New(kubeConfig, cache.Options{})
	if err != nil {
		return nil, err
	}

	if indexer != nil {
		indexer(cache)
	}

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
