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
	flags "github.com/jessevdk/go-flags"
	"github.com/jhump/protoreflect/dynamic"
	"strings"
)

type ModelNameString string

type ModelList struct {
	OutputOptions
	ShowHidden      bool   `long:"showhidden" description:"Show hidden fields in default output"`
	ShowFeedback    bool   `long:"showfeedback" description:"Show feedback fields in default output"`
	ShowBookkeeping bool   `long:"showbookkeeping" description:"Show bookkeeping fields in default output"`
	Filter          string `long:"filter" description:"Comma-separated list of filters"`
	Args            struct {
		ModelName ModelNameString
	} `positional-args:"yes" required:"yes"`
}

type ModelUpdate struct {
	OutputOptions
	Filter    string `long:"filter" description:"Comma-separated list of filters"`
	SetFields string `long:"set-field" description:"Comma-separated list of field=value to set"`
	SetJSON   string `long:"set-json" description:"JSON dictionary to use for settings fields"`
	Args      struct {
		ModelName ModelNameString
	} `positional-args:"yes" required:"yes"`
	IDArgs struct {
		ID []int32
	} `positional-args:"yes" required:"no"`
}

type ModelOpts struct {
	List   ModelList   `command:"list"`
	Update ModelUpdate `command:"update"`
}

var modelOpts = ModelOpts{}

func RegisterModelCommands(parser *flags.Parser) {
	parser.AddCommand("model", "model commands", "Commands to query and manipulate XOS models", &modelOpts)
}

func (options *ModelList) Execute(args []string) error {
	conn, descriptor, err := InitReflectionClient()
	if err != nil {
		return err
	}

	defer conn.Close()

	err = CheckModelName(descriptor, string(options.Args.ModelName))
	if err != nil {
		return err
	}

	queries, err := CommaSeparatedQueryToMap(options.Filter, true)
	if err != nil {
		return err
	}

	models, err := ListOrFilterModels(conn, descriptor, string(options.Args.ModelName), queries)
	if err != nil {
		return err
	}

	field_names := make(map[string]bool)
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

			field_names[field_name] = true
		}
	}

	var default_format strings.Builder
	default_format.WriteString("table")
	first := true
	for field_name, _ := range field_names {
		if first {
			fmt.Fprintf(&default_format, "{{ .%s }}", field_name)
			first = false
		} else {
			fmt.Fprintf(&default_format, "\t{{ .%s }}", field_name)
		}
	}

	FormatAndGenerateOutput(&options.OutputOptions, default_format.String(), "{{.id}}", data)

	return nil
}

func (options *ModelUpdate) Execute(args []string) error {
	conn, descriptor, err := InitReflectionClient()
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
			models[i], err = GetModel(conn, descriptor, modelName, id)
			if err != nil {
				return err
			}
		}
	} else {
		models, err = ListOrFilterModels(conn, descriptor, modelName, queries)
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

	for _, model := range models {
		fields["id"] = model.GetFieldByName("id").(int32)
		UpdateModel(conn, descriptor, modelName, fields)
	}

	count := len(models)
	FormatAndGenerateOutput(&options.OutputOptions, "{{.}} models updated.", "{{.}}", count)

	return nil
}

func (modelName *ModelNameString) Complete(match string) []flags.Completion {
	conn, descriptor, err := InitReflectionClient()
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
