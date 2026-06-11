/*
Visão geral do balancer
	método de balancemento: least connections (escolhe o servidor com menos conexões)

	entrada:
		lista da URL dos servidores ([]string)
		
	funcionamento:
		GetBestServer() retorna o indice do servidor com menos conexões

		UseServer(serverIndex, callback)
			é um wrapper para incrementar e decrementar o contador de conexões.
			incremento ocorre ao chamar UseServer e o decremento ocorre quando
			callback finaliza.

	uso:
		balancer := NewBalancer(servers)
		serverIndex := balancer.GetBestServer()

		func useServer(serverURL string) {
			...
			// bloqueia até que a conexão seja liberada
		}

		err := balancer.UseServer(serverIndex, useServer)
*/

package main

import (
	"fmt"
)

type balancer struct {
	// connections[i] representa a quantide de conexões no servidor serversURL[i]
	serversURL []string
	connections []int
}

func NewBalancer(serversURL []string) *balancer {
	if len(serversURL) == 0 {
		panic("Servers list can not be empty")
	}

	return &balancer {
		serversURL: serversURL,
		connections: make([]int, len(serversURL)),
	}
}

// retorna o indice do servidor com menos conexões
func (bal balancer) GetBestServer() int {
	least_index := 0	

	for i := 1; i < len(bal.serversURL); i++ {
		if bal.connections[i] < bal.connections[least_index] {
			least_index = i
		}
	}

	return least_index
}

/*
	callback é usado para garantir que o decremento ocorre apenas quando
	a função que lida com a proxy entre cliente e servidor termina.

	assim, não há risco do decremento ocorrer mais de uma vez, ou
	por outra parte do cógido que não foi responsável pelo incremento.
*/
func (bal balancer) UseServer(serverIndex int, usageCallback func(serverURL string)) (error) {
	// verifica se o servidor é válido
	if serverIndex >= len(bal.serversURL) || serverIndex < 0 {
		return fmt.Errorf("Server index out of range: %d", serverIndex)
	}

	// chama o callback que lida com a proxy entre cliente e servidor
	bal.connections[serverIndex]++
 	usageCallback(bal.serversURL[serverIndex]) // bloqueia até finalizar

	// callback finalizou
	bal.connections[serverIndex]--

	return nil
}
