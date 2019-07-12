/*
 * Copyright 2019-present Ciena Corporation
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
package order

import (
	"fmt"
	"testing"
)

type SortTestStruct struct {
	One   string
	Two   string
	Three uint64
}

var testList = []SortTestStruct{
	{
		One:   "a",
		Two:   "x",
		Three: 10,
	},
	{
		One:   "a",
		Two:   "c",
		Three: 1,
	},
	{
		One:   "a",
		Two:   "b",
		Three: 2,
	},
	{
		One:   "a",
		Two:   "a",
		Three: 3,
	},
	{
		One:   "b",
		Two:   "a",
		Three: 3,
	},
}

func TestSort(t *testing.T) {
	s, err := Parse("+One,-Two")
	if err != nil {
		t.Errorf("Unable to parse sort specification")
	}
	//fmt.Printf("%#v\n", s)
	o, err := s.Process(testList)
	if err != nil {
		t.Errorf("Sort failed: %s", err.Error())
	}

	fmt.Printf("END: %#v\n", o)
}

func TestSortInt(t *testing.T) {
	s, err := Parse("Three,One")
	if err != nil {
		t.Errorf("Unable to parse sort specification")
	}
	//fmt.Printf("%#v\n", s)
	o, err := s.Process(testList)
	if err != nil {
		t.Errorf("Sort failed: %s", err.Error())
	}

	fmt.Printf("END: %#v\n", o)
}
