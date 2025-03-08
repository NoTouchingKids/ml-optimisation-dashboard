import json
from typing import AsyncIterator
import grpc
import logging
import asyncio
import proto.process_pb2 as pb2
import proto.process_pb2_grpc as pb2_grpc

from service.process_manager import ProcessManager
from service.udp_server import TCPLogServer
from service.health import HealthChecker

logger = logging.getLogger(__file__)


class ProcessService(pb2_grpc.ProcessServiceServicer):
    def __init__(self, process_manager: ProcessManager, udp_server: TCPLogServer):
        self.process_manager = process_manager
        self.udp_server = udp_server

    async def StartProcess(self, request, context):
        try:
            logger.info(f"Received request: {request}")
            config = json.loads(request.payload)
            logger.info(f"Parsed config: {config}")

            # Validate required fields
            if "client_id" not in config:
                raise ValueError("client_id is required")

            if config.get("type", "train") == "train":
                if not config.get("train_start_date") or not config.get(
                    "train_end_date"
                ):
                    raise ValueError("train_start_date and train_end_date are required")

            # Offload the blocking call to the executor
            loop = asyncio.get_running_loop()
            process = await loop.run_in_executor(
                None, self.process_manager.start_process, config["client_id"], config
            )

            return pb2.ProcessResponse(
                client_id=config["client_id"], process_id=process.pid, status="started"
            )
        except Exception as e:
            logger.error(f"Process start failed: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return pb2.ProcessResponse()

    async def StreamLogs(self, request, context) -> AsyncIterator[pb2.LogMessage]:
        # loop = asyncio.get_running_loop()
        try:
            while not context.cancelled():
                try:
                    # Use wait_for to timeout the blocking call, allowing cancellation checks
                    try:
                        # log = await asyncio.wait_for(
                        #     loop.run_in_executor(None, self.udp_server.get_log, 0.1),
                        #     timeout=0.2,
                        # )
                        log = self.udp_server.get_log()
                        # print(log)
                    except asyncio.TimeoutError:
                        log = None

                    if log:
                        yield pb2.LogMessage(**log)
                    else:
                        await asyncio.sleep(0.1)
                except Exception as e:
                    logger.error(f"StreamLogs error: {e}")
                    context.abort(grpc.StatusCode.INTERNAL, str(e))
        except asyncio.CancelledError:
            logger.error("StreamLogs cancelled.", stack_info=True)
            raise


class GRPCServer:
    def __init__(self, address, process_manager, udp_server):
        self.address = address
        self.process_manager = process_manager
        self.udp_server = udp_server
        self.server = None  # We'll create it in start()

    async def start(self):
        # Now that we're in the running event loop, create the server
        self.server = grpc.aio.server()
        pb2_grpc.add_ProcessServiceServicer_to_server(
            ProcessService(self.process_manager, self.udp_server), self.server
        )
        self.server.add_insecure_port(self.address)
        await self.server.start()
        logger.info(f"gRPC server started on {self.address}")
        # Now the server is attached to the correct event loop

    async def stop(self):
        await self.server.stop(0)
