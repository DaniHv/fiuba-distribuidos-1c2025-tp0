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
  
  def serialize_array(parts: 'list[str]') -> bytes:
    """
    Serialize a list of ordered list strings into a sequence of binary data
    bytes representation.
    """

    return b'\0'.join(SBDSerialization.serialize(part) for part in parts)

  def deserialize(data: 'bytes', expected_parts: 'Optional[int]') -> 'list[str]':
    """
    Deserialize a sequence of binary data bytes representation
    into an ordered list of strings.

    If expected_parts is not None, it will check if the number of
    elements in the data matches the expected number of elements.
    """

    parts = data.split(b'\0')
  
    if expected_parts is not None and len(parts) != expected_parts:
      raise ValueError(f'Invalid data format ({data}). Expected {expected_parts} elements, got {len(parts)}.')

    return [part.decode('utf-8') for part in parts]

  def deserialize_to_object(data: 'bytes', structure: 'list[str]') -> 'object':
    """
    Deserialize a sequential binary strings bytes representation
    into an object with the given ordered structure.
    """

    parts = SBDSerialization.deserialize(data, len(structure))

    return { structure[i]: parts[i] for i in range(len(parts)) }

  def deserialize_array(data: 'bytes', expected_parts: int) -> 'list[object]':
    """
    Deserialize an array of sequential binary strings bytes representation
    into a list.
    """

    parts = SBDSerialization.deserialize(data, None)

    if expected_parts is not None and len(parts) % expected_parts != 0:
      raise ValueError(f'Invalid data format ({data}). Expected items of {expected_parts} parts, got {len(parts) / expected_parts} (approximately).')

    return [parts[i:i + expected_parts] for i in range(0, len(parts), expected_parts)]