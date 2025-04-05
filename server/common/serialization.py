from typing import Optional

class SBDSerialization:
  """
  SBDSerialization (Sequential Binary Data Serialization) provides
  binary serialization and deserialization for a list of ordered strings.
  """

  def serialize(parts: 'list[str]') -> bytes:
    """
    Serialize an ordered list of strings into a sequence of binary data
    bytes representation.
    """

    return b'\0'.join(part.encode('utf-8') for part in parts)

  def deserialize(data: 'bytes', expected_elements: 'Optional[int]') -> 'list[str]':
    """
    Deserialize a sequence of binary data bytes representation
    into an ordered list of strings.

    If expected_elements is not None, it will check if the number of
    elements in the data matches the expected number of elements.
    """

    parts = data.split(b'\0')
  
    if expected_elements is not None and len(parts) != expected_elements:
      raise ValueError(f'Invalid data format ({data}). Expected {expected_elements} elements, got {len(parts)}.')

    return [part.decode('utf-8') for part in parts]

  def deserialize_to_object(data: 'bytes', structure: 'list[str]') -> 'object':
    """
    Deserialize a sequential binary strings bytes representation
    into an object with the given ordered structure.
    """

    parts = SBDSerialization.deserialize(data, len(structure))

    return { structure[i]: parts[i] for i in range(len(parts)) }
