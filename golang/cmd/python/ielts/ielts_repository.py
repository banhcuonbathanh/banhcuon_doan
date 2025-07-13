import requests
from python_proto.claude import claude_pb2 as ielts_pb2

class ClaudeRepository:
    def __init__(self):
        self.api_url = "https://api.anthropic.com/v1/messages"  # Replace with the actual Claude API URL
        self.api_key = "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//"  # Replace with your actual API key

    def evaluate_ielts(self, request):
        headers = {
            "Content-Type": "application/json",
            "X-API-Key": self.api_key,
        }

        prompt = f"""
        Evaluate the following IELTS reading response:

        Passage: {request.passage}
        Question: {request.question}
        Student Response: {request.student_response}

        Please consider the following aspects in your evaluation:
        1. Complex Sentences: {request.complex_sentences}
        2. Advanced Vocabulary: {request.advanced_vocabulary}
        3. Cohesive Devices: {request.cohesive_devices}

        Provide a detailed evaluation of the student's response, including strengths and areas for improvement.
        """

        data = {
            "model": "claude-3-sonnet-20240229",
            "messages": [{"role": "user", "content": prompt}],
            "max_tokens": 1000,
        }

        response = requests.post(self.api_url, json=data, headers=headers)
        
        if response.status_code == 200:
            claude_evaluation = response.json()["content"]
            return ielts_pb2.EvaluationResponseFromClaude(evaluation=claude_evaluation)
        else:
            return ielts_pb2.EvaluationResponseFromClaude(evaluation="Error: Unable to get evaluation from Claude.")