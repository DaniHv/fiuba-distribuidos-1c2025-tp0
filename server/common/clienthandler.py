import logging
import json
import select

from common.protocol import MBPSocket, MBPMessage
from common.utils import Bet, store_bets

# Client->Server messages
class PlaceBetMessage:
    def is_of_type(message: 'MBPMessage') -> 'bool':
        return message.action == 'PLACE_BET'
    
    def get_bet(message: 'MBPMessage') -> 'Bet':
        data = json.loads(message.data)

        if not all(key in data for key in ['Agency', 'FirstName', 'LastName', 'Document', 'BirthDate', 'Number']):
            raise ValueError(f'Invalid JSON Bet format ({message.data})')

        return Bet(data['Agency'], data['FirstName'], data['LastName'], data['Document'], data['BirthDate'], data['Number'])

class ProcessBetsMessage:
    def is_of_type(message: 'MBPMessage') -> 'bool':
        return message.action == 'PROCESS_BETS'

class EndBetsMessage:
    def is_of_type(message: 'MBPMessage') -> 'bool':
        return message.action == 'END'

# Server->Client messages
class BetsProcessedMessage:
    def create_message(success: 'bool') -> 'MBPMessage':
        return MBPMessage('BETS_PROCESSED', bytes(json.dumps({'result': 'success' if success else 'fail'}), 'utf-8'))

# Client handler
class ClientHandler:
    def __init__(self, socket: 'MBPSocket'):
        self._socket = socket

    def handle_connection(self):
        self._request_bets()

        self._socket.close()

    def _request_bets(self):
        while True:
            if not self._request_bets_batch():
                break

    # Request and process a batch of bets from the client, returning True if more
    # batches are expected or False if the client has finished.
    def _request_bets_batch(self) -> 'bool':
        bets = []

        try:
            # Obtain batch bets from the client
            while True:
                message = self._socket.receive_message()

                # EndBetsMessage is expected to be preceded by a ProcessBetsMessage,
                # for that reason the loop is stopped as no bets are expected to be
                # pending for processing.
                if EndBetsMessage.is_of_type(message):
                    logging.debug(f'action: end_of_bets | result: in_progress')
                    return False

                if ProcessBetsMessage.is_of_type(message):
                    logging.debug(f'action: process_bets | result: in_progress')
                    break

                if not PlaceBetMessage.is_of_type(message):
                    raise Exception(f'Unexpected action received: {message.action}')

                logging.debug(f'action: apuesta_recibida | result: in_progress | data: {message.data}')
                bets.append(PlaceBetMessage.get_bet(message))

            # Process the batch of bets (store)
            store_bets(bets)
            logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')

            # Notify the client that the bets have been processed
            self._socket.send_message(BetsProcessedMessage.create_message(True))

            return True

        except Exception as e:
            logging.info(f'action: apuesta_recibida | result: fail | cantidad (hasta error): {len(bets)} | error: {e}')
            self._socket.send_message(BetsProcessedMessage.create_message(False))
            return False