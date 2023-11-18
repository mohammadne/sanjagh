package config

type Config struct {
	MinReplication int32 `koanf:"min_replication"`
	MaxReplication int32 `koanf:"max_replication"`
}
