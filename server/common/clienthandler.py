import logging

from common.protocol import MBPSocket, MBPMessage
from common.utils import Bet, store_bets
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

# Client handler
class ClientHandler:
    def __init__(self, socket: 'MBPSocket'):
        self._socket = socket

        self._wait_register()

    def _wait_register(self):
        try: 
            msg = self._socket.receive_message()

            if not RegisterMessage.is_of_type(msg):
                raise Exception(f'Register message with unexpected action received: {msg.action}')

            self.id = RegisterMessage.get_id(msg)

        except Exception as e:
            logging.error(f'action: register | result: fail | error: {e}')

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