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
	flags "github.com/jessevdk/go-flags"
	corderrors "github.com/opencord/cordctl/internal/pkg/error"
	"time"
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
	ChunkSize int `short:"h" long:"chunksize" default:"65536" description:"Chunk size for streaming transfer"`
	Args      struct {
		LocalFileName string
	} `positional-args:"yes" required:"yes"`
}

type BackupRestore struct {
	OutputOptions
	ChunkSize int `short:"h" long:"chunksize" default:"65536" description:"Chunk size for streaming transfer"`
	Args      struct {
		LocalFileName string
	} `positional-args:"yes" required:"yes"`
	CreateURIFunc func() (string, string) // allow override of CreateURIFunc for easy unit testing
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
	conn, descriptor, err := InitClient(INIT_DEFAULT)
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx := context.Background() // TODO: Implement a sync timeout

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
	conditional_printf(!options.Quiet, "Created backup-create operation id=%d uuid=%s\n", backupop["id"], backupop["uuid"])
	conditional_printf(!options.Quiet, "Waiting for sync ")

	// STEP 2: Wait for the operation to complete

	flags := GM_UNTIL_ENACTED | GM_UNTIL_FOUND | Ternary_uint32(options.Quiet, GM_QUIET, 0)
	conn, completed_backupop, err := GetModelWithRetry(ctx, conn, descriptor, "BackupOperation", backupop["id"].(int32), flags)
	if err != nil {
		return err
	}

	defer conn.Close()

	status := completed_backupop.GetFieldByName("status").(string)
	conditional_printf(!options.Quiet, "\nStatus: %s\n", status)

	// we've failed. leave.
	if status != "created" {
		return corderrors.NewInternalError("BackupOp status is %s", status)
	}

	// STEP 3: Retrieve URI
	backupfile_id := completed_backupop.GetFieldByName("file_id").(int32)
	if backupfile_id == 0 {
		return corderrors.NewInternalError("BackupOp.file_id is not set")
	}

	completed_backupfile, err := GetModel(ctx, conn, descriptor, "BackupFile", backupfile_id)
	if err != nil {
		return err
	}

	uri := completed_backupfile.GetFieldByName("uri").(string)
	conditional_printf(!options.Quiet, "URI %s\n", uri)

	// STEP 4: Download the file

	conditional_printf(!options.Quiet, "Downloading %s\n", local_name)

	h, err := DownloadFile(conn, descriptor, uri, local_name)
	if err != nil {
		return err
	}

	// STEP 5: Verify checksum

	if completed_backupfile.GetFieldByName("checksum").(string) != h.GetChecksum() {
		return corderrors.WithStackTrace(&corderrors.ChecksumMismatchError{
			Actual:   h.GetChecksum(),
			Expected: completed_backupfile.GetFieldByName("checksum").(string)})
	}

	// STEP 6: Show results

	data := make([]BackupOutput, 1)
	data[0].Chunks = h.chunks
	data[0].Bytes = h.bytes
	data[0].Status = h.status
	data[0].Checksum = h.GetChecksum()

	FormatAndGenerateOutput(&options.OutputOptions, DEFAULT_BACKUP_FORMAT, "{{.Status}}", data)

	return nil
}

// Create a file:/// URI to use for storing the file in the core
func CreateDynamicURI() (string, string) {
	remote_name := "cordctl-restore-" + time.Now().Format("20060102T150405Z")
	uri := "file:///var/run/xos/backup/local/" + remote_name
	return remote_name, uri
}

func (options *BackupRestore) Execute(args []string) error {
	conn, descriptor, err := InitClient(INIT_DEFAULT)
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx := context.Background() // TODO: Implement a sync timeout

	local_name := options.Args.LocalFileName

	var remote_name, uri string
	if options.CreateURIFunc != nil {
		remote_name, uri = options.CreateURIFunc()
	} else {
		remote_name, uri = CreateDynamicURI()
	}

	// STEP 1: Upload the file

	h, upload_result, err := UploadFile(conn, descriptor, local_name, uri, options.ChunkSize)
	if err != nil {
		return err
	}

	upload_status := GetEnumValue(upload_result, "status")
	if upload_status != "SUCCESS" {
		return corderrors.NewInternalError("Upload status was %s", upload_status)
	}

	// STEP 2: Verify checksum

	if upload_result.GetFieldByName("checksum").(string) != h.GetChecksum() {
		return corderrors.WithStackTrace(&corderrors.ChecksumMismatchError{
			Expected: h.GetChecksum(),
			Actual:   upload_result.GetFieldByName("checksum").(string)})
	}

	// STEP 2: Create a BackupFile object

	backupfile := make(map[string]interface{})
	backupfile["name"] = remote_name
	backupfile["uri"] = uri
	backupfile["checksum"] = h.GetChecksum()
	err = CreateModel(conn, descriptor, "BackupFile", backupfile)
	if err != nil {
		return err
	}
	conditional_printf(!options.Quiet, "Created backup file %d\n", backupfile["id"])

	// STEP 3: Create a BackupOperation object

	backupop := make(map[string]interface{})
	backupop["operation"] = "restore"
	backupop["file_id"] = backupfile["id"]
	err = CreateModel(conn, descriptor, "BackupOperation", backupop)
	if err != nil {
		return err
	}
	conditional_printf(!options.Quiet, "Created backup-restore operation id=%d uuid=%s\n", backupop["id"], backupop["uuid"])

	conditional_printf(!options.Quiet, "Waiting for completion ")

	// STEP 4: Wait for completion

	flags := GM_UNTIL_ENACTED | GM_UNTIL_FOUND | GM_UNTIL_STATUS | Ternary_uint32(options.Quiet, GM_QUIET, 0)
	queries := map[string]string{"uuid": backupop["uuid"].(string)}
	conn, completed_backupop, err := FindModelWithRetry(ctx, conn, descriptor, "BackupOperation", queries, flags)
	if err != nil {
		return err
	}

	defer conn.Close()

	conditional_printf(!options.Quiet, "\n")

	// STEP 5: Show results

	data := make([]BackupOutput, 1)
	data[0].Checksum = upload_result.GetFieldByName("checksum").(string)
	data[0].Chunks = int(upload_result.GetFieldByName("chunks_received").(int32))
	data[0].Bytes = int(upload_result.GetFieldByName("bytes_received").(int32))

	if completed_backupop.GetFieldByName("status") == "restored" {
		data[0].Status = "SUCCESS"
	} else {
		data[0].Status = "FAILURE"
	}

	FormatAndGenerateOutput(&options.OutputOptions, DEFAULT_BACKUP_FORMAT, "{{.Status}}", data)

	return nil
}
