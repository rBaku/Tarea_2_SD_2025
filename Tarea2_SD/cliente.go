package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	pb "Tarea2_SD/emergencia"

	"google.golang.org/grpc"
)

type Emergencia struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Magnitude int     `json:"magnitude"`
}

func cargarEmergencias(path string) ([]Emergencia, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error al abrir archivo: %v", err)
	}
	defer file.Close()

	var emergencias []Emergencia
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&emergencias); err != nil {
		return nil, fmt.Errorf("error al decodificar JSON: %v", err)
	}

	return emergencias, nil
}

func containsExtinguido(msg string) bool {
	return strings.Contains(msg, "ha sido extinguido")
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Uso: ./cliente emergencia.json")
		return
	}

	emergencias, err := cargarEmergencias(os.Args[1])
	if err != nil {
		log.Fatalf("Error cargando emergencias: %v", err)
	}

	conn1, err := grpc.Dial("10.10.28.57:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("No se pudo conectar al servicio de asignación: %v", err)
	}
	defer conn1.Close()
	client := pb.NewAsignadorClient(conn1)

	conn2, err := grpc.Dial("10.10.28.56:50053", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("No se pudo conectar al servicio de monitoreo: %v", err)
	}
	defer conn2.Close()
	monitoreo := pb.NewMonitoreoClient(conn2)

	ctxMonitoreo, cancelMonitoreo := context.WithCancel(context.Background())
	defer cancelMonitoreo()

	stream, err := monitoreo.StreamMensajes(ctxMonitoreo, &pb.Vacio{})
	if err != nil {
		log.Fatalf("Error conectando con monitoreo: %v", err)
	}

	done := make(chan bool)

	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				if ctxMonitoreo.Err() != nil {
					return
				}
				log.Printf("Error recibiendo actualización de monitoreo: %v", err)
				return
			}
			fmt.Println(msg.Contenido)
			if containsExtinguido(msg.Contenido) {
				done <- true
			}
		}
	}()

	for _, e := range emergencias {
		fmt.Printf("\nEmergencia actual : %s magnitud %d en x=%d , y=%d\n", e.Name, e.Magnitude, int(e.Latitude), int(e.Longitude))

		_, err := client.EnviarEmergencias(context.Background(), &pb.EmergenciasRequest{
			Emergencias: []*pb.Emergencia{
				{
					Name:      e.Name,
					Latitude:  float32(e.Latitude),
					Longitude: float32(e.Longitude),
					Magnitude: int32(e.Magnitude),
				},
			},
		})
		if err != nil {
			log.Fatalf("Error al enviar emergencia: %v", err)
		}

		<-done
	}

	cancelMonitoreo()
	time.Sleep(100 * time.Millisecond)
}
