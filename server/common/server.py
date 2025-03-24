import logging
import signal
import select
import os
import json
from common.protocol import MBPSocket, MBPMessage
from common.utils import Bet, store_bets

def deserialize_bet(json_str):
    data = json.loads(json_str)

    if not all(key in data for key in ['Agency', 'FirstName', 'LastName', 'Document', 'BirthDate', 'Number']):
        raise ValueError(f'Invalid JSON Bet format ({json_str})')

    return Bet(data['Agency'], data['FirstName'], data['LastName'], data['Document'], data['BirthDate'], data['Number'])

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
        try:
            message = client_sock.receive_message()

            if message.action != 'PLACE_BET':
                raise Exception('Unexpected action received')

            bet = deserialize_bet(message.data)
            store_bets([bet])

            confirmation = MBPMessage('STORED_BET')
            client_sock.send_message(confirmation)

            logging.info(f'action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}')

        except Exception as e:
            logging.info(f'action: apuesta_almacenada | result: fail | error: {e}')

        client_sock.close()
