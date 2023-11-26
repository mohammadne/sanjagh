package server

import "fmt"

type Config struct {
	TLSCert string `mapstructure:"tlsCert"`
	TLSKey  string `mapstructure:"tlsKey"`
}

func (c *Config) Validate() error {
	if c.TLSCert == "" || c.TLSKey == "" {
		return fmt.Errorf("serverConfig.tlsCert or serverConfig.tlsKey is empty")
	}
	return nil
}
