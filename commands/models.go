/*
 * Copyright 2019-present Ciena Corporation
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
	//"fmt"
	 "context"
	 "cordctl/format"
	 "github.com/fullstorydev/grpcurl"
	 flags "github.com/jessevdk/go-flags"
	 "github.com/jhump/protoreflect/dynamic"
	 pbdescriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
 )
 
 const (
	 DEFAULT_MODEL_FORMAT = "table{{ .id }}\t{{ .name }}"
 )
 
 type ModelList struct {
	 OutputOptions
	 Args struct {
		ModelName  string
     } `positional-args:"yes" required:"yes"`
 }
 
 type ModelListOutput struct {
	 Id          string
	 Name        string
 }
 
 type ModelOpts struct {
	 List    ModelList    `command:"list"`
 }
 
 var modelOpts = ModelOpts{}

 // Eventually check feedback state and hidden options.
 var ignore_fields = map[string]bool{"backend_need_delete_policy": true,
									 "no_policy": true,
									 "mount_data_sets": true,
									 "backend_code": true,
									 "backend_need_delete": true,
									 "backend_need_reap": true,
									 "leaf_model_name": true,
									 "no_sync": true,
									 "policy_status": true,
									 "policy_code": true,
									 "backend_register": true,
									 "self_content_type_id": true,}
 
 func RegisterModelCommands(parser *flags.Parser) {
	 parser.AddCommand("model", "model commands", "Commands to query and manipulate XOS models", &modelOpts)
 }

 func (options *ModelList) Execute(args []string) error {
 
	 conn, err := NewConnection()
	 if err != nil {
		 return err
	 }
	 defer conn.Close()

	 // TODO: Validate ModelName

	 method_name := "xos.xos/List" + options.Args.ModelName
 
	 descriptor, method, err := GetReflectionMethod(conn, method_name)
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
 
	 items, err := d.TryGetFieldByName("items")
	 if err != nil {
		 return err
	 }

	 field_names := make(map[string]bool)
	 data := make([]map[string]interface{}, len(items.([]interface{})))
	 for i, item := range items.([]interface{}) {
		val := item.(*dynamic.Message)
		data[i] = make(map[string]interface{})
		for _,field_desc := range val.GetKnownFields() {
			field_name := field_desc.GetName()
			field_type := field_desc.GetType()

			_, ignore := ignore_fields[field_name]
			if ignore {
				continue
			}

			if field_desc.IsRepeated() {
				 continue
			}
			switch field_type {
				case pbdescriptor.FieldDescriptorProto_TYPE_STRING:
					data[i][field_name] = val.GetFieldByName(field_name).(string)
				case pbdescriptor.FieldDescriptorProto_TYPE_INT32:
					data[i][field_name] = val.GetFieldByName(field_name).(int32)
				case pbdescriptor.FieldDescriptorProto_TYPE_BOOL:
					data[i][field_name] = val.GetFieldByName(field_name).(bool)
//				case pbdescriptor.FieldDescriptorProto_TYPE_DOUBLE:
//					data[i][field_name] = val.GetFieldByName(field_name).(double)
			}

			field_names[field_name] = true
		}
	}

	default_format := "table"
	for field_name,_ := range field_names {
		default_format = default_format + "{{ ." + field_name + "}}\t"
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = default_format
	}
	if options.Quiet {
		outputFormat = "{{.Id}}"
	}
 
	 result := CommandResult{
		 Format:   format.Format(outputFormat),
		 OutputAs: toOutputType(options.OutputAs),
		 Data:     data,
	 }
 
	 GenerateOutput(&result)
	 return nil
 }
 