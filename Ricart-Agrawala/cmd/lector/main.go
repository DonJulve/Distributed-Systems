package main


import (
    "practica2/ms"
    "practica2/mr"
    "practica2/ra"
    "sync"
    "practica2/gf"
    "os"
    "fmt"
    "strconv"
    "math/rand"
    "time"
    "github.com/DistributedClocks/GoVector/govec"
)

func go_reader(radb *ra.RASharedDB, myFile string) {
  for{
    radb.PreProtocol()
    
    radb.Logger.LogLocalEvent("Reading", govec.GetDefaultLogOptions())
    _ = gf.ReadFile(myFile)
    time.Sleep(time.Duration(1+rand.Intn(2))* time.Second)
    radb.Logger.LogLocalEvent("Read Complete", govec.GetDefaultLogOptions())
    
    radb.PostProtocol()
  }
}

func main() {
 time.Sleep(2 * time.Second)
 meString := os.Args[1]
 fmt.Println("Proceso lector con pid " + meString)
 me, _ := strconv.Atoi(meString)
 myFile := "fichero_" + meString + ".txt"
 usersFile := "../../ms/users.txt"
 gf.CreateFile(myFile)
 
 reqch := make(chan mr.Request)
 repch := make(chan mr.Reply)
 barch := make(chan bool)
 
 messageTypes := []ms.Message{mr.Request{}, mr.Reply{}, mr.Update{},mr.Controller{}}
 msgs := ms.New(me, usersFile, messageTypes)


 go mr.ReceiveMessage(&msgs, myFile, reqch, repch, barch)
 
 radb := ra.New(&msgs, me, usersFile, "read", &reqch, &repch)
 
 msgs.Send(radb.Size(), mr.Controller{})
 <- barch
 var wg sync.WaitGroup
 wg.Add(1)
 go go_reader(radb,myFile)
 wg.Wait()
}