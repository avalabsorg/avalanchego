PKG_ROOT=/tmp/avalanchego
DEBIAN_BASE_DIR=$PKG_ROOT/debian
AVALANCHE_BUILD_BIN_DIR=$DEBIAN_BASE_DIR/usr/local/bin
AVALANCHE_LIB_DIR=$DEBIAN_BASE_DIR/usr/local/lib/avalanchego
TEMPLATE=.github/workflows/debian/template 
DEBIAN_CONF=$DEBIAN_BASE_DIR/DEBIAN

mkdir -p $DEBIAN_BASE_DIR
mkdir -p $DEBIAN_CONF
mkdir -p $AVALANCHE_BUILD_BIN_DIR
mkdir -p $AVALANCHE_LIB_DIR

OK=`cp ./build/avalanchego $AVALANCHE_BUILD_BIN_DIR`
if [[ $OK -ne 0 ]]; then
  exit $OK;
fi
OK=`cp ./build/plugins/evm $AVALANCHE_LIB_DIR`
if [[ $OK -ne 0 ]]; then
  exit $OK;
fi
OK=`cp $TEMPLATE/control-$ARCH $DEBIAN_CONF/control`
if [[ $OK -ne 0 ]]; then
  exit $OK;
fi

echo "Build debian package..."
cd $PKG_ROOT
echo "Tag: $TAG"
VER=$TAG
if [[ $TAG =~ ^v ]]; then
  VER=$(echo $TAG | tr -d 'v')
fi
NEW_VERSION_STRING="Version: $VER"
sed -i "s/Version.*/$NEW_VERSION_STRING/g" debian/DEBIAN/control
dpkg-deb --build debian avalanchego-linux-$ARCH-$TAG.deb
aws s3 cp avalanchego-linux-$ARCH-$TAG.deb s3://$BUCKET/linux/deb/
