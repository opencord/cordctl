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
	"errors"
	"fmt"
	flags "github.com/jessevdk/go-flags"
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

/* Command processors */

func (options *TransferUpload) Execute(args []string) error {

	conn, descriptor, err := InitReflectionClient()
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

	h, upload_result, err := UploadFile(conn, descriptor, local_name, uri, options.ChunkSize)
	if err != nil {
		return err
	}

	if upload_result.GetFieldByName("checksum").(string) != h.GetChecksum() {
		return fmt.Errorf("Checksum mismatch, expected=%s, received=%s",
			h.GetChecksum(),
			upload_result.GetFieldByName("checksum").(string))
	}

	data := make([]TransferOutput, 1)
	data[0].Checksum = upload_result.GetFieldByName("checksum").(string)
	data[0].Chunks = int(upload_result.GetFieldByName("chunks_received").(int32))
	data[0].Bytes = int(upload_result.GetFieldByName("bytes_received").(int32))
	data[0].Status = GetEnumValue(upload_result, "status")

	FormatAndGenerateOutput(&options.OutputOptions, DEFAULT_TRANSFER_FORMAT, "{{.Status}}", data)

	return nil
}

func IsFileUri(s string) bool {
	return strings.HasPrefix(s, "file://")
}

func (options *TransferDownload) Execute(args []string) error {
	conn, descriptor, err := InitReflectionClient()
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

	h, err := DownloadFile(conn, descriptor, uri, local_name)
	if err != nil {
		return err
	}

	data := make([]TransferOutput, 1)
	data[0].Chunks = h.chunks
	data[0].Bytes = h.bytes
	data[0].Status = h.status
	data[0].Checksum = h.GetChecksum()

	FormatAndGenerateOutput(&options.OutputOptions, DEFAULT_TRANSFER_FORMAT, "{{.Status}}", data)

	return nil
}
