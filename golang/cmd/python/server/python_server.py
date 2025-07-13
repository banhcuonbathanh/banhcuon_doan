from concurrent import futures
import grpc
import sys
import os

# Add the parent directory to the Python path
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from python_proto import helloworld_pb2, helloworld_pb2_grpc 
from python_proto.claude import claude_pb2, claude_pb2_grpc  
import google.protobuf

print(f"Runtime protobuf version: {google.protobuf.__version__}")
print(f"Generated protobuf version: {helloworld_pb2.DESCRIPTOR.GetOptions().SerializeToString()}")

class Greeter(helloworld_pb2_grpc.GreeterServicer):
    def SayHello(self, request, context):
        return helloworld_pb2.HelloReply(message=f'Hello, {request.name}!')

class IELTSService(claude_pb2_grpc.IELTSServiceServicer):
    def EvaluateIELTS(self, request, context):
        # Here you would implement the logic to evaluate the IELTS response
        # For now, we'll just return an empty response
        return claude_pb2.EvaluationResponseFromToDataBase(
            student_response=request.student_response,
            passage=request.passage,
            question=request.question,
            complex_sentences=request.complex_sentences,
            advanced_vocabulary=request.advanced_vocabulary,
            cohesive_devices=request.cohesive_devices,
            evaluation_from_claude="Evaluation not implemented yet."
        )

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    helloworld_pb2_grpc.add_GreeterServicer_to_server(Greeter(), server)
    claude_pb2_grpc.add_IELTSServiceServicer_to_server(IELTSService(), server)
    server.add_insecure_port('[::]:50052')  # Changed port to 50052
    print("Server starting on port 50052...")
    server.start()
    server.wait_for_termination()

if __name__ == '__main__':
    serve()