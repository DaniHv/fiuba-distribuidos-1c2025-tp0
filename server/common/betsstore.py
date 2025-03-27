import threading

from common.utils import Bet, store_bets

class BetsSharedStore:
  """
  Thread-safe store for lottery bets.
  """
  def __init__(self):
    self._lock = threading.Lock()

  # This method is used to store a bet in the shared store.
  # This methods can be called by multiple threads at the same time, and will
  # block until the bets are stored.
  def store_bets(self, bets: 'list[Bet]'):
    with self._lock.acquire():
      store_bets(bets)