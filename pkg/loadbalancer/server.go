package loadbalancer

type ServerInstance struct {
	ID     string
	Host   string
	Port   string
	Active bool
}
