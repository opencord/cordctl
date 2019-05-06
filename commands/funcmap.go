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
	"fmt"
	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

type MethodNotFoundError struct {
	Name string
}

func (e *MethodNotFoundError) Error() string {
	return fmt.Sprintf("Method '%s' not found in function map", e.Name)
}

type MethodVersionNotFoundError struct {
	Name    string
	Version string
}

func (e *MethodVersionNotFoundError) Error() string {
	return fmt.Sprintf("Method '%s' does not have a verison for '%s' specfied in function map", e.Name, e.Version)
}

type DescriptorNotFoundError struct {
	Version string
}

func (e *DescriptorNotFoundError) Error() string {
	return fmt.Sprintf("Protocol buffer descriptor for API version '%s' not found", e.Version)
}

type UnableToParseDescriptorErrror struct {
	err     error
	Version string
}

func (e *UnableToParseDescriptorErrror) Error() string {
	return fmt.Sprintf("Unable to parse protocal buffer descriptor for version '%s': %s", e.Version, e.err)
}

func GetReflectionMethod(conn *grpc.ClientConn, name string) (grpcurl.DescriptorSource, string, error) {
	refClient := grpcreflect.NewClient(context.Background(), reflectpb.NewServerReflectionClient(conn))
	defer refClient.Reset()

	desc := grpcurl.DescriptorSourceFromServer(context.Background(), refClient)

	return desc, name, nil
}

func GetEnumValue(val *dynamic.Message, name string) string {
	return val.FindFieldDescriptorByName(name).GetEnumType().
		FindValueByNumber(val.GetFieldByName(name).(int32)).GetName()
}
