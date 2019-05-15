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
	"context"
	"github.com/fullstorydev/grpcurl"
	flags "github.com/jessevdk/go-flags"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/opencord/cordctl/cli/version"
	"github.com/opencord/cordctl/format"
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
}

type VersionOutput struct {
	Client VersionDetails     `json:"client"`
	Server CoreVersionDetails `json:"server"`
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
	},
}

func RegisterVersionCommands(parent *flags.Parser) {
	parent.AddCommand("version", "display version", "Display client version", &versionOpts)
}

const DefaultFormat = `Client:
 Version         {{.Client.Version}}
 Go version:     {{.Client.GoVersion}}
 Git commit:     {{.Client.GitCommit}}
 Git dirty:      {{.Client.GitDirty}}
 Built:          {{.Client.BuildTime}}
 OS/Arch:        {{.Client.Os}}/{{.Client.Arch}}

Server:
 Version         {{.Server.Version}}
 Python version: {{.Server.PythonVersion}}
 Git commit:     {{.Server.GitCommit}}
 Built:          {{.Server.BuildTime}}
 OS/Arch:        {{.Server.Os}}/{{.Server.Arch}}
`

func (options *VersionOpts) Execute(args []string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	descriptor, method, err := GetReflectionMethod(conn, "xos.utility.GetVersion")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	headers := GenerateHeaders()

	h := &RpcEventHandler{}
	err = grpcurl.InvokeRPC(ctx, descriptor, conn, method, headers, h, h.GetParams)
	if err != nil {
		return err
	}

	if h.Status != nil && h.Status.Err() != nil {
		return h.Status.Err()
	}

	d, err := dynamic.AsDynamicMessage(h.Response)
	if err != nil {
		return err
	}

	versionInfo.Server.Version = d.GetFieldByName("version").(string)
	versionInfo.Server.PythonVersion = d.GetFieldByName("pythonVersion").(string)
	versionInfo.Server.GitCommit = d.GetFieldByName("gitCommit").(string)
	versionInfo.Server.BuildTime = d.GetFieldByName("buildTime").(string)
	versionInfo.Server.Os = d.GetFieldByName("os").(string)
	versionInfo.Server.Arch = d.GetFieldByName("arch").(string)

	result := CommandResult{
		Format:   format.Format(DefaultFormat),
		OutputAs: toOutputType(options.OutputAs),
		Data:     versionInfo,
	}

	GenerateOutput(&result)
	return nil
}
