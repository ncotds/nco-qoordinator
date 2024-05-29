package app_test

import (
	"context"
	"fmt"

	"github.com/ncotds/nco-qoordinator/pkg/app"
)

func ExampleWithLogAttrs() {
	ctxFirst := app.WithLogAttrs(context.Background(), app.Attrs{"k0": 0})

	ctx := ctxFirst
	ctx = app.WithLogAttrs(ctx, app.Attrs{"k1": 1})
	ctx = app.WithLogAttrs(ctx, app.Attrs{"k2": 2})
	ctx = app.WithLogAttrs(ctx, app.Attrs{"k3": 3})

	ctxFinal := ctx

	fmt.Printf("first ctx: %#v\n", app.LogAttrs(ctxFirst))
	fmt.Printf("last ctx: %#v\n", app.LogAttrs(ctxFinal))
	// Output:
	// first ctx: app.Attrs{"k0":0}
	// last ctx: app.Attrs{"k0":0, "k1":1, "k2":2, "k3":3}
}

func ExampleLogAttrs() {
	ctx := app.WithLogAttrs(context.Background(), app.Attrs{"k1": 1})

	attrs := app.LogAttrs(ctx)

	fmt.Printf("%#v\n", attrs)
	// Output
	// app.Attrs{"k1":1}
}
