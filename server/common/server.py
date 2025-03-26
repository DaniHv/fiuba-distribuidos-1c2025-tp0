import logging
import signal
import select
import os
import json
from common.protocol import MBPSocket
from common.utils import Bet
from common.clienthandler import ClientHandler

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = MBPSocket()
        self._server_socket.listen(port, listen_backlog)

        # Initialize graceful shutdown.
        self.shutdown_r, self.shutdown_w = os.pipe()
        signal.signal(signal.SIGTERM, self.__shutdown_signal)
        signal.signal(signal.SIGINT, self.__shutdown_signal)

    def __shutdown_signal(self, signum, frame):
        logging.info("action: graceful_shutdown | result: in_progress | signal: %s", signum)

        self.shutdown = True
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

                    handler = ClientHandler(client_sock)
                    handler.handle_connection()

                elif r == self.shutdown_r:
                    self._server_socket.close()
                    logging.info("action: graceful_shutdown | result: success")

                    return
