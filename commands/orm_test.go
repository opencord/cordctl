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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDecodeOperator(t *testing.T) {
	value, operator, invert, err := DecodeOperator("=something")
	assert.Equal(t, value, "something")
	assert.Equal(t, operator, "EQUAL")
	assert.Equal(t, invert, false)
	assert.Equal(t, err, nil)

	value, operator, invert, err = DecodeOperator("!=something")
	assert.Equal(t, value, "something")
	assert.Equal(t, operator, "EQUAL")
	assert.Equal(t, invert, true)
	assert.Equal(t, err, nil)

	value, operator, invert, err = DecodeOperator(">3")
	assert.Equal(t, value, "3")
	assert.Equal(t, operator, "GREATER_THAN")
	assert.Equal(t, invert, false)
	assert.Equal(t, err, nil)

	value, operator, invert, err = DecodeOperator(">=3")
	assert.Equal(t, value, "3")
	assert.Equal(t, operator, "GREATER_THAN_OR_EQUAL")
	assert.Equal(t, invert, false)
	assert.Equal(t, err, nil)

	value, operator, invert, err = DecodeOperator("<3")
	assert.Equal(t, value, "3")
	assert.Equal(t, operator, "LESS_THAN")
	assert.Equal(t, invert, false)
	assert.Equal(t, err, nil)

	value, operator, invert, err = DecodeOperator("<=3")
	assert.Equal(t, value, "3")
	assert.Equal(t, operator, "LESS_THAN_OR_EQUAL")
	assert.Equal(t, invert, false)
	assert.Equal(t, err, nil)
}

func TestCommaSeparatedQueryStringsToMap(t *testing.T) {
	m, err := CommaSeparatedQueryToMap("foo=7,bar!=stuff, x = 5, y= 27", true)
	assert.Equal(t, err, nil)
	assert.Equal(t, m["foo"], "=7")
	assert.Equal(t, m["bar"], "!=stuff")
	assert.Equal(t, m["x"], "= 5")
	assert.Equal(t, m["y"], "= 27")
}

func TestTypeConvert(t *testing.T) {
	conn, descriptor, err := InitReflectionClient()
	assert.Equal(t, err, nil)
	defer conn.Close()

	v, err := TypeConvert(descriptor, "Site", "id", "7")
	assert.Equal(t, err, nil)
	assert.Equal(t, v, int32(7))

	v, err = TypeConvert(descriptor, "Site", "name", "foo")
	assert.Equal(t, err, nil)
	assert.Equal(t, v, "foo")

	v, err = TypeConvert(descriptor, "Site", "enacted", "123.4")
	assert.Equal(t, err, nil)
	assert.Equal(t, v, 123.4)
}

func TestCheckModelName(t *testing.T) {
	conn, descriptor, err := InitReflectionClient()
	assert.Equal(t, err, nil)
	defer conn.Close()

	err = CheckModelName(descriptor, "Slice")
	assert.Equal(t, err, nil)

	err = CheckModelName(descriptor, "DoesNotExist")
	assert.Equal(t, err.Error(), "Model DoesNotExist does not exist. Use `cordctl models available` to get a list of available models")
}

func TestCreateModel(t *testing.T) {
	conn, descriptor, err := InitReflectionClient()
	assert.Equal(t, err, nil)
	defer conn.Close()

	m := make(map[string]interface{})
	m["name"] = "mockslice3"
	m["site_id"] = int32(1)

	err = CreateModel(conn, descriptor, "Slice", m)
	assert.Equal(t, err, nil)

	assert.Equal(t, m["id"], int32(3))
}

func TestUpdateModel(t *testing.T) {
	conn, descriptor, err := InitReflectionClient()
	assert.Equal(t, err, nil)
	defer conn.Close()

	m := make(map[string]interface{})
	m["id"] = int32(1)
	m["name"] = "mockslice1_newname"

	err = UpdateModel(conn, descriptor, "Slice", m)
	assert.Equal(t, err, nil)
}

func TestGetModel(t *testing.T) {
	conn, descriptor, err := InitReflectionClient()
	assert.Equal(t, err, nil)
	defer conn.Close()

	m, err := GetModel(context.Background(), conn, descriptor, "Slice", int32(1))
	assert.Equal(t, err, nil)

	assert.Equal(t, m.GetFieldByName("id").(int32), int32(1))
	assert.Equal(t, m.GetFieldByName("name").(string), "mockslice1")
}

func TestListModels(t *testing.T) {
	conn, descriptor, err := InitReflectionClient()
	assert.Equal(t, err, nil)
	defer conn.Close()

	m, err := ListModels(context.Background(), conn, descriptor, "Slice")
	assert.Equal(t, err, nil)

	assert.Equal(t, len(m), 2)
	assert.Equal(t, m[0].GetFieldByName("id").(int32), int32(1))
	assert.Equal(t, m[0].GetFieldByName("name").(string), "mockslice1")
	assert.Equal(t, m[1].GetFieldByName("id").(int32), int32(2))
	assert.Equal(t, m[1].GetFieldByName("name").(string), "mockslice2")
}

func TestFilterModels(t *testing.T) {
	conn, descriptor, err := InitReflectionClient()
	assert.Equal(t, err, nil)
	defer conn.Close()

	qm := map[string]string{"id": "=1"}

	m, err := FilterModels(context.Background(), conn, descriptor, "Slice", FILTER_DEFAULT, qm)
	assert.Equal(t, err, nil)

	assert.Equal(t, len(m), 1)
	assert.Equal(t, m[0].GetFieldByName("id").(int32), int32(1))
	assert.Equal(t, m[0].GetFieldByName("name").(string), "mockslice1")
}

func TestDeleteModel(t *testing.T) {
	conn, descriptor, err := InitReflectionClient()
	assert.Equal(t, err, nil)
	defer conn.Close()

	err = DeleteModel(conn, descriptor, "Slice", int32(1))
	assert.Equal(t, err, nil)
}
