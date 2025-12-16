package sfid

import (
	"github.com/bwmarrin/snowflake"
	"net"
	"sync/atomic"
)

var (
	node      *snowflake.Node
	sfSetting atomic.Bool
)

func settingSnowflake(nodeId int) {
	if nodeId == 0 {
		v, e := lower16BitPrivateIP(net.InterfaceAddrs)
		if e != nil {
			panic(e)
		}
		nodeId = v
	}

	if sfSetting.CompareAndSwap(false, true) {
		n, e := snowflake.NewNode(int64(nodeId))
		if e != nil {
			panic(e)
		}

		node = n
	}
}

func snowflakeNext() int64 {
	return node.Generate().Int64()
}
