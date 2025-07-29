from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class AgentState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    IN_TRANSIT: _ClassVar[AgentState]
    IN_STATION: _ClassVar[AgentState]
    IN_BASE: _ClassVar[AgentState]
    ON_ERROR: _ClassVar[AgentState]
IN_TRANSIT: AgentState
IN_STATION: AgentState
IN_BASE: AgentState
ON_ERROR: AgentState

class MoveCommand(_message.Message):
    __slots__ = ("stationId",)
    STATIONID_FIELD_NUMBER: _ClassVar[int]
    stationId: str
    def __init__(self, stationId: _Optional[str] = ...) -> None: ...

class Response(_message.Message):
    __slots__ = ("ok", "onError")
    OK_FIELD_NUMBER: _ClassVar[int]
    ONERROR_FIELD_NUMBER: _ClassVar[int]
    ok: bool
    onError: str
    def __init__(self, ok: bool = ..., onError: _Optional[str] = ...) -> None: ...

class Empty(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class Status(_message.Message):
    __slots__ = ("state", "stationId", "batteryState")
    STATE_FIELD_NUMBER: _ClassVar[int]
    STATIONID_FIELD_NUMBER: _ClassVar[int]
    BATTERYSTATE_FIELD_NUMBER: _ClassVar[int]
    state: AgentState
    stationId: str
    batteryState: float
    def __init__(self, state: _Optional[_Union[AgentState, str]] = ..., stationId: _Optional[str] = ..., batteryState: _Optional[float] = ...) -> None: ...
