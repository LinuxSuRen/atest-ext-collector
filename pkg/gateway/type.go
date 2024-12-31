package gateway

type Gateway struct {
	Servers []Server `yaml:"servers"`
}

type Server struct {
	Protocol string  `yaml:"protocol"`
	Domain   string  `yaml:"domain"`
	Routes   []Route `yaml:"routes"`
}

type Route struct {
	Path      string `yaml:"path"`
	ProxyPass string `yaml:"proxyPass"`
}
