import functools
import time
from typing import TypeVar, Callable, Any
import logging

T = TypeVar("T")


def retry(
    max_attempts: int = 3,
    delay: float = 1.0,
    backoff: float = 2.0,
    exceptions: tuple = (Exception,),
) -> Callable:
    def decorator(func: Callable[..., T]) -> Callable[..., T]:
        @functools.wraps(func)
        def wrapper(*args: Any, **kwargs: Any) -> T:
            last_exception = None
            attempt = 0
            current_delay = delay

            while attempt < max_attempts:
                try:
                    return func(*args, **kwargs)
                except exceptions as e:
                    attempt += 1
                    last_exception = e

                    if attempt == max_attempts:
                        break

                    logging.warning(
                        f"Attempt {attempt} failed: {str(e)}. "
                        f"Retrying in {current_delay} seconds..."
                    )

                    time.sleep(current_delay)
                    current_delay *= backoff

            raise last_exception

        return wrapper

    return decorator


class ProcessError(Exception):
    """Base exception for process-related errors"""

    pass


class ModelError(ProcessError):
    """Raised when model operations fail"""

    pass


class DataError(ProcessError):
    """Raised when data operations fail"""

    pass
