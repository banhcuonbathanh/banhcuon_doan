# rag-tutorial-v2

1. python3 -m venv env
2. source env/bin/activate
3. deactivate
   4.pip3 install -r requirements.txt
   5.pip install --upgrade pip

python --version

-----------------------------------------package------------------------------

pip install grpcio grpcio-tools

================================grpc--------------------------------

python -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. proto/helloworld.proto
# Activate your virtual environment if not already activated
   source env/bin/activate  # or source .venv/bin/activate, depending on your setup

   # Check Python version
   python --version

   # Verify grpcio installation
   pip show grpcio

   # Check where grpc is installed
   python -c "import grpc; print(grpc.__file__)"

   # Try importing grpc in Python interactive shell
   python -c "import grpc; print('grpc import successful')"

   ===============



   python server/greeter_server.py