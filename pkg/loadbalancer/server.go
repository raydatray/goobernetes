package loadbalancer

type ServerInstance struct {
	ID     string
	Host   string
	Port   int
	Active bool
}
