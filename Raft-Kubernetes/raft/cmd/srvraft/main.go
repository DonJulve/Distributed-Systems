package main

import (
	//"errors"
	"fmt"
	//"log"
 //"math/rand"
	"net"
	"net/rpc"
	"os"
	"raft/internal/raft"
	"raft/internal/comun/rpctimeout"
	"raft/internal/comun/check"
	"strconv"
	"time"
	"strings"
)


func main() {
	meText := strings.ReplaceAll(os.Args[1], "raft-", "")
	// obtener entero de indice de este nodo
	me, err := strconv.Atoi(meText)
	check.CheckError(err, "Main, mal numero entero de indice de nodo:")
 
 //fmt.Println("Replica escucha en :", me, " de ", os.Args[2:])

	barrierPoint := os.Args[2]

	var nodos []rpctimeout.HostPort
	// Resto de argumento son los end points como strings
	// De todas la replicas-> pasarlos a HostPort
	for _, endPoint := range os.Args[3:] {
		 nodos = append(nodos, rpctimeout.HostPort(endPoint))
	}
 fmt.Println("Replica escucha en :", me, " de ", os.Args[3:])
 
 canalAplicarOperacion := make(chan raft.AplicaOperacion, 1000)
 database := make(map[string]string)
 
 

	// Parte Servidor
	nr := raft.NuevoNodo(nodos, me, canalAplicarOperacion,barrierPoint)
	rpc.Register(nr)
 
   go aplicarOperacion(database,canalAplicarOperacion)
   
   //go someterOperacion(nr)
	
	fmt.Println("Replica escucha en :", me, " de ", os.Args[3:])

	l, err := net.Listen("tcp", os.Args[3:][me])
	check.CheckError(err, "Main listen error:")
 fmt.Println("Replica escucha en :", me, " de ", os.Args[3:])

  for{
	  rpc.Accept(l)
  }
}

func aplicarOperacion(database map[string]string, canal chan raft.AplicaOperacion){
  for{
    operacion := <- canal
    
    if operacion.Operacion.Operacion == "leer" {
      operacion.Operacion.Valor = database[operacion.Operacion.Clave]
    } else if operacion.Operacion.Operacion == "escribir" {
      database[operacion.Operacion.Clave] = operacion.Operacion.Valor
      operacion.Operacion.Valor = "ok"
    }
    canal <- operacion
  }
}

func someterOperacion(nr *raft.NodoRaft){
  var reply raft.ResultadoRemoto
  
  time.Sleep(20*time.Second)
  go nr.Nodos[nr.IdLider].CallTimeout("NodoRaft.SometerOperacionRaft",raft.TipoOperacion{"escribir","keyA","a"},&reply,25000*time.Millisecond)
  go nr.Nodos[nr.IdLider].CallTimeout("NodoRaft.SometerOperacionRaft",raft.TipoOperacion{"leer","keyA",""},&reply,25000*time.Millisecond)
  go nr.Nodos[nr.IdLider].CallTimeout("NodoRaft.SometerOperacionRaft",raft.TipoOperacion{"escribir","keyB","b"},&reply,25000*time.Millisecond)
  go nr.Nodos[nr.IdLider].CallTimeout("NodoRaft.SometerOperacionRaft",raft.TipoOperacion{"leer","keyB",""},&reply,25000*time.Millisecond)

}











