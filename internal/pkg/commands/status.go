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
	corderrors "github.com/opencord/cordctl/internal/pkg/error"
	"strings"
)

const StatusListFormat = "table{{ .Component }}\t{{ .Name }}\t{{ .Version }}\t{{ .Connection }}\t{{ .Status }}"

type StatusListOpts struct {
	ListOutputOptions
	Filter string `short:"f" long:"filter" description:"Comma-separated list of filters"`
}

type StatusOpts struct {
	List StatusListOpts `command:"list"`
}

var statusOpts = StatusOpts{}

func RegisterStatusCommands(parser *flags.Parser) {
	parser.AddCommand("status", "status commands", "Commands to query status of various subsystems", &statusOpts)
}

func (options *StatusListOpts) Execute(args []string) error {
	conn, descriptor, err := InitClient(INIT_DEFAULT)
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := GrpcTimeoutContext(context.Background())
	defer cancel()

	headers := GenerateHeaders()

	var components []map[string]string

	// TODO(smbaker): Consider using David's client-side filtering can be used so we get filtering
	// by other fields (status, etc) for free.

	if options.Filter == "" || strings.Contains(strings.ToLower(options.Filter), "database") {
		h := &RpcEventHandler{}
		err = grpcurl.InvokeRPC(ctx, descriptor, conn, "xos.utility.GetDatabaseInfo", headers, h, h.GetParams)
		if err != nil {
			return corderrors.RpcErrorToCordError(err)
		}

		if h.Status != nil && h.Status.Err() != nil {
			return corderrors.RpcErrorToCordError(h.Status.Err())
		}

		d, err := dynamic.AsDynamicMessage(h.Response)
		if err != nil {
			return err
		}

		db_map := make(map[string]string)
		db_map["Component"] = "Database"
		db_map["Name"] = d.GetFieldByName("name").(string)
		db_map["Version"] = d.GetFieldByName("version").(string)
		db_map["Connection"] = d.GetFieldByName("connection").(string)
		db_map["Status"] = GetEnumValue(d, "status")

		components = append(components, db_map)
	}

	FormatAndGenerateListOutput(&options.ListOutputOptions, StatusListFormat, StatusListFormat, components)

	return nil
}
