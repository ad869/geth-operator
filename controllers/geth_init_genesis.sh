#!/bin/sh

set -e

if [ ! -d $DATA_PATH/geth ]; then
    echo "initializing geth genesis block"
    geth init --datadir $DATA_PATH $CONFIG_PATH/genesis.json
else
    echo "genesis block has been initialized before!"
fi

cp $CONFIG_PATH/static-nodes.json $DATA_PATH/geth/static-nodes.json
cp $SECRETS_PATH/account.key $DATA_PATH/geth/nodekey
