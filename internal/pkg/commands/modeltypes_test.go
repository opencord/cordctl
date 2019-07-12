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
	"bytes"
	"github.com/opencord/cordctl/pkg/testutils"
	"testing"
)

func TestModelTypeList(t *testing.T) {
	// use `python -m json.tool` to pretty-print json
	expected := `ANIPort
AddressPool
AttWorkflowDriverService
AttWorkflowDriverServiceInstance
AttWorkflowDriverWhiteListEntry
BNGPortMapping
BackupFile
BackupOperation
BandwidthProfile
ComputeServiceInstance
FabricCrossconnectService
FabricCrossconnectServiceInstance
FabricIpAddress
FabricService
Flavor
Image
InterfaceType
KubernetesConfigMap
KubernetesConfigVolumeMount
KubernetesData
KubernetesResourceInstance
KubernetesSecret
KubernetesSecretVolumeMount
KubernetesService
KubernetesServiceInstance
NNIPort
Network
NetworkParameter
NetworkParameterType
NetworkSlice
NetworkTemplate
Node
NodeLabel
NodeToSwitchPort
OLTDevice
ONOSApp
ONOSService
ONUDevice
PONPort
Port
PortBase
PortInterface
Principal
Privilege
RCORDIpAddress
RCORDService
RCORDSubscriber
Role
Service
ServiceAttribute
ServiceDependency
ServiceGraphConstraint
ServiceInstance
ServiceInstanceAttribute
ServiceInstanceLink
ServiceInterface
ServicePort
Site
Slice
Switch
SwitchPort
Tag
TechnologyProfile
TrustDomain
UNIPort
User
VOLTService
VOLTServiceInstance
XOSCore
XOSGuiExtension
`

	got := new(bytes.Buffer)
	OutputStream = got

	var options ModelTypeOpts
	err := options.List.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	testutils.AssertStringEqual(t, got.String(), expected)
}
