package main

import (
	"github.com/in4it/openvpn/pkg/api"
)

func main() {
	config := api.Config{Port: "8080"}
	api.NewServer(config).Start()
}
