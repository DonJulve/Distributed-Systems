/*
* AUTOR: Rafael Tolosana Calasanz
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: ricart-agrawala.go
* DESCRIPCIÓN: Implementación del algoritmo de Ricart-Agrawala Generalizado en Go
*/
package ra

import (
    "practica2/ms"
    "practica2/mr"
    "sync"
    "strconv"
    "github.com/DistributedClocks/GoVector/govec"
    "github.com/DistributedClocks/GoVector/govec/vclock"
)



type Exclusion struct {
   Op1 string
   Op2 string
}



type RASharedDB struct {
    OutRepCnt   int
    ReqCS       bool
    RepDefd     []int
    ms          *ms.MessageSystem
    done        chan bool
    chrep       chan bool
    req chan mr.Request
    rep chan mr.Reply
    op string
    Mutex       sync.Mutex // mutex para proteger concurrencia sobre las variables
    exclude map[Exclusion]bool
    Logger *govec.GoLog
    LoggerTicket *govec.GoLog
    // TODO: completar
}

func (ra RASharedDB) Size () (int){
  return len(ra.ms.GetPeers())
}


func New(messages *ms.MessageSystem,me int, usersFile string,op string,creq *chan mr.Request,crep *chan mr.Reply) (*RASharedDB) {
    Logger := govec.InitGoVector(strconv.Itoa(me), strconv.Itoa(me),govec.GetDefaultConfig())
    LoggerTicket := govec.InitGoVector(strconv.Itoa(me), strconv.Itoa(me),govec.GetDefaultConfig())
    msgs := messages
    ra := RASharedDB{0, false, make([]int, len(msgs.GetPeers())-1), msgs,make(chan bool),  make(chan bool),(*creq),(*crep),op, sync.Mutex{}, make(map[Exclusion]bool),Logger,LoggerTicket}
    ra.exclude[Exclusion{"read", "read"}] = false
    ra.exclude[Exclusion{"read", "write"}] = true
    ra.exclude[Exclusion{"write", "read"}] = true
    ra.exclude[Exclusion{"write", "write"}] = true
    go requestReceived(&ra)
    go replyReceived(&ra)

    // TODO completar
    return &ra
}

//Pre: Verdad
//Post: Realiza  el  PreProtocol  para el  algoritmo de
//      Ricart-Agrawala Generalizado
func (ra *RASharedDB) PreProtocol(){
    // TODO completar

    //ra.Logger.LogLocalEvent("Start PreProtocol", govec.GetDefaultLogOptions())

    ra.Mutex.Lock()

    ra.ReqCS = true

    messagePayload := []byte("sample-payload")

    out := ra.Logger.PrepareSend("Send request", messagePayload, govec.GetDefaultLogOptions())

    *ra.LoggerTicket = *ra.Logger

    ra.Mutex.Unlock()

    N := len(ra.ms.GetPeers())-1
    ra.OutRepCnt = N-1


    for j := 1;j <= N;j++ {
      if ra.ms.GetMe() != j {
        ra.ms.Send(j,mr.Request{out,ra.ms.GetMe(), ra.op})
      }
    }

    <-ra.chrep

    //ra.Logger.LogLocalEvent("End PreProtocol", govec.GetDefaultLogOptions())
}

//Pre: Verdad
//Post: Realiza  el  PostProtocol  para el  algoritmo de
//      Ricart-Agrawala Generalizado
func (ra *RASharedDB) PostProtocol(){
    // TODO completar

    //ra.Logger.LogLocalEvent("Start PostProtocol", govec.GetDefaultLogOptions())

    ra.ReqCS = false

    N := len(ra.ms.GetPeers())-1


    for j := 1;j <= N;j++ {
      if ra.RepDefd[j-1]==1 {
        ra.RepDefd[j-1] = 0
        ra.ms.Send(j,mr.Reply{})
      }
    }

    //ra.Logger.LogLocalEvent("End PostProtocol", govec.GetDefaultLogOptions())
}

func requestReceived(ra *RASharedDB) {
  for {
     request := <- ra.req
     var defer_it bool
     messagePayload := []byte("sample-payload")
     ra.Logger.UnpackReceive("Recive request", request.Buf, &messagePayload,govec.GetDefaultLogOptions())
     vc := ra.LoggerTicket.GetCurrentVC()
     other, _ := vclock.FromBytes(request.Buf)
     ra.Mutex.Lock()
     defer_it = ra.ReqCS && happensBefore(vc, other, ra.ms.GetMe(), request.Pid) && ra.exclude[Exclusion{ra.op, request.Op}]
     ra.Mutex.Unlock()
     if defer_it {
       ra.RepDefd[request.Pid-1] = 1
     } else {
       ra.ms.Send(request.Pid, mr.Reply{})
     }
  }
}





func replyReceived(ra *RASharedDB) {
  for {
    <-ra.rep
    ra.OutRepCnt = ra.OutRepCnt - 1
    if ra.OutRepCnt == 0 {
      ra.chrep <- true
    }
  }
}


func (ra *RASharedDB) Stop(){
    ra.ms.Stop()
    ra.done <- true
}

func happensBefore(vc vclock.VClock, other vclock.VClock, i int, j int) bool {
   if vc.Compare(other, vclock.Descendant) {
     return true
   } else if vc.Compare(other, vclock.Concurrent) {
     return i < j
   } else {
     return false
   }
}
