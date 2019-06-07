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
	"strings"
)

const DatabaseInfoFormat = `Database:
  Name:         {{.Database.Name}}
  Location:     {{.Database.Endpoint}}
  Version:      {{.Database.Version}}
  Status:       {{.Database.Status}}
`

type StatusListOpts struct {
	OutputOptions
	Filter string `short:"f" long:"filter" description:"Comma-separated list of filters"`
}

type DatabaseInfoOutput struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	Version  string `json:"version"`
	Status   string `json:"status"`
}

type StatusListOutput struct {
	Database DatabaseInfoOutput
}

type StatusOpts struct {
	List StatusListOpts `command:"list"`
}

var statusOpts = StatusOpts{}
var statusListOutput = StatusListOutput{}

func RegisterStatusCommands(parser *flags.Parser) {
	parser.AddCommand("status", "status commands", "Commands to query status of various subsystems", &statusOpts)
}

func (options *StatusListOpts) Execute(args []string) error {
	conn, descriptor, err := InitReflectionClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	headers := GenerateHeaders()

	default_format := ""

	if options.Filter == "" || strings.Contains(strings.ToLower(options.Filter), "database") {
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

		statusListOutput.Database.Name = d.GetFieldByName("name").(string)
		statusListOutput.Database.Endpoint = d.GetFieldByName("endpoint").(string)
		statusListOutput.Database.Version = d.GetFieldByName("version").(string)
		statusListOutput.Database.Status = GetEnumValue(d, "status")

		default_format += DatabaseInfoFormat
	}

	FormatAndGenerateOutput(&options.OutputOptions, default_format, default_format, statusListOutput)

	return nil
}
