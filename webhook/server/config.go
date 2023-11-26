package server

import "fmt"

type Config struct {
	ListenAddr string `mapstructure:"listenAddr"`
	TLSCert    string `mapstructure:"tlsCert"`
	TLSKey     string `mapstructure:"tlsKey"`
	Path       string `mapstructure:"path"`
}

func (c *Config) Validate() error {
	if c.ListenAddr == "" {
		return fmt.Errorf("serverConfig.listenAddr is empty")
	}
	if c.TLSCert == "" || c.TLSKey == "" {
		return fmt.Errorf("serverConfig.tlsCert or serverConfig.tlsKey is empty")
	}
	if c.Path == "" {
		return fmt.Errorf("serverConfig.path is empty")
	}
	return nil
}
