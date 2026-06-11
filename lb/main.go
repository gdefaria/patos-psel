package main

import (
	"net"
	"fmt"
	"encoding/json"
	"os"
)

type Config struct {
	Host string `json:"host"`
	Port int `json:"port"`
	Servers []string `json:"servers"`
}

func loadConfig() *Config {
	// carrega o arquivo de configuração
	fileContent, err := os.ReadFile("./config.json")

	if err != nil {
		panic(fmt.Errorf("Failed to open config.json: %w", err))
	}

	var c Config
	err = json.Unmarshal(fileContent, &c)

	if err != nil {
		panic(fmt.Errorf("Failed to parse config.json: %w", err))
	}

	/*
		caso existam campos não setados no arquivo config.json,
		json.Unmarshal usa por padrão a estrutura não inicialidada.
		é importante verificar se o usuário setou os valores necessários.

		note que não é necessário verificar se c.Host != ""
		pois, caso seja uma string vazia, net.Listen 
		escutará em todas as interfaces.
	*/
	if c.Port == 0 {
		panic("[config.json] Invalid port")
	}

	if len(c.Servers) == 0 {
		panic("[config.json] Invalid servers list")
	}

	return &c
}

func main() {
	// carrega o arquivo ./config.json
	config := loadConfig()

	// inicializa o load balancer
	balancer := NewBalancer(config.Servers)

	// inicializa o listener
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Printf("Listening at %s:%d\n", config.Host, config.Port)

	// aceita novas conexões
	for {
		clientConn, err := listener.Accept()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to new client: %v\n", err)
			return
		}

		// instancia uma reverse proxy entre o cliente e o melhor servidor
		go func() {
			serverIndex := balancer.GetBestServer()
			
			balancer.UseServer(serverIndex, func(serverURL string) {
				proxy, err := NewReverseProxy(clientConn, serverURL)

				if err != nil {
					fmt.Fprintf(
						os.Stderr,
						"Failed to setup reverse proxy between (client)[%s] and [%s](server): %v\n",
						clientConn.RemoteAddr().String(),
						serverURL,
						err,
					)

					return
				}

				// redireciona os fluxos entre uma socket e outra
				proxy.Pipe()
			})
		}()
	}

}
