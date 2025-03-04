package config

import (
	"fmt"
	"github.com/habiliai/apidepot/pkg/internal/constants"
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"github.com/pkg/errors"
	"time"
)

type (
	SMTPConfig struct {
		Host       string
		Port       int
		Username   string
		Password   string
		AdminEmail string
	}

	RegionalStackConfig struct {
		Scheme string
		Domain string
	}

	StackConfig struct {
		ForceDelete     bool
		SkipHealthCheck bool
		Seoul           RegionalStackConfig
		Singapore       RegionalStackConfig
	}

	RegionalS3Config struct {
		Endpoint string
	}

	S3Config struct {
		AccessKey string
		SecretKey string
		Seoul     RegionalS3Config
		Singapore RegionalS3Config
	}

	RegionalDBConfig struct {
		Host     string
		Port     int
		User     string
		Name     string
		Password string
	}

	DBConfig struct {
		PingTimeout     string
		AutoMigration   bool
		MaxIdleConns    int
		MaxOpenConns    int
		ConnMaxLifetime string

		Seoul     RegionalDBConfig
		Singapore RegionalDBConfig
	}

	RegionalKubernetesConfig struct {
		Context string
	}

	KubernetesConfig struct {
		KubeConfig string
		QPS        float32
		Burst      int

		Seoul     RegionalKubernetesConfig
		Singapore RegionalKubernetesConfig
	}

	ApiDepotServerConfig struct {
		Address      string
		Port         int
		WebPort      int
		Secure       bool
		IncludeDebug bool
		LogLevel     string

		Scan struct {
			Timeout time.Duration
		}

		DB   DBConfig
		Stoa struct {
			URL      string
			AnonKey  string
			AdminKey string
		}
		Github struct {
			AppId         string
			ClientId      string
			ClientSecret  string
			AppPrivateKey string
		}

		K8s KubernetesConfig

		Stack StackConfig
		SMTP  SMTPConfig
		S3    S3Config
	}
)

func (c *ApiDepotServerConfig) String() string {
	return fmt.Sprintf(
		"Address: %s, Port: %d, WebPort: %d, Secure: %t, IncludeDebug: %t, LogLevel: %s, DB: %v, Stoa: %v",
		c.Address, c.Port, c.WebPort, c.Secure, c.IncludeDebug, c.LogLevel, c.DB, c.Stoa,
	)
}

func (c RegionalDBConfig) GetURI() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable&search_path=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, constants.ApiDepotSchema,
	)
}

func (c RegionalDBConfig) WithDBName(dbName string) RegionalDBConfig {
	c.Name = dbName
	return c
}

func (c *ApiDepotServerConfig) Validate() error {
	if c.S3.SecretKey == "" {
		return errors.New("s3.secretKey is required")
	}

	if c.S3.AccessKey == "" {
		return errors.New("s3.accessKey is required")
	}

	return nil
}

func (c S3Config) GetRegionalConfig(zone tcltypes.InstanceZone) RegionalS3Config {
	switch zone {
	case tcltypes.InstanceZoneOciApSeoul:
		return c.Seoul
	case tcltypes.InstanceZoneOciSingapore:
		return c.Singapore
	default:
		panic(fmt.Sprintf("invalid zone: %s", zone))
	}
}

func (c StackConfig) GetRegionalConfig(zone tcltypes.InstanceZone) RegionalStackConfig {
	switch zone {
	case tcltypes.InstanceZoneOciApSeoul:
		return c.Seoul
	case tcltypes.InstanceZoneOciSingapore:
		return c.Singapore
	default:
		panic(fmt.Sprintf("invalid zone: %s", zone))
	}
}

func (c KubernetesConfig) GetRegionalConfig(zone tcltypes.InstanceZone) RegionalKubernetesConfig {
	switch zone {
	case tcltypes.InstanceZoneOciApSeoul:
		return c.Seoul
	case tcltypes.InstanceZoneOciSingapore:
		return c.Singapore
	default:
		panic(fmt.Sprintf("invalid zone: %s", zone))
	}
}

func (c DBConfig) GetRegionalConfig(zone tcltypes.InstanceZone) RegionalDBConfig {
	switch zone {
	case tcltypes.InstanceZoneOciApSeoul:
		return c.Seoul
	case tcltypes.InstanceZoneOciSingapore:
		return c.Singapore
	default:
		panic(fmt.Sprintf("invalid zone: %s", zone))
	}
}
