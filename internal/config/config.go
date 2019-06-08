package config

import (
	"errors"
	"fmt"

	"github.com/spf13/pflag"
	"toolman.org/base/basecfg"
	"toolman.org/base/toolman/v2"
)

const (
	defaultCommand  = "gogetter"
	defaultHostname = "www.toolman.org"
	requireOauth    = false
)

type Config struct {
	Hostname      string      `cfg:"hostname"`
	Port          int64       `cfg:"port"`
	Socket        string      `cfg:"socket"`
	LogDir        string      `cfg:"logdir"`
	ClientID      string      `cfg:"client-id"`
	IntegrationID int         `cfg:"integration-id"`
	HookSecret    string      `cfg:"hook-secret"`
	Trans         []*TransDef `cfg:"translators"`
	APIKey        string      `cfg:"api-key"`

	*basecfg.Config
}

type TransDef struct {
	Prefix string   `cfg:"prefix"`
	Owners []string `cfg:"owners,flow"`
}

func New() *Config {
	pflag.ErrHelp = errors.New("")

	c := &Config{
		Hostname: defaultHostname,
	}

	c.Config = basecfg.New(commandName(), basecfg.Base(c))

	return c
}

func (c *Config) FlagSet(fs *pflag.FlagSet) {
	fs.StringVar(&c.Hostname, "hostname", c.Hostname, "Service's public hostname (for callback URL)")

	fs.Int64Var(&c.Port, "port", 0, "TCP Listen Port")
	fs.StringVar(&c.Socket, "socket", "", "FastCGI Unix-Domain Socket")
}

func (c *Config) Validate() error {
	if f := pflag.Lookup("log_dir"); f != nil {
		fmt.Printf("\n#####[log_dir]: %#v\n\n", f)
	}

	c.LogDir = c.deriveLogDir()

	if c.APIKey == "" {
		return errors.New("config has no Github API Key")
	}

	if c.Port == 0 && c.Socket == "" {
		return errors.New("must specify one of --port or --socket")
	}

	if c.Port != 0 && c.Socket != "" {
		return errors.New("only one of --port or --socket may be specified")
	}

	if c.IntegrationID == 0 {
		return errors.New("missing integration ID")
	}

	if requireOauth {
		var missing []string
		for k, v := range map[string]string{"client-id": c.ClientID, "hook-secret": c.HookSecret} {
			if v == "" {
				missing = append(missing, k)
			}
		}

		if len(missing) != 0 {
			return fmt.Errorf("missing config parameters: %q", missing)
		}
	}

	return nil
}

const logDirFlag = "log_dir"

func (c *Config) deriveLogDir() string {
	if pflag.CommandLine.Changed(logDirFlag) {
		ld, _ := pflag.CommandLine.GetString(logDirFlag)
		return ld
	}

	return c.LogDir
}

func commandName() string {
	if cmd := toolman.CommandName(); cmd != "" && cmd != "debug" {
		return cmd
	}
	return defaultCommand
}
