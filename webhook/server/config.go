package server

import "fmt"

type Config struct {
	TLS struct {
		Certificate string `koanf:"certificate"`
		PrivateKey  string `koanf:"private_key"`
	} `koanf:"tls"`
}

func (c *Config) Validate() error {
	if c.TLS.Certificate == "" || c.TLS.PrivateKey == "" {
		return fmt.Errorf("TLS Certificate or PrivateKey is empty")
	}
	return nil
}
