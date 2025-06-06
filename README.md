# Tarea_2_SD_2025

## Integrantes

| Nombre                 | Rol                      |
|------------------------|--------------------------|
| Luis López Aguilera    | 202173049-4              |
| Sofia Bastias Valdés   | 201973041-K              |
| Rodolfo Osorio Verdejo | 201973014-2|             |

## Instrucciones de uso
Con respecto a las máquinas y sus componentes tenemos:

| Máquina                 | Componente              |
|------------------------|--------------------------|
| (MV1)Terminada en 56        | Cliente y Servicio de monitoreo   |
| (MV2)Terminada en 57        | Servicio de asignación, Servicio de registro y MongoDB  |
| (MV3)Terminada en 58        | Sistema de drones|             |

## Consideracion
El archivo database.mongo contiene los 3 drones iniciales que se insertan en la colección `drones` de la base de datos `emergencias_db`.  
Puede ser utilizado con `mongoimport` para poblar manualmente la base si se requiere, aunque NO es necesario ya que drones.go inserta los drones en
la coleccion 'drones' de la base de datos 'emergencias_db', si se prefiere poblarla con este metodo usar el comando 'mongoimport --db emergencias_db --collection drones --file database.mongo' y asegurese de estar en el path Tarea2_SD.

### Comandos a ejecutar
(Completar)
En este orden
1. **En 58 (MV3)**
   ```bash
   go run Tarea_2_SD_2025/Tarea2_SD/drones.go

2. **En 57 (MV2):**
   ```bash
   go run Tarea_2_SD_2025/Tarea2_SD/asignacion.go
   python3 Tarea_2_SD_2025/Tarea2_SD/registro.py
   
3. **En 56 (MV1):**
   ```bash
   go run Tarea_2_SD_2025/Tarea2_SD/monitoreo.go
   
4. **En 56 nuevamente**
  ```bash
  Tarea_2_SD_2025/Tarea2_SD/ go run cliente.go emergencia.json

