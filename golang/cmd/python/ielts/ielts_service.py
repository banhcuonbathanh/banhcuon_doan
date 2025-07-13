from python_proto import ielts_pb2 as ielts_pb2
from python_proto import ielts_pb2_grpc as ielts_pb2_grpc
from ielts.ielts_repository import ClaudeRepository

class IELTSService(ielts_pb2_grpc.IELTSServiceServicer):
    def __init__(self):
        self.claude_repository = ClaudeRepository()

    def EvaluateIELTS(self, request, context):
        # Prepare the request for Claude
        claude_request = ielts_pb2.EvaluationRequestToClaude(
            student_response=request.student_response,
            passage=request.passage,
            question=request.question,
            complex_sentences=request.complex_sentences,
            advanced_vocabulary=request.advanced_vocabulary,
            cohesive_devices=request.cohesive_devices
        )

        # Send request to Claude and get the evaluation
        claude_response = self.claude_repository.evaluate_ielts(claude_request)

        # Prepare the response
        response = ielts_pb2.EvaluationResponseFromToDataBase(
            student_response=request.student_response,
            passage=request.passage,
            question=request.question,
            complex_sentences=request.complex_sentences,
            advanced_vocabulary=request.advanced_vocabulary,
            cohesive_devices=request.cohesive_devices,
            evaluation_from_claude=claude_response.evaluation
        )

        return response