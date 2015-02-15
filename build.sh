#!/bin/bash
TEMP_DIR=./builds
DEB_DIR=./debs
VERSION=$TRAVIS_BUILD_NUMBER
VERSION=3
echo "Building #$VERSION"

rm -rf $TEMP_DIR
rm -rf $DEB_DIR
mkdir -p $TEMP_DIR
mkdir -p $DEB_DIR
mkdir -p $TEMP_DIR/opt/hello
mkdir -p $TEMP_DIR/etc/hello

go get ./...
gox -osarch="linux/amd64" -output $TEMP_DIR/opt/hello/kenko
#touch $TEMP_DIR/etc/hello/purokishi.yml
cp kenko.conf $TEMP_DIR/etc/hello/kenko.conf

if $(gem list fpm -i) == "true"; then
    echo "fpm found"
else
    echo "installing fpm"
    gem install fpm --no-ri --no-rdoc
fi;

fpm --force -s dir -C $TEMP_DIR -t deb --name "kenko" --version $VERSION --config-files etc/hello .
mv *.deb ./debs/
