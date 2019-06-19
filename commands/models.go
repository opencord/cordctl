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
	"fmt"
	"github.com/fullstorydev/grpcurl"
	flags "github.com/jessevdk/go-flags"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc"
	"sort"
	"strings"
	"time"
)

const (
	DEFAULT_CREATE_FORMAT = "table{{ .Id }}\t{{ .Message }}"
	DEFAULT_DELETE_FORMAT = "table{{ .Id }}\t{{ .Message }}"
	DEFAULT_UPDATE_FORMAT = "table{{ .Id }}\t{{ .Message }}"
	DEFAULT_SYNC_FORMAT   = "table{{ .Id }}\t{{ .Message }}"
)

type ModelNameString string

type ModelList struct {
	ListOutputOptions
	ShowHidden      bool   `long:"showhidden" description:"Show hidden fields in default output"`
	ShowFeedback    bool   `long:"showfeedback" description:"Show feedback fields in default output"`
	ShowBookkeeping bool   `long:"showbookkeeping" description:"Show bookkeeping fields in default output"`
	Filter          string `short:"f" long:"filter" description:"Comma-separated list of filters"`
	State           string `short:"s" long:"state" description:"Filter model state [DEFAULT | ALL | DIRTY | DELETED | DIRTYPOL | DELETEDPOL]"`
	Args            struct {
		ModelName ModelNameString
	} `positional-args:"yes" required:"yes"`
}

type ModelUpdate struct {
	OutputOptions
	Unbuffered  bool          `short:"u" long:"unbuffered" description:"Do not buffer console output and suppress default output processor"`
	Filter      string        `short:"f" long:"filter" description:"Comma-separated list of filters"`
	SetFields   string        `long:"set-field" description:"Comma-separated list of field=value to set"`
	SetJSON     string        `long:"set-json" description:"JSON dictionary to use for settings fields"`
	Sync        bool          `long:"sync" description:"Synchronize before returning"`
	SyncTimeout time.Duration `long:"synctimeout" default:"600s" description:"Timeout for --sync option"`
	Args        struct {
		ModelName ModelNameString
	} `positional-args:"yes" required:"yes"`
	IDArgs struct {
		ID []int32
	} `positional-args:"yes" required:"no"`
}

type ModelDelete struct {
	OutputOptions
	Unbuffered bool   `short:"u" long:"unbuffered" description:"Do not buffer console output and suppress default output processor"`
	Filter     string `short:"f" long:"filter" description:"Comma-separated list of filters"`
	All        bool   `short:"a" long:"all" description:"Operate on all models"`
	Args       struct {
		ModelName ModelNameString
	} `positional-args:"yes" required:"yes"`
	IDArgs struct {
		ID []int32
	} `positional-args:"yes" required:"no"`
}

type ModelCreate struct {
	OutputOptions
	Unbuffered  bool          `short:"u" long:"unbuffered" description:"Do not buffer console output"`
	SetFields   string        `long:"set-field" description:"Comma-separated list of field=value to set"`
	SetJSON     string        `long:"set-json" description:"JSON dictionary to use for settings fields"`
	Sync        bool          `long:"sync" description:"Synchronize before returning"`
	SyncTimeout time.Duration `long:"synctimeout" default:"600s" description:"Timeout for --sync option"`
	Args        struct {
		ModelName ModelNameString
	} `positional-args:"yes" required:"yes"`
}

type ModelSync struct {
	OutputOptions
	Unbuffered  bool          `short:"u" long:"unbuffered" description:"Do not buffer console output and suppress default output processor"`
	Filter      string        `short:"f" long:"filter" description:"Comma-separated list of filters"`
	SyncTimeout time.Duration `long:"synctimeout" default:"600s" description:"Timeout for synchronization"`
	All         bool          `short:"a" long:"all" description:"Operate on all models"`
	Args        struct {
		ModelName ModelNameString
	} `positional-args:"yes" required:"yes"`
	IDArgs struct {
		ID []int32
	} `positional-args:"yes" required:"no"`
}

type ModelSetDirty struct {
	OutputOptions
	Unbuffered bool   `short:"u" long:"unbuffered" description:"Do not buffer console output and suppress default output processor"`
	Filter     string `short:"f" long:"filter" description:"Comma-separated list of filters"`
	All        bool   `short:"a" long:"all" description:"Operate on all models"`
	Args       struct {
		ModelName ModelNameString
	} `positional-args:"yes" required:"yes"`
	IDArgs struct {
		ID []int32
	} `positional-args:"yes" required:"no"`
}

type ModelOpts struct {
	List     ModelList     `command:"list"`
	Update   ModelUpdate   `command:"update"`
	Delete   ModelDelete   `command:"delete"`
	Create   ModelCreate   `command:"create"`
	Sync     ModelSync     `command:"sync"`
	SetDirty ModelSetDirty `command:"setdirty"`
}

type ModelStatusOutputRow struct {
	Id      interface{} `json:"id"`
	Message string      `json:"message"`
}

type ModelStatusOutput struct {
	Rows       []ModelStatusOutputRow
	Unbuffered bool
}

var modelOpts = ModelOpts{}

func RegisterModelCommands(parser *flags.Parser) {
	parser.AddCommand("model", "model commands", "Commands to query and manipulate XOS models", &modelOpts)
}

// Initialize ModelStatusOutput structure, creating a row for each model that will be output
func InitModelStatusOutput(unbuffered bool, count int) ModelStatusOutput {
	return ModelStatusOutput{Rows: make([]ModelStatusOutputRow, count), Unbuffered: unbuffered}
}

// Update model status output row for the model
//    If unbuffered is set then we will output directly to the console. Regardless of the unbuffered
//    setting, we always update the row, as callers may check that row for status.
// Args:
//    output - ModelStatusOutput struct to update
//    i - index of row to update
//    id - id of model, <nil> if no model exists
//    status - status text to set if there is no error
//    errror - if non-nil, then apply error text instead of status text
//    final - true if successful status should be reported, false if successful status is yet to come

func UpdateModelStatusOutput(output *ModelStatusOutput, i int, id interface{}, status string, err error, final bool) {
	if err != nil {
		if output.Unbuffered {
			fmt.Printf("%v: %s\n", id, HumanReadableError(err))
		}
		output.Rows[i] = ModelStatusOutputRow{Id: id, Message: HumanReadableError(err)}
	} else {
		if output.Unbuffered && final {
			fmt.Println(id)
		}
		output.Rows[i] = ModelStatusOutputRow{Id: id, Message: status}
	}
}

// Convert a user-supplied state filter argument to the appropriate enum name
func GetFilterKind(kindArg string) (string, error) {
	kindMap := map[string]string{
		"default":   FILTER_DEFAULT,
		"all":       FILTER_ALL,
		"dirty":     FILTER_DIRTY,
		"deleted":   FILTER_DELETED,
		"dirtypol":  FILTER_DIRTYPOL,
		"deletedpo": FILTER_DELETEDPOL,
	}

	// If no arg then use default
	if kindArg == "" {
		return kindMap["default"], nil
	}

	val, ok := kindMap[strings.ToLower(kindArg)]
	if !ok {
		return "", fmt.Errorf("Failed to understand model state %s", kindArg)
	}

	return val, nil
}

// Common processing for commands that take a modelname and a list of ids or a filter
func GetIDList(conn *grpc.ClientConn, descriptor grpcurl.DescriptorSource, modelName string, ids []int32, filter string, all bool) ([]int32, error) {
	err := CheckModelName(descriptor, modelName)
	if err != nil {
		return nil, err
	}

	// we require exactly one of ID, --filter, or --all
	exclusiveCount := 0
	if len(ids) > 0 {
		exclusiveCount++
	}
	if filter != "" {
		exclusiveCount++
	}
	if all {
		exclusiveCount++
	}

	if (exclusiveCount == 0) || (exclusiveCount > 1) {
		return nil, fmt.Errorf("Use either an ID, --filter, or --all to specify which models to operate on")
	}

	queries, err := CommaSeparatedQueryToMap(filter, true)
	if err != nil {
		return nil, err
	}

	if len(ids) > 0 {
		// do nothing
	} else {
		models, err := ListOrFilterModels(context.Background(), conn, descriptor, modelName, FILTER_DEFAULT, queries)
		if err != nil {
			return nil, err
		}
		ids = make([]int32, len(models))
		for i, model := range models {
			ids[i] = model.GetFieldByName("id").(int32)
		}
		if len(ids) == 0 {
			return nil, fmt.Errorf("Filter matches no objects")
		} else if len(ids) > 1 {
			if !Confirmf("Filter matches %d objects. Continue [y/n] ? ", len(models)) {
				return nil, fmt.Errorf("Aborted by user")
			}
		}
	}

	return ids, nil
}

func (options *ModelList) Execute(args []string) error {
	conn, descriptor, err := InitClient(INIT_DEFAULT)
	if err != nil {
		return err
	}

	defer conn.Close()

	err = CheckModelName(descriptor, string(options.Args.ModelName))
	if err != nil {
		return err
	}

	filterKind, err := GetFilterKind(options.State)
	if err != nil {
		return err
	}

	queries, err := CommaSeparatedQueryToMap(options.Filter, true)
	if err != nil {
		return err
	}

	models, err := ListOrFilterModels(context.Background(), conn, descriptor, string(options.Args.ModelName), filterKind, queries)
	if err != nil {
		return err
	}

	var field_names []string
	data := make([]map[string]interface{}, len(models))
	for i, val := range models {
		data[i] = make(map[string]interface{})
		for _, field_desc := range val.GetKnownFields() {
			field_name := field_desc.GetName()

			isGuiHidden := strings.Contains(field_desc.GetFieldOptions().String(), "1005:1")
			isFeedback := strings.Contains(field_desc.GetFieldOptions().String(), "1006:1")
			isBookkeeping := strings.Contains(field_desc.GetFieldOptions().String(), "1007:1")

			if isGuiHidden && (!options.ShowHidden) {
				continue
			}

			if isFeedback && (!options.ShowFeedback) {
				continue
			}

			if isBookkeeping && (!options.ShowBookkeeping) {
				continue
			}

			if field_desc.IsRepeated() {
				continue
			}

			data[i][field_name] = val.GetFieldByName(field_name)

			// Every row has the same set of known field names, so it suffices to use the names
			// from the first row.
			if i == 0 {
				field_names = append(field_names, field_name)
			}
		}
	}

	// Sort field names, making sure "id" appears first
	sort.SliceStable(field_names, func(i, j int) bool {
		if field_names[i] == "id" {
			return true
		} else if field_names[j] == "id" {
			return false
		} else {
			return (field_names[i] < field_names[j])
		}
	})

	var default_format strings.Builder
	default_format.WriteString("table")
	for i, field_name := range field_names {
		if i == 0 {
			fmt.Fprintf(&default_format, "{{ .%s }}", field_name)
		} else {
			fmt.Fprintf(&default_format, "\t{{ .%s }}", field_name)
		}
	}

	FormatAndGenerateListOutput(&options.ListOutputOptions, default_format.String(), "{{.id}}", data)

	return nil
}

func (options *ModelUpdate) Execute(args []string) error {
	conn, descriptor, err := InitClient(INIT_DEFAULT)
	if err != nil {
		return err
	}

	defer conn.Close()

	err = CheckModelName(descriptor, string(options.Args.ModelName))
	if err != nil {
		return err
	}

	if (len(options.IDArgs.ID) == 0 && len(options.Filter) == 0) ||
		(len(options.IDArgs.ID) != 0 && len(options.Filter) != 0) {
		return fmt.Errorf("Use either an ID or a --filter to specify which models to update")
	}

	queries, err := CommaSeparatedQueryToMap(options.Filter, true)
	if err != nil {
		return err
	}

	updates, err := CommaSeparatedQueryToMap(options.SetFields, true)
	if err != nil {
		return err
	}

	modelName := string(options.Args.ModelName)

	var models []*dynamic.Message

	if len(options.IDArgs.ID) > 0 {
		models = make([]*dynamic.Message, len(options.IDArgs.ID))
		for i, id := range options.IDArgs.ID {
			models[i], err = GetModel(context.Background(), conn, descriptor, modelName, id)
			if err != nil {
				return err
			}
		}
	} else {
		models, err = ListOrFilterModels(context.Background(), conn, descriptor, modelName, FILTER_DEFAULT, queries)
		if err != nil {
			return err
		}
	}

	if len(models) == 0 {
		return fmt.Errorf("Filter matches no objects")
	} else if len(models) > 1 {
		if !Confirmf("Filter matches %d objects. Continue [y/n] ? ", len(models)) {
			return fmt.Errorf("Aborted by user")
		}
	}

	fields := make(map[string]interface{})

	if len(options.SetJSON) > 0 {
		fields["_json"] = []byte(options.SetJSON)
	}

	for fieldName, value := range updates {
		value = value[1:]
		proto_value, err := TypeConvert(descriptor, modelName, fieldName, value)
		if err != nil {
			return err
		}
		fields[fieldName] = proto_value
	}

	modelStatusOutput := InitModelStatusOutput(options.Unbuffered, len(models))
	for i, model := range models {
		id := model.GetFieldByName("id").(int32)
		fields["id"] = id
		err := UpdateModel(conn, descriptor, modelName, fields)

		UpdateModelStatusOutput(&modelStatusOutput, i, id, "Updated", err, !options.Sync)
	}

	if options.Sync {
		ctx, cancel := context.WithTimeout(context.Background(), options.SyncTimeout)
		defer cancel()
		for i, model := range models {
			id := model.GetFieldByName("id").(int32)
			if modelStatusOutput.Rows[i].Message == "Updated" {
				conditional_printf(!options.Quiet, "Wait for sync: %d ", id)
				conn, _, err = GetModelWithRetry(ctx, conn, descriptor, modelName, id, GM_UNTIL_ENACTED|Ternary_uint32(options.Quiet, GM_QUIET, 0))
				conditional_printf(!options.Quiet, "\n")
				UpdateModelStatusOutput(&modelStatusOutput, i, id, "Enacted", err, true)
			}
		}
	}

	if !options.Unbuffered {
		FormatAndGenerateOutput(&options.OutputOptions, DEFAULT_UPDATE_FORMAT, DEFAULT_UPDATE_FORMAT, modelStatusOutput.Rows)
	}

	return nil
}

func (options *ModelDelete) Execute(args []string) error {
	conn, descriptor, err := InitClient(INIT_DEFAULT)
	if err != nil {
		return err
	}

	defer conn.Close()

	modelName := string(options.Args.ModelName)
	ids, err := GetIDList(conn, descriptor, modelName, options.IDArgs.ID, options.Filter, options.All)
	if err != nil {
		return err
	}

	modelStatusOutput := InitModelStatusOutput(options.Unbuffered, len(ids))
	for i, id := range ids {
		err = DeleteModel(conn, descriptor, modelName, id)
		UpdateModelStatusOutput(&modelStatusOutput, i, id, "Deleted", err, true)
	}

	if !options.Unbuffered {
		FormatAndGenerateOutput(&options.OutputOptions, DEFAULT_DELETE_FORMAT, DEFAULT_DELETE_FORMAT, modelStatusOutput.Rows)
	}

	return nil
}

func (options *ModelCreate) Execute(args []string) error {
	conn, descriptor, err := InitClient(INIT_DEFAULT)
	if err != nil {
		return err
	}

	defer conn.Close()

	err = CheckModelName(descriptor, string(options.Args.ModelName))
	if err != nil {
		return err
	}

	updates, err := CommaSeparatedQueryToMap(options.SetFields, true)
	if err != nil {
		return err
	}

	modelName := string(options.Args.ModelName)

	fields := make(map[string]interface{})

	if len(options.SetJSON) > 0 {
		fields["_json"] = []byte(options.SetJSON)
	}

	for fieldName, value := range updates {
		value = value[1:]
		proto_value, err := TypeConvert(descriptor, modelName, fieldName, value)
		if err != nil {
			return err
		}
		fields[fieldName] = proto_value
	}

	modelStatusOutput := InitModelStatusOutput(options.Unbuffered, 1)

	err = CreateModel(conn, descriptor, modelName, fields)
	UpdateModelStatusOutput(&modelStatusOutput, 0, fields["id"], "Created", err, !options.Sync)

	if options.Sync {
		ctx, cancel := context.WithTimeout(context.Background(), options.SyncTimeout)
		defer cancel()
		if modelStatusOutput.Rows[0].Message == "Created" {
			id := fields["id"].(int32)
			conditional_printf(!options.Quiet, "Wait for sync: %d ", id)
			conn, _, err = GetModelWithRetry(ctx, conn, descriptor, modelName, id, GM_UNTIL_ENACTED|Ternary_uint32(options.Quiet, GM_QUIET, 0))
			conditional_printf(!options.Quiet, "\n")
			UpdateModelStatusOutput(&modelStatusOutput, 0, id, "Enacted", err, true)
		}
	}

	if !options.Unbuffered {
		FormatAndGenerateOutput(&options.OutputOptions, DEFAULT_CREATE_FORMAT, DEFAULT_CREATE_FORMAT, modelStatusOutput.Rows)
	}

	return nil
}

func (options *ModelSync) Execute(args []string) error {
	conn, descriptor, err := InitClient(INIT_DEFAULT)
	if err != nil {
		return err
	}

	defer conn.Close()

	modelName := string(options.Args.ModelName)
	ids, err := GetIDList(conn, descriptor, modelName, options.IDArgs.ID, options.Filter, options.All)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), options.SyncTimeout)
	defer cancel()

	modelStatusOutput := InitModelStatusOutput(options.Unbuffered, len(ids))
	for i, id := range ids {
		conditional_printf(!options.Quiet, "Wait for sync: %d ", id)
		conn, _, err = GetModelWithRetry(ctx, conn, descriptor, modelName, id, GM_UNTIL_ENACTED|Ternary_uint32(options.Quiet, GM_QUIET, 0))
		conditional_printf(!options.Quiet, "\n")
		UpdateModelStatusOutput(&modelStatusOutput, i, id, "Enacted", err, true)
	}

	if !options.Unbuffered {
		FormatAndGenerateOutput(&options.OutputOptions, DEFAULT_SYNC_FORMAT, DEFAULT_SYNC_FORMAT, modelStatusOutput.Rows)
	}

	return nil
}

func (options *ModelSetDirty) Execute(args []string) error {
	conn, descriptor, err := InitClient(INIT_DEFAULT)
	if err != nil {
		return err
	}

	defer conn.Close()

	modelName := string(options.Args.ModelName)
	ids, err := GetIDList(conn, descriptor, modelName, options.IDArgs.ID, options.Filter, options.All)
	if err != nil {
		return err
	}

	modelStatusOutput := InitModelStatusOutput(options.Unbuffered, len(ids))
	for i, id := range ids {
		updateMap := map[string]interface{}{"id": id}
		err := UpdateModel(conn, descriptor, modelName, updateMap)
		UpdateModelStatusOutput(&modelStatusOutput, i, id, "Dirtied", err, true)
	}

	if !options.Unbuffered {
		FormatAndGenerateOutput(&options.OutputOptions, DEFAULT_SYNC_FORMAT, DEFAULT_SYNC_FORMAT, modelStatusOutput.Rows)
	}

	return nil
}

func (modelName *ModelNameString) Complete(match string) []flags.Completion {
	conn, descriptor, err := InitClient(INIT_DEFAULT)
	if err != nil {
		return nil
	}

	defer conn.Close()

	models, err := GetModelNames(descriptor)
	if err != nil {
		return nil
	}

	list := make([]flags.Completion, 0)
	for k := range models {
		if strings.HasPrefix(k, match) {
			list = append(list, flags.Completion{Item: k})
		}
	}

	return list
}
