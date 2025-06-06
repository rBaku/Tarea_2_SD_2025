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

### Comandos a ejecutar
(Completar)
En este orden
1. **En 57 (MV2):**
   ```bash
   python3 Tarea_2_SD_2025/Tarea2_SD/registro.py
   go run Tarea_2_SD_2025/Tarea2_SD/asignacion.go
2. **En 56 (MV1):**
   ```bash
   go run Tarea_2_SD_2025/Tarea2_SD/monitoreo.go
3. **En 58 (MV3)**
   ```bash
   go run Tarea_2_SD_2025/Tarea2_SD/drones.go
4. **En 56 nuevamente**
  ```bash
  Tarea_2_SD_2025/Tarea2_SD/ go run cliente.go emergencia.json
