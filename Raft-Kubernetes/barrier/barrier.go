/*
* NIPS: 848431 (Saúl Caballero Luca) y 840710 (Javier Julve Yubero)
*/
package main

import (
	"bufio"
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"fmt"
)

func readEndpoints(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error Open ",filename)
		fmt.Println(err)
		return nil, err
	}
	defer file.Close()

	var endpoints []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
		if line != "" {
			endpoints = append(endpoints, line)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error")
		return nil, err
	}
	return endpoints, nil
}

func handleConnection(conn net.Conn, endpoints []string, connChan chan bool, barrierChan chan bool, received *map[string]bool, mu *sync.Mutex, muResponse *sync.Mutex, n int,response string) {
	defer conn.Close()
	buf := make([]byte, 1024)
	length, err := conn.Read(buf)
	if err != nil {
		return
	}
	msg := string(buf[:length])
	msg = strings.TrimSpace(msg)
	msg =  strings.ReplaceAll(msg, "\r\n", "")

	isContained := false
	for i := 0; i<n; i++ {
		endpoints[i] = strings.TrimSpace(endpoints[i])
		endpoints[i] =  strings.ReplaceAll(endpoints[i], "\r\n", "")

		if strings.Compare(msg, endpoints[i]) == 0 {
			isContained = true
			break
		}
	}

	if !isContained {
		fmt.Println("Not contained")
		return
	}

	fmt.Println(msg + " recibido")

	mu.Lock()
	(*received)[msg] = true

	if len(*received) == n-1 {
		
		for i := 0; i<n-2; i++ {
			connChan <- true
		}

		mu.Unlock()
	} else {
		mu.Unlock()


		<-connChan
	}

	

	muResponse.Lock()

	
	_, err = conn.Write([]byte(response))
	delete(*received,msg)

	muResponse.Unlock()

	time.Sleep(2 * time.Second)

	conn.Close()

	if err != nil {
		fmt.Println(err)
		return
	}

	muResponse.Lock()

	if len(*received) == 0 {
		
		barrierChan <- true

	}

	muResponse.Unlock()
}

// Get enpoints (IP adresse:port for each distributed process)
func getEndpoints() ([]string, int, error) {
	endpointsFile := os.Args[1]
	lineNumber, err := strconv.Atoi(os.Args[2])
	if err != nil || lineNumber < 1 {
		return nil, 0, errors.New("Invalid line number")
	}

	endpoints, err := readEndpoints(endpointsFile)
	if err != nil {
		return nil, 0, err
	}

	if lineNumber > len(endpoints) {
		return nil, 0, errors.New("Line number out of range")
	}

	return endpoints, lineNumber, nil
}

func acceptAndHandleConnections(listener net.Listener, endpoints []string,connChan chan bool, quitChannel chan bool, barrierChan chan bool, receivedMap *map[string]bool, mu *sync.Mutex, muResponse *sync.Mutex, n int,response string) {
	for {
		select {
		case <-quitChannel:
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			go handleConnection(conn, endpoints,connChan, barrierChan, receivedMap, mu, muResponse, n,response)
		}
	}
}

func main() {
	if len(os.Args) != 4 {
		return
	}

	response := os.Args[3]

	endPoints, lineNumber, err := getEndpoints()
	if err != nil {
		fmt.Println("Error leyendo endpoints")
		return
	}

	localEndpoint := endPoints[lineNumber-1]
	listener, err := net.Listen("tcp", localEndpoint)
	if err != nil {
		fmt.Println("Error creando socket")
		fmt.Println(err)
		return
	}
	defer listener.Close()

	var mu sync.Mutex
	var muResponse sync.Mutex

	quitChannel := make(chan bool)
	receivedMap := make(map[string]bool)
	barrierChan := make(chan bool)
	connChan := make(chan bool)

	go acceptAndHandleConnections(listener, endPoints, connChan,quitChannel, barrierChan, &receivedMap, &mu,&muResponse, len(endPoints), response)

	<-barrierChan

	time.Sleep(2 * time.Second)

	close(quitChannel)

	time.Sleep(2 * time.Second)

}
