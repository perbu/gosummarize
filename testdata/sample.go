package testdata

import (
	"fmt"
	"strings"
)

// PublicStruct is a sample, exported struct with documentation
type PublicStruct struct {
	// PublicField is an exported field
	PublicField  string
	privateField int
}

// StructWithInterface is a struct that implements an interface
type StructWithInterface struct {
	Data []string
}

// PublicInterface is an exported interface
type PublicInterface interface {
	DoSomething() error
	ProcessData(input string) string
}

// PublicConst is an exported constant with documentation
const PublicConst = "This is a public constant"

// MultipleConsts defines multiple constants
const (
	// FirstConst is the first constant
	FirstConst = 1
	// SecondConst is the second constant
	SecondConst  = 2
	privateConst = 3
)

// PublicVar is an exported variable
var PublicVar = map[string]int{
	"one": 1,
	"two": 2,
}

// DoSomething implements the PublicInterface
func (s *StructWithInterface) DoSomething() error {
	if len(s.Data) == 0 {
		return fmt.Errorf("no data available")
	}
	return nil
}

// ProcessData processes the input string and returns a modified version
// It implements the PublicInterface
func (s *StructWithInterface) ProcessData(input string) string {
	s.Data = append(s.Data, input)
	return strings.ToUpper(input)
}

// PublicFunction is an exported function with parameters and return values
// It demonstrates a function with documentation
func PublicFunction(input string, count int) (string, error) {
	if count <= 0 {
		return "", fmt.Errorf("count must be positive")
	}
	return strings.Repeat(input, count), nil
}

func privateFunction() {
	// This function won't be included in the summary
	fmt.Println("This is private")
}
