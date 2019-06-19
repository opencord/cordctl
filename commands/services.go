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

const (
	DEFAULT_SERVICE_FORMAT = "table{{ .Name }}\t{{.Version}}\t{{.State}}"
)

type ServiceList struct {
	ListOutputOptions
}

type ServiceListOutput struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	State   string `json:"state"`
}

type ServiceOpts struct {
	List ServiceList `command:"list"`
}

var serviceOpts = ServiceOpts{}

func RegisterServiceCommands(parser *flags.Parser) {
	parser.AddCommand("service", "service commands", "Commands to query and manipulate dynamically loaded XOS Services", &serviceOpts)
}

func (options *ServiceList) Execute(args []string) error {
	conn, descriptor, err := InitClient(INIT_DEFAULT)
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	headers := GenerateHeaders()

	h := &RpcEventHandler{}
	err = grpcurl.InvokeRPC(ctx, descriptor, conn, "xos.dynamicload.GetLoadStatus", headers, h, h.GetParams)
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

	items, err := d.TryGetFieldByName("services")
	if err != nil {
		return err
	}

	data := make([]ServiceListOutput, len(items.([]interface{})))

	for i, item := range items.([]interface{}) {
		val := item.(*dynamic.Message)
		data[i].Name = val.GetFieldByName("name").(string)
		data[i].Version = val.GetFieldByName("version").(string)
		data[i].State = val.GetFieldByName("state").(string)
	}

	FormatAndGenerateListOutput(&options.ListOutputOptions, DEFAULT_SERVICE_FORMAT, "{{.Name}}", data)

	return nil
}
