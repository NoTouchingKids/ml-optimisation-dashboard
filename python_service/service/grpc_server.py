import json
from typing import Iterator
import grpc
import logging
from concurrent import futures
import proto.process_pb2 as pb2
import proto.process_pb2_grpc as pb2_grpc


class ProcessService(pb2_grpc.ProcessServiceServicer):
    def __init__(self, process_manager, udp_server):
        self.process_manager = process_manager
        self.udp_server = udp_server

    def StartProcess(self, request, context):
        print(request)
        print(context)
        try:
            logging.info(f"Received request: {request}")
            config = json.loads(request.payload)
            logging.info(f"Parsed config: {config}")

            # Validate required fields
            if "client_id" not in config:
                raise ValueError("client_id is required")

            if config.get("type", "train") == "train":
                if not config.get("train_start_date") or not config.get(
                    "train_end_date"
                ):
                    raise ValueError("train_start_date and train_end_date are required")

            process = self.process_manager.start_process(config["client_id"], config)

            return pb2.ProcessResponse(
                client_id=config["client_id"], process_id=process.pid, status="started"
            )
        except Exception as e:
            logging.error(f"Process start failed: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return pb2.ProcessResponse()

    def StreamLogs(self, request, context) -> Iterator[pb2.LogMessage]:
        while context.is_active():
            try:
                log = self.udp_server.get_log(timeout=0.1)  # Non-blocking with timeout

                if log:
                    yield pb2.LogMessage(**log)
            except Exception as e:
                logging.error(f"StreamLogs error: {e}")
                context.abort(grpc.StatusCode.INTERNAL, str(e))


class GRPCServer:
    def __init__(self, address, process_manager, udp_server):
        self.server = grpc.server(futures.ThreadPoolExecutor(max_workers=3))
        self.address = address
        self.service = ProcessService(process_manager, udp_server)
        pb2_grpc.add_ProcessServiceServicer_to_server(self.service, self.server)

    def start(self):
        try:
            self.server.add_insecure_port(self.address)
            self.server.start()
            logging.info(f"gRPC server started on {self.address}")
        except Exception as e:
            logging.error(f"Failed to start gRPC server: {e}")
            raise

    def stop(self):
        self.server.stop(0)
