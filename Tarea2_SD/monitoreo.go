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

func nuevoServidorMonitoreo() *servidorMonitoreo {
	s := &servidorMonitoreo{}
	s.cond = sync.NewCond(&s.mutex)
	return s
}

func (s *servidorMonitoreo) AgregarMensaje(mensaje string) {
	s.mutex.Lock()
	s.mensajes = append(s.mensajes, mensaje)
	s.cond.Broadcast()
	s.mutex.Unlock()
}

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
		time.Sleep(5 * time.Second) // Esperar 5 segundos antes de enviar el siguiente
	}
}

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
