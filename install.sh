#!/bin/sh

# Overridden variables from Go template: {{.Overrides}}

docker_install() {
    echo "🐳 Installing ShellHub using docker method..."

    PREFERRED_HOSTNAME_ARG="-e SHELLHUB_PREFERRED_HOSTNAME=$PREFERRED_HOSTNAME"
    PREFERRED_IDENTITY_ARG="-e SHELLHUB_PREFERRED_IDENTITY=$PREFERRED_IDENTITY"

    echo "📥 Downloading ShellHub container image..."

    {
        docker pull -q shellhubio/agent:$AGENT_VERSION
    } || { echo "❌ Failed to download shellhub container image."; exit 1; }

    echo "🚀 Starting ShellHub container..."

    docker run -d \
       --name=$CONTAINER_NAME \
       --restart=on-failure \
       --privileged \
       --net=host \
       --pid=host \
       -v /:/host \
       -v /dev:/dev \
       -v /var/run/docker.sock:/var/run/docker.sock \
       -v /etc/passwd:/etc/passwd \
       -v /etc/group:/etc/group \
       -v /etc/resolv.conf:/etc/resolv.conf \
       -v /var/run:/var/run \
       -v /var/log:/var/log \
       -e SHELLHUB_SERVER_ADDRESS=$SERVER_ADDRESS \
       -e SHELLHUB_PRIVATE_KEY=/host/etc/shellhub.key \
       -e SHELLHUB_TENANT_ID=$TENANT_ID \
       $PREFERRED_HOSTNAME_ARG \
       $PREFERRED_IDENTITY_ARG \
       shellhubio/agent:$AGENT_VERSION
}

snap_install() {
    echo "📦 Installing ShellHub using snap method..."

    if ! type snap > /dev/null 2>&1; then
        echo "❌ Snap is not installed or not supported on this system."
        exit 1
    fi

    echo "📥 Downloading ShellHub snap package..."

    {
        sudo snap install shellhub-agent --channel=$AGENT_VERSION
    } || { echo "❌ Failed to download and install ShellHub snap package."; exit 1; }

    echo "🚀 Starting ShellHub snap service..."

    {
        sudo snap start shellhub-agent
    } || { echo "❌ Failed to start ShellHub snap service."; exit 1; }
}

standalone_install() {
    echo "🐧 Installing ShellHub using standalone method..."

    INSTALL_DIR="${INSTALL_DIR:-/opt/shellhub}"

    if [ "$(id -u)" -ne 0 ]; then
        printf "⚠️ NOTE: This install method requires root privileges.\n"
        SUDO="sudo"
    fi

    if ! systemctl show-environment > /dev/null 2>&1 ; then
        printf "❌ ERROR: This is not a systemd-based operation system. Unable to proceed with the requested action.\n"
        exit 1
    fi

    echo "📥 Downloading required files..."

    {
        download "https://github.com/opencontainers/runc/releases/download/${RUNC_VERSION}/runc.${RUNC_ARCH}" $TMP_DIR/runc && chmod 755 $TMP_DIR/runc
    } || { rm -rf $TMP_DIR && echo "❌ Failed to download runc binary." && exit 1; }

    {
        download https://raw.githubusercontent.com/shellhub-io/shellhub/${AGENT_VERSION}/agent/packaging/config.json $TMP_DIR/config.json
    } ||  { rm -rf $TMP_DIR && echo "❌ Failed to download OCI runtime spec." && exit 1; }

    {
        download https://raw.githubusercontent.com/shellhub-io/shellhub/${AGENT_VERSION}/agent/packaging/shellhub-agent.service $TMP_DIR/shellhub-agent.service
    } || { rm -rf $TMP_DIR && echo "❌ Failed to download systemd service file." && exit 1; }


    {
        download https://github.com/shellhub-io/shellhub/releases/download/$AGENT_VERSION/rootfs-$AGENT_ARCH.tar.gz $TMP_DIR/rootfs.tar.gz
    } || { rm -rf $TMP_DIR && echo "❌ Failed to download rootfs." && exit 1; }

    echo "📂 Extracting files..."

    {
        mkdir -p $TMP_DIR/rootfs && tar -C $TMP_DIR/rootfs -xzf $TMP_DIR/rootfs.tar.gz && rm -f $TMP_DIR/rootfs.tar.gz
    } || { rm -rf $TMP_DIR && echo "❌ Failed to extract rootfs." && exit 1; }

    rm -f $TMP_DIR/rootfs/.dockerenv

    sed -i "s,__SERVER_ADDRESS__,$SERVER_ADDRESS,g" $TMP_DIR/config.json
    sed -i "s,__TENANT_ID__,$TENANT_ID,g" $TMP_DIR/config.json
    sed -i "s,__ROOT_PATH__,$INSTALL_DIR/rootfs,g" $TMP_DIR/config.json
    sed -i "s,__INSTALL_DIR__,$INSTALL_DIR,g" $TMP_DIR/shellhub-agent.service

    echo "🚀 Starting ShellHub system service..."

    $SUDO cp $TMP_DIR/shellhub-agent.service /etc/systemd/system/shellhub-agent.service
    $SUDO systemctl enable --now shellhub-agent || { rm -rf $TMP_DIR && echo "❌ Failed to enable systemd service."; exit 1; }

    $SUDO rm -rf $INSTALL_DIR
    $SUDO mv $TMP_DIR $INSTALL_DIR
    $SUDO rm -rf $TMP_DIR
}

download() {
    _DOWNLOAD_URL=$1
    _DOWNLOAD_OUTPUT=$2

    if type curl > /dev/null 2>&1; then
        curl -fsSL $_DOWNLOAD_URL --output $_DOWNLOAD_OUTPUT
    elif type wget > /dev/null 2>&1; then
        wget -q -O $_DOWNLOAD_OUTPUT $_DOWNLOAD_URL
    fi
}

http_get() {
    _HTTP_GET_URL=$1

    if type curl > /dev/null 2>&1; then
        curl -sk $_HTTP_GET_URL
    elif type wget > /dev/null 2>&1; then
        wget -q -O - $_HTTP_GET_URL
    fi
}

if [ "$(uname -s)" = "FreeBSD" ]; then
    echo "👹 This system is running FreeBSD."
    echo "❌ ERROR: Automatic installation is not supported on FreeBSD."
    echo
    echo "Please refer to the ShellHub port at https://github.com/shellhub-io/ports"
    exit 1
fi

[ -z "$TENANT_ID" ] && { echo "ERROR: TENANT_ID is missing."; exit 1; }

SERVER_ADDRESS="${SERVER_ADDRESS:-https://cloud.shellhub.io}"
TENANT_ID="${TENANT_ID}"
INSTALL_METHOD="$INSTALL_METHOD"
AGENT_VERSION="${AGENT_VERSION:-$(http_get $SERVER_ADDRESS/info | sed -E 's/.*"version":\s?"?([^,"]*)"?.*/\1/')}"
AGENT_ARCH="$AGENT_ARCH"
CONTAINER_NAME="${CONTAINER_NAME:-shellhub}"
RUNC_VERSION=${RUNC_VERSION:-v1.1.3}
RUNC_ARCH=$RUNC_ARCH
INSTALL_DIR="${INSTALL_DIR:-/opt/shellhub}"
TMP_DIR="${TMP_DIR:-`mktemp -d -t shellhub-installer-XXXXXX`}"

if type docker > /dev/null 2>&1; then
    while :; do
        if $SUDO docker info > /dev/null 2>&1; then
            INSTALL_METHOD="${INSTALL_METHOD:-docker}"
            break
        elif [ "$(id -u)" -ne 0 ]; then
            [ -z "$SUDO" ] && SUDO="sudo" || { SUDO="" && break; }
        fi
    done
fi

if [ -z "$INSTALL_METHOD" ] && type snap > /dev/null 2>&1; then
    INSTALL_METHOD="snap"
fi

INSTALL_METHOD="${INSTALL_METHOD:-standalone}"

# Auto detect arch if it has not already been set
if [ -z "$AGENT_ARCH" ]; then
    case `uname -m` in
        x86_64)
            AGENT_ARCH=amd64
            RUNC_ARCH=amd64
            ;;
        armv6l)
            AGENT_ARCH=arm32v6
            RUNC_ARCH=armel
            ;;
        armv7l)
            AGENT_ARCH=arm32v7
            RUNC_ARCH=armhf
            ;;
        aarch64)
            AGENT_ARCH=arm64v8
            RUNC_ARCH=arm64
    esac
fi

echo "🛠️ Welcome to the ShellHub Agent Installer Script"
echo
echo "📝 Summary of chosen options:"
echo "- Server address: $SERVER_ADDRESS"
echo "- Tenant ID: $TENANT_ID"
echo "- Install method: $INSTALL_METHOD"
echo "- Agent version: $AGENT_VERSION"
echo

case "$INSTALL_METHOD" in
    docker)
        docker_install
        ;;
    snap)
        snap_install
        ;;
    standalone)
        standalone_install
        ;;
    *)
        echo "❌ Install method not supported."
        exit 1
esac
