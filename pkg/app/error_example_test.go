package app_test

import (
	"errors"
	"fmt"

	"github.com/ncotds/nco-qoordinator/pkg/app"
)

func ExampleErr() {
	reason1 := errors.New("foo")
	reason2 := errors.New("bar")
	appErr := app.Err(app.ErrCodeValidation, "baz", reason1, reason2)

	fmt.Printf("appErr code: '%s'\n", appErr.Code())
	fmt.Printf("appErr is reason 1: %v\n", errors.Is(appErr, reason1))
	fmt.Printf("appErr is reason 2: %v\n", errors.Is(appErr, reason2))
	// Output:
	// appErr code: 'ERR_VALIDATION'
	// appErr is reason 1: true
	// appErr is reason 2: true
}

func ExampleError_Code() {
	baseErr := app.Error{}
	appErr := app.Err(app.ErrCodeValidation, "foo")

	fmt.Printf("baseErr code: '%s'\n", baseErr.Code())
	fmt.Printf("appErr code: '%s'\n", appErr.Code())
	// Output:
	// baseErr code: ''
	// appErr code: 'ERR_VALIDATION'
}

func ExampleError_Message() {
	baseErr := app.Error{}
	appErr := app.Err(app.ErrCodeValidation, "bar")

	fmt.Printf("baseErr message: '%s'\n", baseErr.Message())
	fmt.Printf("appErr message: '%s'\n", appErr.Message())
	// Output:
	// baseErr message: ''
	// appErr message: 'bar'
}
