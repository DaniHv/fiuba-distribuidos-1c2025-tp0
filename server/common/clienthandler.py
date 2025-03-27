import logging
import json
import threading
import queue

from common.protocol import MBPSocket, MBPMessage
from common.utils import Bet, store_bets
from common.betsstore import BetsSharedStore

# Client->Server messages
class PlaceBetMessage:
    def is_of_type(message: 'MBPMessage') -> 'bool':
        return message.action == 'PLACE_BET'

    def get_bet(agency_id: 'str', message: 'MBPMessage') -> 'Bet':
        data = json.loads(message.data)

        if not all(key in data for key in ['FirstName', 'LastName', 'Document', 'BirthDate', 'Number']):
            raise ValueError(f'Invalid JSON Bet format ({message.data})')

        return Bet(agency_id, data['FirstName'], data['LastName'], data['Document'], data['BirthDate'], data['Number'])

class ProcessBetsMessage:
    def is_of_type(message: 'MBPMessage') -> 'bool':
        return message.action == 'PROCESS_BETS'

class EndBetsMessage:
    def is_of_type(message: 'MBPMessage') -> 'bool':
        return message.action == 'END'

class RegisterMessage:
    def is_of_type(message: 'MBPMessage') -> 'bool':
        return message.action == 'REGISTER'
    
    def get_id(message: 'MBPMessage') -> int:
        data = json.loads(message.data)

        if not 'ID' in data:
            raise ValueError(f'Invalid JSON Register format ({message.data})')

        return data['ID']

# Server->Client messages
class BetsProcessedMessage:
    def create_message(success: 'bool') -> 'MBPMessage':
        return MBPMessage('BETS_PROCESSED', bytes(json.dumps({'result': 'success' if success else 'fail'}), 'utf-8'))

class BetsResultsMessage:
    def create_message(total: int) -> 'MBPMessage':
        return MBPMessage('WINNERS', bytes(json.dumps({ "Winners": total }), 'utf-8'))

# Client handler
class ClientHandler(threading.Thread):
    def __init__(self, socket: 'MBPSocket', bets_store: 'BetsSharedStore', draw_barrier: 'threading.Barrier'):
        super().__init__()
        self._socket = socket
        self._bets_store = bets_store
        self._draw_barrier = draw_barrier
        self._results_queue = queue.Queue()

    def run(self):
        self._wait_register()
        self._request_bets()
        self._draw_barrier.wait()
        self._send_results_to_client()

    def _wait_register(self):
        try: 
            msg = self._socket.receive_message()

            if not RegisterMessage.is_of_type(msg):
                raise Exception(f'Register message with unexpected action received: {msg.action}')

            self.id = RegisterMessage.get_id(msg)

        except Exception as e:
            logging.error(f'action: register | result: fail | error: {e}')

    def _request_bets(self):
        while True:
            if not self._request_bets_batch():
                break

    def _send_results_to_client(self):
        total = self._results_queue.get()

        try:
            self._socket.send_message(BetsResultsMessage.create_message(total))

            logging.info(f'action: send_results | result: success | cantidad: {total}')
        except Exception as e:
            logging.error(f'action: send_results | result: fail | error: {e}')

    def send_results(self, total: int):
        self._results_queue.put(total)

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

                logging.debug(f'action: single_apuesta_recibida | result: in_progress | data: {message.data}')
                bets.append(PlaceBetMessage.get_bet(self.id, message))

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

    def close(self):
        self._socket.close()