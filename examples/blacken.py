"""A test that checks whether its own code is correctly formatted with black."""

from black import main

main(
    [
        "--color",
        "--check",
        "--config=pyproject.toml",
        __file__,
    ]
)
