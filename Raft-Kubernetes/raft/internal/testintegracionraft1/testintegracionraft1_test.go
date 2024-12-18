package testintegracionraft1

import (
	"raft/internal/comun/check"
	"fmt"
	//"log"
	"math/rand"
	//"os"
	//"path/filepath"
	"strconv"
	"testing"
	"time"

	"raft/internal/despliegue"
	"raft/internal/raft"
	"raft/internal/comun/rpctimeout"
)

const (

	//barrier
	BARRIER = "192.168.3.5"
	PUERTOBARRIER = "31073"

	//hosts
	MAQUINA1      = "192.168.3.6"
	MAQUINA2      = "192.168.3.7"
	MAQUINA3      = "192.168.3.8"

	//puertos
	PUERTOREPLICA1 = "31070"
	PUERTOREPLICA2 = "31071"
	PUERTOREPLICA3 = "31072"

	//nodos replicas
	REPLICA1 = MAQUINA1 + ":" + PUERTOREPLICA1
	REPLICA2 = MAQUINA2 + ":" + PUERTOREPLICA2
	REPLICA3 = MAQUINA3 + ":" + PUERTOREPLICA3

	// paquete main de ejecutables relativos a PATH previo
	EXECREPLICA = "cmd/srvraft/main.go"
	EXECBARRIER = "barrier.go"

	// comandos completo a ejecutar en máquinas remota con ssh. Ejemplo :
	// 				cd $HOME/raft; go run cmd/srvraft/main.go 127.0.0.1:29001

	// Ubicar, en esta constante, nombre de fichero de vuestra clave privada local
	// emparejada con la clave pública en authorized_keys de máquinas remotas

	PRIVKEYFILE = "ed2556"
)

// PATH de los ejecutables de modulo golang de servicio Raft
var PATH_RAFT string = "/home/a840710/practica4/raft"
var PATH_BARRIER string = "/home/a840710/practica4/barrier"

	// go run cmd/srvraft/main.go 0 127.0.0.1:29001 127.0.0.1:29002 127.0.0.1:29003
//var EXECREPLICACMD string = "cd " + PATH + "; /usr/local/go/bin/go run " + EXECREPLICA
var EXECREPLICACMD string = "cd "+PATH_RAFT+";/usr/bin/go mod tidy;/usr/bin/go run " + PATH_RAFT +"/"+EXECREPLICA
var EXEC_BARRIER string = "cd "+PATH_BARRIER+";/usr/bin/go mod tidy;/usr/bin/go run " + PATH_BARRIER +"/"+EXECBARRIER


// TEST primer rango
func TestPrimerasPruebas(t *testing.T) { // (m *testing.M) {
	// <setup code>
	// Crear canal de resultados de ejecuciones ssh en maquinas remotas
	cfg := makeCfgDespliegue(t,
							3,
							BARRIER + ":" + PUERTOBARRIER,
							[]string{REPLICA1, REPLICA2, REPLICA3},
							[]bool{true, true, true})

	// tear down code
	// eliminar procesos en máquinas remotas
	defer cfg.stop()

	// Run test sequence

	// Test1 : No debería haber ningun primario, si SV no ha recibido aún latidos
	t.Run("T1:soloArranqueYparada",
		func(t *testing.T) { cfg.soloArranqueYparadaTest1(t) })

	// Test2 : No debería haber ningun primario, si SV no ha recibido aún latidos
	t.Run("T2:ElegirPrimerLider",
		func(t *testing.T) { cfg.elegirPrimerLiderTest2(t) })

	// Test3: tenemos el primer primario correcto
	t.Run("T3:FalloAnteriorElegirNuevoLider",
		func(t *testing.T) { cfg.falloAnteriorElegirNuevoLiderTest3(t) })

	// Test4: Tres operaciones comprometidas en configuración estable
	t.Run("T4:tresOperacionesComprometidasEstable",
		func(t *testing.T) { cfg.tresOperacionesComprometidasEstable(t) })
}


// TEST primer rango
func TestAcuerdosConFallos(t *testing.T) { // (m *testing.M) {
	// <setup code>
	// Crear canal de resultados de ejecuciones ssh en maquinas remotas
	cfg := makeCfgDespliegue(t,
							3,
							BARRIER + ":" + PUERTOBARRIER,
							[]string{REPLICA1, REPLICA2, REPLICA3},
							[]bool{true, true, true})

	// tear down code
	// eliminar procesos en máquinas remotas
	defer cfg.stop()

	// Test5: Se consigue acuerdo a pesar de desconexiones de seguidor
	t.Run("T5:AcuerdoAPesarDeDesconexionesDeSeguidor ",
		func(t *testing.T) { cfg.AcuerdoApesarDeSeguidor(t) })

	t.Run("T5:SinAcuerdoPorFallos ",
		func(t *testing.T) { cfg.SinAcuerdoPorFallos(t) })

	t.Run("T5:SometerConcurrentementeOperaciones ",
		func(t *testing.T) { cfg.SometerConcurrentementeOperaciones(t) })

}


// ---------------------------------------------------------------------
// 
// Canal de resultados de ejecución de comandos ssh remotos
type canalResultados chan string

func (cr canalResultados) stop() {
	close(cr)

	// Leer las salidas obtenidos de los comandos ssh ejecutados
	for s := range cr {
		fmt.Println(s)
	}
}

// ---------------------------------------------------------------------
// Operativa en configuracion de despliegue y pruebas asociadas
type configDespliegue struct {
	t *testing.T
	conectados []bool
	numReplicas int
	barrier string
	nodosRaft []rpctimeout.HostPort
  idLider int
	cr canalResultados
}

// Crear una configuracion de despliegue
func makeCfgDespliegue(t *testing.T, n int, barrier string,nodosraft []string,
										conectados []bool) *configDespliegue {
	cfg := &configDespliegue{}
	cfg.t = t
	cfg.conectados = conectados
	cfg.numReplicas = n
	cfg.barrier = barrier
	cfg.nodosRaft = rpctimeout.StringArrayToHostPortArray(nodosraft)
  	cfg.idLider = 0
	cfg.cr = make(canalResultados, 2000)
	
	return cfg
}

func (cfg *configDespliegue) stop() {
	//cfg.stopDistributedProcesses()

	time.Sleep(50 * time.Millisecond)

	cfg.cr.stop()
}

// --------------------------------------------------------------------------
// FUNCIONES DE SUBTESTS

// Se pone en marcha una replica ?? - 3 NODOS RAFT
func (cfg *configDespliegue) soloArranqueYparadaTest1(t *testing.T) {
	//t.Skip("SKIPPED soloArranqueYparadaTest1")

	fmt.Println(t.Name(), ".....................")

	cfg.t = t  // Actualizar la estructura de datos de tests para errores

	// Poner en marcha replicas en remoto con un tiempo de espera incluido
	cfg.startDistributedProcesses("OK")

	time.Sleep(3 * time.Second)



  //cfg.obtenerEstadoRemoto(0)
  //cfg.obtenerEstadoRemoto(1)
  //cfg.obtenerEstadoRemoto(2)

	// Comprobar estado replica 0
	cfg.comprobarEstadoRemoto (0, 0, false, -1)

	// Comprobar estado replica 1
	cfg.comprobarEstadoRemoto (1, 0, false, -1)

	// Comprobar estado replica 2
	cfg.comprobarEstadoRemoto (2, 0, false, -1)

	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses()

	fmt.Println(".............", t.Name(), "Superado")
}

// Primer lider en marcha - 3 NODOS RAFT
func (cfg *configDespliegue) elegirPrimerLiderTest2(t *testing.T) {
//	t.Skip("SKIPPED ElegirPrimerLiderTest2")

	fmt.Println(t.Name(), ".....................")

	cfg.startDistributedProcesses("RUN")

	time.Sleep(3 * time.Second)
  
  //time.Sleep(2000 * time.Millisecond)

	// Se ha elegido lider ?
	fmt.Printf("Probando lider en curso\n")
	cfg.pruebaUnLider(3)


	// Parar réplicas alamcenamiento en remoto
	cfg.stopDistributedProcesses()   // Parametros

	fmt.Println(".............", t.Name(), "Superado")
}

// Fallo de un primer lider y reeleccion de uno nuevo - 3 NODOS RAFT
func (cfg *configDespliegue) falloAnteriorElegirNuevoLiderTest3(t *testing.T) {
//	t.Skip("SKIPPED FalloAnteriorElegirNuevoLiderTest3")

	fmt.Println(t.Name(), ".....................")

	cfg.startDistributedProcesses("RUN")

	time.Sleep(3 * time.Second)
 
  //time.Sleep(2000 * time.Millisecond)

	fmt.Printf("Lider inicial\n")
	cfg.pruebaUnLider(3)


	// Desconectar lider
  cfg.stopLeader()
  cfg.restartDistributedProcesses()
	// ???
 
  time.Sleep(5000 * time.Millisecond)

	fmt.Printf("Comprobar nuevo lider\n")
	cfg.pruebaUnLider(3)
	

	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses()  //parametros

	fmt.Println(".............", t.Name(), "Superado")
}

// 3 operaciones comprometidas con situacion estable y sin fallos - 3 NODOS RAFT
func (cfg *configDespliegue) tresOperacionesComprometidasEstable(t *testing.T) {
//	t.Skip("SKIPPED tresOperacionesComprometidasEstable")
 
  fmt.Println(t.Name(), ".....................")

	cfg.startDistributedProcesses("RUN")

	time.Sleep(3 * time.Second)
 
  //time.Sleep(2000 * time.Millisecond)
  
  fmt.Printf("Lider inicial\n")
 
  cfg.pruebaUnLider(3)
  
  fmt.Printf("Escritura 1\n")
  cfg.comprobarOperacion(0,"escribir","keyA","a","ok")
  fmt.Printf("Escritura 2\n")
  cfg.comprobarOperacion(1,"escribir","keyB","b","ok")
  fmt.Printf("Lectura\n")
  cfg.comprobarOperacion(2,"leer","keyA","","a")
  fmt.Printf("Todo correcto\n")
 
  
  cfg.stopDistributedProcesses()  //parametros

	fmt.Println(".............", t.Name(), "Superado")
 
 

	// A completar ???
}

// Se consigue acuerdo a pesar de desconexiones de seguidor -- 3 NODOS RAFT
func(cfg *configDespliegue) AcuerdoApesarDeSeguidor(t *testing.T) {
//	t.Skip("SKIPPED AcuerdoApesarDeSeguidor")

	// A completar ???

	// Comprometer una entrada

	cfg.t = t  // Actualizar la estructura de datos de tests para errores

	//  Obtener un lider y, a continuación desconectar una de los nodos Raft
 
  fmt.Println(t.Name(), ".....................")

	cfg.startDistributedProcesses("RUN")

	time.Sleep(3 * time.Second)
 
  //time.Sleep(2000 * time.Millisecond)
  
  fmt.Printf("Lider inicial\n")
 
  cfg.pruebaUnLider(3)
  
  fmt.Printf("Escritura 1\n")
  cfg.comprobarOperacion(0,"escribir","keyA","a","ok")
  
  
  cfg.stopRandomNode()
  time.Sleep(2*time.Second)
  fmt.Printf("Comprobando Lider actual\n")
  cfg.pruebaUnLider(3)
  
  fmt.Printf("Escritura 2\n")
  cfg.comprobarOperacion(1,"escribir","keyB","b","ok")
  fmt.Printf("Escritura 3\n")
  cfg.comprobarOperacion(2,"escribir","keyA","aa","ok")
  fmt.Printf("Lectura 1\n")
  cfg.comprobarOperacion(3,"leer","keyB","","b")
  
  fmt.Printf("Reconectando procesos caidos\n")
  cfg.restartDistributedProcesses()
  fmt.Printf("Escritura 4\n")
  cfg.comprobarOperacion(4,"escribir","keyB","bb","ok")
  fmt.Printf("Lectura 2\n")
  cfg.comprobarOperacion(5,"leer","keyA","","aa")
  fmt.Printf("Escritura 5\n")
  cfg.comprobarOperacion(6,"escribir","keyB","bbb","ok")
  fmt.Printf("Lectura 3\n")
  cfg.comprobarOperacion(7,"leer","keyB","","bbb")
  
  
  
  fmt.Printf("Todo correcto\n")
 
  
  cfg.stopDistributedProcesses()  //parametros

	fmt.Println(".............", t.Name(), "Superado")


	// Comprobar varios acuerdos con una réplica desconectada

	// reconectar nodo Raft previamente desconectado y comprobar varios acuerdos
}

// NO se consigue acuerdo al desconectarse mayoría de seguidores -- 3 NODOS RAFT
func(cfg *configDespliegue) SinAcuerdoPorFallos(t *testing.T) {
//	t.Skip("SKIPPED SinAcuerdoPorFallos")

	// A completar ???

	// Comprometer una entrada

	//  Obtener un lider y, a continuación desconectar 2 de los nodos Raft

  fmt.Println(t.Name(), ".....................")

	cfg.startDistributedProcesses("RUN")

	time.Sleep(3 * time.Second)
 
  //time.Sleep(2000 * time.Millisecond)
  
  fmt.Printf("Lider inicial\n")
 
  cfg.pruebaUnLider(3)
  
  fmt.Printf("Escritura 1\n")
  cfg.comprobarOperacion(0,"escribir","keyA","a","ok")
  
  fmt.Printf("Tirando procesos followers\n")
  cfg.stopFollowers()
  
  cfg.someterOperacionFallo("escribir","keyA","aa")
  cfg.someterOperacionFallo("escribir","keyB","b")
  cfg.someterOperacionFallo("lectura","keyA","")
  
  
  fmt.Printf("Reconectando procesos caidos\n")
  cfg.restartDistributedProcesses()
  
  fmt.Printf("Escritura 2\n")
  cfg.comprobarOperacion(4,"escribir","keyB","b","ok")
  fmt.Printf("Escritura 3\n")
  cfg.comprobarOperacion(5,"escribir","keyC","c","ok")
  fmt.Printf("Lectura 1\n")
  cfg.comprobarOperacion(6,"leer","keyA","","aa")
  
  fmt.Printf("Todo correcto\n")
 
  
  cfg.stopDistributedProcesses()  //parametros

	fmt.Println(".............", t.Name(), "Superado")

	// Comprobar varios acuerdos con 2 réplicas desconectada

	// reconectar lo2 nodos Raft  desconectados y probar varios acuerdos
}

// Se somete 5 operaciones de forma concurrente -- 3 NODOS RAFT
func(cfg *configDespliegue) SometerConcurrentementeOperaciones(t *testing.T) {
	//t.Skip("SKIPPED SometerConcurrentementeOperaciones")

	// A completar ???

	// un bucle para estabilizar la ejecucion

	// Obtener un lider y, a continuación someter una operacion

  fmt.Println(t.Name(), ".....................")

	cfg.startDistributedProcesses("RUN")

	time.Sleep(3 * time.Second)
 
  //time.Sleep(2000 * time.Millisecond)
  
  fmt.Printf("Lider inicial\n")
 
  cfg.pruebaUnLider(3)
  
  fmt.Printf("Escritura 1\n")
  cfg.comprobarOperacion(0,"escribir","keyA","a","ok")
  
  go cfg.someterOperacion("escribir","keyA","a")
  go cfg.someterOperacion("lectura","keyA","")
  go cfg.someterOperacion("escribir","keyB","b")
  go cfg.someterOperacion("lectura","keyB","")
  go cfg.someterOperacion("escribir","keyC","c")
  
  time.Sleep(40*time.Second)
  
  cfg.comprobarEstadoRegistro(5)
  
  
  fmt.Printf("Todo correcto\n")
 
  
  cfg.stopDistributedProcesses()  //parametros

	fmt.Println(".............", t.Name(), "Superado")

	// Someter 5  operaciones concurrentes

	// Comprobar estados de nodos Raft, sobre todo
	// el avance del mandato en curso e indice de registro de cada uno
	// que debe ser identico entre ellos
}




// --------------------------------------------------------------------------
// FUNCIONES DE APOYO
// Comprobar que hay un solo lider
// probar varias veces si se necesitan reelecciones
func (cfg *configDespliegue) pruebaUnLider(numreplicas int) int {
	for iters := 0; iters < 10; iters++ {
		time.Sleep(500 * time.Millisecond)
		mapaLideres := make(map[int][]int)
		for i := 0; i < numreplicas; i++ {
			if cfg.conectados[i] {
				if _, mandato, eslider, _ := cfg.obtenerEstadoRemoto(i);
																	  eslider {
					mapaLideres[mandato] = append(mapaLideres[mandato], i)
				}
			}
		}

		ultimoMandatoConLider := -1
		for mandato, lideres := range mapaLideres {
			if len(lideres) > 1 {
				cfg.t.Fatalf("mandato %d tiene %d (>1) lideres",
														mandato, len(lideres))
			}
			if mandato > ultimoMandatoConLider {
				ultimoMandatoConLider = mandato
			}
		}

		if len(mapaLideres) != 0 {
			
      cfg.idLider = mapaLideres[ultimoMandatoConLider][0]
			return mapaLideres[ultimoMandatoConLider][0]  // Termina
			
		}
	}
	cfg.t.Fatalf("un lider esperado, ninguno obtenido")
	
	return -1   // Termina
}

func (cfg *configDespliegue) obtenerEstadoRemoto(
										indiceNodo int) (int, int, bool, int) {
	var reply raft.EstadoRemoto
	err := cfg.nodosRaft[indiceNodo].CallTimeout("NodoRaft.ObtenerEstadoNodo",
								raft.Vacio{}, &reply, 5000 * time.Millisecond)
	check.CheckError(err, "Error en llamada RPC ObtenerEstadoRemoto")
 
  cfg.t.Log("Estado replica: ", reply.IdNodo, reply.Mandato, reply.EsLider, reply.IdLider, "\n")

	return reply.IdNodo, reply.Mandato, reply.EsLider, reply.IdLider
}

func (cfg *configDespliegue) obtenerEstadoRegistro(indiceNodo int)(int,int){
  var reply raft.EstadoRegistro
  err := cfg.nodosRaft[indiceNodo].CallTimeout("NodoRaft.ObtenerEstadoRegistro",raft.Vacio{},&reply,5000*time.Millisecond)
  check.CheckError(err,"Error en la llamada RPC ObtenerEstadoRegistro")
  return reply.Index,reply.Term
}

// start  gestor de vistas; mapa de replicas y maquinas donde ubicarlos;
// y lista clientes (host:puerto)
func (cfg *configDespliegue) startDistributedProcesses(response string) {
 //cfg.t.Log("Before starting following distributed processes: ", cfg.nodosRaft)

 //fmt.Println(EXEC_BARRIER + " endpoints.txt 1")

 despliegue.ExecMutipleHosts( EXEC_BARRIER +
	" endpoints.txt 1 "+response,
	[]string{BARRIER}, cfg.cr, PRIVKEYFILE)

	time.Sleep(10 * time.Second)

	for i, endPoint := range cfg.nodosRaft {
		despliegue.ExecMutipleHosts( EXECREPLICACMD +
								" " + strconv.Itoa(i) + " " + cfg.barrier + " "+
								rpctimeout.HostPortArrayToString(cfg.nodosRaft),
								[]string{endPoint.Host()}, cfg.cr, PRIVKEYFILE)

								cfg.conectados[i] = true
                                                               
     /*fmt.Println(EXECREPLICACMD +
								" " + strconv.Itoa(i) + " " +
								rpctimeout.HostPortArrayToString(cfg.nodosRaft),
								[]string{endPoint.Host()})*/

		// dar tiempo para se establezcan las replicas
		//time.Sleep(1500 * time.Millisecond)
	}

	// aproximadamente 500 ms para cada arranque por ssh en portatil
	time.Sleep(10000 * time.Millisecond)
}

func (cfg *configDespliegue) restartDistributedProcesses() {
 //cfg.t.Log("Before starting following distributed processes: ", cfg.nodosRaft)

	for i, endPoint := range cfg.nodosRaft {
		if !cfg.conectados[i]{
			despliegue.ExecMutipleHosts( EXECREPLICACMD +
				" " + strconv.Itoa(i) + " " + cfg.barrier + " "+
				rpctimeout.HostPortArrayToString(cfg.nodosRaft),
				[]string{endPoint.Host()}, cfg.cr, PRIVKEYFILE)

				cfg.conectados[i] = true
                                                                 
       /*fmt.Println(EXECREPLICACMD +
  								" " + strconv.Itoa(i) + " " +
  								rpctimeout.HostPortArrayToString(cfg.nodosRaft),
  								[]string{endPoint.Host()})*/
       cfg.conectados[i] = true
  
  		// dar tiempo para se establezcan las replicas
  		//time.Sleep(1500 * time.Millisecond)
    }
	}

	// aproximadamente 500 ms para cada arranque por ssh en portatil
	time.Sleep(10000 * time.Millisecond)
}

//
func (cfg *configDespliegue) stopDistributedProcesses() {
	var reply raft.Vacio

	for i, endPoint := range cfg.nodosRaft {
		if cfg.conectados[i]{
			err := endPoint.CallTimeout("NodoRaft.ParaNodo",
									raft.Vacio{}, &reply, 10 * time.Millisecond)
			check.CheckError(err, "Error en llamada RPC Para nodo")
		}
	}
}

func (cfg *configDespliegue) stopLeader(){
  var reply raft.Vacio
  for i, endPoint := range cfg.nodosRaft{
    if i== cfg.idLider{
      err := endPoint.CallTimeout("NodoRaft.ParaNodo",raft.Vacio{},&reply,20*time.Millisecond)
      check.CheckError(err,"Error en la llamada RPC Para nodo")
      cfg.conectados[i] = false
    }
  }
}

func (cfg *configDespliegue) stopFollowers(){
  var reply raft.Vacio
  for i, endPoint := range cfg.nodosRaft{
    if i != cfg.idLider{
      err := endPoint.CallTimeout("NodoRaft.ParaNodo",raft.Vacio{},&reply,20*time.Millisecond)
      check.CheckError(err,"Error en la llamada RPC Para nodo")
      cfg.conectados[i] = false
    }
  }
}

func (cfg *configDespliegue) stopRandomNode(){
  var reply raft.Vacio
  
  i := rand.Intn(3)
  
  err := cfg.nodosRaft[i].CallTimeout("NodoRaft.ParaNodo",raft.Vacio{},&reply,20*time.Millisecond)
  check.CheckError(err,"Error en la llamada RPC Para nodo")
  cfg.conectados[i] = false
  fmt.Printf("Nodo %d parado\n",i)
}



// Comprobar estado remoto de un nodo con respecto a un estado prefijado
func (cfg *configDespliegue) comprobarEstadoRemoto(idNodoDeseado int,
				 mandatoDeseado int, esLiderDeseado bool, IdLiderDeseado int) {
	idNodo, mandato, esLider, idLider := cfg.obtenerEstadoRemoto(idNodoDeseado)

	
  //cfg.t.Log("Estado replica: ", idNodo, mandato, esLider, idLider, "\n")

	if idNodo != idNodoDeseado || mandato != mandatoDeseado ||
						esLider != esLiderDeseado || idLider != IdLiderDeseado {
	  cfg.t.Fatalf("Estado incorrecto en replica %d en subtest %s",
													idNodoDeseado, cfg.t.Name())
	}

}

func (cfg *configDespliegue) someterOperacion(operacion string,clave string,valor string) (int,int,bool,int,string){
  
  var reply raft.ResultadoRemoto
  fmt.Printf("Someter operacion a %d\n",cfg.idLider)
  err := cfg.nodosRaft[cfg.idLider].CallTimeout("NodoRaft.SometerOperacionRaft",raft.TipoOperacion{operacion,clave,valor},&reply,60000*time.Millisecond)
  check.CheckError(err, "Error en llamada RPC SometerOperacionRaft")
  
  cfg.idLider = reply.IdLider
  
  return reply.IndiceRegistro, reply.Mandato, reply.EsLider, reply.IdLider, reply.ValorADevolver
}

func (cfg *configDespliegue) someterOperacionFallo(operacion string,clave string,valor string){
  
  var reply raft.ResultadoRemoto
  fmt.Printf("Someter operacion a %d\n",cfg.idLider)
  err := cfg.nodosRaft[cfg.idLider].CallTimeout("NodoRaft.SometerOperacionRaft",raft.TipoOperacion{operacion,clave,valor},&reply,60000*time.Millisecond)
  //check.CheckError(err, "Error en llamada RPC SometerOperacionRaft")
  
  if err == nil{
    cfg.t.Fatalf("Acuerdo conseguido sin mayoria simple en subtest %s",cfg.t.Name())
  } else{
    fmt.Printf("Acuerdo no conseguido satisfactoriamente\n")
  }
  
  
}

func (cfg *configDespliegue) comprobarOperacion(indiceLog int,operacion string,clave string,valor string,valorDevuelto string){
  
  indice, _,esLider,_, valorADevolver := cfg.someterOperacion(operacion,clave,valor)
  
  if indice != indiceLog || valorADevolver != valorDevuelto{
    cfg.t.Fatalf("Operacion no sometida correctamente en indice %d en subtest %s. %d:%d - %s:%s - %t",indiceLog,cfg.t.Name(),indice,indiceLog,valorADevolver,valorDevuelto,esLider)
  }
}

func (cfg *configDespliegue) comprobarEstadoRegistro(indiceLog int){
  
  indices := make([]int,cfg.numReplicas)
  mandatos := make([]int,cfg.numReplicas)
  
  for i,_:= range cfg.nodosRaft{
    indices[i], mandatos[i] = cfg.obtenerEstadoRegistro(i)
  }
  
  fmt.Print(indices[0]," = ",indiceLog,"\n")
  
  if indices[0] != indiceLog{
    cfg.t.Fatalf("Avance de indice de registro incorrecto en subtest %s",cfg.t.Name())
  }
  
  for i := 1;i<cfg.numReplicas;i++ {
    fmt.Print(indices[0]," = ",indices[i]," - ",mandatos[0]," = ",mandatos[i],"\n")
    if indices[0] != indices[i] || mandatos[0] != mandatos[i] {
      cfg.t.Fatalf("No coincide el estado en todos los nodos en subtest %s",cfg.t.Name())
    }
  }
}