package main

import (
	"context"
	"fmt"

	"avidbound.com/zego/zego"
)

func main() {
	ctx := context.TODO()

	testModule := `
		package test
		a := input.a`

	input := `{"a": "test"}`

	query, err := zego.New(
		zego.Query("zego.test.a"),
		zego.Module("test.zego", testModule),
	).PrepareForEval(ctx)

	if err != nil {
		panic(err)
	}

	rs, err := query.Eval(ctx, zego.EvalInput(input))

	if err != nil {
		panic(err)
	}

	fmt.Println(rs[0])
}
