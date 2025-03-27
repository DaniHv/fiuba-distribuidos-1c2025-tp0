# TP0: Docker + Comunicaciones + Concurrencia

## EJ1

Se implementa el script `generar-compose.sh` para generar de forma automática una definición de docker-compose del sistema cliente-servidor con una cantidad variable de clientes (pueden ser cero clientes).
```bash
./generar-compose.sh {{file}} {{clients}}
```

Por ejemplo, para crear un docker-compose para un único cliente, se ejecuta:
```bash
./generar-compose.sh docker-compose-dev.yaml 1
```

Importante: este script es modificar en ejercicios siguientes para satisfacer las necesidades de cada uno de ellos en base a las modificaciones y agregados.

## EJ2:

Para evitar la necesidad de builds ante el cambio de configuraciones (config.ini en server, y config.yaml en clientes), estos archivos de configuración se excluyen de las imágenes de docker haciendo uso del `.dockerignore` de la raíz para evitar modificar los Dockerfiles incluyendo manualmente todo excepto ese archivo (el flag --exclude del comando COPY de docker podría solucionarlo sin necesidad de un .dockerignore, sin embargo este aún no se encuentra en versiones productivas de Docker).

Puesto a que estos archivos de configuración siguen siendo necesarios, los mismos son inyectados al levantar las instancias mediante volumenes definidos en el docker-compose-dev.yaml, los cuales copian el contenido en los contenedores al ser creados.

Adicionalmente, el Dockerfile provisto para el servidor tenía un problema: al hacer COPY de toda la carpeta raíz en lugar de únicamente server/, cambios no relacionados al servidor como modificaciones en los scripts bash, archivos de documentación e incluso cambios en el cliente provocaban también la invalidación de la caché de la imagen del servidor. Por este motivo, se reorganizó la estructura de dependencias de go para llevarlas en su totalidad a /server y aislar el build de los archivos comunes del root.

## EJ3:

Con el objetivo de confirmar el funcionamiento del echo server dentro de la misma red de docker creada mediante docker-compose y de forma aislada al host, se crea el script `validar-echo-server.sh` el cual crea un contenedor de docker conectado a la red `tp0_testing_net`, dentro del cual ejecuta una request TCP al echo server mediante netcat con el que se busca validar si la respuesta recibida corresponde al valor enviado.

Importante: debido a que este script prueba el servidor echo, su uso solo es válido durante este ejercicio 3 y ejercicio 4.

## EJ4:

Para realizar un graceful shutdown en cliente y servidor se agregan handlers para las señales SIGTERM y SIGINT.

En el servidor, se utiliza un `accept` no bloqueante en conjunto a un `select` que permite de forma óptima esperar por una conexión de un cliente y procesar su consulta, o esperar la señal de shutdown y abortar de forma controlada, y ejecutar lo primero que ocurra.

En el cliente, se emplea la misma técnica (un `select`) al finalizar cada una de las operaciones, esperando por la finalización del sleep para continuar con un mensaje siguiente, o abortar de forma controlada.

## EJ5:

### Protocolo de transmisión
Se implementa un protocolo de transmisión denominado "MBP" (Message-based Protocol), el cual brinda comunicación full duplex basada en mensajes individuales sobre una conexión TCP (Similar a WebSockets). Al ser un protocolo full duplex, una misma conexión puede ser usada para el envío y recepción de múltiples mensajes en ambas direcciones, lo que resulta ideal para los siguientes ejercicios.

Este protocolo envía y recibe mensajes conformados por una acción (string, que permite identificar que hacer con el contenido del mensaje, por ejemplo: "PLACE_BET") y una información asociada (data) cuyo contenido es transmitido en binarios sin ninguna serialización/deserialización en esta capa de comunicación. La responsabilidad de definir un formato de datos a enviar y recibir se delega a la implementación de la capa de aplicación (Servidor "Loteria", Cliente "Agencia").

La implementación del protocolo se ha realizado de forma independiente, genérica y no acoplada a la lógica de negocio (agencias / lotería). Tanto en Python como en Go, se han implementado clases que proveen métodos para una utilización de forma -casi- equivalente a implementaciones estándares TCP/UDP en la mayoría de lenguajes. Estos son:

- Connect: Conecta el socket con otro dada su dirección y puerto (Solo implementado en el cliente)
- Listen: Abre un socket en modo listen, permitiendo que otros sockets se conecten a él (Solo implementado en el servidor)
- Accept: Obtiene un nuevo `MBPSocket` para una conexión entrante (Solo implementado en el servidor)
- SendMessage: Envía un `MBPMessage` al socket al conectado, equivalente a un "send" TCP/UDP.
- ReceiveMessage: Recibe un `MBPMessage` desde el socket conectado, equivalente a un "recv" TCP/UDP.
- Close: Cierra el socket.

Importante: Debido a que TCP es un protocolo que garantiza el orden y la entrega de paquetes, se han obviado algunas indicaciones en mensajes de ejercicios futuros tales como indicar la cantidad de apuestas recibidas por parte del servidor al cliente.

### Cliente/Servidor
Tanto el cliente como el servidor son modificados para implementar la lógica de negocio necesaria. La comunicación entre ellos se lleva mediante `MBPMessages` a través de un `MBPSocket`, los cuales hacen uso de información transmitida en formato json (debido a su soporte integrado en go y python en sus respectivas librerías estándar, y su facilidad de uso).

Se definen los mensajes:

- Cliente > Servidor
- `PLACE_BET`: enviado por el cliente al servidor para registrar una apuesta, con información en json `{ "Agency": {{id cliente}}, "FirstName": {{nombre}}, "LastName": {{apellido}}, "Document": {{document}}, "BirthDate": {{nacimiento}}, "Number": {{numero}} }`.

- Servidor > Cliente
- `BETS_PROCESSED`: enviado por el cliente al servidor para registrar una apuesta, con data `{}`.
- `BET` enviado por el servidor al cliente al registrar correctamente una apuesta recibida, sin data adicional.

# EJ6:

Se modifica el servidor para recibir en loop múltiples apuestas (en EJ5 el servido recibe una apuesta y termina la conexión con el cliente) y procesarlas en batch. En el cliente, se leen desde el archivo `bets.csv` en su directorio (relativo al `main.go`) las apuestas necesarias para conformar un batch, evitando almacenar en memoria grandes cantidades de objetos (bets) previo al envio al servidor.

Para este objetivo se implementan mensajes adicionales:

- Cliente -> Servidor:
- `PROCESS_BETS`: Le indica al servidor que ya se enviaron todas las apuestas del batch. Este mensaje no contiene información adicional.
- `END`: Le indica al servidor que no hay más apuestas por ser enviadas. Siempre será precedido por un `PROCESS_BETS`. Este mensaje no contiene información adicional.

- Servidor -> Cliente:
- `BETS_PROCESSED`: Le indica al cliente que su solicitud de procesamiento `PROCESS_BETS` fue exitosa. Este mensaje no contiene información adicional.

Una vez finalizado el envío de todas las apuestas, el cliente se desconecta y el servidor vuelve a estar disponible para recibir nuevas conexiones.

# EJ7:

Para llegar a cabo el procesamiento de los ganadores (sorteo) y notificación a los clientes, se optó por una estrategia de `long polling`. El servidor aceptará conexiones y recibirá las apuestas de cada uno de los clientes conectados de forma serial hasta que todos terminen. Al finalizar la recepción, se libera el thread principal del loop de "accepts" para proceder a realizar el sorteo, manteniendo las referencias de los handlers de cada uno de los clientes (y sus respectivos sockets aún abiertos) para que estos puedan ser notificados al finalizar.

El procesamiento de los ganadores se realiza mediante las funciónes provistas `load_bets`, la cual se itera para obtener cada una de las apuestas realizadas y ejecutar con ellas `has_won` dentro de la misma iteración, evitando almacenar en memoria grandes cantidades de objetos (bets) del archivo bets.csv en su totalidad.

Para este objetivo se implementan mensajes adicionales:

Cliente->Servidor:
- `REGISTER`: Le indica al servidor cual agencia se está registrado, con información en json `{ "ID": {{id cliente}} }`. Anteriormente no resultaba relevante conocer el id del cliente ya que las apuestas enviadas contenian el id, pero al realizar el sorteo resulta necesario conocerlo, planteando un register inicial y evitando la redudancia de los `PLACE_BET`.
- `PLACE_BET`: Se modifica el mensaje del EJ5 para eliminar el dato `Agency` en el json enviado.

Servidor->Cliente:
- `WINNERS`: Le indica al cliente la cantidad de ganadores de apuestas enviadas por él, con información en json `{ "Winners": {{cantidad}} }`

# EJ8:

Se modifica el servidor para establecer conexiones con los clientes, recibir las apuestas y guardarlas de forma paralela. Para ello, se hace uso de multithreading, creando un thread por cada conexión con el cliente.

El thread principal permanece siendo el responsable de aceptar las conexiones entrantes (main socket listen / accept) y crear los threads necesarios para el manejo de cada una de ellas. El sorteo por otra parte, es realizado en el thread principal, debido a que éste no tiene ninguna otra ocupación al no mantenerse aceptando nuevas conexiones una vez que ya se conectaron todos los clientes esperados (debe ser conocido de antemano por la medida de sincronización para el sorteo del EJ7).

Para garantizar un funcionamiento óptimo y esperado y satisfacer los requerimientos solicitados, fueron implementados los siguientes mecanismos de sincronización:

- WriteLock: Puesto que todos los clientes (agencias) requieren almacenar las apuestas en un mismo lugar (bets.csv), se hace uso de `threading.Lock` para garantizar que solo un chunk de apuestas puede ser escrito a la vez. Aclaración: no se modificó la implementación brindada de bets_store, sin embargo resultaría una mejor alternativa mantener el mismo archivo abierto en lugar de abrir y cerrarlo por cada chunk a escribir (que pueden ser muchos) como se hace actualmente.

- Barrier: Puesto que el sorteo de los ganadores debe hacerse recién todos los clientes (agencias) terminaron de enviar sus apuestas, se utiliza una barrera para sincronizar todos los threads de clientes, y del thread principal para proceder con el sorteo en el momento necesario.

Importante: Si bien el uso de threads en python presentan un problema de performance debido al GIL (Global Interpreter Lock) que no permite un cómputo realmente paralelo, las necesidad de esta solución cliente-servidor para el problema planteado involucra casi en su totalidad operaciones de IO, por lo que el uso de threads permite la ejecución de forma concurrente de la recepción y procesamiento de mensajes sacando provecho de los tiempos de espera en lecturas y escrituras.