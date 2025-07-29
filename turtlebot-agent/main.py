
from concurrent import futures
import grpc
import pb.agent_pb2 as agent_pb2
import pb.agent_pb2_grpc as agent_pb2_grpc

class AgentController(agent_pb2_grpc.AgentControlServicer):

    # TODO - Implement here the logic
    def MoveToStation(self, request: agent_pb2.MoveCommand, context):
        print("Sup")
        return agent_pb2.Response(ok=True, onError=f"Got it, moving to {request.stationId}!")
    
    def ReportStatus(self, request: agent_pb2.Empty, context):
        return agent_pb2.Status(stationId="192.168.0.1", batteryState = 15.68, state=agent_pb2.AgentState.IN_BASE)

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=1))
    agent_pb2_grpc.add_AgentControlServicer_to_server(AgentController(), server)

    server.add_insecure_port('[::]:50051')
    server.start()
    print("gRPC server running on port 50051...")
    server.wait_for_termination()

if __name__ == '__main__':
    serve()