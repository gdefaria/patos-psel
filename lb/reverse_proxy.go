/*
Visão geral da reverse proxy
	entrada:
		clientConn (net.Conn) - conexão com cliente recebida com net.Listen

	funcionamento:
		abrir conexão com servidor (serverConn net.Conn)

		ler da conexão com o cliente (<- clientConn)
		escrever na conexão com o servidor o que foi lido (clientConn -> serverConn)

		ler da conexão com o servidor (<- serverConn)
		escrever na conexão com o cliente o que foi lido (serverConn -> clientConn)

	uso:
		reverseProxy, err := NewReverseProxy(clientConn, "host:port")
		(go) reverseProxy.Pipe()
*/

package main

import (
	"net"
	"fmt"
	"io"
)

type reverseProxy struct {
	clientConn net.Conn // recebida por net.Listen
	serverConn net.Conn // criada em NewReverseProxy()
}

func NewReverseProxy(clientConn net.Conn, destination string) (*reverseProxy, error) {
	// abre conexão com servidor de destino
	serverConn, err := net.Dial("tcp", destination)

	if err != nil {
		return nil, fmt.Errorf("Failed to stablish TCP connetcion with server: %w", err)
	}

	return &reverseProxy {
		clientConn: clientConn,
		serverConn: serverConn,
	}, nil
}

// conecta a socket do cliente com a do servidor (bidirecional)
func (r reverseProxy) Pipe() {
	// channel para acompanhar se alguma das conexões encerrou
	connectionClosed := make(chan struct{}, 2)

	// (ler de) servidor -> (escrever em) cliente
	go func() {
		io.Copy(r.clientConn, r.serverConn)
		connectionClosed <- struct{}{}
	}()

	// (ler de) cliente -> (escrever em) servidor
	go func() {
		io.Copy(r.serverConn, r.clientConn)
		connectionClosed <- struct{}{}
	}()

	// log da conexão
	fmt.Printf(
		"Piping (client)[%s] <-> [%s](server)\n",
		r.clientConn.RemoteAddr().String(),
		r.serverConn.RemoteAddr().String(),
	)

	// bloqueia até que uma das conexões seja fechada
	<- connectionClosed

	// fecha a outra conexão, caso esteja aberta
	r.Close()
}

// fecha ambas as sockets
func (r reverseProxy) Close() {
	r.serverConn.Close()
	r.clientConn.Close()

	fmt.Printf(
		"Closed (client)[%s] <-> [%s](server)\n",
		r.clientConn.RemoteAddr().String(),
		r.serverConn.RemoteAddr().String(),
	)
}
