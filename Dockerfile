# Use an appropriate base image with necessary build tools
FROM ubuntu:20.04

# Install dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    cmake \
    git \
    libboost-all-dev \
    libcurl4-openssl-dev \
    liblzma-dev \
    libopus-dev \
    libpcre3-dev \
    libssl-dev \
    libvpx-dev \
    libz-dev \
    pkg-config \
    yasm \
    && rm -rf /var/lib/apt/lists/*

    
# Set the working directory
WORKDIR /usr/local/src

# Clone the SRS repository
RUN git clone https://github.com/ossrs/srs.git

# Build SRS
WORKDIR /usr/local/src/srs/trunk
RUN ./configure --prefix=/usr/local/srs && \
    make && \
    make install

# Expose necessary ports
#1935 for RTMP & WebRTC signaling
#1985 for API and statistics
#8080 for HLS/LL-HLS
#10080 for WebRTC media transport (UDP)
EXPOSE 1935 8080 1985 10080

# Copy your SRS configuration file
COPY ./config/srs.conf /usr/local/srs/conf/srs.conf

# Set the entrypoint to start SRS
ENTRYPOINT ["/usr/local/srs/objs/srs", "-c", "/usr/local/srs/conf/srs.conf"]