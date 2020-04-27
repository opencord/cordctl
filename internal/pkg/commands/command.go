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
	"encoding/json"
	"fmt"
	"github.com/opencord/cordctl/internal/pkg/config"
	corderrors "github.com/opencord/cordctl/internal/pkg/error"
	"github.com/opencord/cordctl/pkg/format"
	"github.com/opencord/cordctl/pkg/order"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"strings"
)

type OutputType uint8

const (
	OUTPUT_TABLE OutputType = iota
	OUTPUT_JSON
	OUTPUT_YAML
)

// Make it easy to override output stream for testing
var OutputStream io.Writer = os.Stdout

var CharReplacer = strings.NewReplacer("\\t", "\t", "\\n", "\n")

type OutputOptions struct {
	Format   string `long:"format" value-name:"FORMAT" default:"" description:"Format to use to output structured data"`
	Quiet    bool   `short:"q" long:"quiet" description:"Output only the IDs of the objects"`
	OutputAs string `short:"o" long:"outputas" default:"table" choice:"table" choice:"json" choice:"yaml" description:"Type of output to generate"`
}

type ListOutputOptions struct {
	OutputOptions
	OrderBy string `short:"r" long:"orderby" default:"" description:"Specify the sort order of the results"`
}

func toOutputType(in string) OutputType {
	switch in {
	case "table":
		fallthrough
	default:
		return OUTPUT_TABLE
	case "json":
		return OUTPUT_JSON
	case "yaml":
		return OUTPUT_YAML
	}
}

type CommandResult struct {
	Format   format.Format
	OrderBy  string
	OutputAs OutputType
	Data     interface{}
}

func NewConnection() (*grpc.ClientConn, error) {
	config.ProcessGlobalOptions()
	clientConn, err := grpc.Dial(config.GlobalConfig.Server, grpc.WithInsecure())
	return clientConn, corderrors.RpcErrorToCordError(err)
}

func GenerateOutput(result *CommandResult) {
	if result != nil && result.Data != nil {
		data := result.Data
		if result.OrderBy != "" {
			s, err := order.Parse(result.OrderBy)
			if err != nil {
				panic(err)
			}
			data, err = s.Process(data)
			if err != nil {
				panic(err)
			}
		}
		if result.OutputAs == OUTPUT_TABLE {
			tableFormat := format.Format(result.Format)
			tableFormat.Execute(OutputStream, true, data)
		} else if result.OutputAs == OUTPUT_JSON {
			asJson, err := json.Marshal(&data)
			if err != nil {
				panic(err)
			}
			fmt.Fprintf(OutputStream, "%s", asJson)
		} else if result.OutputAs == OUTPUT_YAML {
			asYaml, err := yaml.Marshal(&data)
			if err != nil {
				panic(err)
			}
			fmt.Fprintf(OutputStream, "%s", asYaml)
		}
	}
}

// Applies common output options to format and generate output
func FormatAndGenerateOutput(options *OutputOptions, default_format string, quiet_format string, data interface{}) {
	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = default_format
	}
	if options.Quiet {
		outputFormat = quiet_format
	}

	result := CommandResult{
		Format:   format.Format(outputFormat),
		OutputAs: toOutputType(options.OutputAs),
		Data:     data,
	}

	GenerateOutput(&result)
}

// Applies common output options to format and generate output
func FormatAndGenerateListOutput(options *ListOutputOptions, default_format string, quiet_format string, data interface{}) {
	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = default_format
	}
	if options.Quiet {
		outputFormat = quiet_format
	}

	result := CommandResult{
		Format:   format.Format(outputFormat),
		OutputAs: toOutputType(options.OutputAs),
		Data:     data,
		OrderBy:  options.OrderBy,
	}

	GenerateOutput(&result)
}
