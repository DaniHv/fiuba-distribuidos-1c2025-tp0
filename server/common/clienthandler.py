import logging
import threading
import queue

from common.protocol import MBPSocket, MBPMessage
from common.utils import Bet, store_bets
from common.betsstore import BetsSharedStore
from common.serialization import SBDSerialization

# Client->Server messages
class PlaceBetsMessage:
    def is_of_type(message: 'MBPMessage') -> 'bool':
        return message.action == 'PLACE_BETS'
    
    def get_bets(agency_id: 'str', message: 'MBPMessage') -> 'Bet':
        bet_parts = SBDSerialization.deserialize_array(message.data, 5)

        return [Bet(agency_id, bet[0], bet[1], bet[2], bet[3], bet[4]) for bet in bet_parts]

class EndBetsMessage:
    def is_of_type(message: 'MBPMessage') -> 'bool':
        return message.action == 'END'

class RegisterMessage:
    def is_of_type(message: 'MBPMessage') -> 'bool':
        return message.action == 'REGISTER'
    
    def get_id(message: 'MBPMessage') -> int:
        parts = SBDSerialization.deserialize(message.data, 1)

        return parts[0]

# Server->Client messages
class BetsProcessedMessage:
    def create_message(success: 'bool') -> 'MBPMessage':
        return MBPMessage('BETS_PROCESSED', SBDSerialization.serialize(['success' if success else 'fail']))

class BetsResultsMessage:
    def create_message(winners: 'dict') -> 'MBPMessage':
        return MBPMessage('WINNERS', SBDSerialization.serialize_array([[k, str(v)] for k, v in winners.items()]))

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
        winners = self._results_queue.get()

        try:
            self._socket.send_message(BetsResultsMessage.create_message(winners))

            logging.info(f'action: send_results | result: success | client: {self.id} | cantidad: {sum(winners.values())}')
        except Exception as e:
            logging.error(f'action: send_results | result: fail | client: {self.id} | error: {e}')

    def send_results(self, winners: 'dict'):
        self._results_queue.put(winners)

    # Request and process a batch of bets from the client, returning True if more
    # batches are expected or False if the client has finished.
    def _request_bets_batch(self) -> 'bool':
        try:
            # Obtain batch bets from the client
            message = self._socket.receive_message()

            # EndBetsMessage is expected to be preceded by a ProcessBetsMessage,
            # for that reason the loop is stopped as no bets are expected to be
            # pending for processing.
            if EndBetsMessage.is_of_type(message):
                logging.debug(f'action: end_of_bets | result: in_progress')
                return False

            if not PlaceBetsMessage.is_of_type(message):
                raise Exception(f'Unexpected action received: {message.action}')

            # Process the batch of bets (store)
            bets = PlaceBetsMessage.get_bets(self.id, message)
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