import yaml
from dataclasses import dataclass
from typing import Optional


@dataclass
class ServerConfig:
    grpc_address: str
    udp_host: str
    udp_port: int
    log_level: str
    model_path: str


def load_config(path: str = "config.yaml") -> ServerConfig:
    with open(path) as f:
        config = yaml.safe_load(f)

    return ServerConfig(
        grpc_address=config["server"]["grpc_address"],
        udp_host=config["server"]["udp_host"],
        udp_port=config["server"]["udp_port"],
        log_level=config["logging"]["level"],
        model_path=config["model"]["path"],
    )
