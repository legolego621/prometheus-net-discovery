package config

import (
	"fmt"
	"os"
	"prometheus-net-discovery/internal/netops/scanner"

	validator "github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Global   *GlobalConfig      `yaml:"global"   validate:"required"`
	Scanners []*scanner.Scanner `yaml:"scanners" validate:"required,required_ping_or_ports,dive"`
}

type GlobalConfig struct {
	InstanceID string `yaml:"instanceId" validate:"required"`
}

func New() *Config {
	return &Config{}
}

func (c *Config) Load(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(file, c); err != nil {
		return err
	}

	if err := c.validate(); err != nil {
		return fmt.Errorf("failed to validate config: %w", err)
	}

	return nil
}

func (c *Config) validate() error {
	validate := validator.New()

	if err := validate.RegisterValidation("required_ping_or_ports", validateDiscovery); err != nil {
		return err
	}

	return validate.Struct(c)
}

func validateDiscovery(fl validator.FieldLevel) bool {
	scanners, ok := fl.Field().Interface().([]*scanner.Scanner)
	if !ok {
		return false
	}

	for _, scan := range scanners {
		if !scan.Ping && len(scan.Ports) == 0 {
			return false
		}
	}

	return true
}
