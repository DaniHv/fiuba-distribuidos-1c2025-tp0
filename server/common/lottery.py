import logging

from common.clienthandler import ClientHandler
from common.utils import load_bets, has_won

class Lottery:
  clients_done = 0

  def notify_winners(self, clients: 'list[ClientHandler]'):
    """
    Get the winners of the lottery and notify the clients.
    """
    winners_by_client = {}

    logging.debug(f'action: processing_winners | result: pending')

    bets = 0

    for bet in load_bets():
      bets += 1
      if has_won(bet):
        winners_by_client[bet.agency] = winners_by_client.get(bet.agency, 0) + 1

    for client in clients:
      winners = winners_by_client.get(int(client.id), 0)

      client.send_results(winners)
