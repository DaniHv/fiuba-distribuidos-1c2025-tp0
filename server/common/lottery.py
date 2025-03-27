import logging

from common.clienthandler import ClientHandler
from common.utils import load_bets, has_won

class Lottery:
  clients_done = 0

  def notify_winners(self, clients: 'list[ClientHandler]'):
    """
    Process all bets placed by clients, identify winners and notify them
    with ClientHandler.send_results() method. 
    """
    winners_by_client = {}

    for bet in load_bets():
      if has_won(bet):
        client_results = winners_by_client.get(bet.agency, {})
        client_results[bet.document] = client_results.get(bet.document, 0) + 1

        winners_by_client[bet.agency] = client_results

    logging.info('action: sorteo | result: success')

    for client in clients:
      winners = winners_by_client.get(int(client.id), {})

      client.send_results(winners)
