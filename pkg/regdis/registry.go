package regdis

type RegisterParam struct {
	ServiceName string
	Addr        string // 192.168.1.155:8080
	Weight      float32
	Extra       map[string]any
	Metadata    map[string]string
}

type DeregisterParam struct {
	ServiceName string
	Addr        string // 192.168.1.155:8080
	Extra       map[string]any
}

type Registry interface {
	Register(rp RegisterParam) error

	Deregister(dp DeregisterParam) error
}
