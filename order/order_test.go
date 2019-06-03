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
