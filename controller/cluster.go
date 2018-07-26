package controller

type Cluster struct {
	APIURL            string
	ConsoleURL        string
	MetricsURL        string
	LoggingURL        string
	AppDNS            string
	CapacityExhausted bool

	User  string
	Token string
}
