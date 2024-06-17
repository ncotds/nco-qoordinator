package qoordinator

import (
	"testing"

	"github.com/ncotds/nco-qoordinator/tests/e2e/client"
)

func BenchmarkQoordinatorInsert_SingleThread_100rows(b *testing.B) {
	cl, err := NewClient(&client.TestConfig)
	if err != nil {
		b.Fatal("cannot setup client", err.Error())
	}

	client.DoBenchmarkInsert(b, cl, 100, 1)
}

func BenchmarkQoordinatorInsert_MultiThread_100rows(b *testing.B) {
	cl, err := NewClient(&client.TestConfig)
	if err != nil {
		b.Fatal("cannot setup client", err.Error())
	}

	client.DoBenchmarkInsert(b, cl, 100, client.TestConfig.OMNIbus.MaxConnections)
}

func BenchmarkQoordinatorSelect_SingleThread_10rows_100repeat(b *testing.B) {
	cl, err := NewClient(&client.TestConfig)
	if err != nil {
		b.Fatal("cannot setup client", err.Error())
	}

	client.DoBenchmarkSelect(b, cl, 10, 100, 1)
}

func BenchmarkQoordinatorSelect_MultiThread_10rows_100repeat(b *testing.B) {
	cl, err := NewClient(&client.TestConfig)
	if err != nil {
		b.Fatal("cannot setup client", err.Error())
	}

	client.DoBenchmarkSelect(b, cl, 10, 100, client.TestConfig.OMNIbus.MaxConnections)
}

func BenchmarkQoordinatorSelect_SingleThread_100rows_10repeat(b *testing.B) {
	cl, err := NewClient(&client.TestConfig)
	if err != nil {
		b.Fatal("cannot setup client", err.Error())
	}

	client.DoBenchmarkSelect(b, cl, 100, 10, 1)
}

func BenchmarkQoordinatorSelect_MultiThread_100rows_10repeat(b *testing.B) {
	cl, err := NewClient(&client.TestConfig)
	if err != nil {
		b.Fatal("cannot setup client", err.Error())
	}

	client.DoBenchmarkSelect(b, cl, 100, 10, client.TestConfig.OMNIbus.MaxConnections)
}
