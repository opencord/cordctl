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
)

const DatabaseShowFormat = `Name:          {{.Client.Name}}
Endpoint:     {{.Client.Endpoint}}
Version:      {{.Client.Version}}
Status:       {{.Client.Status}}
`

type DatabaseShowOpts struct {
	OutputOptions
}

type DatabaseShowOutput struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	Version  string `json:"version"`
	Status   string `json:"status"`
}

type DatabaseOpts struct {
	Show DatabaseShowOpts `command:"show"`
}

var databaseOpts = DatabaseOpts{}
var databaseShowOutput = DatabaseShowOutput{}

func RegisterDatabaseCommands(parser *flags.Parser) {
	parser.AddCommand("database", "database commands", "Commands to query database information", &databaseOpts)
}

func (options *DatabaseShowOpts) Execute(args []string) error {
	conn, descriptor, err := InitReflectionClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	headers := GenerateHeaders()

	h := &RpcEventHandler{}
	err = grpcurl.InvokeRPC(ctx, descriptor, conn, "xos.utility.GetDatabaseInfo", headers, h, h.GetParams)
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

	databaseShowOutput.Name = d.GetFieldByName("name").(string)
	databaseShowOutput.Endpoint = d.GetFieldByName("endpoint").(string)
	databaseShowOutput.Version = d.GetFieldByName("version").(string)
	databaseShowOutput.Status = GetEnumValue(d, "status")

	FormatAndGenerateOutput(&options.OutputOptions, DatabaseShowFormat, DatabaseShowFormat, databaseShowOutput)

	return nil
}
