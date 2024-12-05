#!/usr/bin/env sh

SKIP_BUILD=${SKIP_BUILD:-""}

# Set up env
export GOPATH=$HOME/go
export GOBIN=$GOPATH/bin
export BUILD_PATH=/kiichain/kiichain3/build
export PATH=$GOBIN:$PATH:/usr/local/go/bin:$BUILD_PATH
echo "export GOPATH=$HOME/go" >> /root/.bashrc
echo "GOBIN=$GOPATH/bin" >> /root/.bashrc
echo "export PATH=$GOBIN:$PATH:/usr/local/go/bin:$BUILD_PATH" >> /root/.bashrc
/bin/bash -c "source /root/.bashrc"
mkdir -p $GOBIN

# Set up cosmovisor env
export DAEMON_NAME="kiichaind"
export DAEMON_HOME="/root/.kiichain3"
export DAEMON_ALLOW_DOWNLOAD_BINARIES="false"
export DAEMON_RESTART_AFTER_UPGRADE="true"
echo "export DAEMON_NAME=$DAEMON_NAME" >> /root/.bashrc
echo "export DAEMON_HOME=$DAEMON_HOME" >> /root/.bashrc
echo "export DAEMON_ALLOW_DOWNLOAD_BINARIES=$DAEMON_ALLOW_DOWNLOAD_BINARIES" >> /root/.bashrc
echo "export DAEMON_RESTART_AFTER_UPGRADE=$DAEMON_RESTART_AFTER_UPGRADE" >> /root/.bashrc

# Step 1 build kiichaind
if [ -z "$SKIP_BUILD" ]
then
  /usr/bin/build.sh
fi

# Run init to set up state sync configurations
/usr/bin/configure_init.sh

# Start the chain
/usr/bin/start_kiichain.sh

tail -f /dev/null
