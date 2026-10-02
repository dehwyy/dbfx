package main

import (
	"github.com/dehwyy/dbfx/pkg/migrate/cli"
	"go.uber.org/fx"
)

func main() {
	fx.New(cli.Module).Run()
}
