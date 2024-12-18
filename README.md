# Sistemas Distribuidos

Este repositorio contiene implementaciones de algoritmos populares de sistemas distribuidos. Incluye el **Ricart-Agrawala** en Go para exclusión mutua distribuida y el **algoritmo de consenso Raft** junto a una **versión con integración de Kubernetes**. Las explicaciones de despliegue de cada algoritmo se encuentran en sus respectivas carpetas.

## Contenido

### **Ricart-Agrawala**:

- Permite la exclusión mutua en sistemas distribuidos sin un coordinador central.
- Utiliza mensajes de solicitud, permiso y liberación para controlar el acceso a recursos compartidos.
- Reduce la cantidad de mensajes necesarios en comparación con otros algoritmos de exclusión mutua distribuidos.
- Garantiza la equidad en el acceso a los recursos entre los nodos participantes.

### **Raft**:

- Proporciona un protocolo de consenso robusto para garantizar la consistencia de datos en sistemas distribuidos.
- Permite la recuperación automática ante fallos mediante la elección de líderes y la replicación de registros.

### **Kubernetes**:

- Simplifica el despliegue y la gestión de clústeres con Kubernetes.
- Ofrece una solución escalable para la gestión y orquestación de contenedores.

## Teclonogías empleadas:

- **Golang**: Lenguaje de programación utilizado para implementar los algoritmos de sistemas distribuidos.
- **Docker**: Herramienta de contenedores utilizada para empaquetar y distribuir aplicaciones.
- **Kubernetes**: Plataforma de orquestación de contenedores utilizada para gestionar clústeres de contenedores.
