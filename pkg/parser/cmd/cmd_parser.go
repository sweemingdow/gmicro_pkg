package cmd

import (
	"flag"
	"github.com/sweemingdow/gmicro_pkg/pkg/utils"
)

var DefaultParseEntry = []CmdParseEntry{
	{
		Name:   "http_port",
		DefVal: "8080",
	},
	{

		Name:   "rpc_port",
		DefVal: "8081",
	},
	{

		Name:   "config",
		DefVal: "./configs/config.yaml",
	},
}

type CmdParseEntry struct {
	Name   string
	DefVal string
}

type CmdParser struct {
	resultMap map[string]string
}

func NewCmdParser() *CmdParser {
	return &CmdParser{}
}

// -env=prod or -env prod
func (cp *CmdParser) Parse(args []CmdParseEntry) {
	argsPtr := make([]*string, len(args))
	for idx, arg := range args {
		argsPtr[idx] = flag.String(arg.Name, arg.DefVal, arg.Name)
	}

	flag.Parse()

	argName2value := make(map[string]string, len(args))
	for idx, ptr := range argsPtr {
		argName2value[args[idx].Name] = *ptr
	}

	cp.resultMap = argName2value
}

func (cp *CmdParser) GetInt(name string) int {
	return utils.A2i(cp.resultMap[name])
}

func (cp *CmdParser) GetString(name string) string {
	return cp.resultMap[name]
}

func (cp *CmdParser) GetBool(name string) bool {
	return utils.A2b(cp.resultMap[name])
}
