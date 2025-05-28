import pika
import json
from pymongo import MongoClient

connection = pika.BlockingConnection(pika.ConnectionParameters("localhost"))
channel = connection.channel()

channel.queue_declare(queue="registro_emergencias")
channel.queue_declare(queue="apagar_emergencias")

client = MongoClient("localhost", 27017)
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