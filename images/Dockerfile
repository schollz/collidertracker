FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

# Install build dependencies and runtime tools
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    build-essential \
    cmake \
    git \
    jackd2 \
    libjack-jackd2-dev \
    libsndfile1-dev \
    libfftw3-dev \
    libxt-dev \
    libavahi-client-dev \
    libudev-dev \
    libicu-dev \
    libreadline-dev \
    pkg-config \
    libncurses5-dev \
    rtkit \
    alsa-utils \
    lame \
    nano \
    dbus \
    dbus-user-session \
    ca-certificates \
    python3 python-is-python3 \
    sudo && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

RUN jackd --version
# RUN cd /tmp && \
#     git clone https://github.com/jackaudio/jack2.git && \
#     cd jack2 && \
#     git checkout v1.9.22 && \
#     # Apply patch to fix deprecated 'imp' usage in Context.py
#     sed -i 's/import os, re, imp, sys/import os, re, sys, types/' waflib/Context.py && \
#     sed -i 's/module = imp.new_module(WSCRIPT_FILE)/module = types.ModuleType(WSCRIPT_FILE)/' waflib/Context.py && \
#     chmod +x waf && \
#     ./waf configure --prefix=/usr && \
#     ./waf && \
#     ./waf install && \
#     cd /tmp && rm -rf jack2 && \
#     ldconfig

# Build SuperCollider 3.14.1 from source (headless)
RUN cd /tmp && \
    git clone --branch Version-3.14.1 --recurse-submodules https://github.com/SuperCollider/SuperCollider.git && \
    cd SuperCollider && \
    mkdir build && cd build && \
    cmake -DCMAKE_BUILD_TYPE=Release \
    -DSC_QT=OFF \
    -DSC_IDE=OFF \
    -DSC_ED=OFF \
    -DSC_EL=OFF \
    -DSC_VIM=OFF \
    -DNATIVE=ON \
    -DSUPERNOVA=OFF \
    -DCMAKE_INSTALL_PREFIX=/usr/local \
    .. && \
    make -j$(nproc) && \
    make install && \
    cd /tmp && rm -rf SuperCollider

RUN adduser --disabled-password --gecos "" scuser && \
    usermod -a -G audio,video scuser && \
    mkdir -p /var/log/icecast && \
    chown scuser:scuser /var/log/icecast && \
    mkdir -p /run/user/1000 && \
    chown scuser:scuser /run/user/1000

RUN echo "@audio   -  rtprio     95" >> /etc/security/limits.conf && \
    echo "@audio   -  memlock    unlimited" >> /etc/security/limits.conf && \
    echo "@audio   -  nice      -10" >> /etc/security/limits.conf

RUN /etc/init.d/dbus start

COPY icecast.xml /etc/icecast2/icecast.xml
COPY darkice.cfg /etc/darkice.cfg
COPY autostart.scd /home/scuser/autostart.scd
COPY start.sh /start.sh

USER scuser

EXPOSE 8000 8001
ENTRYPOINT ["/start.sh"]
