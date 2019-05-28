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
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc"
	"io"
	"strconv"
	"strings"
	"time"
)

// Flags for calling the *WithRetry methods
const GM_QUIET = 1
const GM_UNTIL_FOUND = 2
const GM_UNTIL_ENACTED = 4
const GM_UNTIL_STATUS = 8

type QueryEventHandler struct {
	RpcEventHandler
	Elements map[string]string
	Model    *desc.MessageDescriptor
	Kind     string
	EOF      bool
}

// Separate the operator from the query value.
// For example,
//    "==foo"  --> "EQUAL", "foo"
func DecodeOperator(query string) (string, string, bool, error) {
	if strings.HasPrefix(query, "!=") {
		return strings.TrimSpace(query[2:]), "EQUAL", true, nil
	} else if strings.HasPrefix(query, "==") {
		return "", "", false, errors.New("Operator == is now allowed. Suggest using = instead.")
	} else if strings.HasPrefix(query, "=") {
		return strings.TrimSpace(query[1:]), "EQUAL", false, nil
	} else if strings.HasPrefix(query, ">=") {
		return strings.TrimSpace(query[2:]), "GREATER_THAN_OR_EQUAL", false, nil
	} else if strings.HasPrefix(query, ">") {
		return strings.TrimSpace(query[1:]), "GREATER_THAN", false, nil
	} else if strings.HasPrefix(query, "<=") {
		return strings.TrimSpace(query[2:]), "LESS_THAN_OR_EQUAL", false, nil
	} else if strings.HasPrefix(query, "<") {
		return strings.TrimSpace(query[1:]), "LESS_THAN", false, nil
	} else {
		return strings.TrimSpace(query), "EQUAL", false, nil
	}
}

// Generate the parameters for Query messages.
func (h *QueryEventHandler) GetParams(msg proto.Message) error {
	dmsg, err := dynamic.AsDynamicMessage(msg)
	if err != nil {
		return err
	}

	//fmt.Printf("MessageName: %s\n", dmsg.XXX_MessageName())

	if h.EOF {
		return io.EOF
	}

	// Get the MessageType for the `elements` field
	md := dmsg.GetMessageDescriptor()
	elements_fld := md.FindFieldByName("elements")
	elements_mt := elements_fld.GetMessageType()

	for field_name, element := range h.Elements {
		value, operator, invert, err := DecodeOperator(element)
		if err != nil {
			return err
		}

		nm := dynamic.NewMessage(elements_mt)

		field_descriptor := h.Model.FindFieldByName(field_name)
		if field_descriptor == nil {
			return fmt.Errorf("Field %s does not exist", field_name)
		}

		field_type := field_descriptor.GetType()
		switch field_type {
		case descriptor.FieldDescriptorProto_TYPE_INT32:
			var i int64
			i, err = strconv.ParseInt(value, 10, 32)
			nm.SetFieldByName("iValue", int32(i))
		case descriptor.FieldDescriptorProto_TYPE_UINT32:
			var i int64
			i, err = strconv.ParseInt(value, 10, 32)
			nm.SetFieldByName("iValue", uint32(i))
		case descriptor.FieldDescriptorProto_TYPE_FLOAT:
			err = errors.New("Floating point filters are unsupported")
		case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
			err = errors.New("Floating point filters are unsupported")
		default:
			nm.SetFieldByName("sValue", value)
			err = nil
		}

		if err != nil {
			return err
		}

		nm.SetFieldByName("name", field_name)
		nm.SetFieldByName("invert", invert)
		SetEnumValue(nm, "operator", operator)
		dmsg.AddRepeatedFieldByName("elements", nm)
	}

	SetEnumValue(dmsg, "kind", h.Kind)

	h.EOF = true

	return nil
}

// Take a string list of queries and turns it into a map of queries
func QueryStringsToMap(query_args []string, allow_inequality bool) (map[string]string, error) {
	queries := make(map[string]string)
	for _, query_str := range query_args {
		query_str := strings.TrimSpace(query_str)
		operator_pos := -1
		for i, ch := range query_str {
			if allow_inequality {
				if (ch == '!') || (ch == '=') || (ch == '>') || (ch == '<') {
					operator_pos = i
					break
				}
			} else {
				if ch == '=' {
					operator_pos = i
					break
				}
			}
		}
		if operator_pos == -1 {
			return nil, fmt.Errorf("Illegal operator/value string %s", query_str)
		}
		queries[strings.TrimSpace(query_str[:operator_pos])] = query_str[operator_pos:]
	}
	return queries, nil
}

// Take a string of comma-separated queries and turn it into a map of queries
func CommaSeparatedQueryToMap(query_str string, allow_inequality bool) (map[string]string, error) {
	if query_str == "" {
		return nil, nil
	}

	query_strings := strings.Split(query_str, ",")
	return QueryStringsToMap(query_strings, allow_inequality)
}

// Convert a string into the appropriate gRPC type for a given field
func TypeConvert(source grpcurl.DescriptorSource, modelName string, field_name string, v string) (interface{}, error) {
	model_descriptor, err := source.FindSymbol("xos." + modelName)
	if err != nil {
		return nil, err
	}
	model_md, ok := model_descriptor.(*desc.MessageDescriptor)
	if !ok {
		return nil, fmt.Errorf("Failed to convert model %s to a messagedescriptor", modelName)
	}
	field_descriptor := model_md.FindFieldByName(field_name)
	if field_descriptor == nil {
		return nil, fmt.Errorf("Field %s does not exist in model %s", field_name, modelName)
	}
	field_type := field_descriptor.GetType()

	var result interface{}

	switch field_type {
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		var i int64
		i, err = strconv.ParseInt(v, 10, 32)
		result = int32(i)
	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		var i int64
		i, err = strconv.ParseInt(v, 10, 32)
		result = uint32(i)
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		var f float64
		f, err = strconv.ParseFloat(v, 32)
		result = float32(f)
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		var f float64
		f, err = strconv.ParseFloat(v, 64)
		result = f
	default:
		result = v
		err = nil
	}

	return result, err
}

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

// Update a model in XOS given a map of fields
func UpdateModel(conn *grpc.ClientConn, descriptor grpcurl.DescriptorSource, modelName string, fields map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	headers := GenerateHeaders()

	h := &RpcEventHandler{
		Fields: map[string]map[string]interface{}{"xos." + modelName: fields},
	}
	err := grpcurl.InvokeRPC(ctx, descriptor, conn, "xos.xos.Update"+modelName, headers, h, h.GetParams)
	if err != nil {
		return err
	} else if h.Status != nil && h.Status.Err() != nil {
		return h.Status.Err()
	}

	resp, err := dynamic.AsDynamicMessage(h.Response)
	if err != nil {
		return err
	}

	// TODO: Do we need to do anything with the response?
	_ = resp

	return nil
}

// Get a model from XOS given its ID
func GetModel(ctx context.Context, conn *grpc.ClientConn, descriptor grpcurl.DescriptorSource, modelName string, id int32) (*dynamic.Message, error) {
	ctx, cancel := context.WithTimeout(ctx, GlobalConfig.Grpc.Timeout)
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
func GetModelWithRetry(ctx context.Context, conn *grpc.ClientConn, descriptor grpcurl.DescriptorSource, modelName string, id int32, flags uint32) (*grpc.ClientConn, *dynamic.Message, error) {
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

		model, err := GetModel(ctx, conn, descriptor, modelName, id)
		if err != nil {
			if strings.Contains(err.Error(), "rpc error: code = Unavailable") ||
				strings.Contains(err.Error(), "rpc error: code = Internal desc = stream terminated by RST_STREAM") {
				if !quiet {
					fmt.Print(".")
				}
				select {
				case <-time.After(100 * time.Millisecond):
				case <-ctx.Done():
					return nil, nil, ctx.Err()
				}
				conn.Close()
				conn = nil
				continue
			}

			if until_found && strings.Contains(err.Error(), "rpc error: code = NotFound") {
				if !quiet {
					fmt.Print("x")
				}
				select {
				case <-time.After(100 * time.Millisecond):
				case <-ctx.Done():
					return nil, nil, ctx.Err()
				}
				continue
			}
			return nil, nil, err
		}

		if until_enacted && !IsEnacted(model) {
			if !quiet {
				fmt.Print("o")
			}
			select {
			case <-time.After(100 * time.Millisecond):
			case <-ctx.Done():
				return nil, nil, ctx.Err()
			}
			continue
		}

		if until_status && model.GetFieldByName("status") == nil {
			if !quiet {
				fmt.Print("O")
			}
			select {
			case <-time.After(100 * time.Millisecond):
			case <-ctx.Done():
				return nil, nil, ctx.Err()
			}
			continue
		}

		return conn, model, nil
	}
}

func ItemsToDynamicMessageList(items interface{}) []*dynamic.Message {
	result := make([]*dynamic.Message, len(items.([]interface{})))
	for i, item := range items.([]interface{}) {
		result[i] = item.(*dynamic.Message)
	}
	return result
}

// List all objects of a given model
func ListModels(ctx context.Context, conn *grpc.ClientConn, descriptor grpcurl.DescriptorSource, modelName string) ([]*dynamic.Message, error) {
	ctx, cancel := context.WithTimeout(ctx, GlobalConfig.Grpc.Timeout)
	defer cancel()

	headers := GenerateHeaders()

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

	return ItemsToDynamicMessageList(items), nil
}

// Filter models based on field values
//   queries is a map of <field_name> to <operator><query>
//   For example,
//     map[string]string{"name": "==mysite"}
func FilterModels(ctx context.Context, conn *grpc.ClientConn, descriptor grpcurl.DescriptorSource, modelName string, queries map[string]string) ([]*dynamic.Message, error) {
	ctx, cancel := context.WithTimeout(ctx, GlobalConfig.Grpc.Timeout)
	defer cancel()

	headers := GenerateHeaders()

	model_descriptor, err := descriptor.FindSymbol("xos." + modelName)
	if err != nil {
		return nil, err
	}
	model_md, ok := model_descriptor.(*desc.MessageDescriptor)
	if !ok {
		return nil, errors.New("Failed to convert model to a messagedescriptor")
	}

	h := &QueryEventHandler{
		RpcEventHandler: RpcEventHandler{
			Fields: map[string]map[string]interface{}{"xos.Query": map[string]interface{}{"kind": 0}},
		},
		Elements: queries,
		Model:    model_md,
		Kind:     "DEFAULT",
	}
	err = grpcurl.InvokeRPC(ctx, descriptor, conn, "xos.xos.Filter"+modelName, headers, h, h.GetParams)
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

	return ItemsToDynamicMessageList(items), nil
}

// Call ListModels or FilterModels as appropriate
func ListOrFilterModels(ctx context.Context, conn *grpc.ClientConn, descriptor grpcurl.DescriptorSource, modelName string, queries map[string]string) ([]*dynamic.Message, error) {
	if len(queries) == 0 {
		return ListModels(ctx, conn, descriptor, modelName)
	} else {
		return FilterModels(ctx, conn, descriptor, modelName, queries)
	}
}

// Get a model from XOS given a fieldName/fieldValue
func FindModel(ctx context.Context, conn *grpc.ClientConn, descriptor grpcurl.DescriptorSource, modelName string, queries map[string]string) (*dynamic.Message, error) {
	models, err := FilterModels(ctx, conn, descriptor, modelName, queries)
	if err != nil {
		return nil, err
	}

	if len(models) == 0 {
		return nil, errors.New("rpc error: code = NotFound")
	}

	return models[0], nil
}

// Find a model, but retry under a variety of circumstances
func FindModelWithRetry(ctx context.Context, conn *grpc.ClientConn, descriptor grpcurl.DescriptorSource, modelName string, queries map[string]string, flags uint32) (*grpc.ClientConn, *dynamic.Message, error) {
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

		model, err := FindModel(ctx, conn, descriptor, modelName, queries)
		if err != nil {
			if strings.Contains(err.Error(), "rpc error: code = Unavailable") ||
				strings.Contains(err.Error(), "rpc error: code = Internal desc = stream terminated by RST_STREAM") {
				if !quiet {
					fmt.Print(".")
				}
				select {
				case <-time.After(100 * time.Millisecond):
				case <-ctx.Done():
					return nil, nil, ctx.Err()
				}
				conn.Close()
				conn = nil
				continue
			}

			if until_found && strings.Contains(err.Error(), "rpc error: code = NotFound") {
				if !quiet {
					fmt.Print("x")
				}
				select {
				case <-time.After(100 * time.Millisecond):
				case <-ctx.Done():
					return nil, nil, ctx.Err()
				}
				continue
			}
			return nil, nil, err
		}

		if until_enacted && !IsEnacted(model) {
			if !quiet {
				fmt.Print("o")
			}
			select {
			case <-time.After(100 * time.Millisecond):
			case <-ctx.Done():
				return nil, nil, ctx.Err()
			}
			continue
		}

		if until_status && model.GetFieldByName("status") == nil {
			if !quiet {
				fmt.Print("O")
			}
			select {
			case <-time.After(100 * time.Millisecond):
			case <-ctx.Done():
				return nil, nil, ctx.Err()
			}
			continue
		}

		return conn, model, nil
	}
}

// Get a model from XOS given its ID
func DeleteModel(conn *grpc.ClientConn, descriptor grpcurl.DescriptorSource, modelName string, id int32) error {
	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	headers := GenerateHeaders()

	h := &RpcEventHandler{
		Fields: map[string]map[string]interface{}{"xos.ID": map[string]interface{}{"id": id}},
	}
	err := grpcurl.InvokeRPC(ctx, descriptor, conn, "xos.xos.Delete"+modelName, headers, h, h.GetParams)
	if err != nil {
		return err
	}

	if h.Status != nil && h.Status.Err() != nil {
		return h.Status.Err()
	}

	_, err = dynamic.AsDynamicMessage(h.Response)
	if err != nil {
		return err
	}

	return nil
}

// Takes a *dynamic.Message and turns it into a map of fields to interfaces
//    TODO: Might be more useful to convert the values to strings and ints
func MessageToMap(d *dynamic.Message) map[string]interface{} {
	fields := make(map[string]interface{})
	for _, field_desc := range d.GetKnownFields() {
		field_name := field_desc.GetName()
		fields[field_name] = d.GetFieldByName(field_name)
	}
	return fields
}

// Returns True if a message has been enacted
func IsEnacted(d *dynamic.Message) bool {
	enacted := d.GetFieldByName("enacted").(float64)
	updated := d.GetFieldByName("updated").(float64)

	return (enacted >= updated)
}
