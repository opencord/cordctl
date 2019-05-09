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
	"errors"
	"fmt"
	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc"
	"strings"
	"time"
)

const GM_QUIET = 1
const GM_UNTIL_FOUND = 2
const GM_UNTIL_ENACTED = 4
const GM_UNTIL_STATUS = 8

// Return a list of all available model names
func GetModelNames(source grpcurl.DescriptorSource) (map[string]bool, error) {
	models := make(map[string]bool)
	methods, err := grpcurl.ListMethods(source, "xos.xos")

	if err != nil {
		return nil, err
	}

	for _, method := range methods {
		if strings.HasPrefix(method, "xos.xos.Get") {
			models[method[11:]] = true
		}
	}

	return models, nil
}

// Check to see if a model name is valid
func CheckModelName(source grpcurl.DescriptorSource, name string) error {
	models, err := GetModelNames(source)
	if err != nil {
		return err
	}
	_, present := models[name]
	if !present {
		return errors.New("Model " + name + " does not exist. Use `cordctl models available` to get a list of available models")
	}
	return nil
}

// Create a model in XOS given a map of fields
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

	if resp.HasFieldName("uuid") {
		fields["uuid"] = resp.GetFieldByName("uuid").(string)
	}

	return nil
}

// Get a model from XOS given its ID
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

// Get a model, but retry under a variety of circumstances
func GetModelWithRetry(conn *grpc.ClientConn, descriptor grpcurl.DescriptorSource, modelName string, id int32, flags uint32) (*grpc.ClientConn, *dynamic.Message, error) {
	quiet := (flags & GM_QUIET) != 0
	until_found := (flags & GM_UNTIL_FOUND) != 0
	until_enacted := (flags & GM_UNTIL_ENACTED) != 0
	until_status := (flags & GM_UNTIL_STATUS) != 0

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
			if strings.Contains(err.Error(), "rpc error: code = Unavailable") ||
				strings.Contains(err.Error(), "rpc error: code = Internal desc = stream terminated by RST_STREAM") {
				if !quiet {
					fmt.Print(".")
				}
				time.Sleep(100 * time.Millisecond)
				conn.Close()
				conn = nil
				continue
			}

			if until_found && strings.Contains(err.Error(), "rpc error: code = NotFound") {
				if !quiet {
					fmt.Print("x")
				}
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return nil, nil, err
		}

		if until_enacted && !IsEnacted(model) {
			if !quiet {
				fmt.Print("o")
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if until_status && model.GetFieldByName("status") == nil {
			if !quiet {
				fmt.Print("O")
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}

		return conn, model, nil
	}
}

// Get a model from XOS given a fieldName/fieldValue
func FindModel(conn *grpc.ClientConn, descriptor grpcurl.DescriptorSource, modelName string, fieldName string, fieldValue string) (*dynamic.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	headers := GenerateHeaders()

	// TODO(smbaker): Implement filter the right way

	h := &RpcEventHandler{}
	err := grpcurl.InvokeRPC(ctx, descriptor, conn, "xos.xos.List"+modelName, headers, h, h.GetParams)
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

	items, err := d.TryGetFieldByName("items")
	if err != nil {
		return nil, err
	}

	for _, item := range items.([]interface{}) {
		val := item.(*dynamic.Message)

		if val.GetFieldByName(fieldName).(string) == fieldValue {
			return val, nil
		}

	}

	return nil, errors.New("rpc error: code = NotFound")
}

// Find a model, but retry under a variety of circumstances
func FindModelWithRetry(conn *grpc.ClientConn, descriptor grpcurl.DescriptorSource, modelName string, fieldName string, fieldValue string, flags uint32) (*grpc.ClientConn, *dynamic.Message, error) {
	quiet := (flags & GM_QUIET) != 0
	until_found := (flags & GM_UNTIL_FOUND) != 0
	until_enacted := (flags & GM_UNTIL_ENACTED) != 0
	until_status := (flags & GM_UNTIL_STATUS) != 0

	for {
		var err error

		if conn == nil {
			conn, err = NewConnection()
			if err != nil {
				return nil, nil, err
			}
		}

		model, err := FindModel(conn, descriptor, modelName, fieldName, fieldValue)
		if err != nil {
			if strings.Contains(err.Error(), "rpc error: code = Unavailable") ||
				strings.Contains(err.Error(), "rpc error: code = Internal desc = stream terminated by RST_STREAM") {
				if !quiet {
					fmt.Print(".")
				}
				time.Sleep(100 * time.Millisecond)
				conn.Close()
				conn = nil
				continue
			}

			if until_found && strings.Contains(err.Error(), "rpc error: code = NotFound") {
				if !quiet {
					fmt.Print("x")
				}
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return nil, nil, err
		}

		if until_enacted && !IsEnacted(model) {
			if !quiet {
				fmt.Print("o")
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if until_status && model.GetFieldByName("status") == nil {
			if !quiet {
				fmt.Print("O")
			}
			time.Sleep(100 * time.Millisecond)
			continue
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
