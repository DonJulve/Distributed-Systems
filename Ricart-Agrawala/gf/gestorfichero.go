package gf

import (
  "fmt"
  "io/ioutil"
  "os"
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func ReadFile(file string) string {
    data, err := ioutil.ReadFile(file)
    
    checkError(err)
    
    return string(data)
}

func WriteToFile(file string, dataToWrite string) {
    f, err := os.OpenFile(file,os.O_APPEND|os.O_WRONLY, 0666)
    
    checkError(err)
    
    defer f.Close()
    
    _, err = f.WriteString(dataToWrite)
    
    checkError(err)
}

func CreateFile(file string) {
    _, err := os.Create(file)
    
    checkError(err)
}