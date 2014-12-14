#!/bin/bash
# uncomment line below if you want verbose output
# set+x 

DEPLOYMENT_PATH=$1
REPOSITORY=$2
# variable init
CUR_TIMESTAMP=`date +"%Y%m%d%H%M%S"`

# update code base with remote_cache strategy
if [ -d "$DEPLOYMENT_PATH/shared/cached-copy" ]
then 
	cd "$DEPLOYMENT_PATH/shared/cached-copy"
	git fetch -q origin
	git fetch --tags -q origin
	git rev-list --max-count=1 HEAD | xargs git reset -q --hard
	git clean -q -d -x -f;
else
	git clone -q $REPOSITORY "$DEPLOYMENT_PATH/shared/cached-copy"
	cd "$DEPLOYMENT_PATH/shared/cached-copy"
	git rev-list --max-count=1 HEAD | xargs git checkout -q -b deploy
fi
cp -RPp "$DEPLOYMENT_PATH/shared/cached-copy" "/usr/share/www/agorema.com/releases/$CUR_TIMESTAMP"
git rev-list --max-count=1 HEAD > "/usr/share/www/agorema.com/releases/$CUR_TIMESTAMP/REVISION"
chmod -R g+w "$DEPLOYMENT_PATH/releases/$CUR_TIMESTAMP"


rm -f "$DEPLOYMENT_PATH/current" &&  ln -s "$DEPLOYMENT_PATH/releases/$CUR_TIMESTAMP" "$DEPLOYMENT_PATH/current"
ls -1dt "$DEPLOYMENT_PATH"/releases/* | tail -n +6 |  xargs rm -rf