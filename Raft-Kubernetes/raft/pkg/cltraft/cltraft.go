package main

import (
 "fmt"
 "raft/internal/comun/check"
 "raft/internal/comun/rpctimeout"
 "raft/internal/raft"
 "strconv"
 "time"
)

func main() {
   fmt.Println("Inicio cliente")
   dns := "raft-service.default.svc.cluster.local"
   name := "raft"
   puerto := "6000"
   var direcciones []string
   
   for i := 0; i < 3; i++ {
     nodo := name + "-" + strconv.Itoa(i) + "." + dns + ":" + puerto
     direcciones = append(direcciones, nodo)
   }
   
   var nodos []rpctimeout.HostPort
   
   for _, endPoint := range direcciones {
     nodos = append(nodos, rpctimeout.HostPort(endPoint))
   }

   fmt.Println("Setup realizado")
   
   time.Sleep(10 * time.Second)
   var reply raft.ResultadoRemoto
   operacion1 := raft.TipoOperacion{"escribir", "x", "5"}
   operacion2 := raft.TipoOperacion{"leer", "x", ""}
   fmt.Println("Operacion 1 sometida a 0")
   
   err := nodos[0].CallTimeout("NodoRaft.SometerOperacionRaft", operacion1,&reply, 5000*time.Millisecond)
   check.CheckError(err, "SometerOperacion")
   
   for reply.IdLider == -1 {
     err = nodos[0].CallTimeout("NodoRaft.SometerOperacionRaft", operacion1,&reply, 5000*time.Millisecond)
     check.CheckError(err, "SometerOperacion")
   }
   if !reply.EsLider {
     fmt.Printf("Operacion 1 sometida a %d\n", reply.IdLider)
     err = nodos[reply.IdLider].CallTimeout("NodoRaft.SometerOperacionRaft",operacion1, &reply, 5000*time.Millisecond)
     check.CheckError(err, "SometerOperacion")
   }
   
   fmt.Printf("Valor devuelto: %s\n", reply.ValorADevolver)
   fmt.Printf("Operacion 2 sometida a %d\n", reply.IdLider)
   err = nodos[reply.IdLider].CallTimeout("NodoRaft.SometerOperacionRaft", operacion2, &reply, 5000*time.Millisecond)
   check.CheckError(err, "SometerOperacion")
   fmt.Printf("Valor devuelto: %s\n", reply.ValorADevolver)
}
