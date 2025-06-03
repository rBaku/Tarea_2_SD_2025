# Tarea_2_SD_2025

## Integrantes

| Nombre                 | Rol                      |
|------------------------|--------------------------|
| Luis López Aguilera    | 202173049-4              |
| Sofia Bastias Valdés   | 201973041-K              |
| Rodolfo Osorio Verdejo | 201973014-2|             |

## Instrucciones de uso
El ideal por ahora (favor de cambiar cuando esté listo) de qué máquina virtual será qué tenemos:

| Máquina                 | Componente              |
|------------------------|--------------------------|
| Terminada en 56        | Cliente y Servicio de monitoreo   |
| Terminada en 57        | Servicio de asignación, Servicio de registro y MongoDB  |
| Terminada en 58        | Sistema de drones|             |

### Comandos a ejecutar
(Completar)
En este orden
1. **En 57:**
   ```bash
   python3 path_3/registro.py
   go run path_2/asignacion.go
2. **En 56:**
   ```bash
   go run path_4/monitoreo.go
3. **En 58**
   ```bash
   go run path_5/drones.go
4.**En 56 nuevamente**
  ```bash
  ./cliente emergencia.json
