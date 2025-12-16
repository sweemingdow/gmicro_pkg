package sfid

import (
	"github.com/sony/sonyflake/v2"
	"net"
	"sync/atomic"
	"time"
)

const defaultIns uint64 = 32

var (
	flakes  []*sonyflake.Sonyflake
	counter atomic.Uint64
	setting atomic.Bool
	insSize uint64
)

func settingSonyFlake(ins int) {
	if ins == 0 {
		ins = int(defaultIns)
	}

	if setting.CompareAndSwap(false, true) {
		insSize = uint64(ins)
		v, e := lower16BitPrivateIP(net.InterfaceAddrs)

		for i := 0; i < ins; i++ {
			sf, err := sonyflake.New(sonyflake.Settings{
				TimeUnit: time.Millisecond,
				MachineID: func() (int, error) {
					if e != nil {
						return 0, e
					}
					return v + i, nil
				},
			})

			if err != nil {
				panic(err)
			}

			flakes = append(flakes, sf)
		}
	}
}

func sonyFlakeNext() int64 {
	idx := counter.Add(1) % insSize
	id, _ := flakes[idx].NextID()
	return id
}
