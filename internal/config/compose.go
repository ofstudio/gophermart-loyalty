package config

type CfgFunc func(*Config) (*Config, error)

// Compose - объединяет набор конфигурационных функций в одну
func Compose(fns ...CfgFunc) (*Config, error) {
	var err error
	cfg := &Config{}
	for _, fn := range fns {
		cfg, err = fn(cfg)
		if err != nil {
			return nil, err
		}
	}
	return cfg, nil
}
