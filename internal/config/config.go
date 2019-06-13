// Copyright 2019 Timothy E. Peoples
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

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
	etcdEndpoint    = "https://cfg.toolman.org:2379"
	etcdConfigKey   = "/config/gogetter.yaml"
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

	// c.Config = basecfg.New(commandName(), basecfg.Base(c), basecfg.EtcdProvider(etcdEndpoint, etcdConfigKey))
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
