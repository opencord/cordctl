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
	"github.com/opencord/cordctl/format"
)

const (
	DEFAULT_BACKUP_FORMAT = "table{{ .Status }}\t{{ .Checksum }}\t{{ .Chunks }}\t{{ .Bytes }}"
)

type BackupOutput struct {
	Status   string `json:"status"`
	Checksum string `json:"checksum"`
	Chunks   int    `json:"chunks"`
	Bytes    int    `json:"bytes"`
}

type BackupCreate struct {
	OutputOptions
	ChunkSize int `short:"h" long:"chunksize" default:"65536" description:"Host and port"`
	Args      struct {
		LocalFileName string
	} `positional-args:"yes" required:"yes"`
}

type BackupRestore struct {
	OutputOptions
	Args struct {
		LocalFileName string
	} `positional-args:"yes" required:"yes"`
}

type BackupOpts struct {
	Create  BackupCreate  `command:"create"`
	Restore BackupRestore `command:"restore"`
}

var backupOpts = BackupOpts{}

func RegisterBackupCommands(parser *flags.Parser) {
	parser.AddCommand("backup", "backup management commands", "Commands to create backups and restore backups", &backupOpts)
}

func (options *BackupCreate) Execute(args []string) error {
	conn, descriptor, err := InitReflectionClient()
	if err != nil {
		return err
	}
	defer conn.Close() // This might be closed in WaitForEnactedWithReconnect -- is that a problem?

	// We might close and reopen the connection befor we do the DownloadFile,
	// so make sure we've downloaded the service descriptor.
	_, err = descriptor.FindSymbol("xos.filetransfer")
	if err != nil {
		return err
	}

	local_name := options.Args.LocalFileName

	// STEP 1: Create backup operation
	backupop := make(map[string]interface{})
	backupop["operation"] = "create"
	err = CreateModel(conn, descriptor, "BackupOperation", backupop)
	if err != nil {
		return err
	}
	fmt.Printf("Created backup operation %d\n", backupop["id"])
	fmt.Printf("Waiting for sync ")

	// STEP 2: Wait for the operation to complete
	conn, completed_backupop, err := WaitForEnactedWithReconnect(conn, descriptor, "BackupOperation", backupop["id"].(int32))
	if err != nil {
		return err
	}

	defer conn.Close()

	status := completed_backupop.GetFieldByName("status").(string)
	fmt.Printf("\nStatus: %s\n", status)

	// we've failed. leave.
	if status != "created" {
		return nil
	}

	// STEP 3: Retrieve URI
	backupfile_id := completed_backupop.GetFieldByName("file_id").(int32)
	if backupfile_id == 0 {
		return errors.New("BackupOp.file_id is not set")
	}

	completed_backupfile, err := GetModel(conn, descriptor, "BackupFile", backupfile_id)
	if err != nil {
		return err
	}

	uri := completed_backupfile.GetFieldByName("uri").(string)
	fmt.Printf("URI %s\n", uri)

	// STEP 4: Download the file

	fmt.Printf("Downloading %s\n", local_name)

	h, err := DownloadFile(conn, descriptor, uri, local_name)
	if err != nil {
		return err
	}

	// STEP 5: Show results
	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = DEFAULT_BACKUP_FORMAT
	}
	if options.Quiet {
		outputFormat = "{{.Status}}"
	}

	data := make([]BackupOutput, 1)
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

func (options *BackupRestore) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	local_name := options.Args.LocalFileName

	_ = conn
	_ = local_name

	return nil
}
