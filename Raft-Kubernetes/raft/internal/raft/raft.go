// Escribir vuestro código de funcionalidad Raft en este fichero
//

package raft

//
// API
// ===
// Este es el API que vuestra implementación debe exportar
//
// nodoRaft = NuevoNodo(...)
//   Crear un nuevo servidor del grupo de elección.
//
// nodoRaft.Para()
//   Solicitar la parado de un servidor
//
// nodo.ObtenerEstado() (yo, mandato, esLider)
//   Solicitar a un nodo de elección por "yo", su mandato en curso,
//   y si piensa que es el msmo el lider
//
// nodoRaft.SometerOperacion(operacion interface()) (indice, mandato, esLider)

// type AplicaOperacion


import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
  "math/rand"
	//"crypto/rand"
	"sync"
	"time"
	//"net/rpc"
  "net"
  "strings"

	"raft/internal/comun/rpctimeout"
)


const (
	// Constante para fijar valor entero no inicializado
	IntNOINICIALIZADO = -1

	//  false deshabilita por completo los logs de depuracion
	// Aseguraros de poner kEnableDebugLogs a false antes de la entrega
	kEnableDebugLogs = true

	// Poner a true para logear a stdout en lugar de a fichero
	kLogToStdout = false

	// Cambiar esto para salida de logs en un directorio diferente
	kLogOutputDir = "./logs_raft/"
)

type State int64

const (
	Follower State = 0
	Candidate State = 1
	Leader State = 2
)

type TipoOperacion struct {
	Operacion string  // La operaciones posibles son "leer" y "escribir"
	Clave string
	Valor string    // en el caso de la lectura Valor = ""
}


// A medida que el nodo Raft conoce las operaciones de las  entradas de registro
// comprometidas, envía un AplicaOperacion, con cada una de ellas, al canal
// "canalAplicar" (funcion NuevoNodo) de la maquina de estados 
type AplicaOperacion struct {
	Indice int  // en la entrada de registro
	Operacion TipoOperacion
}

// Tipo de dato Go que representa un solo nodo (réplica) de raft
//
type NodoRaft struct {
	Mux   sync.Mutex       // Mutex para proteger acceso a estado compartido

	// Host:Port de todos los nodos (réplicas) Raft, en mismo orden
	Nodos []rpctimeout.HostPort
	Yo    int           // indice de este nodos en campo array "nodos"
	IdLider int
	// Utilización opcional de este logger para depuración
	// Cada nodo Raft tiene su propio registro de trazas (logs)
	Logger *log.Logger

	// Vuestros datos aqui.
  NodeState State
  Hearthbeat chan bool
  FollowerState chan bool
  LeaderState chan bool
  AplicarOperacion chan AplicaOperacion 
  Committed chan string 

   
  CurrentTerm int
  VotedFor int
  
  //Si es candidato
  NVotes int
  NAnswers int
  
  Log []Entry
  CommitIndex int
  LastApplied int
  NextIndex []int
  MatchIndex []int
  
	
	// mirar figura 2 para descripción del estado que debe mantenre un nodo Raft
}

type Entry struct{
  Index int
  Term int
  Operation TipoOperacion
}



// Creacion de un nuevo nodo de eleccion
//
// Tabla de <Direccion IP:puerto> de cada nodo incluido a si mismo.
//
// <Direccion IP:puerto> de este nodo esta en nodos[yo]
//
// Todos los arrays nodos[] de los nodos tienen el mismo orden

// canalAplicar es un canal donde, en la practica 5, se recogerán las
// operaciones a aplicar a la máquina de estados. Se puede asumir que
// este canal se consumira de forma continúa.
//
// NuevoNodo() debe devolver resultado rápido, por lo que se deberían
// poner en marcha Gorutinas para trabajos de larga duracion
func NuevoNodo(nodos []rpctimeout.HostPort, yo int, 
						canalAplicarOperacion chan AplicaOperacion, barrierPoint string) *NodoRaft {
	nr := &NodoRaft{}
	nr.Nodos = nodos
	nr.Yo = yo
	nr.IdLider = -1
 
  
  nr.NodeState = Follower
  nr.Hearthbeat = make(chan bool)
  nr.FollowerState = make(chan bool)
  nr.LeaderState = make(chan bool)
  
  nr.AplicarOperacion = canalAplicarOperacion
  nr.Committed = make(chan string)

  
   
  nr.CurrentTerm = 0
  nr.VotedFor = -1
  
  nr.NVotes = 0
  nr.NAnswers = 0
  nr.CommitIndex = -1
  nr.LastApplied = -1
  nr.NextIndex = make([]int, 3)
  nr.MatchIndex = make([]int, 3)
 
 

	if kEnableDebugLogs {
		nombreNodo := nodos[yo].Host() + "_" + nodos[yo].Port()
		logPrefix := fmt.Sprintf("%s", nombreNodo)
		
		fmt.Println("LogPrefix: ", logPrefix)

		if kLogToStdout {
			nr.Logger = log.New(os.Stdout, nombreNodo + " -->> ",
								log.Lmicroseconds|log.Lshortfile)
		} else {
			err := os.MkdirAll(kLogOutputDir, os.ModePerm)
			if err != nil {
				panic(err.Error())
			}
			logOutputFile, err := os.OpenFile(fmt.Sprintf("%s/%s.txt",
			  kLogOutputDir, logPrefix), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				panic(err.Error())
			}
			nr.Logger = log.New(logOutputFile, 
						   logPrefix + " -> ", log.Lmicroseconds|log.Lshortfile)
		}
		nr.Logger.Println("logger initialized")
	} else {
		nr.Logger = log.New(ioutil.Discard, "", 0)
	}

	// Añadir codigo de inicialización
  go delay_start(nr, barrierPoint)
  
  go status(nr)
  
	return nr
}

func delay_start(nr *NodoRaft, barrier_point string){
  conn, err := net.Dial("tcp", barrier_point)

  if err != nil {
    fmt.Println("Error connecting to", barrier_point, ":", err)
    time.Sleep(5 * time.Second)
    go raft_execution(nr)
    return
  }

  _, err = conn.Write([]byte(nr.Nodos[nr.Yo]))
  buf := make([]byte, 1024)
	length, _ := conn.Read(buf)
  
  msg := string(buf[:length])
	msg = strings.TrimSpace(msg)
	msg =  strings.ReplaceAll(msg, "\r\n", "")

  if strings.Compare(msg, "RUN") == 0 {
    go raft_execution(nr)
  }
}

func status(nr *NodoRaft){
  for{
    fmt.Print(nr.NodeState," - ")
    fmt.Println(nr.obtenerEstado())
    time.Sleep(3*time.Second)
  }
}

// Metodo Para() utilizado cuando no se necesita mas al nodo
//
// Quizas interesante desactivar la salida de depuracion
// de este nodo
//
func (nr *NodoRaft) para() {
	go func() {time.Sleep(5 * time.Millisecond); os.Exit(0) } ()
}

// Devuelve "yo", mandato en curso y si este nodo cree ser lider
//
// Primer valor devuelto es el indice de este  nodo Raft el el conjunto de nodos 
// la operacion si consigue comprometerse.
// El segundo valor es el mandato en curso
// El tercer valor es true si el nodo cree ser el lider
// Cuarto valor es el lider, es el indice del líder si no es él
func (nr *NodoRaft) obtenerEstado() (int, int, bool, int) {
	var yo int = nr.Yo
	var mandato int = nr.CurrentTerm
	var idLider int = nr.IdLider
  var esLider bool = idLider == yo
	

	return yo, mandato, esLider, idLider
}

// Devuelve el índice de la última entrada consolidada y el mandato de dicha
 // entrada
func (nr *NodoRaft) obtenerEstadoRegistro() (int, int) {
 indice := -1
 mandato := 0

 if len(nr.Log) != 0 {
  indice = nr.CommitIndex
  mandato = nr.Log[indice].Term
 }

 return indice, mandato
}


// El servicio que utilice Raft (base de datos clave/valor, por ejemplo)
// Quiere buscar un acuerdo de posicion en registro para siguiente operacion
// solicitada por cliente.

// Si el nodo no es el lider, devolver falso
// Sino, comenzar la operacion de consenso sobre la operacion y devolver en
// cuanto se consiga
// 
// No hay garantia que esta operacion consiga comprometerse en una entrada de
// de registro, dado que el lider puede fallar y la entrada ser reemplazada
// en el futuro.
// Primer valor devuelto es el indice del registro donde se va a colocar 
// la operacion si consigue comprometerse.
// El segundo valor es el mandato en curso
// El tercer valor es true si el nodo cree ser el lider
// Cuarto valor es el lider, es el indice del líder si no es él
func (nr *NodoRaft) someterOperacion(operacion TipoOperacion) (int, int,
															bool, int, string) {
  //fmt.Println("1")
  //fmt.Println("Lockeo someter")
  nr.Mux.Lock()
	indice := -1
	mandato := -1
	EsLider := nr.IdLider == nr.Yo
	idLider := -1
	valorADevolver := ""
	
  //fmt.Println("2")
	// Vuestro codigo aqui
  if (EsLider) {
    indice = len(nr.Log)
    mandato = nr.CurrentTerm
    
    entry := Entry{indice,mandato,operacion}
    
    nr.Log = append(nr.Log, entry)
    
    //nr.Logger.Printf("(%d,%d,%s,%s,%s)", entry.Index,entry.Term,entry.Operation.Operacion, entry.Operation.Clave,entry.Operation.Valor)
    
    
   // fmt.Println("3")
    idLider = nr.Yo
    
   // fmt.Println("DesLockeo someter")
    nr.Mux.Unlock()
    
   // fmt.Println("\tEsperando confirmacion \n")
    valorADevolver = <- nr.Committed
   // fmt.Println("\tConfirmacion recibida \n")
    
    nr.Logger.Printf("(%d,%d,%s,%s,%s)", entry.Index,entry.Term,entry.Operation.Operacion, entry.Operation.Clave,entry.Operation.Valor)
    
  }else
  {
   // fmt.Println("DesLockeo someter")
    nr.Mux.Unlock()
    idLider = nr.IdLider
  }
  
  
  
	

	return indice, mandato, EsLider, idLider, valorADevolver
}


// -----------------------------------------------------------------------
// LLAMADAS RPC al API
//
// Si no tenemos argumentos o respuesta estructura vacia (tamaño cero)
type Vacio struct{}

func (nr * NodoRaft) ParaNodo(args Vacio, reply *Vacio) error {
	defer nr.para()
	return nil
}

type EstadoParcial struct {
	Mandato	int
	EsLider bool
	IdLider	int
}

type EstadoRemoto struct {
	IdNodo	int
	EstadoParcial
}

func (nr *NodoRaft) ObtenerEstadoNodo(args Vacio, reply *EstadoRemoto) error {
 // fmt.Println("Lockeo estado nodo")
  nr.Mux.Lock()
	reply.IdNodo,reply.Mandato,reply.EsLider,reply.IdLider = nr.obtenerEstado()
 // fmt.Println("DesLockeo estado nodo")
  nr.Mux.Unlock()
	return nil
}

type ResultadoRemoto struct {
	ValorADevolver string
	IndiceRegistro int
	EstadoParcial
}

func (nr *NodoRaft) SometerOperacionRaft(operacion TipoOperacion,
												reply *ResultadoRemoto) error {
//  fmt.Println("Inicion someter raft")
	reply.IndiceRegistro,reply.Mandato, reply.EsLider,
			reply.IdLider,reply.ValorADevolver = nr.someterOperacion(operacion)
	return nil
}

type EstadoRegistro struct {
  Index int
  Term int
}

func (nr *NodoRaft) ObtenerEstadoRegistro(args Vacio, reply *EstadoRegistro) error {
 // fmt.Println("Lockeo estado registro")
  nr.Mux.Lock()
  reply.Index, reply.Term = nr.obtenerEstadoRegistro()
//  fmt.Println("DesLockeo estado registro")
  nr.Mux.Unlock()
  return nil
}

// -----------------------------------------------------------------------
// LLAMADAS RPC protocolo RAFT
//
// Structura de ejemplo de argumentos de RPC PedirVoto.
//
// Recordar
// -----------
// Nombres de campos deben comenzar con letra mayuscula !
//
type ArgsPeticionVoto struct {
	// Vuestros datos aqui
  Term int
  CandidateID int
  LastLogIndex int
  LastLogTerm int
}

// Structura de ejemplo de respuesta de RPC PedirVoto,
//
// Recordar
// -----------
// Nombres de campos deben comenzar con letra mayuscula !
//
//
type RespuestaPeticionVoto struct {
	// Vuestros datos aqui
  Term int
  VoteGranted bool
}

func bestLeader(nr *NodoRaft, lastLogIndex int, lastLogTerm int) bool{
  return (lastLogTerm>nr.Log[len(nr.Log)-1].Term)||(lastLogTerm==nr.Log[len(nr.Log)-1].Term&&lastLogIndex>=len(nr.Log)-1)
}

// Metodo para RPC PedirVoto
//
func (nr *NodoRaft) PedirVoto(peticion *ArgsPeticionVoto,
										reply *RespuestaPeticionVoto) error {
//  fmt.Println("Lockeo pedir voto")
  nr.Mux.Lock()
	// Vuestro codigo aqui
  if((peticion.Term < nr.CurrentTerm)||(peticion.Term == nr.CurrentTerm&&nr.VotedFor!=peticion.CandidateID)){
    reply.Term = nr.CurrentTerm
    reply.VoteGranted = false
  } else if (peticion.Term > nr.CurrentTerm) {
    if len(nr.Log) == 0 || bestLeader(nr,peticion.LastLogTerm,peticion.LastLogIndex){
    
      nr.CurrentTerm = peticion.Term
      nr.VotedFor = peticion.CandidateID
      reply.Term = nr.CurrentTerm
      reply.VoteGranted = true
    
    } else{
        nr.CurrentTerm = peticion.Term
        reply.Term = nr.CurrentTerm
        reply.VoteGranted = false
    }
    
    if nr.NodeState == Candidate || nr.NodeState == Leader {
      nr.FollowerState <- true
    }
  }
//  fmt.Println("DesLockeo pedir voto")
  nr.Mux.Unlock()
	return nil	
}


type ArgAppendEntries struct {
	// Vuestros datos aqui
  Term int
  LeaderID int
  EntryToCommit Entry
  LeaderCommit int
  
  PrevLogIndex int
  PrevLogTerm int
}

type Results struct {
	// Vuestros datos aqui
  Term int
  Success bool
}


// Metodo de tratamiento de llamadas RPC AppendEntries
func (nr *NodoRaft) AppendEntries(args *ArgAppendEntries,
													  results *Results) error {
	// Completar....
 // fmt.Println("Lockeo AppendEntries")
  nr.Mux.Lock()
  
  if(args.Term<nr.CurrentTerm){
    results.Term = nr.CurrentTerm
  }else if(args.Term==nr.CurrentTerm){
  
    nr.IdLider = args.LeaderID
    results.Term = nr.CurrentTerm
    
    if len(nr.Log)==0{
      if args.EntryToCommit != (Entry{}) {
        nr.Log = append(nr.Log, args.EntryToCommit)
      }
      results.Success = true
    }else if !consistentLog(nr,args.PrevLogIndex,args.PrevLogTerm){
      
      results.Success = false
    } else {
      
      if args.EntryToCommit != (Entry{}) {
        nr.Log = nr.Log[0:args.PrevLogIndex+1]
        nr.Log = append(nr.Log, args.EntryToCommit)
      }
      results.Success = true
    }
  
  
    if args.LeaderCommit > nr.CommitIndex {
      
      nr.CommitIndex = min(args.LeaderCommit,len(nr.Log)-1)
    }
  
    
    nr.Hearthbeat <- true
    
  } else {
    nr.IdLider = args.LeaderID
    nr.CurrentTerm = args.Term
    results.Term = nr.CurrentTerm
    if(nr.NodeState == Leader){
      nr.FollowerState <- true
    }else{
      if args.LeaderCommit > nr.CommitIndex {
        nr.CommitIndex = min(args.LeaderCommit,len(nr.Log)-1)
      }
      nr.Hearthbeat <- true
    }
  }
  
 // fmt.Println("DesLockeo AppendEntries")
  nr.Mux.Unlock()

	return nil
}


// ----- Metodos/Funciones a utilizar como clientes
//
//

// Ejemplo de código enviarPeticionVoto
//
// nodo int -- indice del servidor destino en nr.nodos[]
//
// args *RequestVoteArgs -- argumentos par la llamada RPC
//
// reply *RequestVoteReply -- respuesta RPC
//
// Los tipos de argumentos y respuesta pasados a CallTimeout deben ser
// los mismos que los argumentos declarados en el metodo de tratamiento
// de la llamada (incluido si son punteros
//
// Si en la llamada RPC, la respuesta llega en un intervalo de tiempo,
// la funcion devuelve true, sino devuelve false
//
// la llamada RPC deberia tener un timout adecuado.
//
// Un resultado falso podria ser causado por una replica caida,
// un servidor vivo que no es alcanzable (por problemas de red ?),
// una petición perdida, o una respuesta perdida
//
// Para problemas con funcionamiento de RPC, comprobar que la primera letra
// del nombre  todo los campos de la estructura (y sus subestructuras)
// pasadas como parametros en las llamadas RPC es una mayuscula,
// Y que la estructura de recuperacion de resultado sea un puntero a estructura
// y no la estructura misma.
//
func (nr *NodoRaft) enviarPeticionVoto(nodo int, args *ArgsPeticionVoto,
											reply *RespuestaPeticionVoto) bool {
  err := nr.Nodos[nodo].CallTimeout("NodoRaft.PedirVoto",args,reply,20*time.Millisecond)
  
  fmt.Print("\t",nodo," - ",reply.Term," : ",nr.CurrentTerm," - ",reply.VoteGranted,"\n")
  
  if(err!=nil){
    return false
  } else{
    if(reply.Term > nr.CurrentTerm){
      nr.CurrentTerm = reply.Term
      nr.FollowerState <- true
    }else if(reply.VoteGranted){
      nr.NVotes++
      if(nr.NVotes>len(nr.Nodos)/2){
        nr.LeaderState<-true
      }
    }
  }
    
	// Completar....
	
	return true
}

func (nr *NodoRaft) sendHearthbeat(nodo int, args *ArgAppendEntries,
											reply *Results) bool {
  
  err := nr.Nodos[nodo].CallTimeout("NodoRaft.AppendEntries",args,reply,20*time.Millisecond)
  
  if(err!=nil){
    return false
  } else{
    if(reply.Term > nr.CurrentTerm){
   //   fmt.Println("Lockeo sendHearthbeat")
      nr.Mux.Lock()
      nr.CurrentTerm = reply.Term
      nr.IdLider = -1
      nr.FollowerState <- true
  //    fmt.Println("DesLockeo sendHearthbeat")
      nr.Mux.Unlock()
    }
  }
    
	// Completar....
	
	return true
}

func (nr *NodoRaft) newEntry(nodo int,args *ArgAppendEntries,results *Results) bool {
  
  err := nr.Nodos[nodo].CallTimeout("NodoRaft.AppendEntries",args,results,10*time.Millisecond)
  
  if err!=nil {
    return false
  }
  
  
  if results.Success {
  
    
    nr.MatchIndex[nodo] = nr.NextIndex[nodo]
    nr.NextIndex[nodo]++
    
   // fmt.Println("Lockeo newEntry")
    nr.Mux.Lock()
    
    if nr.MatchIndex[nodo] > nr.CommitIndex {
      
      nr.NAnswers++
      
      if nr.NAnswers == len(nr.Nodos)/2 {
        
        nr.CommitIndex++
        nr.NAnswers = 0
      }
    }
 //   fmt.Println("DesLockeo newEntry")
    nr.Mux.Unlock()
    
  } else {
    
    nr.NextIndex[nodo]--
  }
  
  return true
  
}


func raft_execution(nr *NodoRaft){
  for{
      if nr.CommitIndex > nr.LastApplied {
        nr.LastApplied++
        operacion := AplicaOperacion{nr.LastApplied,nr.Log[nr.LastApplied].Operation}
        nr.AplicarOperacion <- operacion
      }
      
      for nr.NodeState == Follower{
          select {
            case <-nr.Hearthbeat:
              nr.NodeState = Follower
            case <- time.After(timeout()):
              nr.IdLider = -1
              nr.NodeState = Candidate
          }
        }
      for nr.NodeState == Candidate{
          if nr.CommitIndex > nr.LastApplied {
            nr.LastApplied++
            operacion := AplicaOperacion{nr.LastApplied,nr.Log[nr.LastApplied].Operation}
            nr.AplicarOperacion <- operacion
          }
          nr.CurrentTerm++
          nr.VotedFor = nr.Yo
          nr.NVotes = 1
          requestVotes(nr)
          timer := time.NewTimer(timeout())
          select{
            case <- nr.Hearthbeat:
              nr.NodeState = Follower
            case <- nr.FollowerState:
              nr.NodeState = Follower
            case <- nr.LeaderState:
              for i:=0;i<len(nr.Nodos);i++{
                if i!= nr.Yo {
                  nr.NextIndex[i] = len(nr.Log)
                  nr.MatchIndex[i] = -1
                }
              }
              nr.NodeState = Leader
            case <- timer.C:
              nr.NodeState = Candidate
          }
        }
      for nr.NodeState == Leader{
          nr.IdLider = nr.Yo
          sendAppendEntries(nr)
          timer := time.NewTimer(50*time.Millisecond)
          select{
            case <- nr.FollowerState:
              nr.NodeState = Follower
            case <- timer.C:
              if nr.CommitIndex > nr.LastApplied {
                nr.LastApplied++
                operacion := AplicaOperacion{nr.LastApplied,nr.Log[nr.LastApplied].Operation}
                nr.AplicarOperacion <- operacion
                
                operacion = <- nr.AplicarOperacion
                nr.Committed <- operacion.Operacion.Valor
              }
              nr.NodeState = Leader
          }
        }
    
  }
}



func requestVotes(nr *NodoRaft){
  var reply RespuestaPeticionVoto
  for i:=0;i<len(nr.Nodos);i++ {
    if(i!=nr.Yo){
      if len(nr.Log)!=0{
        lastLogIndex := len(nr.Log) -1
        lastLogTerm := nr.Log[lastLogIndex].Term
        go nr.enviarPeticionVoto(i,&ArgsPeticionVoto{nr.CurrentTerm,nr.Yo,lastLogIndex,lastLogTerm},&reply)
      }else{
        go nr.enviarPeticionVoto(i,&ArgsPeticionVoto{nr.CurrentTerm,nr.Yo,-1,0},&reply)
      }
    
    
      
    }
  }
}

func sendAppendEntries(nr *NodoRaft){
  var reply Results
  for i:=0;i<len(nr.Nodos);i++{
    if(i!=nr.Yo){
      if len(nr.Log)-1>=nr.NextIndex[i] {
        entry := Entry{nr.NextIndex[i],nr.Log[nr.NextIndex[i]].Term,nr.Log[nr.NextIndex[i]].Operation}
        
        if nr.NextIndex[i] != 0{
          prevLogIndex := nr.NextIndex[i]-1
          prevLogTerm := nr.Log[prevLogIndex].Term
          go nr.newEntry(i,&ArgAppendEntries{nr.CurrentTerm,nr.Yo,entry,nr.CommitIndex,prevLogIndex,prevLogTerm}, &reply)
        } else{
          go nr.newEntry(i,&ArgAppendEntries{nr.CurrentTerm,nr.Yo,entry,nr.CommitIndex,-1,0}, &reply)
        }
      } else {
        if nr.NextIndex[i] != 0{
          prevLogIndex := nr.NextIndex[i]-1
          prevLogTerm := nr.Log[prevLogIndex].Term
          go nr.sendHearthbeat(i,&ArgAppendEntries{nr.CurrentTerm,nr.Yo,Entry{},nr.CommitIndex,prevLogIndex,prevLogTerm}, &reply)
        } else{
          go nr.sendHearthbeat(i,&ArgAppendEntries{nr.CurrentTerm,nr.Yo,Entry{},nr.CommitIndex,-1,0}, &reply)
        }
      }
    
    }
  }
}

func consistentLog(nr *NodoRaft, prevLogIndex int,prevLogTerm int) bool {
  if prevLogIndex > len(nr.Log)-1 {
    return false
  } else if nr.Log[prevLogIndex].Term != prevLogTerm {
    return false
  } else{
    return true
  }
}

func min(a int, b int) int {
  if a<b {
    return a
  }
  return b
}

func timeout()time.Duration{
  return time.Duration(rand.Intn(800)+200)*time.Millisecond
}



