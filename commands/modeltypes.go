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
	flags "github.com/jessevdk/go-flags"
	"sort"
)

const (
	DEFAULT_MODELTYPE_LIST_FORMAT = "{{ . }}"
)

type ModelTypeList struct {
	OutputOptions
}

type ModelTypeOpts struct {
	List ModelTypeList `command:"list"`
}

var modelTypeOpts = ModelTypeOpts{}

func RegisterModelTypeCommands(parser *flags.Parser) {
	parser.AddCommand("modeltype", "model type commands", "Commands to query the types of models", &modelTypeOpts)
}

func (options *ModelTypeList) Execute(args []string) error {
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

	FormatAndGenerateOutput(&options.OutputOptions, DEFAULT_MODELTYPE_LIST_FORMAT, "", model_names)

	return nil
}
