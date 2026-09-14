#!/usr/bin/env fish

podman run --rm -it \
    --hostname sandbox \
    --security-opt no-new-privileges \
    --cap-drop ALL \
    --memory 4g --cpus 4 \
    -v sandbox-home:/root \
    sandbox:latest
