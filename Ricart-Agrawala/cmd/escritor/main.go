package main

import (
    "practica2/ms"
    "practica2/ra"
    "practica2/gf"
    "sync"
    "os"
    "practica2/mr"
    "fmt"
    "strconv"
    "math/rand"
    "time"
    "github.com/DistributedClocks/GoVector/govec"
)

func go_escritor(msgs *ms.MessageSystem, radb *ra.RASharedDB, me int,myFile string,text string) {
  for{
    radb.PreProtocol()
    radb.Logger.LogLocalEvent("Writing", govec.GetDefaultLogOptions())
    gf.WriteToFile(myFile,text)
    time.Sleep(time.Duration(1+rand.Intn(2))* time.Second)
    radb.Logger.LogLocalEvent("Write Complete", govec.GetDefaultLogOptions())
    
    for i := 1; i<= radb.Size()-1 ; i++ {
        if i != me {
          msgs.Send(i, mr.Update{text})
        }
    }
    
    radb.PostProtocol()
  }
}

func main()  {
   time.Sleep(2 * time.Second)
   meString := os.Args[1]
   fmt.Println("Proceso escritor con pid " + meString)
   me, _ := strconv.Atoi(meString)
   text := "z"
   myFile := "fichero_" + meString + ".txt"
   usersFile := "../../ms/users.txt"
   gf.CreateFile(myFile)
   
   reqch := make(chan mr.Request)
   repch := make(chan mr.Reply)
   barch := make(chan bool)
   
   messageTypes := []ms.Message{mr.Request{}, mr.Reply{}, mr.Update{},mr.Controller{}}
   msgs := ms.New(me, usersFile, messageTypes)
  
  
   go mr.ReceiveMessage(&msgs, myFile, reqch, repch, barch)
   
   radb := ra.New(&msgs, me, usersFile, "write", &reqch, &repch)
   
   msgs.Send(radb.Size(), mr.Controller{})
   <- barch
   var wg sync.WaitGroup
   wg.Add(1)
   go go_escritor(&msgs,radb,me,myFile,text)
   wg.Wait()
}

