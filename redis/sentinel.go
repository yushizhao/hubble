package redis

import "github.com/mediocregopher/radix"

func NewSentinelWithPassword(primaryName string, sentinelAddrs []string, SentinelPassword, ServerPassword string) (*radix.Sentinel, error) {
	sentinelConnFunc := func(network, addr string) (radix.Conn, error) {
		return radix.Dial(network, addr,
			// radix.DialTimeout(1*time.Second),
			radix.DialAuthPass(SentinelPassword),
		)
	}

	serverConnFunc := func(network, addr string) (radix.Conn, error) {
		return radix.Dial(network, addr,
			// radix.DialTimeout(1*time.Second),
			radix.DialAuthPass(ServerPassword),
		)
	}

	poolFunc := func(network, addr string) (radix.Client, error) {
		return radix.NewPool(network, addr, 4, radix.PoolConnFunc(serverConnFunc))
	}

	return radix.NewSentinel(primaryName, sentinelAddrs, radix.SentinelConnFunc(sentinelConnFunc), radix.SentinelPoolFunc(poolFunc))
}

func NewSentinels(primaryNames []string, sentinelAddrs []string, SentinelPassword, ServerPassword string) (map[string]*radix.Sentinel, error) {
	namesMapSentinel := make(map[string]*radix.Sentinel)
	var err error

	for _, n := range primaryNames {
		namesMapSentinel[n], err = NewSentinelWithPassword(n, sentinelAddrs, SentinelPassword, ServerPassword)
		if err != nil {
			return namesMapSentinel, err
		}
	}

	return namesMapSentinel, err
}

func GetMaster(sentinel *radix.Sentinel) (radix.Client, error) {
	addr, _ := sentinel.Addrs()
	return sentinel.Client(addr)
}

// always use the first slave
func GetSlave(sentinel *radix.Sentinel) (radix.Client, error) {
	_, addr := sentinel.Addrs()
	return sentinel.Client(addr[0])
}
