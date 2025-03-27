import logging
import signal
import select
import os
import json
from common.protocol import MBPSocket
from common.utils import Bet
from common.clienthandler import ClientHandler
from common.lottery import Lottery


class Server:
    def __init__(self, port, listen_backlog, clients_qty):
        # Initialize server socket
        self._server_socket = MBPSocket()
        self._server_socket.listen(port, listen_backlog)

        # Initialize graceful shutdown.
        self.shutdown_r, self.shutdown_w = os.pipe()
        signal.signal(signal.SIGTERM, self.__shutdown_signal)
        signal.signal(signal.SIGINT, self.__shutdown_signal)

        # Clients waiting for lottery results
        self.clients_qty = clients_qty
        self.clients = []

    def __shutdown_signal(self, signum, frame):
        logging.info("action: graceful_shutdown | result: in_progress | signal: %s", signum)

        self.shutdown = True
        os.write(self.shutdown_w, b'1')

    def __graceful_shutdown(self):
        self._server_socket.close()

        for client in self.clients:
            client.close()

        logging.info("action: graceful_shutdown | result: success")

    def run(self):
        """
        Start listening for connections until a shutdown is requested.
        """

        # Accept connections from agencies until all of them finish sending their bets
        while len(self.clients) is not self.clients_qty:
            logging.info('action: accept_connections | result: in_progress')

            # Wait for new connections or shutdown signal
            readables, _, _ = select.select([self._server_socket, self.shutdown_r], [], [])

            for r in readables:                
                if r == self._server_socket:
                    client_sock, addr = self._server_socket.accept()
                    logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')

                    handler = ClientHandler(client_sock)
                    handler.request_bets()

                    self.clients.append(handler)

                elif r == self.shutdown_r:
                    self.__graceful_shutdown()
                    return

        # Get winners and notify clients
        lottery = Lottery()
        lottery.notify_winners(self.clients)
