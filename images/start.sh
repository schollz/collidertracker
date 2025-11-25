#!/bin/bash

export DISPLAY=:0
export JACK_NO_START_SERVER=1
export JACK_NO_AUDIO_RESERVATION=1
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/usr/local/lib64/

echo "Starting jackd..."
jackd -d dummy -r 48000 &
sleep 2

echo "Starting icecast2..."
icecast2 -c /etc/icecast2/icecast.xml &
sleep 2

echo "Starting darkice..."
darkice -c /etc/darkice.cfg &
sleep 2

echo "Starting sclang..."
sclang /home/scuser/autostart.scd &
sleep 5

echo "Connecting jack ports..."
jack_connect SuperCollider:out1 darkice:left
jack_connect SuperCollider:out2 darkice:right

echo "All services started. Tailing /dev/null to keep container alive."
tail -f /dev/null # stay alive
