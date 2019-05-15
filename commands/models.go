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
	"github.com/opencord/cordctl/format"
	"sort"
	"strings"
)

const (
	DEFAULT_MODEL_AVAILABLE_FORMAT = "{{ . }}"
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

type ModelAvailable struct {
	OutputOptions
}

type ModelOpts struct {
	List      ModelList      `command:"list"`
	Available ModelAvailable `command:"available"`
}

var modelOpts = ModelOpts{}

func RegisterModelCommands(parser *flags.Parser) {
	parser.AddCommand("model", "model commands", "Commands to query and manipulate XOS models", &modelOpts)
}

func (options *ModelAvailable) Execute(args []string) error {
	conn, descriptor, err := InitReflectionClient()
	if err != nil {
		return err
	}

	defer conn.Close()

	models, err := GetModelNames(descriptor)
	if err != nil {
		return err
	}

	model_names := []string{}
	for k := range models {
		model_names = append(model_names, k)
	}

	sort.Strings(model_names)

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = DEFAULT_MODEL_AVAILABLE_FORMAT
	}

	result := CommandResult{
		Format:   format.Format(outputFormat),
		OutputAs: toOutputType(options.OutputAs),
		Data:     model_names,
	}

	GenerateOutput(&result)

	return nil
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

	var models []*dynamic.Message

	queries, err := CommaSeparatedQueryToMap(options.Filter)
	if err != nil {
		return err
	}

	if len(queries) == 0 {
		models, err = ListModels(conn, descriptor, string(options.Args.ModelName))
	} else {
		models, err = FilterModels(conn, descriptor, string(options.Args.ModelName), queries)
	}
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

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = default_format.String()
	}
	if options.Quiet {
		outputFormat = "{{.Id}}"
	}

	result := CommandResult{
		Format:   format.Format(outputFormat),
		OutputAs: toOutputType(options.OutputAs),
		Data:     data,
	}

	GenerateOutput(&result)
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
