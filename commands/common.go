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
	b64 "encoding/base64"
	"fmt"
	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/grpcreflect"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

func GenerateHeaders() []string {
	username := GlobalConfig.Username
	password := GlobalConfig.Password
	sEnc := b64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	headers := []string{"authorization: basic " + sEnc}
	return headers
}

func InitReflectionClient() (*grpc.ClientConn, grpcurl.DescriptorSource, error) {
	conn, err := NewConnection()
	if err != nil {
		return nil, nil, err
	}

	refClient := grpcreflect.NewClient(context.Background(), reflectpb.NewServerReflectionClient(conn))
	defer refClient.Reset()

	descriptor := grpcurl.DescriptorSourceFromServer(context.Background(), refClient)

	return conn, descriptor, nil
}

// A makeshift substitute for C's Ternary operator
func Ternary_uint32(condition bool, value_true uint32, value_false uint32) uint32 {
	if condition {
		return value_true
	} else {
		return value_false
	}
}

// call printf only if visible is True
func conditional_printf(visible bool, format string, args ...interface{}) {
	if visible {
		fmt.Printf(format, args...)
	}
}
