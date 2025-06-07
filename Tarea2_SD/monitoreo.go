package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	pb "Tarea2_SD/emergencia"

	"github.com/streadway/amqp"
	"google.golang.org/grpc"
)

type servidorMonitoreo struct {
	pb.UnimplementedMonitoreoServer
	mensajes []string
	mutex    sync.Mutex
	cond     *sync.Cond
}

// nuevoServidorMonitoreo crea e inicializa una nueva instancia del servidor de monitoreo
//
// Retorna:
//
//	*servidorMonitoreo: Instancia del servidor inicializada
func nuevoServidorMonitoreo() *servidorMonitoreo {
	s := &servidorMonitoreo{}
	s.cond = sync.NewCond(&s.mutex)
	return s
}

// AgregarMensaje añade un nuevo mensaje a la lista de forma segura para concurrencia
// y notifica a las goroutines esperando por nuevos mensajes
//
// Parámetros:
//
//	mensaje string: Mensaje a agregar
func (s *servidorMonitoreo) AgregarMensaje(mensaje string) {
	s.mutex.Lock()
	s.mensajes = append(s.mensajes, mensaje)
	s.cond.Broadcast()
	s.mutex.Unlock()
}

// StreamMensajes implementa el servicio gRPC para streaming de mensajes de monitoreo
//
// Flujo de operación:
// 1. Espera nuevos mensajes (bloqueante)
// 2. Cuando llegan nuevos mensajes, los envía por el stream
// 3. Espera 5 segundos entre cada envío
//
// Parámetros:
// _ *pb.Vacio: Parámetro vacío (no usado)
// stream pb.Monitoreo_StreamMensajesServer: Stream gRPC para enviar mensajes
func (s *servidorMonitoreo) StreamMensajes(_ *pb.Vacio, stream pb.Monitoreo_StreamMensajesServer) error {
	indice := 0
	for {
		s.mutex.Lock()
		for len(s.mensajes) == indice {
			s.cond.Wait()
		}
		msg := s.mensajes[indice]
		indice++
		s.mutex.Unlock()

		err := stream.Send(&pb.MensajeMonitoreo{Contenido: msg})
		if err != nil {
			log.Println("Error enviando mensaje de monitoreo:", err)
			return err
		}
		time.Sleep(5 * time.Second)
	}
}

// main inicia el servidor de monitoreo con el siguiente flujo:
// 1. Establece conexión con RabbitMQ
// 2. Crea instancia del servidor de monitoreo
// 3. Inicia goroutine para consumir mensajes de RabbitMQ
// 4. Configura servidor gRPC en puerto 50053
// 5. Inicia servicio de streaming de mensajes

// El programa termina si ocurren errores en:
// - La conexión a RabbitMQ
// - La inicialización del servidor gRPC
func main() {
	conn, err := amqp.Dial("amqp://rodolfo:123@10.10.28.57:5672/")
	if err != nil {
		log.Fatalf("No se pudo conectar a RabbitMQ: %v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("No se pudo abrir canal en RabbitMQ: %v", err)
	}
	ch.QueueDeclare("acciones_dron", false, false, false, false, nil)

	mon := nuevoServidorMonitoreo()

	go func() {
		msgs, _ := ch.Consume("acciones_dron", "", true, false, false, false, nil)
		for m := range msgs {
			mon.AgregarMensaje(string(m.Body))
		}
	}()

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("No se pudo escuchar en el puerto 50053: %v", err)
	}

	srv := grpc.NewServer()
	pb.RegisterMonitoreoServer(srv, mon)

	fmt.Println("Servicio de monitoreo escuchando en puerto 50053...")
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("Fallo al servir gRPC: %v", err)
	}
}
