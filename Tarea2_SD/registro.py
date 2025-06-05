import pika
import json
from pymongo import MongoClient

# Configuración de credenciales para RabbitMQ
credentials = pika.PlainCredentials('rodolfo', '123')  # Reemplaza con tus credenciales
parameters = pika.ConnectionParameters(
    host="10.10.28.57",
    credentials=credentials
)

# Establecer conexión con RabbitMQ
connection = pika.BlockingConnection(parameters)
channel = connection.channel()

channel.queue_declare(queue="registro_emergencias")
channel.queue_declare(queue="apagar_emergencias")

# Conexión a MongoDB (aquí también podrías necesitar credenciales)
client = MongoClient("10.10.28.57", 27017)
db = client.emergencias_db
col = db.emergencias

def registrar_emergencia(ch, method, properties, body):
    data = json.loads(body)
    existing = col.find_one({"emergency_id": data["emergency_id"]})
    if not existing:
        col.insert_one(data)

def actualizar_estado(ch, method, properties, body):
    data = json.loads(body)
    emergency_id = data["emergency_id"]
    col.update_one({"emergency_id": emergency_id}, {"$set": {"status": "Extinguido"}})

channel.basic_consume(queue="registro_emergencias", on_message_callback=registrar_emergencia, auto_ack=True)
channel.basic_consume(queue="apagar_emergencias", on_message_callback=actualizar_estado, auto_ack=True)

print("Servicio de registro escuchando...")
channel.start_consuming()