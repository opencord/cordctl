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
	"fmt"
	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc"
	"strings"
	"time"
)

func CreateModel(conn *grpc.ClientConn, descriptor grpcurl.DescriptorSource, modelName string, fields map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	headers := GenerateHeaders()

	h := &RpcEventHandler{
		Fields: map[string]map[string]interface{}{"xos." + modelName: fields},
	}
	err := grpcurl.InvokeRPC(ctx, descriptor, conn, "xos.xos.Create"+modelName, headers, h, h.GetParams)
	if err != nil {
		return err
	} else if h.Status != nil && h.Status.Err() != nil {
		return h.Status.Err()
	}

	resp, err := dynamic.AsDynamicMessage(h.Response)
	if err != nil {
		return err
	}

	fields["id"] = resp.GetFieldByName("id").(int32)

	return nil
}

func GetModel(conn *grpc.ClientConn, descriptor grpcurl.DescriptorSource, modelName string, id int32) (*dynamic.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	headers := GenerateHeaders()

	h := &RpcEventHandler{
		Fields: map[string]map[string]interface{}{"xos.ID": map[string]interface{}{"id": id}},
	}
	err := grpcurl.InvokeRPC(ctx, descriptor, conn, "xos.xos.Get"+modelName, headers, h, h.GetParams)
	if err != nil {
		return nil, err
	}

	if h.Status != nil && h.Status.Err() != nil {
		return nil, h.Status.Err()
	}

	d, err := dynamic.AsDynamicMessage(h.Response)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func GetModelWithReconnect(conn *grpc.ClientConn, descriptor grpcurl.DescriptorSource, modelName string, id int32) (*grpc.ClientConn, *dynamic.Message, error) {
	for {
		var err error

		if conn == nil {
			conn, err = NewConnection()
			if err != nil {
				return nil, nil, err
			}
		}

		model, err := GetModel(conn, descriptor, modelName, id)
		if err != nil {
			if strings.Contains(err.Error(), "rpc error: code = Unavailable") {
				fmt.Print(".")
				time.Sleep(100 * time.Millisecond)
				conn.Close()
				conn = nil
				continue
			}
			return nil, nil, err
		}

		return conn, model, nil
	}
}

func MessageToMap(d *dynamic.Message) map[string]interface{} {
	fields := make(map[string]interface{})
	for _, field_desc := range d.GetKnownFields() {
		field_name := field_desc.GetName()
		fields[field_name] = d.GetFieldByName(field_name)
	}
	return fields
}

func IsEnacted(d *dynamic.Message) bool {
	enacted := d.GetFieldByName("enacted").(float64)
	updated := d.GetFieldByName("updated").(float64)

	return (enacted >= updated)
}

func WaitForEnactedWithReconnect(conn *grpc.ClientConn, descriptor grpcurl.DescriptorSource, modelName string, id int32) (*grpc.ClientConn, *dynamic.Message, error) {
	for {
		conn, model, err := GetModelWithReconnect(conn, descriptor, modelName, id)
		if !IsEnacted(model) {
			fmt.Print("o")
			time.Sleep(100 * time.Millisecond)
			continue
		}

		return conn, model, err
	}
}
