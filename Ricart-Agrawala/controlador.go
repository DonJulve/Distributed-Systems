package main

import (
 "fmt"
 "practica2/mr"
 "practica2/ms"
)

func main() {
 me := 6
 usersFile := "./ms/users.txt"
 messageTypes := []ms.Message{mr.Controller{}}
 msgs := ms.New(me, usersFile, messageTypes)
 for i := 1; i <= 5; i++ {
   _ = msgs.Receive()
   fmt.Printf("%d en el controlador\n", i)
 }
 for j := 1; j <= 5; j++ {
   msgs.Send(j, mr.Controller{})
 }
}
