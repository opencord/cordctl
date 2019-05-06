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
	"github.com/opencord/cordctl/cli/version"
	"github.com/opencord/cordctl/format"
)

type VersionDetails struct {
	Version   string `json:"version"`
	GoVersion string `json:"goversion"`
	GitCommit string `json:"gitcommit"`
	BuildTime string `json:"buildtime"`
	Os        string `json:"os"`
	Arch      string `json:"arch"`
}

type VersionOutput struct {
	Client VersionDetails `json:"client"`
}

type VersionOpts struct {
	OutputAs string `short:"o" long:"outputas" default:"table" choice:"table" choice:"json" choice:"yaml" description:"Type of output to generate"`
}

var versionOpts = VersionOpts{}

var versionInfo = VersionOutput{
	Client: VersionDetails{
		Version:   version.Version,
		GoVersion: version.GoVersion,
		GitCommit: version.GitCommit,
		Os:        version.Os,
		Arch:      version.Arch,
		BuildTime: version.BuildTime,
	},
}

func RegisterVersionCommands(parent *flags.Parser) {
	parent.AddCommand("version", "display version", "Display client version", &versionOpts)
}

const DefaultFormat = `Client:
 Version        {{.Client.Version}}
 Go version:    {{.Client.GoVersion}}
 Git commit:    {{.Client.GitCommit}}
 Built:         {{.Client.BuildTime}}
 OS/Arch:       {{.Client.Os}}/{{.Client.Arch}}
`

func (options *VersionOpts) Execute(args []string) error {
	result := CommandResult{
		Format:   format.Format(DefaultFormat),
		OutputAs: toOutputType(options.OutputAs),
		Data:     versionInfo,
	}

	GenerateOutput(&result)
	return nil
}
