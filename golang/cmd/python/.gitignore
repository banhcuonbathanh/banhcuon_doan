from concurrent import futures
import grpc
import sys
import os

# Add the parent directory to the Python path
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from proto import helloworld_pb2 as helloworld_pb2
from proto import helloworld_pb2_grpc as helloworld_pb2_grpc
import google.protobuf

print(f"Runtime protobuf version: {google.protobuf.__version__}")
print(f"Generated protobuf version: {helloworld_pb2.DESCRIPTOR.GetOptions().SerializeToString()}")

class Greeter(helloworld_pb2_grpc.GreeterServicer):
    def SayHello(self, request, context):
        return helloworld_pb2.HelloReply(message=f'Hello, {request.name}!')

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    helloworld_pb2_grpc.add_GreeterServicer_to_server(Greeter(), server)
    server.add_insecure_port('[::]:50052')  # Changed port to 50052
    print("Server starting on port 50052...")
    server.start()
    server.wait_for_termination()

if __name__ == '__main__':
    serve()
