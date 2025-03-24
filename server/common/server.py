import logging
import signal
import select
import os
from common.protocol import MBPSocket, MBPMessage

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = MBPSocket()
        self._server_socket.listen(port, listen_backlog)

        # Initialize graceful shutdown
        self.shutdown_r, self.shutdown_w = os.pipe()
        signal.signal(signal.SIGTERM, self.__shutdown_signal)
        signal.signal(signal.SIGINT, self.__shutdown_signal)

    def __shutdown_signal(self, signum, frame):
        logging.info("action: graceful_shutdown | result: in_progress | signal: %s", signum)

        os.write(self.shutdown_w, b'1')

    def run(self):
        """
        Start listening for connections until a shutdown is requested.
        """

        while True:
            logging.info('action: accept_connections | result: in_progress')

            # Wait for new connections or shutdown signal
            readables, _, _ = select.select([self._server_socket, self.shutdown_r], [], [])
            
            for r in readables:                
                if r == self._server_socket:
                    client_sock, addr = self._server_socket.accept()
                    logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')

                    self.__handle_client_connection(client_sock)

                elif r == self.shutdown_r:
                    self._server_socket.close()
                    logging.info("action: graceful_shutdown | result: success")

                    return

    def __handle_client_connection(self, client_sock: 'MBPSocket'):
        addr = client_sock.getpeername()
        message = None

        try:
            message = client_sock.receive_message()

            logging.info(f'action: receive_message | result: success | ip: {addr[0]} | msg: {message.data.decode("utf-8")}')
        except Exception as e:
            logging.error(f'action: receive_message | result: fail | error: {e}')

        if message is not None:
            try:
                client_sock.send_message(message)
    
                logging.info(f'action: send_message | result: success | ip: {addr[0]} | msg: {message.data.decode("utf-8")}')
            except Exception as e:
                logging.error(f'action: send_message | result: fail | error: {e}')
        
        client_sock.close()
