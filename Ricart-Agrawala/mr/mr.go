
package mr

import (
 "practica2/ms"
 "practica2/gf"
)

type Request struct {
   Buf []byte
   Pid int
   Op string
}

type Reply struct {}

type Update struct {
 Text string
}

type Controller struct {}

func ReceiveMessage(msgs *ms.MessageSystem, myFile string, reqch chan Request, repch chan Reply, barch chan bool) {
for {
 msg := msgs.Receive()
 switch msgType := msg.(type) {
   case Request:
     reqch <- msgType
   case Reply:
     repch <- msgType
   case Update:
     gf.WriteToFile(myFile, msgType.Text)
   case Controller:
     barch <- true
   }
 }
}
