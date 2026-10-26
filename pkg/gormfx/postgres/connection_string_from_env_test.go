package postgres

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func baseEnv(extra map[string]string) map[string]string {
	env := map[string]string{
		EnvPgsqlHost:     "db.internal",
		EnvPgsqlUser:     "app",
		EnvPgsqlPassword: "secret",
	}

	for key, value := range extra {
		env[key] = value
	}

	return env
}

func writeSelfSignedCA(t *testing.T) string {
	t.Helper()

	key, err := ecdsa.GenerateKey(
		elliptic.P256(),
		rand.Reader,
	)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test-ca",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}

	der, err := x509.CreateCertificate(
		rand.Reader,
		template,
		template,
		&key.PublicKey,
		key,
	)
	require.NoError(t, err)

	path := filepath.Join(
		t.TempDir(),
		"ca.pem",
	)

	err = os.WriteFile(
		path,
		pem.EncodeToMemory(
			&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: der,
			},
		),
		0o600,
	)
	require.NoError(t, err)

	return path
}

func TestConnectionStringFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		dbName   string
		env      map[string]string
		wantOK   bool
		wantErr  string
		wantHost string
		wantPort uint16
		wantDB   string
		wantUser string
		wantPass string
	}{
		{
			name:   "disabled without host",
			dbName: "orders",
			env: map[string]string{
				EnvPgsqlPort: "garbage",
			},
		},
		{
			name:   "disabled with blank host",
			dbName: "orders",
			env: map[string]string{
				EnvPgsqlHost: "  \t",
			},
		},
		{
			name:     "default port and db fallback",
			dbName:   "orders",
			env:      baseEnv(nil),
			wantOK:   true,
			wantHost: "db.internal",
			wantPort: 5432,
			wantDB:   "orders",
			wantUser: "app",
			wantPass: "secret",
		},
		{
			name:   "dbname env overrides argument",
			dbName: "orders",
			env: baseEnv(map[string]string{
				EnvPgsqlDBName: " ledger ",
			}),
			wantOK:   true,
			wantHost: "db.internal",
			wantPort: 5432,
			wantDB:   "ledger",
			wantUser: "app",
			wantPass: "secret",
		},
		{
			name:   "dbname from env without argument",
			dbName: "",
			env: baseEnv(map[string]string{
				EnvPgsqlDBName: "ledger",
			}),
			wantOK:   true,
			wantHost: "db.internal",
			wantPort: 5432,
			wantDB:   "ledger",
			wantUser: "app",
			wantPass: "secret",
		},
		{
			name:    "no dbname anywhere",
			dbName:  "",
			env:     baseEnv(nil),
			wantErr: EnvPgsqlDBName,
		},
		{
			name:   "explicit port",
			dbName: "orders",
			env: baseEnv(map[string]string{
				EnvPgsqlPort: " 6432 ",
			}),
			wantOK:   true,
			wantHost: "db.internal",
			wantPort: 6432,
			wantDB:   "orders",
			wantUser: "app",
			wantPass: "secret",
		},
		{
			name:   "ipv6 host",
			dbName: "orders",
			env: baseEnv(map[string]string{
				EnvPgsqlHost: "::1",
			}),
			wantOK:   true,
			wantHost: "::1",
			wantPort: 5432,
			wantDB:   "orders",
			wantUser: "app",
			wantPass: "secret",
		},
		{
			name:   "only username",
			dbName: "orders",
			env: map[string]string{
				EnvPgsqlHost:     "db.internal",
				EnvPgsqlUsername: "svc",
				EnvPgsqlPassword: "secret",
			},
			wantOK:   true,
			wantHost: "db.internal",
			wantPort: 5432,
			wantDB:   "orders",
			wantUser: "svc",
			wantPass: "secret",
		},
		{
			name:   "user and username equal",
			dbName: "orders",
			env: baseEnv(map[string]string{
				EnvPgsqlUsername: "app",
			}),
			wantOK:   true,
			wantHost: "db.internal",
			wantPort: 5432,
			wantDB:   "orders",
			wantUser: "app",
			wantPass: "secret",
		},
		{
			name:   "user and username differ",
			dbName: "orders",
			env: baseEnv(map[string]string{
				EnvPgsqlUsername: "other",
			}),
			wantErr: EnvPgsqlUsername,
		},
		{
			name:   "no user",
			dbName: "orders",
			env: map[string]string{
				EnvPgsqlHost:     "db.internal",
				EnvPgsqlPassword: "secret",
			},
			wantErr: EnvPgsqlUser,
		},
		{
			name:   "empty password",
			dbName: "orders",
			env: map[string]string{
				EnvPgsqlHost: "db.internal",
				EnvPgsqlUser: "app",
			},
			wantErr: EnvPgsqlPassword,
		},
		{
			name:   "password is not trimmed",
			dbName: "orders",
			env: baseEnv(map[string]string{
				EnvPgsqlPassword: "  pw  ",
			}),
			wantOK:   true,
			wantHost: "db.internal",
			wantPort: 5432,
			wantDB:   "orders",
			wantUser: "app",
			wantPass: "  pw  ",
		},
		{
			name:   "port not a number",
			dbName: "orders",
			env: baseEnv(map[string]string{
				EnvPgsqlPort: "54x2",
			}),
			wantErr: EnvPgsqlPort,
		},
		{
			name:   "port zero",
			dbName: "orders",
			env: baseEnv(map[string]string{
				EnvPgsqlPort: "0",
			}),
			wantErr: EnvPgsqlPort,
		},
		{
			name:   "port too large",
			dbName: "orders",
			env: baseEnv(map[string]string{
				EnvPgsqlPort: "65536",
			}),
			wantErr: EnvPgsqlPort,
		},
		{
			name:   "invalid sslmode",
			dbName: "orders",
			env: baseEnv(map[string]string{
				EnvPgsqlSSLMode: "verify",
			}),
			wantErr: EnvPgsqlSSLMode,
		},
		{
			name:   "special characters in password",
			dbName: "orders",
			env: baseEnv(map[string]string{
				EnvPgsqlPassword: `p@:/?#&=% '"\ x`,
			}),
			wantOK:   true,
			wantHost: "db.internal",
			wantPort: 5432,
			wantDB:   "orders",
			wantUser: "app",
			wantPass: `p@:/?#&=% '"\ x`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dsn, ok, err := connectionStringFromEnv(
				tc.dbName,
				envGetter(tc.env),
			)

			if tc.wantErr != "" {
				require.ErrorIs(t, err, ErrInvalidEnv)
				require.ErrorContains(t, err, tc.wantErr)
				require.False(t, ok)
				require.Empty(t, dsn)

				password := tc.env[EnvPgsqlPassword]
				if password != "" {
					require.NotContains(t, err.Error(), password)
				}

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantOK, ok)

			if !tc.wantOK {
				require.Empty(t, dsn)

				return
			}

			cfg, err := pgconn.ParseConfig(dsn)
			require.NoError(t, err)
			require.Equal(t, tc.wantHost, cfg.Host)
			require.Equal(t, tc.wantPort, cfg.Port)
			require.Equal(t, tc.wantDB, cfg.Database)
			require.Equal(t, tc.wantUser, cfg.User)
			require.Equal(t, tc.wantPass, cfg.Password)
		})
	}
}

func TestConnectionStringFromEnvErrorsHideValues(t *testing.T) {
	const password = "Sup3r-S3cret-Pw"

	cases := map[string]map[string]string{
		"port": {
			EnvPgsqlPort: "bad" + password,
		},
		"sslmode": {
			EnvPgsqlSSLMode: "bad" + password,
		},
		"username": {
			EnvPgsqlUsername: "bad" + password,
		},
	}

	for name, extra := range cases {
		t.Run(name, func(t *testing.T) {
			env := baseEnv(extra)
			env[EnvPgsqlPassword] = password

			_, _, err := connectionStringFromEnv(
				"orders",
				envGetter(env),
			)
			require.Error(t, err)
			require.NotContains(t, err.Error(), password)
		})
	}
}

func TestConnectionStringFromEnvSSL(t *testing.T) {
	caPath := writeSelfSignedCA(t)

	dsn, ok, err := connectionStringFromEnv(
		"orders",
		envGetter(
			baseEnv(map[string]string{
				EnvPgsqlSSLMode:     "verify-full",
				EnvPgsqlSSLRootCert: caPath,
			}),
		),
	)
	require.NoError(t, err)
	require.True(t, ok)

	cfg, err := pgconn.ParseConfig(dsn)
	require.NoError(t, err)
	require.NotNil(t, cfg.TLSConfig)
	require.Equal(t, "db.internal", cfg.TLSConfig.ServerName)
}

func TestConnectionStringFromEnvOmitsEmptySSL(t *testing.T) {
	dsn, ok, err := connectionStringFromEnv(
		"orders",
		envGetter(baseEnv(nil)),
	)
	require.NoError(t, err)
	require.True(t, ok)
	require.NotContains(t, dsn, "ssl")
}

func TestConnectionStringFromEnvPublic(t *testing.T) {
	t.Setenv(EnvPgsqlHost, "")

	dsn, ok, err := ConnectionStringFromEnv("orders")
	require.NoError(t, err)
	require.False(t, ok)
	require.Empty(t, dsn)
}
