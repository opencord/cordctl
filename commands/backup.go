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
	defer conn.Close()

	local_name := options.Args.LocalFileName

	_ = conn
	_ = local_name

	// Create backup operation
	backupop := make(map[string]interface{})
	backupop["operation"] = "create"
	err = CreateModel(conn, descriptor, "BackupOperation", backupop)
	if err != nil {
		return err
	}
	fmt.Printf("Created backup operation %d\n", backupop["id"])
	fmt.Printf("Waiting for sync ")

	// Wait for the operation to complete
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

	// TODO:
	//    (optionally) make a backupfile
	//    make a backupoperation
	//    execute the backupoperation
	//    wait for completion
	//    download the file

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
