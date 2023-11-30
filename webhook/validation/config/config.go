package config

type Config struct {
	Replication struct {
		Maximum int32 `koanf:"maximum"`
		Minimum int32 `koanf:"minimum"`
	} `koanf:"replication"`
}
