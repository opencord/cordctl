/*
 * Portions copyright 2019-present Open Networking Foundation
 * Original copyright 2019-present Ciena Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

type OutputType uint8

const (
	OUTPUT_TABLE OutputType = iota
	OUTPUT_JSON
	OUTPUT_YAML

	CORE_VERSION_CONSTRAINT = ">= 3, < 5" // Support XOS major versions 3 and 4
)

var CharReplacer = strings.NewReplacer("\\t", "\t", "\\n", "\n")

type GrpcConfigSpec struct {
	Timeout time.Duration `yaml:"timeout"`
}

type TlsConfigSpec struct {
	UseTls bool   `yaml:"useTls"`
	CACert string `yaml:"caCert"`
	Cert   string `yaml:"cert"`
	Key    string `yaml:"key"`
	Verify string `yaml:"verify"`
}

type GlobalConfigSpec struct {
	Server   string        `yaml:"server"`
	Username string        `yaml:"username"`
	Password string        `yaml:"password"`
	Protoset string        `yaml:"protoset"`
	Tls      TlsConfigSpec `yaml:"tls"`
	Grpc     GrpcConfigSpec
}

var GlobalConfig = GlobalConfigSpec{
	Server: "localhost",
	Tls: TlsConfigSpec{
		UseTls: false,
	},
	Grpc: GrpcConfigSpec{
		Timeout: time.Second * 10,
	},
}

var GlobalOptions struct {
	Config   string `short:"c" long:"config" env:"CORDCONFIG" value-name:"FILE" default:"" description:"Location of client config file"`
	Server   string `short:"s" long:"server" default:"" value-name:"SERVER:PORT" description:"IP/Host and port of XOS"`
	Username string `short:"u" long:"username" value-name:"USERNAME" default:"" description:"Username to authenticate with XOS"`
	Password string `short:"p" long:"password" value-name:"PASSWORD" default:"" description:"Password to authenticate with XOS"`
	Protoset string `long:"protoset" value-name:"FILENAME" description:"Load protobuf definitions from protoset instead of reflection api"`
	Debug    bool   `short:"d" long:"debug" description:"Enable debug mode"`
	UseTLS   bool   `long:"tls" description:"Use TLS"`
	CACert   string `long:"tlscacert" value-name:"CA_CERT_FILE" description:"Trust certs signed only by this CA"`
	Cert     string `long:"tlscert" value-name:"CERT_FILE" description:"Path to TLS vertificate file"`
	Key      string `long:"tlskey" value-name:"KEY_FILE" description:"Path to TLS key file"`
	Verify   bool   `long:"tlsverify" description:"Use TLS and verify the remote"`
	Yes      bool   `short:"y" long:"yes" description:"answer yes to any confirmation prompts"`
}

func ProcessGlobalOptions() {
	if len(GlobalOptions.Config) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Printf("Unable to discover the users home directory: %s\n", err)
		}
		GlobalOptions.Config = fmt.Sprintf("%s/.cord/config", home)
	}

	info, err := os.Stat(GlobalOptions.Config)
	if err == nil && !info.IsDir() {
		configFile, err := ioutil.ReadFile(GlobalOptions.Config)
		if err != nil {
			log.Printf("configFile.Get err   #%v ", err)
		}
		err = yaml.Unmarshal(configFile, &GlobalConfig)
		if err != nil {
			log.Fatalf("Unmarshal: %v", err)
		}
	}

	// Override from environment
	//    in particualr, for passing env vars via `go test`
	env_server, present := os.LookupEnv("CORDCTL_SERVER")
	if present {
		GlobalConfig.Server = env_server
	}
	env_username, present := os.LookupEnv("CORDCTL_USERNAME")
	if present {
		GlobalConfig.Username = env_username
	}
	env_password, present := os.LookupEnv("CORDCTL_PASSWORD")
	if present {
		GlobalConfig.Password = env_password
	}
	env_protoset, present := os.LookupEnv("CORDCTL_PROTOSET")
	if present {
		GlobalConfig.Protoset = env_protoset
	}

	// Override from command line
	if GlobalOptions.Server != "" {
		GlobalConfig.Server = GlobalOptions.Server
	}
	if GlobalOptions.Username != "" {
		GlobalConfig.Username = GlobalOptions.Username
	}
	if GlobalOptions.Password != "" {
		GlobalConfig.Password = GlobalOptions.Password
	}
	if GlobalOptions.Protoset != "" {
		GlobalConfig.Protoset = GlobalOptions.Protoset
	}

	// Generate error messages for required settings
	if GlobalConfig.Server == "" {
		log.Fatal("Server is not set. Please update config file or use the -s option")
	}
	if GlobalConfig.Username == "" {
		log.Fatal("Username is not set. Please update config file or use the -u option")
	}
	if GlobalConfig.Password == "" {
		log.Fatal("Password is not set. Please update config file or use the -p option")
	}
	//Try to resolve hostname if provided for the server
	if host, port, err := net.SplitHostPort(GlobalConfig.Server); err == nil {
		if addrs, err := net.LookupHost(host); err == nil {
			GlobalConfig.Server = net.JoinHostPort(addrs[0], port)
		}
	}
}
