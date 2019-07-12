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
package commands

import (
	flags "github.com/jessevdk/go-flags"
	"github.com/opencord/cordctl/internal/pkg/cli/version"
	"github.com/opencord/cordctl/pkg/format"
)

type VersionDetails struct {
	Version   string `json:"version"`
	GoVersion string `json:"goversion"`
	GitCommit string `json:"gitcommit"`
	GitDirty  string `json:"gitdirty"`
	BuildTime string `json:"buildtime"`
	Os        string `json:"os"`
	Arch      string `json:"arch"`
}

type CoreVersionDetails struct {
	Version       string `json:"version"`
	PythonVersion string `json:"goversion"`
	GitCommit     string `json:"gitcommit"`
	BuildTime     string `json:"buildtime"`
	Os            string `json:"os"`
	Arch          string `json:"arch"`
	DjangoVersion string `json:"djangoversion"`
}

type VersionOutput struct {
	Client VersionDetails     `json:"client"`
	Server CoreVersionDetails `json:"server"`
}

type VersionOpts struct {
	OutputAs   string `short:"o" long:"outputas" default:"table" choice:"table" choice:"json" choice:"yaml" description:"Type of output to generate"`
	ClientOnly bool   `short:"c" long:"client-only" description:"Print only client version"`
}

var versionOpts = VersionOpts{}

var versionInfo = VersionOutput{
	Client: VersionDetails{
		Version:   version.Version,
		GoVersion: version.GoVersion,
		GitCommit: version.GitCommit,
		GitDirty:  version.GitDirty,
		Os:        version.Os,
		Arch:      version.Arch,
		BuildTime: version.BuildTime,
	},
	Server: CoreVersionDetails{
		Version:       "unknown",
		PythonVersion: "unknown",
		GitCommit:     "unknown",
		Os:            "unknown",
		Arch:          "unknown",
		BuildTime:     "unknown",
		DjangoVersion: "unknown",
	},
}

func RegisterVersionCommands(parent *flags.Parser) {
	parent.AddCommand("version", "display version", "Display client version", &versionOpts)
}

const ClientFormat = `Client:
 Version         {{.Client.Version}}
 Go version:     {{.Client.GoVersion}}
 Git commit:     {{.Client.GitCommit}}
 Git dirty:      {{.Client.GitDirty}}
 Built:          {{.Client.BuildTime}}
 OS/Arch:        {{.Client.Os}}/{{.Client.Arch}}
`
const ServerFormat = `
Server:
 Version         {{.Server.Version}}
 Python version: {{.Server.PythonVersion}}
 Django version: {{.Server.DjangoVersion}}
 Git commit:     {{.Server.GitCommit}}
 Built:          {{.Server.BuildTime}}
 OS/Arch:        {{.Server.Os}}/{{.Server.Arch}}
`

const DefaultFormat = ClientFormat + ServerFormat

func (options *VersionOpts) Execute(args []string) error {
	if !options.ClientOnly {
		conn, descriptor, err := InitClient(INIT_NO_VERSION_CHECK)
		if err != nil {
			return err
		}
		defer conn.Close()

		d, err := GetVersion(conn, descriptor)
		if err != nil {
			return err
		}

		versionInfo.Server.Version = d.GetFieldByName("version").(string)
		versionInfo.Server.PythonVersion = d.GetFieldByName("pythonVersion").(string)
		versionInfo.Server.GitCommit = d.GetFieldByName("gitCommit").(string)
		versionInfo.Server.BuildTime = d.GetFieldByName("buildTime").(string)
		versionInfo.Server.Os = d.GetFieldByName("os").(string)
		versionInfo.Server.Arch = d.GetFieldByName("arch").(string)

		// djangoVersion was added to GetVersion() in xos-core 3.3.1-dev
		djangoVersion, err := d.TryGetFieldByName("djangoVersion")
		if err == nil {
			versionInfo.Server.DjangoVersion = djangoVersion.(string)
		}
	}

	result := CommandResult{
		// Format:   format.Format(DefaultFormat),
		OutputAs: toOutputType(options.OutputAs),
		Data:     versionInfo,
	}

	if options.ClientOnly {
		result.Format = format.Format(ClientFormat)
	} else {
		result.Format = format.Format(DefaultFormat)
	}

	GenerateOutput(&result)
	return nil
}
