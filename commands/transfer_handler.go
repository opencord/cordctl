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
	"crypto/sha256"
	"fmt"
	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc"
	"hash"
	"io"
	"os"
)

/* Handlers for streaming upload and download */

type DownloadHandler struct {
	RpcEventHandler
	f      *os.File
	chunks int
	bytes  int
	status string
	hash   hash.Hash
}

type UploadHandler struct {
	RpcEventHandler
	chunksize int
	f         *os.File
	uri       string
	hash      hash.Hash
}

func (h *DownloadHandler) OnReceiveResponse(m proto.Message) {
	d, err := dynamic.AsDynamicMessage(m)
	if err != nil {
		h.status = "ERROR"
		// TODO(smbaker): How to raise an exception?
		return
	}
	chunk := d.GetFieldByName("chunk").(string)
	io.WriteString(h.hash, chunk)
	h.f.Write([]byte(chunk))
	h.chunks += 1
	h.bytes += len(chunk)
}

func (h *DownloadHandler) GetChecksum() string {
	return fmt.Sprintf("sha256:%x", h.hash.Sum(nil))
}

func (h *UploadHandler) GetParams(msg proto.Message) error {
	dmsg, err := dynamic.AsDynamicMessage(msg)
	if err != nil {
		return err
	}

	//fmt.Printf("streamer, MessageName: %s\n", dmsg.XXX_MessageName())

	block := make([]byte, h.chunksize)
	bytes_read, err := h.f.Read(block)

	if err == io.EOF {
		h.f.Close()
		//fmt.Print("EOF\n")
		return err
	}

	if err != nil {
		//fmt.Print("ERROR!\n")
		return err
	}

	chunk := string(block[:bytes_read])
	io.WriteString(h.hash, chunk)

	dmsg.TrySetFieldByName("uri", h.uri)
	dmsg.TrySetFieldByName("chunk", chunk)

	return nil
}

func (h *UploadHandler) GetChecksum() string {
	return fmt.Sprintf("sha256:%x", h.hash.Sum(nil))
}

func UploadFile(conn *grpc.ClientConn, descriptor grpcurl.DescriptorSource, local_name string, uri string, chunkSize int) (*UploadHandler, *dynamic.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	headers := GenerateHeaders()

	f, err := os.Open(local_name)
	if err != nil {
		return nil, nil, err
	}

	h := &UploadHandler{uri: uri,
		f:         f,
		chunksize: chunkSize,
		hash:      sha256.New()}

	err = grpcurl.InvokeRPC(ctx, descriptor, conn, "xos.filetransfer/Upload", headers, h, h.GetParams)
	if err != nil {
		return nil, nil, err
	}
	if h.Status.Err() != nil {
		return nil, nil, h.Status.Err()
	}
	d, err := dynamic.AsDynamicMessage(h.Response)
	if err != nil {
		return nil, nil, err
	}

	return h, d, err
}

func DownloadFile(conn *grpc.ClientConn, descriptor grpcurl.DescriptorSource, uri string, local_name string) (*DownloadHandler, error) {
	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	headers := GenerateHeaders()

	f, err := os.Create(local_name)
	if err != nil {
		return nil, err
	}

	dm := make(map[string]interface{})
	dm["uri"] = uri

	h := &DownloadHandler{
		RpcEventHandler: RpcEventHandler{
			Fields: map[string]map[string]interface{}{"xos.FileRequest": dm},
		},
		f:      f,
		hash:   sha256.New(),
		status: "SUCCESS"}

	err = grpcurl.InvokeRPC(ctx, descriptor, conn, "xos.filetransfer/Download", headers, h, h.GetParams)
	if err != nil {
		return nil, err
	}

	if h.Status.Err() != nil {
		return nil, h.Status.Err()
	}

	return h, err
}
