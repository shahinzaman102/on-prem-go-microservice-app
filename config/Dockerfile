FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive
ENV MINIKUBE_VERSION=v1.36.0
ENV GOLANG_VERSION=1.24.3

# Install base dependencies
RUN apt-get update && apt-get install -y \
    curl \
    bash \
    iptables \
    conntrack \
    socat \
    docker.io \
    gettext-base \
    ca-certificates \
    # for test & scan dependencies
    git \
    make \
    tar \
    gzip \
    && rm -rf /var/lib/apt/lists/*

# Install kubectl
RUN curl -LO "https://dl.k8s.io/release/$(curl -Ls https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl && \
    rm kubectl

# Install Minikube
RUN curl -Lo /usr/local/bin/minikube https://storage.googleapis.com/minikube/releases/${MINIKUBE_VERSION}/minikube-linux-amd64 && \
    chmod +x /usr/local/bin/minikube

# Install Go (for test & scan dependencies)
RUN curl -LO https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    rm go${GOLANG_VERSION}.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:$PATH"

# Install Trivy (for test & scan dependencies)
RUN curl -sfL https://github.com/aquasecurity/trivy/releases/download/v0.34.0/trivy_0.34.0_Linux-64bit.tar.gz \
    | tar xz -C /usr/local/bin

# Safe user setup
ARG USER_UID=1010
ARG USER_GID=1010
ARG DOCKER_GID=984
ARG USERNAME=gitlab

RUN \
    if ! getent group docker >/dev/null; then groupadd -g ${DOCKER_GID} docker; fi && \
    if ! getent group ${USER_GID} >/dev/null; then groupadd -g ${USER_GID} hostgroup; else groupadd hostgroup; fi && \
    if ! id -u ${USERNAME} >/dev/null 2>&1 && ! getent passwd ${USER_UID} >/dev/null; then \
    useradd -m -s /bin/bash -u ${USER_UID} -g hostgroup -G docker ${USERNAME}; \
    else \
    echo "User '${USERNAME}' or UID ${USER_UID} already exists. Skipping useradd."; \
    fi

# Don't switch user here — let GitLab control that
# USER ${USERNAME}

CMD ["/bin/bash"]
