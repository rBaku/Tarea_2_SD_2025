package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"time"

	pb "Tarea2_SD/emergencia"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

type servidorDron struct {
	pb.UnimplementedDronServer
	canal   *amqp.Channel
	mongoDB *mongo.Collection
}

// insertarDrones inicializa la base de datos con drones disponibles si no existen
//
// Parámetros:
//
//	col *mongo.Collection: Colección MongoDB donde insertar los drones
func insertarDrones(col *mongo.Collection) {
	drones := []interface{}{
		bson.M{"id": "dron01", "latitude": 0.0, "longitude": 0.0, "status": "available"},
		bson.M{"id": "dron02", "latitude": 0.0, "longitude": 0.0, "status": "available"},
		bson.M{"id": "dron03", "latitude": 0.0, "longitude": 0.0, "status": "available"},
	}
	for _, d := range drones {
		id := d.(bson.M)["id"]
		count, _ := col.CountDocuments(context.TODO(), bson.M{"id": id})
		if count == 0 {
			col.InsertOne(context.TODO(), d)
		}
	}
}

// conectarMongo establece conexión con MongoDB y asegura que existan drones iniciales
//
// Retorna:
//
//	*mongo.Collection: Referencia a la colección de drones
func conectarMongo() *mongo.Collection {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://10.10.28.57:27017"))
	if err != nil {
		log.Fatal("Mongo error:", err)
	}
	col := client.Database("emergencias_db").Collection("drones")
	insertarDrones(col)
	return col
}

// conectarRabbit establece conexión con RabbitMQ y declara las colas necesarias
//
// Retorna:
//
//	*amqp.Channel: Canal de comunicación RabbitMQ
func conectarRabbit() *amqp.Channel {
	conn, _ := amqp.Dial("amqp://rodolfo:123@10.10.28.57:5672/")
	ch, _ := conn.Channel()
	ch.QueueDeclare("acciones_dron", false, false, false, false, nil)
	ch.QueueDeclare("apagar_emergencias", false, false, false, false, nil)
	ch.QueueDeclare("fin_emergencia", false, false, false, false, nil)
	return ch
}

// publicarTexto envía un mensaje de texto a una cola RabbitMQ específica
//
// Parámetros:
//
//	ch *amqp.Channel: Canal RabbitMQ
//	cola string: Nombre de la cola destino
//	msg string: Mensaje a enviar
func publicarTexto(ch *amqp.Channel, cola, msg string) {
	ch.Publish("", cola, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(msg),
	})
}

// publicarJSON serializa datos a JSON y los publica en una cola RabbitMQ
//
// Parámetros:
//
//	ch *amqp.Channel: Canal RabbitMQ
//	cola string: Nombre de la cola destino
//	data interface{}: Datos a serializar como JSON
func publicarJSON(ch *amqp.Channel, cola string, data interface{}) {
	body, _ := json.Marshal(data)
	ch.Publish("", cola, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

// publicarCada5Segundos envía mensajes periódicos durante un tiempo determinado
//
// Parámetros:
//
//	duracion time.Duration: Tiempo total de envío
//	mensaje string: Contenido a enviar
//	canal *amqp.Channel: Canal RabbitMQ a usar
func publicarCada5Segundos(duracion time.Duration, mensaje string, canal *amqp.Channel) {
	intervalo := 5 * time.Second
	total := int(duracion / intervalo)
	resto := duracion % intervalo

	if total == 0 {
		publicarTexto(canal, "acciones_dron", mensaje)
		time.Sleep(duracion)
		return
	}

	for i := 0; i < total; i++ {
		publicarTexto(canal, "acciones_dron", mensaje)
		time.Sleep(intervalo)
	}

	if resto > 0 {
		publicarTexto(canal, "acciones_dron", mensaje)
		time.Sleep(resto)
	}
}

// AtenderEmergencia implementa el servicio gRPC para manejo de emergencias por drones
//
// Flujo de operaciones:
// 1. Actualiza estado del dron a "unavailable"
// 2. Calcula tiempo de desplazamiento según distancia
// 3. Publica actualizaciones periódicas del estado
// 4. Al finalizar, actualiza posición y estado del dron
// 5. Notifica finalización de emergencia
//
// Parámetros:
//
//	ctx context.Context: Contexto de ejecución
//	e *pb.EmergenciaAsignada: Datos de la emergencia asignada
//
// Retorna:
//
//	*pb.Respuesta: Confirmación de operación
//	error: Posible error durante el proceso
func (s *servidorDron) AtenderEmergencia(ctx context.Context, e *pb.EmergenciaAsignada) (*pb.Respuesta, error) {
	dronID := e.DronId
	fmt.Printf("%s atendiendo emergencia: %s\n", dronID, e.Name)

	var dron struct {
		ID        string  `bson:"id"`
		Latitude  float64 `bson:"latitude"`
		Longitude float64 `bson:"longitude"`
	}
	err := s.mongoDB.FindOne(context.TODO(), bson.M{"id": dronID}).Decode(&dron)
	if err != nil {
		dron.Latitude, dron.Longitude = 0, 0
	}

	s.mongoDB.UpdateOne(context.TODO(), bson.M{"id": dronID}, bson.M{"$set": bson.M{"status": "unavailable"}})

	dx := math.Abs(float64(e.Latitude) - dron.Latitude)
	dy := math.Abs(float64(e.Longitude) - dron.Longitude)
	distancia := dx + dy
	duracionDesplazamiento := time.Duration(distancia * 0.5 * float64(time.Second))
	duracionApagado := time.Duration(e.Magnitude) * 2 * time.Second

	publicarTexto(s.canal, "acciones_dron", fmt.Sprintf("Se ha asignado %s a la emergencia", dronID))
	publicarCada5Segundos(duracionDesplazamiento, "Dron en camino a emergencia...", s.canal)
	publicarCada5Segundos(duracionApagado, "Dron apagando emergencia...", s.canal)
	publicarTexto(s.canal, "acciones_dron", fmt.Sprintf("%s ha sido extinguido por %s", e.Name, dronID))

	s.mongoDB.UpdateOne(context.TODO(), bson.M{"id": dronID}, bson.M{"$set": bson.M{
		"latitude":  e.Latitude,
		"longitude": e.Longitude,
		"status":    "available",
	}})

	publicarJSON(s.canal, "apagar_emergencias", bson.M{"emergency_id": e.EmergencyId})
	publicarJSON(s.canal, "fin_emergencia", bson.M{"emergency_id": e.EmergencyId})

	return &pb.Respuesta{Mensaje: "Emergencia atendida correctamente"}, nil
}

// main inicia el servidor gRPC del servicio de drones
//
// Configura:
// 1. Conexión a MongoDB (colección drones)
// 2. Conexión a RabbitMQ (canal de mensajería)
// 3. Servidor gRPC escuchando en puerto 50052
func main() {
	lis, _ := net.Listen("tcp", ":50052")
	grpcServer := grpc.NewServer()
	canal := conectarRabbit()
	mongo := conectarMongo()

	pb.RegisterDronServer(grpcServer, &servidorDron{canal: canal, mongoDB: mongo})
	fmt.Println("Servicio de drones escuchando en puerto 50052...")
	grpcServer.Serve(lis)
}
