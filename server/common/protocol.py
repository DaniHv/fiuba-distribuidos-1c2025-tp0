import socket

class MBPMessage:
    """
    MBPMessage (Message-based Protocol Message) is the data representation of MBPSocket message
    containing an action and data.

    The action is a string (utf8) representing the action to be taken by the counterpart,
    cannot contain spaces or newlines, for example: EXECUTE_SOMETHING. Can be thought as
    the combination of VERB + URL in an HTTP Request.

    The data is a bytes stream containing the data to be sent or received. MBPSocket doesn't
    enforce any restrictions on the data, it could be a plain text, json or any
    other binary data such as protobufs. Can be though as the body of an HTTP Request.
    """
    def __init__(self, action: 'str', data: 'bytes' = b''):
        if ' ' in action or '\n' in action:
            raise ValueError("Action cannot contain spaces or newlines")

        self.action = action
        self.data = data

    def to_bytes(self):
        return bytes(f"{self.action} ", 'utf-8') + self.data
    
    def from_bytes(data: 'bytes') -> 'MBPMessage':
        action, data = data.split(b' ', 1)

        return MBPMessage(action.decode('utf8'), data)

class MBPSocket:
    """
    MBPSocket (Message-based Protocol Socket) is a TCP-Based, full-duplex and message-oriented protocol.

    Messages received and sent are represented by MBPMessage objects,
    containing an action and data, both strings.

    Server socket accepts connections non-blocking. Child sockets sends and receives messages blocking.

    Note: "connect" method is not implemented since it won't be required, since this will be used
    as server-only.
    """

    listening = False
    buffer = bytes()

    def __init__(self):
        pass

    def listen(self, port, listen_backlog):
        """
        Listen for incoming connections on a given port.
        """

        if self.listening:
            raise Exception("Cannot listen in an already listening socket.")

        self._socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._socket.bind(('', port))
        self._socket.listen(listen_backlog)
        self._socket.setblocking(False)
        self.listening = True

    def accept(self) -> 'MBPSocket':
        """"
        Accept a new connection on a listening socket.
        """

        if not self.listening:
            raise Exception("Cannot accept connections in a non listening socket.")
        
        c, addr = self._socket.accept()
        c.setblocking(True)

        client_socket = MBPSocket()
        client_socket._socket = c

        return client_socket, addr
    
    def send_message(self, message: 'MBPMessage'):
        """
        Receive a message (MBPMessage) to the connected socket.

        Only available in non-listening sockets.
        """

        if self.listening:
            raise Exception("Cannot send message in a listening socket")
        
        self._socket.sendall(message.to_bytes() + b'\n')

    def receive_message(self) -> 'MBPMessage':
        """
        Receive a message (MBPMessage) from the connected socket.

        Only available in non-listening sockets.
        """

        if self.listening:
            raise Exception("Cannot receive message in a listening socket")

        line = self._receive_line()

        return MBPMessage.from_bytes(line)

    def getpeername(self):
        return self._socket.getpeername()

    def close(self):
        self._socket.shutdown(socket.SHUT_RDWR)
        self._socket.close()
        self.listening = False

    def fileno(self):
        return self._socket.fileno()
    
    # Read a line from the socket, buffering until a newline is found.
    def _receive_line(self):
        while True:
            if b'\n' in self.buffer:
                index = self.buffer.index(b'\n')
                line = self.buffer[:index]
                self.buffer = self.buffer[index+1:]
                return line

            new_chunk = self._socket.recv(1024)

            if len(new_chunk) == 0:
                raise Exception(f'Unexpected EOF while reading line. Current buffer: {self.buffer}, chunk: {new_chunk}')

            self.buffer += new_chunk
