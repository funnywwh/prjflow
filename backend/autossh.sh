#!/bin/bash

export AUTOSSH_POLL=60
autossh -M 0 -f -N -R 8080:localhost:8080 aiweb