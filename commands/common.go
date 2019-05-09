/*
 * Copyright 2019-present Open Networking Foundation
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
	"golang.org/x/net/context"
	"errors"
	b64 "encoding/base64"
	"github.com/fullstorydev/grpcurl"
	"strings"
	"google.golang.org/grpc"
	"github.com/jhump/protoreflect/grpcreflect"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

func GenerateHeaders() []string {
	username := GlobalConfig.Username
	password := GlobalConfig.Password
	sEnc := b64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	headers := []string{"authorization: basic " + sEnc}
	return headers
}

func GetModelNames(source grpcurl.DescriptorSource) (map[string]bool, error) {
	models := make(map[string]bool)
	methods, err := grpcurl.ListMethods(source, "xos.xos")

	if err != nil {
		return nil, err
	}

	for _,method := range methods {
	    if strings.HasPrefix(method, "xos.xos.Get") {
			models[method[11:]] = true
		}
	}

	return models, nil
}

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

func InitReflectionClient() (*grpc.ClientConn, grpcurl.DescriptorSource, error) {
	conn, err := NewConnection()
	if err != nil {
		return nil, nil, err
	}
	//defer conn.Close()

	refClient := grpcreflect.NewClient(context.Background(), reflectpb.NewServerReflectionClient(conn))
	defer refClient.Reset()

	descriptor := grpcurl.DescriptorSourceFromServer(context.Background(), refClient)

	return conn, descriptor, nil
}
