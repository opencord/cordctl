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
	"bufio"
	b64 "encoding/base64"
	"fmt"
	"github.com/fullstorydev/grpcurl"
	versionUtils "github.com/hashicorp/go-version"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/opencord/cordctl/internal/pkg/config"
	corderrors "github.com/opencord/cordctl/internal/pkg/error"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"log"
	"os"
	"strings"
)

// Flags for calling the InitReflectionClient Method
const (
	INIT_DEFAULT          = 0
	INIT_NO_VERSION_CHECK = 1 // Do not check whether server is allowed version
)

func GenerateHeaders() []string {
	username := config.GlobalConfig.Username
	password := config.GlobalConfig.Password
	sEnc := b64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	headers := []string{"authorization: basic " + sEnc}
	return headers
}

// Perform the GetVersion API call on the core to get the version
func GetVersion(conn *grpc.ClientConn, descriptor grpcurl.DescriptorSource) (*dynamic.Message, error) {
	ctx, cancel := GrpcTimeoutContext(context.Background())
	defer cancel()

	headers := GenerateHeaders()

	h := &RpcEventHandler{}
	err := grpcurl.InvokeRPC(ctx, descriptor, conn, "xos.utility.GetVersion", headers, h, h.GetParams)
	if err != nil {
		return nil, corderrors.RpcErrorToCordError(err)
	}

	if h.Status != nil && h.Status.Err() != nil {
		return nil, corderrors.RpcErrorToCordError(h.Status.Err())
	}

	d, err := dynamic.AsDynamicMessage(h.Response)

	return d, err
}

// Initialize client connection
//    flags is a set of optional flags that may influence how the connection is setup
//        INIT_DEFAULT - default behavior (0)
//        INIT_NO_VERSION_CHECK - do not perform core version check

func InitClient(flags uint32) (*grpc.ClientConn, grpcurl.DescriptorSource, error) {
	conn, err := NewConnection()
	if err != nil {
		return nil, nil, err
	}

	refClient := grpcreflect.NewClient(context.Background(), reflectpb.NewServerReflectionClient(conn))
	defer refClient.Reset()

	// Intended method of use is to download the protos via reflection API. Loading the
	// protos from a file is supported for unit testing, as the mock server does not
	// support the reflection API.

	var descriptor grpcurl.DescriptorSource
	if config.GlobalConfig.Protoset != "" {
		descriptor, err = grpcurl.DescriptorSourceFromProtoSets(config.GlobalConfig.Protoset)
		if err != nil {
			return nil, nil, err
		}
	} else {
		descriptor = grpcurl.DescriptorSourceFromServer(context.Background(), refClient)
	}

	if flags&INIT_NO_VERSION_CHECK == 0 {
		d, err := GetVersion(conn, descriptor)
		if err != nil {
			return nil, nil, err
		}
		// Note: NewVersion doesn't like the `-dev` suffix, so strip it off.
		serverVersion, err := versionUtils.NewVersion(strings.Split(d.GetFieldByName("version").(string), "-")[0])
		if err != nil {
			return nil, nil, err
		}

		constraint, err := versionUtils.NewConstraint(config.CORE_VERSION_CONSTRAINT)
		if err != nil {
			return nil, nil, err
		}

		if !constraint.Check(serverVersion) {
			return nil, nil, corderrors.WithStackTrace(&corderrors.VersionConstraintError{
				Name:       "xos-core",
				Version:    serverVersion.String(),
				Constraint: config.CORE_VERSION_CONSTRAINT})
		}

	}

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

// Print a confirmation prompt and get a response from the user
func Confirmf(format string, args ...interface{}) bool {
	if config.GlobalOptions.Yes {
		return true
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		msg := fmt.Sprintf(format, args...)
		fmt.Print(msg)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

// Returns a context used for gRPC timeouts
func GrpcTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, config.GlobalConfig.Grpc.Timeout)
}
