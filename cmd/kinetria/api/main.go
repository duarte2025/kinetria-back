package main

import (
	"go.uber.org/fx"
)

var (
	AppName     = "kinetria"
	BuildCommit = "undefined"
	BuildTag    = "undefined"
	BuildTime   = "undefined"
)

func main() {
	fx.New(
		// TODO: Adicionar módulos base quando disponíveis
		// xfx.BaseModule(),
		// xbuild.Module(AppName, BuildCommit, BuildTime, BuildTag),
		// xlog.Module(),
		// xtelemetry.Module(),
		// xhttp.Module(),
		// xhealth.Module(),

		fx.Provide(
		// TODO: Adicionar providers
		),
	).Run()
}
