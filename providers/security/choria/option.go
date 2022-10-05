// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package choria

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"

	"github.com/choria-io/go-choria/config"
	"github.com/choria-io/go-choria/inter"
	"github.com/choria-io/go-choria/tlssetup"
	"github.com/sirupsen/logrus"
)

// Option is a function that can configure the Security Provider
type Option func(*ChoriaSecurity) error

// BuildInfoProvider provides info about the build
type BuildInfoProvider interface {
	ClientIdentitySuffix() string
}

// WithChoriaConfig optionally configures the Security Provider from settings found in a typical Choria configuration
func WithChoriaConfig(c *config.Config) Option {
	return func(s *ChoriaSecurity) error {
		cfg := Config{
			TLSConfig:             tlssetup.TLSConfig(c),
			RemoteSignerURL:       c.Choria.RemoteSignerURL,
			RemoteSignerTokenFile: c.Choria.RemoteSignerTokenFile,
		}

		for _, signer := range c.Choria.ChoriaSecurityTrustedSigners {
			pk, err := hex.DecodeString(signer)
			if err != nil {
				return fmt.Errorf("invalid ed25519 public key: %v: %v", signer, err)
			}
			if len(pk) != ed25519.PublicKeySize {
				return fmt.Errorf("invalid ed25519 public key size: %v: %v", signer, len(pk))
			}

			cfg.TrustedTokenSigners = append(cfg.TrustedTokenSigners, pk)
		}

		if c.Choria.ChoriaSecurityIdentity != "" {
			cfg.Identity = c.Choria.ChoriaSecurityIdentity
		} else if !c.InitiatedByServer {
			userEnvVar := "USER"
			if runtime.GOOS == "windows" {
				userEnvVar = "USERNAME"
			}

			u, ok := os.LookupEnv(userEnvVar)
			if !ok {
				return fmt.Errorf("could not determine client identity, ensure %s environment variable is set", userEnvVar)
			}

			cfg.Identity = u
		}

		s.conf = &cfg

		return nil
	}
}

// WithTokenFile sets the path to the JWT token stored in a file
func WithTokenFile(f string) Option {
	return func(s *ChoriaSecurity) error {
		s.conf.TokenFile = f
		return nil
	}
}

// WithSeedFile sets the path to the ed25519 seed stored in a file
func WithSeedFile(f string) Option {
	return func(s *ChoriaSecurity) error {
		s.conf.SeedFile = f
		return nil
	}
}

// WithSigner configures a remote request signer
func WithSigner(signer inter.RequestSigner) Option {
	return func(s *ChoriaSecurity) error {
		s.conf.RemoteSigner = signer

		return nil
	}
}

// WithConfig optionally configures the Security Provider using its native configuration format
func WithConfig(c *Config) Option {
	return func(s *ChoriaSecurity) error {
		s.conf = c

		if s.conf.TLSConfig == nil {
			s.conf.TLSConfig = tlssetup.TLSConfig(nil)
		}

		return nil
	}
}

// WithLog configures a logger for the Security Provider
func WithLog(l *logrus.Entry) Option {
	return func(s *ChoriaSecurity) error {
		s.log = l.WithFields(logrus.Fields{"security": "choria"})

		if s.conf.TLSConfig == nil {
			s.conf.TLSConfig = tlssetup.TLSConfig(nil)
		}

		return nil
	}
}
