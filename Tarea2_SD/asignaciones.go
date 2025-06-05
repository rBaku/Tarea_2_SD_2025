package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"net"
	"sync"
	"time"

	pb "Tarea2_SD/emergencia"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/streadway/amqp"
	"google.golang.org/grpc"
)

type servidorAsignador struct {
	pb.UnimplementedAsignadorServer
	dronActual int
	mu         sync.Mutex
	mongoDB    *mongo.Collection
	canal      *amqp.Channel
}

var nombresDrones = []string{"dron01", "dron02", "dron03"}

func conectarMongo() *mongo.Collection {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://10.10.28.57:27017"))
	if err != nil {
		log.Fatalf("Error conectando a MongoDB: %v", err)
	}
	return client.Database("emergencias_db").Collection("drones")
}

func conectarRabbit() *amqp.Channel {
	conn, err := amqp.Dial("amqp://guest:guest@10.10.28.57:5672/")
	if err != nil {
		log.Fatalf("Error conectando a RabbitMQ: %v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error creando canal RabbitMQ: %v", err)
	}
	ch.QueueDeclare("registro_emergencias", false, false, false, false, nil)
	ch.QueueDeclare("fin_emergencia", false, false, false, false, nil)
	return ch
}

func publicarJSON(ch *amqp.Channel, cola string, data interface{}) {
	body, _ := json.Marshal(data)
	ch.Publish("", cola, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

func (s *servidorAsignador) EnviarEmergencias(ctx context.Context, req *pb.EmergenciasRequest) (*pb.Respuesta, error) {
	for _, e := range req.Emergencias {
		s.mu.Lock()

		dron := obtenerDronMasCercano(s.mongoDB, e.Latitude, e.Longitude)

		s.mongoDB.UpdateOne(context.TODO(), bson.M{"id": dron.ID}, bson.M{"$set": bson.M{"status": "busy"}})

		doc := bson.M{
			"emergency_id": obtenerNuevoID(),
			"name":         e.Name,
			"latitude":     e.Latitude,
			"longitude":    e.Longitude,
			"magnitude":    e.Magnitude,
			"status":       "En curso",
		}
		s.mongoDB.Database().Collection("emergencias").InsertOne(context.TODO(), doc)

		publicarJSON(s.canal, "registro_emergencias", doc)

		log.Printf("Emergencia enviada a registro: %s (ID: %d)", e.Name, doc["emergency_id"])

		conn, err := grpc.Dial("10.10.28.58:50052", grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Error conectando con dron: %v", err)
		}
		defer conn.Close()
		dronClient := pb.NewDronClient(conn)

		_, err = dronClient.AtenderEmergencia(context.Background(), &pb.EmergenciaAsignada{
			EmergencyId: int32(doc["emergency_id"].(int)),
			Name:        e.Name,
			Latitude:    e.Latitude,
			Longitude:   e.Longitude,
			Magnitude:   e.Magnitude,
			DronId:      dron.ID,
		})
		if err != nil {
			log.Printf("Error enviando emergencia al dron: %v", err)
		}

		s.mu.Unlock()

		_, err = s.canal.Consume("fin_emergencia", "", true, false, false, false, nil)
		if err != nil {
			log.Fatalf("Error esperando fin_emergencia: %v", err)
		}

		time.Sleep(500 * time.Millisecond)
	}

	return &pb.Respuesta{Mensaje: "Emergencias procesadas correctamente"}, nil
}

func obtenerDronMasCercano(col *mongo.Collection, x, y float32) struct{ ID string } {
	cursor, _ := col.Find(context.TODO(), bson.M{"status": "available"})
	var drones []bson.M
	cursor.All(context.TODO(), &drones)

	var minDist float64 = math.MaxFloat64
	var elegido string
	for _, d := range drones {
		lat := d["latitude"].(float64)
		long := d["longitude"].(float64)
		dist := math.Sqrt(math.Pow(float64(x)-lat, 2) + math.Pow(float64(y)-long, 2))
		if dist < minDist {
			minDist = dist
			elegido = d["id"].(string)
		}
	}
	return struct{ ID string }{ID: elegido}
}

var idActual = 0

func obtenerNuevoID() int {
	idActual++
	return idActual
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Error escuchando: %v", err)
	}

	grpcServer := grpc.NewServer()
	s := &servidorAsignador{
		dronActual: 0,
		mongoDB:    conectarMongo(),
		canal:      conectarRabbit(),
	}
	pb.RegisterAsignadorServer(grpcServer, s)

	log.Println("Servidor de asignación escuchando en puerto 50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Error al iniciar servidor gRPC: %v", err)
	}
}
