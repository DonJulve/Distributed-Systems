package main

import(
  "bufio"
  "os"
  "os/exec"
  "fmt"
  "strings"
  "sync"
  "time"
  "strconv"
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func waitprocess(){
    for{
      time.Sleep(2 * time.Second)
    }
}

func printprocess (cmd exec.Cmd){
  stdout, err := cmd.StdoutPipe()
  cmd.Stderr = cmd.Stdout
  
  checkError(err)
  
  err = cmd.Start()
  checkError(err)
  
  defer cmd.Wait()
  
  for {
    tmp := make([]byte, 1024)
    _, err := stdout.Read(tmp)
    fmt.Print(string(tmp))
    if err != nil {
        break
    }
  }
}

func main() {
  var wg sync.WaitGroup

  file, err := os.Open("./ms/users.txt")
  checkError(err)
  
  
  defer file.Close()

  scanner := bufio.NewScanner(file)
  
  
  i:=1
  
  for scanner.Scan() {
      worker := strings.Split(scanner.Text(),":")
      
      if(i%2!=0){
        wg.Add(1)

        go func(i int, worker []string){
          defer wg.Done()
          cmd := exec.Command("ssh", "a840710@"+worker[0], "cd practica2/cmd/escritor ;  /usr/local/go/bin/go run main.go " + strconv.Itoa(i))
          
          printprocess(*cmd)
        }(i, worker)
        
        
      } else if(i!=6){
        wg.Add(1)

        go func(i int, worker []string){
          defer wg.Done()
          cmd := exec.Command("ssh", "a840710@"+worker[0], "cd practica2/cmd/lector ;  /usr/local/go/bin/go run main.go " + strconv.Itoa(i))
          
          printprocess(*cmd)
        }(i, worker)
        
        
      }else{
        go func(worker []string){
          defer wg.Wait()
          cmd := exec.Command("ssh", "a840710@"+worker[0], "cd practica2 ;  /usr/local/go/bin/go run controlador.go " + strconv.Itoa(i))
          
          printprocess(*cmd)
        }(worker)
        
      }
      
      i++
      
  }
  
  var wgMain sync.WaitGroup
  wgMain.Add(1)
  go waitprocess()
  wgMain.Wait()
   
}