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
	flags "github.com/jessevdk/go-flags"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/opencord/cordctl/format"
	"io"
	"os"
	"strings"
)

const (
	DEFAULT_TRANSFER_FORMAT = "table{{ .Status }}\t{{ .Checksum }}\t{{ .Chunks }}\t{{ .Bytes }}"
)

type TransferOutput struct {
	Status   string `json:"status"`
	Checksum string `json:"checksum"`
	Chunks   int    `json:"chunks"`
	Bytes    int    `json:"bytes"`
}

type TransferUpload struct {
	OutputOptions
	ChunkSize int `short:"h" long:"chunksize" default:"65536" description:"Host and port"`
	Args      struct {
		LocalFileName string
		URI           string
	} `positional-args:"yes" required:"yes"`
}

type TransferDownload struct {
	OutputOptions
	Args struct {
		URI           string
		LocalFileName string
	} `positional-args:"yes" required:"yes"`
}

type TransferOpts struct {
	Upload   TransferUpload   `command:"upload"`
	Download TransferDownload `command:"download"`
}

var transferOpts = TransferOpts{}

func RegisterTransferCommands(parser *flags.Parser) {
	parser.AddCommand("transfer", "file transfer commands", "Commands to transfer files to and from XOS", &transferOpts)
}

/* Handlers for streaming upload and download */

type DownloadHandler struct {
	RpcEventHandler
	f      *os.File
	chunks int
	bytes  int
	status string
}

type UploadHandler struct {
	RpcEventHandler
	chunksize int
	f         *os.File
	uri       string
}

func (h *DownloadHandler) OnReceiveResponse(m proto.Message) {
	d, err := dynamic.AsDynamicMessage(m)
	if err != nil {
		h.status = "ERROR"
		// TODO(smbaker): How to raise an exception?
		return
	}
	chunk := d.GetFieldByName("chunk").(string)
	h.f.Write([]byte(chunk))
	h.chunks += 1
	h.bytes += len(chunk)
}

func (h *UploadHandler) GetParams(msg proto.Message) error {
	dmsg, err := dynamic.AsDynamicMessage(msg)
	if err != nil {
		return err
	}

	fmt.Printf("streamer, MessageName: %s\n", dmsg.XXX_MessageName())

	block := make([]byte, h.chunksize)
	bytes_read, err := h.f.Read(block)

	if err == io.EOF {
		h.f.Close()
		fmt.Print("EOF\n")
		return err
	}

	if err != nil {
		fmt.Print("ERROR!\n")
		return err
	}

	dmsg.TrySetFieldByName("uri", h.uri)
	dmsg.TrySetFieldByName("chunk", string(block[:bytes_read]))

	return nil
}

/* Command processors */

func (options *TransferUpload) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	local_name := options.Args.LocalFileName
	uri := options.Args.URI

	descriptor, method, err := GetReflectionMethod(conn, "xos.filetransfer/Upload")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	headers := GenerateHeaders()

	f, err := os.Open(local_name)
	if err != nil {
		return err
	}

	h := &UploadHandler{uri: uri, f: f, chunksize: options.ChunkSize}

	err = grpcurl.InvokeRPC(ctx, descriptor, conn, method, headers, h, h.GetParams)
	if err != nil {
		return err
	}
	d, err := dynamic.AsDynamicMessage(h.Response)
	if err != nil {
		return err
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = DEFAULT_TRANSFER_FORMAT
	}
	if options.Quiet {
		outputFormat = "{{.Status}}"
	}

	data := make([]TransferOutput, 1)
	data[0].Checksum = d.GetFieldByName("checksum").(string)
	data[0].Chunks = int(d.GetFieldByName("chunks_received").(int32))
	data[0].Bytes = int(d.GetFieldByName("bytes_received").(int32))
	data[0].Status = GetEnumValue(d, "status")

	result := CommandResult{
		Format:   format.Format(outputFormat),
		OutputAs: toOutputType(options.OutputAs),
		Data:     data,
	}

	GenerateOutput(&result)

	return nil
}

func IsFileUri(s string) bool {
	return strings.HasPrefix(s, "file://")
}

func (options *TransferDownload) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	local_name := options.Args.LocalFileName
	uri := options.Args.URI

	if IsFileUri(local_name) {
		return errors.New("local_name argument should not be a uri")
	}

	if !IsFileUri(uri) {
		return errors.New("uri argument should be a file:// uri")
	}

	descriptor, method, err := GetReflectionMethod(conn, "xos.filetransfer/Download")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	headers := GenerateHeaders()

	f, err := os.Create(local_name)
	if err != nil {
		return err
	}

	dm := make(map[string]interface{})
	dm["uri"] = uri

	h := &DownloadHandler{
		RpcEventHandler: RpcEventHandler{
			Fields: map[string]map[string]interface{}{"xos.FileRequest": dm},
		},
		f:      f,
		chunks: 0,
		bytes:  0,
		status: "SUCCESS"}

	err = grpcurl.InvokeRPC(ctx, descriptor, conn, method, headers, h, h.GetParams)
	if err != nil {
		return err
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = DEFAULT_TRANSFER_FORMAT
	}
	if options.Quiet {
		outputFormat = "{{.Status}}"
	}

	data := make([]TransferOutput, 1)
	data[0].Chunks = h.chunks
	data[0].Bytes = h.bytes
	data[0].Status = h.status

	result := CommandResult{
		Format:   format.Format(outputFormat),
		OutputAs: toOutputType(options.OutputAs),
		Data:     data,
	}

	GenerateOutput(&result)

	return nil
}
