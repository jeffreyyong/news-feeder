package apppostgres

// PGConfig to connect to postgres database.
type PGConfig struct {
	dsn string
}

// PGOption is an option
type PGOption func(c *PGConfig)

// WithDSN allows the caller to specify the connection string or uri.
func WithDSN(dsn string) PGOption {
	return func(c *PGConfig) {
		c.dsn = dsn
	}
}
